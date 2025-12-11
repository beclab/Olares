package controllers

var NodeIP string

var (
	BflStatefulSetName = "bfl"

	BflAnnotationAppCache = "appcache_hostpath"

	BflAnnotationDbData = "dbdata_hostpath"

	AppSubDirs = map[string][]int{
		"launcher": {65532, 65532},
	}

	DbDataSubDirs = map[string][]int{
		"mdbdata":        {1001, 1001},
		"mdbdata-config": {1001, 1001},
	}
)
