package ble

type Peripheral interface {
	WriteCharacteristic(uuid string, data []byte) error
	SubscribeCharacteristic(uuid string, onNotify func(data []byte)) error
}
