package utils

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/jaypipes/ghw"
	"github.com/jaypipes/ghw/pkg/pci"
	"github.com/jaypipes/pcidb"
)

func TestCollectGpuInfos_PutsNvidiaCardsFirst(t *testing.T) {
	cards := []*ghw.GraphicsCard{
		gpuCard("Intel Corporation", "Arc Graphics"),
		gpuCard("NVIDIA Corporation", "GeForce RTX 4070"),
		gpuCard("AMD", "Radeon RX 7900"),
		gpuCard("NVIDIA Corporation", "GeForce RTX 4090"),
	}

	got := collectGpuInfos(cards)
	want := []string{
		"NVIDIA Corporation GeForce RTX 4070",
		"NVIDIA Corporation GeForce RTX 4090",
		"Intel Corporation Arc Graphics",
		"AMD Radeon RX 7900",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("collectGpuInfos() = %#v, want %#v", got, want)
	}
}

func TestCollectGpuInfos_SkipsIncompleteCards(t *testing.T) {
	cards := []*ghw.GraphicsCard{
		nil,
		{DeviceInfo: nil},
		{DeviceInfo: &pci.Device{Vendor: &pcidb.Vendor{Name: "Intel"}}},
		{DeviceInfo: &pci.Device{Product: &pcidb.Product{Name: "Arc"}}},
		gpuCard("AMD", "Radeon RX 7900"),
	}

	got := collectGpuInfos(cards)
	want := []string{"AMD Radeon RX 7900"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("collectGpuInfos() = %#v, want %#v", got, want)
	}
}

func TestCollectGpuInfos_Empty(t *testing.T) {
	got := collectGpuInfos(nil)
	if got == nil {
		t.Fatal("collectGpuInfos(nil) returned nil, want empty slice")
	}
	if len(got) != 0 {
		t.Fatalf("collectGpuInfos(nil) = %#v, want empty slice", got)
	}
}

func TestFilterSriovVirtualFunctions_RemovesVF(t *testing.T) {
	sysfsDevices := t.TempDir()
	vfAddress := "0000:00:02.1"
	vfPath := filepath.Join(sysfsDevices, vfAddress)
	if err := os.Mkdir(vfPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("../0000:00:02.0", filepath.Join(vfPath, "physfn")); err != nil {
		t.Fatal(err)
	}

	pf := gpuCard("Intel Corporation", "Arrow Lake-S Graphics")
	pf.Address = "0000:00:02.0"
	vf := gpuCard("Intel Corporation", "Arrow Lake-S Graphics")
	vf.Address = vfAddress

	got := filterSriovVirtualFunctions([]*ghw.GraphicsCard{pf, vf}, sysfsDevices)
	want := []*ghw.GraphicsCard{pf}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("filterSriovVirtualFunctions() = %#v, want %#v", got, want)
	}
}

func TestFilterSriovVirtualFunctions_KeepsUnconfirmedCards(t *testing.T) {
	sysfsDevices := t.TempDir()
	pf := gpuCard("Intel Corporation", "Arrow Lake-S Graphics")
	pf.Address = "0000:00:02.0"
	unknown := gpuCard("NVIDIA Corporation", "GeForce RTX 4070")
	unknown.Address = ""

	got := filterSriovVirtualFunctions([]*ghw.GraphicsCard{pf, unknown}, sysfsDevices)
	want := []*ghw.GraphicsCard{pf, unknown}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("filterSriovVirtualFunctions() = %#v, want %#v", got, want)
	}
}

func gpuCard(vendor, product string) *ghw.GraphicsCard {
	return &ghw.GraphicsCard{
		DeviceInfo: &pci.Device{
			Vendor:  &pcidb.Vendor{Name: vendor},
			Product: &pcidb.Product{Name: product},
		},
	}
}
