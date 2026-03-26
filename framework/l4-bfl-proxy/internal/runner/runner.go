package runner

import "context"

type Runner interface {
	Name() string
	Start(ctx context.Context) error
}
