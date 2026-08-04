package connector

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// KernelVersion is a parsed Linux kernel version (major.minor).
type KernelVersion struct {
	Major int
	Minor int
}

// String renders the version as "major.minor" (e.g. "6.9").
func (v KernelVersion) String() string {
	return fmt.Sprintf("%d.%d", v.Major, v.Minor)
}

// semver returns the version as a semver value (patch fixed at 0), so kernel
// versions can be compared with the shared Masterminds/semver library.
func (v KernelVersion) semver() *semver.Version {
	return semver.New(uint64(v.Major), uint64(v.Minor), 0, "", "")
}

// Compare returns -1 if v < o, 0 if v == o, and 1 if v > o.
func (v KernelVersion) Compare(o KernelVersion) int {
	return v.semver().Compare(o.semver())
}

// Sentinel values used in the Intel "Initial support" / "Full support" columns.
const (
	intelSupportUnavailable = "Unavailable"
	intelSupportOutOfTree   = "Out-of-tree"
)

// intelGPUSupportInfo holds the per-GPU columns transcribed verbatim from the
// Intel docs. Only initial/full drive the kernel lookup today; name,
// architecture and codename are carried for potential future use.
type intelGPUSupportInfo struct {
	name         string
	architecture string
	codename     string
	initial      string
	full         string
}

// intelGPUSupportRow groups one or more PCI device ids that share the same
// support columns (multi-id rows in the source tables). name, architecture and
// codename are carried verbatim from the source tables for potential future use
// (e.g. user-facing messages), even though only initial/full drive the kernel
// lookup today.
type intelGPUSupportRow struct {
	ids          []string
	name         string
	architecture string
	codename     string
	initial      string
	full         string
}

