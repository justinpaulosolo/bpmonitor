package ble

import "tinygo.org/x/bluetooth"

var (
	ServiceUUID          = mustParseUUID("0000a610-0000-1000-8000-00805f9b34fb")
	CharNotifyDataUUID   = mustParseUUID("0000a621-0000-1000-8000-00805f9b34fb")
	CharIndicateDataUUID = mustParseUUID("0000a620-0000-1000-8000-00805f9b34fb")
	CharWriteAckUUID     = mustParseUUID("0000a622-0000-1000-8000-00805f9b34fb")
	CharNotifyAckUUID    = mustParseUUID("0000a625-0000-1000-8000-00805f9b34fb")
	CharWriteCommandUUID = mustParseUUID("0000a624-0000-1000-8000-00805f9b34fb")
)

func mustParseUUID(s string) bluetooth.UUID {
	uuid, err := bluetooth.ParseUUID(s)
	if err != nil {
		panic(err)
	}
	return uuid
}
