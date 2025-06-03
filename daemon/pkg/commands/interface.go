package commands

import (
	"context"
	"errors"
)

type Interface interface {
	OperationName() Operations
	Execute(ctx context.Context, param any) (res any, err error)
}

type Operation struct {
	Name Operations
}

func (o *Operation) OperationName() Operations {
	return o.Name
}

func (o *Operation) Execute(_ context.Context, _ any) (res any, err error) {
	return nil, errors.New("not implement")
}
