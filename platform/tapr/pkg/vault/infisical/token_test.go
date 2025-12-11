package infisical

import (
	"testing"
	"time"

	"k8s.io/klog/v2"
)

func TestIssueToekn(t *testing.T) {
	token := &tokenIssuer{}
	at, e := token.issueToken("6481c167ee527c75e9622fab", "bn8gxanmn2xjwdek6dtjqonnafvsoutj", 10*24*time.Hour, "accessToken", "aaa", "aaa")
	if e != nil {
		klog.Error(e)
		t.Fail()
	} else {
		t.Log("jwt: ", at)
	}
}
