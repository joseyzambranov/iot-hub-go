package usecases

import (
	"context"
	"log"
	"time"
	
	"iot-hub-go/internal/domain/entities"
	"iot-hub-go/internal/domain/repositories"
)

const MAX_MESSAGES_PER_MINUTE = 20

type RateLimiter struct {
	deviceRepo repositories.DeviceRepository
}

func NewRateLimiter(deviceRepo repositories.DeviceRepository) *RateLimiter {
	return &RateLimiter{
		deviceRepo: deviceRepo,
	}
}

func (r *RateLimiter) CheckRateLimit(ctx context.Context, deviceID string) (bool, error) {
	device, err := r.deviceRepo.GetDevice(ctx, deviceID)
	if err != nil {
		device = entities.NewDevice(deviceID, "")
	}
	
	now := time.Now()
	rateLimitInfo := device.RateLimit
	
	if now.Sub(rateLimitInfo.LastReset) >= time.Minute {
		rateLimitInfo.Count = 0
		rateLimitInfo.LastReset = now
		rateLimitInfo.Blocked = false
	}
	
	if rateLimitInfo.Count >= MAX_MESSAGES_PER_MINUTE {
		rateLimitInfo.Blocked = true
		log.Printf("🚫 RATE LIMIT: Dispositivo %s bloqueado por exceder %d mensajes/min", deviceID, MAX_MESSAGES_PER_MINUTE)
		r.deviceRepo.UpdateDevice(ctx, device)
		return false, nil
	}
	
	rateLimitInfo.Count++
	r.deviceRepo.UpdateDevice(ctx, device)
	return true, nil
}