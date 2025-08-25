package amd

import _ "embed"

// Source of truth file: put your AMD plugin YAML here and keep it versioned.
//go:generate bash -c "mkdir -p assets && cp -f ../../../infrastructure/gpu/.olares/config/amd/amd-gpu-device-plugin.yaml assets/amd-gpu-device-plugin.yaml"

//go:embed assets/amd-gpu-device-plugin.yaml
var amdManifest []byte

func mustManifestYAML() []byte {
	if len(amdManifest) == 0 {
		panic("embedded AMD GPU manifest is empty; run `go generate ./cli/pkg/gpu/amd` to populate it")
	}
	return amdManifest
}
