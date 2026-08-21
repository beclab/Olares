package router

import (
	"strings"
	"testing"
)

// A key appearing without being asked for is the surprise this notice exists
// to defuse, and each of these is a question the reader has next. Two of them
// were missing while the message only described what had already happened: how
// much authority the new key carries, and how to run the same command without
// creating one.
func TestTheMintedKeyNoticeAnswersWhatWhyAndHowToAvoid(t *testing.T) {
	notice := mintedKeyNotice(&createdKey{
		apiKeyView: apiKeyView{Name: "olares-cli on laptop", KeyPrefix: "sk-abcd"},
	})

	for _, want := range []struct{ what, substr string }{
		{"the name, so a key list is matchable against it", `"olares-cli on laptop"`},
		{"the prefix, for the same reason", "sk-abcd"},
		{"why there was no alternative", "cannot use the profile session"},
		{"how much authority it has", "no expiry, no quota and no model restriction"},
		{"how to end it", "key revoke"},
		{"how to drop only this machine's copy", "key current --forget"},
		{"how not to create one at all", "--api-key"},
		{"the scriptable way not to create one", dataPlaneKeyEnv},
	} {
		if !strings.Contains(notice, want.substr) {
			t.Errorf("the notice does not give %s (%q):\n%s", want.what, want.substr, notice)
		}
	}
}

// Router echoes the name back, but nothing here can require that of every
// version, and a notice naming a key as "" is worse than one repeating the
// name we sent.
func TestTheNoticeNamesTheKeyEvenIfRouterDoesNot(t *testing.T) {
	notice := mintedKeyNotice(&createdKey{apiKeyView: apiKeyView{KeyPrefix: "sk-abcd"}})
	if strings.Contains(notice, `""`) {
		t.Errorf("the notice names the key as an empty string:\n%s", notice)
	}
}
