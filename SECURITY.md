# Security

## Reporting a vulnerability

Please report security issues **privately** — open a
[GitHub security advisory](https://github.com/darktweek/cairn/security/advisories/new)
or email the maintainer rather than filing a public issue. Include reproduction
steps and the affected version/commit. You'll get an acknowledgement and a fix
timeline.

## Security model (summary)

| Area | Implementation |
|---|---|
| Passwords | Argon2id (t=1, m=64 MB, p=4) |
| Sessions | random token, **SHA-256 hashed at rest**; `HttpOnly` + `Secure` + `SameSite=Strict` cookie; server-side expiry; configurable lifetime |
| 2FA | TOTP (RFC 6238) for invited/email-verified signups; secrets AES-256-GCM encrypted |
| Password reset | hashed, single-use, expiring token; all sessions invalidated on reset |
| Authorization | permission-checked admin routes; collection ACL (`view`/`edit`/`manage`) enforced server-side; forbidden/unknown → `404` |
| RBAC guard-rails | grant only permissions you hold; last owner protected; can't modify a more-privileged user |
| Injection | 100% parameterized SQL (no string-built queries) |
| XSS | strict CSP (`default-src 'self'`, no `unsafe-inline` script); DOM built with `textContent` |
| CSRF | `SameSite=Strict` session cookie; CORS locked to `CAIRN_BASE_URL` |
| SSRF | outbound calls (OIDC, SMTP) use admin-configured endpoints only |
| Uploads | magic-byte validation, server-generated filenames, per-user quotas |
| Media | path-traversal-safe (`filepath.Clean` + root-prefix check), behind auth |
| Headers | CSP, `X-Frame-Options: DENY`, `X-Content-Type-Options`, `Referrer-Policy`, `Permissions-Policy` |
| Rate limiting | per-account + per-IP on login; register/forgot 3/hour |
| Container | `scratch` base, no shell, zero CGO |
| Dependencies | `govulncheck` clean |

## Pre-release review (2026-06-28)

A code-level review plus live probing covered: authentication & sessions,
RBAC/collection authorization (IDOR, privilege escalation), SQL injection, XSS,
CSRF, SSRF, security headers, file uploads, secret handling, and dependency
CVEs (`govulncheck`). No critical code-level findings. Items addressed:

- **Fixed** — an actor can no longer modify a user holding permissions they
  lack (privilege guard added to role assignment).
- **Fixed** — role badges no longer use `innerHTML` (defence in depth).
- **Config** — open registration + first-user-becomes-owner: see hardening
  below. The public read-only share-link feature was removed entirely.

## Hardening a public-facing instance

1. Create your **owner** account first, then `CAIRN_OPEN_REGISTRATION=false`.
2. Serve over **HTTPS behind a reverse proxy** (Secure cookies, HSTS at proxy).
3. `CAIRN_TRUSTED_PROXY=true` only behind that proxy; otherwise `false`.
4. Strong `CAIRN_SESSION_SECRET` (≥ 32 chars).
5. Shorter sessions: `CAIRN_SESSION_LIFETIME_DAYS=7`.
6. Prefer **invitations** (TOTP-enforced) over open registration.

See the [Security section of the README](README.md#security) for the full table.
