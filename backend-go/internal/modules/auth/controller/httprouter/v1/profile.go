package v1

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/adexcell/crm-ai/internal/infra/http/middleware"
	"github.com/adexcell/crm-ai/internal/infra/http/request"
	"github.com/adexcell/crm-ai/internal/infra/http/response"
	"github.com/adexcell/crm-ai/internal/modules/auth/domain"
	"github.com/adexcell/crm-ai/internal/modules/auth/dto"
)

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	userID, currentOrgID := claims.UserID, claims.OrgID

	profile, err := h.service.GetProfile(r.Context(), userID, currentOrgID)
	if err != nil {
		if errors.Is(err, domain.ErrUserInactive) {
			response.Error(w, http.StatusForbidden, domain.CodeUserInactive, "User account is deactivated", nil)
			return
		}
		if errors.Is(err, domain.ErrNotFound) {
			response.NotFound(w, "User not found")
			return
		}
		slog.ErrorContext(r.Context(), "failed to get profile", "error", err)
		response.InternalError(w, "Failed to get profile")
		return
	}

	response.JSON(w, http.StatusOK, profile)
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	var input dto.UpdateProfileRequest
	if !request.Decode(w, r, &input) {
		return
	}

	userResp, err := h.service.UpdateProfile(r.Context(), userID, input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			response.BadRequest(w, domain.CodeInvalidInput, "invalid input")
			return
		}
		if errors.Is(err, domain.ErrUserInactive) {
			response.Error(w, http.StatusForbidden, domain.CodeUserInactive, "User account is deactivated", nil)
			return
		}
		if errors.Is(err, domain.ErrNotFound) {
			response.NotFound(w, "User not found")
			return
		}
		slog.ErrorContext(r.Context(), "failed to update profile", "error", err)
		response.InternalError(w, "Failed to update profile")
		return
	}

	response.JSON(w, http.StatusOK, userResp)
}

func (h *Handler) SwitchOrg(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	userID := claims.UserID

	var input dto.SwitchOrgRequest
	if !request.Decode(w, r, &input) {
		return
	}

	token, err := h.service.SwitchOrganization(r.Context(), userID, input.OrganizationID)
	if err != nil {
		if errors.Is(err, domain.ErrUserInactive) {
			response.Error(w, http.StatusForbidden, domain.CodeUserInactive, "User account is deactivated", nil)
			return
		}
		if errors.Is(err, domain.ErrNotFound) {
			response.NotFound(w, "User not found")
			return
		}
		if errors.Is(err, domain.ErrForbidden) {
			slog.WarnContext(r.Context(), "security alert: unauthorized organization switch attempt",
				"user_id", userID,
				"target_org_id", input.OrganizationID,
			)
			response.Forbidden(w, "Access denied: you are not a member of this organization")
			return
		}
		slog.ErrorContext(r.Context(), "failed to switch organization", "error", err)
		response.InternalError(w, "Failed to switch organization")
		return
	}

	response.JSON(w, http.StatusOK, token)
}
