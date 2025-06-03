package mountsmb

type Param struct {
	MountBaseDir string
	SmbPath      string // e.g. //my-window-server/share-path
	User         string
	Password     string
}
