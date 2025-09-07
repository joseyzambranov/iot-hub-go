package entities

import (
	"testing"
	"time"
)

func TestNewDevice(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		deviceType string
	}{
		{
			name:       "create new device",
			id:         "device123",
			deviceType: "sensor",
		},
		{
			name:       "create device with empty type",
			id:         "device456",
			deviceType: "",
		},
		{
			name:       "create device with different type",
			id:         "camera001",
			deviceType: "camera",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now()
			device := NewDevice(tt.id, tt.deviceType)
			after := time.Now()

			// Check basic properties
			if device.ID != tt.id {
				t.Errorf("NewDevice().ID = %v, want %v", device.ID, tt.id)
			}
			if device.Type != tt.deviceType {
				t.Errorf("NewDevice().Type = %v, want %v", device.Type, tt.deviceType)
			}

			// Check time is set reasonably
			if device.LastSeen.Before(before) || device.LastSeen.After(after) {
				t.Errorf("NewDevice().LastSeen = %v, want between %v and %v", device.LastSeen, before, after)
			}

			// Check RateLimit is initialized
			if device.RateLimit == nil {
				t.Error("NewDevice().RateLimit should not be nil")
			} else {
				if device.RateLimit.Count != 0 {
					t.Errorf("NewDevice().RateLimit.Count = %v, want 0", device.RateLimit.Count)
				}
				if device.RateLimit.Blocked != false {
					t.Errorf("NewDevice().RateLimit.Blocked = %v, want false", device.RateLimit.Blocked)
				}
				if device.RateLimit.LastReset.Before(before) || device.RateLimit.LastReset.After(after) {
					t.Errorf("NewDevice().RateLimit.LastReset = %v, want between %v and %v", device.RateLimit.LastReset, before, after)
				}
			}

			// Check Behavior is initialized
			if device.Behavior == nil {
				t.Error("NewDevice().Behavior should not be nil")
			} else {
				if device.Behavior.AccessAttempts == nil {
					t.Error("NewDevice().Behavior.AccessAttempts should not be nil")
				}
				if len(device.Behavior.AccessAttempts) != 0 {
					t.Errorf("NewDevice().Behavior.AccessAttempts length = %v, want 0", len(device.Behavior.AccessAttempts))
				}
			}

			// Check quarantine status
			if device.Quarantined != false {
				t.Errorf("NewDevice().Quarantined = %v, want false", device.Quarantined)
			}
		})
	}
}

func TestDevice_Struct(t *testing.T) {
	now := time.Now()
	
	device := &Device{
		ID:       "test123",
		Type:     "sensor",
		LastSeen: now,
		RateLimit: &DeviceRateLimit{
			Count:     5,
			LastReset: now,
			Blocked:   true,
		},
		Behavior: &DeviceBehavior{
			LastSeen:       now,
			MessageCount:   10,
			AvgTemperature: 25.5,
			AvgBattery:     80.0,
			AccessAttempts: []int{1, 2, 3},
			AnomalyCount:   2,
		},
		Quarantined:    true,
		QuarantineTime: now,
	}

	// Test that all fields are properly set
	if device.ID != "test123" {
		t.Errorf("Device.ID = %v, want test123", device.ID)
	}
	if device.Type != "sensor" {
		t.Errorf("Device.Type = %v, want sensor", device.Type)
	}
	if !device.LastSeen.Equal(now) {
		t.Errorf("Device.LastSeen = %v, want %v", device.LastSeen, now)
	}
	if device.RateLimit.Count != 5 {
		t.Errorf("Device.RateLimit.Count = %v, want 5", device.RateLimit.Count)
	}
	if !device.RateLimit.Blocked {
		t.Errorf("Device.RateLimit.Blocked = %v, want true", device.RateLimit.Blocked)
	}
	if device.Behavior.MessageCount != 10 {
		t.Errorf("Device.Behavior.MessageCount = %v, want 10", device.Behavior.MessageCount)
	}
	if device.Behavior.AvgTemperature != 25.5 {
		t.Errorf("Device.Behavior.AvgTemperature = %v, want 25.5", device.Behavior.AvgTemperature)
	}
	if len(device.Behavior.AccessAttempts) != 3 {
		t.Errorf("Device.Behavior.AccessAttempts length = %v, want 3", len(device.Behavior.AccessAttempts))
	}
	if !device.Quarantined {
		t.Errorf("Device.Quarantined = %v, want true", device.Quarantined)
	}
}

func TestDeviceRateLimit_Struct(t *testing.T) {
	now := time.Now()
	
	rateLimit := &DeviceRateLimit{
		Count:     10,
		LastReset: now,
		Blocked:   false,
	}

	if rateLimit.Count != 10 {
		t.Errorf("DeviceRateLimit.Count = %v, want 10", rateLimit.Count)
	}
	if !rateLimit.LastReset.Equal(now) {
		t.Errorf("DeviceRateLimit.LastReset = %v, want %v", rateLimit.LastReset, now)
	}
	if rateLimit.Blocked {
		t.Errorf("DeviceRateLimit.Blocked = %v, want false", rateLimit.Blocked)
	}
}

func TestDeviceBehavior_Struct(t *testing.T) {
	now := time.Now()
	attempts := []int{1, 2, 3, 4, 5}
	
	behavior := &DeviceBehavior{
		LastSeen:       now,
		MessageCount:   100,
		AvgTemperature: 22.5,
		AvgBattery:     75.0,
		AccessAttempts: attempts,
		AnomalyCount:   3,
	}

	if !behavior.LastSeen.Equal(now) {
		t.Errorf("DeviceBehavior.LastSeen = %v, want %v", behavior.LastSeen, now)
	}
	if behavior.MessageCount != 100 {
		t.Errorf("DeviceBehavior.MessageCount = %v, want 100", behavior.MessageCount)
	}
	if behavior.AvgTemperature != 22.5 {
		t.Errorf("DeviceBehavior.AvgTemperature = %v, want 22.5", behavior.AvgTemperature)
	}
	if behavior.AvgBattery != 75.0 {
		t.Errorf("DeviceBehavior.AvgBattery = %v, want 75.0", behavior.AvgBattery)
	}
	if len(behavior.AccessAttempts) != 5 {
		t.Errorf("DeviceBehavior.AccessAttempts length = %v, want 5", len(behavior.AccessAttempts))
	}
	if behavior.AnomalyCount != 3 {
		t.Errorf("DeviceBehavior.AnomalyCount = %v, want 3", behavior.AnomalyCount)
	}

	// Test slice equality
	for i, attempt := range behavior.AccessAttempts {
		if attempt != attempts[i] {
			t.Errorf("DeviceBehavior.AccessAttempts[%d] = %v, want %v", i, attempt, attempts[i])
		}
	}
}