// intelGPUSupportRows is transcribed from the two Intel driver-support tables
// (as of 2026-07). Kernel numbers are guidance only: vendor kernels may
// backport support to older kernels, matching the disclaimer on the source
// pages.
//   - i915: https://dgpu-docs.intel.com/overview/supported-hardware/i915-driver-gpus.html
//   - xe:   https://dgpu-docs.intel.com/overview/supported-hardware/xe-driver-gpus.html
var intelGPUSupportRows = []intelGPUSupportRow{
	// i915 KMD driver GPUs
	{ids: []string{"7D51"}, name: "Intel® Graphics", architecture: "Xe-LPG", codename: "Arrow Lake-H", initial: "Unavailable", full: "Ubuntu 24.04+ (6.9)"},
	{ids: []string{"7D67"}, name: "Intel® Graphics", architecture: "Xe-LPG", codename: "Arrow Lake-S", initial: "Unavailable", full: "Ubuntu 24.04+ (6.7)"},
	{ids: []string{"7D41"}, name: "Intel® Graphics", architecture: "Xe-LPG", codename: "Arrow Lake-U", initial: "Unavailable", full: "Ubuntu 24.04+ (6.9)"},
	{ids: []string{"7DD5"}, name: "Intel® Graphics", architecture: "Xe-LPG", codename: "Meteor Lake", initial: "Kernel 6.0", full: "Ubuntu 24.04+ (6.7)"},
	{ids: []string{"7D45"}, name: "Intel® Graphics", architecture: "Xe-LPG", codename: "Meteor Lake", initial: "Kernel 6.0", full: "Ubuntu 24.04+ (6.7)"},
	{ids: []string{"7D40"}, name: "Intel® Graphics", architecture: "Xe-LPG", codename: "Meteor Lake", initial: "Kernel 6.0", full: "Ubuntu 24.04+ (6.7)"},
	{ids: []string{"7D55"}, name: "Intel® Arc™ Graphics", architecture: "Xe-LPG", codename: "Meteor Lake", initial: "Kernel 6.0", full: "Ubuntu 24.04+ (6.7)"},
	{ids: []string{"0BD5"}, name: "Intel® Data Center GPU Max 1550", architecture: "Xe-HPC", codename: "Ponte Vecchio", initial: "Out-of-tree", full: "Out-of-tree"},
	{ids: []string{"0BDA"}, name: "Intel® Data Center GPU Max 1100", architecture: "Xe-HPC", codename: "Ponte Vecchio", initial: "Out-of-tree", full: "Out-of-tree"},
	{ids: []string{"56C0"}, name: "Intel® Data Center GPU Flex 170", architecture: "Xe-HPG", codename: "Alchemist", initial: "Out-of-tree", full: "Out-of-tree"},
	{ids: []string{"56C1"}, name: "Intel® Data Center GPU Flex 140", architecture: "Xe-HPG", codename: "Alchemist", initial: "Out-of-tree", full: "Out-of-tree"},
	{ids: []string{"5690"}, name: "Intel® Arc™ A770M Graphics", architecture: "Xe-HPG", codename: "Alchemist", initial: "Kernel 5.19", full: "Ubuntu 24.04+ (6.2)"},
	{ids: []string{"5691"}, name: "Intel® Arc™ A730M Graphics", architecture: "Xe-HPG", codename: "Alchemist", initial: "Kernel 5.19", full: "Ubuntu 24.04+ (6.2)"},
	{ids: []string{"5696"}, name: "Intel® Arc™ A570M Graphics", architecture: "Xe-HPG", codename: "Alchemist", initial: "Kernel 5.19", full: "Ubuntu 24.04+ (6.2)"},
	{ids: []string{"5692"}, name: "Intel® Arc™ A550M Graphics", architecture: "Xe-HPG", codename: "Alchemist", initial: "Kernel 5.19", full: "Ubuntu 24.04+ (6.2)"},
	{ids: []string{"5697"}, name: "Intel® Arc™ A530M Graphics", architecture: "Xe-HPG", codename: "Alchemist", initial: "Kernel 5.19", full: "Ubuntu 24.04+ (6.2)"},
	{ids: []string{"5693"}, name: "Intel® Arc™ A370M Graphics", architecture: "Xe-HPG", codename: "Alchemist", initial: "Kernel 5.19", full: "Ubuntu 24.04+ (6.2)"},
	{ids: []string{"5694"}, name: "Intel® Arc™ A350M Graphics", architecture: "Xe-HPG", codename: "Alchemist", initial: "Kernel 5.19", full: "Ubuntu 24.04+ (6.2)"},
	{ids: []string{"56A0"}, name: "Intel® Arc™ A770 Graphics", architecture: "Xe-HPG", codename: "Alchemist", initial: "Kernel 6.0", full: "Ubuntu 24.04+ (6.2)"},
	{ids: []string{"56A1"}, name: "Intel® Arc™ A750 Graphics", architecture: "Xe-HPG", codename: "Alchemist", initial: "Kernel 6.0", full: "Ubuntu 24.04+ (6.2)"},
	{ids: []string{"56A2"}, name: "Intel® Arc™ A580 Graphics", architecture: "Xe-HPG", codename: "Alchemist", initial: "Kernel 6.0", full: "Ubuntu 24.04+ (6.2)"},
	{ids: []string{"56A5"}, name: "Intel® Arc™ A380 Graphics", architecture: "Xe-HPG", codename: "Alchemist", initial: "Kernel 6.0", full: "Ubuntu 24.04+ (6.2)"},
	{ids: []string{"56A6"}, name: "Intel® Arc™ A310 Graphics", architecture: "Xe-HPG", codename: "Alchemist", initial: "Kernel 6.0", full: "Ubuntu 24.04+ (6.2)"},
	{ids: []string{"56B3"}, name: "Intel® Arc™ Pro A60 Graphics", architecture: "Xe-HPG", codename: "Alchemist", initial: "Kernel 6.0", full: "Ubuntu 24.04+ (6.2)"},
	{ids: []string{"56B2"}, name: "Intel® Arc™ Pro A60M Graphics", architecture: "Xe-HPG", codename: "Alchemist", initial: "Kernel 5.19", full: "Ubuntu 24.04+ (6.2)"},
	{ids: []string{"56B1"}, name: "Intel® Arc™ Pro A40/A50 Graphics", architecture: "Xe-HPG", codename: "Alchemist", initial: "Kernel 6.0", full: "Ubuntu 24.04+ (6.2)"},
	{ids: []string{"56B0"}, name: "Intel® Arc™ Pro A30M Graphics", architecture: "Xe-HPG", codename: "Alchemist", initial: "Kernel 5.19", full: "Ubuntu 24.04+ (6.2)"},
	{ids: []string{"56BA"}, name: "Intel® Arc™ A380E Graphics", architecture: "Xe-HPG", codename: "Alchemist", initial: "Unavailable", full: "Ubuntu 24.04+ (6.7)"},
	{ids: []string{"56BC"}, name: "Intel® Arc™ A370E Graphics", architecture: "Xe-HPG", codename: "Alchemist", initial: "Unavailable", full: "Ubuntu 24.04+ (6.7)"},
	{ids: []string{"56BD"}, name: "Intel® Arc™ A350E Graphics", architecture: "Xe-HPG", codename: "Alchemist", initial: "Unavailable", full: "Ubuntu 24.04+ (6.7)"},
	{ids: []string{"56BB"}, name: "Intel® Arc™ A310E Graphics", architecture: "Xe-HPG", codename: "Alchemist", initial: "Unavailable", full: "Ubuntu 24.04+ (6.7)"},
	{ids: []string{"A780"}, name: "Intel® UHD Graphics 770", architecture: "Xe", codename: "Raptor Lake-S", initial: "Unavailable", full: "Ubuntu 22.04+ (5.17)"},
	{ids: []string{"A781", "A788", "A789"}, name: "Intel® UHD Graphics", architecture: "Xe", codename: "Raptor Lake-S", initial: "Unavailable", full: "Ubuntu 22.04+ (5.17)"},
	{ids: []string{"A78A"}, name: "Intel® UHD Graphics", architecture: "Xe", codename: "Raptor Lake-S", initial: "Unavailable", full: "Ubuntu 22.04+ (5.19)"},
	{ids: []string{"A782"}, name: "Intel® UHD Graphics 730", architecture: "Xe", codename: "Raptor Lake-S", initial: "Unavailable", full: "Ubuntu 22.04+ (5.17)"},
	{ids: []string{"A78B"}, name: "Intel® UHD Graphics", architecture: "Xe", codename: "Raptor Lake-S", initial: "Unavailable", full: "Ubuntu 22.04+ (5.19)"},
	{ids: []string{"A783"}, name: "Intel® UHD Graphics 710", architecture: "Xe", codename: "Raptor Lake-S", initial: "Unavailable", full: "Ubuntu 22.04+ (5.17)"},
	{ids: []string{"A7A0", "A7A1"}, name: "Intel® Iris® Xe Graphics", architecture: "Xe", codename: "Raptor Lake-P", initial: "Unavailable", full: "Ubuntu 22.04+ (5.19)"},
	{ids: []string{"A7A8"}, name: "Intel® UHD Graphics", architecture: "Xe", codename: "Raptor Lake-P", initial: "Unavailable", full: "Ubuntu 22.04+ (5.19)"},
	{ids: []string{"A7AA"}, name: "Intel® Graphics", architecture: "Xe", codename: "Raptor Lake-P", initial: "Unavailable", full: "Ubuntu 24.04+ (6.7)"},
	{ids: []string{"A7AB"}, name: "Intel® Graphics", architecture: "Xe", codename: "Raptor Lake-P", initial: "Unavailable", full: "Ubuntu 24.04+ (6.7)"},
	{ids: []string{"A7AC"}, name: "Intel® Graphics", architecture: "Xe", codename: "Raptor Lake-U", initial: "Unavailable", full: "Ubuntu 24.04+ (6.7)"},
	{ids: []string{"A7AD"}, name: "Intel® Graphics", architecture: "Xe", codename: "Raptor Lake-U", initial: "Unavailable", full: "Ubuntu 24.04+ (6.7)"},
	{ids: []string{"A7A9"}, name: "Intel® UHD Graphics", architecture: "Xe", codename: "Raptor Lake-P", initial: "Unavailable", full: "Ubuntu 22.04+ (5.19)"},
	{ids: []string{"A721"}, name: "Intel® UHD Graphics", architecture: "Xe", codename: "Raptor Lake-P", initial: "Unavailable", full: "Ubuntu 22.04+ (5.19)"},
	{ids: []string{"4905"}, name: "Intel® Iris® Xe MAX Graphics", architecture: "Xe", codename: "DG1", initial: "Kernel 5.16", full: "Ubuntu 25.10+ (6.17)"},
	{ids: []string{"4907"}, name: "Intel Server GPU SG-18M", architecture: "Xe", codename: "DG1", initial: "Kernel 5.16", full: "Ubuntu 25.10+ (6.17)"},
	{ids: []string{"4908"}, name: "Intel® Iris® Xe Graphics", architecture: "Xe", codename: "DG1", initial: "Kernel 5.16", full: "Ubuntu 25.10+ (6.17)"},
	{ids: []string{"4909"}, name: "Intel® Iris® Xe MAX 100 Graphics", architecture: "Xe", codename: "DG1", initial: "Kernel 5.16", full: "Ubuntu 25.10+ (6.17)"},
	{ids: []string{"4680", "4690"}, name: "Intel® UHD Graphics 770", architecture: "Xe", codename: "Alder Lake-S", initial: "Kernel 5.13", full: "Ubuntu 22.04+ (5.16)"},
	{ids: []string{"4688"}, name: "Intel® UHD Graphics 770", architecture: "Xe", codename: "Alder Lake-S", initial: "Kernel 5.14", full: "Ubuntu 22.04+ (5.16)"},
	{ids: []string{"468A"}, name: "Intel® UHD Graphics 770", architecture: "Xe", codename: "Alder Lake-S", initial: "Unavailable", full: "Ubuntu 22.04+ (5.16)"},
	{ids: []string{"468B"}, name: "Intel® UHD Graphics 770", architecture: "Xe", codename: "Alder Lake-S", initial: "Unavailable", full: "Ubuntu 24.04+ (6.1)"},
	{ids: []string{"4682", "4692"}, name: "Intel® UHD Graphics 730", architecture: "Xe", codename: "Alder Lake-S", initial: "Kernel 5.13", full: "Ubuntu 22.04+ (5.16)"},
	{ids: []string{"4693"}, name: "Intel® UHD Graphics 710", architecture: "Xe", codename: "Alder Lake-S", initial: "Kernel 5.13", full: "Ubuntu 22.04+ (5.16)"},
	{ids: []string{"46D3"}, name: "Intel® Graphics", architecture: "Xe", codename: "Twin Lake", initial: "Unavailable", full: "Ubuntu 24.04+ (6.9)"},
	{ids: []string{"46D4"}, name: "Intel® Graphics", architecture: "Xe", codename: "Twin Lake", initial: "Unavailable", full: "Ubuntu 24.04+ (6.9)"},
	{ids: []string{"46D0"}, name: "Intel® UHD Graphics", architecture: "Xe", codename: "Alder Lake-N", initial: "Unavailable", full: "Ubuntu 22.04+ (5.18)"},
	{ids: []string{"46D1"}, name: "Intel® UHD Graphics", architecture: "Xe", codename: "Alder Lake-N", initial: "Unavailable", full: "Ubuntu 22.04+ (5.18)"},
	{ids: []string{"46D2"}, name: "Intel® UHD Graphics", architecture: "Xe", codename: "Alder Lake-N", initial: "Unavailable", full: "Ubuntu 22.04+ (5.18)"},
	{ids: []string{"4626", "4628", "462A"}, name: "Intel® UHD Graphics", architecture: "Xe", codename: "Alder Lake-P", initial: "Kernel 5.14", full: "Ubuntu 22.04+ (5.17)"},
	{ids: []string{"46A2", "46B3", "46C2"}, name: "Intel® UHD Graphics", architecture: "Xe", codename: "Alder Lake-P", initial: "Kernel 5.14", full: "Ubuntu 22.04+ (5.17)"},
	{ids: []string{"46A3", "46B2", "46C3"}, name: "Intel® UHD Graphics", architecture: "Xe", codename: "Alder Lake-P", initial: "Kernel 5.14", full: "Ubuntu 22.04+ (5.17)"},
	{ids: []string{"46A0", "46B0", "46C0"}, name: "Intel® Iris® Xe Graphics", architecture: "Xe", codename: "Alder Lake-P", initial: "Kernel 5.14", full: "Ubuntu 22.04+ (5.17)"},
	{ids: []string{"46A6", "46AA", "46A8"}, name: "Intel® Iris® Xe Graphics", architecture: "Xe", codename: "Alder Lake-P", initial: "Kernel 5.14", full: "Ubuntu 22.04+ (5.17)"},
	{ids: []string{"46A1", "46B1", "46C1"}, name: "Intel® Iris® Xe Graphics", architecture: "Xe", codename: "Alder Lake-P", initial: "Kernel 5.14", full: "Ubuntu 22.04+ (5.17)"},
	{ids: []string{"4C8A"}, name: "Intel® UHD Graphics 750", architecture: "Xe", codename: "Rocket Lake", initial: "Kernel 5.9", full: "5.13"},
	{ids: []string{"4C8B"}, name: "Intel® UHD Graphics 730", architecture: "Xe", codename: "Rocket Lake", initial: "Kernel 5.9", full: "5.13"},
	{ids: []string{"4C90", "4C9A"}, name: "Intel® UHD Graphics P750", architecture: "Xe", codename: "Rocket Lake", initial: "Kernel 5.9", full: "5.13"},
	{ids: []string{"4E71"}, name: "Intel® UHD Graphics", architecture: "Xe", codename: "Jasper Lake", initial: "Kernel 5.6", full: "Ubuntu 22.04+ (5.15)"},
	{ids: []string{"4E61"}, name: "Intel® UHD Graphics", architecture: "Xe", codename: "Jasper Lake", initial: "Kernel 5.6", full: "Ubuntu 22.04+ (5.15)"},
	{ids: []string{"4E57"}, name: "Intel® UHD Graphics", architecture: "Xe", codename: "Jasper Lake", initial: "Kernel 5.9", full: "Ubuntu 22.04+ (5.15)"},
	{ids: []string{"4E55"}, name: "Intel® UHD Graphics", architecture: "Xe", codename: "Jasper Lake", initial: "Kernel 5.9", full: "Ubuntu 22.04+ (5.15)"},
	{ids: []string{"4E51"}, name: "Intel® UHD Graphics", architecture: "Xe", codename: "Jasper Lake", initial: "Kernel 5.6", full: "Ubuntu 22.04+ (5.15)"},
	{ids: []string{"4557"}, name: "Intel® UHD Graphics", architecture: "Xe", codename: "Elkhart Lake", initial: "Kernel 5.9", full: "Ubuntu 22.04+ (5.15)"},
	{ids: []string{"4555"}, name: "Intel® UHD Graphics", architecture: "Xe", codename: "Elkhart Lake", initial: "Kernel 5.9", full: "Ubuntu 22.04+ (5.15)"},
	{ids: []string{"4571"}, name: "Intel® UHD Graphics", architecture: "Xe", codename: "Elkhart Lake", initial: "Kernel 5.2", full: "Ubuntu 22.04+ (5.15)"},
	{ids: []string{"4551"}, name: "Intel® UHD Graphics", architecture: "Xe", codename: "Elkhart Lake", initial: "Kernel 5.2", full: "Ubuntu 22.04+ (5.15)"},
	{ids: []string{"4541"}, name: "Intel® UHD Graphics", architecture: "Xe", codename: "Elkhart Lake", initial: "Kernel 5.2", full: "Ubuntu 22.04+ (5.15)"},
	{ids: []string{"9A59"}, name: "Intel® UHD Graphics", architecture: "Xe", codename: "Tiger Lake", initial: "Kernel 5.4", full: "5.7"},
	{ids: []string{"9A78"}, name: "Intel® UHD Graphics", architecture: "Xe", codename: "Tiger Lake", initial: "Kernel 5.4", full: "5.7"},
	{ids: []string{"9A60", "9A70"}, name: "Intel® UHD Graphics", architecture: "Xe", codename: "Tiger Lake", initial: "Kernel 5.4", full: "5.7"},
	{ids: []string{"9A68"}, name: "Intel® UHD Graphics", architecture: "Xe", codename: "Tiger Lake", initial: "Kernel 5.4", full: "5.7"},
	{ids: []string{"9A40", "9A49"}, name: "Intel® Iris® Xe Graphics", architecture: "Xe", codename: "Tiger Lake", initial: "Kernel 5.4", full: "5.7"},

	// Xe KMD driver GPUs
	{ids: []string{"E223"}, name: "Intel® Arc™ Pro B70 Graphics", architecture: "Xe2", codename: "Battlemage", initial: "Unavailable", full: "Ubuntu 25.10+ (6.17)"},
	{ids: []string{"E222"}, name: "Intel® Arc™ Pro B65 Graphics", architecture: "Xe2", codename: "Battlemage", initial: "Unavailable", full: "Ubuntu 25.10+ (6.17)"},
	{ids: []string{"B080", "B082", "B084", "B086"}, name: "Intel® Arc™ B390 GPU", architecture: "Xe3", codename: "Panther Lake", initial: "Kernel 6.13", full: "Ubuntu 25.10+ (6.17)"},
	{ids: []string{"B081", "B083", "B085", "B087"}, name: "Intel® Arc™ B370 GPU", architecture: "Xe3", codename: "Panther Lake", initial: "Kernel 6.15", full: "Ubuntu 25.10+ (6.17)"},
	{ids: []string{"B090", "B0A0"}, name: "Intel® Graphics", architecture: "Xe3", codename: "Panther Lake", initial: "Kernel 6.13", full: "Ubuntu 25.10+ (6.17)"},
	{ids: []string{"E212"}, name: "Intel® Arc™ Pro B50 Graphics", architecture: "Xe2", codename: "Battlemage", initial: "Kernel 6.11", full: "Ubuntu 24.04+ (6.14)"},
	{ids: []string{"E211"}, name: "Intel® Arc™ Pro B60 Graphics", architecture: "Xe2", codename: "Battlemage", initial: "Unavailable", full: "Ubuntu 24.04+ (6.15)"},
	{ids: []string{"E20B"}, name: "Intel® Arc™ B580 Graphics", architecture: "Xe2", codename: "Battlemage", initial: "Kernel 6.11", full: "Ubuntu 24.04+ (6.12)"},
	{ids: []string{"E20C"}, name: "Intel® Arc™ B570 Graphics", architecture: "Xe2", codename: "Battlemage", initial: "Kernel 6.11", full: "Ubuntu 24.04+ (6.12)"},
	{ids: []string{"64A0"}, name: "Intel® Arc™ Graphics", architecture: "Xe2", codename: "Lunar Lake", initial: "Kernel 6.8", full: "Ubuntu 24.04+ (6.11)"},
	{ids: []string{"6420"}, name: "Intel® Graphics", architecture: "Xe2", codename: "Lunar Lake", initial: "Kernel 6.8", full: "Ubuntu 24.04+ (6.11)"},
}

