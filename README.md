# Cairn

A self-hosted personal start page — new-tab replacement for your browser.  
Clock, ambient effects, wallpapers, configurable search, bookmarks, multi-user with admin panel.

> **No cloud. No tracking. Runs entirely on your own machine or server.**

This project is **100% vibecoded** — designed and built in conversation with AI (Claude Code), iterated feature by feature on a homelab. Craftsmanship for the joy of it, not for scale.

💬 **Questions, ideas, help?** Join the Discord → **https://discord.gg/2S9v5MKfGg**

---

## What it looks like

*The home — clock, search, your wallpaper (image or video), rain effect:*

![Home](docs/screenshot-home.png)

*The hub (`!menu` in the search bar) — liquid-glass tiles over your background:*

![Hub](docs/screenshot-hub.png)

*The login:*

![Login](docs/screenshot-login.png)

Type `!bm` in the search bar to open the bookmark manager.  
Type `!menu` (or your configured bang) for the full-screen hub.

---

## Features

| Feature | Details |
|---|---|
| **Clock & date** | Large italic clock, ISO week number |
| **Wallpapers** | Images & videos, pin favorite, random rotation, adaptive light/dark theme via luminance sampling |
| **Ambient effects** | Rain and dust canvas animations (opt-in per-user) |
| **Search** | DuckDuckGo by default — Google, Brave, Bing, Kagi, or custom URL; per-user "open in new tab" option |
| **Bangs** | `!bm` bookmarks, `!g` Google, `!yt` YouTube, `!gh` GitHub, `!hub` full menu, + all DDG bangs |
| **Bookmarks** | Collections with a nested folder tree, tags, import/export Netscape format (Chrome/Firefox/Safari/Edge), mobile bookmarklet |
| **Collaboration** | Share collections with individual users or whole groups — `view` / `edit` / `manage` permission levels, "shared" indicators, optional email notification |
| **Roles & permissions (RBAC)** | Bitwarden-style roles over a granular permission catalog; seeded `owner` / `admin` / `user` plus custom roles; a user can hold **several** roles (effective permissions = union) |
| **Groups / teams** | Bundle users into groups and grant a whole team access to a collection |
| **TOTP / 2FA** | RFC 6238, server-generated QR code, secrets encrypted at rest — enforced for invited / email-verified signups |
| **Multi-user** | First account created becomes the instance **owner** automatically |
| **Invitations** | Admin invites by email (with delivery confirmation) **or** by copying the invite link — works without SMTP; open registration toggle |
| **SSO / OIDC** | OpenID Connect (Authentik, Keycloak, Authelia, Google…) — JIT account provisioning |
| **Admin panel** | User management, storage quotas, upload limits, SMTP test, audit log, pending registrations |
| **Email** | Account setup & invitation emails; SMTP supports implicit TLS (port 465) and STARTTLS (port 587); configurable via env or admin UI |
| **Audit log** | All security and admin actions logged and attributed to users; GDPR-compliant (no personal content logged, metadata preserved on account deletion) |

---

## Requirements

**Docker** (Compose recommended) — that's it. A prebuilt multi-arch image
(`linux/amd64`, `linux/arm64`) is published to GHCR, so nothing is compiled on
your machine:

```
ghcr.io/darktweek/cairn:latest   # or pin a version, e.g. :v0.2.3
```

---

## Quick start

### Option A — one command

```bash
docker run -d --name cairn -p 8080:8080 \
  -e CAIRN_BASE_URL=http://localhost:8080 \
  -e CAIRN_SESSION_SECRET="$(openssl rand -base64 32)" \
  -v cairn_data:/data \
  ghcr.io/darktweek/cairn:latest
```

Open **http://localhost:8080**, click **Sign in → Register** — the **first account
becomes the instance owner**. That's it.

### Option B — Docker Compose (recommended)

Create a `compose.yaml`:

