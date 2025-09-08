package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"iot-hub-go/internal/domain/entities"
)

var ErrDeviceNotFound = errors.New("device not found")

// Mock repositories for testing
type mockDeviceRepository struct {
	devices        map[string]*entities.Device
	quarantined    map[string]string
}

func newMockDeviceRepository() *mockDeviceRepository {
	return &mockDeviceRepository{
		devices:     make(map[string]*entities.Device),
		quarantined: make(map[string]string),
	}
}

func (m *mockDeviceRepository) GetDevice(ctx context.Context, deviceID string) (*entities.Device, error) {
	device, exists := m.devices[deviceID]
	if !exists {
		return nil, ErrDeviceNotFound
	}
	return device, nil
}

func (m *mockDeviceRepository) SaveDevice(ctx context.Context, device *entities.Device) error {
	m.devices[device.ID] = device
	return nil
}

func (m *mockDeviceRepository) UpdateDevice(ctx context.Context, device *entities.Device) error {
	m.devices[device.ID] = device
	return nil
}

func (m *mockDeviceRepository) GetQuarantinedDevices(ctx context.Context) ([]*entities.Device, error) {
	var devices []*entities.Device
	for deviceID := range m.quarantined {
		if device, exists := m.devices[deviceID]; exists {
			devices = append(devices, device)
		}
	}
	return devices, nil
}

func (m *mockDeviceRepository) QuarantineDevice(ctx context.Context, deviceID, reason string) error {
	m.quarantined[deviceID] = reason
	return nil
}

func (m *mockDeviceRepository) IsDeviceQuarantined(ctx context.Context, deviceID string) (bool, error) {
	_, exists := m.quarantined[deviceID]
	return exists, nil
}

func (m *mockDeviceRepository) ReleaseFromQuarantine(ctx context.Context, deviceID string) error {
	delete(m.quarantined, deviceID)
	return nil
}

func (m *mockDeviceRepository) CleanExpiredQuarantines(ctx context.Context, duration time.Duration) error {
	return nil
}

type mockAnomalyRepository struct {
	anomalies []*entities.Anomaly
}

func newMockAnomalyRepository() *mockAnomalyRepository {
	return &mockAnomalyRepository{
		anomalies: make([]*entities.Anomaly, 0),
	}
}

func (m *mockAnomalyRepository) SaveAnomaly(ctx context.Context, anomaly *entities.Anomaly) error {
	m.anomalies = append(m.anomalies, anomaly)
	return nil
}

func (m *mockAnomalyRepository) GetAnomaliesByDevice(ctx context.Context, deviceID string, since time.Time) ([]*entities.Anomaly, error) {
	var deviceAnomalies []*entities.Anomaly
	for _, anomaly := range m.anomalies {
		if anomaly.DeviceID == deviceID && anomaly.Timestamp.After(since) {
			deviceAnomalies = append(deviceAnomalies, anomaly)
		}
	}
	return deviceAnomalies, nil
}

func (m *mockAnomalyRepository) GetAnomaliesByType(ctx context.Context, anomalyType entities.AnomalyType, since time.Time) ([]*entities.Anomaly, error) {
	var typeAnomalies []*entities.Anomaly
	for _, anomaly := range m.anomalies {
		if anomaly.Type == anomalyType && anomaly.Timestamp.After(since) {
			typeAnomalies = append(typeAnomalies, anomaly)
		}
	}
	return typeAnomalies, nil
}

func (m *mockAnomalyRepository) CountAnomaliesByDevice(ctx context.Context, deviceID string, since time.Time) (int, error) {
	count := 0
	for _, anomaly := range m.anomalies {
		if anomaly.DeviceID == deviceID && anomaly.Timestamp.After(since) {
			count++
		}
	}
	return count, nil
}

type mockNotificationService struct {
	anomalyAlerts     []*entities.Anomaly
	quarantineAlerts  []string
}

func newMockNotificationService() *mockNotificationService {
	return &mockNotificationService{
		anomalyAlerts:    make([]*entities.Anomaly, 0),
		quarantineAlerts: make([]string, 0),
	}
}

func (m *mockNotificationService) SendAnomalyAlert(ctx context.Context, anomaly *entities.Anomaly) error {
	m.anomalyAlerts = append(m.anomalyAlerts, anomaly)
	return nil
}

