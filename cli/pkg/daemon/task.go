package daemon

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/action"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/util"
	"github.com/beclab/Olares/cli/pkg/daemon/templates"
	"github.com/beclab/Olares/cli/pkg/manifest"
	"github.com/beclab/Olares/cli/pkg/utils"
	"github.com/pkg/errors"
)

type InstallTerminusdBinary struct {
	common.KubeAction
	manifest.ManifestAction
}

func (g *InstallTerminusdBinary) Execute(runtime connector.Runtime) error {
	if err := utils.ResetTmpDir(runtime); err != nil {
		return err
	}

	binary, err := g.Manifest.Get("olaresd")
	if err != nil {
		return fmt.Errorf("get kube binary olaresd info failed: %w", err)
	}

	path := binary.FilePath(g.BaseDir)

	dst := filepath.Join(common.TmpDir, binary.Filename)
	if err := runtime.GetRunner().Scp(path, dst); err != nil {
		return errors.Wrap(errors.WithStack(err), "sync olaresd tar.gz failed")
	}

	installCmd := fmt.Sprintf("tar -zxf %s && cp -f olaresd /usr/local/bin/ && chmod +x /usr/local/bin/olaresd && rm -rf olaresd*", dst)
	if _, err := runtime.GetRunner().SudoCmd(installCmd, false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "install olaresd binaries failed")
	}
	return nil
}

type UpdateOlaresdServiceEnv struct {
	common.KubeAction
}

func (a *UpdateOlaresdServiceEnv) Execute(runtime connector.Runtime) error {
	envFilePath := filepath.Join("/etc/systemd/system/", templates.TerminusdEnv.Name())
	versionKey := "INSTALLED_VERSION"
	updateVersionCMD := fmt.Sprintf("sed -i '/%s/c\\%s=%s' %s ", versionKey, versionKey, a.KubeConf.Arg.OlaresVersion, envFilePath)
	if _, err := runtime.GetRunner().SudoCmd(updateVersionCMD, false, false); err != nil {
		return fmt.Errorf("update olaresd env failed: %v", err)
	}
	return nil
}

type GenerateTerminusdServiceEnv struct {
	common.KubeAction
}

func (g *GenerateTerminusdServiceEnv) Execute(runtime connector.Runtime) error {
	var baseDir = runtime.GetBaseDir()
	templateAction := action.Template{
		Name:     "OlaresdServiceEnv",
		Template: templates.TerminusdEnv,
		Dst:      filepath.Join("/etc/systemd/system/", templates.TerminusdEnv.Name()),
		Data: util.Data{
			"Version":         g.KubeConf.Arg.OlaresVersion,
			"KubeType":        g.KubeConf.Arg.Kubetype,
			"RegistryMirrors": g.KubeConf.Arg.RegistryMirrors,
			"BaseDir":         baseDir,
			"GpuEnable":       utils.FormatBoolToInt(g.KubeConf.Arg.GPU.Enable),
			"TokenMaxAge":     g.KubeConf.Arg.TokenMaxAge,
			"MarketProvider":  g.KubeConf.Arg.MarketProvider,
		},
		PrintContent: true,
	}

	templateAction.Init(nil, nil)
	if err := templateAction.Execute(runtime); err != nil {
		return err
	}
	return nil
}

type GenerateTerminusdService struct {
	common.KubeAction
}

func (g *GenerateTerminusdService) Execute(runtime connector.Runtime) error {
	templateAction := action.Template{
		Name:         "OlaresdService",
		Template:     templates.TerminusdService,
		Dst:          filepath.Join("/etc/systemd/system/", templates.TerminusdService.Name()),
		Data:         util.Data{},
		PrintContent: true,
	}

	templateAction.Init(nil, nil)
	if err := templateAction.Execute(runtime); err != nil {
		return err
	}
	return nil
}

type EnableTerminusdService struct {
	common.KubeAction
}

func (e *EnableTerminusdService) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("systemctl enable --now olaresd",
		false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "enable olaresd failed")
	}
	return nil
}

type DisableTerminusdService struct {
	common.KubeAction
}

func (s *DisableTerminusdService) Execute(runtime connector.Runtime) error {
	stdout, _ := runtime.GetRunner().SudoCmd("systemctl is-active olaresd", false, false)
	if stdout == "active" {
		if _, err := runtime.GetRunner().SudoCmd("systemctl disable --now olaresd", false, true); err != nil {
			return errors.Wrap(errors.WithStack(err), "disable olaresd failed")
		}
	}
	return nil
}

type UninstallTerminusd struct {
	common.KubeAction
}

func (r *UninstallTerminusd) Execute(runtime connector.Runtime) error {
	var olaresdFiles []string
	svcpath := filepath.Join("/etc/systemd/system", templates.TerminusdService.Name())
	svcenvpath := filepath.Join("/etc/systemd/system", templates.TerminusdEnv.Name())
	binPath := "/usr/local/bin/olaresd"
	olaresdFiles = append(olaresdFiles, svcpath, svcenvpath, binPath)
	for _, pidFile := range []string{"installing.pid", "changingip.pid"} {
		olaresdFiles = append(olaresdFiles, filepath.Join(runtime.GetBaseDir(), pidFile))
	}
	for _, f := range olaresdFiles {
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("rm -rf %s", f), false, false); err != nil {
			return errors.Wrap(errors.WithStack(err), "remove olaresd failed")
		}
	}
	return nil
}

type CheckTerminusdService struct {
}

func (c *CheckTerminusdService) Execute() error {
	cmd := exec.Command("/bin/sh", "-c", "systemctl list-unit-files --no-legend --no-pager -l | grep olaresd")
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	return nil
}
