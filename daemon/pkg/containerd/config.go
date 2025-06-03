package containerd

import (
	"bytetrade.io/web3os/terminusd/pkg/utils"
	"errors"
	"fmt"
	criconfig "github.com/containerd/containerd/pkg/cri/config"
	"github.com/containerd/containerd/plugin"
	serverconfig "github.com/containerd/containerd/services/server/config"
	"github.com/pelletier/go-toml"
	"k8s.io/klog/v2"
	"os"
)

const (
	DefaultContainerdConfigPath = "/etc/containerd/config.toml"
)

func getConfig() (*serverconfig.Config, error) {
	config := &serverconfig.Config{}
	err := serverconfig.LoadConfig(DefaultContainerdConfigPath, config)
	if err != nil {
		return nil, err
	}

	if config.Plugins == nil {
		config.Plugins = make(map[string]toml.Tree)
	}

	return config, nil
}

// newDefaultCRIPlugin returns an instance of CRI plugin with the default configuration.
// It's not actually used, and just for convenient plugin identification like plugin.URI and config decode
func newDefaultCRIPlugin() *plugin.Registration {
	criDefaultPluginConfig := criconfig.DefaultConfig()
	return &plugin.Registration{
		Type:   plugin.GRPCPlugin,
		ID:     "cri",
		Config: &criDefaultPluginConfig,
	}
}

func getCRIPluginConfig(config *serverconfig.Config) (*criconfig.PluginConfig, error) {
	if config == nil {
		return nil, errors.New("nil containerd config")
	}
	criPlugin := newDefaultCRIPlugin()
	criPluginConfigInterface, err := config.Decode(criPlugin)
	if err != nil {
		return nil, fmt.Errorf("failed to load cri plugin config from containerd config: %v", err)
	}
	criPluginConfig, ok := criPluginConfigInterface.(*criconfig.PluginConfig)
	if !ok {
		return nil, errors.New("failed to decode cri plugin config: type mismatch")
	}
	if criPluginConfig.Registry.Mirrors == nil {
		criPluginConfig.Registry.Mirrors = make(map[string]criconfig.Mirror)
	}
	return criPluginConfig, nil
}

func updateCRIPluginConfig(config *serverconfig.Config, criPluginConfig *criconfig.PluginConfig) error {
	backupPath := DefaultContainerdConfigPath + ".bak"
	if err := utils.CopyFile(DefaultContainerdConfigPath, backupPath); err != nil {
		klog.Errorf("failed to backup containerd config: %v", err)
		return err
	}
	defer os.Remove(backupPath)

	if criPluginConfig.Registry.ConfigPath != "" {
		// reset config path as it will mask the other options
		// we do not set mirrors in the config path
		// because image-service expects an explicit inline config in the Mirrors field
		// as of now
		criPluginConfig.Registry.ConfigPath = ""
	}

	criPlugin := newDefaultCRIPlugin()
	criPluginConfigBytes, err := toml.Marshal(criPluginConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal cri plugin config: %v", err)
	}
	criPluginConfigTree, err := toml.LoadBytes(criPluginConfigBytes)
	if err != nil {
		return fmt.Errorf("failed to load cri plugin config to toml tree: %v", err)
	}
	config.Plugins[criPlugin.URI()] = *criPluginConfigTree
	configFile, err := os.OpenFile(DefaultContainerdConfigPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open containerd config file for writing: %v", err)
	}
	defer configFile.Close()
	if err := toml.NewEncoder(configFile).Encode(config); err != nil {
		rollbackErr := utils.CopyFile(backupPath, DefaultContainerdConfigPath)
		if rollbackErr != nil {
			klog.Errorf("failed to rollback containerd config: %v", rollbackErr)
		}
		return fmt.Errorf("failed to encode new containerd config: %v", err)
	}

	return nil
}
