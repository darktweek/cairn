package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/darktweek/cairn/internal/model"
)

type CollectionRepository interface {
	Create(ctx context.Context, c *model.Collection) error
	GetByID(ctx context.Context, id string) (*model.Collection, error)
	GetByPublicToken(ctx context.Context, token string) (*model.Collection, error)
	SetPublicToken(ctx context.Context, id string, token *string) error
	Update(ctx context.Context, c *model.Collection) error
	Delete(ctx context.Context, id string) error
	// ListAccessible returns every collection the user can access, each with its
	// effective Perm, owner username and bookmark count. Phase 1: owned only.
	ListAccessible(ctx context.Context, userID string) ([]*model.Collection, error)
	// GetOrCreatePersonal returns the user's personal collection, creating it lazily.
	GetOrCreatePersonal(ctx context.Context, userID string) (*model.Collection, error)
	// EffectivePerm returns "", view, edit or manage for (user, collection).
	EffectivePerm(ctx context.Context, userID, collectionID string) (string, error)
	// AccessibleCollectionIDs returns the IDs of every collection the user can read.
	AccessibleCollectionIDs(ctx context.Context, userID string) ([]string, error)

	// ListShares returns the per-user grants on a collection (excluding the owner).
	ListShares(ctx context.Context, collectionID string) ([]*model.CollectionShare, error)
	// SetShare upserts a user's permission on a collection.
	SetShare(ctx context.Context, collectionID, userID, perm string) error
	// RemoveShare revokes a user's access to a collection.
	RemoveShare(ctx context.Context, collectionID, userID string) error

	// Group shares (Phase 4).
	ListGroupShares(ctx context.Context, collectionID string) ([]*model.CollectionGroupShare, error)
	SetGroupShare(ctx context.Context, collectionID, groupID, perm string) error
	RemoveGroupShare(ctx context.Context, collectionID, groupID string) error
}

type sqliteCollectionRepo struct {
	db *sql.DB
}

func newSQLiteCollectionRepo(db *sql.DB) CollectionRepository {
	return &sqliteCollectionRepo{db: db}
}

func (r *sqliteCollectionRepo) Create(ctx context.Context, c *model.Collection) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO collections (id, owner_id, name, description, color, icon, is_personal, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.ID, c.OwnerID, c.Name, nullStr(c.Description), nullStr(c.Color), nullStr(c.Icon),
		boolInt(c.IsPersonal), c.CreatedAt.Unix(), c.UpdatedAt.Unix(),
	)
	if err != nil {
		return fmt.Errorf("collection create: %w", err)
	}
	return nil
}

func (r *sqliteCollectionRepo) GetByID(ctx context.Context, id string) (*model.Collection, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, owner_id, name, description, color, icon, is_personal, created_at, updated_at, public_token
		FROM collections WHERE id = ?`, id)
	return scanCollection(row)
}

func (r *sqliteCollectionRepo) GetByPublicToken(ctx context.Context, token string) (*model.Collection, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, owner_id, name, description, color, icon, is_personal, created_at, updated_at, public_token
		FROM collections WHERE public_token = ?`, token)
	return scanCollection(row)
}

func (r *sqliteCollectionRepo) SetPublicToken(ctx context.Context, id string, token *string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE collections SET public_token = ?, updated_at = ? WHERE id = ?`,
		token, time.Now().Unix(), id)
	return err
}

func (r *sqliteCollectionRepo) Update(ctx context.Context, c *model.Collection) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE collections SET name = ?, description = ?, color = ?, icon = ?, updated_at = ?
		WHERE id = ?`,
		c.Name, nullStr(c.Description), nullStr(c.Color), nullStr(c.Icon), time.Now().Unix(), c.ID,
	)
	if err != nil {
		return fmt.Errorf("collection update: %w", err)
	}
	return nil
}

