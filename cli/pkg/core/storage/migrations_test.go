package storage

import (
	"fmt"
	"strings"
	"testing"
)

func TestMatch(t *testing.T) {
	var m = "._V0001.initial_schema.down.sql"
	m = strings.TrimPrefix(m, "._")
	var ok = reMigration.MatchString(m)
	fmt.Println("---ok---", ok)
}
