package terminus

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/pkg/errors"
)

const linkerdIdentityCertDirName = "linkerd-identity-certs"

func ensureLinkerdIdentityCerts(certDir string) error {
	if linkerdIdentityCertsPresent(certDir) {
		return nil
	}
	if err := os.MkdirAll(certDir, 0o700); err != nil {
		return err
	}
	script := linkerdIdentityCertScript(filepath.Dir(certDir))
	if script == "" {
		return errors.New("linkerd identity cert generator script not found (framework/app-gateway/hack/generate-linkerd-identity-certs.sh)")
	}
	logger.InfoInstallationProgress("Generating Linkerd identity certificates (ECDSA P-256, CA 30y / issuer 3y) ...")
	cmd := exec.Command(script, certDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "generate linkerd identity certificates")
	}
	if !linkerdIdentityCertsPresent(certDir) {
		return errors.New("linkerd identity cert generation did not produce ca.crt/issuer.crt/issuer.key")
	}
	return nil
}

func linkerdIdentityCertsPresent(certDir string) bool {
	for _, name := range []string{"ca.crt", "ca.key", "issuer.crt", "issuer.key"} {
		if _, err := os.Stat(filepath.Join(certDir, name)); err != nil {
			return false
		}
	}
	return true
}

func linkerdIdentityCertScript(vendorDir string) string {
	if vendorDir == "" {
		return ""
	}
	p := filepath.Join(vendorDir, "generate-linkerd-identity-certs.sh")
	if st, err := os.Stat(p); err == nil && !st.IsDir() {
		return p
	}
	return ""
}