func (m *mockNotificationService) SendQuarantineAlert(ctx context.Context, deviceID, reason string) error {
	m.quarantineAlerts = append(m.quarantineAlerts, deviceID+": "+reason)
	return nil
}

func TestSensorDataProcessor_RateLimiting(t *testing.T) {
	deviceRepo := newMockDeviceRepository()
	anomalyRepo := newMockAnomalyRepository()
	notificationSvc := newMockNotificationService()
	
	processor := NewSensorDataProcessor(deviceRepo, anomalyRepo, notificationSvc)
	ctx := context.Background()
	
	deviceID := "test-device-001"
	now := time.Now().Unix()

	// Create valid sensor data
	sensorData := &entities.SensorData{
		DeviceID:    deviceID,
		Timestamp:   now,
		Temperature: 25.0,
	}

	// Make requests within rate limit (10 per minute)
	for i := 0; i < 10; i++ {
		err := processor.ProcessSensorData(ctx, sensorData)
		if err != nil {
			t.Errorf("Request %d should be allowed, but got error: %v", i+1, err)
		}
	}

	// 11th request should be blocked by rate limiter
	err := processor.ProcessSensorData(ctx, sensorData)
	if err == nil {
		t.Error("11th request should be blocked by rate limiter")
	}

	// Verify anomaly was created for rate limiting
	if len(anomalyRepo.anomalies) == 0 {
		t.Error("Expected anomaly to be created for rate limiting violation")
	}

	// Verify device was quarantined
	isQuarantined, _ := deviceRepo.IsDeviceQuarantined(ctx, deviceID)
	if !isQuarantined {
		t.Error("Device should be quarantined after rate limit violation")
	}

	// Verify notification was sent
	if len(notificationSvc.anomalyAlerts) == 0 {
		t.Error("Expected anomaly alert to be sent")
	}
	if len(notificationSvc.quarantineAlerts) == 0 {
		t.Error("Expected quarantine alert to be sent")
	}
}

func TestSensorDataProcessor_AnomalyDetection(t *testing.T) {
	deviceRepo := newMockDeviceRepository()
	anomalyRepo := newMockAnomalyRepository()
	notificationSvc := newMockNotificationService()
	
	processor := NewSensorDataProcessor(deviceRepo, anomalyRepo, notificationSvc)
	ctx := context.Background()
	
	tests := []struct {
		name           string
		sensorData     *entities.SensorData
		expectAnomaly  bool
		anomalyType    entities.AnomalyType
	}{
		{
			name: "extreme high temperature",
			sensorData: &entities.SensorData{
				DeviceID:    "temp-sensor-001",
				Timestamp:   time.Now().Unix(),
				Temperature: 75.0, // Above 50°C threshold
			},
			expectAnomaly: true,
			anomalyType:   entities.AnomalyTemperature,
		},
		{
			name: "extreme low temperature",
			sensorData: &entities.SensorData{
				DeviceID:    "temp-sensor-002",
				Timestamp:   time.Now().Unix(),
				Temperature: -15.0, // Below -10°C threshold
			},
			expectAnomaly: true,
			anomalyType:   entities.AnomalyTemperature,
		},
		{
			name: "critical battery level",
			sensorData: &entities.SensorData{
				DeviceID:     "battery-sensor-001",
				Timestamp:    time.Now().Unix(),
				BatteryLevel: 5.0, // Below 10% threshold
			},
			expectAnomaly: true,
			anomalyType:   entities.AnomalyBattery,
		},
		{
			name: "multiple access attempts",
			sensorData: &entities.SensorData{
				DeviceID:       "access-sensor-001",
				Timestamp:      time.Now().Unix(),
				AccessAttempts: 10, // Above 5 threshold
			},
			expectAnomaly: true,
			anomalyType:   entities.AnomalyAccessAttempts,
		},
		{
			name: "weak signal strength",
			sensorData: &entities.SensorData{
				DeviceID:       "signal-sensor-001",
				Timestamp:      time.Now().Unix(),
				SignalStrength: 15.0, // Below 20% threshold
			},
			expectAnomaly: true,
			anomalyType:   entities.AnomalySignalStrength,
		},
		{
			name: "normal sensor data",
			sensorData: &entities.SensorData{
				DeviceID:       "normal-sensor-001",
				Timestamp:      time.Now().Unix(),
				Temperature:    25.0,
				BatteryLevel:   80.0,
				AccessAttempts: 1,
				SignalStrength: 75.0,
			},
			expectAnomaly: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset anomaly repository for each test
			anomalyRepo.anomalies = make([]*entities.Anomaly, 0)
			
			err := processor.ProcessSensorData(ctx, tt.sensorData)
			if err != nil {
				t.Errorf("ProcessSensorData() error = %v", err)
				return
			}

			if tt.expectAnomaly {
				if len(anomalyRepo.anomalies) == 0 {
					t.Errorf("Expected anomaly to be detected, but none found")
					return
				}
				
				found := false
				for _, anomaly := range anomalyRepo.anomalies {
					if anomaly.Type == tt.anomalyType {
						found = true
						break
					}
				}
				
				if !found {
					t.Errorf("Expected anomaly type %v, but not found in anomalies", tt.anomalyType)
				}
			} else {
				if len(anomalyRepo.anomalies) > 0 {
					t.Errorf("Expected no anomalies, but found %d", len(anomalyRepo.anomalies))
				}
			}
		})
	}
}

