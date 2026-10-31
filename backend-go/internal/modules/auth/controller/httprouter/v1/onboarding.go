package v1

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/adexcell/crm-ai/internal/infra/http/request"
	"github.com/adexcell/crm-ai/internal/infra/http/response"
	"github.com/adexcell/crm-ai/internal/modules/auth/domain"
	"github.com/adexcell/crm-ai/internal/modules/auth/dto"
)

func (h *Handler) EmailSendCode(w http.ResponseWriter, r *http.Request) {
	var input dto.EmailSendCodeRequest
	if !request.Decode(w, r, &input) {
		return
	}

	err := h.service.EmailSendCode(r.Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			response.Error(w, http.StatusConflict, domain.CodeUserAlreadyExists, err.Error(), nil)
			return
		}
		slog.ErrorContext(r.Context(), "failed to send email code", "error", err)
		response.InternalError(w, "Failed to send code")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Verification code sent successfully",
	})
}

func (h *Handler) EmailVerifyCode(w http.ResponseWriter, r *http.Request) {
	var input dto.EmailVerifyCodeRequest
	if !request.Decode(w, r, &input) {
		return
	}

	err := h.service.EmailVerifyCode(r.Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCode) {
			response.BadRequest(w, domain.CodeInvalidOTP, err.Error())
			return
		}
		if errors.Is(err, domain.ErrCodeNotFound) {
			response.BadRequest(w, domain.CodeOTPNotFound, err.Error())
			return
		}
		slog.ErrorContext(r.Context(), "failed to verify code", "error", err)
		response.InternalError(w, "")
		return
	}

	response.JSON(w, http.StatusOK, map[string]bool{
		"verify": true,
	})
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var input dto.RegisterRequest
	if !request.Decode(w, r, &input) {
		return
	}
	userAgent := r.UserAgent()
	clientIP := r.RemoteAddr

	output, rawRefreshToken, err := h.service.Register(r.Context(), input, userAgent, clientIP)
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			response.Error(w, http.StatusConflict, domain.CodeUserAlreadyExists, err.Error(), nil)
			return
		}
		slog.ErrorContext(r.Context(), "failed to process register", "error", err)
		response.InternalError(w, "Failed to register")
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

	response.JSON(w, http.StatusCreated, output)
}

// Invite — заглушка для приглашения сотрудников по ссылке (будет реализован позже)
func (h *Handler) Invite(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusNotImplemented, map[string]string{
		"message": "Invite flow is not implemented yet",
	})
}
