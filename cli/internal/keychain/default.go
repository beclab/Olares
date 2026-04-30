package keychain

// defaultKeychain forwards KeychainAccess calls to the package-level Get/Set/
// Remove functions, which in turn dispatch to the per-platform backend via
// build-tag-gated platformGet/Set/Remove implementations.
type defaultKeychain struct{}

func (d *defaultKeychain) Get(service, account string) (string, error) {
	return Get(service, account)
}

func (d *defaultKeychain) Set(service, account, value string) error {
	return Set(service, account, value)
}

func (d *defaultKeychain) Remove(service, account string) error {
	return Remove(service, account)
}

// Default returns a KeychainAccess backed by the real platform keychain. It is
// the only place production code should construct a KeychainAccess; tests can
// substitute their own implementation when injected through this seam.
func Default() KeychainAccess {
	return &defaultKeychain{}
}
