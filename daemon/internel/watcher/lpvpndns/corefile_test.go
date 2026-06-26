package lpvpndns

import "testing"

func TestParsePodCIDRFromKubeServicesIPTables(t *testing.T) {
	cases := []struct {
		name   string
		output string
		want   string
		wantOK bool
	}{
		{
			name:   "standard masq rule",
			output: `-A KUBE-SERVICES ! -s 10.233.64.0/18 -d 10.233.0.1/32 -p tcp -j KUBE-MARK-MASQ`,
			want:   "10.233.64.0/18",
			wantOK: true,
		},
		{
			name:   "no masq line",
			output: `-A KUBE-SERVICES -d 10.233.0.1/32 -j KUBE-SVC-XYZ`,
			want:   "",
			wantOK: false,
		},
		{
			name:   "empty",
			output: "",
			want:   "",
			wantOK: false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := parsePodCIDRFromKubeServicesIPTables(c.output)
			if got != c.want || ok != c.wantOK {
				t.Fatalf("parsePodCIDRFromKubeServicesIPTables = (%q,%v), want (%q,%v)", got, ok, c.want, c.wantOK)
			}
		})
	}
}