func TestSensorDataProcessor_BehaviorAnalysis(t *testing.T) {
	deviceRepo := newMockDeviceRepository()
	anomalyRepo := newMockAnomalyRepository()
	notificationSvc := newMockNotificationService()
	
	processor := NewSensorDataProcessor(deviceRepo, anomalyRepo, notificationSvc)
	ctx := context.Background()
	
	deviceID := "behavior-test-device"
	now := time.Now().Unix()

	// First, send normal temperature data to establish baseline
	normalData := &entities.SensorData{
		DeviceID:    deviceID,
		Timestamp:   now,
		Temperature: 25.0,
	}
	
	err := processor.ProcessSensorData(ctx, normalData)
	if err != nil {
		t.Errorf("Error processing normal data: %v", err)
		return
	}

	// Reset anomalies to test behavior analysis
	anomalyRepo.anomalies = make([]*entities.Anomaly, 0)

	// Send data with drastic temperature change
	drasticChangeData := &entities.SensorData{
		DeviceID:    deviceID,
		Timestamp:   now + 1,
		Temperature: 50.0, // 25°C difference from baseline
	}
	
	err = processor.ProcessSensorData(ctx, drasticChangeData)
	if err != nil {
		t.Errorf("Error processing drastic change data: %v", err)
		return
	}

	// Should detect behavior pattern anomaly
	found := false
	for _, anomaly := range anomalyRepo.anomalies {
		if anomaly.Type == entities.AnomalyBehaviorPattern {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("Expected behavior pattern anomaly for drastic temperature change")
	}
}

func TestSensorDataProcessor_InvalidData(t *testing.T) {
	deviceRepo := newMockDeviceRepository()
	anomalyRepo := newMockAnomalyRepository()
	notificationSvc := newMockNotificationService()
	
	processor := NewSensorDataProcessor(deviceRepo, anomalyRepo, notificationSvc)
	ctx := context.Background()

	// Test with invalid sensor data
	invalidData := &entities.SensorData{
		DeviceID:    "", // Invalid empty device ID
		Timestamp:   time.Now().Unix(),
		Temperature: 25.0,
	}

	err := processor.ProcessSensorData(ctx, invalidData)
	if err == nil {
		t.Error("Expected error for invalid sensor data")
	}

	// Verify quarantine alert was sent
	if len(notificationSvc.quarantineAlerts) == 0 {
		t.Error("Expected quarantine alert for invalid data")
	}
}

func TestSensorDataProcessor_QuarantinedDevice(t *testing.T) {
	deviceRepo := newMockDeviceRepository()
	anomalyRepo := newMockAnomalyRepository()
	notificationSvc := newMockNotificationService()
	
	processor := NewSensorDataProcessor(deviceRepo, anomalyRepo, notificationSvc)
	ctx := context.Background()
	
	deviceID := "quarantined-device"
	
	// Quarantine the device first
	deviceRepo.QuarantineDevice(ctx, deviceID, "test quarantine")

	// Try to process data from quarantined device
	sensorData := &entities.SensorData{
		DeviceID:    deviceID,
		Timestamp:   time.Now().Unix(),
		Temperature: 25.0,
	}

	err := processor.ProcessSensorData(ctx, sensorData)
	if err == nil {
		t.Error("Expected error when processing data from quarantined device")
	}
}