package utils

import (
	"bytes"
	"fmt"
	"io/ioutil"
)

var (
	serviceAccountDir = "/var/run/secrets/kubernetes.io/serviceaccount"

	serviceAccountNamespace = fmt.Sprintf("%s/namespace", serviceAccountDir)

	serviceAccountToken = fmt.Sprintf("%s/token", serviceAccountDir)
)

func Namespace() string {
	bs, err := ioutil.ReadFile(serviceAccountNamespace)
	if err != nil {
		return ""
	}

	if bs == nil || len(bs) == 0 {
		return ""
	}
	bsLen := len(bs)

	var bf bytes.Buffer
	for i := 0; i < bsLen; i++ {
		c := bs[i]
		if c != 32 && c != 10 && c != 13 {
			bf.WriteByte(c)
		}
	}
	return bf.String()
}
