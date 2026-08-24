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

// PublishCatalogDeclaration declares this version's catalog apps and nothing
// else, which is what an upgrade has to say: the charts are on the network
// rather than on the medium, and whatever the device already installed for an
// earlier release keeps its own declaration.
func PublishCatalogDeclaration(rootDir, osVersion string) error {
	declaration := DeclarationV2{
		SchemaVersion: DeclarationSchemaVersion,
		OSVersion:     osVersion,
		Apps:          catalogDeclarationApps(),
	}
	declaration.GeneratedAt = generatedNow()
	if err := ValidateDeclaration(&declaration); err != nil {
		return err
	}
	return publishDeclaration(nil, rootDir, declaration)
}
