package response

const (
	TokenInvalidErrCode = 100001
)

type TokenError struct {
	Text string
}

func (e TokenError) Error() string {
	return e.Text
}
