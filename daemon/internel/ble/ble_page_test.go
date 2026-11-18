package ble

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/beclab/Olares/daemon/internel/wifi"
)

func makeAdvertisedAPs(n int) []advertisedAP {
	aps := make([]advertisedAP, n)
	for i := range aps {
		aps[i] = advertisedAP{
			SSID:          fmt.Sprintf("network-%03d", i),
			Strength:      byte(100 - i%100),
			SecurityTypes: []wifi.Security{},
		}
	}
	return aps
}

func decodePages(t *testing.T, pages [bleAPPageCount]string) []advertisedAP {
	t.Helper()
	var all []advertisedAP
	for i, page := range pages {
		if len(page) > bleAPPageSize {
			t.Fatalf("page %d is %d bytes, exceeds %d", i+1, len(page), bleAPPageSize)
		}
		var decoded []advertisedAP
		if err := json.Unmarshal([]byte(page), &decoded); err != nil {
			t.Fatalf("page %d is not valid JSON: %v", i+1, err)
		}
		all = append(all, decoded...)
	}
	return all
}

func TestPaginateAPsFewEntries(t *testing.T) {
	pages := paginateAPs(makeAdvertisedAPs(3))
	if got := decodePages(t, pages); len(got) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(got))
	}
	for i := 1; i < bleAPPageCount; i++ {
		if pages[i] != "[]" {
			t.Fatalf("page %d should be empty, got %s", i+1, pages[i])
		}
	}
}

func TestPaginateAPsEmpty(t *testing.T) {
	for i, page := range paginateAPs(nil) {
		if page != "[]" {
			t.Fatalf("page %d should be empty, got %s", i+1, page)
		}
	}
}

func TestPaginateAPsPreservesOrderAcrossPages(t *testing.T) {
	aps := makeAdvertisedAPs(40)
	pages := paginateAPs(aps)
	if pages[1] == "[]" {
		t.Fatal("expected entries to spill into the second page")
	}
	got := decodePages(t, pages)
	if len(got) != len(aps) {
		t.Fatalf("expected %d entries, got %d", len(aps), len(got))
	}
	for i := range got {
		if got[i].SSID != aps[i].SSID {
			t.Fatalf("entry %d: expected %s, got %s", i, aps[i].SSID, got[i].SSID)
		}
	}
}

func TestPaginateAPsDropsOverflow(t *testing.T) {
	aps := makeAdvertisedAPs(500)
	pages := paginateAPs(aps)
	got := decodePages(t, pages)
	if len(got) == 0 || len(got) >= len(aps) {
		t.Fatalf("expected overflow to be dropped, got %d of %d", len(got), len(aps))
	}
	for i := range got {
		if got[i].SSID != aps[i].SSID {
			t.Fatalf("entry %d: expected %s, got %s", i, aps[i].SSID, got[i].SSID)
		}
	}
	if pages[bleAPPageCount-1] == "[]" {
		t.Fatal("expected the last page to be used")
	}
}
