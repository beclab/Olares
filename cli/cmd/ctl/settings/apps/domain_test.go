package apps

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunDomainSet_RMWPreservesUntouchedFields(t *testing.T) {
	doer := &fakeDoer{}
	doer.enqueueEnvelope(SetupDomain{
		ThirdLevelDomain: "myhost",
		ThirdPartyDomain: "files.example.com",
		Cert:             "PEM-CERT",
		Key:              "PEM-KEY",
		CnameStatus:      "ThirdParty",
	})
	doer.enqueueEmptyEnvelope()

	// Only --third-level is updated; third-party + cert + key must
	// survive the merge.
	flags := domainSetFlags{
		thirdLevel:    "newhost",
		thirdLevelSet: true,
	}
	if err := runDomainSetWithDoer(context.Background(), doer, "files", "file", flags); err != nil {
		t.Fatalf("runDomainSetWithDoer: %v", err)
	}
	if len(doer.calls) != 2 {
		t.Fatalf("want GET + POST (2 calls), got %d", len(doer.calls))
	}
	post := doer.calls[1].body.(setupDomainBody)
	if post.ThirdLevelDomain != "newhost" {
		t.Errorf("third-level not applied: %q", post.ThirdLevelDomain)
	}
	if post.ThirdPartyDomain != "files.example.com" {
		t.Errorf("third-party should be preserved, got %q", post.ThirdPartyDomain)
	}
	if post.Cert != "PEM-CERT" || post.Key != "PEM-KEY" {
		t.Errorf("cert/key should be preserved across RMW; got cert=%q key=%q", post.Cert, post.Key)
	}
}

func TestRunDomainSet_ClearThirdPartyAlsoClearsCertKey(t *testing.T) {
	doer := &fakeDoer{}
	doer.enqueueEnvelope(SetupDomain{
		ThirdLevelDomain: "myhost",
		ThirdPartyDomain: "files.example.com",
		Cert:             "PEM-CERT",
		Key:              "PEM-KEY",
	})
	doer.enqueueEmptyEnvelope()

	flags := domainSetFlags{clearThirdParty: true}
	if err := runDomainSetWithDoer(context.Background(), doer, "files", "file", flags); err != nil {
		t.Fatalf("runDomainSetWithDoer: %v", err)
	}
	post := doer.calls[1].body.(setupDomainBody)
	if post.ThirdPartyDomain != "" || post.Cert != "" || post.Key != "" {
		t.Errorf("clear-third-party must zero domain+cert+key; got %+v", post)
	}
	if post.ThirdLevelDomain != "myhost" {
		t.Errorf("third-level should NOT be cleared by --clear-third-party; got %q", post.ThirdLevelDomain)
	}
}

func TestRunDomainSet_CertKeyFromFile(t *testing.T) {
	tmp := t.TempDir()
	certPath := filepath.Join(tmp, "fullchain.pem")
	keyPath := filepath.Join(tmp, "privkey.pem")
	if err := os.WriteFile(certPath, []byte("---CERT BODY---\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyPath, []byte("---KEY BODY---\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	doer := &fakeDoer{}
	doer.enqueueEnvelope(SetupDomain{}) // empty current setup
	doer.enqueueEmptyEnvelope()

	flags := domainSetFlags{
		thirdParty:    "files.example.com",
		thirdPartySet: true,
		certFile:      certPath,
		keyFile:       keyPath,
	}
	if err := runDomainSetWithDoer(context.Background(), doer, "files", "file", flags); err != nil {
		t.Fatalf("runDomainSetWithDoer: %v", err)
	}
	post := doer.calls[1].body.(setupDomainBody)
	if !strings.Contains(post.Cert, "CERT BODY") {
		t.Errorf("cert file contents not in body: %q", post.Cert)
	}
	if !strings.Contains(post.Key, "KEY BODY") {
		t.Errorf("key file contents not in body: %q", post.Key)
	}
}

func TestRunDomainSet_RejectsConflictingFlags(t *testing.T) {
	cases := []struct {
		name    string
		flags   domainSetFlags
		errSub  string
	}{
		{
			name:   "third-level and clear-third-level mutex",
			flags:  domainSetFlags{thirdLevel: "x", thirdLevelSet: true, clearThirdLevel: true},
			errSub: "mutually exclusive",
		},
		{
			name:   "third-party requires cert and key",
			flags:  domainSetFlags{thirdParty: "x.com", thirdPartySet: true},
			errSub: "--cert-file and --key-file",
		},
		{
			name:   "clear-third-party with cert/key not allowed",
			flags:  domainSetFlags{clearThirdParty: true, certFile: "/dev/null"},
			errSub: "cannot be combined",
		},
		{
			name:   "nothing to do",
			flags:  domainSetFlags{},
			errSub: "nothing to do",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			doer := &fakeDoer{}
			err := runDomainSetWithDoer(context.Background(), doer, "files", "file", c.flags)
			if err == nil {
				t.Fatal("want validation err, got nil")
			}
			if !strings.Contains(err.Error(), c.errSub) {
				t.Errorf("err=%q want substring %q", err.Error(), c.errSub)
			}
			if len(doer.calls) != 0 {
				t.Errorf("validation should reject before any wire call; got calls=%+v", doer.calls)
			}
		})
	}
}

func TestRunDomainSet_GetFailureBubbles(t *testing.T) {
	doer := &fakeDoer{}
	// Server says the entrance does not exist or user is forbidden:
	// the GET fails with a non-zero envelope and we should NOT POST.
	doer.responses = append(doer.responses, []byte(`{"code":403,"message":"forbidden"}`))

	flags := domainSetFlags{thirdLevel: "x", thirdLevelSet: true}
	err := runDomainSetWithDoer(context.Background(), doer, "files", "file", flags)
	if err == nil {
		t.Fatal("want err on GET failure, got nil")
	}
	if !strings.Contains(err.Error(), "forbidden") {
		t.Errorf("err=%q does not surface upstream message", err.Error())
	}
	if len(doer.calls) != 1 {
		t.Errorf("want exactly 1 call (the failed GET), got %d (%+v)", len(doer.calls), doer.calls)
	}
}
