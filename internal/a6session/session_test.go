package a6session

import (
	"bytes"
	"encoding/hex"
	"testing"
	"time"

	"github.com/justinpaulosolo/bpmonitor/internal/a6protocol"
)

type recordedWrite struct {
	characteristic string
	data           []byte
}

func TestHandleDataNotify_LoginRequest(t *testing.T) {
	key, _ := hex.DecodeString("5472616e7374656b4136b8b77d120e86")
	incomming, _ := hex.DecodeString("101200072aa71876ba21fbf06e38abbdff7b71b1")
	wantOutgoing, _ := hex.DecodeString("101200086f3721583f14da81efc85d12e4db164c")

	var calls []recordedWrite
	fakeWrite := func(characteristic string, data []byte) error {
		calls = append(calls, recordedWrite{characteristic, data})
		return nil
	}

	session := NewSession(fakeWrite, key, time.Now) // time doesn't matter for this path
	_, err := session.HandleDataNotify(incomming)
	if err != nil {
		t.Fatalf("HandleDataNotify( ... ) returned an error: %v", err)
	}

	if len(calls) != 2 {
		t.Fatalf("Write called %d times, want 2", len(calls))
	}

	wantAck := recordedWrite{CharAck, []byte{0x00, 0x01, 0x01}}
	if calls[0].characteristic != wantAck.characteristic || !bytes.Equal(calls[0].data, wantAck.data) {
		t.Errorf("first write = %+v, want %+v", calls[0], wantAck)
	}

	if calls[1].characteristic != CharCommand {
		t.Errorf("second write characteristic = %q, want %q", calls[1].characteristic, CharCommand)
	}
	if !bytes.Equal(calls[1].data, wantOutgoing) {
		t.Errorf("second write data = %x, want %x", calls[1].data, wantOutgoing)
	}

	if session.step != 1 {
		t.Errorf("session.step = %d, want 1", session.step)
	}
	if session.userSlot != 2 {
		t.Errorf("session.userSlot = %d, want 2", session.userSlot)
	}
}

func TestHandleAckNotify_StepTwo_SendsSyncRequest(t *testing.T) {
	key, _ := hex.DecodeString("5472616e7374656b4136b8b77d120e86")
	incoming, _ := hex.DecodeString("000101")
	wantOutgoing, _ := hex.DecodeString("1012490135d853fa2536cd32400a7f171d8ae484")

	var calls []recordedWrite
	fakeWrite := func(characteristic string, data []byte) error {
		calls = append(calls, recordedWrite{characteristic, data})
		return nil
	}

	session := NewSession(fakeWrite, key, time.Now) // time doesn't matter for this path
	session.step = 2
	session.userSlot = 2

	_, err := session.HandleAckNotify(incoming)
	if err != nil {
		t.Fatalf("HandleAckNotify(...) returned error: %v", err)
	}

	if len(calls) != 2 {
		t.Fatalf("write called %d times, want 2", len(calls))
	}

	wantAck := recordedWrite{CharAck, []byte{0x00, 0x01, 0x01}}
	if calls[0].characteristic != wantAck.characteristic || !bytes.Equal(calls[0].data, wantAck.data) {
		t.Errorf("first write = %+v, want %+v", calls[0], wantAck)
	}

	if calls[1].characteristic != CharCommand {
		t.Errorf("second write characteristic = %q, want %q", calls[1].characteristic, CharCommand)
	}
	if !bytes.Equal(calls[1].data, wantOutgoing) {
		t.Errorf("second write data = %x, want %x", calls[1].data, wantOutgoing)
	}

	if session.step != 3 {
		t.Errorf("session.step = %d, want 3", session.step)
	}
}

