package ble

import (
	"fmt"
	"testing"
	"time"

	"k8s.io/klog/v2"
	"tinygo.org/x/bluetooth"
)

func TestUUID(t *testing.T) {
	fmt.Printf("Service UUID: %s\n", OlaersServiceUUID.String())
	fmt.Printf("Connect Characteristic UUID: %s\n", ConnectWifiUUID.String())
	fmt.Printf("ListAP Characteristic UUID: %s\n", ListAPUUID.String())
	fmt.Printf("TerminusInfo Characteristic UUID: %s\n", TerminusInfoUUID.String())
}

func TestScan(t *testing.T) {
	var adapter = bluetooth.DefaultAdapter

	// Enable BLE interface.
	must("enable BLE stack", adapter.Enable())

	// Start scanning.
	println("scanning....")
	var deviceUUID bluetooth.Address
	err := adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
		if device.HasServiceUUID(OlaersServiceUUID) {
			// if device.LocalName() == "olares2" {
			//if device.Address.String() == "46f0ed5c-8fd5-408d-b7c5-d4c2b43fec43" {
			println("found device:", device.Address.String(), device.RSSI, device.LocalName())
			deviceUUID = device.Address
			// adapter.StopScan()
		}
	})
	must("start scan", err)

	device, err := adapter.Connect(deviceUUID, bluetooth.ConnectionParams{ConnectionTimeout: bluetooth.NewDuration(time.Minute)})
	if err != nil {
		panic(err)
	}

	println("connected to ", deviceUUID.String())

	println("discovering services/characteristics")
	//srvcs, err := device.DiscoverServices([]bluetooth.UUID{OlaersServiceUUID})
	srvcs, err := device.DiscoverServices(nil)
	must("discover services", err)

	if len(srvcs) == 0 {
		panic("could not find terminus info service")
	}

	srvc := srvcs[0]

	for _, srvc := range srvcs {
		println("found service", srvc.UUID().String())
	}

	//	chars, err := srvc.DiscoverCharacteristics([]bluetooth.UUID{ListAPUUID})
	chars, err := srvc.DiscoverCharacteristics(nil)
	if err != nil {
		println(err)
	}

	if len(chars) == 0 {
		panic("could not find terminus info characteristic")
	}

	// char := chars[0]

	for _, char := range chars {
		println("found characteristic", char.UUID().String())

		// char.EnableNotifications(func(buf []byte) {
		// 	println("data:", uint8(buf[1]))
		// })

		data := make([]byte, 4096)
		n, err := char.Read(data)
		if err != nil {
			klog.Error(err)
		}
		fmt.Printf("read %d bytes, %s", n, string(data))
	}
}

func must(action string, err error) {
	if err != nil {
		panic("failed to " + action + ": " + err.Error())
	}
}
