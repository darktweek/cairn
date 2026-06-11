# Cairn — Spec technique complète

*État au 09 juin 2026 — voir « Écarts connus » en fin de document pour les évolutions postérieures (10 juin 2026).*

---

## 0. Produit — Ce qu'est Cairn

### Origine

Cairn est l'évolution de **go-startpage** — une start page minimaliste single-file (HTML/CSS/JS vanilla) publiée en snapshot public sur `github.com/darktweek/go-startpage`. Le projet a outgrown le single-file et est devenu une webapp multi-user full-stack, tout en conservant l'esthétique zen d'origine.

### Philosophie

- **Minimaliste et zen** — pas de bruit visuel, pas de favicons, pas de couleurs criardes
- **Self-hosted** — conçu pour tourner sur un homelab, pas pour le cloud public
- **Craftsmanship** — bien fait pour le plaisir de bien faire, pas pour scaler
- **Friction zéro** — chaque interaction doit être naturelle et rapide

---

### UX — Ce que voit l'utilisateur

#### Page principale

```
┌─────────────────────────────────────────────────────┐
│                                                     │
│         [background : image ou vidéo zen]           │
│              [animation pluie canvas]               │
│                                                     │
│                   12:34                             │
│              Lundi 09 juin 2026 · S23               │
│                                                     │
│              [ barre de recherche ]                 │
│                                                     │
└─────────────────────────────────────────────────────┘
```

Éléments présents :
- Horloge large, italic, serif (Cormorant Garant)
- Date + numéro de semaine ISO
- Barre de recherche DuckDuckGo avec autocomplete
- Support des bangs (`!g`, `!yt`, `!gh`, etc.)
- Animation pluie HTML5 canvas en fond
- Background : image ou vidéo, sélectionné aléatoirement parmi les wallpapers de l'user
- Thème jour/nuit adaptatif (luminance sampling sur le background)
- Fallback : gradient CSS animé si aucun wallpaper

#### Backgrounds

- Chaque user uploade ses propres wallpapers (images + vidéos)
- Limite configurable globalement et par user (défaut : 10)
- Un wallpaper peut être **pinné** → affiché en priorité
- Sans wallpaper pinné → sélection aléatoire à chaque chargement
- Formats acceptés : `.jpg .jpeg .png .webp .avif .mp4 .webm`
- Fichiers servis via auth → jamais publics

#### Gestion des bookmarks

La search bar **morphe** en interface de gestion complète via le bang `!bm` (ou `!edit`) :

```
!bm  →  la search bar se transforme en panel full-page

┌─────────────────────────────────────────────────────┐
│  [+ Ajouter]  [Importer]  [Exporter]   [✕ Fermer]  │
│                                                     │
│  Dev/Go                                             │
│    ├── https://go.dev — The Go Programming Language │
│    └── https://pkg.go.dev — Go Packages             │
│                                                     │
│  Perso                                              │
│    └── https://example.com — Mon site               │
│                                                     │
│  #homelab  #selfhosted                              │
│    └── https://noted.com — Noted                    │
└─────────────────────────────────────────────────────┘
```

- Organisation : dossiers (chemin materialisé) + tags libres
- Dossiers préservés à l'import/export (format Netscape `bookmarks.html`)
- Drag & drop pour réordonner (champ `sort` en DB)
- Pas de favicons
- Import : `bookmarks.html` natif Chrome/Firefox/Safari/Edge
- Export : même format Netscape → réimportable partout
- **Bookmarklet** : code JS glissé dans la barre de favoris → sauvegarde la page active en un clic, marche sur desktop et mobile Safari

#### Moteur de recherche

Configurable par user depuis le panel. DuckDuckGo par défaut.

| Moteur | URL |
|---|---|
| DuckDuckGo (défaut) | `https://duckduckgo.com/?q=` |
| Google | `https://www.google.com/search?q=` |
| Brave | `https://search.brave.com/search?q=` |
| Bing | `https://www.bing.com/search?q=` |
| Kagi | `https://kagi.com/search?q=` |
| Custom | URL définie par l'user (doit se terminer par `=`) |

- Stocké en DB par user (colonne `search_engine` sur `users`)
- Les bangs DDG (`!g`, `!yt`, etc.) fonctionnent **quel que soit le moteur choisi** — ils passent toujours via DDG
- Moteur custom : l'user saisit l'URL complète ex: `https://search.example.com/?q=`

#### Bangs custom

| Bang | Action |
|---|---|
| `!bm` ou `!edit` | Ouvre le panel bookmark |
| `!g` | Google |
| `!yt` | YouTube |
| `!gh` | GitHub |
| `!hub` | Docker Hub |
| + tous les bangs DuckDuckGo standard | |

#### Authentification

- Page de login : intégrée à l'esthétique zen (pas une page séparée générique)
- Pas d'inscription publique — l'admin crée les comptes
- **Premier user à se connecter = admin automatiquement**
- TOTP optionnel (QR code + app authenticator)
- Sessions visibles et révocables par l'user

#### Panel user (accessible depuis la page principale)

```
Mon compte
  ├── Modifier username / email
  ├── Changer mot de passe
  ├── Moteur de recherche : choisir parmi liste ou custom
  ├── TOTP : activer / désactiver
  ├── Sessions actives : voir / révoquer
  ├── Mon bookmarklet : copier / révoquer
  ├── Mes wallpapers : gérer (upload, pin, supprimer, réordonner)
  └── Historique de connexions
```

#### Panel admin

```
Administration
  ├── Stats : users, bookmarks, wallpapers, taille DB
  ├── Users : liste, suspendre, activer, supprimer
  ├── Quota wallpapers : global et par user
  └── Audit log : toutes les actions de tous les users
```

---

### Flow premier démarrage

1. L'app démarre, DB vide
2. Premier user qui s'inscrit → role `admin` automatique
3. Il arrive sur la page zen avec gradient CSS (pas de wallpaper encore)
4. Il configure son compte, uploade ses wallpapers, crée ses bookmarks
5. Il peut inviter d'autres users via le panel admin

---

## 0. Contexte

