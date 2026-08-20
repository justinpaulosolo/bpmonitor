package ble

import (
	"fmt"
	"strings"
	"time"

	"github.com/justinpaulosolo/bpmonitor/internal/a6session"
	"tinygo.org/x/bluetooth"
)

type CaptureEvent struct {
	Status  string
	Reading *a6session.Reading
	Error   error
}

type found struct {
	address bluetooth.Address
	key     []byte
}

type notification struct {
	kind    string
	payload []byte
}

const deviceNameHint = "0661"

func Capture() (<-chan CaptureEvent, error) {
	adapter := bluetooth.DefaultAdapter
	if err := adapter.Enable(); err != nil {
		return nil, err
	}

	events := make(chan CaptureEvent, 10)

	go func() {
		defer close(events)

		events <- CaptureEvent{Status: "scanning"}

		result := make(chan found, 1)
		err := adapter.Scan(func(a *bluetooth.Adapter, r bluetooth.ScanResult) {
			name := r.LocalName()
			if !strings.Contains(name, deviceNameHint) {
				return
			}

			key, ok := DeriveKeyFromManufacturerData(r.ManufacturerData())
			if !ok {
				return
			}
			a.StopScan()
			result <- found{address: r.Address, key: key}
		})
		if err != nil {
			events <- CaptureEvent{Error: err}
			return
		}
		f := <-result

		events <- CaptureEvent{Status: "found device, connecting"}

		device, err := adapter.Connect(f.address, bluetooth.ConnectionParams{})
		if err != nil {
			events <- CaptureEvent{Error: err}
			return
		}

		services, err := device.DiscoverServices([]bluetooth.UUID{ServiceUUID})
		if err != nil {
			events <- CaptureEvent{Error: err}
			return
		}
		if len(services) == 0 {
			events <- CaptureEvent{Error: fmt.Errorf("no services discovered")}
			return
		}

		chars, err := services[0].DiscoverCharacteristics([]bluetooth.UUID{
			CharNotifyDataUUID,
			CharIndicateDataUUID,
			CharWriteAckUUID,
			CharNotifyAckUUID,
			CharWriteCommandUUID})
		if err != nil {
			events <- CaptureEvent{Error: err}
			return
		}

		var notifyChar, indicateChar, ackChar, notifyAckChar, commandChar bluetooth.DeviceCharacteristic

		for _, c := range chars {
			switch c.UUID() {
			case CharNotifyDataUUID:
				notifyChar = c
			case CharIndicateDataUUID:
				indicateChar = c
			case CharWriteAckUUID:
				ackChar = c
			case CharNotifyAckUUID:
				notifyAckChar = c
			case CharWriteCommandUUID:
				commandChar = c
			}
		}

		events <- CaptureEvent{Status: "connected, subscribing"}

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

		notifications := make(chan notification, 10)

		if err := notifyChar.EnableNotifications(func(data []byte) {
			notifications <- notification{kind: "data", payload: data}
		}); err != nil {
			events <- CaptureEvent{Error: err}
			return
		}

		if err := indicateChar.EnableNotifications(func(data []byte) {
			notifications <- notification{kind: "indicate", payload: data}
		}); err != nil {
			events <- CaptureEvent{Error: err}
			return
		}

		if err := notifyAckChar.EnableNotifications(func(data []byte) {
			notifications <- notification{kind: "ack", payload: data}
		}); err != nil {
			events <- CaptureEvent{Error: err}
			return
		}

		events <- CaptureEvent{Status: "waiting for reading"}

		for {
			select {
			case n := <-notifications:
				switch n.kind {
				case "data":
					reading, err := session.HandleDataNotify(n.payload)
					if err != nil {
						events <- CaptureEvent{Error: err}
						return
					}
					if reading != nil {
						events <- CaptureEvent{Reading: reading}
						return
					}
				case "indicate":
					reading, err := session.HandleIndicate(n.payload)
					if err != nil {
						events <- CaptureEvent{Error: err}
						return
					}
					if reading != nil {
						events <- CaptureEvent{Reading: reading}
						return
					}
				case "ack":
					_, err := session.HandleAckNotify(n.payload)
					if err != nil {
						events <- CaptureEvent{Error: err}
						return
					}
				}
			case <-time.After(60 * time.Second):
				events <- CaptureEvent{Error: fmt.Errorf("timeout")}
				return
			}
		}
	}()

	return events, nil
}
