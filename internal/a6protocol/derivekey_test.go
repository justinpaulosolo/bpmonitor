package a6protocol

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestDeriveKey(t *testing.T) {
	// AD Length: everything after this is the advertisement data
	// AD Type: Manufacturer Specific Data
	// Company ID: 0x12AE (A&D)
	// Device Extra: 0x567801 (A&D)
	// Identifier: 0xB8B77D120E86 (A&D)

	const (
		adLength    = "0C" // 12 bytes of advertisement data
		adType      = "FF" // FF == Manufacturer Specific Data
		companyID   = "AE12"
		deviceExtra = "567801"
		identifier  = "B8B77D120E86"
		keyPrefix   = "5472616e7374656b4136"
	)

	input, err := hex.DecodeString(adLength + adType + companyID + deviceExtra + identifier)
	if err != nil {
		t.Fatalf("Failed to decode input hex string: %v", err)
	}
	want, err := hex.DecodeString(keyPrefix + identifier)
	if err != nil {
		t.Fatalf("Failed to decode want hex string: %v", err)
	}
	key, ok := DeriveKey(input)
	if !ok {
		t.Fatalf("DeriveKey returned ok=false, want true")
	}
	if !bytes.Equal(key, want) {
		t.Errorf("DeriveKey returned key=%x, want %x", key, want)
	}
}

func TestDeriveKey_TooShort(t *testing.T) {
	const (
		adLength  = "05" // 5 bytes of advertisement data
		adType    = "FF" // FF == Manufacturer Specific Data
		companyID = "AE12"
		shortData = "9999" // Only 2 bytes here
	)
	input, _ := hex.DecodeString(adLength + adType + companyID + shortData)

	key, ok := DeriveKey(input)
	if ok {
		t.Fatalf("DeriveKey returned ok=true, want false for too short input")
	}
	if key != nil {
		t.Errorf("DeriveKey returned key=%x, want nil for too short input", key)
	}
}
