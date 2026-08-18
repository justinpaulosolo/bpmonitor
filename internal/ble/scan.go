package ble

import "tinygo.org/x/bluetooth"

func DeriveKeyFromManufacturerData(mfgData []bluetooth.ManufacturerDataElement) ([]byte, bool) {
	for _, m := range mfgData {
		if m.CompanyID == 0x12AE {
			identifier := m.Data[3:9]
			return append([]byte("TranstekA6"), identifier...), true
		}
	}
	return nil, false
}
