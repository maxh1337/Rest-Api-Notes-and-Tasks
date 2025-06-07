package entities

const (
	// Auth errors
	ErrorCodeInvalidCredentials  = "INVALID_CREDENTIALS"
	ErrorCodeTokenCreationFailed = "TOKEN_CREATION_FAILED"
	ErrorCodeRefreshTokenMissing = "REFRESH_TOKEN_MISSING"
	ErrorCodeSessionIDMissing    = "SESSION_ID_MISSING"
	ErrorCode2FARequired         = "2FA_REQUIRED"
	ErrorCodeTokensMismatch      = "TOKENS_MISMATCH"

	// 2FA errors
	ErrorCode2FACodeExpired    = "2FA_CODE_EXPIRED"
	ErrorCode2FACodeInvalid    = "2FA_CODE_INVALID"
	ErrorCode2FACodeRevoked    = "2FA_CODE_REVOKED"
	ErrorCode2FATokenMismatch  = "2FA_TOKEN_MISMATCH"
	ErrorCode2FAAlreadyEnabled = "2FA_ALREADY_ENABLED"

	// Session errors
	ErrorCodeSessionWrongDevice = "SESSION_WRONG_DEVICE"
	ErrorCodeSessionExpired     = "SESSION_EXPIRED"
	ErrorCodeSessionNotFound    = "SESSION_NOT_FOUND"
	ErrorCodeInvalidSessionID   = "INVALID_SESSION_ID"

	// User errors
	ErrorCodeEmailTaken         = "EMAIL_ALREADY_TAKEN"
	ErrorCodeUsernameTaken      = "USERNAME_ALREADY_TAKEN"
	ErrorCodeUserNotFound       = "USER_NOT_FOUND"
	ErrorCodeInvalidEmailFormat = "INVALID_EMAIL_FORMAT"
	ErrorCodeCantChangePhone2FA = "CANT_CHANGE_PHONE_2FA"

	// General errors
	ErrorCodeValidationFailed = "VALIDATION_FAILED"
	ErrorCodeInternalError    = "INTERNAL_ERROR"
	ErrorCodeUnauthorized     = "UNAUTHORIZED"
	ErrorCodeForbidden        = "FORBIDDEN"
	ErrorCodeInvalidInput     = "INVALID_INPUT"
)

type APIError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (e APIError) Error() string {
	return e.Message
}

func NewAPIError(code, message string, details ...map[string]interface{}) *APIError {
	err := &APIError{
		Code:    code,
		Message: message,
	}
	if len(details) > 0 {
		err.Details = details[0]
	}
	return err
}

type ValidationError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields"`
}

func (e ValidationError) Error() string {
	return e.Message
}

func NewValidationError(fields map[string]string) *ValidationError {
	return &ValidationError{
		Code:    ErrorCodeValidationFailed,
		Message: "Validation failed",
		Fields:  fields,
	}
}

type TwoFactorRequiredError struct {
	Code           string `json:"code"`
	Message        string `json:"message"`
	TwoFactorToken string `json:"two_factor_token,omitempty"`
	Required       bool   `json:"required"`
}

func (e TwoFactorRequiredError) Error() string {
	return e.Message
}

func NewTwoFactorRequiredError() *TwoFactorRequiredError {
	return &TwoFactorRequiredError{
		Code:     ErrorCode2FARequired,
		Message:  "Two-factor authentication required",
		Required: true,
	}
}

var ErrorMapper = map[error]*APIError{
	// Auth errors
	ErrInvalidCredentials:         NewAPIError(ErrorCodeInvalidCredentials, "Invalid username/email or password"),
	ErrFailedToCreateAccessToken:  NewAPIError(ErrorCodeTokenCreationFailed, "Failed to create access token"),
	ErrFailedToCreateRefreshToken: NewAPIError(ErrorCodeTokenCreationFailed, "Failed to create refresh token"),
	ErrRefreshTokenNotProvided:    NewAPIError(ErrorCodeRefreshTokenMissing, "Refresh token not provided"),
	ErrSessionIDTokenNotProvided:  NewAPIError(ErrorCodeSessionIDMissing, "Session ID not provided"),
	ErrTokensMismatch:             NewAPIError(ErrorCodeTokensMismatch, "Token mismatch detected"),

	// 2FA errors
	Err2FACodeInvalidOrExpired:    NewAPIError(ErrorCode2FACodeExpired, "2FA code has expired"),
	Err2FACodeInvalid:             NewAPIError(ErrorCode2FACodeInvalid, "Invalid 2FA code"),
	Err2FACodeRevoked:             NewAPIError(ErrorCode2FACodeRevoked, "2FA code has been revoked"),
	Err2FASessionAndTokenMismatch: NewAPIError(ErrorCode2FATokenMismatch, "2FA token mismatch"),
	Err2FAAlreadyEnabled:          NewAPIError(ErrorCode2FAAlreadyEnabled, "Two-factor authentication is already enabled"),
	Err2FADisabled:                NewAPIError(ErrorCode2FAAlreadyEnabled, "Two-factor authentication is disabled"),

	// Session errors
	ErrSessionBelongsToAnotherDevice: NewAPIError(ErrorCodeSessionWrongDevice, "Session belongs to another device"),
	ErrSessionExpired:                NewAPIError(ErrorCodeSessionExpired, "Session has expired, please login again"),
	ErrSessionNotFound:               NewAPIError(ErrorCodeSessionNotFound, "Session not found"),
	ErrInvalidSessionID:              NewAPIError(ErrorCodeInvalidSessionID, "Invalid session ID"),

	// User errors
	ErrEmailAlreadyTaken:    NewAPIError(ErrorCodeEmailTaken, "Email address is already taken"),
	ErrUsernameAlreadyTaken: NewAPIError(ErrorCodeUsernameTaken, "Username is already taken"),
	ErrUserNotFound:         NewAPIError(ErrorCodeUserNotFound, "User not found"),
	ErrInvalidEmailFormat:   NewAPIError(ErrorCodeInvalidEmailFormat, "Invalid email format"),
	ErrCantChangePhone2FA:   NewAPIError(ErrorCodeCantChangePhone2FA, "Cannot change phone number while 2FA is active"),
}

func ConvertError(err error) error {
	if err == nil {
		return nil
	}

	if apiErr, ok := err.(*APIError); ok {
		return apiErr
	}

	if validationErr, ok := err.(*ValidationError); ok {
		return validationErr
	}

	if twoFactorErr, ok := err.(*TwoFactorRequiredError); ok {
		return twoFactorErr
	}

	if apiErr, exists := ErrorMapper[err]; exists {
		return &APIError{
			Code:    apiErr.Code,
			Message: apiErr.Message,
			Details: apiErr.Details,
		}
	}

	return NewAPIError(ErrorCodeInternalError, "Internal server error")
}

func IsErrorCode(err error, code string) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Code == code
	}
	return false
}
