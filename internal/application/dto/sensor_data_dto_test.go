package dto

import (
	"testing"
	"time"

	"iot-hub-go/internal/domain/entities"
)

func TestSensorDataDTO_ToEntity(t *testing.T) {
	timestamp := time.Now().Unix()
	motionDetected := true
	recording := false
	locked := true

	dto := &SensorDataDTO{
		DeviceID:       "device123",
		DeviceType:     "sensor",
		Timestamp:      timestamp,
		SecurityLevel:  "high",
		Temperature:    25.5,
		Humidity:       60.0,
		MotionDetected: &motionDetected,
		Recording:      &recording,
		BatteryLevel:   80.0,
		Locked:         &locked,
		AccessAttempts: 0,
		SignalStrength: 75.0,
	}

	entity := dto.ToEntity()

	// Verify all fields are correctly mapped
	if entity.DeviceID != dto.DeviceID {
		t.Errorf("ToEntity().DeviceID = %v, want %v", entity.DeviceID, dto.DeviceID)
	}
	if entity.DeviceType != dto.DeviceType {
		t.Errorf("ToEntity().DeviceType = %v, want %v", entity.DeviceType, dto.DeviceType)
	}
	if entity.Timestamp != dto.Timestamp {
		t.Errorf("ToEntity().Timestamp = %v, want %v", entity.Timestamp, dto.Timestamp)
	}
	if entity.SecurityLevel != dto.SecurityLevel {
		t.Errorf("ToEntity().SecurityLevel = %v, want %v", entity.SecurityLevel, dto.SecurityLevel)
	}
	if entity.Temperature != dto.Temperature {
		t.Errorf("ToEntity().Temperature = %v, want %v", entity.Temperature, dto.Temperature)
	}
	if entity.Humidity != dto.Humidity {
		t.Errorf("ToEntity().Humidity = %v, want %v", entity.Humidity, dto.Humidity)
	}
	if entity.MotionDetected != dto.MotionDetected {
		t.Errorf("ToEntity().MotionDetected = %v, want %v", entity.MotionDetected, dto.MotionDetected)
	}
	if entity.Recording != dto.Recording {
		t.Errorf("ToEntity().Recording = %v, want %v", entity.Recording, dto.Recording)
	}
	if entity.BatteryLevel != dto.BatteryLevel {
		t.Errorf("ToEntity().BatteryLevel = %v, want %v", entity.BatteryLevel, dto.BatteryLevel)
	}
	if entity.Locked != dto.Locked {
		t.Errorf("ToEntity().Locked = %v, want %v", entity.Locked, dto.Locked)
	}
	if entity.AccessAttempts != dto.AccessAttempts {
		t.Errorf("ToEntity().AccessAttempts = %v, want %v", entity.AccessAttempts, dto.AccessAttempts)
	}
	if entity.SignalStrength != dto.SignalStrength {
		t.Errorf("ToEntity().SignalStrength = %v, want %v", entity.SignalStrength, dto.SignalStrength)
	}
}

func TestFromSensorDataEntity(t *testing.T) {
	timestamp := time.Now().Unix()
	motionDetected := false
	recording := true
	locked := false

	entity := &entities.SensorData{
		DeviceID:       "device456",
		DeviceType:     "camera",
		Timestamp:      timestamp,
		SecurityLevel:  "medium",
		Temperature:    22.0,
		Humidity:       55.0,
		MotionDetected: &motionDetected,
		Recording:      &recording,
		BatteryLevel:   90.0,
		Locked:         &locked,
		AccessAttempts: 2,
		SignalStrength: 85.0,
	}

	dto := FromSensorDataEntity(entity)

	// Verify all fields are correctly mapped
	if dto.DeviceID != entity.DeviceID {
		t.Errorf("FromSensorDataEntity().DeviceID = %v, want %v", dto.DeviceID, entity.DeviceID)
	}
	if dto.DeviceType != entity.DeviceType {
		t.Errorf("FromSensorDataEntity().DeviceType = %v, want %v", dto.DeviceType, entity.DeviceType)
	}
	if dto.Timestamp != entity.Timestamp {
		t.Errorf("FromSensorDataEntity().Timestamp = %v, want %v", dto.Timestamp, entity.Timestamp)
	}
	if dto.SecurityLevel != entity.SecurityLevel {
		t.Errorf("FromSensorDataEntity().SecurityLevel = %v, want %v", dto.SecurityLevel, entity.SecurityLevel)
	}
	if dto.Temperature != entity.Temperature {
		t.Errorf("FromSensorDataEntity().Temperature = %v, want %v", dto.Temperature, entity.Temperature)
	}
	if dto.Humidity != entity.Humidity {
		t.Errorf("FromSensorDataEntity().Humidity = %v, want %v", dto.Humidity, entity.Humidity)
	}
	if dto.MotionDetected != entity.MotionDetected {
		t.Errorf("FromSensorDataEntity().MotionDetected = %v, want %v", dto.MotionDetected, entity.MotionDetected)
	}
	if dto.Recording != entity.Recording {
		t.Errorf("FromSensorDataEntity().Recording = %v, want %v", dto.Recording, entity.Recording)
	}
	if dto.BatteryLevel != entity.BatteryLevel {
		t.Errorf("FromSensorDataEntity().BatteryLevel = %v, want %v", dto.BatteryLevel, entity.BatteryLevel)
	}
	if dto.Locked != entity.Locked {
		t.Errorf("FromSensorDataEntity().Locked = %v, want %v", dto.Locked, entity.Locked)
	}
	if dto.AccessAttempts != entity.AccessAttempts {
		t.Errorf("FromSensorDataEntity().AccessAttempts = %v, want %v", dto.AccessAttempts, entity.AccessAttempts)
	}
	if dto.SignalStrength != entity.SignalStrength {
		t.Errorf("FromSensorDataEntity().SignalStrength = %v, want %v", dto.SignalStrength, entity.SignalStrength)
	}
}

