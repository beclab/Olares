package wizard

import "fmt"

// AuthError represents authentication errors
type AuthError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Data    any       `json:"data,omitempty"`
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func NewAuthError(code ErrorCode, message string, data any) *AuthError {
	return &AuthError{
		Code:    code,
		Message: message,
		Data:    data,
	}
}
