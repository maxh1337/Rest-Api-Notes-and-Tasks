package entities

import (
	"errors"

	"github.com/google/uuid"
)

var (
	Err2FACodeInvalidOrExpired    = errors.New("2FA code expired")
	Err2FACodeInvalid             = errors.New("2FA code invalid")
	Err2FACodeRevoked             = errors.New("2FA code revoked")
	Err2FASessionAndTokenMismatch = errors.New("provided token mismatch with token from 2FA session data")
	Err2FAAlreadyEnabled          = errors.New("2FA already enabled")
	Err2FADisabled                = errors.New("2FA disabled")
)

type TwoFASessionData struct {
	Code      string              `json:"code"`
	Context   TwoFASessionContext `json:"context"`
	UserID    uuid.UUID           `json:"user_id"`
	Token     string              `json:"token,omitempty"`
	ExpiresAt int64               `json:"expires_at"`
}

type TwoFASessionContext string

const (
	TwoFAContextLogin  TwoFASessionContext = "login"
	TwoFAContextEnable TwoFASessionContext = "enable"
)

type SendCodeRequest struct {
	PhoneNumber string `json:"phone_number"`
	Code        string `json:"code"`
}

type RequestStatus struct {
	RequestID        string  `json:"request_id"`
	PhoneNumber      string  `json:"phone_number"`
	RequestCost      float64 `json:"request_cost"`
	IsRefunded       bool    `json:"is_refunded"`
	RemainingBalance float64 `json:"remaining_balance"`
	DeliveryStatus   DeliveryStatus
}

type DeliveryStatus struct {
	Status    Status `json:"status"`
	UpdatedAt int    `json:"updated_at"`
}

type Status string

const (
	Sent      Status = "sent"
	Delivered Status = "delivered"
	Read      Status = "read"
	Expired   Status = "expired"
	Revoked   Status = "revoked"
)
