package utils

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
)

// What a node holds back is the sum of the things that live outside
// kubepods.slice, each term given the shape its cost actually has.
//
// A share of RAM is the wrong shape for most of them. A k3s server costs about
// the same on a small machine as on a large one, so a percentage under-reserves
// small machines badly; and because a percentage and a floor can only be
// combined with max(), on a large machine the percentage has to cover the
// constant daemon cost as well, leaving less for the kernel exactly where slab
// is biggest. Adding the parts up instead lets each one be what it is: flat for
// the daemons, growing with the machine for the kernel.
const (
	// k3sServerMemoryMi carries the API server, scheduler, controller managers
	// and kubelet. This is the one daemon cost that does creep up with the size
	// of the cluster, so it is set above where a large one lands rather than at
	// where a small one does. etcd is not in here: Olares runs it as its own
	// unit, see below.
	k3sServerMemoryMi = 2048

	// k3sAgentMemoryMi is kubelet and kube-proxy only, an order of magnitude
	// less than a server.
	k3sAgentMemoryMi = 256

	// containerdMemoryMi covers containerd and every containerd-shim, which
	// share its cgroup, so this one does grow with pod count.
	containerdMemoryMi = 512

	// etcdMemoryMi covers the datastore, which on Olares is a standalone
	// etcd.service rather than the one embedded in k3s: k3s-server holds client
	// connections to 127.0.0.1:2379 and /registry lives in that etcd, while
	// k3s's own db/etcd directory holds nothing but a name file.
	//
	// Sized against etcd's whole working set rather than its heap, which is the
	// smaller part of it: the bbolt database and the write-ahead log are read
	// through mmap, so they are file-backed, and losing them to reclaim puts a
	// disk read in front of every write etcd fsyncs.
	etcdMemoryMi = 512

	// otherHostDaemonsMemoryMi is journald, sshd, olaresd and the JuiceFS client.
	//
	// They are reserved for but deliberately not protected: none of them is on
	// the path that relieves memory pressure, so losing one to swap costs
	// latency rather than the cluster. Note that sitting inside a protected
	// system.slice does not protect them by itself, since a child with
	// memory.min=0 draws nothing from protection its parent has not handed down.
	otherHostDaemonsMemoryMi = 256

	// minKernelMemoryMi and kernelMemoryPercent size slab, page tables and enough
	// page cache that the host is not reclaiming its own working set. Two points
	// of RAM is roughly double what the kernel holds unreclaimably, and the
	// difference is the page cache headroom.
	minKernelMemoryMi   = 1024
	kernelMemoryPercent = 2

	// maxZramReservePercent caps the ZRAM term on its own rather than capping the
	// total, which a cap on the total cannot do without cutting into the terms
	// above. Olares sizes a ZRAM device at half of RAM, so uncapped this one
	// term would dominate everything else.
	maxZramReservePercent = 5
)

// etcdUnitPath is the unit pkg/etcd/module.go's GenerateETCDService writes, and
// the evidence NodeRunsControlPlane confirms a declared master against.
const etcdUnitPath = "/etc/systemd/system/etcd.service"

