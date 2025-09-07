package dto

import "iot-hub-go/internal/domain/entities"

type SensorDataDTO struct {
	DeviceID       string  `json:"device_id"`
	DeviceType     string  `json:"device_type,omitempty"`
	Timestamp      int64   `json:"timestamp"`
	SecurityLevel  string  `json:"security_level,omitempty"`
	Temperature    float64 `json:"temperature,omitempty"`
	Humidity       float64 `json:"humidity,omitempty"`
	MotionDetected *bool   `json:"motion_detected,omitempty"`
	Recording      *bool   `json:"recording,omitempty"`
	BatteryLevel   float64 `json:"battery_level,omitempty"`
	Locked         *bool   `json:"locked,omitempty"`
	AccessAttempts int     `json:"access_attempts,omitempty"`
	SignalStrength float64 `json:"signal_strength,omitempty"`
}

func (dto *SensorDataDTO) ToEntity() *entities.SensorData {
	return &entities.SensorData{
		DeviceID:       dto.DeviceID,
		DeviceType:     dto.DeviceType,
		Timestamp:      dto.Timestamp,
		SecurityLevel:  dto.SecurityLevel,
		Temperature:    dto.Temperature,
		Humidity:       dto.Humidity,
		MotionDetected: dto.MotionDetected,
		Recording:      dto.Recording,
		BatteryLevel:   dto.BatteryLevel,
		Locked:         dto.Locked,
		AccessAttempts: dto.AccessAttempts,
		SignalStrength: dto.SignalStrength,
	}
}

func FromSensorDataEntity(entity *entities.SensorData) *SensorDataDTO {
	return &SensorDataDTO{
		DeviceID:       entity.DeviceID,
		DeviceType:     entity.DeviceType,
		Timestamp:      entity.Timestamp,
		SecurityLevel:  entity.SecurityLevel,
		Temperature:    entity.Temperature,
		Humidity:       entity.Humidity,
		MotionDetected: entity.MotionDetected,
		Recording:      entity.Recording,
		BatteryLevel:   entity.BatteryLevel,
		Locked:         entity.Locked,
		AccessAttempts: entity.AccessAttempts,
		SignalStrength: entity.SignalStrength,
	}
}