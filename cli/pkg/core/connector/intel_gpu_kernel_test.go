package connector

import "testing"

func TestIntelGPUMinKernel(t *testing.T) {
	tests := []struct {
		name      string
		pciID     string
		wantMin   KernelVersion
		wantOOT   bool
		wantFound bool
	}{
		// Initial support == "Unavailable" -> use Full support number.
		{name: "unavailable uses full (i915)", pciID: "A780", wantMin: KernelVersion{5, 17}, wantFound: true},
		{name: "unavailable uses full paren 6.9", pciID: "7D51", wantMin: KernelVersion{6, 9}, wantFound: true},
		{name: "unavailable uses full 6.17 (xe)", pciID: "E223", wantMin: KernelVersion{6, 17}, wantFound: true},
		{name: "unavailable full 6.15 (xe)", pciID: "E211", wantMin: KernelVersion{6, 15}, wantFound: true},

		// Initial support present -> use Initial support number.
		{name: "initial kernel 6.0", pciID: "7D55", wantMin: KernelVersion{6, 0}, wantFound: true},
		{name: "initial kernel 5.19", pciID: "5690", wantMin: KernelVersion{5, 19}, wantFound: true},
		{name: "initial kernel 6.11 (xe)", pciID: "E20B", wantMin: KernelVersion{6, 11}, wantFound: true},

		// Bare Full support number (no Ubuntu prefix / parentheses).
		{name: "bare full 5.13", pciID: "468A", wantMin: KernelVersion{5, 16}, wantFound: true}, // Unavailable -> full 5.16
		{name: "initial 5.9 not bare full", pciID: "4C8A", wantMin: KernelVersion{5, 9}, wantFound: true},
		{name: "initial 5.4 not bare full 5.7", pciID: "9A59", wantMin: KernelVersion{5, 4}, wantFound: true},

		// Multi-id row members resolve individually.
		{name: "multi-id member A788", pciID: "A788", wantMin: KernelVersion{5, 17}, wantFound: true},
		{name: "multi-id member B084 (xe)", pciID: "B084", wantMin: KernelVersion{6, 13}, wantFound: true},

		// Case-insensitive / whitespace-trimmed input.
		{name: "lowercase input", pciID: "7d55", wantMin: KernelVersion{6, 0}, wantFound: true},
		{name: "padded input", pciID: "  7D55  ", wantMin: KernelVersion{6, 0}, wantFound: true},

		// Out-of-tree rows.
		{name: "out-of-tree ponte vecchio", pciID: "0BD5", wantOOT: true, wantFound: true},
		{name: "out-of-tree flex", pciID: "56C1", wantOOT: true, wantFound: true},

		// Unknown id.
		{name: "unknown id", pciID: "FFFF", wantFound: false},
		{name: "empty id", pciID: "", wantFound: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMin, gotOOT, gotFound := IntelGPUMinKernel(tt.pciID)
			if gotFound != tt.wantFound {
				t.Fatalf("found = %v, want %v", gotFound, tt.wantFound)
			}
			if gotOOT != tt.wantOOT {
				t.Fatalf("outOfTree = %v, want %v", gotOOT, tt.wantOOT)
			}
			if tt.wantFound && !tt.wantOOT && gotMin != tt.wantMin {
				t.Fatalf("min = %s, want %s", gotMin, tt.wantMin)
			}
		})
	}
}

func TestParseKernelToken(t *testing.T) {
	tests := []struct {
		in      string
		want    KernelVersion
		wantErr bool
	}{
		{in: "Kernel 6.0", want: KernelVersion{6, 0}},
		{in: "Kernel 5.19", want: KernelVersion{5, 19}},
		{in: "Ubuntu 24.04+ (6.9)", want: KernelVersion{6, 9}},
		{in: "Ubuntu 22.04+ (5.17)", want: KernelVersion{5, 17}},
		{in: "5.13", want: KernelVersion{5, 13}},
		{in: "5.7", want: KernelVersion{5, 7}},
		{in: "nonsense", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := parseKernelToken(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Fatalf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestKernelVersionCompare(t *testing.T) {
	tests := []struct {
		a, b KernelVersion
		want int
	}{
		{KernelVersion{6, 0}, KernelVersion{6, 0}, 0},
		{KernelVersion{5, 19}, KernelVersion{6, 0}, -1},
		{KernelVersion{6, 0}, KernelVersion{5, 19}, 1},
		{KernelVersion{6, 9}, KernelVersion{6, 17}, -1},
		{KernelVersion{6, 17}, KernelVersion{6, 9}, 1},
		{KernelVersion{6, 2}, KernelVersion{6, 2}, 0},
	}
	for _, tt := range tests {
		if got := tt.a.Compare(tt.b); got != tt.want {
			t.Errorf("%s.Compare(%s) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}
