package command

import (
	"os"
	"os/exec"
)

const (
	DefaultFrpcBinary  = "/frpc"
	DefaultFrpcCfgPath = "/frpc.toml"
)

type FrpcCommand struct {
	binary   string
	confPath string
}

func NewFrpcCommand() *FrpcCommand {
	return &FrpcCommand{
		binary:   DefaultFrpcBinary,
		confPath: DefaultFrpcCfgPath,
	}
}

func (f *FrpcCommand) StartCmd(args ...string) *exec.Cmd {
	var cmdArgs []string

	cmdArgs = append(cmdArgs, "-c", f.confPath)
	cmdArgs = append(cmdArgs, args...)
	return exec.Command(f.binary, cmdArgs...)
}

func (f *FrpcCommand) output(args ...string) ([]byte, error) {
	return exec.Command(f.binary, args...).CombinedOutput()
}

func (f *FrpcCommand) Reload() ([]byte, error) {
	return f.output("reload", "-c", f.confPath)
}

func (f *FrpcCommand) VerifyConfig(confPath string) ([]byte, error) {
	if confPath == "" {
		confPath = f.confPath
	}
	return f.output("verify", "-c", confPath)
}

func (f *FrpcCommand) WriteConfig(content []byte) error {
	confFile, err := os.OpenFile(f.confPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	_, err = confFile.Write(content)
	return err
}
