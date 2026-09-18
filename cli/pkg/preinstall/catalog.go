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

// CatalogPolicy says whether a publish declares the catalog apps of the release
// it publishes for. It is stated by every caller rather than defaulted: a run
// that declared the wrong list would still succeed, and the device would be left
// preinstalling apps nobody asked it to, or none of the ones it was meant to.
type CatalogPolicy string

const (
	// DeclareCatalogApps belongs to an upgrade. The device is already running
	// and can reach the catalog, so the apps this release expects are declared
	// whether or not any medium carried them.
	DeclareCatalogApps = CatalogPolicy("declare")
	// OmitCatalogApps belongs to a CLI install, which declares what the medium
	// carries and nothing else.
	OmitCatalogApps = CatalogPolicy("omit")
)

func (p CatalogPolicy) Valid() bool {
	return p == DeclareCatalogApps || p == OmitCatalogApps
}

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
