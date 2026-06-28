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

// GroupService manages teams. All mutating operations are gated at the route
// level by the groups.manage instance permission.
type GroupService interface {
	List(ctx context.Context) ([]*model.Group, error)
	ListMine(ctx context.Context, userID string) ([]*model.Group, error)
	Create(ctx context.Context, actorID, name string) (*model.Group, error)
	Rename(ctx context.Context, id, name string) error
	Delete(ctx context.Context, id string) error
	ListMembers(ctx context.Context, groupID string) ([]*model.GroupMember, error)
	AddMember(ctx context.Context, groupID, userID, role string) error
	RemoveMember(ctx context.Context, groupID, userID string) error
}

type groupService struct {
	repos *repository.Repositories
}

func newGroupService(repos *repository.Repositories) GroupService {
	return &groupService{repos: repos}
}

func (s *groupService) List(ctx context.Context) ([]*model.Group, error) {
	return s.repos.Groups.ListAll(ctx)
}

func (s *groupService) ListMine(ctx context.Context, userID string) ([]*model.Group, error) {
	return s.repos.Groups.ListByMember(ctx, userID)
}

func (s *groupService) Create(ctx context.Context, actorID, name string) (*model.Group, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("%w: group name required", ErrInvalidInput)
	}
	now := time.Now()
	g := &model.Group{
		ID:        ulid.Make().String(),
		Name:      name,
		OwnerID:   actorID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repos.Groups.Create(ctx, g); err != nil {
		return nil, err
	}
	// The creator joins as a group admin.
	if err := s.repos.Groups.SetMember(ctx, g.ID, actorID, "admin"); err != nil {
		return nil, err
	}
	_ = s.repos.Audit.Log(ctx, &model.AuditEntry{
		ID: ulid.Make().String(), UserID: &actorID, Action: "group_created",
		Metadata: map[string]any{"group": g.ID, "name": name}, CreatedAt: now,
	})
	return g, nil
}

func (s *groupService) Rename(ctx context.Context, id, name string) error {
	g, err := s.repos.Groups.GetByID(ctx, id)
	if err != nil {
		return ErrNotFound
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("%w: group name required", ErrInvalidInput)
	}
	g.Name = name
	return s.repos.Groups.Update(ctx, g)
}

func (s *groupService) Delete(ctx context.Context, id string) error {
	if _, err := s.repos.Groups.GetByID(ctx, id); err != nil {
		return ErrNotFound
	}
	return s.repos.Groups.Delete(ctx, id)
}

func (s *groupService) ListMembers(ctx context.Context, groupID string) ([]*model.GroupMember, error) {
	if _, err := s.repos.Groups.GetByID(ctx, groupID); err != nil {
		return nil, ErrNotFound
	}
	return s.repos.Groups.ListMembers(ctx, groupID)
}

func (s *groupService) AddMember(ctx context.Context, groupID, userID, role string) error {
	if role != "admin" && role != "member" {
		role = "member"
	}
	if _, err := s.repos.Groups.GetByID(ctx, groupID); err != nil {
		return ErrNotFound
	}
	if _, err := s.repos.Users.GetByID(ctx, userID); err != nil {
		return fmt.Errorf("%w: unknown user", ErrInvalidInput)
	}
	return s.repos.Groups.SetMember(ctx, groupID, userID, role)
}

func (s *groupService) RemoveMember(ctx context.Context, groupID, userID string) error {
	return s.repos.Groups.RemoveMember(ctx, groupID, userID)
}
