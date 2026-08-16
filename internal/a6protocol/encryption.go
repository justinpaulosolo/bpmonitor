package a6protocol

import (
	"crypto/aes"
	"fmt"
)

const (
	headerLength = 4
	blockLength  = 16
	frameLength  = headerLength + blockLength
	frameType    = 0x12
)

func EncryptFrame(plainFrame []byte, key []byte) ([]byte, error) {
	if len(plainFrame) < headerLength {
		return nil, fmt.Errorf("a6protocol: plainFrame too short: got %d bytes, want %d", len(plainFrame), headerLength)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("a6protocol: %w", err)
	}

	header := make([]byte, headerLength)
	copy(header, plainFrame[:headerLength])
	header[1] = frameType

	payload := make([]byte, blockLength)
	copy(payload, plainFrame[headerLength:])

	encrypted := make([]byte, blockLength)
	block.Encrypt(encrypted, payload)

	return append(header, encrypted...), nil
}

func DecryptFrame(frame []byte, key []byte) ([]byte, error) {
	if len(frame) < frameLength {
		return frame, nil
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("a6protocol: %w", err)
	}
	header := make([]byte, headerLength)
	copy(header, frame[:headerLength])

	decrypted := make([]byte, blockLength)
	block.Decrypt(decrypted, frame[headerLength:frameLength])

	return append(header, decrypted...), nil
}
