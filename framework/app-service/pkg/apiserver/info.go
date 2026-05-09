package apiserver

import (
	"io/fs"
	"path/filepath"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/oac"
)

/*
OlaresManifest.yaml

OlaresManifest.version: v1
metadata:
  name: <chart name>
  description: <desc>
  icon: <icon file uri>
  appid: <app register id>
  version: <app version>
  title: <app title>
*/

func (h *Handler) readAppInfo(dir fs.FileInfo) (*appcfg.AppConfiguration, error) {
	chartDir := filepath.Join(appcfg.ChartsPath, dir.Name())
	cfg, err := oac.LoadAppConfiguration(chartDir)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
