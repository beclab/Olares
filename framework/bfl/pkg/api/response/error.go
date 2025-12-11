package response

import "fmt"

type TokenValidationError struct {
	innerErrs []error
	text      string
}

func (e TokenValidationError) Error() string {
	text := fmt.Sprintf("TokenValidationError: %s", e.text)
	if len(e.innerErrs) > 0 {
		if e.text == "" {
			text += e.innerErrs[0].Error()
		} else {
			text += fmt.Sprintf(", %v", e.innerErrs[0])
		}
	}
	return text
}

func (e TokenValidationError) InnerError() error {
	if len(e.innerErrs) > 0 {
		return e.innerErrs[0]
	}
	return nil
}

func (e TokenValidationError) Text() string {
	return e.text
}

func NewTokenValidationError(text string, errs ...error) TokenValidationError {
	return TokenValidationError{innerErrs: errs, text: text}
}
