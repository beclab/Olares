package doctor

import (
	"testing"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func TestDoctorCommandRegistersImages(t *testing.T) {
	cmd := NewDoctorCommand(cmdutil.NewFactory())

	if got := cmd.Name(); got != "doctor" {
		t.Fatalf("command name = %q, want doctor", got)
	}
	if sub, _, err := cmd.Find([]string{"images"}); err != nil || sub == nil || sub.Name() != "images" {
		t.Fatalf("doctor images not registered: sub=%v err=%v", sub, err)
	}
	if sub, _, err := cmd.Find([]string{"thirdleveldomain"}); err != nil || sub == nil || sub.Name() != "thirdleveldomain" {
		t.Fatalf("doctor thirdleveldomain not registered: sub=%v err=%v", sub, err)
	}
}