// NodeRunsControlPlane reports whether a node should be sized for a k3s server.
//
// The roles on a host describe the topology the installer intends rather than
// what the node is running, and they are conclusive in one direction only.
// Worker without master is certain: that layout comes from the multi-node branch
// of pkg/common/loader.go, which is reached only when a master address was
// configured, meaning this host was told to join one.
//
// Master on its own is the weaker claim, because routine commands hand it to
// machines that are not one:
//
//   - olares-cli prepare clears the master config before building the runtime
//     (pkg/pipelines/prepare_system.go), so the loader takes its single-node
//     branch and gives the local host every role, master included — on every
//     machine, including one about to become a worker.
//   - A worker whose master.conf has gone missing falls back to that same
//     all-roles layout, and that is what upgrade then sees.
//   - GetLocalHost() hands back a host with no roles at all when the OS hostname
//     does not match the configured one.
//
// A declared master is therefore confirmed against the node itself. Olares always
// installs its own etcd, always as a unit of its own, and only on control-plane
// nodes — GenerateETCDService in pkg/etcd/module.go targets the etcd role alone.
// It is also in place before every path that reaches this code:
// etcd.ConfigureModule precedes k3s.InitClusterModule, k3s.JoinNodesModule and
// kubernetes.InitKubernetesModule in pkg/phase/cluster/create_cluster.go, while
// the change-ip and upgrade paths only regenerate the unit of a node that is
// already installed. So a missing etcd.service means this node is not a control
// plane.
//
// Failing to look, on the other hand, says nothing either way, and is answered
// with control plane because the two mistakes are not equal: sizing an agent as a
// server costs that worker schedulable memory, while sizing a server as an agent
// under-reserves the control plane by the same amount, and that is the failure
// this calculation exists to prevent.
func NodeRunsControlPlane(runtime connector.Runtime) bool {
	if declaredAgent(runtime.RemoteHost()) {
		return false
	}
	exists, err := runtime.GetRunner().FileExist(etcdUnitPath)
	if err != nil {
		logger.Warnf("failed to look for %s, sizing this node as a control plane: %v", etcdUnitPath, err)
		return true
	}
	return exists
}

// declaredAgent reports whether a host's roles say it is certainly not a control
// plane. It is the half of NodeRunsControlPlane that needs no host to answer.
func declaredAgent(host connector.Host) bool {
	return host.IsRole(common.Worker) && !host.IsRole(common.Master)
}

// HostDaemonsMemoryMi is the part of the reserve that host daemons hold as
// anonymous memory, and so the value for system.slice's memory.min.
//
// It deliberately excludes the kernel's share: slab and page tables are charged
// to no cgroup, so protecting them through system.slice is not possible and
// counting them there would only overcommit the protection.
func HostDaemonsMemoryMi(runsControlPlane bool) int64 {
	return K3sMemoryMi(runsControlPlane) + containerdMemoryMi +
		EtcdMemoryMi(runsControlPlane) + otherHostDaemonsMemoryMi
}

// EtcdMemoryMi is the memory.min for etcd.service, and zero on a node that does
// not run one.
//
// Of everything protected here this is the one least able to tolerate being
// swapped. etcd fsyncs its write-ahead log on the path of every write and holds
// a 100ms heartbeat against a 1s election timeout, so paging it back in does not
// slow the cluster down so much as it takes the API server away: the leader
// misses heartbeats, an election starts, and every write blocks until it ends.
//
// It also cannot be left out of the set while its siblings are in it. Protection
// does not reduce how much the kernel has to reclaim, only where it takes it
// from, so protecting k3s and containerd and not etcd aims reclaim at etcd.
func EtcdMemoryMi(runsControlPlane bool) int64 {
	if !runsControlPlane {
		return 0
	}
	return etcdMemoryMi
}

// kernelMemoryMi is the kernel's share of the reserve: slab, page tables and
// enough page cache for the host not to reclaim its own working set.
//
// This is the one part that tracks the size of the machine, because slab grows
// with the number of pods and pod count grows with RAM.
func kernelMemoryMi(memTotalMi int64) int64 {
	if scaled := memTotalMi * kernelMemoryPercent / 100; scaled > minKernelMemoryMi {
		return scaled
	}
	return minKernelMemoryMi
}

// K3sMemoryMi is the memory.min for k3s.service. A server carries the control
// plane; an agent is only kubelet and kube-proxy.
func K3sMemoryMi(runsControlPlane bool) int64 {
	if runsControlPlane {
		return k3sServerMemoryMi
	}
	return k3sAgentMemoryMi
}

