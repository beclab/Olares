package appcfg

import (
	"path/filepath"
	"testing"
)

func TestSubChartPath(t *testing.T) {
	cases := []struct {
		name string
		cfg  ApplicationConfig
		want string
	}{
		{"cache layout follows ChartsName",
			ApplicationConfig{AppName: "notesabc", RawAppName: "notes", ChartsName: "charts/market.olares/alice/notes/1.0.0-aaaaaaaaaaaa/notes"},
			filepath.Join("charts/market.olares/alice/notes/1.0.0-aaaaaaaaaaaa/notes", "server")},
		{"legacy AM keeps its flat path",
			ApplicationConfig{AppName: "notes", ChartsName: "./charts/notes"},
			filepath.Join("./charts/notes", "server")},
		{"no ChartsName falls back to the raw app name",
			ApplicationConfig{AppName: "notesabc", RawAppName: "notes"},
			ChartPath("notes", "server")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := SubChartPath(&tc.cfg, "server"); got != tc.want {
				t.Fatalf("SubChartPath = %s, want %s", got, tc.want)
			}
		})
	}
}
