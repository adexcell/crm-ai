package dto

import (
	"time"

	"github.com/adexcell/crm-ai/internal/modules/auth/domain"
	"github.com/google/uuid"
)

type RegisterRequest struct {
	Email            string `json:"email"             validate:"required,email"`
	Password         string `json:"password"          validate:"required,gte=8,lte=128"`
	FirstName        string `json:"first_name"        validate:"required,gte=1,lte=100"`
	LastName         string `json:"last_name"         validate:"required,gte=1,lte=100"`
	OrganizationName string `json:"organization_name" validate:"required,gte=2,lte=100"`
}

type AuthResponse struct {
	User   UserResponse  `json:"user"`
	Tokens TokenResponse `json:"tokens"`
}

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,gte=8,lte=128"`
}

type UserResponse struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	IsActive    bool      `json:"is_active"`
	IsSuperuser bool      `json:"is_superuser"`
	CreatedAt   time.Time `json:"created_at"`
}

type OrganizationResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type OrganizationMemberResponse struct {
	ID             uuid.UUID               `json:"id"`
	OrganizationID uuid.UUID               `json:"organization_id"`
	UserID         uuid.UUID               `json:"user_id"`
	Role           domain.OrganizationRole `json:"role"`
	JoinedAt       time.Time               `json:"joined_at"`
}

type TokenResponse struct {
	AccessToken     string `json:"access_token"`
	TokenType       string `json:"token_type" default:"bearer"`
	ExpiresIn       int64  `json:"expires_in"`
	RawRefreshToken string `json:"-"`
}

type EmailSendCodeRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type EmailVerifyCodeRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,len=6"`
}

type SessionResponse struct {
	ID        uuid.UUID `json:"id"`
	UserAgent string    `json:"user_agent"`
	IPAddress string    `json:"ip_address"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	IsCurrent bool      `json:"is_current"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password" validate:"required,gte=8,lte=128"`
}

type UpdateProfileRequest struct {
	FirstName string `json:"first_name" validate:"required,gte=1,lte=100"`
	LastName  string `json:"last_name"  validate:"required,gte=1,lte=100"`
}

type SwitchOrgRequest struct {
	OrganizationID uuid.UUID `json:"organization_id" validate:"required"`
}

type UserMembershipResponse struct {
	OrganizationID   uuid.UUID               `json:"organization_id"`
	OrganizationName string                  `json:"organization_name"`
	Slug             string                  `json:"slug"`
	Role             domain.OrganizationRole `json:"role"`
}

type ProfileResponse struct {
	User                UserResponse             `json:"user"`
	CurrentOrganization *UserMembershipResponse  `json:"current_organization"`
	Memberships         []UserMembershipResponse `json:"memberships"`
}
