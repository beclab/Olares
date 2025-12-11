package v1

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestXxx(t *testing.T) {
	str := `{"f1":null}`

	var s struct {
		F1 *int64 `json:"f1"`
	}

	err := json.Unmarshal([]byte(str), &s)
	if err != nil {
		panic(err)
	}

	if s.F1 == nil {
		fmt.Print("nil value")
	}

	fmt.Printf("value, %+v", s)
}
