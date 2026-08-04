package search

import (
	"strings"
	"testing"
)

func TestDecodeMyAppsResponse(t *testing.T) {
	t.Parallel()

	const method = "POST"
	const path = "/server/myApps"

	t.Run("envelope success", func(t *testing.T) {
		raw := []byte(`{"code":0,"data":[{"id":"wise","title":"Wise","state":"running","entrances":[]}]}`)
		apps, err := decodeMyAppsResponse(method, path, raw)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(apps) != 1 || apps[0].ID != "wise" {
			t.Fatalf("apps = %#v", apps)
		}
	})

	t.Run("envelope upstream error", func(t *testing.T) {
		raw := []byte(`{"code":500,"message":"internal error"}`)
		_, err := decodeMyAppsResponse(method, path, raw)
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "upstream returned code 500: internal error") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("bare array", func(t *testing.T) {
		raw := []byte(`[{"id":"drive","title":"Drive","state":"running","entrances":[]}]`)
		apps, err := decodeMyAppsResponse(method, path, raw)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(apps) != 1 || apps[0].ID != "drive" {
			t.Fatalf("apps = %#v", apps)
		}
	})
}