func TestHandleAckNotify_StepOne_SendsSetTime(t *testing.T) {
	key, _ := hex.DecodeString("5472616e7374656b4136b8b77d120e86")
	incoming, _ := hex.DecodeString("000101")

	// A fixed, known instant with an explicit non-UTC offset -- using a
	// real local timezone here would make the test's expected value depend
	// on whatever machine runs it. FixedZone keeps it fully deterministic
	// AND actually exercises the offset-addition logic (if that math were
	// missing, this test would catch it -- a UTC-only fixture wouldn't).
	fixedTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.FixedZone("TEST", 3600))
	fakeNow := func() time.Time { return fixedTime }

	var calls []recordedWrite
	fakeWrite := func(characteristic string, data []byte) error {
		calls = append(calls, recordedWrite{characteristic, data})
		return nil
	}

	session := NewSession(fakeWrite, key, fakeNow)
	session.step = 1

	_, err := session.HandleAckNotify(incoming)
	if err != nil {
		t.Fatalf("HandleAckNotify(...) returned error: %v", err)
	}

	if len(calls) != 2 {
		t.Fatalf("write called %d times, want 2", len(calls))
	}

	wantAck := recordedWrite{CharAck, []byte{0x00, 0x01, 0x01}}
	if calls[0].characteristic != wantAck.characteristic || !bytes.Equal(calls[0].data, wantAck.data) {
		t.Errorf("first write = %+v, want %+v", calls[0], wantAck)
	}
	if calls[1].characteristic != CharCommand {
		t.Errorf("second write characteristic = %q, want %q", calls[1].characteristic, CharCommand)
	}

	// Decrypt the sent command back to inspect it, rather than hardcoding
	// an expected ciphertext -- the exact bytes depend on the timestamp,
	// which we control via fakeNow, but let's verify by decrypting, not by
	// guessing what AES would output.
	decrypted, err := a6protocol.DecryptFrame(calls[1].data, key)
	if err != nil {
		t.Fatalf("failed to decrypt sent command: %v", err)
	}

	cmd := uint16(decrypted[2])<<8 | uint16(decrypted[3])
	if cmd != 0x1102 {
		t.Errorf("cmd = %#04x, want 0x1102", cmd)
	}

	payload := decrypted[4:9] // 0x01 + 4-byte timestamp = 5 bytes
	if payload[0] != 0x01 {
		t.Errorf("payload[0] = %#02x, want 0x01", payload[0])
	}

	gotTimestamp := uint32(payload[1])<<24 | uint32(payload[2])<<16 | uint32(payload[3])<<8 | uint32(payload[4])

	_, offsetSeconds := fixedTime.Zone()
	wantTimestamp := uint32(fixedTime.Unix() + int64(offsetSeconds) - 1262304000)

	if gotTimestamp != wantTimestamp {
		t.Errorf("timestamp = %d, want %d", gotTimestamp, wantTimestamp)
	}

	if session.step != 2 {
		t.Errorf("session.step = %d, want 2", session.step)
	}
}

func TestHandleDataNotify_Measurement_NeedsFollowUp(t *testing.T) {
	key, _ := hex.DecodeString("5472616e7374656b4136b8b77d120e86")
	incoming, _ := hex.DecodeString("20124902c4f5aecf77bdc3d9679ed9caf2e77560")

	var calls []recordedWrite
	fakeWrite := func(characteristic string, data []byte) error {
		calls = append(calls, recordedWrite{characteristic, data})
		return nil
	}

	session := NewSession(fakeWrite, key, time.Now)

	reading, err := session.HandleDataNotify(incoming)
	if err != nil {
		t.Fatalf("HandleDataNotify(...) returned error: %v", err)
	}
	if reading != nil {
		t.Fatalf("reading = %+v, want nil (this measurement needs a follow-up indicate frame)", reading)
	}

	if len(calls) != 1 {
		t.Fatalf("write called %d times, want 1 (just the ack)", len(calls))
	}
	wantAck := recordedWrite{CharAck, []byte{0x00, 0x01, 0x01}}
	if calls[0].characteristic != wantAck.characteristic || !bytes.Equal(calls[0].data, wantAck.data) {
		t.Errorf("write = %+v, want %+v", calls[0], wantAck)
	}

	if session.partial == nil {
		t.Fatalf("session.partial = nil, want populated")
	}
	if session.partial.Systolic != 128 {
		t.Errorf("partial.Systolic = %d, want 128", session.partial.Systolic)
	}
	if session.partial.Diastolic != 85 {
		t.Errorf("partial.Diastolic = %d, want 85", session.partial.Diastolic)
	}
	if session.partial.Pulse != 97 {
		t.Errorf("partial.Pulse = %d, want 97", session.partial.Pulse)
	}
}

func TestHandleIndicate_MergesWithPartial(t *testing.T) {
	key, _ := hex.DecodeString("5472616e7374656b4136b8b77d120e86")
	incoming, _ := hex.DecodeString("2112021f0befea2c4c55a33fb2afd65ddf1dfc8e")

	var calls []recordedWrite
	fakeWrite := func(characteristic string, data []byte) error {
		calls = append(calls, recordedWrite{characteristic, data})
		return nil
	}

	session := NewSession(fakeWrite, key, time.Now)
	session.partial = &a6protocol.Measurement{
		Systolic:     128,
		Diastolic:    85,
		MeanPressure: 106,
		Status:       0,
		Pulse:        97,
	}

	reading, err := session.HandleIndicate(incoming)
	if err != nil {
		t.Fatalf("HandleIndicate(...) returned error: %v", err)
	}
	if reading == nil {
		t.Fatalf("reading = nil, want a completed Reading")
	}

	if reading.Systolic != 128 || reading.Diastolic != 85 || reading.MeanPressure != 106 || reading.Pulse != 97 {
		t.Errorf("reading fields = %+v, want the stored partial's values", reading)
	}
	if reading.UserSlot == nil || *reading.UserSlot != 2 {
		t.Errorf("UserSlot = %v, want 2", reading.UserSlot)
	}
	if reading.DeviceTime == nil || *reading.DeviceTime != 1786837980 {
		t.Errorf("DeviceTime = %v, want 1786837980", reading.DeviceTime)
	}

	if len(calls) != 1 {
		t.Fatalf("write called %d times, want 1 (just the ack)", len(calls))
	}

	if session.partial != nil {
		t.Errorf("session.partial = %+v, want nil after emitting the reading", session.partial)
	}
}
