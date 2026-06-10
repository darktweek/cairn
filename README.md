# Cairn

Start page personnelle, self-hosted, multi-utilisateur.  
Horloge, pluie, fond d'écran, moteur de recherche configurable, marque-pages.

---

## Aperçu

```
┌─────────────────────────────────────────────────────┐
│                                                     │
│              [fond d'écran ou vidéo]                │
│              [animation pluie canvas]               │
│                                                     │
│                    12:34                            │
│           Lundi 09 juin 2026 · S23                  │
│                                                     │
│           [ barre de recherche ]                    │
│                                                     │
└─────────────────────────────────────────────────────┘
```

Tapez `!bm` dans la barre de recherche pour ouvrir le gestionnaire de marque-pages.

---

## Fonctionnalités

- **Horloge** — grande, italique, Cormorant Garant
- **Date** avec numéro de semaine ISO
- **Fond d'écran** — images et vidéos, sélection aléatoire, épinglage, thème clair/sombre adaptatif par analyse de luminance
- **Pluie** — animation canvas en fond
- **Recherche** — DuckDuckGo par défaut, moteur configurable par utilisateur (Google, Brave, Bing, Kagi, custom)
- **Bangs** — `!bm`/`!edit` (marque-pages), `!g`, `!yt`, `!gh`, `!hub`, + tous les bangs DDG
- **Marque-pages** — dossiers, tags, import/export Netscape (Chrome/Firefox/Safari/Edge), bookmarklet mobile
- **TOTP** — optionnel, RFC 6238, compatible toutes apps authenticator
- **Multi-utilisateur** — le premier compte créé devient admin automatiquement
- **Panel admin** — stats, gestion des comptes, journal d'audit

---

## Stack

| Composant | Choix |
|---|---|
| Langage | Go 1.23 |
| Router | chi v5 |
| Base de données | SQLite WAL |
| Driver SQLite | modernc.org/sqlite (pure Go, zéro CGO) |
| Image Docker | `scratch` (~5 MB) |
| Frontend | Vanilla JS, zéro dépendance |

---

## Démarrage rapide

### Prérequis

- Docker + Docker Compose

### 1. Cloner le dépôt

```bash
git clone https://github.com/darktweek/cairn.git
cd cairn
```

### 2. Créer le fichier `.env`

```bash
cp .env.example .env
```

Éditez `.env` et renseignez les trois variables obligatoires :

```bash
CAIRN_SESSION_SECRET=      # générer : openssl rand -base64 32
CAIRN_SMTP_USER=           # adresse email SMTP
CAIRN_SMTP_PASS=           # mot de passe SMTP
```

### 3. Adapter `compose.yaml`

Remplacez les valeurs suivantes dans `compose.yaml` selon votre environnement :

```yaml
CAIRN_BASE_URL:  "https://go.example.com"   # votre URL publique
CAIRN_SMTP_HOST: "smtp.example.com"
CAIRN_SMTP_FROM: "cairn@example.com"
```

### 4. Démarrer

```bash
docker compose up -d
```

L'application est disponible sur `http://localhost:8080`.

### 5. Premier utilisateur

Rendez-vous sur `http://localhost:8080` et créez un compte via **Se connecter**.  
**Le premier compte créé obtient automatiquement le rôle admin.**

---

## Configuration complète

Toute la configuration se fait via variables d'environnement.  
Les valeurs sensibles vont dans `.env` (gitignored).