**Cairn** — webapp self-hosted, multi-user, start page + dashboard personnel.
Servi à `go.dotnot.be` sur homelab Ginyu, derrière Traefik + Cloudflare Tunnel.
Auth via Authentik (`auth.dotnot.be`) en SSO/OIDC optionnel.
Pas d'inscription publique — gestion users via l'app uniquement.
Motivation : apprendre et construire proprement, pas pour scaler.

---

## 1. Stack technique

| Composant | Choix | Raison |
|---|---|---|
| Langage | Go 1.22+ | Compile rapide, binaire statique, maintenable |
| Framework HTTP | `chi v5` | 100% compatible net/http, léger, pas de lock-in |
| Base de données | SQLite en WAL mode | Fichier unique, backup trivial, zéro serveur |
| Driver SQLite | `modernc.org/sqlite` | Pure Go, zéro CGO, compatible scratch |
| Query layer | `sqlc` | SQL pur → code Go typé généré |
| Migrations | `goose` | SQLite + Postgres supportés, embedded |
| IDs | ULID | Triable chronologiquement, pas d'int exposé |
| Password | Argon2id (RFC 9106) | Standard actuel |
| Sessions | HMAC-SHA256 | Cookie HttpOnly, token jamais stocké brut |
| TOTP | RFC 6238 | Compatible toutes apps authenticator |
| Image Docker | `scratch` | Zéro OS, zéro shell, surface d'attaque nulle |

---

## 2. Architecture générale

```
Browser
  │
  ├── GET /        → index.html (servi par Go)
  └── /api/*       → handlers chi
                        │
                        └── Repository → SQLite

/media/*           → middleware auth → fichier physique
```

Un seul binaire, un seul container, un seul port (8080).

### Pattern des couches

```
Request
  → chi Router
    → Middleware global (logger, headers, recoverer)
      → Middleware de route (auth, admin, rate limit, body limit)
        → Validation (struct tags, go-playground/validator)
          → Handler (parse → appelle service → retourne JSON)
            → Service (logique métier)
              → Repository (interface → SQLite)
                → SQLite WAL
```

### Scalabilité future

L'architecture est compatible Kubernetes sans modification du code :
- Toute la config via env vars
- Stateless sauf `/data` (un seul PVC à gérer)
- Image multi-arch (amd64 + arm64)
- Healthcheck HTTP sur `GET /healthcheck`
- Séparation config (non sensible) / secrets (`.env` gitignored)

---

## 3. Structure des dossiers

```
cairn/
├── cmd/
│   └── cairn/
│       └── main.go
│
├── internal/
│   ├── config/
│   │   └── config.go
│   │
│   ├── db/
│   │   ├── db.go
│   │   ├── migrate.go
│   │   └── migrations/
│   │       ├── 001_init_users.sql
│   │       ├── 002_init_sessions.sql
│   │       ├── 003_init_totp.sql
│   │       ├── 004_init_bookmarks.sql
│   │       ├── 005_init_tags.sql
│   │       ├── 006_init_wallpapers.sql
│   │       └── 007_init_audit_log.sql
│   │
│   ├── model/
│   │   ├── user.go
│   │   ├── bookmark.go
│   │   ├── wallpaper.go
│   │   └── audit.go
│   │
│   ├── repository/
│   │   ├── repository.go
│   │   ├── user.go
│   │   ├── session.go
│   │   ├── bookmark.go
│   │   ├── tag.go
│   │   ├── wallpaper.go
│   │   ├── totp.go
│   │   └── audit.go
│   │
│   ├── service/
│   │   ├── service.go
│   │   ├── auth.go
│   │   ├── user.go
│   │   ├── bookmark.go
│   │   ├── wallpaper.go
│   │   ├── admin.go
│   │   ├── email.go
│   │   └── templates/
│   │       ├── password_reset.html
│   │       └── welcome.html
│   │
│   ├── handler/
│   │   ├── handler.go
│   │   ├── auth.go
│   │   ├── user.go
│   │   ├── bookmark.go
│   │   ├── tag.go
│   │   ├── wallpaper.go
│   │   └── admin.go
│   │
│   └── middleware/
│       ├── middleware.go
│       ├── headers.go
│       ├── auth.go
│       ├── admin.go
│       ├── bookmarklet.go
│       ├── ratelimit.go
│       ├── cors.go
│       └── bodylimit.go
│
├── web/
│   └── static/
│       └── index.html
│
├── data/                ← volume Docker (gitignored)
│   ├── db.sqlite
│   └── media/
│       └── {user-id}/
│
├── go.mod
├── go.sum
├── Dockerfile
├── compose.yaml
├── .env                 ← gitignored, secrets uniquement
└── .gitignore
```

---

## 4. Config — `internal/config/config.go`

### Struct

```go
type Config struct {
    // Serveur
    Addr    string
    Env     string  // "production" | "development"
    BaseURL string

    // Base de données
    DBPath    string
    MediaPath string

    // Sécurité
    SessionSecret string  // min 32 chars

    // Limites
    DefaultWallpaperLimit    int
    MaxUploadSize            int64
    BookmarkletTokenLifetime int  // jours

    // TOTP
    TOTPIssuer string

    // Proxy
    TrustedProxy bool

    // SMTP
    SMTPHost string
    SMTPPort int
    SMTPUser string
    SMTPPass string
    SMTPFrom string
    SMTPTLS  bool
}
```

### Variables d'environnement

| Variable | Défaut | Obligatoire | Description |
|---|---|---|---|
| `CAIRN_ADDR` | `:8080` | non | Adresse d'écoute |
| `CAIRN_ENV` | `production` | non | `production` ou `development` |
| `CAIRN_BASE_URL` | — | **oui** | URL publique ex: `https://go.dotnot.be` |
| `CAIRN_DB_PATH` | `/data/db.sqlite` | non | Chemin SQLite |
| `CAIRN_MEDIA_PATH` | `/data/media` | non | Racine media |
| `CAIRN_SESSION_SECRET` | — | **oui** | Clé HMAC, min 32 chars |
| `CAIRN_DEFAULT_WALLPAPER_LIMIT` | `10` | non | Limite globale wallpapers |
| `CAIRN_MAX_UPLOAD_SIZE` | `52428800` | non | 50MB en bytes |
| `CAIRN_BOOKMARKLET_TOKEN_LIFETIME` | `90` | non | Durée token bookmarklet en jours |
| `CAIRN_TOTP_ISSUER` | `Cairn` | non | Nom affiché dans l'app authenticator |
| `CAIRN_TRUSTED_PROXY` | `true` | non | Lire X-Forwarded-For / CF-Connecting-IP |
| `CAIRN_SMTP_HOST` | — | **oui** | Serveur SMTP |
| `CAIRN_SMTP_PORT` | `587` | non | Port SMTP |
| `CAIRN_SMTP_USER` | — | **oui** | Utilisateur SMTP |
| `CAIRN_SMTP_PASS` | — | **oui** | Mot de passe SMTP |
| `CAIRN_SMTP_FROM` | — | **oui** | Adresse expéditeur |
| `CAIRN_SMTP_TLS` | `true` | non | STARTTLS activé |

