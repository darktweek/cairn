package service

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/darktweek/cairn/internal/config"
	"github.com/darktweek/cairn/internal/model"
	"github.com/darktweek/cairn/internal/repository"
	"github.com/oklog/ulid/v2"
	"golang.org/x/net/html"
)

// BookmarkInput carries the mutable fields of a bookmark create/update request.
type BookmarkInput struct {
	URL          string
	Title        string
	Hidden       bool
	CollectionID string // empty = the user's personal collection
	FolderID     *string
	Tags         []string
}

type BookmarkService interface {
	Create(ctx context.Context, userID string, in BookmarkInput) (*model.Bookmark, error)
	Update(ctx context.Context, userID, bookmarkID string, in BookmarkInput) error
	Delete(ctx context.Context, userID, bookmarkID string) error
	List(ctx context.Context, userID string, filter repository.BookmarkFilter) ([]*model.Bookmark, int, error)
	ListInCollection(ctx context.Context, userID, collectionID string, filter repository.BookmarkFilter) ([]*model.Bookmark, int, error)
	UpdateSort(ctx context.Context, userID string, ids []string) error
	ImportNetscape(ctx context.Context, userID string, data []byte) (imported, skipped int, err error)
	ExportNetscape(ctx context.Context, userID string) ([]byte, error)
	GenerateBookmarklet(ctx context.Context, userID, ip, userAgent string) (string, error)
	ListTags(ctx context.Context, userID string) ([]*model.Tag, error)
	DeleteTag(ctx context.Context, userID, tagID string) error
	ClearAll(ctx context.Context, userID string) error
}

type bookmarkService struct {
	repos *repository.Repositories
	cfg   *config.Config
	auth  AuthService
}

func newBookmarkService(repos *repository.Repositories, cfg *config.Config, auth AuthService) BookmarkService {
	return &bookmarkService{repos: repos, cfg: cfg, auth: auth}
}

// requirePerm checks the user holds at least `need` on the collection. No access
// at all is reported as ErrNotFound (so the collection's existence stays hidden);
// insufficient-but-present access is ErrForbidden.
func (s *bookmarkService) requirePerm(ctx context.Context, userID, collectionID, need string) error {
	perm, err := s.repos.Collections.EffectivePerm(ctx, userID, collectionID)
	if err != nil {
		return err
	}
	if perm == "" {
		return ErrNotFound
	}
	if !model.PermAtLeast(perm, need) {
		return ErrForbidden
	}
	return nil
}

// resolveTarget returns the target collection ID (defaulting to the user's
// personal collection) and validates that the optional folder lives in it.
func (s *bookmarkService) resolveTarget(ctx context.Context, userID, collectionID string, folderID *string) (string, *string, error) {
	if collectionID == "" {
		personal, err := s.repos.Collections.GetOrCreatePersonal(ctx, userID)
		if err != nil {
			return "", nil, err
		}
		collectionID = personal.ID
	}
	if err := s.requirePerm(ctx, userID, collectionID, model.PermEdit); err != nil {
		return "", nil, err
	}
	if folderID != nil {
		if *folderID == "" {
			folderID = nil
		} else {
			f, err := s.repos.Folders.GetByID(ctx, *folderID)
			if err != nil || f.CollectionID != collectionID {
				return "", nil, fmt.Errorf("%w: folder does not belong to collection", ErrInvalidInput)
			}
		}
	}
	return collectionID, folderID, nil
}