// ContainerdMemoryMi is the memory.min for containerd.service.
//
// Every containerd-shim shares this cgroup, and a shim is what actually stops a
// container, so this is on the path kubelet takes to evict. Letting the shims be
// swapped out is what turns a node under pressure into one that has decided to
// evict but cannot carry it out.
func ContainerdMemoryMi() int64 {
	return containerdMemoryMi
}

// KubeReservedMemory returns the value for kubelet's kube-reserved memory, as a
// Kubernetes quantity string: k3s and containerd, the Kubernetes daemons.
//
// kubelet adds kube-reserved and system-reserved together and subtracts the sum
// from capacity (GetNodeAllocatableReservation in
// pkg/kubelet/cm/node_container_manager_linux.go), so which of the two carries a
// given megabyte changes nothing. They become separate cgroup limits only when
// --enforce-node-allocatable names them and --kube-reserved-cgroup points at a
// cgroup, neither of which Olares sets. The split is therefore documentation,
// and it is worth keeping honest for the next person reading the flags.
//
// Enforcing them is deliberately not done: enforcement means a memory.max on
// those cgroups, which would turn the memory.min protection installed beside
// this into a hard cap that OOM-kills k3s instead of keeping it resident.
func KubeReservedMemory(runsControlPlane bool) string {
	return fmt.Sprintf("%dMi", kubeReservedMemoryMi(runsControlPlane))
}

// SystemReservedMemory returns the value for kubelet's system-reserved memory,
// as a Kubernetes quantity string: the OS side of the reserve, meaning the
// remaining host daemons, the kernel, and swap that pods are not allowed to use.
//
// What this replaces was a share of RAM — three points of it, capped at five and
// floored at 500Mi — which under-reserved every machine small enough for the
// floor to decide, since a control plane's host side runs to gigabytes there just
// as it does anywhere else. Where the share did reach a large enough number it
// got there partly by counting the swap device, so much of the reserve stood for
// disk rather than for anything the host holds in RAM.
//
// Under-reserving is not merely an accounting error. The reserve is what sets
// kubepods.slice's memory.max, so a node that holds back less than its host side
// needs has licensed pods to grow into memory the host is using, and it can then
// commit itself past the point where reclaim keeps up. That state is
// unrecoverable in a way a per-pod OOM kill is not: every cgroup is still inside
// its own limit, so nothing in the accounting says anything is wrong and nothing
// gets chosen for eviction.
//
// Nothing extra is held back for GPU nodes. The host pages behind a managed
// allocation are charged to the cgroup that asked for them, so a pod using
// unified memory counts against its own limit and against kubepods.slice like
// any other pod, and the eviction signals see it.
//
// Only ZRAM swap is reserved for, and only when pods may not use it. A ZRAM
// device holds its pages in RAM, Olares sizes it at half of physical memory and
// sets no mem_limit (pkg/bootstrap/os/templates/swap.go), so pages that are
// incompressible enough can cost the machine that entire half — real RAM that no
// longer holds anything schedulable. A swap file on a disk takes no RAM at all,
// and kubelet never counts swap as available either way, so reserving against
// one only throws capacity away, by as much as the file is large.
//
// Determining the host's memory needs a command on that host, so a failure here
// is reported as a warning and the part of the reserve that can be had without
// reaching it is returned, rather than failing the install outright.
func SystemReservedMemory(runtime connector.Runtime, podSwapEnabled bool) string {
	out, err := runtime.GetRunner().Cmd("cat /proc/meminfo", false, false)
	if err == nil {
		var memTotal int64
		memTotal, err = parseMemTotal(out)
		if err == nil {
			var zramKiB int64
			if !podSwapEnabled {
				zramKiB = zramSwapKiB(runtime)
			}
			return fmt.Sprintf("%dMi", systemReservedMemoryMi(memTotal, zramKiB))
		}
	}
	// Falling back to the old flat 250Mi would leave the node in the state this
	// calculation exists to prevent. Everything but the swap term can be had
	// without reaching the host, so fall back to that.
	fallbackMi := otherHostDaemonsMemoryMi + minKernelMemoryMi
	logger.Warnf("failed to read host memory, reserving %dMi for the system: %v", fallbackMi, err)
	return fmt.Sprintf("%dMi", fallbackMi)
}