### Règles de validation

- Champs obligatoires vides → erreur explicite au démarrage
- `CAIRN_SESSION_SECRET` < 32 chars → erreur
- `CAIRN_ENV` != `production|development` → erreur
- Pas de dépendance externe (stdlib uniquement)

---

## 5. Base de données — `internal/db/`

### `db.go` — Fonction `Open(path string) (*sql.DB, error)`

PRAGMAs à configurer à l'ouverture :

```sql
PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;
PRAGMA busy_timeout = 5000;
PRAGMA synchronous = NORMAL;
PRAGMA cache_size = -64000;
PRAGMA temp_store = MEMORY;
```

Pool de connexions :
```
SetMaxOpenConns(1)      ← SQLite = 1 writer max
SetMaxIdleConns(1)
SetConnMaxLifetime(0)   ← connexion persistante
```

### `migrate.go` — Fonction `Migrate(db *sql.DB) error`

- Goose en mode embedded (`go:embed migrations`)
- `goose.SetDialect("sqlite3")`
- `goose.Up()` au démarrage → applique toutes les migrations pending
- Zéro fichiers externes au runtime

### Migrations

#### `001_init_users.sql`

```sql
-- +goose Up
CREATE TABLE users (
    id                TEXT    NOT NULL PRIMARY KEY,
    username          TEXT    NOT NULL UNIQUE,
    email             TEXT    NOT NULL UNIQUE COLLATE NOCASE,
    password          TEXT    NOT NULL,
    role              TEXT    NOT NULL DEFAULT 'user'
                              CHECK(role IN ('user', 'admin')),
    is_active         INTEGER NOT NULL DEFAULT 1
                              CHECK(is_active IN (0, 1)),
    wallpaper_limit   INTEGER,
    search_engine     TEXT    NOT NULL DEFAULT 'duckduckgo'
                              CHECK(search_engine IN (
                                  'duckduckgo', 'google', 'brave',
                                  'bing', 'kagi', 'custom'
                              )),
    search_engine_url TEXT,   -- rempli uniquement si search_engine = 'custom'
    created_at        INTEGER NOT NULL,
    updated_at        INTEGER NOT NULL,
    deleted_at        INTEGER
);

CREATE INDEX idx_users_email    ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_active   ON users(is_active) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE users;
```

#### `002_init_sessions.sql`

```sql
-- +goose Up
CREATE TABLE sessions (
    id              TEXT    NOT NULL PRIMARY KEY,
    user_id         TEXT    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash      TEXT    NOT NULL UNIQUE,
    user_agent      TEXT,
    ip              TEXT,
    expires_at      INTEGER NOT NULL,
    created_at      INTEGER NOT NULL,
    is_bookmarklet  INTEGER NOT NULL DEFAULT 0
                            CHECK(is_bookmarklet IN (0, 1))
);

CREATE INDEX idx_sessions_user_id    ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- +goose Down
DROP TABLE sessions;
```

#### `003_init_totp.sql`

```sql
-- +goose Up
CREATE TABLE totp_secrets (
    user_id     TEXT    NOT NULL PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    secret      TEXT    NOT NULL,
    is_verified INTEGER NOT NULL DEFAULT 0
                        CHECK(is_verified IN (0, 1)),
    created_at  INTEGER NOT NULL
);

-- +goose Down
DROP TABLE totp_secrets;
```

#### `004_init_bookmarks.sql`

```sql
-- +goose Up
CREATE TABLE bookmarks (
    id          TEXT    NOT NULL PRIMARY KEY,
    user_id     TEXT    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    url         TEXT    NOT NULL,
    title       TEXT    NOT NULL,
    folder      TEXT,
    sort        INTEGER NOT NULL DEFAULT 0,
    created_at  INTEGER NOT NULL,
    updated_at  INTEGER NOT NULL
);

CREATE INDEX idx_bookmarks_user_id ON bookmarks(user_id);
CREATE INDEX idx_bookmarks_folder  ON bookmarks(user_id, folder);
CREATE INDEX idx_bookmarks_sort    ON bookmarks(user_id, sort);

-- +goose Down
DROP TABLE bookmarks;
```

#### `005_init_tags.sql`

```sql
-- +goose Up
CREATE TABLE tags (
    id      TEXT NOT NULL PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name    TEXT NOT NULL COLLATE NOCASE,
    UNIQUE(user_id, name)
);

CREATE INDEX idx_tags_user_id ON tags(user_id);

CREATE TABLE bookmark_tags (
    bookmark_id TEXT NOT NULL REFERENCES bookmarks(id) ON DELETE CASCADE,
    tag_id      TEXT NOT NULL REFERENCES tags(id)      ON DELETE CASCADE,
    PRIMARY KEY (bookmark_id, tag_id)
);

-- +goose Down
DROP TABLE bookmark_tags;
DROP TABLE tags;
```

#### `006_init_wallpapers.sql`

```sql
-- +goose Up
CREATE TABLE wallpapers (
    id          TEXT    NOT NULL PRIMARY KEY,
    user_id     TEXT    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    filename    TEXT    NOT NULL,
    is_pinned   INTEGER NOT NULL DEFAULT 0
                        CHECK(is_pinned IN (0, 1)),
    sort        INTEGER NOT NULL DEFAULT 0,
    created_at  INTEGER NOT NULL,
    UNIQUE(user_id, filename)
);

CREATE INDEX idx_wallpapers_user_id ON wallpapers(user_id);

-- +goose Down
DROP TABLE wallpapers;
```