func (r *sqliteCollectionRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM collections WHERE id = ?`, id)
	return err
}

func (r *sqliteCollectionRepo) ListAccessible(ctx context.Context, userID string) ([]*model.Collection, error) {
	// When the admin-override policy is on and the user can manage all collections,
	// every collection is listed; otherwise owned + directly shared collections.
	override, err := r.adminOverride(ctx, userID)
	if err != nil {
		return nil, err
	}

	var query string
	var args []any
	sharedExpr := `(c.owner_id <> ?
	     OR EXISTS(SELECT 1 FROM collection_shares s2 WHERE s2.collection_id = c.id)
	     OR EXISTS(SELECT 1 FROM collection_group_shares g2 WHERE g2.collection_id = c.id)
	     OR c.public_token IS NOT NULL)`

	if override {
		query = `
		SELECT c.id, c.owner_id, c.name, c.description, c.color, c.icon, c.is_personal,
		       c.created_at, c.updated_at, u.username,
		       (SELECT COUNT(*) FROM bookmarks b WHERE b.collection_id = c.id) AS cnt,
		       'manage' AS perm,
		       ` + sharedExpr + ` AS shared,
		       (c.public_token IS NOT NULL) AS is_public
		FROM collections c JOIN users u ON u.id = c.owner_id
		ORDER BY c.is_personal DESC, c.name COLLATE NOCASE ASC`
		args = []any{userID}
	} else {
		query = `
		SELECT c.id, c.owner_id, c.name, c.description, c.color, c.icon, c.is_personal,
		       c.created_at, c.updated_at, u.username,
		       (SELECT COUNT(*) FROM bookmarks b WHERE b.collection_id = c.id) AS cnt,
		       CASE WHEN c.owner_id = ? THEN 'manage'
		            ELSE COALESCE(cs.perm, gshare.perm) END AS perm,
		       ` + sharedExpr + ` AS shared,
		       (c.public_token IS NOT NULL) AS is_public
		FROM collections c
		JOIN users u ON u.id = c.owner_id
		LEFT JOIN collection_shares cs ON cs.collection_id = c.id AND cs.user_id = ?
		LEFT JOIN (
		    SELECT cgs.collection_id AS cid, MAX(
		        CASE cgs.perm WHEN 'manage' THEN 3 WHEN 'edit' THEN 2 ELSE 1 END
		    ) AS rank,
		    CASE MAX(CASE cgs.perm WHEN 'manage' THEN 3 WHEN 'edit' THEN 2 ELSE 1 END)
		        WHEN 3 THEN 'manage' WHEN 2 THEN 'edit' ELSE 'view' END AS perm
		    FROM collection_group_shares cgs
		    JOIN group_members gm ON gm.group_id = cgs.group_id
		    WHERE gm.user_id = ?
		    GROUP BY cgs.collection_id
		) gshare ON gshare.cid = c.id
		WHERE c.owner_id = ? OR cs.user_id IS NOT NULL OR gshare.cid IS NOT NULL
		ORDER BY (c.owner_id = ?) DESC, c.is_personal DESC, c.name COLLATE NOCASE ASC`
		args = []any{userID, userID, userID, userID, userID, userID}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("collection list: %w", err)
	}
	defer rows.Close()

	var out []*model.Collection
	for rows.Next() {
		c, err := scanCollectionListWithPerm(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *sqliteCollectionRepo) GetOrCreatePersonal(ctx context.Context, userID string) (*model.Collection, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, owner_id, name, description, color, icon, is_personal, created_at, updated_at, public_token
		FROM collections WHERE owner_id = ? AND is_personal = 1 LIMIT 1`, userID)
	c, err := scanCollection(row)
	if err == nil {
		return c, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	now := time.Now()
	c = &model.Collection{
		ID:         newID(),
		OwnerID:    userID,
		Name:       "Personal",
		IsPersonal: true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := r.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (r *sqliteCollectionRepo) EffectivePerm(ctx context.Context, userID, collectionID string) (string, error) {
	var ownerID string
	err := r.db.QueryRowContext(ctx,
		`SELECT owner_id FROM collections WHERE id = ?`, collectionID).Scan(&ownerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("collection effective perm: %w", err)
	}
	if ownerID == userID {
		return model.PermManage, nil
	}

	best := ""
	// Direct share.
	var perm sql.NullString
	if err := r.db.QueryRowContext(ctx,
		`SELECT perm FROM collection_shares WHERE collection_id = ? AND user_id = ?`,
		collectionID, userID).Scan(&perm); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("share perm: %w", err)
	}
	if perm.Valid && model.PermRank(perm.String) > model.PermRank(best) {
		best = perm.String
	}

	// Group shares (Phase 4) union in here.
	if gp, err := r.groupSharePerm(ctx, userID, collectionID); err != nil {
		return "", err
	} else if model.PermRank(gp) > model.PermRank(best) {
		best = gp
	}

	// Admin-override policy: manage on any collection.
	if best != model.PermManage {
		if ok, err := r.adminOverride(ctx, userID); err != nil {
			return "", err
		} else if ok {
			best = model.PermManage
		}
	}
	return best, nil
}

func (r *sqliteCollectionRepo) AccessibleCollectionIDs(ctx context.Context, userID string) ([]string, error) {
	cols, err := r.ListAccessible(ctx, userID)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(cols))
	for _, c := range cols {
		ids = append(ids, c.ID)
	}
	return ids, nil
}

