package cluster

import (
	"github.com/beclab/Olares/cli/pkg/bootstrap/os"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/module"
	"github.com/beclab/Olares/cli/pkg/core/pipeline"
	"github.com/beclab/Olares/cli/pkg/gpu"
	"github.com/beclab/Olares/cli/pkg/k3s"
	"github.com/beclab/Olares/cli/pkg/kubernetes"
	"github.com/beclab/Olares/cli/pkg/manifest"
	"github.com/beclab/Olares/cli/pkg/storage"
	"github.com/beclab/Olares/cli/pkg/terminus"
)

func AddNodePhase(runtime *common.KubeRuntime) *pipeline.Pipeline {
	var err error
	var manifestMap manifest.InstallationManifest
	manifestMap, err = manifest.ReadAll(runtime.Arg.Manifest)
	if err != nil {
		logger.Fatal(err)
	}

	var m []module.Module
	m = append(m,
		&terminus.GetMasterInfoModule{},
		&terminus.CheckPreparedModule{Force: true},
		&storage.InstallJuiceFsModule{
			ManifestModule: manifest.ManifestModule{
				Manifest: manifestMap,
				BaseDir:  runtime.GetBaseDir(),
			},
		},
		&AddNodeModule{
			ManifestModule: manifest.ManifestModule{
				Manifest: manifestMap,
				BaseDir:  runtime.GetBaseDir(),
			},
		},
	)

	m = append(m, &terminus.SaveMasterHostConfigModule{}, &terminus.InstalledModule{})

	return &pipeline.Pipeline{
		Name:    "Add Worker Node To The Cluster",
		Modules: m,
		Runtime: runtime,
	}
}

type AddNodeModule struct {
	common.KubeModule
	manifest.ManifestModule
	underlyingModules []module.TaskModule
}

func (m *AddNodeModule) Init() {
	m.Name = "JoinKubernetesCluster"
	if m.KubeConf.Arg.Kubetype == common.K8s {
		m.underlyingModules = []module.TaskModule{
			&kubernetes.StatusModule{},
			&os.ConfigureOSModule{},
			&kubernetes.InstallKubeBinariesModule{
				ManifestModule: m.ManifestModule,
			},
			&kubernetes.JoinNodesModule{},
		}
	} else {
		m.underlyingModules = []module.TaskModule{
			&k3s.StatusModule{},
			&os.ConfigureOSModule{},
			&k3s.InstallKubeBinariesModule{
				ManifestModule: m.ManifestModule,
			},
			&k3s.JoinNodesModule{},
		}
	}
	m.underlyingModules = append(m.underlyingModules, &gpu.NodeLabelingModule{})
	for _, underlyingModule := range m.underlyingModules {
		underlyingModule.Default(m.Runtime, m.PipelineCache, m.ModuleCache)
		underlyingModule.AutoAssert()
		underlyingModule.Init()
		m.Tasks = append(m.Tasks, underlyingModule.GetTasks()...)
	}
}
