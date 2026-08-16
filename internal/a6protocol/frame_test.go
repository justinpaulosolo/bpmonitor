package a6protocol

import (
	"bytes"
	"testing"
)

func TestFrame(t *testing.T) {
	tests := []struct {
		name    string
		cmd     uint16
		payload []byte
		want    []byte
	}{
		{
			name:    "empty payload",
			cmd:     0x0008,
			payload: []byte{},
			want:    []byte{0x10, 0x02, 0x00, 0x08},
		},
		{
			name:    "payload with data",
			cmd:     0x1102,
			payload: []byte{0x01, 0x02, 0x03},
			want:    []byte{0x10, 0x05, 0x11, 0x02, 0x01, 0x02, 0x03},
		},
		{
			name:    "nil payload",
			cmd:     0x4901,
			payload: nil,
			want:    []byte{0x10, 0x02, 0x49, 0x01},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Frame(tt.cmd, tt.payload)
			if !bytes.Equal(got, tt.want) {
				t.Errorf("Frame(%#04x, %v) = % x, want % x", tt.cmd, tt.payload, got, tt.want)
			}
		})
	}
}
