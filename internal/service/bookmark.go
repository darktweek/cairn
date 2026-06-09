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

type BookmarkService interface {
	Create(ctx context.Context, userID, rawURL, title string, folder *string, tags []string) (*model.Bookmark, error)
	Update(ctx context.Context, userID, bookmarkID, rawURL, title string, folder *string, tags []string) error
	Delete(ctx context.Context, userID, bookmarkID string) error
	List(ctx context.Context, userID string, filter repository.BookmarkFilter) ([]*model.Bookmark, int, error)
	UpdateSort(ctx context.Context, userID string, ids []string) error
	ImportNetscape(ctx context.Context, userID string, data []byte) (imported, skipped int, err error)
	ExportNetscape(ctx context.Context, userID string) ([]byte, error)
	GenerateBookmarklet(ctx context.Context, userID, ip, userAgent string) (string, error)
	ListTags(ctx context.Context, userID string) ([]*model.Tag, error)
	DeleteTag(ctx context.Context, userID, tagID string) error
}

type bookmarkService struct {
	repos *repository.Repositories
	cfg   *config.Config
	auth  AuthService
}

func newBookmarkService(repos *repository.Repositories, cfg *config.Config, auth AuthService) BookmarkService {
	return &bookmarkService{repos: repos, cfg: cfg, auth: auth}
}

func (s *bookmarkService) Create(ctx context.Context, userID, rawURL, title string, folder *string, tags []string) (*model.Bookmark, error) {
	if !isValidURL(rawURL) {
		return nil, fmt.Errorf("%w: invalid URL", ErrInvalidInput)
	}

	now := time.Now()
	b := &model.Bookmark{
		ID:        ulid.Make().String(),
		UserID:    userID,
		URL:       rawURL,
		Title:     title,
		Folder:    folder,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repos.Bookmarks.Create(ctx, b); err != nil {
		return nil, err
	}

	if err := s.applyTags(ctx, b.ID, userID, tags); err != nil {
		return nil, err
	}

	return s.repos.Bookmarks.GetByID(ctx, b.ID, userID)
}

func (s *bookmarkService) Update(ctx context.Context, userID, bookmarkID, rawURL, title string, folder *string, tags []string) error {
	if !isValidURL(rawURL) {
		return fmt.Errorf("%w: invalid URL", ErrInvalidInput)
	}

	b, err := s.repos.Bookmarks.GetByID(ctx, bookmarkID, userID)
	if err != nil {
		return ErrNotFound
	}

	b.URL = rawURL
	b.Title = title
	b.Folder = folder
	b.UpdatedAt = time.Now()

	if err := s.repos.Bookmarks.Update(ctx, b); err != nil {
		return err
	}

	return s.applyTags(ctx, bookmarkID, userID, tags)
}

func (s *bookmarkService) Delete(ctx context.Context, userID, bookmarkID string) error {
	return s.repos.Bookmarks.Delete(ctx, bookmarkID, userID)
}

func (s *bookmarkService) List(ctx context.Context, userID string, filter repository.BookmarkFilter) ([]*model.Bookmark, int, error) {
	return s.repos.Bookmarks.ListByUser(ctx, userID, filter)
}

func (s *bookmarkService) UpdateSort(ctx context.Context, userID string, ids []string) error {
	return s.repos.Bookmarks.UpdateSort(ctx, userID, ids)
}

func (s *bookmarkService) ImportNetscape(ctx context.Context, userID string, data []byte) (int, int, error) {
	items, err := parseNetscape(data)
	if err != nil {
		return 0, 0, fmt.Errorf("%w: %s", ErrInvalidInput, err.Error())
	}

	now := time.Now()
	var toInsert []*model.Bookmark
	skipped := 0

	for _, item := range items {
		if !isValidURL(item.url) {
			skipped++
			continue
		}
		var folder *string
		if item.folder != "" {
			f := item.folder
			folder = &f
		}
		b := &model.Bookmark{
			ID:        ulid.Make().String(),
			UserID:    userID,
			URL:       item.url,
			Title:     item.title,
			Folder:    folder,
			CreatedAt: now,
			UpdatedAt: now,
		}
		toInsert = append(toInsert, b)
	}

	if len(toInsert) == 0 {
		return 0, skipped, nil
	}

	if err := s.repos.Bookmarks.BulkCreate(ctx, userID, toInsert); err != nil {
		return 0, 0, err
	}

	imported := len(toInsert)
	_ = s.repos.Audit.Log(ctx, &model.AuditEntry{
		ID:     ulid.Make().String(),
		UserID: &userID,
		Action: "bookmark_import",
		Metadata: map[string]any{
			"imported": imported,
			"skipped":  skipped,
		},
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

	type entry struct {
		URL     string
		Title   string
		Folder  string
		AddDate int64
	}

	var entries []entry
	for _, b := range bookmarks {
		folder := ""
		if b.Folder != nil {
			folder = *b.Folder
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
	url    string
	title  string
	folder string
}

// parseNetscape parses the Netscape bookmarks HTML format.
func parseNetscape(data []byte) ([]netscapeItem, error) {
	doc, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	var items []netscapeItem
	var walk func(n *html.Node, folder string)
	walk = func(n *html.Node, folder string) {
		if n.Type == html.ElementNode {
			switch strings.ToUpper(n.Data) {
			case "H3":
				// folder name = text content of H3
				folder = textContent(n)
			case "A":
				item := netscapeItem{folder: folder}
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
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c, folder)
		}
	}

	walk(doc, "")
	return items, nil
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
