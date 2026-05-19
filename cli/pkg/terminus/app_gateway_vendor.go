package terminus

import (
	"os"
	"path/filepath"
)

// ResolveAppGatewayVendorDir returns wizard/config/app-gateway-vendor (or the vendor dir itself)
// when markerFile exists under that directory. installerOrVendorDir may be the installer root (.dist)
// or the vendor directory. Falls back to OLARES_INSTALLER_DIR.
func ResolveAppGatewayVendorDir(installerOrVendorDir, markerFile string) string {
	for _, dir := range appGatewayVendorDirCandidates(installerOrVendorDir) {
		if markerFilePresent(dir, markerFile) {
			return dir
		}
	}
	return ""
}

func appGatewayVendorDirCandidates(installerOrVendorDir string) []string {
	var out []string
	if installerOrVendorDir != "" {
		out = append(out, filepath.Clean(installerOrVendorDir))
		out = append(out, filepath.Join(installerOrVendorDir, "wizard", "config", appGatewayVendorDirName))
	}
	if d := os.Getenv("OLARES_INSTALLER_DIR"); d != "" {
		out = append(out, filepath.Join(d, "wizard", "config", appGatewayVendorDirName))
	}
	return out
}

func markerFilePresent(vendorDir, markerFile string) bool {
	if vendorDir == "" || markerFile == "" {
		return false
	}
	st, err := os.Stat(filepath.Join(vendorDir, markerFile))
	return err == nil && !st.IsDir()
}