#### `007_init_audit_log.sql`

```sql
-- +goose Up
CREATE TABLE audit_log (
    id         TEXT NOT NULL PRIMARY KEY,
    user_id    TEXT REFERENCES users(id) ON DELETE SET NULL,
    action     TEXT NOT NULL
                    CHECK(action IN (
                        'login', 'logout', 'login_failed',
                        'password_change', 'totp_enabled', 'totp_disabled',
                        'user_created', 'user_deleted', 'user_suspended',
                        'bookmark_import', 'wallpaper_upload', 'wallpaper_delete'
                    )),
    ip         TEXT,
    user_agent TEXT,
    metadata   TEXT,
    created_at INTEGER NOT NULL
);

CREATE INDEX idx_audit_log_user_id    ON audit_log(user_id);
CREATE INDEX idx_audit_log_created_at ON audit_log(created_at);
CREATE INDEX idx_audit_log_action     ON audit_log(action);

-- +goose Down
DROP TABLE audit_log;
```

---

## 6. Modèles — `internal/model/`

```go
// user.go
type User struct {
    ID             string
    Username       string
    Email          string
    Password       string     // hash argon2id, jamais exposé JSON
    Role           string     // "user" | "admin"
    IsActive       bool
    WallpaperLimit  *int       // nil = valeur globale config
    SearchEngine    string     // "duckduckgo" | "google" | "brave" | "bing" | "kagi" | "custom"
    SearchEngineURL *string    // rempli uniquement si SearchEngine = "custom"
    CreatedAt       time.Time
    UpdatedAt       time.Time
    DeletedAt       *time.Time
}

type Session struct {
    ID            string
    UserID        string
    TokenHash     string
    UserAgent     string
    IP            string
    ExpiresAt     time.Time
    CreatedAt     time.Time
    IsBookmarklet bool
}

type TOTPSecret struct {
    UserID     string
    Secret     string  // chiffré AES-GCM
    IsVerified bool
    CreatedAt  time.Time
}

// bookmark.go
type Bookmark struct {
    ID        string
    UserID    string
    URL       string
    Title     string
    Folder    *string   // nil = pas de dossier, "Dev/Go" = chemin materialisé
    Sort      int
    Tags      []Tag
    CreatedAt time.Time
    UpdatedAt time.Time
}

type Tag struct {
    ID     string
    UserID string
    Name   string
}

// wallpaper.go
type Wallpaper struct {
    ID        string
    UserID    string
    Filename  string
    IsPinned  bool
    Sort      int
    CreatedAt time.Time
}

// audit.go
type AuditEntry struct {
    ID        string
    UserID    *string
    Action    string
    IP        string
    UserAgent string
    Metadata  map[string]any  // sérialisé JSON en DB
    CreatedAt time.Time
}

type AdminStats struct {
    TotalUsers      int
    ActiveUsers     int
    TotalBookmarks  int
    TotalWallpapers int
    DBSizeBytes     int64
}
```

---

## 7. Repository — `internal/repository/`

### Pattern général

```go
// Interface définie dans chaque fichier
// Struct sqliteXxxRepo qui implémente l'interface
// Constructeur newSQLiteXxxRepo(db *sql.DB)
// Toutes les queries filtrent WHERE deleted_at IS NULL sauf admin explicite
```

### `repository.go`

```go
type Repositories struct {
    Users      UserRepository
    Sessions   SessionRepository
    Bookmarks  BookmarkRepository
    Tags       TagRepository
    Wallpapers WallpaperRepository
    TOTP       TOTPRepository
    Audit      AuditRepository
}

func New(db *sql.DB) *Repositories
```

### `user.go` — Interface `UserRepository`

```
Create(ctx, user *model.User) error
GetByID(ctx, id string) (*model.User, error)
GetByEmail(ctx, email string) (*model.User, error)
GetByUsername(ctx, username string) (*model.User, error)
Update(ctx, user *model.User) error
SoftDelete(ctx, id string) error
List(ctx, offset, limit int) ([]*model.User, int, error)
Count(ctx) (int, error)
IsFirstUser(ctx) (bool, error)
```

Notes :
- `GetByEmail` et `GetByUsername` case-insensitive (COLLATE NOCASE en DB)
- `List` retourne aussi le total pour pagination
- `IsFirstUser` → `SELECT COUNT(*) = 0`

### `session.go` — Interface `SessionRepository`

```
Create(ctx, session *model.Session) error
GetByTokenHash(ctx, hash string) (*model.Session, error)
DeleteByID(ctx, id string) error
DeleteByUserID(ctx, userID string) error
DeleteExpired(ctx) error
ListByUserID(ctx, userID string) ([]*model.Session, error)
```

Notes :
- `GetByTokenHash` vérifie aussi `expires_at > now()`
- `DeleteExpired` appelé au démarrage + toutes les heures via goroutine

### `bookmark.go` — Interface `BookmarkRepository`

```
Create(ctx, bookmark *model.Bookmark) error
GetByID(ctx, id, userID string) (*model.Bookmark, error)
Update(ctx, bookmark *model.Bookmark) error
Delete(ctx, id, userID string) error
ListByUser(ctx, userID string, filter BookmarkFilter) ([]*model.Bookmark, int, error)
UpdateSort(ctx, userID string, ids []string) error
CountByUser(ctx, userID string) (int, error)
AddTag(ctx, bookmarkID, tagID string) error
RemoveTag(ctx, bookmarkID, tagID string) error
SetTags(ctx, bookmarkID string, tagIDs []string) error
BulkCreate(ctx, userID string, bookmarks []*model.Bookmark) error

type BookmarkFilter struct {
    Folder *string
    TagID  *string
    Search string
    Offset int
    Limit  int
}
```

Notes :
- `GetByID` et `Delete` vérifient toujours `userID` → pas d'accès cross-user
- `Search` utilise `LIKE '%term%'` (SQLite FTS5 optionnel plus tard)
- `BulkCreate` via transaction unique pour l'import

### `tag.go` — Interface `TagRepository`

