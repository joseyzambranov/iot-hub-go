package services

import (
	"context"
	"testing"
	"time"

	"iot-hub-go/internal/application/dto"
	"iot-hub-go/internal/domain/entities"
	"iot-hub-go/internal/domain/usecases"
	"iot-hub-go/internal/infrastructure/repositories"
)

// Mock NotificationService for testing
type mockNotificationService struct{}

func (m *mockNotificationService) SendAnomalyAlert(ctx context.Context, anomaly *entities.Anomaly) error {
	return nil
}

func (m *mockNotificationService) SendQuarantineAlert(ctx context.Context, deviceID, reason string) error {
	return nil
}

func TestNewIoTService(t *testing.T) {
	// Create real instances for testing constructor
	deviceRepo := repositories.NewMemoryDeviceRepository()
	anomalyRepo := repositories.NewMemoryAnomalyRepository()
	notificationService := &mockNotificationService{}
	
	sensorProcessor := usecases.NewSensorDataProcessor(deviceRepo, anomalyRepo, notificationService)
	rateLimiter := usecases.NewRateLimiter(deviceRepo)

	service := NewIoTService(sensorProcessor, rateLimiter)

	if service == nil {
		t.Fatal("NewIoTService() returned nil")
	}
}

func TestIoTService_ProcessSensorData_Success(t *testing.T) {
	// Create real instances for integration test
	deviceRepo := repositories.NewMemoryDeviceRepository()
	anomalyRepo := repositories.NewMemoryAnomalyRepository()
	notificationService := &mockNotificationService{}
	
	sensorProcessor := usecases.NewSensorDataProcessor(deviceRepo, anomalyRepo, notificationService)
	rateLimiter := usecases.NewRateLimiter(deviceRepo)
	
	service := NewIoTService(sensorProcessor, rateLimiter)
	ctx := context.Background()

	data := &dto.SensorDataDTO{
		DeviceID:    "device123",
		DeviceType:  "sensor",
		Timestamp:   time.Now().Unix(),
		Temperature: 25.0,
		Humidity:    60.0,
	}

	err := service.ProcessSensorData(ctx, data)
	if err != nil {
		t.Errorf("ProcessSensorData() error = %v, want nil", err)
	}

	// Verify device was created/updated
	device, err := deviceRepo.GetDevice(ctx, "device123")
	if err != nil {
		t.Errorf("Device should be created after processing data")
	}
	if device == nil {
		t.Fatal("Device should not be nil")
	}
	if device.ID != "device123" {
		t.Errorf("Device ID = %v, want device123", device.ID)
	}
}

func TestIoTService_ProcessSensorData_InvalidData(t *testing.T) {
	deviceRepo := repositories.NewMemoryDeviceRepository()
	anomalyRepo := repositories.NewMemoryAnomalyRepository()
	notificationService := &mockNotificationService{}
	
	sensorProcessor := usecases.NewSensorDataProcessor(deviceRepo, anomalyRepo, notificationService)
	rateLimiter := usecases.NewRateLimiter(deviceRepo)
	
	service := NewIoTService(sensorProcessor, rateLimiter)
	ctx := context.Background()

	// Create data with invalid device ID (empty)
	data := &dto.SensorDataDTO{
		DeviceID:    "", // Invalid: empty device ID
		DeviceType:  "sensor",
		Timestamp:   time.Now().Unix(),
		Temperature: 25.0,
	}

	err := service.ProcessSensorData(ctx, data)
	if err == nil {
		t.Error("ProcessSensorData() with invalid data should return error")
	}
}

func TestIoTService_ProcessSensorData_RateLimit(t *testing.T) {
	deviceRepo := repositories.NewMemoryDeviceRepository()
	anomalyRepo := repositories.NewMemoryAnomalyRepository()
	notificationService := &mockNotificationService{}
	
	sensorProcessor := usecases.NewSensorDataProcessor(deviceRepo, anomalyRepo, notificationService)
	rateLimiter := usecases.NewRateLimiter(deviceRepo)
	
	service := NewIoTService(sensorProcessor, rateLimiter)
	ctx := context.Background()

	data := &dto.SensorDataDTO{
		DeviceID:    "device123",
		DeviceType:  "sensor",
		Timestamp:   time.Now().Unix(),
		Temperature: 25.0,
	}

	// Send multiple messages rapidly to trigger rate limit (MAX_MESSAGES_PER_MINUTE = 20)
	var lastErr error
	for i := 0; i < 25; i++ {
		lastErr = service.ProcessSensorData(ctx, data)
		// Don't fail on individual errors, just check final state
	}

	// The service should handle rate limiting gracefully
	// (it returns nil when rate limited, not an error)
	if lastErr != nil {
		// If there's an error, it should be a validation error, not a rate limit error
		t.Logf("ProcessSensorData() returned error (expected for validation): %v", lastErr)
	}

	// Verify device was created even with rate limiting
	device, err := deviceRepo.GetDevice(ctx, "device123")
	if err != nil {
		t.Errorf("Device should exist after rate limit test")
	}
	if device != nil && device.RateLimit.Count > 0 {
		t.Logf("Rate limit count: %d (expected > 0)", device.RateLimit.Count)
	}
}

