package utils

import (
	"testing"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/stretchr/testify/assert"
)

const gi = 1024 * 1024 // KiB in a GiB

// TestDeclaredAgent covers the role layouts olares-cli actually produces. It is
// the half of NodeRunsControlPlane that needs no host to answer, and the only
// half that can settle the question on its own: a host's roles state the
// topology the installer intends rather than what the node is running, so only
// worker-without-master is conclusive and everything else goes on to be
// confirmed against the node's etcd unit.
func TestDeclaredAgent(t *testing.T) {
	tests := []struct {
		name     string
		roles    []string
		expected bool
	}{
		{
			// The one conclusive case: joining a worker leaves master.conf in
			// place, so the loader's multi-node branch gives the local host
			// worker without master. No round trip is spent on it.
			name:     "a joined worker is the only certain agent",
			roles:    []string{common.Worker, common.K8s},
			expected: true,
		},
		{
			// pkg/pipelines/prepare_system.go clears the master config first, so
			// pkg/common/loader.go takes its single-node branch and hands out
			// every role — on a machine that may well be destined to be a worker.
			// This is the layout the etcd unit has to arbitrate.
			name:     "prepare and single-node install carry every role",
			roles:    []string{common.Master, common.Worker, common.ETCD, common.Registry, common.K8s},
			expected: false,
		},
		{
			// GetLocalHost() falls through to a host with no roles when the OS
			// hostname stops matching the configured one.
			name:     "an unidentified host is not conclusively an agent",
			roles:    nil,
			expected: false,
		},
		{
			name:     "a master that is not also a worker",
			roles:    []string{common.Master, common.ETCD, common.K8s},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host := connector.NewHost()
			for _, role := range tt.roles {
				host.SetRole(role)
			}
			assert.Equal(t, tt.expected, declaredAgent(host))
		})
	}
}

// TestKubeReservedMemoryMi checks the bucket that holds the Kubernetes daemons:
// flat in the size of the machine, and the only one that varies by role.
func TestKubeReservedMemoryMi(t *testing.T) {
	assert.Equal(t, int64(k3sServerMemoryMi+containerdMemoryMi+etcdMemoryMi), kubeReservedMemoryMi(true))
	assert.Equal(t, int64(k3sAgentMemoryMi+containerdMemoryMi), kubeReservedMemoryMi(false))

	// An agent costs an order of magnitude less than a server, which is the whole
	// reason for telling the two apart.
	assert.Greater(t, kubeReservedMemoryMi(true), kubeReservedMemoryMi(false))
}

// TestEtcdMemoryMi pins that only a control plane pays for the datastore. An
// agent has no etcd.service, so reserving or protecting for one there would take
// memory from pods to guard a unit that does not exist.
func TestEtcdMemoryMi(t *testing.T) {
	assert.Equal(t, int64(etcdMemoryMi), EtcdMemoryMi(true))
	assert.Zero(t, EtcdMemoryMi(false))

	// etcd's working set is several times its heap, most of it the bbolt database
	// and write-ahead log it reads through mmap. Protecting only the heap would
	// leave the pages it fsyncs against exposed to reclaim, so the value has to
	// clear a whole working set rather than an anonymous-memory figure.
	assert.Greater(t, EtcdMemoryMi(true), int64(500),
		"must cover etcd's working set, not just its heap")
}

func TestSystemReservedMemoryMi(t *testing.T) {
	tests := []struct {
		name        string
		memTotalKiB int64
		zramKiB     int64
		expectedMi  int64
	}{
		{
			// 2% of 16Gi is 327Mi, under the kernel floor, so the floor decides.
			name:        "a small machine is held up by the kernel floor",
			memTotalKiB: 16 * gi,
			expectedMi:  otherHostDaemonsMemoryMi + minKernelMemoryMi,
		},
		{
			// 2% of 96Gi is 1966Mi, past the floor, so the kernel's share starts
			// tracking the machine: roughly half of it covers what the kernel
			// holds unreclaimably and the rest is page cache headroom.
			name:        "a large machine scales past the floor",
			memTotalKiB: 96 * gi,
			expectedMi:  otherHostDaemonsMemoryMi + 96*1024*2/100,
		},
		{
			// A swap file on a disk costs no RAM, so nothing is added for one.
			// The previous arithmetic counted the swap device into the reserve,
			// spending capacity on disk instead of on anything the host holds.
			name:        "a disk swap file adds nothing",
			memTotalKiB: 46 * gi,
			zramKiB:     0,
			expectedMi:  otherHostDaemonsMemoryMi + minKernelMemoryMi,
		},
		{
			// ZRAM holds its pages in RAM and Olares sets no mem_limit, so pages
			// incompressible enough cost the machine the device's full size.
			name:        "a small ZRAM device is reserved in full",
			memTotalKiB: 96 * gi,
			zramKiB:     2 * gi,
			expectedMi:  otherHostDaemonsMemoryMi + 96*1024*2/100 + 2*1024,
		},
		{
			// Olares sizes ZRAM at half of RAM, which is why this term is capped
			// on its own instead of being allowed into the total.
			name:        "the default half-of-RAM ZRAM device is capped",
			memTotalKiB: 96 * gi,
			zramKiB:     48 * gi,
			expectedMi:  otherHostDaemonsMemoryMi + 96*1024*2/100 + 96*1024*5/100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedMi,
				systemReservedMemoryMi(tt.memTotalKiB, tt.zramKiB))
		})
	}
}

