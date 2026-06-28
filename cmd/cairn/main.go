package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/darktweek/cairn/internal/config"
	"github.com/darktweek/cairn/internal/db"
	"github.com/darktweek/cairn/internal/handler"
	"github.com/darktweek/cairn/internal/middleware"
	"github.com/darktweek/cairn/internal/model"
	"github.com/darktweek/cairn/internal/repository"
	"github.com/darktweek/cairn/internal/service"
	cairnweb "github.com/darktweek/cairn/web"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		if err := healthcheck(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	setupLogger(cfg.Env)

	slog.Info("starting cairn", "version", version, "addr", cfg.Addr, "env", cfg.Env)

	database, err := db.Open(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("db open: %w", err)
	}
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		return fmt.Errorf("db migrate: %w", err)
	}
	slog.Info("migrations applied")

	repos := repository.New(database)
	svcs := service.New(repos, cfg)
	h := handler.New(svcs, strings.HasPrefix(cfg.BaseURL, "https://"), cfg.SessionLifetimeDays)

	// Background: purge expired sessions and pending registrations every hour.
	go func() {
		t := time.NewTicker(time.Hour)
		for range t.C {
			if err := repos.Sessions.DeleteExpired(context.Background()); err != nil {
				slog.Error("purge sessions", "err", err)
			}
			if err := repos.PendingRegistrations.DeleteExpired(context.Background()); err != nil {
				slog.Error("purge pending registrations", "err", err)
			}
		}
	}()

	r := buildRouter(cfg, h, svcs)

	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", cfg.Addr)
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case sig := <-quit:
		slog.Info("shutdown signal received", "signal", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	slog.Info("cairn stopped")
	return nil
}

