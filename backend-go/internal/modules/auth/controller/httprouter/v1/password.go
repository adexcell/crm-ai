package v1

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/adexcell/crm-ai/internal/infra/http/request"
	"github.com/adexcell/crm-ai/internal/infra/http/response"
	"github.com/adexcell/crm-ai/internal/modules/auth/domain"
	"github.com/adexcell/crm-ai/internal/modules/auth/dto"
)

func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var input dto.ForgotPasswordRequest
	if !request.Decode(w, r, &input) {
		return
	}

	err := h.service.ForgotPassword(r.Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.JSON(w, http.StatusOK, map[string]string{
				"message": "Если аккаунт с таким email существует, мы отправили инструкцию по сбросу пароля",
			})
			return
		}
		slog.ErrorContext(r.Context(), "failed to process forgot password", "error", err)
		response.InternalError(w, "")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Если аккаунт с таким email существует, мы отправили инструкцию по сбросу пароля",
	})

}

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var input dto.ResetPasswordRequest
	if !request.Decode(w, r, &input) {
		return
	}
	if err := h.service.ResetPassword(r.Context(), input); err != nil {
		if errors.Is(err, domain.ErrInvalidToken) || errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusBadRequest, domain.CodeInvalidToken, "Invalid or expired reset token", nil)
			return
		}
		slog.ErrorContext(r.Context(), "failed to process reset password", "error", err)
		response.InternalError(w, "Failed to reset password")
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Password has been reset successfully. Please login with your new password.",
	})
}
