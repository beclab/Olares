package action

import (
	"embed"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/util"
)

var scripts embed.FS

func Assets() embed.FS {
	return scripts
}

type Script struct {
	BaseAction
	Name        string
	File        string
	Args        []string
	Envs        map[string]string
	PrintOutput bool
	PrintLine   bool
	Ignore      bool
}

func (s *Script) Execute(runtime connector.Runtime) error {
	if s.Ignore {
		return nil
	}

	if !util.IsExist(s.File) {
		return errors.New(fmt.Sprintf("script file %s not exist", s.File))
	}

	var envs string
	if s.Envs != nil && len(s.Envs) > 0 {
		for k, v := range s.Envs {
			envs += fmt.Sprintf("export %s=%s;", k, v)
		}
	}

	var cmd = fmt.Sprintf("%s bash %s %s", envs, s.File, strings.Join(s.Args, " "))
	_, err := runtime.GetRunner().SudoCmd(cmd, s.PrintOutput, s.PrintLine)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("exec script %s failed, args: %v", s.File, s.Args))
	}

	return nil
}