func (r *sqliteCollectionRepo) ListShares(ctx context.Context, collectionID string) ([]*model.CollectionShare, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT cs.user_id, u.username, cs.perm
		FROM collection_shares cs JOIN users u ON u.id = cs.user_id
		WHERE cs.collection_id = ?
		ORDER BY u.username COLLATE NOCASE ASC`, collectionID)
	if err != nil {
		return nil, fmt.Errorf("list shares: %w", err)
	}
	defer rows.Close()
	var out []*model.CollectionShare
	for rows.Next() {
		var s model.CollectionShare
		if err := rows.Scan(&s.UserID, &s.Username, &s.Perm); err != nil {
			return nil, err
		}
		out = append(out, &s)
	}
	return out, rows.Err()
}

func (r *sqliteCollectionRepo) SetShare(ctx context.Context, collectionID, userID, perm string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO collection_shares (collection_id, user_id, perm, created_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(collection_id, user_id) DO UPDATE SET perm = excluded.perm`,
		collectionID, userID, perm, time.Now().Unix())
	if err != nil {
		return fmt.Errorf("set share: %w", err)
	}
	return nil
}

func (r *sqliteCollectionRepo) RemoveShare(ctx context.Context, collectionID, userID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM collection_shares WHERE collection_id = ? AND user_id = ?`, collectionID, userID)
	return err
}

// adminOverride reports whether the admin-manage-all policy is enabled AND the
// user holds the collections.manage_all permission.
func (r *sqliteCollectionRepo) adminOverride(ctx context.Context, userID string) (bool, error) {
	var val sql.NullString
	if err := r.db.QueryRowContext(ctx,
		`SELECT value FROM settings WHERE key = ?`, model.PolicyAdminManageAllCollections,
	).Scan(&val); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, fmt.Errorf("policy read: %w", err)
	}
	if val.String != "true" {
		return false, nil
	}
	var n int
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM role_permissions rp JOIN users u ON u.role_id = rp.role_id
		WHERE u.id = ? AND rp.permission = ?`, userID, model.PermCollectionsManageAll,
	).Scan(&n); err != nil {
		return false, fmt.Errorf("manage_all check: %w", err)
	}
	return n > 0, nil
}

