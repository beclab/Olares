package mountnfs

type Param struct {
	Server       string
	MountBaseDir string
	MountPath    string
	NfsPath      string // e.g. /opt/nfs-share
	// User         string
	// Password     string
}
