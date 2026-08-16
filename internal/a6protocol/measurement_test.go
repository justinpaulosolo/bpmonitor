package a6protocol

import (
	"encoding/hex"
	"testing"
)

func TestParseMeasurement_Valid(t *testing.T) {
	data, _ := hex.DecodeString("2012490200000000005e00800055006a00000061")
	got, ok := ParseMeasurement(data)

	if !ok {
		t.Fatalf("ParseMeasurement(...) returned ok=false, want true")
	}
	if got.Systolic != 128 {
		t.Errorf("Systolic = %d, want 128", got.Systolic)
	}
	if got.Diastolic != 85 {
		t.Errorf("Diastolic = %d, want 85", got.Diastolic)
	}
	if got.MeanPressure != 106 {
		t.Errorf("MeanPressure = %d, want 106", got.MeanPressure)
	}
	if got.Status != 0 {
		t.Errorf("Status = %d, want 0", got.Status)
	}
	if got.Pulse != 97 {
		t.Errorf("Pulse = %d, want 97", got.Pulse)
	}
	if got.NeedsFollowUp != true {
		t.Errorf("NeedsFollowUp = %v, want true", got.NeedsFollowUp)
	}
}

func TestParseMeasurement_TooShort(t *testing.T) {
	data, _ := hex.DecodeString("201249020000000000") // 9 bytes, nowhere near enough

	_, ok := ParseMeasurement(data)

	if ok {
		t.Fatalf("ParseMeasurement(...) returned ok=true for too-short input, want false")
	}
}

func TestParseMeasurement_PulseFlagOff(t *testing.T) {
	const (
		header    = "20124902"
		unused    = "0000"
		flagBits  = "00000000" // flag has NO bits set -- 0x02 is off
		systolic  = "0000"
		diastolic = "0000"
		mean      = "0000"
		status    = "0000"
		pulseJunk = "ffff" // present in the bytes, should still be ignored
	)
	data, _ := hex.DecodeString(header + unused + flagBits + systolic + diastolic + mean + status + pulseJunk)

	got, ok := ParseMeasurement(data)

	if !ok {
		t.Fatalf("ParseMeasurement(...) returned ok=false, want true")
	}
	if got.Pulse != 0 {
		t.Errorf("Pulse = %d, want 0 when flag&0x02 is unset (even though raw bytes were 0xffff)", got.Pulse)
	}
}
