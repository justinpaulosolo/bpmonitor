package a6protocol

const a6EpochOfsset = 1262304000 // A6 epoch starts at 2010-01-01 00:00:00 UTC, which is 1262304000 seconds

type IndicateData struct {
	UserSlot   *int
	DeviceTime *uint32
}

func ParseIndicate(data []byte) (IndicateData, bool) {
	if len(data) < 3 {
		return IndicateData{}, false
	}

	var result IndicateData

	offset := 2 // Skip the first two bytes (header)

	if len(data) >= offset+1 {
		userSlot := int(data[offset])
		result.UserSlot = &userSlot
		offset += 1
	}

	if len(data) >= offset+4 {
		rawTimestamp := uint32(data[offset])<<24 | uint32(data[offset+1])<<16 | uint32(data[offset+2])<<8 | uint32(data[offset+3])
		deviceTime := rawTimestamp + a6EpochOfsset
		result.DeviceTime = &deviceTime
		offset += 4
	}

	return result, true
}