func buildRouter(cfg *config.Config, h *handler.Handler, svcs *service.Services) http.Handler {
	r := chi.NewRouter()

	// Global middleware.
	r.Use(chimw.RequestID)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(middleware.SecureHeaders)
	r.Use(middleware.CORS(cfg.BaseURL, cfg.Env))
	// Global body cap. JSON APIs are capped tight (1 MB); the two routes that
	// legitimately accept large bodies — wallpaper upload and bookmark import —
	// are excluded here and set their own (bigger) limit further down.
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			largeUpload := req.Method == http.MethodPost &&
				(req.URL.Path == "/api/wallpapers" || req.URL.Path == "/api/bookmarks/import")
			if !largeUpload {
				req.Body = http.MaxBytesReader(w, req.Body, 1<<20) // 1 MB for JSON
			}
			next.ServeHTTP(w, req)
		})
	})

	// Rate limit configs.
	// IP-level is only a coarse spray guard: behind Docker NAT every client
	// shares one bucket. The real brute-force lock is per-account in
	// AuthService.Login (10/5min per identifier).
	loginRL := middleware.RateLimit(
		middleware.RateLimitConfig{Max: 30, Window: 5 * time.Minute},
		cfg.TrustedProxy,
	)
	registerRL := middleware.RateLimit(
		middleware.RateLimitConfig{Max: 3, Window: time.Hour},
		cfg.TrustedProxy,
	)
	forgotRL := middleware.RateLimit(
		middleware.RateLimitConfig{Max: 3, Window: time.Hour},
		cfg.TrustedProxy,
	)

	// Healthcheck — public, intentionally not logged.
	r.Get("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","version":%q}`, version)
	})

	// Auth routes — public with rate limiting.
	// Public: validate invite token (legacy).
	r.Get("/api/auth/invite/{token}", h.ValidateInvite)
	// Public: invite-based setup flow.
	r.Post("/api/auth/invite/{token}/prepare", h.PrepareInviteSetup)
	r.Post("/api/auth/invite/{token}/complete", h.CompleteInviteSetup)

	// Public: open-registration setup flow.
	r.Post("/api/auth/request-registration", h.RequestRegistration)
	r.Get("/api/auth/setup", h.ValidateSetupToken)
	r.Post("/api/auth/complete-setup", h.CompleteSetup)

	// Public: open registration flag (used by login page to show/hide register link).
	r.Get("/api/auth/config", h.PublicAuthConfig)

	// Public: SSO / OIDC.
	r.Get("/api/auth/sso/config", h.SSOConfig)
	r.Get("/api/auth/sso/login", h.SSOLogin)
	r.Get("/api/auth/sso/callback", h.SSOCallback)

	r.Group(func(r chi.Router) {
		r.Post("/api/auth/register", func(w http.ResponseWriter, req *http.Request) {
			registerRL(http.HandlerFunc(h.RegisterWithInviteCheck)).ServeHTTP(w, req)
		})
		r.Post("/api/auth/login", func(w http.ResponseWriter, req *http.Request) {
			loginRL(http.HandlerFunc(h.Login)).ServeHTTP(w, req)
		})
		r.Post("/api/auth/forgot-password", func(w http.ResponseWriter, req *http.Request) {
			forgotRL(http.HandlerFunc(h.ForgotPassword)).ServeHTTP(w, req)
		})
		r.Post("/api/auth/reset-password", h.ResetPassword)
	})

	// Bookmarklet quick-save — special auth from body token.
	r.With(middleware.BookmarkletAuth(svcs.Auth)).
		Post("/api/bookmarks/quick", h.QuickBookmark)

	// Authenticated routes.
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(svcs.Auth))

		r.Post("/api/auth/logout", h.Logout)

		// Me
		r.Get("/api/me", h.GetMe)
		r.Put("/api/me", h.UpdateProfile)
		r.Put("/api/me/password", h.ChangePassword)
		r.Put("/api/me/locale", h.UpdateLocale)
		r.Put("/api/me/search-engine", h.UpdateSearchEngine)
		r.Get("/api/me/sessions", h.ListSessions)
		r.Delete("/api/me/sessions/{id}", h.RevokeSession)
		r.Delete("/api/me/sessions", h.RevokeAllSessions)
		r.Get("/api/me/audit", h.GetMyAuditLog)
		r.Get("/api/me/stats", h.GetMyStats)
		r.Get("/api/me/prefs", h.GetMyPrefs)
		r.Put("/api/me/prefs", h.SetMyPrefs)
		r.Delete("/api/me", h.DeleteAccount)

		// TOTP
		r.Post("/api/me/totp", h.BeginTOTP)
		r.Put("/api/me/totp", h.ConfirmTOTP)
		r.Delete("/api/me/totp", h.DisableTOTP)

		// Bookmarklet
		r.Get("/api/me/bookmarklet", h.GetBookmarklet)
		r.Delete("/api/me/bookmarklet", h.RevokeBookmarklet)

		// Bookmarks
		r.Get("/api/bookmarks", h.ListBookmarks)
		r.Post("/api/bookmarks", h.CreateBookmark)
		r.Get("/api/bookmarks/{id}", h.GetBookmark)
		r.Put("/api/bookmarks/{id}", h.UpdateBookmark)
		r.Delete("/api/bookmarks/{id}", h.DeleteBookmark)
		r.Put("/api/bookmarks/sort", h.UpdateBookmarkSort)
		r.With(middleware.BodyLimit(10<<20)).Post("/api/bookmarks/import", h.ImportBookmarks)
		r.Get("/api/bookmarks/export", h.ExportBookmarks)

		// Collections & folders
		r.Get("/api/collections", h.ListCollections)
		r.Post("/api/collections", h.CreateCollection)
		r.Get("/api/collections/{id}", h.GetCollection)
		r.Put("/api/collections/{id}", h.UpdateCollection)
		r.Delete("/api/collections/{id}", h.DeleteCollection)
		r.Get("/api/collections/{id}/folders", h.ListCollectionFolders)
		r.Post("/api/collections/{id}/folders", h.CreateCollectionFolder)
		r.Get("/api/collections/{id}/shares", h.ListCollectionShares)
		r.Post("/api/collections/{id}/shares", h.SetCollectionShare)
		r.Delete("/api/collections/{id}/shares/{userId}", h.RemoveCollectionShare)
		r.Post("/api/collections/{id}/group-shares", h.SetCollectionGroupShare)
		r.Delete("/api/collections/{id}/group-shares/{groupId}", h.RemoveCollectionGroupShare)
		r.Put("/api/folders/{id}", h.UpdateFolder)
		r.Delete("/api/folders/{id}", h.DeleteFolder)
		r.Get("/api/users/search", h.SearchUsers)

		// Groups: listing is open to authenticated users (share picker);
		// mutations require the groups.manage permission.
		r.Get("/api/groups", h.ListGroups)
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(model.PermGroupsManage))
			r.Get("/api/admin/groups", h.AdminListGroups)
			r.Post("/api/groups", h.CreateGroup)
			r.Put("/api/groups/{id}", h.UpdateGroup)
			r.Delete("/api/groups/{id}", h.DeleteGroup)
			r.Get("/api/groups/{id}/members", h.ListGroupMembers)
			r.Post("/api/groups/{id}/members", h.AddGroupMember)
			r.Delete("/api/groups/{id}/members/{userId}", h.RemoveGroupMember)
		})

		// Tags
		r.Get("/api/tags", h.ListTags)
		r.Delete("/api/tags/{id}", h.DeleteTag)

		// Wallpapers
		r.Get("/api/wallpapers", h.ListWallpapers)
		r.With(middleware.UserBodyLimit(cfg.MaxUploadSize)).Post("/api/wallpapers", h.UploadWallpaper)
		r.Delete("/api/wallpapers/{id}", h.DeleteWallpaper)
		r.Put("/api/wallpapers/{id}/pin", h.SetWallpaperPinned)
		r.Put("/api/wallpapers/sort", h.UpdateWallpaperSort)

		// Media — served from filesystem.
		r.Get("/media/{userID}/{filename}", h.ServeMedia)

		// Admin — coarse guard: reachable by anyone holding an admin-area permission.
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAnyPermission(
				model.PermUsersManage, model.PermSettingsManage, model.PermAuditView,
				model.PermRolesManage, model.PermGroupsManage,
			))

			r.Get("/api/admin/stats", h.AdminGetStats)
			r.Get("/api/admin/permissions", h.ListPermissions)

			// User management.
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequirePermission(model.PermUsersManage))
				r.Get("/api/admin/users", h.AdminListUsers)
				r.Get("/api/admin/users/{id}", h.AdminGetUser)
				r.Get("/api/admin/users/{id}/stats", h.AdminGetUserStats)
				r.Put("/api/admin/users/{id}/suspend", h.AdminSuspendUser)
				r.Put("/api/admin/users/{id}/activate", h.AdminActivateUser)
				r.Delete("/api/admin/users/{id}", h.AdminDeleteUser)
				r.Put("/api/admin/users/{id}/wallpaper-limit", h.AdminSetWallpaperLimit)
				r.Put("/api/admin/users/{id}/upload-size-limit", h.AdminSetUploadSizeLimit)
				r.Put("/api/admin/users/{id}/storage-quota", h.AdminSetStorageQuota)
				r.Put("/api/admin/users/{id}/roles", h.SetUserRoles)
				r.Get("/api/admin/pending-registrations", h.AdminListPendingRegistrations)
				r.Delete("/api/admin/pending-registrations/{id}", h.AdminRevokePendingRegistration)
				r.Get("/api/admin/invitations", h.AdminListInvitations)
				r.Post("/api/admin/invitations", h.AdminCreateInvitation)
				r.Delete("/api/admin/invitations/{id}", h.AdminRevokeInvitation)
				r.Post("/api/admin/invitations/{id}/resend", h.AdminResendInvitation)
			})

			// Audit log.
			r.With(middleware.RequirePermission(model.PermAuditView)).
				Get("/api/admin/audit", h.AdminGetAuditLog)

			// Roles (RBAC). Listing is also available to user-managers so they
			// can populate the role picker; mutations require roles.manage.
			r.With(middleware.RequireAnyPermission(model.PermRolesManage, model.PermUsersManage)).
				Get("/api/admin/roles", h.ListRoles)
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequirePermission(model.PermRolesManage))
				r.Post("/api/admin/roles", h.CreateRole)
				r.Put("/api/admin/roles/{id}", h.UpdateRole)
				r.Delete("/api/admin/roles/{id}", h.DeleteRole)
			})

			// Instance settings.
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequirePermission(model.PermSettingsManage))
				r.Get("/api/admin/settings/registration", h.AdminGetRegistrationSettings)
				r.Put("/api/admin/settings/registration", h.AdminSetRegistrationSettings)
				r.Get("/api/admin/settings/menu", h.AdminGetMenuSettings)
				r.Put("/api/admin/settings/menu", h.AdminSetMenuSettings)
				r.Get("/api/admin/settings/sso", h.AdminGetSSOSettings)
				r.Put("/api/admin/settings/sso", h.AdminSetSSOSettings)
				r.Get("/api/admin/settings/system", h.AdminGetSystemSettings)
				r.Put("/api/admin/settings/system", h.AdminSetSystemSettings)
				r.Post("/api/admin/settings/smtp/test", h.AdminTestSMTP)
				r.Get("/api/admin/settings/policies", h.GetPolicies)
				r.Put("/api/admin/settings/policies", h.SetPolicy)
			})
		})
	})

	// Static assets embedded in the binary — served directly by name to avoid
	// http.FileServer URL path matching issues with chi's router.
	static, _ := fs.Sub(cairnweb.Static, "static")
	r.Get("/style.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, static, "style.css")
	})
	r.Get("/app.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, static, "app.js")
	})
	r.Get("/i18n.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, static, "i18n.js")
	})
	// SPA fallback: all unknown routes serve index.html.
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, static, "index.html")
	})

	return r
}

func setupLogger(env string) {
	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}

	if env == "development" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))
}

func healthcheck() error {
	resp, err := http.Get("http://localhost:8080/healthcheck")
	if err != nil {
		return fmt.Errorf("healthcheck request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("healthcheck status: %d", resp.StatusCode)
	}
	return nil
}
