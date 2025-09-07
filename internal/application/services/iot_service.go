package services

import (
	"context"
	
	"iot-hub-go/internal/application/dto"
	"iot-hub-go/internal/domain/usecases"
)

type IoTService struct {
	sensorProcessor *usecases.SensorDataProcessor
	rateLimiter     *usecases.RateLimiter
}

func NewIoTService(
	sensorProcessor *usecases.SensorDataProcessor,
	rateLimiter *usecases.RateLimiter,
) *IoTService {
	return &IoTService{
		sensorProcessor: sensorProcessor,
		rateLimiter:     rateLimiter,
	}
}

func (s *IoTService) ProcessSensorData(ctx context.Context, data *dto.SensorDataDTO) error {
	allowed, err := s.rateLimiter.CheckRateLimit(ctx, data.DeviceID)
	if err != nil {
		return err
	}
	
	if !allowed {
		return nil
	}
	
	sensorData := data.ToEntity()
	return s.sensorProcessor.ProcessSensorData(ctx, sensorData)
}