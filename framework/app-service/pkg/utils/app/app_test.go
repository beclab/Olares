package app

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestGetFirstSubDir(t *testing.T) {
	tests := []struct {
		fullPath string
		basePath string
		expected string
	}{
		{
			fullPath: "/olares/userdata/Cache/pvc-appcache-olares-yeaoioao6ib76mgo/dify/volumes/nginx/claim8",
			basePath: "/olares/userdata/Cache/pvc-appcache-olares-yeaoioao6ib76mgo",
			expected: "/olares/userdata/Cache/pvc-appcache-olares-yeaoioao6ib76mgo/dify",
		},
		{
			fullPath: "/olares/userdata/Cache/pvc-appcache-olares-yeaoioao6ib76mgo/dify/volumes/nginx/claim8",
			basePath: "/olares/userdata/Cache/pvc-appcache-olares-yeaoioao6ib76mgo/",
			expected: "/olares/userdata/Cache/pvc-appcache-olares-yeaoioao6ib76mgo/dify",
		},
		{
			fullPath: "/olares/userdata/Cache/pvc-appcache-olares-yeaoioao6ib76mgo/",
			basePath: "/olares/userdata/Cache/pvc-appcache-olares-yeaoioao6ib76mgo/",
			expected: "",
		},
		{
			fullPath: "/olares/userdata/Cache/pvc-appcache-olares-yeaoioao6ib76mgo",
			basePath: "/olares/userdata/Cache/pvc-appcache-olares-yeaoioao6ib76mgo",
			expected: "",
		},
		{
			fullPath: "/olares/userdata/Cache/pvc-appcache-olares-yeaoioao6ib76mgo/",
			basePath: "/olares/userdata/Cache/pvc-appcache-olares-yeaoioao6ib76mgo",
			expected: "",
		},
		{
			fullPath: "/olares/userdata/Cache/pvc-appcache-olares-yeaoioao6ib76mgo",
			basePath: "/olares/userdata/Cache/pvc-appcache-olares-yeaoioao6ib76mgo/",
			expected: "",
		},
		{
			fullPath: "/some/other/path",
			basePath: "/olares/userdata/Cache/pvc-appcache-olares-yeaoioao6ib76mgo",
			expected: "",
		},
		{
			fullPath: "/some/other/path",
			basePath: "/olares/userdata/Cache/pvc-appcache-olares-yeaoioao6ib76mgo/",
			expected: "",
		},
		{
			fullPath: "/some/other/path",
			basePath: "/olares/userdata/Cache/pvc-appcache-olares-yeaoioao6ib76mgo",
			expected: "",
		},
		{
			fullPath: "/olares/userdata/Cache/pvc-appcache-olares-yeaoioao6ib76mgo/dify",
			basePath: "/olares/userdata/Cache/pvc-appcache-olares-yeaoioao6ib76mgo",
			expected: "/olares/userdata/Cache/pvc-appcache-olares-yeaoioao6ib76mgo/dify",
		},
		{
			fullPath: "/olares/userdata/Cache/pvc-appcache-olares-yeaoioao6ib76mgo/dify",
			basePath: "/olares/userdata/Cache/pvc-appcache-olares-yeaoioao6ib76mgo",
			expected: "/olares/userdata/Cache/pvc-appcache-olares-yeaoioao6ib76mgo/dify",
		},
		{
			fullPath: "/olares/userdata/Cache/pvc-appcache-olares-yeaoioao6ib76mgo/dify/volumes/nginx/c6",
			basePath: "/olares/userdata/Cache/pvc-appcache-olares-yeaoioao6ib76mgo",
			expected: "/olares/userdata/Cache/pvc-appcache-olares-yeaoioao6ib76mgo/dify",
		},
		{
			fullPath: "/olares/userdata/Cache/pvc-appcache-olares-yeaoioao6ib76mgo/dify/volumes/nginx/c6",
			basePath: "",
			expected: "",
		},
		{
			fullPath: "/olares/userdata/Cache/pvc-appcache-olares-yeaoioao6ib76mgo/dify/volumes/nginx/c6",
			basePath: "/",
			expected: "",
		},
		{
			fullPath: "/olares/userdata/Cache/pvc-appcache-olares-yeaoioao6ib76mgo/dify/volumes/nginx/c6",
			basePath: "/olares",
			expected: "",
		},
	}
	for _, tt := range tests {
		result := GetFirstSubDir(tt.fullPath, tt.basePath)
		if result != tt.expected {
			assert.Equal(t, tt.expected, result)
		}
	}
}