// intelGPUSupportByID maps each individual PCI device id (uppercase) to its
// support columns, expanded from intelGPUSupportRows.
var intelGPUSupportByID = func() map[string]intelGPUSupportInfo {
	m := make(map[string]intelGPUSupportInfo)
	for _, r := range intelGPUSupportRows {
		for _, id := range r.ids {
			m[strings.ToUpper(id)] = intelGPUSupportInfo{
				name:         r.name,
				architecture: r.architecture,
				codename:     r.codename,
				initial:      r.initial,
				full:         r.full,
			}
		}
	}
	return m
}()

var (
	kernelParenRe = regexp.MustCompile(`\(([^)]+)\)`)
	kernelNumRe   = regexp.MustCompile(`(\d+)\.(\d+)`)
)

// ParseKernelVersion extracts the major.minor KernelVersion from an arbitrary
// kernel string, e.g. the running kernel "6.8.0-45-generic" -> 6.8.
func ParseKernelVersion(s string) (KernelVersion, error) {
	return parseKernelToken(s)
}

// parseKernelToken extracts a KernelVersion from the three formats observed in
// the Intel tables:
//   - "Kernel 6.0"          -> 6.0
//   - "Ubuntu 24.04+ (6.9)" -> 6.9 (number inside parentheses)
//   - "5.13"                -> 5.13 (bare number)
func parseKernelToken(s string) (KernelVersion, error) {
	s = strings.TrimSpace(s)
	// Prefer the parenthesised kernel number so we don't accidentally match the
	// Ubuntu release (e.g. "24.04") that precedes it.
	if m := kernelParenRe.FindStringSubmatch(s); m != nil {
		s = m[1]
	}
	m := kernelNumRe.FindStringSubmatch(s)
	if m == nil {
		return KernelVersion{}, fmt.Errorf("cannot parse kernel version from %q", s)
	}
	major, err := strconv.Atoi(m[1])
	if err != nil {
		return KernelVersion{}, fmt.Errorf("invalid major version in %q: %w", s, err)
	}
	minor, err := strconv.Atoi(m[2])
	if err != nil {
		return KernelVersion{}, fmt.Errorf("invalid minor version in %q: %w", s, err)
	}
	return KernelVersion{Major: major, Minor: minor}, nil
}