```
Create(ctx, tag *model.Tag) (*model.Tag, error)
GetOrCreate(ctx, userID, name string) (*model.Tag, error)
ListByUser(ctx, userID string) ([]*model.Tag, error)
Delete(ctx, id, userID string) error
```

Notes :
- `GetOrCreate` → `INSERT OR IGNORE` + `SELECT`
- Suppression d'un tag → `bookmark_tags` nettoyé via CASCADE

### `wallpaper.go` — Interface `WallpaperRepository`

```
Create(ctx, wallpaper *model.Wallpaper) error
GetByID(ctx, id, userID string) (*model.Wallpaper, error)
ListByUser(ctx, userID string) ([]*model.Wallpaper, error)
Delete(ctx, id, userID string) error
UpdateSort(ctx, userID string, ids []string) error
SetPinned(ctx, id, userID string, pinned bool) error
CountByUser(ctx, userID string) (int, error)
```

Notes :
- La limite est vérifiée dans le **service**, pas le repository
- `Delete` vérifie `userID` — suppression fichier physique dans le service

### `totp.go` — Interface `TOTPRepository`

```
Create(ctx, userID, encryptedSecret string) error
GetByUserID(ctx, userID string) (*model.TOTPSecret, error)
Verify(ctx, userID string) error
Delete(ctx, userID string) error
```

Notes :
- Le secret est chiffré AES-GCM **avant** stockage (dans le service)
- Le repository ne voit que des strings opaques

### `audit.go` — Interface `AuditRepository`

```
Log(ctx, entry *model.AuditEntry) error
ListByUser(ctx, userID string, offset, limit int) ([]*model.AuditEntry, int, error)
List(ctx, offset, limit int, filter AuditFilter) ([]*model.AuditEntry, int, error)

type AuditFilter struct {
    UserID *string
    Action *string
    From   *time.Time
    To     *time.Time
}
```

Notes :
- `Log` est fire-and-forget → erreur loggée slog mais jamais bloquante
- `List` pour admin, `ListByUser` pour panel user

---

## 8. Service — `internal/service/`

```go
type Services struct {
    Auth      AuthService
    User      UserService
    Bookmark  BookmarkService
    Wallpaper WallpaperService
    Admin     AdminService
    Email     EmailService
}

func New(repos *repository.Repositories, cfg *config.Config) *Services
```

### `auth.go` — Interface `AuthService`

```
Login(ctx, email, password, ip, userAgent string) (*model.Session, string, error)
Logout(ctx, sessionID string) error
LogoutAll(ctx, userID string) error
ValidateSession(ctx, token string) (*model.User, *model.Session, error)
BeginTOTP(ctx, userID string) (secret, qrCodeURL string, error)
ConfirmTOTP(ctx, userID, code string) error
ValidateTOTP(ctx, userID, code string) (bool, error)
DisableTOTP(ctx, userID, password string) error
ForgotPassword(ctx, email string) error
ResetPassword(ctx, token, newPassword string) error
```

