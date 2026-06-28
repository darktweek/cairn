package service

import (
	"context"
	"errors"
	"testing"

	"github.com/darktweek/cairn/internal/model"
	"github.com/darktweek/cairn/internal/repository"
)

func TestGroupShareGrantsAccessToMembers(t *testing.T) {
	auth, repos := newTestAuth(t)
	bm := newBookmarkService(repos, nil, auth)
	cols := newCollectionService(repos, nil)
	groups := newGroupService(repos)
	ctx := context.Background()

	owner := makeUser(t, repos, "owner", "owner@example.com")
	member := makeUser(t, repos, "member", "member@example.com")
	outsider := makeUser(t, repos, "outsider", "outsider@example.com")

	col, err := cols.Create(ctx, owner.ID, CollectionInput{Name: "Team space"})
	if err != nil {
		t.Fatalf("create collection: %v", err)
	}
	if _, err := bm.Create(ctx, owner.ID, BookmarkInput{URL: "https://a.com", Title: "A", CollectionID: col.ID}); err != nil {
		t.Fatalf("seed bookmark: %v", err)
	}

	g, err := groups.Create(ctx, owner.ID, "Team")
	if err != nil {
		t.Fatalf("create group: %v", err)
	}
	if err := groups.AddMember(ctx, g.ID, member.ID, "member"); err != nil {
		t.Fatalf("add member: %v", err)
	}

	// Share the collection with the group as editor.
	if err := cols.SetGroupShare(ctx, owner.ID, col.ID, g.ID, model.PermEdit); err != nil {
		t.Fatalf("group share: %v", err)
	}

	// Member can now write; outsider still has no access.
	if _, err := bm.Create(ctx, member.ID, BookmarkInput{URL: "https://b.com", Title: "B", CollectionID: col.ID}); err != nil {
		t.Fatalf("member write via group: %v", err)
	}
	if _, _, err := bm.ListInCollection(ctx, outsider.ID, col.ID, repository.BookmarkFilter{Limit: 10}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("outsider: got %v, want ErrNotFound", err)
	}

	// The collection appears in the member's accessible list with edit perm.
	list, err := cols.List(ctx, member.ID)
	if err != nil {
		t.Fatalf("member list: %v", err)
	}
	found := false
	for _, c := range list {
		if c.ID == col.ID {
			found = true
			if c.Perm != model.PermEdit {
				t.Fatalf("group-shared perm = %q, want edit", c.Perm)
			}
		}
	}
	if !found {
		t.Fatal("group-shared collection missing from member's list")
	}

	// Removing the member revokes access.
	if err := groups.RemoveMember(ctx, g.ID, member.ID); err != nil {
		t.Fatalf("remove member: %v", err)
	}
	if _, _, err := bm.ListInCollection(ctx, member.ID, col.ID, repository.BookmarkFilter{Limit: 10}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("ex-member: got %v, want ErrNotFound", err)
	}
}
