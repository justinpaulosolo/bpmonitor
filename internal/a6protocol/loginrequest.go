package a6protocol

type LoginRequest struct {
	RandomBytes  []byte
	UserSlot     int
	BatteryLevel int
}

const (
	randomBytesLength  = 6
	userSlotLength     = 1
	batteryLevelLength = 1
	loginRequestLength = headerLength + randomBytesLength + userSlotLength + batteryLevelLength
)

func ParseLoginRequest(data []byte) (LoginRequest, bool) {
	if len(data) < loginRequestLength {
		return LoginRequest{}, false
	}

	randomBytes := data[headerLength : headerLength+randomBytesLength]
	userSlot := int(data[headerLength+randomBytesLength])
	batteryLevel := int(data[headerLength+randomBytesLength+userSlotLength])

	return LoginRequest{
		RandomBytes:  randomBytes,
		UserSlot:     userSlot,
		BatteryLevel: batteryLevel,
	}, true
}
