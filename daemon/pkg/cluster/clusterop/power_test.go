package clusterop

import (
	"context"
	"errors"
	"testing"

	clistate "github.com/beclab/Olares/cli/pkg/daemon/state"
	"github.com/beclab/Olares/daemon/pkg/commands"
)

type fakeCommand struct {
	commands.Operation
	executed bool
	err      error
}

func (f *fakeCommand) Execute(context.Context, any) (any, error) {
	f.executed = true
	return nil, f.err
}

func (f *fakeCommand) OperationName() commands.Operations { return f.Operation.Name }

func withPowerCommands(t *testing.T, reboot, shutdown *fakeCommand, validate error) {
	t.Helper()
	prevNew, prevValidate := newPowerCommand, validatePowerOp
	newPowerCommand = func(op Type) commands.Interface {
		switch op {
		case TypeShutdown:
			return shutdown
		case TypeReboot:
			return reboot
		default:
			return nil
		}
	}
	validatePowerOp = func(commands.Interface) error { return validate }
	t.Cleanup(func() {
		newPowerCommand, validatePowerOp = prevNew, prevValidate
	})
	withHostPower(t, nil)
}

// withHostPower replaces this node's own answer to "can a power command issued
// here reach the machine". It is the check the execution point makes for
// itself, rather than trusting the master's precheck.
func withHostPower(t *testing.T, refusal error) {
	t.Helper()
	prev := hostPowerSupport
	hostPowerSupport = func(Type) error { return refusal }
	t.Cleanup(func() { hostPowerSupport = prev })
}

func powerCode(t *testing.T, err error) string {
	t.Helper()
	var pe *PowerError
	if !errors.As(err, &pe) {
		t.Fatalf("error %v is not a PowerError, so no caller can map it to a stable code", err)
	}
	return pe.Code
}

func TestPowerHostRunsTheCommandForTheOperation(t *testing.T) {
	rb := &fakeCommand{Operation: commands.Operation{Name: commands.Reboot}}
	sd := &fakeCommand{Operation: commands.Operation{Name: commands.Shutdown}}
	withPowerCommands(t, rb, sd, nil)

	if err := PowerHost(context.Background(), TypeShutdown); err != nil {
		t.Fatalf("PowerHost: %v", err)
	}
	if !sd.executed || rb.executed {
		t.Errorf("shutdown=%v reboot=%v, want only the shutdown", sd.executed, rb.executed)
	}
}

// The daemon refuses a power command in states where it already refuses the
// single-node one. A cluster operation is not a way around that.
func TestPowerHostHonoursTheDaemonState(t *testing.T) {
	rb := &fakeCommand{Operation: commands.Operation{Name: commands.Reboot}}
	sd := &fakeCommand{Operation: commands.Operation{Name: commands.Shutdown}}
	withPowerCommands(t, rb, sd, errors.New("operation is not allowed while installing"))

	err := PowerHost(context.Background(), TypeReboot)
	if err == nil {
		t.Fatal("a forbidden state still powered the machine")
	}
	if rb.executed {
		t.Error("the command ran despite the state check failing")
	}
}

func TestPowerHostRefusesAnUnknownOperation(t *testing.T) {
	rb := &fakeCommand{Operation: commands.Operation{Name: commands.Reboot}}
	sd := &fakeCommand{Operation: commands.Operation{Name: commands.Shutdown}}
	withPowerCommands(t, rb, sd, nil)

	err := PowerHost(context.Background(), Type("halt"))
	if err == nil {
		t.Fatal("an unknown operation was accepted")
	}
	if got := powerCode(t, err); got != CodeUnsupportedOperation {
		t.Errorf("code = %q, want %s", got, CodeUnsupportedOperation)
	}
	if rb.executed || sd.executed {
		t.Error("an unknown operation reached a power command")
	}
}

func TestPowerHostReportsACommandFailure(t *testing.T) {
	rb := &fakeCommand{Operation: commands.Operation{Name: commands.Reboot}, err: errors.New("reboot: command not found")}
	withPowerCommands(t, rb, &fakeCommand{}, nil)

	err := PowerHost(context.Background(), TypeReboot)
	if err == nil {
		t.Fatal("a failing power command was reported as accepted")
	}
	if got := powerCode(t, err); got != CodeHostPowerFailed {
		t.Errorf("code = %q, want %s", got, CodeHostPowerFailed)
	}
}

