package preinstall

import (
	_ "embed"
	"fmt"
)

// catalogAppsData lists the apps every device running this version is expected
// to have, whatever medium it arrived on. It is compiled in rather than fetched:
// the list belongs to this release, so the binary that installs the release is
// the thing that knows it.
//
//go:embed catalog-apps.json
var catalogAppsData []byte

var catalogApps []DeclarationAppV2

func init() {
	var file struct {
		Apps []DeclarationAppV2 `json:"apps"`
	}
	if err := strictDecode(catalogAppsData, &file); err != nil {
		panic(fmt.Sprintf("decode embedded catalog apps: %v", err))
	}
	for index := range file.Apps {
		file.Apps[index].ChartSource = ChartSourceCatalog
	}
	catalogApps = file.Apps
}

func catalogDeclarationApps() []DeclarationAppV2 {
	return append([]DeclarationAppV2(nil), catalogApps...)
}
