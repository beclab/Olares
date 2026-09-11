package app

import (
	"context"
	"strings"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
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
			expected: "dify",
		},
		{
			fullPath: "/olares/userdata/Cache/pvc-appcache-olares-yeaoioao6ib76mgo/dify/volumes/nginx/claim8",
			basePath: "/olares/userdata/Cache/pvc-appcache-olares-yeaoioao6ib76mgo/",
			expected: "dify",
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
			expected: "dify",
		},
		{
			fullPath: "/olares/userdata/Cache/pvc-appcache-olares-yeaoioao6ib76mgo/dify",
			basePath: "/olares/userdata/Cache/pvc-appcache-olares-yeaoioao6ib76mgo",
			expected: "dify",
		},
		{
			fullPath: "/olares/userdata/Cache/pvc-appcache-olares-yeaoioao6ib76mgo/dify/volumes/nginx/c6",
			basePath: "/olares/userdata/Cache/pvc-appcache-olares-yeaoioao6ib76mgo",
			expected: "dify",
		},
		{
			fullPath: "/olares/userdata/Cache/pvc-appcache-olares-yeaoioao6ib76mgo/dify/volumes/nginx/c6",
			basePath: "",
			expected: "",
		},
		{
			fullPath: "/olares/userdata/Cache/pvc-appcache-olares-yeaoioao6ib76mgo/dify/volumes/nginx/c6",
			basePath: "/",
			expected: "olares",
		},
		{
			fullPath: "/olares/userdata/Cache/pvc-appcache-olares-yeaoioao6ib76mgo/dify/volumes/nginx/c6",
			basePath: "/olares",
			expected: "userdata",
		},
	}
	for _, tt := range tests {
		result := GetFirstSubDir(tt.fullPath, tt.basePath)
		if result != tt.expected {
			assert.Equal(t, tt.expected, result)
		}
	}
}

// ConfigOptions.Version only reaches the chart repo through
// GetIndexAndDownloadChart, so it is inert whenever NeedDownloadChart is false:
// the config is read from ./charts/{RawAppName}, a path that carries no version.
// A caller who sets Version there gets whatever the last download left on disk
// while believing it pinned a version, which is how a versioned upgrade came to
// deploy a different version's templates.
func TestVersionIsInertWithoutNeedDownloadChart(t *testing.T) {
	// A missing manifest is fine: getAppConfigFromConfigurationFile returns the
	// resolved chartPath alongside the error, which is the value under test.
	// RepoURL stays empty to prove nothing tried to reach the repo.
	resolve := func(version string) string {
		_, chartPath, err := getAppConfigFromRepo(context.Background(), &ConfigOptions{
			App:        "clitest",
			RawAppName: "clitest",
			Version:    version,
		})
		assert.Assert(t, err != nil, "expected the missing manifest to surface an error")
		return chartPath
	}

	want := appcfg.AppChartPath("clitest")
	for _, version := range []string{"", "0.1.2", "0.1.3"} {
		if got := resolve(version); got != want {
			t.Fatalf("Version %q resolved chart path %q, want %q", version, got, want)
		}
	}
	if strings.Contains(want, "0.1.") {
		t.Fatalf("chart path %q is version-scoped; this test no longer describes the code", want)
	}
}