| Variable | Défaut | Requis | Description |
|---|---|---|---|
| `CAIRN_ADDR` | `:8080` | non | Adresse d'écoute |
| `CAIRN_ENV` | `production` | non | `production` ou `development` |
| `CAIRN_BASE_URL` | — | **oui** | URL publique, ex: `https://go.example.com` |
| `CAIRN_DB_PATH` | `/data/db.sqlite` | non | Chemin de la base SQLite |
| `CAIRN_MEDIA_PATH` | `/data/media` | non | Répertoire des fonds d'écran |
| `CAIRN_SESSION_SECRET` | — | **oui** | Clé HMAC, minimum 32 caractères |
| `CAIRN_DEFAULT_WALLPAPER_LIMIT` | `10` | non | Limite de fonds d'écran par utilisateur |
| `CAIRN_MAX_UPLOAD_SIZE` | `52428800` | non | Taille max upload en octets (50 MB) |
| `CAIRN_BOOKMARKLET_TOKEN_LIFETIME` | `90` | non | Durée de vie du token bookmarklet (jours) |
| `CAIRN_TOTP_ISSUER` | `Cairn` | non | Nom affiché dans l'app authenticator |
| `CAIRN_TRUSTED_PROXY` | `true` | non | Lire `CF-Connecting-IP` / `X-Forwarded-For` |
| `CAIRN_MENU_BANG` | — | non | Bang du menu plein écran. Vide = éditable dans l'admin (défaut `!menu`) |
| `CAIRN_OIDC_ISSUER` | — | non | URL issuer OIDC. Si défini, verrouille la config SSO (sinon éditable en admin) |
| `CAIRN_OIDC_CLIENT_ID` | — | non | Client ID OIDC |
| `CAIRN_OIDC_CLIENT_SECRET` | — | non | Client Secret OIDC (mettre dans `.env`) |
| `CAIRN_OIDC_PROVIDER_NAME` | `SSO` | non | Nom affiché sur le bouton « Se connecter avec … » |
| `CAIRN_OIDC_SCOPES` | `openid profile email` | non | Scopes demandés |
| `CAIRN_SMTP_HOST` | — | **oui** | Serveur SMTP |
| `CAIRN_SMTP_PORT` | `587` | non | Port SMTP |
| `CAIRN_SMTP_USER` | — | **oui** | Utilisateur SMTP |
| `CAIRN_SMTP_PASS` | — | **oui** | Mot de passe SMTP |
| `CAIRN_SMTP_FROM` | — | **oui** | Adresse expéditeur |
| `CAIRN_SMTP_TLS` | `true` | non | STARTTLS activé |

---

## Menu

Cairn n'a pas de barre de navigation. Le menu s'ouvre en tapant un **bang** dans
la barre de recherche (défaut `!menu`) : il morph en hub plein écran avec des
tuiles (Marque-pages, Compte, Administration, Déconnexion).

Le bang est configurable : via `CAIRN_MENU_BANG` (verrouillé) ou depuis
**Admin → Réglages → Menu**.

---

## SSO (OpenID Connect)

Cairn s'intègre à n'importe quel fournisseur OIDC standard (Authentik, Keycloak,
Authelia, Google…) via le flux Authorization Code + PKCE, sans dépendance externe.

Quand un provider est configuré, la page de connexion affiche un bouton
**« Se connecter avec <nom du provider> »** au-dessus du formulaire classique.
Les comptes sont provisionnés à la volée (JIT) au premier login, liés par email.
Si aucun provider n'est configuré : email + mot de passe + TOTP (MFA optionnel).

**Configuration** — deux options :

1. **Via le compose** (`CAIRN_OIDC_*`) — verrouille la config.
2. **Via l'admin** — **Admin → Réglages → SSO** si rien n'est défini dans le compose.

**Redirect URI** à déclarer côté provider :

```
<CAIRN_BASE_URL>/api/auth/sso/callback
```

Exemple pour Authentik : issuer `https://auth.example.com/application/o/<slug>/`.

---

## Derrière un reverse proxy

### Traefik + Cloudflare Tunnel

`compose.yaml` inclut les labels Traefik prêts à l'emploi. Adaptez `Host(...)` à votre domaine :

```yaml
labels:
  traefik.enable: "true"
  traefik.http.routers.cairn.rule: "Host(`go.example.com`)"
  traefik.http.routers.cairn.entrypoints: "websecure"
  traefik.http.routers.cairn.tls.certresolver: "cloudflare"
```

Activez `CAIRN_TRUSTED_PROXY=true` pour que les IPs clients soient correctement lues depuis `CF-Connecting-IP`.

### Nginx

```nginx
location / {
    proxy_pass         http://127.0.0.1:8080;
    proxy_set_header   Host              $host;
    proxy_set_header   X-Forwarded-For   $remote_addr;
}
```

