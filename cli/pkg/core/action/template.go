/*
 Copyright 2021 The KubeSphere Authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package action

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"text/template"

	"bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"github.com/pkg/errors"
)

type Template struct {
	BaseAction
	Name         string
	Template     *template.Template
	Dst          string
	Data         util.Data
	PrintContent bool
}

func (t *Template) Execute(runtime connector.Runtime) error {
	templateStr, err := util.Render(t.Template, t.Data)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("render template %s failed", t.Template.Name()))
	}

	if t.PrintContent {
		logger.Infof("template %s result: %s", t.Name, templateStr)
	}

	if !util.IsExist(runtime.GetHostWorkDir()) {
		util.Mkdir(runtime.GetHostWorkDir())
	}

	var fileMode fs.FileMode = common.FileMode0644
	if runtime.GetSystemInfo().IsDarwin() {
		fileMode = common.FileMode0755
	}
	fileName := filepath.Join(runtime.GetHostWorkDir(), t.Template.Name())
	if err := util.WriteFile(fileName, []byte(templateStr), fileMode); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("write file %s failed", fileName))
	}

	if err := runtime.GetRunner().SudoScp(fileName, t.Dst); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("scp file %s to remote %s failed", fileName, t.Dst))
	}

	return nil
}