// minKernelForSupport applies the rule:
//   - Initial support is "Unavailable"  -> minimum kernel is the Full support number.
//   - Initial support is "Out-of-tree"  -> outOfTree (no upstream kernel applies).
//   - otherwise                         -> minimum kernel is the Initial support number.
func minKernelForSupport(initial, full string) (min KernelVersion, outOfTree bool, err error) {
	initial = strings.TrimSpace(initial)
	full = strings.TrimSpace(full)

	if strings.EqualFold(initial, intelSupportOutOfTree) || strings.EqualFold(full, intelSupportOutOfTree) {
		return KernelVersion{}, true, nil
	}
	if strings.EqualFold(initial, intelSupportUnavailable) {
		v, err := parseKernelToken(full)
		return v, false, err
	}
	v, err := parseKernelToken(initial)
	return v, false, err
}

// IntelGPUMinKernel returns the minimum upstream Linux kernel version required
// to support the Intel GPU identified by its PCI device id (case-insensitive,
// e.g. "7d55" or "7D55" from lspci's "[8086:7d55]").
//
//   - found=false    -> the id is not present in the i915/xe tables.
//   - outOfTree=true -> the GPU requires the out-of-tree intel-i915-dkms module;
//     no upstream kernel version applies and min is the zero value.
func IntelGPUMinKernel(pciID string) (min KernelVersion, outOfTree bool, found bool) {
	info, ok := intelGPUSupportByID[strings.ToUpper(strings.TrimSpace(pciID))]
	if !ok {
		return KernelVersion{}, false, false
	}
	v, oot, err := minKernelForSupport(info.initial, info.full)
	if err != nil {
		return KernelVersion{}, false, false
	}
	return v, oot, true
}
