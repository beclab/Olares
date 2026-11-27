package terminus

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestValidateUserChartSourcesBeforeCopy(t *testing.T) {
	root := t.TempDir()
	createChart := func(dir string) {
		t.Helper()
		dir = filepath.Join(root, dir)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "Chart.yaml"), []byte("apiVersion: v2\nname: fixture\nversion: 0.1.0\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	createChart("launcher")
	if err := validateUserChartSources(root); err == nil {
		t.Fatal("missing apps bundle was accepted")
	}
	createChart("apps/wizard")
	createChart("apps/auth")
	if err := os.Mkdir(filepath.Join(root, "apps", "system-apps"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := validateUserChartSources(root); err == nil {
		t.Fatal("incomplete chart directory was accepted")
	}
	createChart("apps/system-apps")
	if err := validateUserChartSources(root); err == nil {
		t.Fatal("bundle missing systemserver was accepted")
	}
	createChart("apps/systemserver")
	if err := validateUserChartSources(root); err != nil {
		t.Fatal(err)
	}
}

func TestPublishUserChartsDoesNotSwitchOnPartialCopy(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "publish", true: "failed copy"}[fail], func(t *testing.T) {
			var commands []string
			stage := "/userapps/.bundle.abcd1234"
			err := publishUserChartBundle("/installer/wizard/config", "app-service-0", func(args ...string) (string, error) {
				switch {
				case args[0] == "cp":
					tree := filepath.Base(args[1])
					commands = append(commands, tree)
					if args[2] != "app-service-0:"+stage {
						t.Fatalf("copy bypassed staging: %v", args)
					}
					if tree == "apps" && fail {
						return "", errors.New("interrupted transfer")
					}
				case args[5] == "mktemp":
					commands = append(commands, "stage")
					return stage + "\n", nil
				case args[5] == "rm":
					commands = append(commands, "cleanup")
					if args[7] != stage {
						t.Fatal("cleanup must only remove the staging directory")
					}
				case args[5] == "/busybox/sh":
					commands = append(commands, "publish")
				default:
					t.Fatalf("unexpected command: %v", args)
				}
				return "", nil
			})
			want := []string{"stage", "launcher", "apps", "publish"}
			if fail {
				want[3] = "cleanup"
			}
			if (err != nil) != fail || !slices.Equal(commands, want) {
				t.Fatalf("publish ran %v, %v; want %v, failure=%t", commands, err, want, fail)
			}
		})
	}
}