// kubeReservedMemoryMi is the Kubernetes daemons' share, which does not depend on
// the size of the machine.
func kubeReservedMemoryMi(runsControlPlane bool) int64 {
	return K3sMemoryMi(runsControlPlane) + containerdMemoryMi + EtcdMemoryMi(runsControlPlane)
}

// systemReservedMemoryMi is the arithmetic behind SystemReservedMemory, in MiB.
// memTotalKiB and zramKiB are in KiB, the unit both /proc files report; zramKiB
// is expected to be zero when pods are allowed to swap.
//
// The role does not appear here: only the Kubernetes daemons differ between a
// server and an agent, and the OS side of the reserve does not.
func systemReservedMemoryMi(memTotalKiB, zramKiB int64) int64 {
	memMi := memTotalKiB / 1024
	totalMi := otherHostDaemonsMemoryMi + kernelMemoryMi(memMi)

	zramMi := zramKiB / 1024
	// Capped on its own rather than as part of the total, which could only be
	// done by cutting into the terms it is added to.
	if capMi := memMi * maxZramReservePercent / 100; zramMi > capMi {
		zramMi = capMi
	}
	return totalMi + zramMi
}

// totalReservedMemoryMi is what the node holds back altogether, across both
// buckets. Nothing configures kubelet with this; it exists so the invariant
// against HostDaemonsMemoryMi can be stated and tested in one place.
func totalReservedMemoryMi(memTotalKiB, zramKiB int64, runsControlPlane bool) int64 {
	return kubeReservedMemoryMi(runsControlPlane) +
		systemReservedMemoryMi(memTotalKiB, zramKiB)
}

// zramSwapKiB returns the combined size of the node's ZRAM swap devices, in KiB.
//
// A failure to read /proc/swaps is reported and treated as no ZRAM, both because
// that is the common configuration and because refusing to install over it would
// be a worse trade than reserving nothing extra.
func zramSwapKiB(runtime connector.Runtime) int64 {
	out, err := runtime.GetRunner().Cmd("cat /proc/swaps", false, false)
	if err != nil {
		logger.Warnf("failed to read the host's swap devices, reserving nothing for ZRAM: %v", err)
		return 0
	}
	return parseZramSwapKiB(out)
}

// parseZramSwapKiB sums the sizes of the ZRAM devices in /proc/swaps, whose
// columns are Filename, Type, Size, Used and Priority, with Size in KiB.
func parseZramSwapKiB(procSwaps string) int64 {
	var total int64
	scanner := bufio.NewScanner(strings.NewReader(procSwaps))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 3 || !strings.HasPrefix(fields[0], "/dev/zram") {
			continue
		}
		size, err := strconv.ParseInt(fields[2], 10, 64)
		if err != nil {
			logger.Warnf("unexpected size for swap device %s in /proc/swaps: %q", fields[0], fields[2])
			continue
		}
		total += size
	}
	return total
}

// parseMemTotal returns MemTotal in KiB, the unit /proc/meminfo reports it in.
func parseMemTotal(out string) (int64, error) {
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		key, value, ok := strings.Cut(scanner.Text(), ":")
		if !ok || key != "MemTotal" {
			continue
		}
		// The value looks like "  98231692 kB".
		fields := strings.Fields(value)
		if len(fields) == 0 {
			continue
		}
		memTotal, err := strconv.ParseInt(fields[0], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("unexpected MemTotal in /proc/meminfo: %q", strings.TrimSpace(value))
		}
		if memTotal <= 0 {
			return 0, fmt.Errorf("no usable MemTotal in /proc/meminfo: %q", strings.TrimSpace(value))
		}
		return memTotal, nil
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	return 0, fmt.Errorf("no MemTotal in /proc/meminfo")
}
