# Changelog

All notable changes to Cairn are documented here. Format loosely follows
[Keep a Changelog](https://keepachangelog.com/); versions are date-stamped.

## v0.2.5 — Custom branding — 2026-06-30

### Added
- **Custom site name**: admins can set the name shown in the browser tab, the
  login page title, and all outgoing emails (Admin → Settings → Site name).
  Defaults to `Cairn` when not set.
- **Custom favicon**: admins can provide a favicon URL (Admin → Settings →
  Favicon URL). The browser tab icon updates immediately without a page reload.
- Both values are exposed on the public `/api/auth/config` endpoint so the
  branding is applied even before login.

## v0.2.4 — First-account SMTP bypass — 2026-06-30

### Added
- **First-account setup without SMTP**: when no user exists yet, the registration
  setup URL is printed prominently in the container logs (WARN level) so a fresh
  install can be bootstrapped without configuring a mail server first.

## v0.2.3 — Source link & CI — 2026-06-30

### Added
- A discreet **"Source code (AGPL-3.0)"** link on the login page and in the hub,
  offering the source to users of the running instance (AGPL §13).

### CI
- Bumped GitHub Actions to Node 24-compatible majors (checkout v7, setup-go v6,
  docker setup/login/metadata/build-push v4–v7) — clears the Node 20 deprecation
  warnings.

## v0.2.2 — Invite without SMTP — 2026-06-29

### Added
- Creating or resending an invitation now returns the **invite link**, shown in
  the admin Invitations tab with a **Copy link** button — so you can onboard
  users even without SMTP configured (SMTP only *delivers* invites by email).

### Docs
- Corrected the "No SMTP" guidance (there is no direct user-creation endpoint;
  use open registration or the copyable invite link).

## v0.2.1 — Security hardening — 2026-06-28

### Security
- Open registration **auto-disables once the first (owner) account exists**,
  closing the "first signup wins owner" race on a freshly-exposed instance.
  An admin can re-enable it from Admin → Settings.
- JSON request bodies are capped at 1 MB (the wallpaper-upload and bookmark-
  import routes keep their own larger limits).

### Repo
- Stopped tracking local artifacts (`.DS_Store`, `.claude/`) and extended
  `.gitignore` ahead of going public.
- Licensed under **GNU AGPL-v3** with a **commercial** option (dual licensing);
  added `LICENSE`, `LICENSING.md`, `CONTRIBUTING.md`.

## v0.2.0 — Collaboration & RBAC — 2026-06-28

The big one: bookmark collaboration and a Bitwarden-style authorization model.

### Added
- **Collections** as first-class, shareable containers with a **nested folder
  tree**. Every user gets one personal collection; legacy folder strings are
  migrated into it automatically.
- **Collection sharing** with individual users at three permission levels —
  `view` / `edit` / `manage`. A 👥 indicator marks shared collections; an
  optional email notifies the new collaborator.
- **Groups / teams** — bundle users and share a collection with a whole team.
- **Instance RBAC** — roles built from a granular permission catalog
  (`audit.view`, `users.manage`, `roles.manage`, `collections.create`, …).
  Seeded `owner` / `admin` / `user` plus admin-defined **custom roles**. A user
  can hold **several roles** (effective permissions = union).
- **GDPR policies** (admin, all off by default): allow admins to manage all
  collections (audited), restrict collection creation, restrict deletion.
- Per-user **"open searches in a new tab"** preference.
- Configurable **session lifetime** via `CAIRN_SESSION_LIFETIME_DAYS`.

### Changed
- The admin area is now gated by **fine-grained permissions** instead of a
  binary admin flag; the admin menu appears for any admin-area permission.
- The first account created becomes the instance **owner** (was: admin).

### Security
- RBAC guard-rails: anti-escalation (grant only permissions you hold), last
  owner cannot be demoted, and **an actor cannot modify a more-privileged user**.
- Collection access is enforced server-side; forbidden/unknown resources return
  `404` to avoid leaking their existence.
- Role badges are rendered with `textContent` (no `innerHTML`) — defence in
  depth on top of the strict CSP.
- Pre-release security review completed (see [`SECURITY.md`](SECURITY.md));
  `govulncheck` reports no known vulnerabilities.

### Database
- Migrations `018`–`024` (collections, folders, roles, role permissions,
  collection shares, groups, group shares, multi-role junction). Applied
  automatically on start; backward-compatible, no manual steps.

## v0.1.0 — Initial release

Self-hosted start page: clock, wallpapers, search + bangs, bookmarks (folders,
tags, import/export, bookmarklet), TOTP 2FA, multi-user with invitations,
SSO/OIDC, admin panel, audit log, SMTP email.
