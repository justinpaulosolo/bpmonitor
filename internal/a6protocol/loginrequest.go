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

	maxUserSlot     = 8
	maxBatteryLevel = 100
)

func ParseLoginRequest(data []byte) (LoginRequest, bool) {
	if len(data) < loginRequestLength {
		return LoginRequest{}, false
	}

	randomBytes := data[headerLength : headerLength+randomBytesLength]
	userSlot := int(data[headerLength+randomBytesLength])
	batteryLevel := int(data[headerLength+randomBytesLength+userSlotLength])

	if userSlot < 0 || userSlot > 8 {
		userSlot = 0 // Clamp to 0 if out of range
	}

	return LoginRequest{
		RandomBytes:  randomBytes,
		UserSlot:     userSlot,
		BatteryLevel: batteryLevel,
	}, true
}
