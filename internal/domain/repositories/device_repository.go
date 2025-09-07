package repositories

import (
	"context"
	"time"
	
	"iot-hub-go/internal/domain/entities"
)

type DeviceRepository interface {
	GetDevice(ctx context.Context, deviceID string) (*entities.Device, error)
	SaveDevice(ctx context.Context, device *entities.Device) error
	UpdateDevice(ctx context.Context, device *entities.Device) error
	GetQuarantinedDevices(ctx context.Context) ([]*entities.Device, error)
	IsDeviceQuarantined(ctx context.Context, deviceID string) (bool, error)
	QuarantineDevice(ctx context.Context, deviceID string, reason string) error
	ReleaseFromQuarantine(ctx context.Context, deviceID string) error
	CleanExpiredQuarantines(ctx context.Context, duration time.Duration) error
}