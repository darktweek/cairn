package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/darktweek/cairn/internal/model"
	"github.com/darktweek/cairn/internal/repository"
	"github.com/oklog/ulid/v2"
)

type CollectionInput struct {
	Name        string
	Description string
	Color       string
	Icon        string
}

type FolderInput struct {
	Name     string
	ParentID *string
	Sort     int
}

type CollectionService interface {
	List(ctx context.Context, userID string) ([]*model.Collection, error)
	Get(ctx context.Context, userID, id string) (*model.Collection, error)
	Create(ctx context.Context, userID string, in CollectionInput) (*model.Collection, error)
	Update(ctx context.Context, userID, id string, in CollectionInput) error
	Delete(ctx context.Context, userID, id string) error

	ListFolders(ctx context.Context, userID, collectionID string) ([]*model.Folder, error)
	CreateFolder(ctx context.Context, userID, collectionID string, in FolderInput) (*model.Folder, error)
	UpdateFolder(ctx context.Context, userID, folderID string, in FolderInput) error
	DeleteFolder(ctx context.Context, userID, folderID string) error

	ListShares(ctx context.Context, userID, collectionID string) ([]*model.CollectionShare, error)
	SetShare(ctx context.Context, userID, collectionID, targetUserID, perm string) error
	RemoveShare(ctx context.Context, userID, collectionID, targetUserID string) error
	ListGroupShares(ctx context.Context, userID, collectionID string) ([]*model.CollectionGroupShare, error)
	SetGroupShare(ctx context.Context, userID, collectionID, groupID, perm string) error
	RemoveGroupShare(ctx context.Context, userID, collectionID, groupID string) error
	SearchUsers(ctx context.Context, q string) ([]*model.User, error)

	Policies(ctx context.Context) (map[string]bool, error)
	SetPolicy(ctx context.Context, actorID, key string, value bool) error
}

// policyKeys is the set of instance policies exposed to admins.
var policyKeys = []string{
	model.PolicyAdminManageAllCollections,
	model.PolicyRestrictCollectionCreate,
	model.PolicyRestrictCollectionDelete,
}

type collectionService struct {
	repos *repository.Repositories
	email EmailService
}

func newCollectionService(repos *repository.Repositories, email EmailService) CollectionService {
	return &collectionService{repos: repos, email: email}
}

func (s *collectionService) requirePerm(ctx context.Context, userID, collectionID, need string) (string, error) {
	perm, err := s.repos.Collections.EffectivePerm(ctx, userID, collectionID)
	if err != nil {
		return "", err
	}
	if perm == "" {
		return "", ErrNotFound
	}
	if !model.PermAtLeast(perm, need) {
		return "", ErrForbidden
	}
	return perm, nil
}

func (s *collectionService) List(ctx context.Context, userID string) ([]*model.Collection, error) {
	// Ensure the personal collection exists so the UI always has a home.
	if _, err := s.repos.Collections.GetOrCreatePersonal(ctx, userID); err != nil {
		return nil, err
	}
	return s.repos.Collections.ListAccessible(ctx, userID)
}

func (s *collectionService) Get(ctx context.Context, userID, id string) (*model.Collection, error) {
	perm, err := s.requirePerm(ctx, userID, id, model.PermView)
	if err != nil {
		return nil, err
	}
	c, err := s.repos.Collections.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	c.Perm = perm
	return c, nil
}

// policyOn reports whether an instance policy is enabled.
func (s *collectionService) policyOn(ctx context.Context, key string) bool {
	v, err := s.repos.Settings.Get(ctx, key)
	return err == nil && v == "true"
}

// userHasPerm reports whether a user's role grants an instance permission.
func (s *collectionService) userHasPerm(ctx context.Context, userID, perm string) bool {
	perms, err := s.repos.Roles.PermissionsForUser(ctx, userID)
	if err != nil {
		return false
	}
	for _, p := range perms {
		if p == perm {
			return true
		}
	}
	return false
}

func (s *collectionService) Create(ctx context.Context, userID string, in CollectionInput) (*model.Collection, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: name required", ErrInvalidInput)
	}
	if s.policyOn(ctx, model.PolicyRestrictCollectionCreate) && !s.userHasPerm(ctx, userID, model.PermCollectionsCreate) {
		return nil, fmt.Errorf("%w: collection creation is restricted", ErrForbidden)
	}
	now := time.Now()
	c := &model.Collection{
		ID:          ulid.Make().String(),
		OwnerID:     userID,
		Name:        name,
		Description: in.Description,
		Color:       in.Color,
		Icon:        in.Icon,
		IsPersonal:  false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repos.Collections.Create(ctx, c); err != nil {
		return nil, err
	}
	c.Perm = model.PermManage
	return c, nil
}

