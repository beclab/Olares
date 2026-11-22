package apiserver

import (
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/oac"
	"k8s.io/klog/v2"
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

// registryApps loads every chart under root, in both the legacy flat layout
// (root/<app>) and the per-version cache layout
// (root/<source>/<owner>/<app>/<version>-<sha>/<app>). When the same app
// appears more than once, the most recently written chart wins.
func registryApps(root string) (map[string]*appcfg.AppConfiguration, error) {
	type found struct {
		cfg *appcfg.AppConfiguration
		mod time.Time
	}
	apps := make(map[string]found)
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if p == root {
				return err
			}
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		if d.Name() == ".tmp" {
			return filepath.SkipDir
		}
		fi, statErr := os.Stat(filepath.Join(p, oac.ManifestFileName))
		if statErr != nil {
			return nil
		}
		cfg, loadErr := oac.LoadAppConfiguration(p)
		if loadErr != nil {
			klog.Errorf("Failed to read app chart %s err=%v", p, loadErr)
			return filepath.SkipDir
		}
		name := cfg.Metadata.Name
		if prev, ok := apps[name]; !ok || fi.ModTime().After(prev.mod) {
			apps[name] = found{cfg, fi.ModTime()}
		}
		// A chart root's own sub-directories are sub-charts, not apps.
		return filepath.SkipDir
	})
	if err != nil {
		return nil, err
	}
	out := make(map[string]*appcfg.AppConfiguration, len(apps))
	for name, f := range apps {
		out[name] = f.cfg
	}
	return out, nil
}
