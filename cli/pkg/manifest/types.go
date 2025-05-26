package manifest

type fileUrl struct {
	Url      string
	Checksum string // md5 checksum
}

type itemUrl struct {
	AMD64 fileUrl
	ARM64 fileUrl
}

type ManifestItem struct {
	Filename  string
	Path      string
	Type      string
	URL       itemUrl
	ImageName string
	FileID    string
}

type InstallationManifest map[string]*ManifestItem

type ManifestModule struct {
	Manifest InstallationManifest
	BaseDir  string
}

type ManifestAction struct {
	Manifest InstallationManifest
	BaseDir  string
}
