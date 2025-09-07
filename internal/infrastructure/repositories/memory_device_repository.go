package repositories

import (
	"context"
	"fmt"
	"sync"
	"time"
	
	"iot-hub-go/internal/domain/entities"
	"iot-hub-go/internal/domain/repositories"
)

type MemoryDeviceRepository struct {
	devices            map[string]*entities.Device
	quarantinedDevices map[string]time.Time
	mutex              sync.RWMutex
}

func NewMemoryDeviceRepository() repositories.DeviceRepository {
	return &MemoryDeviceRepository{
		devices:            make(map[string]*entities.Device),
		quarantinedDevices: make(map[string]time.Time),
	}
}

func (r *MemoryDeviceRepository) GetDevice(ctx context.Context, deviceID string) (*entities.Device, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	device, exists := r.devices[deviceID]
	if !exists {
		return nil, fmt.Errorf("device not found")
	}
	
	return device, nil
}

func (r *MemoryDeviceRepository) SaveDevice(ctx context.Context, device *entities.Device) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	r.devices[device.ID] = device
	return nil
}

func (r *MemoryDeviceRepository) UpdateDevice(ctx context.Context, device *entities.Device) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	r.devices[device.ID] = device
	return nil
}

func (r *MemoryDeviceRepository) GetQuarantinedDevices(ctx context.Context) ([]*entities.Device, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	var quarantined []*entities.Device
	for deviceID := range r.quarantinedDevices {
		if device, exists := r.devices[deviceID]; exists {
			quarantined = append(quarantined, device)
		}
	}
	
	return quarantined, nil
}

func (r *MemoryDeviceRepository) IsDeviceQuarantined(ctx context.Context, deviceID string) (bool, error) {
	r.mutex.RLock()
	quarantineTime, exists := r.quarantinedDevices[deviceID]
	r.mutex.RUnlock()
	
	if !exists {
		return false, nil
	}
	
	const QUARANTINE_DURATION = 5 * time.Minute
	if time.Since(quarantineTime) > QUARANTINE_DURATION {
		r.mutex.Lock()
		if quarantineTime, exists := r.quarantinedDevices[deviceID]; exists {
			if time.Since(quarantineTime) > QUARANTINE_DURATION {
				delete(r.quarantinedDevices, deviceID)
			}
		}
		r.mutex.Unlock()
		return false, nil
	}
	
	return true, nil
}

func (r *MemoryDeviceRepository) QuarantineDevice(ctx context.Context, deviceID string, reason string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	r.quarantinedDevices[deviceID] = time.Now()
	return nil
}

func (r *MemoryDeviceRepository) ReleaseFromQuarantine(ctx context.Context, deviceID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	delete(r.quarantinedDevices, deviceID)
	return nil
}

func (r *MemoryDeviceRepository) CleanExpiredQuarantines(ctx context.Context, duration time.Duration) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	now := time.Now()
	for deviceID, quarantineTime := range r.quarantinedDevices {
		if now.Sub(quarantineTime) > duration {
			delete(r.quarantinedDevices, deviceID)
		}
	}
	
	return nil
}