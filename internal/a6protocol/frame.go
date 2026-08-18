package a6protocol

func Frame(cmd uint16, payload []byte) []byte {
	// [0x10] [length] [cmd hi] [cmd lo] [payload as N bytes]
	length := byte(2 + len(payload))

	frame := make([]byte, 0, 4+len(payload))
	frame = append(frame, 0x10)                    // header
	frame = append(frame, length)                  // length
	frame = append(frame, byte(cmd>>8), byte(cmd)) // cmd
	frame = append(frame, payload...)              // payload

	return frame
}
