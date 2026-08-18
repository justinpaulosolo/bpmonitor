package a6session

type Reading struct {
	Systolic     int
	Diastolic    int
	MeanPressure int
	Pulse        int
	Status       int
	UserSlot     *int
	DeviceTime   *uint32
}
