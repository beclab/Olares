package response

import "fmt"

// TokenValidationError represents an error that occurs during token validation.
type TokenValidationError struct {
	innerErrs []error
	text      string
}

// Error returns merged multiple inner errors.
func (e TokenValidationError) Error() string {
	text := fmt.Sprintf("TokenValidationError: %s", e.text)
	if len(e.innerErrs) > 0 {
		text += fmt.Sprintf(", %v", e.innerErrs[0])
	}
	return text
}

// InnerError returns the first inner error from the TokenValidationError.
// If there are no inner errors, it returns nil.
func (e TokenValidationError) InnerError() error {
	if len(e.innerErrs) > 0 {
		return e.innerErrs[0]
	}
	return nil
}

func (e TokenValidationError) Text() string {
	return e.text
}

// NewTokenValidationError constructs a new TokenValidationError.
func NewTokenValidationError(text string, errs ...error) TokenValidationError {
	return TokenValidationError{innerErrs: errs, text: text}
}
