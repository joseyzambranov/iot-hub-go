package entities

import "time"

type DeviceRateLimit struct {
	Count     int
	LastReset time.Time
	Blocked   bool
}

type DeviceBehavior struct {
	LastSeen       time.Time
	MessageCount   int
	AvgTemperature float64
	AvgBattery     float64
	AccessAttempts []int
	AnomalyCount   int
}

type Device struct {
	ID         string
	Type       string
	LastSeen   time.Time
	RateLimit  *DeviceRateLimit
	Behavior   *DeviceBehavior
	Quarantined bool
	QuarantineTime time.Time
}

func NewDevice(id, deviceType string) *Device {
	now := time.Now()
	return &Device{
		ID:       id,
		Type:     deviceType,
		LastSeen: now,
		RateLimit: &DeviceRateLimit{
			Count:     0,
			LastReset: now,
			Blocked:   false,
		},
		Behavior: &DeviceBehavior{
			AccessAttempts: make([]int, 0),
		},
	}
}