// groupSharePerm returns the strongest permission granted to the user on the
// collection via any group they belong to.
func (r *sqliteCollectionRepo) groupSharePerm(ctx context.Context, userID, collectionID string) (string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT cgs.perm
		FROM collection_group_shares cgs
		JOIN group_members gm ON gm.group_id = cgs.group_id
		WHERE cgs.collection_id = ? AND gm.user_id = ?`, collectionID, userID)
	if err != nil {
		return "", fmt.Errorf("group share perm: %w", err)
	}
	defer rows.Close()
	best := ""
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return "", err
		}
		if model.PermRank(p) > model.PermRank(best) {
			best = p
		}
	}
	return best, rows.Err()
}

func (r *sqliteCollectionRepo) ListGroupShares(ctx context.Context, collectionID string) ([]*model.CollectionGroupShare, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT cgs.group_id, g.name, cgs.perm
		FROM collection_group_shares cgs JOIN groups g ON g.id = cgs.group_id
		WHERE cgs.collection_id = ?
		ORDER BY g.name COLLATE NOCASE ASC`, collectionID)
	if err != nil {
		return nil, fmt.Errorf("list group shares: %w", err)
	}
	defer rows.Close()
	var out []*model.CollectionGroupShare
	for rows.Next() {
		var s model.CollectionGroupShare
		if err := rows.Scan(&s.GroupID, &s.GroupName, &s.Perm); err != nil {
			return nil, err
		}
		out = append(out, &s)
	}
	return out, rows.Err()
}

func (r *sqliteCollectionRepo) SetGroupShare(ctx context.Context, collectionID, groupID, perm string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO collection_group_shares (collection_id, group_id, perm, created_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(collection_id, group_id) DO UPDATE SET perm = excluded.perm`,
		collectionID, groupID, perm, time.Now().Unix())
	if err != nil {
		return fmt.Errorf("set group share: %w", err)
	}
	return nil
}

func (r *sqliteCollectionRepo) RemoveGroupShare(ctx context.Context, collectionID, groupID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM collection_group_shares WHERE collection_id = ? AND group_id = ?`, collectionID, groupID)
	return err
}

func scanCollection(s scanner) (*model.Collection, error) {
	var c model.Collection
	var desc, color, icon, publicToken sql.NullString
	var isPersonal int
	var createdAt, updatedAt int64

	err := s.Scan(&c.ID, &c.OwnerID, &c.Name, &desc, &color, &icon, &isPersonal, &createdAt, &updatedAt, &publicToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("scan collection: %w", err)
	}
	c.Description = desc.String
	c.Color = color.String
	c.Icon = icon.String
	c.IsPersonal = isPersonal != 0
	c.PublicToken = publicToken.String
	c.CreatedAt = time.Unix(createdAt, 0)
	c.UpdatedAt = time.Unix(updatedAt, 0)
	return &c, nil
}

func scanCollectionListWithPerm(s scanner) (*model.Collection, error) {
	var c model.Collection
	var desc, color, icon, perm sql.NullString
	var isPersonal, shared, isPublic int
	var createdAt, updatedAt int64

	err := s.Scan(&c.ID, &c.OwnerID, &c.Name, &desc, &color, &icon, &isPersonal,
		&createdAt, &updatedAt, &c.OwnerUsername, &c.BookmarkCount, &perm, &shared, &isPublic)
	if err != nil {
		return nil, fmt.Errorf("scan collection list: %w", err)
	}
	c.Description = desc.String
	c.Color = color.String
	c.Icon = icon.String
	c.IsPersonal = isPersonal != 0
	c.Perm = perm.String
	c.Shared = shared != 0
	c.IsPublic = isPublic != 0
	c.CreatedAt = time.Unix(createdAt, 0)
	c.UpdatedAt = time.Unix(updatedAt, 0)
	return &c, nil
}

func scanCollectionList(s scanner) (*model.Collection, error) {
	var c model.Collection
	var desc, color, icon sql.NullString
	var isPersonal int
	var createdAt, updatedAt int64

	err := s.Scan(&c.ID, &c.OwnerID, &c.Name, &desc, &color, &icon, &isPersonal,
		&createdAt, &updatedAt, &c.OwnerUsername, &c.BookmarkCount)
	if err != nil {
		return nil, fmt.Errorf("scan collection list: %w", err)
	}
	c.Description = desc.String
	c.Color = color.String
	c.Icon = icon.String
	c.IsPersonal = isPersonal != 0
	c.CreatedAt = time.Unix(createdAt, 0)
	c.UpdatedAt = time.Unix(updatedAt, 0)
	return &c, nil
}