func (s *bookmarkService) Create(ctx context.Context, userID string, in BookmarkInput) (*model.Bookmark, error) {
	if !isValidURL(in.URL) {
		return nil, fmt.Errorf("%w: invalid URL", ErrInvalidInput)
	}

	collectionID, folderID, err := s.resolveTarget(ctx, userID, in.CollectionID, in.FolderID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	b := &model.Bookmark{
		ID:           ulid.Make().String(),
		UserID:       userID,
		CollectionID: collectionID,
		FolderID:     folderID,
		URL:          in.URL,
		Title:        in.Title,
		Hidden:       in.Hidden,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repos.Bookmarks.Create(ctx, b); err != nil {
		return nil, err
	}
	if err := s.applyTags(ctx, b.ID, userID, in.Tags); err != nil {
		return nil, err
	}
	return s.repos.Bookmarks.GetByID(ctx, b.ID)
}

func (s *bookmarkService) Update(ctx context.Context, userID, bookmarkID string, in BookmarkInput) error {
	if !isValidURL(in.URL) {
		return fmt.Errorf("%w: invalid URL", ErrInvalidInput)
	}

	b, err := s.repos.Bookmarks.GetByID(ctx, bookmarkID)
	if err != nil {
		return ErrNotFound
	}
	// Must be able to edit the bookmark's current collection.
	if err := s.requirePerm(ctx, userID, b.CollectionID, model.PermEdit); err != nil {
		return err
	}

	// A move keeps the current collection when none is supplied.
	target := in.CollectionID
	if target == "" {
		target = b.CollectionID
	}
	collectionID, folderID, err := s.resolveTarget(ctx, userID, target, in.FolderID)
	if err != nil {
		return err
	}

	b.CollectionID = collectionID
	b.FolderID = folderID
	b.URL = in.URL
	b.Title = in.Title
	b.Hidden = in.Hidden
	b.UpdatedAt = time.Now()

	if err := s.repos.Bookmarks.Update(ctx, b); err != nil {
		return err
	}
	return s.applyTags(ctx, bookmarkID, userID, in.Tags)
}

func (s *bookmarkService) Delete(ctx context.Context, userID, bookmarkID string) error {
	b, err := s.repos.Bookmarks.GetByID(ctx, bookmarkID)
	if err != nil {
		return ErrNotFound
	}
	if err := s.requirePerm(ctx, userID, b.CollectionID, model.PermEdit); err != nil {
		return err
	}
	return s.repos.Bookmarks.Delete(ctx, bookmarkID)
}

func (s *bookmarkService) List(ctx context.Context, userID string, filter repository.BookmarkFilter) ([]*model.Bookmark, int, error) {
	return s.repos.Bookmarks.ListByUser(ctx, userID, filter)
}

func (s *bookmarkService) ListInCollection(ctx context.Context, userID, collectionID string, filter repository.BookmarkFilter) ([]*model.Bookmark, int, error) {
	if err := s.requirePerm(ctx, userID, collectionID, model.PermView); err != nil {
		return nil, 0, err
	}
	return s.repos.Bookmarks.ListByCollection(ctx, collectionID, filter)
}

func (s *bookmarkService) UpdateSort(ctx context.Context, userID string, ids []string) error {
	// Verify the user can edit every bookmark before reordering.
	for _, id := range ids {
		b, err := s.repos.Bookmarks.GetByID(ctx, id)
		if err != nil {
			return ErrNotFound
		}
		if err := s.requirePerm(ctx, userID, b.CollectionID, model.PermEdit); err != nil {
			return err
		}
	}
	return s.repos.Bookmarks.UpdateSort(ctx, ids)
}

func (s *bookmarkService) ImportNetscape(ctx context.Context, userID string, data []byte) (int, int, error) {
	items, err := parseNetscape(data)
	if err != nil {
		return 0, 0, fmt.Errorf("%w: %s", ErrInvalidInput, err.Error())
	}

	personal, err := s.repos.Collections.GetOrCreatePersonal(ctx, userID)
	if err != nil {
		return 0, 0, err
	}

	now := time.Now()
	// folderCache maps a null-byte-joined path to its folder ID.
	folderCache := map[string]string{}

	// getOrCreateFolder resolves (and caches) a folder for the given path,
	// creating parent folders as needed.
	var getOrCreateFolder func(path []string) (string, error)
	getOrCreateFolder = func(path []string) (string, error) {
		key := strings.Join(path, "\x00")
		if id, ok := folderCache[key]; ok {
			return id, nil
		}
		var parentID *string
		if len(path) > 1 {
			pid, err := getOrCreateFolder(path[:len(path)-1])
			if err != nil {
				return "", err
			}
			parentID = &pid
		}
		f, err := s.repos.Folders.GetOrCreate(ctx, personal.ID, parentID, path[len(path)-1])
		if err != nil {
			return "", err
		}
		folderCache[key] = f.ID
		return f.ID, nil
	}

	var toInsert []*model.Bookmark
	skipped := 0

	for _, item := range items {
		if !isValidURL(item.url) {
			skipped++
			continue
		}
		var folderID *string
		if len(item.folderPath) > 0 {
			id, err := getOrCreateFolder(item.folderPath)
			if err != nil {
				return 0, 0, err
			}
			folderID = &id
		}
		toInsert = append(toInsert, &model.Bookmark{
			ID:           ulid.Make().String(),
			UserID:       userID,
			CollectionID: personal.ID,
			FolderID:     folderID,
			URL:          item.url,
			Title:        item.title,
			CreatedAt:    now,
			UpdatedAt:    now,
		})
	}

	if len(toInsert) == 0 {
		return 0, skipped, nil
	}
	if err := s.repos.Bookmarks.BulkCreate(ctx, toInsert); err != nil {
		return 0, 0, err
	}

	imported := len(toInsert)
	_ = s.repos.Audit.Log(ctx, &model.AuditEntry{
		ID:        ulid.Make().String(),
		UserID:    &userID,
		Action:    "bookmark_import",
		Metadata:  map[string]any{"imported": imported, "skipped": skipped},
		CreatedAt: now,
	})

	return imported, skipped, nil
}

var netscapeTmpl = template.Must(template.New("netscape").Parse(`<!DOCTYPE NETSCAPE-Bookmark-file-1>
<!-- This is an automatically generated file.
     It will be read and overwritten.
     DO NOT EDIT! -->
<META HTTP-EQUIV="Content-Type" CONTENT="text/html; charset=UTF-8">
<TITLE>Bookmarks</TITLE>
<H1>Bookmarks</H1>
<DL><p>
{{- range .}}
{{- if .Folder}}
    <DT><H3>{{.Folder}}</H3>
    <DL><p>
{{- end}}
    <DT><A HREF="{{.URL}}" ADD_DATE="{{.AddDate}}">{{.Title}}</A>
{{- if .Folder}}
    </DL><p>
{{- end}}
{{end}}
</DL>
`))

func (s *bookmarkService) ExportNetscape(ctx context.Context, userID string) ([]byte, error) {
	bookmarks, _, err := s.repos.Bookmarks.ListByUser(ctx, userID, repository.BookmarkFilter{Limit: 100000})
	if err != nil {
		return nil, err
	}

	// Resolve folder names across the user's collections for the folder headings.
	folderNames, err := s.userFolderNames(ctx, userID)
	if err != nil {
		return nil, err
	}

	type entry struct {
		URL     string
		Title   string
		Folder  string
		AddDate int64
	}

	var entries []entry
	for _, b := range bookmarks {
		folder := ""
		if b.FolderID != nil {
			folder = folderNames[*b.FolderID]
		}
		entries = append(entries, entry{
			URL:     b.URL,
			Title:   b.Title,
			Folder:  folder,
			AddDate: b.CreatedAt.Unix(),
		})
	}

	var buf bytes.Buffer
	if err := netscapeTmpl.Execute(&buf, entries); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// userFolderNames builds a folderID -> name map across the user's collections.
func (s *bookmarkService) userFolderNames(ctx context.Context, userID string) (map[string]string, error) {
	ids, err := s.repos.Collections.AccessibleCollectionIDs(ctx, userID)
	if err != nil {
		return nil, err
	}
	names := map[string]string{}
	for _, cid := range ids {
		folders, err := s.repos.Folders.ListByCollection(ctx, cid)
		if err != nil {
			return nil, err
		}
		for _, f := range folders {
			names[f.ID] = f.Name
		}
	}
	return names, nil
}

func (s *bookmarkService) GenerateBookmarklet(ctx context.Context, userID, ip, userAgent string) (string, error) {
	token, err := s.auth.CreateBookmarkletSession(ctx, userID, ip, userAgent)
	if err != nil {
		return "", err
	}

	js := fmt.Sprintf(`javascript:(function(){`+
		`var u=encodeURIComponent(location.href);`+
		`var t=encodeURIComponent(document.title);`+
		`fetch('%s/api/bookmarks/quick',{`+
		`method:'POST',`+
		`headers:{'Content-Type':'application/json'},`+
		`body:JSON.stringify({url:u,title:t,token:'%s'})`+
		`}).then(function(){alert('Bookmark sauvegardé')}).catch(function(){alert('Erreur')});`+
		`})();`,
		s.cfg.BaseURL, token,
	)

	return js, nil
}

func (s *bookmarkService) ListTags(ctx context.Context, userID string) ([]*model.Tag, error) {
	return s.repos.Tags.ListByUser(ctx, userID)
}

func (s *bookmarkService) DeleteTag(ctx context.Context, userID, tagID string) error {
	return s.repos.Tags.Delete(ctx, tagID, userID)
}

func (s *bookmarkService) ClearAll(ctx context.Context, userID string) error {
	if err := s.repos.Bookmarks.DeleteAllByUser(ctx, userID); err != nil {
		return err
	}
	_ = s.repos.Audit.Log(ctx, &model.AuditEntry{
		ID:        ulid.Make().String(),
		UserID:    &userID,
		Action:    "bookmark_clear_all",
		CreatedAt: time.Now(),
	})
	return nil
}

func (s *bookmarkService) applyTags(ctx context.Context, bookmarkID, userID string, tagNames []string) error {
	var tagIDs []string
	for _, name := range tagNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		t, err := s.repos.Tags.GetOrCreate(ctx, userID, name)
		if err != nil {
			return err
		}
		tagIDs = append(tagIDs, t.ID)
	}
	return s.repos.Bookmarks.SetTags(ctx, bookmarkID, tagIDs)
}

// netscapeItem holds a parsed bookmark from Netscape HTML.
type netscapeItem struct {
	url        string
	title      string
	folderPath []string // ordered path segments; nil = no folder
}

// parseNetscape parses the Netscape bookmarks HTML format.
// It reconstructs the full folder hierarchy, not just the immediate parent.
//
// Note: Go's html.Parse (HTML5) places the nested <DL> *inside* the <DT>,
// so the actual tree is DT > [H3, DL] rather than DT and DL being siblings.
func parseNetscape(data []byte) ([]netscapeItem, error) {
	doc, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	var items []netscapeItem
	var walk func(n *html.Node, path []string)
	walk = func(n *html.Node, path []string) {
		if n.Type != html.ElementNode {
			return
		}
		switch strings.ToUpper(n.Data) {
		case "A":
			item := netscapeItem{folderPath: path}
			for _, attr := range n.Attr {
				if strings.ToUpper(attr.Key) == "HREF" {
					item.url = attr.Val
				}
			}
			item.title = textContent(n)
			if item.url != "" {
				items = append(items, item)
			}
			return // don't recurse into <A>

		case "DT":
			if h3 := firstChildByTag(n, "H3"); h3 != nil {
				// Folder heading: add segment to path, then recurse into DT's
				// children (the nested <DL> lives inside this same <DT>).
				folderPath := append(append([]string(nil), path...), textContent(h3))
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if c.Type == html.ElementNode && strings.ToUpper(c.Data) == "H3" {
						continue // already captured above
					}
					walk(c, folderPath)
				}
				return
			}
			// DT containing a bookmark — recurse normally.
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c, path)
		}
	}

	for c := doc.FirstChild; c != nil; c = c.NextSibling {
		walk(c, nil)
	}
	return items, nil
}

// firstChildByTag returns the first direct element child of n with the given tag name.
func firstChildByTag(n *html.Node, tag string) *html.Node {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && strings.ToUpper(c.Data) == tag {
			return c
		}
	}
	return nil
}

func textContent(n *html.Node) string {
	var sb strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			sb.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.TrimSpace(sb.String())
}

// isValidURL is a lightweight check: must have scheme + host.
func isValidURL(rawURL string) bool {
	if rawURL == "" {
		return false
	}
	lower := strings.ToLower(rawURL)
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}
