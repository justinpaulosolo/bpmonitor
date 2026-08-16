package a6session

import (
	"time"

	"github.com/justinpaulosolo/bpmonitor/internal/a6protocol"
)

const a6EpochOffset = 1262304000

type WriteFunc func(characteristic string, data []byte) error

type Session struct {
	write    WriteFunc
	now      func() time.Time
	step     int
	key      []byte
	userSlot int
	partial  *a6protocol.Measurement
}

const (
	CharAck     = "ack"
	CharCommand = "command"
)

func NewSession(write WriteFunc, key []byte, now func() time.Time) *Session {
	return &Session{
		write: write,
		key:   key,
		now:   now,
	}
}

func (s *Session) HandleDataNotify(data []byte) error {
	decrypted, err := a6protocol.DecryptFrame(data, s.key)
	if err != nil {
		return err
	}
	// cmd bytes 2-3 of decrypted data indicate the command type

	if len(decrypted) < 4 {
		return nil // ignore invalid data
	}

	cmd := decrypted[2:4]

	switch {
	case cmd[0] == 0x00 && cmd[1] == 0x07:
		// Login request
		loginReq, ok := a6protocol.ParseLoginRequest(decrypted)
		if !ok {
			return nil // ignore invalid login request
		}

		s.userSlot = loginReq.UserSlot

		err := s.write(CharAck, []byte{0x00, 0x01, 0x01}) // send ack
		if err != nil {
			return err
		}

		responsePayload := []byte{0x01}
		responsePayload = append(responsePayload, loginReq.RandomBytes...)
		responsePayload = append(responsePayload, 0x00, 0x02)

		plain := a6protocol.Frame(0x0008, responsePayload)

		encrypted, err := a6protocol.EncryptFrame(plain, s.key)
		if err != nil {
			return err
		}

		err = s.write(CharCommand, encrypted)
		if err != nil {
			return err
		}

		s.step = 1
	}

	return nil
}

func (s *Session) HandleAckNotify(data []byte) error {
	err := s.write(CharAck, []byte{0x00, 0x01, 0x01}) // send ack
	if err != nil {
		return err
	}

	switch s.step {
	case 1:
		_, offsetSeconds := s.now().Zone() // get the current time offset in seconds
		deviceTime := uint32(s.now().Unix() + int64(offsetSeconds) - a6EpochOffset)

		payload := []byte{0x01}
		payload = append(payload, byte(deviceTime>>24), byte(deviceTime>>16), byte(deviceTime>>8), byte(deviceTime))

		frame := a6protocol.Frame(0x1102, payload)
		encrypted, err := a6protocol.EncryptFrame(frame, s.key)
		if err != nil {
			return err
		}

		err = s.write(CharCommand, encrypted)
		if err != nil {
			return err
		}
		s.step = 2

	case 2:
		payload := []byte{byte(s.userSlot), 0x01}

		frame := a6protocol.Frame(0x4901, payload)
		encrypted, err := a6protocol.EncryptFrame(frame, s.key)
		if err != nil {
			return err
		}

		err = s.write(CharCommand, encrypted)
		if err != nil {
			return err
		}
		s.step = 3

	default:
		// Handle other steps or ignore
	}
	return nil
}