// Inside a container the command acts on the container and leaves the machine
// running, which is the worst possible answer to "is this node off yet". The
// node that would run it is the one that has to say so: a master's precheck
// can be stale, wrong, or simply not have happened.
func TestPowerHostRefusesWhereTheCommandCannotReachTheMachine(t *testing.T) {
	for _, ty := range []Type{TypeReboot, TypeShutdown} {
		t.Run(string(ty), func(t *testing.T) {
			rb := &fakeCommand{Operation: commands.Operation{Name: commands.Reboot}}
			sd := &fakeCommand{Operation: commands.Operation{Name: commands.Shutdown}}
			withPowerCommands(t, rb, sd, nil)
			withHostPower(t, &PowerError{Code: CodePowerUnsupported, Message: "olaresd runs in a container on this node"})

			err := PowerHost(context.Background(), ty)
			if err == nil {
				t.Fatal("a containerized node powered something")
			}
			if got := powerCode(t, err); got != CodePowerUnsupported {
				t.Errorf("code = %q, want %s", got, CodePowerUnsupported)
			}
			if rb.executed || sd.executed {
				t.Error("the command ran on a node that cannot reach its own machine")
			}
		})
	}
}

// The support check comes before the state check and before the command is
// even built: nothing about a node that cannot power itself should depend on
// the daemon happening to be in a state that would have allowed it.
func TestPowerHostChecksItsOwnSupportFirst(t *testing.T) {
	rb := &fakeCommand{Operation: commands.Operation{Name: commands.Reboot}}
	withPowerCommands(t, rb, &fakeCommand{}, errors.New("operation is not allowed while installing"))
	withHostPower(t, &PowerError{Code: CodePowerUnsupported, Message: "olaresd runs in a container on this node"})

	err := PowerHost(context.Background(), TypeReboot)
	if got := powerCode(t, err); got != CodePowerUnsupported {
		t.Errorf("code = %q, want %s", got, CodePowerUnsupported)
	}
}

// hostPowerSupportIn is the production check. A container is refused without
// asking anything else: the command would act on the container and leave the
// machine running.
func TestHostPowerSupportRefusesAContainer(t *testing.T) {
	withCommandOnPath(t, true)
	mode := "docker"

	err := hostPowerSupportIn(clistate.State{ContainerMode: &mode}, TypeReboot)
	if err == nil {
		t.Fatal("a containerized daemon declared it can power its host")
	}
	if got := powerCode(t, err); got != CodePowerUnsupported {
		t.Errorf("code = %q, want %s", got, CodePowerUnsupported)
	}
}

func TestHostPowerSupportNeedsTheCommandItWouldRun(t *testing.T) {
	withCommandOnPath(t, false)

	err := hostPowerSupportIn(clistate.State{}, TypeReboot)
	if err == nil {
		t.Fatal("a node with no reboot command declared it can reboot")
	}
	if got := powerCode(t, err); got != CodePowerUnsupported {
		t.Errorf("code = %q, want %s", got, CodePowerUnsupported)
	}
}

func TestHostPowerSupportAcceptsAHostWithTheCommand(t *testing.T) {
	withCommandOnPath(t, true)

	for _, ty := range []Type{TypeReboot, TypeShutdown} {
		if err := hostPowerSupportIn(clistate.State{}, ty); err != nil {
			t.Errorf("%s refused on a host that can do it: %v", ty, err)
		}
	}
}

func TestHostPowerSupportRefusesAnUnknownOperation(t *testing.T) {
	withCommandOnPath(t, true)

	err := hostPowerSupportIn(clistate.State{}, Type("halt"))
	if got := powerCode(t, err); got != CodeUnsupportedOperation {
		t.Errorf("code = %q, want %s", got, CodeUnsupportedOperation)
	}
}

func withCommandOnPath(t *testing.T, found bool) {
	t.Helper()
	prev := lookPowerCommand
	lookPowerCommand = func(name string) (string, error) {
		if !found {
			return "", errors.New("exec: " + name + ": executable file not found in $PATH")
		}
		return "/sbin/" + name, nil
	}
	t.Cleanup(func() { lookPowerCommand = prev })
}
