package a6protocol

import "encoding/binary"

const (
	adTypeManufacturerSpecificData = 0xFF
	companyIDANDMedical            = 0x12AE
	identifierOffset               = 5
	identifierLength               = 6
	minManufacturerDataLength      = identifierOffset + identifierLength
	keyPrefix                      = "TranstekA6"
)

func DeriveKey(scanRecord []byte) (key []byte, ok bool) {
	for i := 0; i < len(scanRecord); {
		length := int(scanRecord[i])
		if length == 0 {
			break
		}
		end := i + 1 + length
		if end > len(scanRecord) || i+2 > len(scanRecord) {
			break
		}

		adType := scanRecord[i+1]
		data := scanRecord[i+2 : end]

		if adType == adTypeManufacturerSpecificData && len(data) >= minManufacturerDataLength {
			companyID := binary.LittleEndian.Uint16(data[0:2])
			if companyID == companyIDANDMedical {
				identifier := data[identifierOffset : identifierOffset+identifierLength]
				key := append([]byte(keyPrefix), identifier...)
				return key, true
			}
		}
		i = end
	}

	return nil, false
}
