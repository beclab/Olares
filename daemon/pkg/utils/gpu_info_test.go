package utils

import (
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

func gpuCard(vendor, product string) *ghw.GraphicsCard {
	return &ghw.GraphicsCard{
		DeviceInfo: &pci.Device{
			Vendor:  &pcidb.Vendor{Name: vendor},
			Product: &pcidb.Product{Name: product},
		},
	}
}
