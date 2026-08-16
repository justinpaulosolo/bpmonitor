package a6protocol

type Measurement struct {
	Systolic      int
	Diastolic     int
	MeanPressure  int
	Status        int
	Pulse         int
	NeedsFollowUp bool
}

/*
┌───────┬───────────────────────────────────────────────┐
│ Bytes │                     Field                     │
├───────┼───────────────────────────────────────────────┤
│ 0-5   │ header + unused (ignore)                      │
├───────┼───────────────────────────────────────────────┤
│ 6-9   │ flag, 4 bytes big-endian                      │
├───────┼───────────────────────────────────────────────┤
│ 10-11 │ systolic, 2 bytes big-endian                  │
├───────┼───────────────────────────────────────────────┤
│ 12-13 │ diastolic                                     │
├───────┼───────────────────────────────────────────────┤
│ 14-15 │ mean pressure                                 │
├───────┼───────────────────────────────────────────────┤
│ 16-17 │ status                                        │
├───────┼───────────────────────────────────────────────┤
│ 18-19 │ pulse — only meaningful if flag & 0x02 is set │
└───────┴───────────────────────────────────────────────┘
*/

func ParseMeasurement(data []byte) (Measurement, bool) {
	if len(data) < 20 {
		return Measurement{}, false
	}

	flag := int(data[6])<<24 | int(data[7])<<16 | int(data[8])<<8 | int(data[9])
	systolic := int(data[10])<<8 | int(data[11])
	diastolic := int(data[12])<<8 | int(data[13])
	meanPressure := int(data[14])<<8 | int(data[15])
	status := int(data[16])<<8 | int(data[17])
	pulse := int(data[18])<<8 | int(data[19])

	needsFollowUp := (flag & 0x0C) != 0

	if (flag & 0x02) == 0 {
		pulse = 0
	}

	return Measurement{
		Systolic:      systolic,
		Diastolic:     diastolic,
		MeanPressure:  meanPressure,
		Status:        status,
		Pulse:         pulse,
		NeedsFollowUp: needsFollowUp,
	}, true
}