Logique `Login()` :
1. `GetByEmail` → user non trouvé → erreur générique (pas d'énumération)
2. `user.IsActive = false` → erreur générique
3. Vérifier password argon2id
4. Si TOTP vérifié → valider code TOTP
5. Créer session : token = `crypto/rand` 32 bytes, `tokenHash = SHA-256(token)`
6. Insérer session en DB
7. Log audit `login`
8. Retourner session + token brut

Logique `ValidateSession()` :
1. `tokenHash = SHA-256(token cookie)`
2. `GetByTokenHash` → pas trouvé ou expiré → erreur
3. `GetByID(session.UserID)` → user inactif ou supprimé → erreur
4. Retourner user + session

Paramètres sécurité :
- Argon2id : `time=1, memory=64MB, threads=4`
- Token session : 32 bytes `crypto/rand` → base64url
- Cookie : `HttpOnly, Secure, SameSite=Strict, Path=/`
- Session lifetime : 30 jours
- `ForgotPassword` toujours `202` même si email inconnu

### `user.go` — Interface `UserService`

```
Register(ctx, username, email, password, ip, userAgent string) (*model.User, error)
GetByID(ctx, id string) (*model.User, error)
UpdateProfile(ctx, userID, username, email string) error
ChangePassword(ctx, userID, currentPassword, newPassword string) error
UpdateSearchEngine(ctx, userID, engine string, customURL *string) error
GetAuditLog(ctx, userID string, offset, limit int) ([]*model.AuditEntry, int, error)
```

Logique `Register()` :
1. `IsFirstUser()` → role = `admin`, sinon role = `user`
2. Valider username (alphanumérique, 3-32 chars)
3. Valider email format
4. Valider password (min 12 chars)
5. Hash argon2id
6. Générer ULID
7. `Create` en DB
8. Log audit `user_created`

### `bookmark.go` — Interface `BookmarkService`

```
Create(ctx, userID, url, title string, folder *string, tags []string) (*model.Bookmark, error)
Update(ctx, userID, bookmarkID, url, title string, folder *string, tags []string) error
Delete(ctx, userID, bookmarkID string) error
List(ctx, userID string, filter repository.BookmarkFilter) ([]*model.Bookmark, int, error)
UpdateSort(ctx, userID string, ids []string) error
ImportNetscape(ctx, userID string, data []byte) (imported, skipped int, error)
ExportNetscape(ctx, userID string) ([]byte, error)
GenerateBookmarklet(ctx, userID string) (string, error)
```

Logique `ImportNetscape()` :
1. Parser HTML format Netscape (`stdlib net/html`)
2. Extraire `DT/A` → url, title, `ADD_DATE`
3. Extraire `H3` parent → folder (chemin materialisé `"Dossier/Sous-dossier"`)
4. Pour chaque bookmark : valider URL, `GetOrCreate` tags, `BulkCreate` en transaction
5. Skipped = URLs invalides ou doublons
6. Log audit `bookmark_import` avec metadata `{imported, skipped}`

Logique `GenerateBookmarklet()` :
1. Générer token longue durée (90 jours) avec flag `is_bookmarklet = true`
2. Stocker en session
3. Retourner code JS avec token pré-authentifié :

```javascript
javascript:(function(){
    var u=encodeURIComponent(location.href);
    var t=encodeURIComponent(document.title);
    fetch('https://go.dotnot.be/api/bookmarks/quick',{
        method:'POST',
        headers:{'Content-Type':'application/json'},
        body:JSON.stringify({url:u,title:t,token:'TOKEN_ICI'})
    }).then(()=>alert('Bookmark sauvegardé'))
      .catch(()=>alert('Erreur'));
})();
```

### `wallpaper.go` — Interface `WallpaperService`

```
Upload(ctx, userID string, filename string, data []byte) (*model.Wallpaper, error)
Delete(ctx, userID, wallpaperID string) error
List(ctx, userID string) ([]*model.Wallpaper, error)
SetPinned(ctx, userID, wallpaperID string, pinned bool) error
UpdateSort(ctx, userID string, ids []string) error
```

Logique `Upload()` :
1. `CountByUser` → vérifier limite (`user.WallpaperLimit ?? config.DefaultWallpaperLimit`)
2. Valider extension : `.jpg .jpeg .png .webp .avif .mp4 .webm`
3. Valider taille max (`CAIRN_MAX_UPLOAD_SIZE`)
4. Valider magic bytes (pas juste l'extension)
5. Générer filename sécurisé : `{ulid}.{ext}`
6. Écrire dans `/data/media/{userID}/{ulid}.{ext}`
7. Insérer en DB
8. Log audit `wallpaper_upload`

Logique `Delete()` :
1. `GetByID` + vérifier `userID`
2. Supprimer fichier physique
3. Supprimer en DB
4. Log audit `wallpaper_delete`

### `admin.go` — Interface `AdminService`

```
ListUsers(ctx, offset, limit int) ([]*model.User, int, error)
GetUser(ctx, userID string) (*model.User, error)
SuspendUser(ctx, adminID, userID string) error
ActivateUser(ctx, adminID, userID string) error
DeleteUser(ctx, adminID, userID string) error
SetWallpaperLimit(ctx, adminID, userID string, limit *int) error
GetAuditLog(ctx, offset, limit int, filter repository.AuditFilter) ([]*model.AuditEntry, int, error)
GetStats(ctx) (*model.AdminStats, error)
```

Logique `SuspendUser()` :
1. Vérifier `adminID != userID`
2. Vérifier que le target n'est pas le dernier admin
3. `Update is_active = 0`
4. `DeleteByUserID(sessions)` → force déconnexion immédiate
5. Log audit `user_suspended`

Logique `DeleteUser()` :
1. Mêmes vérifications que `SuspendUser`
2. Supprimer tous les fichiers media du user
3. SoftDelete en DB (`deleted_at = now`)
4. `DeleteByUserID(sessions)`
5. Log audit `user_deleted`

### `email.go` — Interface `EmailService`

```
SendPasswordReset(ctx, email, token string) error
SendWelcome(ctx, email, username string) error
```

Notes :
- `net/smtp` stdlib, zéro dépendance externe
- STARTTLS si `CAIRN_SMTP_TLS = true`
- Templates HTML embarqués via `go:embed`
- Timeout connexion SMTP : 10 secondes
- Erreurs loggées slog, jamais propagées au caller HTTP

---

## 9. Handler + Routes — `internal/handler/`

### Format erreur standardisé

```json
{
    "error": "message lisible",
    "code":  "ERROR_CODE"
}
```

Codes : `INVALID_INPUT` (400), `UNAUTHORIZED` (401), `FORBIDDEN` (403),
`NOT_FOUND` (404), `CONFLICT` (409), `RATE_LIMITED` (429), `INTERNAL` (500)

### Table des routes complète

```
// Statique
GET  /*                                  → index.html

// Auth — public
POST /api/auth/register                  → Register
POST /api/auth/login                     → Login
POST /api/auth/logout                    → Logout (auth requise)
POST /api/auth/forgot-password           → ForgotPassword
POST /api/auth/reset-password            → ResetPassword

// User — authentifié
GET    /api/me                           → GetMe
PUT    /api/me                           → UpdateProfile
PUT    /api/me/password                  → ChangePassword
GET    /api/me/sessions                  → ListSessions
DELETE /api/me/sessions/{id}             → RevokeSession
DELETE /api/me/sessions                  → RevokeAllSessions
GET    /api/me/audit                     → GetMyAuditLog

// TOTP — authentifié
POST   /api/me/totp                      → BeginTOTP
PUT    /api/me/totp                      → ConfirmTOTP
DELETE /api/me/totp                      → DisableTOTP

// Bookmarklet — authentifié
GET    /api/me/bookmarklet               → GetBookmarklet
DELETE /api/me/bookmarklet               → RevokeBookmarklet

// Bookmarks — authentifié
GET    /api/bookmarks                    → ListBookmarks
POST   /api/bookmarks                    → CreateBookmark
GET    /api/bookmarks/{id}               → GetBookmark
PUT    /api/bookmarks/{id}               → UpdateBookmark
DELETE /api/bookmarks/{id}               → DeleteBookmark
PUT    /api/bookmarks/sort               → UpdateBookmarkSort
POST   /api/bookmarks/import             → ImportBookmarks
GET    /api/bookmarks/export             → ExportBookmarks

// Bookmarks quick — session bookmarklet uniquement
POST   /api/bookmarks/quick              → QuickBookmark

// Tags — authentifié
GET    /api/tags                         → ListTags
DELETE /api/tags/{id}                    → DeleteTag

// Wallpapers — authentifié
GET    /api/wallpapers                   → ListWallpapers
POST   /api/wallpapers                   → UploadWallpaper
DELETE /api/wallpapers/{id}              → DeleteWallpaper
PUT    /api/wallpapers/{id}/pin          → SetPinned
PUT    /api/wallpapers/sort              → UpdateWallpaperSort

// Media — authentifié
GET    /media/{userID}/{filename}        → ServeMedia

// Admin — authentifié + role admin
GET    /api/admin/users                  → AdminListUsers
GET    /api/admin/users/{id}             → AdminGetUser
PUT    /api/admin/users/{id}/suspend     → AdminSuspendUser
PUT    /api/admin/users/{id}/activate    → AdminActivateUser
DELETE /api/admin/users/{id}             → AdminDeleteUser
PUT    /api/admin/users/{id}/wallpaper-limit → AdminSetWallpaperLimit
GET    /api/admin/audit                  → AdminGetAuditLog
GET    /api/admin/stats                  → AdminGetStats

// Healthcheck — public, non loggé
GET    /healthcheck                      → { status: "ok", version: "x.x.x" }
```

### Montage chi

```
Router chi
│
├── Middleware globaux : RequestID, Logger, Recoverer, SecureHeaders, CORS, BodyLimit(1MB)
│
├── GET /* → frontend statique
│
├── /api/auth/* → RateLimit → handlers publics
│
├── /api/* → Auth
│   ├── /api/me/*
│   ├── /api/bookmarks/*
│   ├── /api/tags/*
│   └── /api/wallpapers/*
│
├── /api/bookmarks/quick → BookmarkletAuth
│
├── /api/admin/* → Auth + Admin
│
└── /media/* → Auth → ServeMedia
```

---

## 10. Middleware — `internal/middleware/`

### Globaux

```
RequestID     → génère X-Request-ID unique par requête
Logger        → slog structuré : method, path, status, latence, request_id
Recoverer     → catch panics → 500 propre + log stack trace
SecureHeaders → headers sécurité sur chaque réponse
```

### Headers sécurité

```
X-Content-Type-Options:  nosniff
X-Frame-Options:         DENY
Referrer-Policy:         strict-origin-when-cross-origin
Permissions-Policy:      geolocation=(), camera=(), microphone=()
Content-Security-Policy: default-src 'self';
                         img-src 'self' data:;
                         media-src 'self';
                         style-src 'self' fonts.googleapis.com;
                         font-src fonts.gstatic.com;
```

### Auth

```
1. Lire cookie cairn_session
2. Si absent → 401
3. SHA-256(token) → ValidateSession
4. Si invalide/expiré/user inactif → 401
5. Injecter user + session dans contexte chi
```

Helpers : `UserFromCtx(ctx)`, `SessionFromCtx(ctx)`

### Admin

```
1. UserFromCtx → user.Role != "admin" → 403
2. Monté après Auth obligatoirement
```

### BookmarkletAuth

```
1. Lire token dans body JSON
2. ValidateSession → vérifier is_bookmarklet == true
3. Droits restreints : POST /api/bookmarks/quick uniquement
```

### RateLimit

```
Login :           5 tentatives / 15 minutes / IP
Register :        3 tentatives / heure / IP
Forgot-password : 3 tentatives / heure / IP

Stockage : sync.Map en mémoire (pas de Redis)
Clé : SHA-256(IP)
Algorithme : sliding window
Réponse 429 + header Retry-After
IP : CF-Connecting-IP si CAIRN_TRUSTED_PROXY = true
```

### CORS

```
Allowed origins : CAIRN_BASE_URL uniquement
                  + localhost si CAIRN_ENV = development
Allowed methods : GET, POST, PUT, DELETE, OPTIONS
Allow credentials : true
Preflight OPTIONS → 204 sans Auth
```

### BodyLimit

```
Routes JSON standard :  1MB
Routes upload media :   CAIRN_MAX_UPLOAD_SIZE (défaut 50MB)
Route import bookmarks: 10MB
Si dépassé → 413 avant tout parsing
```

---

## 11. Logs

### Niveau applicatif (audit_log en DB)

Actions loggées :
```
login, logout, login_failed, password_change,
totp_enabled, totp_disabled, user_created, user_deleted,
user_suspended, bookmark_import, wallpaper_upload, wallpaper_delete
```

Visibilité :
- User : ses propres entrées uniquement
- Admin : toutes les entrées

### Niveau serveur (slog structuré → stdout → docker logs)

```
Erreurs HTTP 4xx/5xx
Erreurs DB
Erreurs SMTP
Panics récupérés
Démarrage / arrêt
Migrations
```

Format JSON en production, texte en development.

---

## 12. Sécurité — résumé

| Point | Implémentation |
|---|---|
| Passwords | Argon2id, time=1, memory=64MB, threads=4 |
| Sessions | HMAC-SHA256, token jamais stocké brut, cookie HttpOnly+Secure+SameSite=Strict |
| TOTP | RFC 6238, secret chiffré AES-GCM en DB |
| Pas d'énumération | register/login/forgot toujours réponse générique |
| Isolation users | userID vérifié dans chaque query repo |
| Media | Servis via middleware auth, jamais publics |
| Upload | Magic bytes validés, filename généré serveur, taille limitée |
| Path traversal | Impossible — chi extrait les params, pas de concat string |
| Rate limiting | Par IP hashée, sliding window, en mémoire |
| Headers HTTP | CSP, X-Frame-Options, HSTS via Traefik |
| Container | scratch, read-only FS, no-new-privileges, CAP_DROP ALL |
| SQL | Queries paramétrées via sqlc, zéro injection possible |
| Soft delete | Users supprimés conservés pour intégrité audit_log |

---

## 13. Déploiement

### `Dockerfile`

```dockerfile
# Stage 1 — Build
FROM golang:1.22-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux \
    go build \
    -ldflags="-s -w -X main.version=$(git describe --tags --always)" \
    -trimpath \
    -o cairn \
    ./cmd/cairn

# Stage 2 — Image finale
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/cairn /cairn

VOLUME ["/data"]
EXPOSE 8080

ENTRYPOINT ["/cairn"]
```

### `compose.yaml`

```yaml
name: cairn

services:
  cairn:
    image: cairn:latest
    container_name: cairn
    restart: unless-stopped

    security_opt:
      - no-new-privileges:true
    cap_drop:
      - ALL
    read_only: true
    tmpfs:
      - /tmp:size=64m,mode=1777

    volumes:
      - cairn_data:/data

    environment:
      CAIRN_ADDR:                       ":8080"
      CAIRN_ENV:                        "production"
      CAIRN_BASE_URL:                   "https://go.dotnot.be"
      CAIRN_DB_PATH:                    "/data/db.sqlite"
      CAIRN_MEDIA_PATH:                 "/data/media"
      CAIRN_DEFAULT_WALLPAPER_LIMIT:    "10"
      CAIRN_MAX_UPLOAD_SIZE:            "52428800"
      CAIRN_TOTP_ISSUER:                "Cairn"
      CAIRN_BOOKMARKLET_TOKEN_LIFETIME: "90"
      CAIRN_TRUSTED_PROXY:              "true"
      CAIRN_SMTP_HOST:                  "smtp.example.com"
      CAIRN_SMTP_PORT:                  "587"
      CAIRN_SMTP_FROM:                  "cairn@example.com"
      CAIRN_SMTP_TLS:                   "true"

    env_file:
      - .env  # CAIRN_SESSION_SECRET, CAIRN_SMTP_USER, CAIRN_SMTP_PASS

    networks:
      - traefik_proxy

    labels:
      traefik.enable: "true"
      traefik.http.routers.cairn.rule: "Host(`go.dotnot.be`)"
      traefik.http.routers.cairn.entrypoints: "websecure"
      traefik.http.routers.cairn.tls.certresolver: "cloudflare"
      traefik.http.services.cairn.loadbalancer.server.port: "8080"
      # Authentik SSO optionnel :
      # traefik.http.routers.cairn.middlewares: "authentik@file"

    healthcheck:
      test: ["/cairn", "healthcheck"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 10s

volumes:
  cairn_data:
    driver: local

networks:
  traefik_proxy:
    external: true
```

### `.env` (gitignored)

```bash
CAIRN_SESSION_SECRET=generer-avec-openssl-rand-base64-32
CAIRN_SMTP_USER=user@example.com
CAIRN_SMTP_PASS=mot-de-passe-smtp
```

### `.gitignore`

```
.env
data/
*.sqlite
*.sqlite-wal
*.sqlite-shm
```

---

## 14. Compatibilité Kubernetes

L'architecture est compatible k8s sans modification du code :

| Point | Status |
|---|---|
| Config via env vars uniquement | ✅ → ConfigMap + Secret k8s |
| Stateless sauf `/data` | ✅ → un seul PVC |
| Image multi-arch (amd64 + arm64) | ✅ → `docker buildx --platform linux/amd64,linux/arm64` |
| Healthcheck HTTP `/healthcheck` | ✅ → `livenessProbe` + `readinessProbe` |
| Secrets séparés de la config | ✅ → `.env` → Secret k8s |
| Binaire statique scratch | ✅ → portable partout |

Les manifests k8s ne sont pas dans le scope MVP — la migration sera triviale le moment venu.

---

## 15. Dépendances Go

```
github.com/go-chi/chi/v5          ← router HTTP
modernc.org/sqlite                 ← driver SQLite pure Go
github.com/pressly/goose/v3        ← migrations
github.com/oklog/ulid/v2           ← génération IDs
golang.org/x/crypto                ← argon2id
github.com/go-playground/validator/v10 ← validation structs
```

---

## 16. Ordre de développement recommandé

1. `go.mod` + structure dossiers
2. `internal/config/`
3. `internal/db/` + migrations
4. `internal/model/`
5. `internal/repository/`
6. `internal/service/auth.go` + `user.go`
7. `internal/middleware/`
8. `internal/handler/auth.go` + routing de base
9. Tests : login, session, logout
10. `internal/service/bookmark.go` + handler
11. Import/export Netscape + bookmarklet
12. `internal/service/wallpaper.go` + handler + ServeMedia
13. `internal/service/admin.go` + handler admin
14. `internal/service/email.go` + templates
15. Frontend `index.html` adapté pour l'API
16. Dockerfile + compose.yaml
17. Tests sécu : `govulncheck`, `trivy`
18. Build multi-arch + push
19. Repo GitHub `darktweek/cairn`
```

---

## 17. Écarts connus — évolutions postérieures à la spec (10 juin 2026)

La spec ci-dessus décrit le MVP. Les évolutions suivantes sont implémentées
et **font foi** lorsqu'elles contredisent les sections précédentes :

### Inscription & invitations (remplace §0 « Pas d'inscription publique »)
- Inscription ouverte optionnelle (toggle admin + `CAIRN_OPEN_REGISTRATION`) :
  username+email → email de validation avec lien 24h → setup TOTP (obligatoire)
  + mot de passe → compte créé. Table `pending_registrations` (migration 012).
- Invitations par email (admin) : lien → username + TOTP + mot de passe.
  Tables/colonnes : migrations 008-013.
- Vue admin des demandes en attente avec révocation.

### Suppression de comptes (remplace soft delete §8 AdminService)
- `DeleteUser` admin = **hard delete RGPD** (purge DB transactionnelle +
  suppression des médias). Le soft delete reste pour la suspension.
- Auto-suppression par l'utilisateur (`DELETE /api/me` avec mot de passe).

### Limites de stockage (complète §4 Config)
- `CAIRN_MAX_UPLOAD_SIZE` (50 Mo) : taille max d'**un fichier**, override
  par user (`users.upload_size_limit`, route admin dédiée).
- `CAIRN_STORAGE_QUOTA` (200 Mo) : stockage média **total** par user,
  override par user (`users.storage_quota`, migration 014).
- La route d'upload applique la limite par user au niveau HTTP
  (`middleware.UserBodyLimit`), exclue du cap global.

### Audit (remplace §11 liste fermée d'actions)
- La contrainte `CHECK(action IN ...)` de la migration 007 a été supprimée
  (migration 015) — le vocabulaire des actions appartient à l'application.
- Actions ajoutées : `login_sso`, `user_created_sso`,
  `register_blocked_duplicate_email`, `registration_requested`,
  `registration_completed`, `registration_revoked`, `invitation_sent`,
  `invitation_revoked`.

### Divers
- SSO/OIDC optionnel (config env ou admin), réglages SMTP éditables en
  admin si non verrouillés par l'env, locale par user (fr/en), menu hub
  (`CAIRN_MENU_BANG`), effets visuels pluie + poussière, thème adaptatif
  par luminance (crop central 200px, ré-échantillonnage vidéo 10s).
- tmpfs `/tmp` du compose à 256m (spool multipart des gros uploads).
- `/media/{userID}/*` : vérification de propriété (user = userID ou admin).
