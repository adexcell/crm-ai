// Package domain
package domain

import (
	"fmt"
	"time"

	"github.com/adexcell/crm-ai/internal/infra/validation"
	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Email        string `validate:"required,email,lte=255"`
	PasswordHash string `validate:"required,lte=255"`
	FirstName    string `validate:"required,lte=100"`
	LastName     string `validate:"required,lte=100"`
	IsActive     bool
	IsSuperuser  bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewUser(email, pwdHash, firstName, lastName string) (User, error) {
	user := User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: pwdHash,
		FirstName:    firstName,
		LastName:     lastName,
		IsActive:     true,
		IsSuperuser:  false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := validation.Validate(user)
	if err != nil {
		return User{}, fmt.Errorf("validating new user: %w", err)
	}

	return user, nil
}

type Organization struct {
	ID        uuid.UUID
	Name      string
	Slug      string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewOrganization(name, slug string) (Organization, error) {
	organization := Organization{
		ID:        uuid.New(),
		Name:      name,
		Slug:      slug,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := validation.Validate(organization)
	if err != nil {
		return Organization{}, fmt.Errorf("validating new organization: %w", err)
	}

	return organization, nil
}

type OrganizationRole string

const (
	RoleOwner   OrganizationRole = "owner"
	RoleAdmin   OrganizationRole = "admin"
	RoleManager OrganizationRole = "manager"
	RoleMember  OrganizationRole = "member"
)

func (r OrganizationRole) IsValid() bool {
	switch r {
	case RoleOwner, RoleAdmin, RoleManager, RoleMember:
		return true
	default:
		return false
	}
}

type OrganizationMember struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	UserID         uuid.UUID
	Role           OrganizationRole
	JoinedAt       time.Time
}

func NewOrganizationMember(orgID, userID uuid.UUID, role OrganizationRole) (OrganizationMember, error) {
	if role == "" {
		role = RoleMember
	}
	if !role.IsValid() {
		return OrganizationMember{}, fmt.Errorf("invalid organization role: %s", role)
	}

	orgMember := OrganizationMember{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         userID,
		Role:           role,
		JoinedAt:       time.Now(),
	}

	err := validation.Validate(orgMember)
	if err != nil {
		return OrganizationMember{}, fmt.Errorf("validating new organization member: %w", err)
	}

	return orgMember, nil
}

type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	UserAgent *string
	IPAddress *string
	ExpiresAt time.Time
	IsRevoked bool
	CreatedAt time.Time
}

func NewRefreshToken(
	userID uuid.UUID,
	token string,
	userAgent,
	IPAddress *string,
	expiresAt time.Time,
) (RefreshToken, error) {
	refresh := RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: token,
		UserAgent: userAgent,
		IPAddress: IPAddress,
		ExpiresAt: expiresAt,
		IsRevoked: false,
		CreatedAt: time.Now(),
	}

	err := validation.Validate(refresh)
	if err != nil {
		return RefreshToken{}, fmt.Errorf("validating new refresh token: %w", err)
	}

	return refresh, nil
}

type UserMembership struct {
	OrganizationID   uuid.UUID
	OrganizationName string
	Slug             string
	Role             OrganizationRole
}

func NewUserMembership(orgID uuid.UUID, orgName, slug string, role OrganizationRole) *UserMembership {
	return &UserMembership{
		OrganizationID:   orgID,
		OrganizationName: orgName,
		Slug:             slug,
		Role:             role,
	}
}
