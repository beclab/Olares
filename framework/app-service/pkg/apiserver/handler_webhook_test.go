package apiserver

import (
	"strconv"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/compute"
)

func TestHamiGPUMemoryLimit(t *testing.T) {
	const gi int64 = 1024 * 1024 * 1024

	tests := []struct {
		name      string
		req       compute.Requirement
		wantBytes int64
		wantNil   bool
	}{
		{
			name: "limitedGpu set and greater than requiredGpu",
			req: compute.Requirement{
				RequiredGPU: 1 * gi,
				LimitedGPU:  24 * gi,
			},
			wantBytes: 24 * gi,
		},
		{
			name: "limitedGpu unset falls back to requiredGpu",
			req: compute.Requirement{
				RequiredGPU: 8 * gi,
			},
			wantBytes: 8 * gi,
		},
		{
			name:    "requiredGpu zero and limitedGpu zero returns nil",
			req:     compute.Requirement{},
			wantNil: true,
		},
		{
			name: "limitedGpu equal to requiredGpu",
			req: compute.Requirement{
				RequiredGPU: 12 * gi,
				LimitedGPU:  12 * gi,
			},
			wantBytes: 12 * gi,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hamiGPUMemoryLimit(tt.req)
			if tt.wantNil {
				if got != nil {
					t.Fatalf("hamiGPUMemoryLimit() = %q, want nil", *got)
				}
				return
			}
			if got == nil {
				t.Fatal("hamiGPUMemoryLimit() = nil, want non-nil")
			}
			want := strconv.FormatInt(tt.wantBytes/1024/1024, 10)
			if *got != want {
				t.Fatalf("hamiGPUMemoryLimit() = %q, want %q", *got, want)
			}
		})
	}
}
