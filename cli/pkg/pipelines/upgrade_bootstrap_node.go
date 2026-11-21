package pipelines

import (
	"context"
	"fmt"
	"path"

	"github.com/beclab/Olares/cli/pkg/bootstrap/download"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/module"
	"github.com/beclab/Olares/cli/pkg/core/pipeline"
	"github.com/beclab/Olares/cli/pkg/daemon"
	"github.com/beclab/Olares/cli/pkg/manifest"
	"github.com/beclab/Olares/cli/pkg/terminus"
	"github.com/beclab/Olares/cli/pkg/utils"
	"github.com/beclab/Olares/cli/version"
	"github.com/spf13/viper"
)

// BootstrapNodeComponent is the one release component bootstrap-node fetches.
const BootstrapNodeComponent = "olaresd"

// UpgradeBootstrapNodePipeline brings this machine's olaresd to the version
// of this olares-cli: it fetches the release wizard for the manifest, only the
// olaresd component of that release, and replaces and restarts olaresd.
//
// The source version's olaresd, or the legacy Job patch, calls it right after
// installing this binary, so its command line is frozen; see
// NewCmdUpgradeBootstrapNode. Restarting olaresd usually ends the process
// that called it, which is why every step here is safe to repeat.
func UpgradeBootstrapNodePipeline(ctx context.Context) error {
	target := viper.GetString(common.FlagVersion)
	if target != version.VERSION {
		return fmt.Errorf("this olares-cli is %s and cannot bootstrap a node to %s", version.VERSION, target)
	}

	arg := common.NewArgument()
	arg.SetOlaresVersion(target)
	arg.SetOlaresCDNService(viper.GetString(common.FlagCDNService))
	if !utils.CheckUrl(arg.OlaresCDNService) {
		return fmt.Errorf("invalid cdn service")
	}

	runtime, err := common.NewKubeRuntime(*arg)
	if err != nil {
		return err
	}
	manifestPath := path.Join(runtime.GetInstallerDir(), "installation.manifest")

	fetch := &pipeline.Pipeline{
		Name: "Fetch Olaresd For Upgrade",
		Modules: []module.Module{
			&terminus.InstallWizardDownloadModule{
				Version:     target,
				CDNService:  arg.OlaresCDNService,
				UrlOverride: viper.GetString(common.FlagURLOverride),
			},
			&download.PackageDownloadModule{
				Manifest:   manifestPath,
				BaseDir:    runtime.GetBaseDir(),
				CDNService: arg.OlaresCDNService,
				Only:       []string{BootstrapNodeComponent},
			},
		},
		Runtime: runtime,
	}
	if err := fetch.Start(ctx); err != nil {
		return fmt.Errorf("fetch olaresd %s: %w", target, err)
	}

	manifestMap, err := manifest.ReadAll(manifestPath)
	if err != nil {
		return fmt.Errorf("read the %s manifest: %w", target, err)
	}
	replace := &pipeline.Pipeline{
		Name: "Prepare Olaresd daemon",
		Modules: []module.Module{
			&daemon.ReplaceOlaresdBinaryModule{
				ManifestModule: manifest.ManifestModule{Manifest: manifestMap, BaseDir: runtime.GetBaseDir()},
			},
		},
		Runtime: runtime,
	}
	if err := replace.Start(ctx); err != nil {
		return fmt.Errorf("replace olaresd with %s: %w", target, err)
	}
	return nil
}
