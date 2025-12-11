package uuid

import (
	"github.com/google/uuid"
)

func NewUUID() string {
	return uuid.New().String()
}

func IsValid(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}
