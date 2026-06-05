package files

import (
	"context"
	"testing"
)

// TestIsCommonFrontendPath pins the drive/Common recogniser the
// version gate keys on: only fileType=="drive" AND extend=="Common"
// (case-sensitive) counts. drive/Home, drive/Data and every other
// namespace are NOT Common.
func TestIsCommonFrontendPath(t *testing.T) {
	common := [][2]string{
		{"drive", "Common"},
	}
	notCommon := [][2]string{
		{"drive", "Home"},
		{"drive", "Data"},
		{"drive", "common"}, // case-sensitive: lowercase is not Common
		{"cache", "Common"}, // wrong fileType
		{"external", "node-1"},
		{"sync", "repo"},
		{"", ""},
	}
	for _, c := range common {
		if !isCommonFrontendPath(c[0], c[1]) {
			t.Errorf("isCommonFrontendPath(%q,%q) = false, want true", c[0], c[1])
		}
	}
	for _, c := range notCommon {
		if isCommonFrontendPath(c[0], c[1]) {
			t.Errorf("isCommonFrontendPath(%q,%q) = true, want false", c[0], c[1])
		}
	}
}

// TestRequireCommonBackendVersion_NoopWhenNotCommon confirms the gate
// short-circuits (no factory / network touch) when the operation does
// not involve drive/Common — exercised here with a nil Factory, which
// would panic if the function tried to resolve the backend version.
func TestRequireCommonBackendVersion_NoopWhenNotCommon(t *testing.T) {
	if err := requireCommonBackendVersion(context.Background(), nil, false); err != nil {
		t.Errorf("requireCommonBackendVersion(_, nil, false) = %v, want nil", err)
	}
}
