// Package chart ports devbox's docker-compose -> Olares chart scaffolding into
// olares-cli. It runs the same kompose conversion devbox uses, then writes an
// Olares chart layout (templates/ + OlaresManifest.yaml + Chart.yaml +
// values.yaml) to a local directory. Unlike devbox it never talks to a running
// cluster: namespace scoping and manifest schema are resolved statically /
// from flags instead of from the live Terminus CR.
package chart

import (
	"path/filepath"

	"github.com/kubernetes/kompose/pkg/kobject"
	"github.com/kubernetes/kompose/pkg/loader"
	"github.com/kubernetes/kompose/pkg/transformer"
	"github.com/kubernetes/kompose/pkg/transformer/kubernetes"
	"k8s.io/apimachinery/pkg/runtime"
)

// composeToK8s loads one or more docker-compose files and transforms every
// service into Kubernetes objects via kompose, entirely in memory: no files
// are written and kompose's PrintList stdout/file side effects are skipped.
func composeToK8s(opt kobject.ConvertOptions) ([]runtime.Object, error) {
	l, err := loader.GetLoader("compose")
	if err != nil {
		return nil, err
	}
	komposeObject := kobject.KomposeObject{
		ServiceConfigs: make(map[string]kobject.ServiceConfig),
	}
	komposeObject, err = l.LoadFile(opt.InputFiles, opt.Profiles, opt.NoInterpolate)
	if err != nil {
		return nil, err
	}
	komposeObject.Namespace = opt.Namespace

	workDir, err := transformer.GetComposeFileDir(opt.InputFiles)
	if err != nil {
		return nil, err
	}
	// kompose records env_file paths as absolute; rewrite to compose-relative
	// so the rendered chart stays portable.
	for _, service := range komposeObject.ServiceConfigs {
		if len(service.EnvFile) == 0 {
			continue
		}
		for i, envFile := range service.EnvFile {
			if !filepath.IsAbs(envFile) {
				continue
			}
			relPath, err := filepath.Rel(workDir, envFile)
			if err != nil {
				return nil, err
			}
			service.EnvFile[i] = filepath.ToSlash(relPath)
		}
	}

	t := &kubernetes.Kubernetes{Opt: opt}
	return t.Transform(komposeObject, opt)
}