func TestSensorDataDTO_ToEntity_WithNilPointers(t *testing.T) {
	dto := &SensorDataDTO{
		DeviceID:       "device123",
		DeviceType:     "sensor",
		Timestamp:      time.Now().Unix(),
		MotionDetected: nil,
		Recording:      nil,
		Locked:         nil,
	}

	entity := dto.ToEntity()

	if entity.MotionDetected != nil {
		t.Errorf("ToEntity().MotionDetected = %v, want nil", entity.MotionDetected)
	}
	if entity.Recording != nil {
		t.Errorf("ToEntity().Recording = %v, want nil", entity.Recording)
	}
	if entity.Locked != nil {
		t.Errorf("ToEntity().Locked = %v, want nil", entity.Locked)
	}
}

func TestFromSensorDataEntity_WithNilPointers(t *testing.T) {
	entity := &entities.SensorData{
		DeviceID:       "device456",
		DeviceType:     "camera",
		Timestamp:      time.Now().Unix(),
		MotionDetected: nil,
		Recording:      nil,
		Locked:         nil,
	}

	dto := FromSensorDataEntity(entity)

	if dto.MotionDetected != nil {
		t.Errorf("FromSensorDataEntity().MotionDetected = %v, want nil", dto.MotionDetected)
	}
	if dto.Recording != nil {
		t.Errorf("FromSensorDataEntity().Recording = %v, want nil", dto.Recording)
	}
	if dto.Locked != nil {
		t.Errorf("FromSensorDataEntity().Locked = %v, want nil", dto.Locked)
	}
}

func TestSensorDataDTO_BiDirectionalMapping(t *testing.T) {
	timestamp := time.Now().Unix()
	motionDetected := true
	recording := false
	locked := true

	originalDTO := &SensorDataDTO{
		DeviceID:       "device789",
		DeviceType:     "smart_lock",
		Timestamp:      timestamp,
		SecurityLevel:  "high",
		Temperature:    28.5,
		Humidity:       45.0,
		MotionDetected: &motionDetected,
		Recording:      &recording,
		BatteryLevel:   65.0,
		Locked:         &locked,
		AccessAttempts: 3,
		SignalStrength: 95.0,
	}

	// Convert DTO -> Entity -> DTO
	entity := originalDTO.ToEntity()
	resultDTO := FromSensorDataEntity(entity)

	// Verify the result matches the original
	if resultDTO.DeviceID != originalDTO.DeviceID {
		t.Errorf("Bidirectional mapping DeviceID = %v, want %v", resultDTO.DeviceID, originalDTO.DeviceID)
	}
	if resultDTO.DeviceType != originalDTO.DeviceType {
		t.Errorf("Bidirectional mapping DeviceType = %v, want %v", resultDTO.DeviceType, originalDTO.DeviceType)
	}
	if resultDTO.Timestamp != originalDTO.Timestamp {
		t.Errorf("Bidirectional mapping Timestamp = %v, want %v", resultDTO.Timestamp, originalDTO.Timestamp)
	}
	if resultDTO.Temperature != originalDTO.Temperature {
		t.Errorf("Bidirectional mapping Temperature = %v, want %v", resultDTO.Temperature, originalDTO.Temperature)
	}
	if resultDTO.Humidity != originalDTO.Humidity {
		t.Errorf("Bidirectional mapping Humidity = %v, want %v", resultDTO.Humidity, originalDTO.Humidity)
	}
	if *resultDTO.MotionDetected != *originalDTO.MotionDetected {
		t.Errorf("Bidirectional mapping MotionDetected = %v, want %v", *resultDTO.MotionDetected, *originalDTO.MotionDetected)
	}
	if *resultDTO.Recording != *originalDTO.Recording {
		t.Errorf("Bidirectional mapping Recording = %v, want %v", *resultDTO.Recording, *originalDTO.Recording)
	}
	if resultDTO.BatteryLevel != originalDTO.BatteryLevel {
		t.Errorf("Bidirectional mapping BatteryLevel = %v, want %v", resultDTO.BatteryLevel, originalDTO.BatteryLevel)
	}
	if *resultDTO.Locked != *originalDTO.Locked {
		t.Errorf("Bidirectional mapping Locked = %v, want %v", *resultDTO.Locked, *originalDTO.Locked)
	}
	if resultDTO.AccessAttempts != originalDTO.AccessAttempts {
		t.Errorf("Bidirectional mapping AccessAttempts = %v, want %v", resultDTO.AccessAttempts, originalDTO.AccessAttempts)
	}
	if resultDTO.SignalStrength != originalDTO.SignalStrength {
		t.Errorf("Bidirectional mapping SignalStrength = %v, want %v", resultDTO.SignalStrength, originalDTO.SignalStrength)
	}
}