---

## Fonds d'écran

### Formats acceptés

| Type | Extensions |
|---|---|
| Image | `.jpg` `.jpeg` `.png` `.webp` `.avif` |
| Vidéo | `.mp4` `.webm` |

Taille max : 50 MB (configurable via `CAIRN_MAX_UPLOAD_SIZE`).

### Thème adaptatif

Cairn échantillonne la luminance du fond d'écran actif et bascule automatiquement entre un thème clair et sombre pour garantir la lisibilité.

---

## Bookmarklet

Sauvegardez la page courante en un clic depuis n'importe quel navigateur, y compris mobile Safari.

1. **Compte → Bookmarklet → Générer un bookmarklet**
2. Copiez le lien et glissez-le dans votre barre de favoris
3. Cliquez le favori sur n'importe quelle page pour la sauvegarder dans Cairn

---

## Moteurs de recherche

| Moteur | Identifiant |
|---|---|
| DuckDuckGo (défaut) | `duckduckgo` |
| Google | `google` |
| Brave Search | `brave` |
| Bing | `bing` |
| Kagi | `kagi` |
| Personnalisé | `custom` + URL se terminant par `=` |

Les bangs DuckDuckGo (`!g`, `!yt`, `!gh`, `!hub`, etc.) fonctionnent quel que soit le moteur configuré.

---

## Développement local

Go n'est pas requis sur la machine hôte — tout passe par Docker.

```bash
# Compiler
docker run --rm \
  -v "$(pwd)":/app \
  -v cairn-gomod-cache:/root/go/pkg/mod \
  -w /app \
  golang:1.23-alpine \
  go build ./...

# Vérifier les vulnérabilités
docker run --rm \
  -v "$(pwd)":/app \
  -v cairn-gomod-cache:/root/go/pkg/mod \
  -w /app \
  golang:1.23-alpine \
  sh -c "go install golang.org/x/vuln/cmd/govulncheck@v1.1.3 && govulncheck ./..."

# Build multi-arch local (sans push)
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --tag cairn:dev \
  .
```

### Mode development

Positionnez `CAIRN_ENV=development` pour obtenir des logs texte au lieu de JSON et autoriser `localhost:3000` dans les origines CORS.

---

## Sécurité

| Point | Implémentation |
|---|---|
| Mots de passe | Argon2id, time=1, memory=64 MB, threads=4 |
| Sessions | SHA-256(token brut), cookie `HttpOnly` + `Secure` + `SameSite=Strict` |
| TOTP | RFC 6238, secret chiffré AES-256-GCM en base |
| Rate limiting | Sliding window par IP hashée (SHA-256), en mémoire |
| Isolation utilisateurs | `userID` vérifié dans chaque requête repository |
| Médias | Servis derrière l'authentification, jamais publics |
| Uploads | Magic bytes validés, nom de fichier généré côté serveur |
| Container | `scratch`, FS read-only, `no-new-privileges`, `CAP_DROP ALL` |
| Headers HTTP | CSP, `X-Frame-Options: DENY`, `Referrer-Policy`, `Permissions-Policy` |

---

## Structure du projet

```
cairn/
├── cmd/cairn/          — point d'entrée (main.go, router, graceful shutdown)
├── internal/
│   ├── config/         — chargement et validation de la configuration
│   ├── db/             — ouverture SQLite + migrations goose embarquées
│   ├── model/          — structs de données
│   ├── repository/     — accès base de données (interfaces + SQLite)
│   ├── service/        — logique métier (auth, bookmarks, wallpapers, admin…)
│   ├── handler/        — handlers HTTP JSON
│   └── middleware/     — auth, admin, rate limit, CORS, headers, bookmarklet
├── web/static/
│   ├── index.html      — shell HTML
│   ├── style.css       — styles (CSS variables, thème adaptatif)
│   └── app.js          — SPA vanilla JS
├── .env.example        — template variables d'environnement
├── Dockerfile          — multi-stage build → image scratch
└── compose.yaml        — déploiement production avec Traefik
```

---

## Licence

MIT
