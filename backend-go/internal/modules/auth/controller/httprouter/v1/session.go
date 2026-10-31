package v1

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/adexcell/crm-ai/internal/infra/http/middleware"
	"github.com/adexcell/crm-ai/internal/infra/http/request"
	"github.com/adexcell/crm-ai/internal/infra/http/response"
	"github.com/adexcell/crm-ai/internal/modules/auth/domain"
	"github.com/adexcell/crm-ai/internal/modules/auth/dto"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var input dto.LoginRequest
	if !request.Decode(w, r, &input) {
		return
	}

	userAgent := r.UserAgent()
	clientIP := r.RemoteAddr

	authResp, rawRefreshToken, err := h.service.Login(r.Context(), input, userAgent, clientIP)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			response.Error(w, http.StatusUnauthorized, domain.CodeInvalidCredentials, "Invalid email or password", nil)
			return
		}
		if errors.Is(err, domain.ErrUserInactive) {
			response.Error(w, http.StatusForbidden, domain.CodeUserInactive, "User account is deactivated", nil)
			return
		}
		response.InternalError(w, "Failed to login")
		return
	}

	refreshDays := h.cfg.JWT.RefreshTokenExpireDays
	maxAgeSeconds := refreshDays * 24 * 60 * 60
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    rawRefreshToken,
		Path:     "/api/v1/auth",
		HttpOnly: true,
		Secure:   h.cfg.App.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   maxAgeSeconds,
		Expires:  time.Now().Add(time.Duration(maxAgeSeconds) * time.Second),
	})

	response.JSON(w, http.StatusOK, authResp)
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil || cookie.Value == "" {
		response.Unauthorized(w, "Refresh token is missing")
		return
	}

	userAgent := r.UserAgent()
	clientIP := r.RemoteAddr

	tokenResp, newRawRefreshToken, err := h.service.RefreshToken(r.Context(), cookie.Value, userAgent, clientIP)
	if err != nil {
		if errors.Is(err, domain.ErrTokenReplayDetected) {
			http.SetCookie(w, &http.Cookie{
				Name:     "refresh_token",
				Value:    "",
				Path:     "/api/v1/auth",
				HttpOnly: true,
				MaxAge:   -1,
			})
			slog.WarnContext(r.Context(), "security alert: refresh token reuse detected",
				"ip", clientIP,
				"user_agent", userAgent,
			)
			response.Error(w, http.StatusUnauthorized, domain.CodeTokenReplayDetected, "Session compromised, all sessions revoked", nil)
			return
		}
		if errors.Is(err, domain.ErrInvalidToken) {
			http.SetCookie(w, &http.Cookie{
				Name:     "refresh_token",
				Value:    "",
				Path:     "/api/v1/auth",
				HttpOnly: true,
				MaxAge:   -1,
			})
			response.Unauthorized(w, "Session expired, please login again")
			return
		}
		if errors.Is(err, domain.ErrUserInactive) {
			response.Error(w, http.StatusForbidden, domain.CodeUserInactive, "User account is deactivated", nil)
			return
		}
		slog.ErrorContext(r.Context(), "failed to refresh token", "err", err)
		response.InternalError(w, "Failed to refresh token")
		return
	}

	refreshDays := h.cfg.JWT.RefreshTokenExpireDays
	maxAgeSeconds := refreshDays * 24 * 60 * 60

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    newRawRefreshToken,
		Path:     "/api/v1/auth",
		HttpOnly: true,
		Secure:   h.cfg.App.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   maxAgeSeconds,
		Expires:  time.Now().Add(time.Duration(maxAgeSeconds) * time.Second),
	})

	response.JSON(w, http.StatusOK, tokenResp)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		slog.DebugContext(r.Context(), "r.Cookie error", "err", err)
	}
	if cookie != nil && cookie.Value != "" {
		err = h.service.Logout(r.Context(), cookie.Value)
		if err != nil {
			slog.DebugContext(r.Context(), "service.Logout()", "err", err)
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/v1/auth",
		HttpOnly: true,
		MaxAge:   -1,
	})

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Logged out successfully",
	})
}

func (h *Handler) GetSessions(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Unathorized")
		return
	}

	rawRefreshToken := ""
	if cookie, err := r.Cookie("refresh_token"); err == nil {
		rawRefreshToken = cookie.Value
	}

	sessions, err := h.service.GetSessions(r.Context(), userID, rawRefreshToken)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to get sessions", "error", err)
		response.InternalError(w, "Failed to get sessions")
		return
	}

	response.JSON(w, http.StatusOK, sessions)
}

func (h *Handler) RevokeAllSessions(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Unauthorized")
		return
	}
	if err := h.service.RevokeAllSessions(r.Context(), userID); err != nil {
		slog.ErrorContext(r.Context(), "failed to revoke all sessions", "error", err)
		response.InternalError(w, "Failed to revoke all sessions")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/v1/auth",
		HttpOnly: true,
		MaxAge:   -1,
	})
	response.JSON(w, http.StatusOK, map[string]string{
		"message": "All sessions revoked successfully",
	})
}

func (h *Handler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	sessionIDStr := chi.URLParam(r, "id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_SESSION_ID", "Invalid session UUID format", nil)
		return
	}
	if err := h.service.RevokeSession(r.Context(), userID, sessionID); err != nil {
		slog.ErrorContext(r.Context(), "failed to revoke session", "error", err)
		response.InternalError(w, "Failed to revoke session")
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Session revoked successfully",
	})
}
