package a6protocol

import (
	"encoding/hex"
	"testing"
)

func TestParseLoginRequest_Valid(t *testing.T) {
	data, _ := hex.DecodeString("10120007000000000000023c0000000000000000")

	got, ok := ParseLoginRequest(data)
	if !ok {
		t.Fatalf("ParseLoginRequest returned ok=false, want true")
	}
	if len(got.RandomBytes) != 6 {
		t.Errorf("RandomBytes length = %d, want 6", len(got.RandomBytes))
	}
	if got.UserSlot != 2 {
		t.Errorf("UserSlot = %d, want 2", got.UserSlot)
	}
	if got.BatteryLevel != 60 {
		t.Errorf("BatteryLevel = %d, want 60", got.BatteryLevel)
	}
}

func TestParseLoginRequest_TooShort(t *testing.T) {
	data, _ := hex.DecodeString("10120007000000000000")

	_, ok := ParseLoginRequest(data)
	if ok {
		t.Fatalf("ParseLoginRequest returned ok=true, want false for too short input")
	}
}

func TestParseLoginRequest_ClampsInvalidUserSlot(t *testing.T) {
	// header(4) + randomBytes(6, arbitrary) + rawSlot=0x0C(12, out of 0-8) + battery(0x50)
	data, _ := hex.DecodeString("10120007aaaaaaaaaaaa0c50")

	got, ok := ParseLoginRequest(data)

	if !ok {
		t.Fatalf("ParseLoginRequest(...) returned ok=false, want true")
	}
	if got.UserSlot != 0 {
		t.Errorf("UserSlot = %d, want 0 (clamped from out-of-range raw value)", got.UserSlot)
	}
}
