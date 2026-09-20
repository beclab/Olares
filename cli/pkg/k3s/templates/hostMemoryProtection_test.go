package templates

import (
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The drop-ins are all called 10-olares-memory.conf on the node, but
// action.Template stages each one locally under its template's Name() before
// copying it. Sharing a Name() there means two of the three quietly overwrite
// the third, and every value still looks right in review.
func TestHostMemoryProtectionTemplateNamesAreDistinct(t *testing.T) {
	seen := map[string]string{}
	for _, tmpl := range []*template.Template{
		HostMemoryProtectionSlice,
		HostMemoryProtectionK3s,
		HostMemoryProtectionContainerd,
		HostMemoryProtectionEtcd,
	} {
		name := tmpl.Name()
		assert.Empty(t, seen[name], "template name %q is used twice; the staged files would collide", name)
		seen[name] = name
	}
	assert.Len(t, seen, 4)
}

func TestHostMemoryProtectionRenders(t *testing.T) {
	tests := []struct {
		name     string
		tmpl     *template.Template
		data     map[string]interface{}
		section  string
		contains []string
	}{
		{
			name:     "slice carries only the ancestor protection",
			tmpl:     HostMemoryProtectionSlice,
			data:     map[string]interface{}{"SystemSliceMemoryMin": "2304M"},
			section:  "[Slice]",
			contains: []string{"MemoryMin=2304M"},
		},
		{
			name:     "k3s stops swapping and keeps its text resident",
			tmpl:     HostMemoryProtectionK3s,
			data:     map[string]interface{}{"K3sMemoryMin": "1536M"},
			section:  "[Service]",
			contains: []string{"MemoryMin=1536M", "MemorySwapMax=0"},
		},
		{
			name:     "containerd covers itself and the shims",
			tmpl:     HostMemoryProtectionContainerd,
			data:     map[string]interface{}{"ContainerdMemoryMin": "512M"},
			section:  "[Service]",
			contains: []string{"MemoryMin=512M", "MemorySwapMax=0"},
		},
		{
			name:     "etcd protects the datastore it fsyncs against",
			tmpl:     HostMemoryProtectionEtcd,
			data:     map[string]interface{}{"EtcdMemoryMin": "512M"},
			section:  "[Service]",
			contains: []string{"MemoryMin=512M", "MemorySwapMax=0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out strings.Builder
			require.NoError(t, tt.tmpl.Execute(&out, tt.data))
			rendered := out.String()

			assert.Contains(t, rendered, tt.section)
			for _, want := range tt.contains {
				assert.Contains(t, rendered, want)
			}
			// An unresolved key renders as "<no value>", which systemd would
			// reject at load time and only on the node.
			assert.NotContains(t, rendered, "<no value>")
		})
	}
}

// A slice unit takes [Slice], a service takes [Service]; systemd ignores a
// resource setting that lands in the wrong section without failing the unit, so
// the mistake would show up only as memory protection that never applied.
func TestHostMemoryProtectionSectionsMatchUnitType(t *testing.T) {
	var slice strings.Builder
	require.NoError(t, HostMemoryProtectionSlice.Execute(&slice, map[string]interface{}{
		"SystemSliceMemoryMin": "2304M",
	}))
	assert.NotContains(t, slice.String(), "[Service]")

	for _, tmpl := range []*template.Template{
		HostMemoryProtectionK3s, HostMemoryProtectionContainerd, HostMemoryProtectionEtcd,
	} {
		var out strings.Builder
		require.NoError(t, tmpl.Execute(&out, map[string]interface{}{
			"K3sMemoryMin": "1536M", "ContainerdMemoryMin": "512M", "EtcdMemoryMin": "512M",
		}))
		assert.NotContains(t, out.String(), "[Slice]")
	}
}
