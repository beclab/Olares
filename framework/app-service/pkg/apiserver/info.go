package apiserver

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"

	"bytetrade.io/web3os/app-service/pkg/appcfg"
	"bytetrade.io/web3os/app-service/pkg/utils/config"

	"gopkg.in/yaml.v2"
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
	cfgFileName := fmt.Sprintf("%s/%s/%s", appcfg.ChartsPath, dir.Name(), config.AppCfgFileName)

	f, err := os.Open(cfgFileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	info, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var appCfg appcfg.AppConfiguration
	if err = yaml.Unmarshal(info, &appCfg); err != nil {
		return nil, err
	}

	// cache app icon data
	// var icon string
	// iconData, found := imageCache.Get(appCfg.Metadata.Icon)
	// if !found {
	// 	icon, err = readImageToBase64(fmt.Sprintf("%s/%s", ChartsPath, dir.Name()), appCfg.Metadata.Icon)
	// 	if err != nil {
	// 		klog.Errorf("get app icon error: %s", err)
	// 	} else {
	// 		imageCache.Set(appCfg.Metadata.Icon, icon, cache.DefaultExpiration)
	// 	}
	// } else {
	// 	icon = iconData.(string)
	// }

	return &appCfg, nil
}
