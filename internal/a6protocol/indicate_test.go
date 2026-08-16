package a6protocol

import (
	"encoding/hex"
	"testing"
)

func TestParseIndicate_Valid(t *testing.T) {
	data, _ := hex.DecodeString("2112021f43c0dc3eb4a3b4000000000000000000")

	got, ok := ParseIndicate(data)

	if !ok {
		t.Fatalf("ParseIndicate(...) returned ok=false, want true")
	}
	if got.UserSlot == nil {
		t.Fatalf("UserSlot = nil, want 2")
	}
	if *got.UserSlot != 2 {
		t.Errorf("UserSlot = %d, want 2", *got.UserSlot)
	}
	if got.DeviceTime == nil {
		t.Fatalf("DeviceTime = nil, want 1786837980")
	}
	if *got.DeviceTime != 1786837980 {
		t.Errorf("DeviceTime = %d, want 1786837980", *got.DeviceTime)
	}
}

func TestParseIndicate_TooShort(t *testing.T) {
	data, _ := hex.DecodeString("2112") // only 2 bytes, need at least 3

	_, ok := ParseIndicate(data)

	if ok {
		t.Fatalf("ParseIndicate(...) returned ok=true for too-short input, want false")
	}
}

func TestParseIndicate_UserSlotOnly_NoDeviceTime(t *testing.T) {
	data, _ := hex.DecodeString("211202") // exactly 3 bytes: enough for user slot, not device time

	got, ok := ParseIndicate(data)

	if !ok {
		t.Fatalf("ParseIndicate(...) returned ok=false, want true")
	}
	if got.UserSlot == nil {
		t.Fatalf("UserSlot = nil, want 2")
	}
	if *got.UserSlot != 2 {
		t.Errorf("UserSlot = %d, want 2", *got.UserSlot)
	}
	if got.DeviceTime != nil {
		t.Errorf("DeviceTime = %d, want nil (not enough bytes for it)", *got.DeviceTime)
	}
}
