package a6protocol

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestDecryptFrame_KnownVector(t *testing.T) {
	frame, _ := hex.DecodeString("101200072aa71876ba21fbf06e38abbdff7b71b1")
	key, _ := hex.DecodeString("5472616e7374656b4136b8b77d120e86")
	want, _ := hex.DecodeString("10120007000000000000023c0000000000000000")

	got, err := DecryptFrame(frame, key)
	if err != nil {
		t.Fatalf("DecryptFrame(...) failed: %v", err)
	}

	if !bytes.Equal(got, want) {
		t.Errorf("DecryptFrame(...) = %x, want %x", got, want)
	}
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	plainFrame, _ := hex.DecodeString("10120007000000000000023c0000000000000000")
	key, _ := hex.DecodeString("5472616e7374656b4136b8b77d120e86")

	encrypted, err := EncryptFrame(plainFrame, key)
	if err != nil {
		t.Fatalf("EncryptFrame(...) failed: %v", err)
	}
	decrypted, err := DecryptFrame(encrypted, key)
	if err != nil {
		t.Fatalf("DecryptFrame(...) failed: %v", err)
	}

	if !bytes.Equal(decrypted, plainFrame) {
		t.Errorf("Round-trip failed: got %x, want %x", decrypted, plainFrame)
	}
}

func TestEncryptFrame_TooShort(t *testing.T) {
	plainFrame := []byte{0x10, 0x12, 0x00} // Only 3 bytes, should be at least 4
	key, _ := hex.DecodeString("5472616e7374656b4136b8b77d120e86")

	_, err := EncryptFrame(plainFrame, key)
	if err == nil {
		t.Fatalf("EncryptFrame(...) did not return an error for too short input")
	}
}