func (s *collectionService) Update(ctx context.Context, userID, id string, in CollectionInput) error {
	if _, err := s.requirePerm(ctx, userID, id, model.PermManage); err != nil {
		return err
	}
	c, err := s.repos.Collections.GetByID(ctx, id)
	if err != nil {
		return err
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return fmt.Errorf("%w: name required", ErrInvalidInput)
	}
	c.Name = name
	c.Description = in.Description
	c.Color = in.Color
	c.Icon = in.Icon
	return s.repos.Collections.Update(ctx, c)
}

func (s *collectionService) Delete(ctx context.Context, userID, id string) error {
	if _, err := s.requirePerm(ctx, userID, id, model.PermManage); err != nil {
		return err
	}
	c, err := s.repos.Collections.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if c.IsPersonal {
		return fmt.Errorf("%w: cannot delete the personal collection", ErrForbidden)
	}
	if s.policyOn(ctx, model.PolicyRestrictCollectionDelete) && !s.userHasPerm(ctx, userID, model.PermCollectionsDeleteAny) {
		return fmt.Errorf("%w: collection deletion is restricted", ErrForbidden)
	}
	return s.repos.Collections.Delete(ctx, id)
}

func (s *collectionService) ListFolders(ctx context.Context, userID, collectionID string) ([]*model.Folder, error) {
	if _, err := s.requirePerm(ctx, userID, collectionID, model.PermView); err != nil {
		return nil, err
	}
	return s.repos.Folders.ListByCollection(ctx, collectionID)
}

