package action

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestFU(t *testing.T) {
	var a = "/etc/kubernetes/addons/kubesphere.yaml"
	var b = filepath.Dir(a)
	fmt.Println(b)
}
