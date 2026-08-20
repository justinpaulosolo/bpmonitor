package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/justinpaulosolo/bpmonitor/internal/a6session"
	"github.com/justinpaulosolo/bpmonitor/internal/ble"
	"github.com/justinpaulosolo/bpmonitor/internal/storage"
	"tinygo.org/x/bluetooth"
)

const deviceName = "0661"

type found struct {
	address bluetooth.Address
	key     []byte
}

type notification struct {
	kind    string
	payload []byte
}

func main() {
	adapter := bluetooth.DefaultAdapter

	if err := adapter.Enable(); err != nil {
		fmt.Println("failed to enable adapter:", err)
		return
	}

	store, err := storage.Open("bpmonitor.db")
	if err != nil {
		fmt.Println("failed to open storage:", err)
		return
	}
	defer store.Close()

	fmt.Println("Scanning for a device with", deviceName, "in its name...")

	result := make(chan found, 1)

	err = adapter.Scan(func(a *bluetooth.Adapter, r bluetooth.ScanResult) {
		name := r.LocalName()
		if !strings.Contains(name, deviceName) {
			return
		}

		fmt.Printf("Found %q (%s):\n", name, r.Address.String())
		fmt.Println("--- Inspecting advertisement payload ---")

		mfgData := r.ManufacturerData()
		for _, m := range mfgData {
			fmt.Printf(" CompanyID = 0x%04X (%d) Data = % x\n", m.CompanyID, m.CompanyID, m.Data)
		}

		key, ok := ble.DeriveKeyFromManufacturerData(r.ManufacturerData())
		if !ok {
			fmt.Println("could not derive key from advertisement")
		}

		a.StopScan()
		result <- found{
			address: r.Address,
			key:     key,
		}
	})

	f := <-result

	fmt.Printf("Derived Key: %x\n", f.key)

	// Open Connection
	device, err := adapter.Connect(f.address, bluetooth.ConnectionParams{})
	if err != nil {
		fmt.Println("failed to connect to device:", err)
		return
	}

	services, err := device.DiscoverServices([]bluetooth.UUID{ble.ServiceUUID})
	if err != nil {
		fmt.Println("failed to discover services:", err)
		return
	}

	if len(services) == 0 {
		fmt.Println("no services discovered")
	}

	fmt.Println("--- Serivices ---")
	for _, service := range services {
		fmt.Println(service)
	}

	chars, err := services[0].DiscoverCharacteristics([]bluetooth.UUID{ble.CharNotifyDataUUID, ble.CharIndicateDataUUID, ble.CharWriteAckUUID, ble.CharNotifyAckUUID, ble.CharWriteCommandUUID})
	if err != nil {
		fmt.Println("failed to discover characteristics:", err)
		return
	}

	fmt.Println("--- Characteristics ---")
	for _, char := range chars {
		fmt.Println(char)
	}

	var notifyChar, indicateChar, ackChar, notifyAckChar, commandChar bluetooth.DeviceCharacteristic
	for _, c := range chars {
		switch c.UUID() {
		case ble.CharNotifyDataUUID:
			notifyChar = c
		case ble.CharIndicateDataUUID:
			indicateChar = c
		case ble.CharWriteAckUUID:
			ackChar = c
		case ble.CharNotifyAckUUID:
			notifyAckChar = c
		case ble.CharWriteCommandUUID:
			commandChar = c
		}
	}

	fmt.Println("All characteristics matched, subscribing...")

	writeFunc := func(characteristic string, data []byte) error {
		switch characteristic {
		case a6session.CharAck:
			_, err := ackChar.WriteWithoutResponse(data)
			return err
		case a6session.CharCommand:
			_, err := commandChar.WriteWithoutResponse(data)
			return err
		}
		return fmt.Errorf("unknown characteristic: %s", characteristic)
	}

	session := a6session.NewSession(writeFunc, f.key, time.Now)

	events := make(chan notification, 10)

	if err := notifyChar.EnableNotifications(func(data []byte) {
		events <- notification{kind: "data", payload: data}
	}); err != nil {
		fmt.Println("failed to subscribe to notify data: ", err)
		return
	}

	if err := indicateChar.EnableNotifications(func(data []byte) {
		events <- notification{kind: "indicate", payload: data}
	}); err != nil {
		fmt.Println("failed to subscribe to indicate data: ", err)
		return
	}

	if err := notifyAckChar.EnableNotifications(func(data []byte) {
		events <- notification{kind: "ack", payload: data}
	}); err != nil {
		fmt.Println("failed to subscribe to notify ack data: ", err)
		return
	}

	for {
		select {
		case ev := <-events:
			switch ev.kind {
			case "data":
				reading, err := session.HandleDataNotify(ev.payload)
				if err != nil {
					fmt.Println("HandleDataNotify error:", err)
					continue
				}
				if reading != nil {
					fmt.Println("Received reading:", *reading)
					return
				}
			case "indicate":
				reading, err := session.HandleIndicate(ev.payload)
				if err != nil {
					fmt.Println("HandleIndicate error:", err)
					continue
				}
				if reading != nil {
					fmt.Println("Received reading:", *reading)
					fmt.Printf("systolic=%d userSlot=%d deviceTime=%d\n", reading.Systolic, reading.UserSlot, reading.DeviceTime)
					return
				}
			case "ack":
				if _, err := session.HandleAckNotify(ev.payload); err != nil {
					fmt.Println("HandleAckNotify error:", err)
				}
			}
		case <-time.After(60 * time.Second):
			fmt.Println("timeout waiting for notifications")
			return
		}
	}
}
