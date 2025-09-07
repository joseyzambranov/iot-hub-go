package entities

import (
	"fmt"
	"time"
)

type AnomalyType string

const (
	AnomalyTemperature     AnomalyType = "temperature"
	AnomalyBattery         AnomalyType = "battery"
	AnomalyAccessAttempts  AnomalyType = "access_attempts"
	AnomalySignalStrength  AnomalyType = "signal_strength"
	AnomalyBehaviorPattern AnomalyType = "behavior_pattern"
)

type Anomaly struct {
	ID          string
	DeviceID    string
	Type        AnomalyType
	Description string
	Value       interface{}
	Timestamp   time.Time
	Severity    string
}

func NewAnomaly(deviceID string, anomalyType AnomalyType, description string, value interface{}) *Anomaly {
	return &Anomaly{
		ID:          fmt.Sprintf("%s_%s_%d", deviceID, anomalyType, time.Now().Unix()),
		DeviceID:    deviceID,
		Type:        anomalyType,
		Description: description,
		Value:       value,
		Timestamp:   time.Now(),
		Severity:    "medium",
	}
}