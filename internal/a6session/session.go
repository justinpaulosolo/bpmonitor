package a6session

import (
	"time"

	"github.com/justinpaulosolo/bpmonitor/internal/a6protocol"
)

const a6EpochOffset = 1262304000

type WriteFunc func(characteristic string, data []byte) error

/*
NOTES:
1. Device sends login request (cmd 0x0007) -> HandleDataNotify acks it, replies with
cmd 0x0008 (echoing the randome bytes), set step = 1
2. Device acks -> HandleAckNotify at step 1 sends set-time (cmd 0x1102), set step = 2
3. Device acks again -> HandleAckNotify at step 2 requests sync (cmd 0x4901), set step = 3
4. Device sends measurement (cmd 0x4902) -> stored in s.partial. If it has a follow-up flag,
it waits for the indicate frame; otherwise it returns the reading immediately.
5. The indicate frame arrives -> HandleIndicate combines the partial measurement with the user slot +
timestamp and returns the final Reading.
*/

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

func (s *Session) HandleDataNotify(data []byte) (*Reading, error) {
	decrypted, err := a6protocol.DecryptFrame(data, s.key)
	if err != nil {
		return nil, err
	}
	// cmd bytes 2-3 of decrypted data indicate the command type

	if len(decrypted) < 4 {
		return nil, nil // ignore invalid data
	}

	cmd := decrypted[2:4]

	switch {
	case cmd[0] == 0x00 && cmd[1] == 0x07:
		// Login request
		loginReq, ok := a6protocol.ParseLoginRequest(decrypted)
		if !ok {
			return nil, nil // ignore invalid login request
		}

		s.userSlot = loginReq.UserSlot

		err := s.write(CharAck, []byte{0x00, 0x01, 0x01}) // send ack
		if err != nil {
			return nil, err
		}

		responsePayload := []byte{0x01}
		responsePayload = append(responsePayload, loginReq.RandomBytes...)
		responsePayload = append(responsePayload, 0x00, 0x02)

		plain := a6protocol.Frame(0x0008, responsePayload)

		encrypted, err := a6protocol.EncryptFrame(plain, s.key)
		if err != nil {
			return nil, err
		}

		err = s.write(CharCommand, encrypted)
		if err != nil {
			return nil, err
		}

		s.step = 1
	case cmd[0] == 0x49 && cmd[1] == 0x02:
		measurement, ok := a6protocol.ParseMeasurement(decrypted)
		if !ok {
			return nil, nil
		}

		err := s.write(CharAck, []byte{0x00, 0x01, 0x01})
		if err != nil {
			return nil, err
		}
		s.partial = &measurement

		if measurement.NeedsFollowUp {
			return nil, nil
		}
		return &Reading{
			Systolic:     measurement.Systolic,
			Diastolic:    measurement.Diastolic,
			MeanPressure: measurement.MeanPressure,
			Pulse:        measurement.Pulse,
			Status:       measurement.Status,
			UserSlot:     nil,
			DeviceTime:   nil,
		}, nil
	}

	return nil, nil
}

func (s *Session) HandleAckNotify(data []byte) (*Reading, error) {
	err := s.write(CharAck, []byte{0x00, 0x01, 0x01}) // send ack
	if err != nil {
		return nil, err
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
			return nil, err
		}

		err = s.write(CharCommand, encrypted)
		if err != nil {
			return nil, err
		}
		s.step = 2

	case 2:
		payload := []byte{byte(s.userSlot), 0x01}

		frame := a6protocol.Frame(0x4901, payload)
		encrypted, err := a6protocol.EncryptFrame(frame, s.key)
		if err != nil {
			return nil, err
		}

		err = s.write(CharCommand, encrypted)
		if err != nil {
			return nil, err
		}
		s.step = 3

	default:
		// Handle other steps or ignore
	}
	return nil, nil
}

func (s *Session) HandleIndicate(data []byte) (*Reading, error) {
	if s.partial == nil {
		return nil, nil
	}

	decrypted, err := a6protocol.DecryptFrame(data, s.key)
	if err != nil {
		return nil, err
	}

	indicate, ok := a6protocol.ParseIndicate(decrypted)
	if !ok {
		return nil, nil
	}

	err = s.write(CharAck, []byte{0x00, 0x01, 0x01})
	if err != nil {
		return nil, err
	}
	reading := &Reading{
		Systolic:     s.partial.Systolic,
		Diastolic:    s.partial.Diastolic,
		MeanPressure: s.partial.MeanPressure,
		Pulse:        s.partial.Pulse,
		Status:       s.partial.Status,
		UserSlot:     indicate.UserSlot,
		DeviceTime:   indicate.DeviceTime,
	}
	s.partial = nil
	return reading, nil
}
