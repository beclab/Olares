package appinstaller

import (
	"reflect"
	"strings"
	"testing"
	"text/template"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
)

// buildExposePortsValues is how a chart learns the host port the platform
// allocated for a `ports[]` entry. The map is keyed by port name and holds
// the exposePort exactly as SetExposePorts left it on the app config.
func TestBuildExposePortsValues(t *testing.T) {
	cases := []struct {
		name  string
		ports []appcfg.ServicePort
		want  map[string]interface{}
	}{
		{name: "no ports", ports: nil, want: map[string]interface{}{}},
		{
			name: "allocated and pinned ports",
			ports: []appcfg.ServicePort{
				{Name: "rtc", Host: "lares", Port: 7882, Protocol: "udp", ExposePort: 47882},
				{Name: "rtc-tcp", Host: "lares", Port: 7881, Protocol: "tcp", ExposePort: 31234},
			},
			want: map[string]interface{}{"rtc": int32(47882), "rtc-tcp": int32(31234)},
		},
		{
			name:  "not yet allocated renders as zero",
			ports: []appcfg.ServicePort{{Name: "game", Host: "server", Port: 7777, Protocol: "udp"}},
			want:  map[string]interface{}{"game": int32(0)},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := buildExposePortsValues(tc.ports); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("buildExposePortsValues() = %#v, want %#v", got, tc.want)
			}
		})
	}
}

// The value must render as a plain number so a chart can pass it straight
// into a container argument or env var.
func TestExposePortsValuesRenderInTemplate(t *testing.T) {
	values := map[string]interface{}{
		"exposePorts": buildExposePortsValues([]appcfg.ServicePort{{Name: "rtc", Port: 7882, Protocol: "udp", ExposePort: 47882}}),
	}
	tpl := template.Must(template.New("t").Parse(`--rtc.udp_port={{ .Values.exposePorts.rtc }}`))
	var out strings.Builder
	if err := tpl.Execute(&out, map[string]interface{}{"Values": values}); err != nil {
		t.Fatal(err)
	}
	if out.String() != "--rtc.udp_port=47882" {
		t.Fatalf("rendered %q", out.String())
	}
}