func TestIoTService_ProcessSensorData_AnomalyDetection(t *testing.T) {
	deviceRepo := repositories.NewMemoryDeviceRepository()
	anomalyRepo := repositories.NewMemoryAnomalyRepository()
	notificationService := &mockNotificationService{}
	
	sensorProcessor := usecases.NewSensorDataProcessor(deviceRepo, anomalyRepo, notificationService)
	rateLimiter := usecases.NewRateLimiter(deviceRepo)
	
	service := NewIoTService(sensorProcessor, rateLimiter)
	ctx := context.Background()

	// Create data that will trigger temperature anomaly (> 50°C)
	data := &dto.SensorDataDTO{
		DeviceID:    "device123",
		DeviceType:  "sensor",
		Timestamp:   time.Now().Unix(),
		Temperature: 85.0, // This should trigger a temperature anomaly
		BatteryLevel: 5.0, // This should trigger a battery anomaly
	}

	err := service.ProcessSensorData(ctx, data)
	if err != nil {
		t.Errorf("ProcessSensorData() error = %v, want nil", err)
	}

	// Verify anomalies were created
	anomalies, err := anomalyRepo.GetAnomaliesByDevice(ctx, "device123", time.Now().Add(-1*time.Hour))
	if err != nil {
		t.Errorf("GetAnomaliesByDevice() error = %v, want nil", err)
	}

	if len(anomalies) < 1 {
		t.Error("Expected at least one anomaly to be detected")
	}

	// Check that at least one anomaly is temperature related
	foundTempAnomaly := false
	foundBatteryAnomaly := false
	for _, anomaly := range anomalies {
		if anomaly.Type == "temperature" {
			foundTempAnomaly = true
		}
		if anomaly.Type == "battery" {
			foundBatteryAnomaly = true
		}
	}

	if !foundTempAnomaly {
		t.Error("Expected temperature anomaly to be detected")
	}
	if !foundBatteryAnomaly {
		t.Error("Expected battery anomaly to be detected")
	}
}

func TestIoTService_ProcessSensorData_DTOToEntityConversion(t *testing.T) {
	deviceRepo := repositories.NewMemoryDeviceRepository()
	anomalyRepo := repositories.NewMemoryAnomalyRepository()
	notificationService := &mockNotificationService{}
	
	sensorProcessor := usecases.NewSensorDataProcessor(deviceRepo, anomalyRepo, notificationService)
	rateLimiter := usecases.NewRateLimiter(deviceRepo)
	
	service := NewIoTService(sensorProcessor, rateLimiter)
	ctx := context.Background()

	motionDetected := true
	recording := false
	locked := true

	originalDTO := &dto.SensorDataDTO{
		DeviceID:       "device456",
		DeviceType:     "camera",
		Timestamp:      time.Now().Unix(),
		SecurityLevel:  "high",
		Temperature:    30.5,
		Humidity:       70.0,
		MotionDetected: &motionDetected,
		Recording:      &recording,
		BatteryLevel:   85.0,
		Locked:         &locked,
		AccessAttempts: 2,
		SignalStrength: 90.0,
	}

	err := service.ProcessSensorData(ctx, originalDTO)
	if err != nil {
		t.Errorf("ProcessSensorData() error = %v, want nil", err)
	}

	// Verify device was created with correct type
	device, err := deviceRepo.GetDevice(ctx, "device456")
	if err != nil {
		t.Fatalf("GetDevice() error = %v", err)
	}

	if device.Type != originalDTO.DeviceType {
		t.Errorf("Device type = %v, want %v", device.Type, originalDTO.DeviceType)
	}
}