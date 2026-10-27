package controllers

var NodeIP string

var (
	AppSubDirs = map[string][]int{
		"launcher": {65532, 65532},
	}

	DbDataSubDirs = map[string][]int{
		"mdbdata":        {1001, 1001},
		"mdbdata-config": {1001, 1001},
	}
)

const (
	UserStorageRoleLabel          = "bytetrade.io/user-storage"
	StorageRelativePathAnnotation = "bytetrade.io/storage-relative-path"
	InitializedNodesAnnotation    = "bytetrade.io/initialized-nodes"
)