// TestReserveCoversTypicalNodes is the check that matters in practice: across the
// range of machines Olares runs on, the reserve has to come out above what a node
// of that shape holds outside kubepods.slice, which is the host daemons plus the
// memory the kernel will not give back.
//
// The floors below are the host-side footprint such a node reaches once its
// cluster is populated, and they grow with the machine because pod count does.
func TestReserveCoversTypicalNodes(t *testing.T) {
	tests := []struct {
		name       string
		memTotalGi int64
		runsCP     bool
		footprint  int64 // host daemon anon + unreclaimable kernel, in MiB
	}{
		{"a small control plane", 16, true, 2200},
		{"a mid-sized control plane", 48, true, 2700},
		{"a large control plane", 96, true, 3300},
		{"an agent", 32, false, 400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total := totalReservedMemoryMi(tt.memTotalGi*gi, 0, tt.runsCP)
			assert.GreaterOrEqual(t, total, tt.footprint,
				"%s (%dGi) reserves %dMi against a host-side footprint of %dMi",
				tt.name, tt.memTotalGi, total, tt.footprint)
		})
	}
}

// TestReserveCoversProtection pins the invariant that makes the reserve and the
// systemd drop-ins one policy instead of two: system.slice is told to hold
// HostDaemonsMemoryMi, so the reserve must be at least that much. Were it less,
// the kernel would be protecting memory the scheduler had already promised to
// pods, and cgroup v2 answers that with sustained OOM kills.
func TestReserveCoversProtection(t *testing.T) {
	for _, runsCP := range []bool{true, false} {
		for _, memGi := range []int64{8, 16, 32, 64, 96, 256} {
			for _, zramGi := range []int64{0, 4, memGi / 2} {
				total := totalReservedMemoryMi(memGi*gi, zramGi*gi, runsCP)
				assert.GreaterOrEqual(t, total, HostDaemonsMemoryMi(runsCP),
					"controlPlane=%v mem=%dGi zram=%dGi: reserve %dMi is below the %dMi system.slice is asked to protect",
					runsCP, memGi, zramGi, total, HostDaemonsMemoryMi(runsCP))
			}
		}
	}
}

// TestHostDaemonsCoversItsChildren checks the other half of the same hierarchy:
// the two services protected inside system.slice may not together ask for more
// than system.slice itself protects, or the kernel prorates them down and the
// values stop meaning what they say.
func TestHostDaemonsCoversItsChildren(t *testing.T) {
	for _, runsCP := range []bool{true, false} {
		children := K3sMemoryMi(runsCP) + ContainerdMemoryMi() + EtcdMemoryMi(runsCP)
		assert.LessOrEqual(t, children, HostDaemonsMemoryMi(runsCP),
			"controlPlane=%v: the protected units ask for %dMi inside a %dMi slice",
			runsCP, children, HostDaemonsMemoryMi(runsCP))
	}
}

func TestParseZramSwapKiB(t *testing.T) {
	tests := []struct {
		name       string
		procSwaps  string
		expectedKi int64
	}{
		{
			// The common case: Olares sets up a swap file, not a ZRAM device.
			name: "a disk swap file is not ZRAM",
			procSwaps: "Filename\t\t\t\tType\t\tSize\t\tUsed\t\tPriority\n" +
				"/swap.img                               file\t\t8388604\t\t678496\t\t-2\n",
			expectedKi: 0,
		},
		{
			name: "the device olares-cli configures",
			procSwaps: "Filename\t\t\t\tType\t\tSize\t\tUsed\t\tPriority\n" +
				"/dev/zram0                              partition\t8388604\t\t0\t\t100\n",
			expectedKi: 8388604,
		},
		{
			// The header must not be mistaken for a device, and a file alongside
			// ZRAM must not be counted.
			name: "only the ZRAM devices are summed",
			procSwaps: "Filename\t\t\t\tType\t\tSize\t\tUsed\t\tPriority\n" +
				"/dev/zram0                              partition\t1048576\t\t0\t\t100\n" +
				"/swap.img                               file\t\t8388604\t\t0\t\t-2\n" +
				"/dev/zram1                              partition\t1048576\t\t0\t\t100\n",
			expectedKi: 2097152,
		},
		{
			name:       "no swap at all",
			procSwaps:  "Filename\t\t\t\tType\t\tSize\t\tUsed\t\tPriority\n",
			expectedKi: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedKi, parseZramSwapKiB(tt.procSwaps))
		})
	}
}

func TestParseMemTotal(t *testing.T) {
	memTotal, err := parseMemTotal("MemTotal:       98231692 kB\nMemFree:         1234 kB\n")
	assert.NoError(t, err)
	assert.Equal(t, int64(98231692), memTotal)

	_, err = parseMemTotal("MemFree:         1234 kB\n")
	assert.Error(t, err)

	_, err = parseMemTotal("MemTotal:       not-a-number kB\n")
	assert.Error(t, err)
}