func (s *collectionService) CreateFolder(ctx context.Context, userID, collectionID string, in FolderInput) (*model.Folder, error) {
	if _, err := s.requirePerm(ctx, userID, collectionID, model.PermEdit); err != nil {
		return nil, err
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: name required", ErrInvalidInput)
	}
	parentID, err := s.validateParent(ctx, collectionID, in.ParentID)
	if err != nil {
		return nil, err
	}
	f := &model.Folder{
		ID:           ulid.Make().String(),
		CollectionID: collectionID,
		ParentID:     parentID,
		Name:         name,
		Sort:         in.Sort,
		CreatedAt:    time.Now(),
	}
	if err := s.repos.Folders.Create(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}

func (s *collectionService) UpdateFolder(ctx context.Context, userID, folderID string, in FolderInput) error {
	f, err := s.repos.Folders.GetByID(ctx, folderID)
	if err != nil {
		return ErrNotFound
	}
	if _, err := s.requirePerm(ctx, userID, f.CollectionID, model.PermEdit); err != nil {
		return err
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return fmt.Errorf("%w: name required", ErrInvalidInput)
	}
	parentID, err := s.validateParent(ctx, f.CollectionID, in.ParentID)
	if err != nil {
		return err
	}
	if parentID != nil && *parentID == folderID {
		return fmt.Errorf("%w: folder cannot be its own parent", ErrInvalidInput)
	}
	f.Name = name
	f.ParentID = parentID
	f.Sort = in.Sort
	return s.repos.Folders.Update(ctx, f)
}

func (s *collectionService) DeleteFolder(ctx context.Context, userID, folderID string) error {
	f, err := s.repos.Folders.GetByID(ctx, folderID)
	if err != nil {
		return ErrNotFound
	}
	if _, err := s.requirePerm(ctx, userID, f.CollectionID, model.PermEdit); err != nil {
		return err
	}
	return s.repos.Folders.Delete(ctx, folderID)
}

func (s *collectionService) ListShares(ctx context.Context, userID, collectionID string) ([]*model.CollectionShare, error) {
	if _, err := s.requirePerm(ctx, userID, collectionID, model.PermManage); err != nil {
		return nil, err
	}
	return s.repos.Collections.ListShares(ctx, collectionID)
}

func (s *collectionService) SetShare(ctx context.Context, userID, collectionID, targetUserID, perm string) error {
	if _, err := s.requirePerm(ctx, userID, collectionID, model.PermManage); err != nil {
		return err
	}
	if perm != model.PermView && perm != model.PermEdit && perm != model.PermManage {
		return fmt.Errorf("%w: invalid permission", ErrInvalidInput)
	}
	col, err := s.repos.Collections.GetByID(ctx, collectionID)
	if err != nil {
		return err
	}
	if col.IsPersonal {
		return fmt.Errorf("%w: the personal collection cannot be shared", ErrForbidden)
	}
	if targetUserID == col.OwnerID {
		return fmt.Errorf("%w: the owner already has full access", ErrInvalidInput)
	}
	target, err := s.repos.Users.GetByID(ctx, targetUserID)
	if err != nil {
		return fmt.Errorf("%w: unknown user", ErrInvalidInput)
	}
	if err := s.repos.Collections.SetShare(ctx, collectionID, target.ID, perm); err != nil {
		return err
	}
	_ = s.repos.Audit.Log(ctx, &model.AuditEntry{
		ID:        ulid.Make().String(),
		UserID:    &userID,
		Action:    "collection_shared",
		Metadata:  map[string]any{"collection": collectionID, "target": target.ID, "perm": perm},
		CreatedAt: time.Now(),
	})
	// Best-effort notification to the new collaborator.
	if s.email != nil && target.Email != "" {
		sharer, _ := s.repos.Users.GetByID(ctx, userID)
		sharerName := userID
		if sharer != nil {
			sharerName = sharer.Username
		}
		_ = s.email.SendCollectionShared(ctx, target.Email, sharerName, col.Name)
	}
	return nil
}

func (s *collectionService) RemoveShare(ctx context.Context, userID, collectionID, targetUserID string) error {
	if _, err := s.requirePerm(ctx, userID, collectionID, model.PermManage); err != nil {
		return err
	}
	if err := s.repos.Collections.RemoveShare(ctx, collectionID, targetUserID); err != nil {
		return err
	}
	_ = s.repos.Audit.Log(ctx, &model.AuditEntry{
		ID:        ulid.Make().String(),
		UserID:    &userID,
		Action:    "collection_unshared",
		Metadata:  map[string]any{"collection": collectionID, "target": targetUserID},
		CreatedAt: time.Now(),
	})
	return nil
}

func (s *collectionService) ListGroupShares(ctx context.Context, userID, collectionID string) ([]*model.CollectionGroupShare, error) {
	if _, err := s.requirePerm(ctx, userID, collectionID, model.PermManage); err != nil {
		return nil, err
	}
	return s.repos.Collections.ListGroupShares(ctx, collectionID)
}

func (s *collectionService) SetGroupShare(ctx context.Context, userID, collectionID, groupID, perm string) error {
	if _, err := s.requirePerm(ctx, userID, collectionID, model.PermManage); err != nil {
		return err
	}
	if perm != model.PermView && perm != model.PermEdit && perm != model.PermManage {
		return fmt.Errorf("%w: invalid permission", ErrInvalidInput)
	}
	col, err := s.repos.Collections.GetByID(ctx, collectionID)
	if err != nil {
		return err
	}
	if col.IsPersonal {
		return fmt.Errorf("%w: the personal collection cannot be shared", ErrForbidden)
	}
	if _, err := s.repos.Groups.GetByID(ctx, groupID); err != nil {
		return fmt.Errorf("%w: unknown group", ErrInvalidInput)
	}
	if err := s.repos.Collections.SetGroupShare(ctx, collectionID, groupID, perm); err != nil {
		return err
	}
	_ = s.repos.Audit.Log(ctx, &model.AuditEntry{
		ID: ulid.Make().String(), UserID: &userID, Action: "collection_shared",
		Metadata: map[string]any{"collection": collectionID, "group": groupID, "perm": perm}, CreatedAt: time.Now(),
	})
	return nil
}

func (s *collectionService) RemoveGroupShare(ctx context.Context, userID, collectionID, groupID string) error {
	if _, err := s.requirePerm(ctx, userID, collectionID, model.PermManage); err != nil {
		return err
	}
	return s.repos.Collections.RemoveGroupShare(ctx, collectionID, groupID)
}

func (s *collectionService) SearchUsers(ctx context.Context, q string) ([]*model.User, error) {
	q = strings.TrimSpace(q)
	if len(q) < 1 {
		return nil, nil
	}
	return s.repos.Users.Search(ctx, q, 10)
}

func (s *collectionService) Policies(ctx context.Context) (map[string]bool, error) {
	out := map[string]bool{}
	for _, k := range policyKeys {
		v, err := s.repos.Settings.Get(ctx, k)
		if err != nil { // missing key = policy off (default)
			out[k] = false
			continue
		}
		out[k] = v == "true"
	}
	return out, nil
}

func (s *collectionService) SetPolicy(ctx context.Context, actorID, key string, value bool) error {
	valid := false
	for _, k := range policyKeys {
		if k == key {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("%w: unknown policy", ErrInvalidInput)
	}
	str := "false"
	if value {
		str = "true"
	}
	if err := s.repos.Settings.Set(ctx, key, str); err != nil {
		return err
	}
	_ = s.repos.Audit.Log(ctx, &model.AuditEntry{
		ID:        ulid.Make().String(),
		UserID:    &actorID,
		Action:    "policy_changed",
		Metadata:  map[string]any{"policy": key, "value": value},
		CreatedAt: time.Now(),
	})
	return nil
}

// validateParent ensures a parent folder, if given, lives in the same collection.
func (s *collectionService) validateParent(ctx context.Context, collectionID string, parentID *string) (*string, error) {
	if parentID == nil || *parentID == "" {
		return nil, nil
	}
	parent, err := s.repos.Folders.GetByID(ctx, *parentID)
	if err != nil || parent.CollectionID != collectionID {
		return nil, fmt.Errorf("%w: parent folder not in collection", ErrInvalidInput)
	}
	return parentID, nil
}
