package domain

import "errors"

// Доменные ошибки уровня бизнес-логики.
var (
	ErrUserAlreadyExists   = errors.New("user with this email already exists")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrUserInactive        = errors.New("user account is deactivated")
	ErrInvalidToken        = errors.New("invalid or expired token")
	ErrInvalidHashFormat   = errors.New("invalid hash format")
	ErrTokenReplayDetected = errors.New("security alert: refresh token reuse detected, session terminated")
	ErrNotFound            = errors.New("resource not found")
	ErrCodeNotFound        = errors.New("verification code not found or expired")
	ErrInvalidCode         = errors.New("invalid verification code")
	ErrInvalidInput        = errors.New("invalid input")
	ErrForbidden           = errors.New("access denied: forbidden")
)

// Машиночитаемые строковые коды для API Envelope
const (
	CodeUserAlreadyExists   = "USER_ALREADY_EXISTS"
	CodeInvalidCredentials  = "INVALID_CREDENTIALS"
	CodeUserInactive        = "USER_INACTIVE"
	CodeInvalidToken        = "INVALID_TOKEN"
	CodeTokenReplayDetected = "TOKEN_REPLAY_DETECTED"
	CodeNotFound            = "RESOURCE_NOT_FOUND"
	CodeOTPNotFound         = "OTP_NOT_FOUND"
	CodeInvalidOTP          = "INVALID_OTP"
	CodeInvalidInput        = "INVALID_JSON"
	CodeForbidden           = "FORBIDDEN"
)