```yaml
services:
  cairn:
    image: ghcr.io/darktweek/cairn:latest
    container_name: cairn
    restart: unless-stopped
    ports:
      - "8080:8080"                              # remove when behind a reverse proxy
    environment:
      CAIRN_BASE_URL: "http://localhost:8080"    # your public URL in production
      CAIRN_SESSION_SECRET: "change-me"          # openssl rand -base64 32
    volumes:
      - cairn_data:/data
volumes:
  cairn_data:
```

```bash
docker compose up -d
```

> **No SMTP?** You can run entirely without it. Create your **owner** account from
> the *Register* link (open registration is on by default and auto-disables once
> that first account exists). To add more users, go to **Admin → Invitations** and
> **copy the invite link** shown right after creating one — no email needed. SMTP
> only *delivers* invitations and password resets automatically.

Pin a version with `:v0.2.3` instead of `:latest`. See the
[Configuration reference](#configuration-reference) for all environment variables.

---

## Set it as your new-tab page

### Chrome / Edge / Brave
Install the [New Tab Redirect](https://chrome.google.com/webstore/detail/new-tab-redirect/icpgjfneehieebagbmdbhnlpiopdcmna) extension and point it to `http://localhost:8080`.

### Firefox
Use [New Tab Homepage](https://addons.mozilla.org/en-US/firefox/addon/new-tab-homepage/) or set it in `about:preferences` under Home.

### Safari
Settings → General → New tabs open with → Homepage → set your URL.

---

## Configuration reference

All configuration is via environment variables.  
Sensitive values go in `.env` (already gitignored).

### Core

| Variable | Default | Required | Description |
|---|---|---|---|
| `CAIRN_ADDR` | `:8080` | no | Listen address |
| `CAIRN_ENV` | `production` | no | `production` or `development` |
| `CAIRN_BASE_URL` | — | **yes** | Your public URL, e.g. `https://start.example.com` |
| `CAIRN_DB_PATH` | `/data/db.sqlite` | no | SQLite database path |
| `CAIRN_MEDIA_PATH` | `/data/media` | no | Wallpaper storage directory |
| `CAIRN_SESSION_SECRET` | — | **yes** | HMAC key, minimum 32 characters |

### Uploads & storage

| Variable | Default | Description |
|---|---|---|
| `CAIRN_DEFAULT_WALLPAPER_LIMIT` | `10` | Max wallpapers per user |
| `CAIRN_MAX_UPLOAD_SIZE` | `52428800` | Max single file size in bytes (50 MB) |
| `CAIRN_STORAGE_QUOTA` | `209715200` | Max total media per user in bytes (200 MB) |

Per-user overrides for both limits are available in the admin panel.

### SMTP (email)

| Variable | Default | Required | Description |
|---|---|---|---|
| `CAIRN_SMTP_HOST` | — | **yes** | SMTP server hostname |
| `CAIRN_SMTP_PORT` | `587` | no | SMTP port |
| `CAIRN_SMTP_USER` | — | **yes** | SMTP username |
| `CAIRN_SMTP_PASS` | — | **yes** | SMTP password |
| `CAIRN_SMTP_FROM` | — | **yes** | Sender address |
| `CAIRN_SMTP_TLS` | `true` | no | Enable TLS (port 465 = implicit TLS/SMTPS; other ports = STARTTLS) |

SMTP can also be configured entirely from the admin UI if not set via environment.  
Use the **Test** button in Admin → Settings → SMTP to verify delivery before sending real invitations.

### SSO / OpenID Connect (optional)

| Variable | Default | Description |
|---|---|---|
| `CAIRN_OIDC_ISSUER` | — | OIDC issuer URL. If set, locks SSO config (otherwise editable in admin) |
| `CAIRN_OIDC_CLIENT_ID` | — | OIDC client ID |
| `CAIRN_OIDC_CLIENT_SECRET` | — | OIDC client secret (put in `.env`) |
| `CAIRN_OIDC_PROVIDER_NAME` | `SSO` | Label shown on the "Sign in with …" button |
| `CAIRN_OIDC_SCOPES` | `openid profile email` | Requested scopes |

**Redirect URI to register with your provider:**
```
<CAIRN_BASE_URL>/api/auth/sso/callback
```

**Provisioning:** accounts are created automatically on first SSO login (JIT). Any user who can authenticate with your provider will get a Cairn account — restrict access at the provider level (groups, policies) if needed. Existing accounts are matched by email.

### Misc

| Variable | Default | Description |
|---|---|---|
| `CAIRN_OPEN_REGISTRATION` | `true` | Allow public self-registration. **Set to `false` for any public-facing instance** — see [Security](#security) |
| `CAIRN_SESSION_LIFETIME_DAYS` | `30` | Browser session lifetime, in days (drives both the DB session and the cookie max-age) |
| `CAIRN_INVITE_LIFETIME` | `72` | Invitation link lifetime, in hours |
| `CAIRN_TRUSTED_PROXY` | `true` | Trust `X-Forwarded-For` for the client IP. Keep `true` **only** behind a reverse proxy; set `false` if Cairn is exposed directly |
| `CAIRN_MENU_BANG` | — | Bang that opens the full-screen menu (default `!menu`, editable in admin if not set here) |
| `CAIRN_TOTP_ISSUER` | `Cairn` | Name shown in your authenticator app |
| `CAIRN_BOOKMARKLET_TOKEN_LIFETIME` | `90` | Bookmarklet token lifetime in days (a bookmarklet token is a full-access session token — revoke from the account panel if leaked) |

---

## Behind a reverse proxy

Serve Cairn over HTTPS behind your proxy: **remove the `ports` mapping** from the
`cairn` service and let the proxy reach it. Keep `CAIRN_TRUSTED_PROXY=true` (the
default) so client IPs are read from `X-Forwarded-For`, and set `CAIRN_BASE_URL`
to your public `https://…` URL (this also makes session cookies `Secure`).

### Traefik

```yaml
services:
  cairn:
    image: ghcr.io/darktweek/cairn:latest
    restart: unless-stopped
    environment:
      CAIRN_BASE_URL: "https://start.example.com"
      CAIRN_SESSION_SECRET: "change-me"
    volumes:
      - cairn_data:/data
    networks:
      - traefik_proxy
    labels:
      traefik.enable: "true"
      traefik.http.routers.cairn.rule: "Host(`start.example.com`)"
      traefik.http.routers.cairn.entrypoints: "websecure"
      traefik.http.routers.cairn.tls.certresolver: "cloudflare"
      traefik.http.services.cairn.loadbalancer.server.port: "8080"

volumes:
  cairn_data:
networks:
  traefik_proxy:
    external: true
```

### Nginx

Keep `ports: - "127.0.0.1:8080:8080"` on the service, then:

```nginx
location / {
    proxy_pass         http://127.0.0.1:8080;
    proxy_set_header   Host              $host;
    proxy_set_header   X-Forwarded-For   $remote_addr;
}
```

---

## Wallpapers

**Accepted formats:**

| Type | Extensions |
|---|---|
| Image | `.jpg` `.jpeg` `.png` `.webp` `.avif` |
| Video | `.mp4` `.webm` |

**Adaptive theme:** Cairn samples the luminance of your active wallpaper and automatically switches between light and dark text for readability.

**Single pin:** Only one wallpaper can be pinned as favorite at a time. Pinning a new one automatically unpins the previous.

---

## Bookmarklet

Save any page in one click from any browser, including mobile Safari:

1. Go to **Account → Bookmarklet → Generate bookmarklet**
2. Drag the link to your browser's bookmarks bar
3. Click it on any page to save it to Cairn

---

## Registrations & invitations

Registration is controlled by `CAIRN_OPEN_REGISTRATION` (and the **Admin → Settings → Registration** toggle).

- **Invite-only (recommended):** the admin sends invite links by email; invited signups are **TOTP-enforced**.
- **Open registration:** anyone can self-register with username + email + password. These accounts are **password-only** (no TOTP, no email verification).

> ⚠️ The **first account created becomes the instance `owner`** (full control). Open registration **auto-disables once that first account exists**, so create your owner account first. For a public-facing instance, also start with `CAIRN_OPEN_REGISTRATION=false` — see [Security](#security).

---

## Collaboration, roles & permissions

Cairn ships a Bitwarden-style authorization model with two independent layers.

**Collections & sharing.** Bookmarks live in *collections* (each with a nested folder tree). Every user has one personal collection. Any other collection can be shared with:

- **individual users**, or
- **groups** (teams),

at one of three permission levels: **view** (read-only), **edit** (add/modify/delete bookmarks & folders), or **manage** (edit + manage the collection's shares). The collection picker marks shared collections with a 👥 indicator. Sharing can optionally email the new collaborator.

**Instance roles (RBAC).** What a user may do at the system level is governed by *roles*, each a bundle of fine-grained permissions from a fixed catalog:

`audit.view` · `bookmarks.import_export` · `collections.create` · `collections.manage_all` · `collections.delete_any` · `groups.manage` · `users.manage` · `settings.manage` · `roles.manage`

Three system roles are seeded — **owner** (all permissions), **admin** (day-to-day administration), **user** (no instance permissions) — and admins with `roles.manage` can create **custom roles** from **Admin → Roles**. A user can hold **multiple roles**; their effective permissions are the union. Guard-rails: an actor can only grant permissions they hold, can't strip the last owner, and can't modify a user more privileged than themselves.

**GDPR / privacy policies.** Under **Admin → Settings**, three opt-in policies (all **off** by default) let you: allow admins to manage *all* collections (logged in the audit trail), restrict collection creation, and restrict collection deletion. By default an admin **cannot** see other users' private collections.

---

## Build from source (contributors)

You don't need Go installed — everything builds in Docker.

```bash
git clone https://github.com/darktweek/cairn.git
cd cairn

# Build the image locally
docker build -t cairn:dev .

# …or run the CI pipeline (build, vet, test)
docker run --rm -v "$PWD":/src -w /src golang:1.26 \
  sh -c 'go build ./... && go vet ./... && go test ./...'
```

To run your local build with Compose, add `build: .` (and `image: cairn:dev`) to
the `cairn` service. Set `CAIRN_ENV: development` for human-readable logs.

See **[CONTRIBUTING.md](CONTRIBUTING.md)** for the dev workflow and contributor
license terms.

---

## Project structure

```
cairn/
├── cmd/cairn/          — entrypoint (main.go, router, graceful shutdown)
├── internal/
│   ├── config/         — config loading and validation
│   ├── db/             — SQLite setup + embedded goose migrations
│   ├── model/          — data structs
│   ├── repository/     — database access (interfaces + SQLite implementations)
│   ├── service/        — business logic (auth, bookmarks, collections, rbac, groups, wallpapers, admin…)
│   ├── handler/        — HTTP JSON handlers
│   └── middleware/     — auth, permission/RBAC, rate limit, CORS, headers, bookmarklet
├── web/static/
│   ├── index.html      — HTML shell
│   ├── style.css       — styles (CSS variables, adaptive theme)
│   └── app.js          — vanilla JS SPA (zero dependencies)
├── .env.example        — environment variable template
├── Dockerfile          — multi-stage build → ~5 MB scratch image
├── compose.yaml        — deployment (pulls the GHCR image)
└── .github/workflows/  — CI (build/vet/test) + Release (GHCR image on tags)
```

---

## Security

| Area | Implementation |
|---|---|
| Passwords | Argon2id (time=1, memory=64 MB, threads=4) |
| Sessions | SHA-256 hashed tokens, `HttpOnly` + `Secure` + `SameSite=Strict` cookies, configurable lifetime |
| TOTP | RFC 6238 for invited/email-verified signups — server-generated QR code, secrets encrypted AES-256-GCM at rest |
| Rate limiting | Two-layer: per-account (10/5 min) + per-IP fallback (30/5 min); register/forgot 3/hour |
| Authorization | Permission-checked admin routes; RBAC anti-escalation (grant only what you hold), last-owner guard, no modifying a more-privileged user; collection ACL enforced server-side (`view`/`edit`/`manage`), forbidden/unknown returns `404` to hide existence |
| User isolation | Data scoped per user / per collection; media served behind auth and path-traversal-safe |
| CORS | Locked to `CAIRN_BASE_URL`; cross-origin requests get no `Access-Control-Allow-Origin` |
| Dependencies | `govulncheck` clean; pure-Go, zero CGO |
| Uploads | Magic bytes validated, server-generated filenames |
| Container | `scratch` base, read-only FS, `no-new-privileges`, `CAP_DROP ALL` |
| HTTP headers | CSP, `X-Frame-Options: DENY`, `Referrer-Policy`, `Permissions-Policy` |
| Audit log | All security events logged (login, logout, password change, TOTP, account lifecycle, admin actions) |
| GDPR | Hard delete purges all user data and media; audit entries retain username in metadata, `user_id` set to NULL |
| Email TLS | Port 465 → implicit TLS (SMTPS); port 587 → STARTTLS upgrade |

### Hardening a public-facing instance

Before exposing Cairn to the internet:

1. **Create your owner account first** — open registration auto-disables once it exists. For belt-and-braces, also set `CAIRN_OPEN_REGISTRATION=false`.
2. Serve **over HTTPS behind a reverse proxy** so session cookies are `Secure`; let the proxy add HSTS.
3. Keep `CAIRN_TRUSTED_PROXY=true` **only** behind that proxy — otherwise set it to `false` to prevent `X-Forwarded-For` spoofing.
4. Use a strong random `CAIRN_SESSION_SECRET` (≥ 32 chars, e.g. `openssl rand -base64 32`).
5. Consider a shorter session: `CAIRN_SESSION_LIFETIME_DAYS=7`.
6. Prefer **invitations** (TOTP-enforced) over open registration.

A full pre-release review (auth, sessions, RBAC/ACL, injection, XSS, CSRF, SSRF, headers, uploads, dependencies) is summarised in [`SECURITY.md`](SECURITY.md).

---

## Tech stack

| Component | Choice |
|---|---|
| Language | Go 1.26 |
| Router | chi v5 |
| Database | SQLite (WAL mode) |
| SQLite driver | modernc.org/sqlite (pure Go, zero CGO) |
| Migrations | goose (embedded) |
| QR code | skip2/go-qrcode (server-side PNG for TOTP setup) |
| Docker image | `scratch` (~5 MB) |
| Frontend | Vanilla JS, zero dependencies |

---

## Community

- 💬 **Discord** — questions, ideas, support: https://discord.gg/2S9v5MKfGg
- 🐛 **Issues** — bug reports & feature requests on GitHub
- 🤝 **Contributing** — see [`CONTRIBUTING.md`](CONTRIBUTING.md)

---

## License

Cairn is **dual-licensed**:

- **GNU AGPL-v3** for open-source / self-hosted use — see [`LICENSE`](LICENSE).
  If you run a modified version as a network service, the AGPL (§13) requires
  you to make your source modifications available to its users.
- **Commercial license** — to embed Cairn in a proprietary product, or to offer
  it as a hosted/SaaS service *without* the AGPL's source-disclosure
  obligations. See [`LICENSING.md`](LICENSING.md).

© 2026 darktweek. Contributions are accepted under [`CONTRIBUTING.md`](CONTRIBUTING.md).
