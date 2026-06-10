package handler

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/darktweek/cairn/internal/middleware"
	"github.com/darktweek/cairn/internal/model"
	"github.com/darktweek/cairn/internal/service"
)

// ValidateInvite — public: GET /api/auth/invite/{token}
func (h *Handler) ValidateInvite(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	inv, err := h.Invitation.Validate(r.Context(), token)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"email":      inv.Email,
		"expires_at": inv.ExpiresAt.Unix(),
	})
}

// RegisterWithInviteCheck replaces Register with open-registration / invite-token enforcement.
func (h *Handler) RegisterWithInviteCheck(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username    string `json:"username"`
		Email       string `json:"email"`
		Password    string `json:"password"`
		InviteToken string `json:"invite_token"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, fmt.Errorf("%w: invalid JSON", service.ErrInvalidInput))
		return
	}

	open, err := h.Invitation.IsOpenRegistration(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	if !open {
		if body.InviteToken == "" {
			writeError(w, fmt.Errorf("%w: inscription sur invitation uniquement", service.ErrForbidden))
			return
		}
		inv, err := h.Invitation.Consume(r.Context(), body.InviteToken)
		if err != nil {
			writeError(w, err)
			return
		}
		if inv.Email != body.Email {
			writeError(w, fmt.Errorf("%w: email ne correspond pas à l'invitation", service.ErrForbidden))
			return
		}
	}

	user, err := h.User.Register(r.Context(), body.Username, body.Email, body.Password, clientIP(r), r.UserAgent())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"id": user.ID, "username": user.Username,
		"email": user.Email, "role": user.Role,
	})
}

// AdminGetRegistrationSettings — GET /api/admin/settings/registration
func (h *Handler) AdminGetRegistrationSettings(w http.ResponseWriter, r *http.Request) {
	open, err := h.Invitation.IsOpenRegistration(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"open_registration": open})
}

// AdminSetRegistrationSettings — PUT /api/admin/settings/registration
func (h *Handler) AdminSetRegistrationSettings(w http.ResponseWriter, r *http.Request) {
	var body struct {
		OpenRegistration bool `json:"open_registration"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, fmt.Errorf("%w: invalid JSON", service.ErrInvalidInput))
		return
	}
	if err := h.Invitation.SetOpenRegistration(r.Context(), body.OpenRegistration); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"open_registration": body.OpenRegistration})
}

// AdminGetMenuSettings — GET /api/admin/settings/menu
func (h *Handler) AdminGetMenuSettings(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"menu_bang": h.Settings.MenuBang(r.Context()),
		"locked":    h.Settings.MenuBangLocked(),
	})
}

// AdminSetMenuSettings — PUT /api/admin/settings/menu
func (h *Handler) AdminSetMenuSettings(w http.ResponseWriter, r *http.Request) {
	var body struct {
		MenuBang string `json:"menu_bang"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, fmt.Errorf("%w: invalid JSON", service.ErrInvalidInput))
		return
	}
	if err := h.Settings.SetMenuBang(r.Context(), body.MenuBang); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"menu_bang": h.Settings.MenuBang(r.Context()),
		"locked":    h.Settings.MenuBangLocked(),
	})
}

// AdminCreateInvitation — POST /api/admin/invitations
func (h *Handler) AdminCreateInvitation(w http.ResponseWriter, r *http.Request) {
	admin := middleware.UserFromCtx(r.Context())
	var body struct {
		Email string `json:"email"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, fmt.Errorf("%w: invalid JSON", service.ErrInvalidInput))
		return
	}
	inv, _, err := h.Invitation.Create(r.Context(), admin.ID, body.Email)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, fmtInvitation(inv))
}

// AdminListInvitations — GET /api/admin/invitations
func (h *Handler) AdminListInvitations(w http.ResponseWriter, r *http.Request) {
	invs, err := h.Invitation.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]map[string]any, 0, len(invs))
	for _, inv := range invs {
		out = append(out, fmtInvitation(inv))
	}
	writeJSON(w, http.StatusOK, out)
}

// AdminRevokeInvitation — DELETE /api/admin/invitations/{id}
func (h *Handler) AdminRevokeInvitation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.Invitation.Revoke(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// AdminResendInvitation — POST /api/admin/invitations/{id}/resend
func (h *Handler) AdminResendInvitation(w http.ResponseWriter, r *http.Request) {
	admin := middleware.UserFromCtx(r.Context())
	id := chi.URLParam(r, "id")
	if _, err := h.Invitation.Resend(r.Context(), id, admin.ID); err != nil {
		writeError(w, err)
		return
	}
	// Return updated invitation.
	invs, err := h.Invitation.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	for _, inv := range invs {
		if inv.ID == id {
			writeJSON(w, http.StatusOK, fmtInvitation(inv))
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func fmtInvitation(inv *model.Invitation) map[string]any {
	out := map[string]any{
		"id":         inv.ID,
		"email":      inv.Email,
		"created_by": inv.CreatedBy,
		"expires_at": inv.ExpiresAt.Unix(),
		"created_at": inv.CreatedAt.Unix(),
		"used":       inv.IsUsed(),
		"expired":    inv.IsExpired(),
		"pending":    inv.IsPending(),
	}
	if inv.UsedAt != nil {
		out["used_at"] = inv.UsedAt.Unix()
	}
	return out
}
