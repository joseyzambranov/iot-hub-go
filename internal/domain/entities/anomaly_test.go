package entities

import (
	"strings"
	"testing"
	"time"
)

func TestNewAnomaly(t *testing.T) {
	tests := []struct {
		name          string
		deviceID      string
		anomalyType   AnomalyType
		description   string
		value         interface{}
		expectedType  AnomalyType
		expectedSeverity string
	}{
		{
			name:          "temperature anomaly",
			deviceID:      "device123",
			anomalyType:   AnomalyTemperature,
			description:   "Temperature too high",
			value:         150.0,
			expectedType:  AnomalyTemperature,
			expectedSeverity: "medium",
		},
		{
			name:          "battery anomaly",
			deviceID:      "device456",
			anomalyType:   AnomalyBattery,
			description:   "Low battery level",
			value:         5.0,
			expectedType:  AnomalyBattery,
			expectedSeverity: "medium",
		},
		{
			name:          "access attempts anomaly",
			deviceID:      "device789",
			anomalyType:   AnomalyAccessAttempts,
			description:   "Too many access attempts",
			value:         1500,
			expectedType:  AnomalyAccessAttempts,
			expectedSeverity: "medium",
		},
		{
			name:          "signal strength anomaly",
			deviceID:      "device999",
			anomalyType:   AnomalySignalStrength,
			description:   "Weak signal strength",
			value:         10.0,
			expectedType:  AnomalySignalStrength,
			expectedSeverity: "medium",
		},
		{
			name:          "behavior pattern anomaly",
			deviceID:      "device111",
			anomalyType:   AnomalyBehaviorPattern,
			description:   "Unusual behavior detected",
			value:         "pattern_xyz",
			expectedType:  AnomalyBehaviorPattern,
			expectedSeverity: "medium",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now()
			anomaly := NewAnomaly(tt.deviceID, tt.anomalyType, tt.description, tt.value)
			after := time.Now()

			// Check basic properties
			if anomaly.DeviceID != tt.deviceID {
				t.Errorf("NewAnomaly().DeviceID = %v, want %v", anomaly.DeviceID, tt.deviceID)
			}
			if anomaly.Type != tt.expectedType {
				t.Errorf("NewAnomaly().Type = %v, want %v", anomaly.Type, tt.expectedType)
			}
			if anomaly.Description != tt.description {
				t.Errorf("NewAnomaly().Description = %v, want %v", anomaly.Description, tt.description)
			}
			if anomaly.Value != tt.value {
				t.Errorf("NewAnomaly().Value = %v, want %v", anomaly.Value, tt.value)
			}
			if anomaly.Severity != tt.expectedSeverity {
				t.Errorf("NewAnomaly().Severity = %v, want %v", anomaly.Severity, tt.expectedSeverity)
			}

			// Check timestamp is set reasonably
			if anomaly.Timestamp.Before(before) || anomaly.Timestamp.After(after) {
				t.Errorf("NewAnomaly().Timestamp = %v, want between %v and %v", anomaly.Timestamp, before, after)
			}

			// Check ID format: deviceID_anomalyType_timestamp
			expectedIDPrefix := tt.deviceID + "_" + string(tt.anomalyType) + "_"
			if !strings.HasPrefix(anomaly.ID, expectedIDPrefix) {
				t.Errorf("NewAnomaly().ID = %v, want to start with %v", anomaly.ID, expectedIDPrefix)
			}

			// ID should be unique (contains timestamp)
			parts := strings.Split(anomaly.ID, "_")
			if len(parts) < 3 {
				t.Errorf("NewAnomaly().ID format incorrect: %v", anomaly.ID)
			}
		})
	}
}

func TestAnomalyType_Constants(t *testing.T) {
	// Test that all anomaly type constants are defined correctly
	expectedTypes := map[AnomalyType]string{
		AnomalyTemperature:     "temperature",
		AnomalyBattery:         "battery",
		AnomalyAccessAttempts:  "access_attempts",
		AnomalySignalStrength:  "signal_strength",
		AnomalyBehaviorPattern: "behavior_pattern",
	}

	for anomalyType, expectedValue := range expectedTypes {
		if string(anomalyType) != expectedValue {
			t.Errorf("AnomalyType %v = %v, want %v", anomalyType, string(anomalyType), expectedValue)
		}
	}
}

func TestAnomaly_Struct(t *testing.T) {
	now := time.Now()
	
	anomaly := &Anomaly{
		ID:          "device123_temperature_1609459200",
		DeviceID:    "device123",
		Type:        AnomalyTemperature,
		Description: "Temperature spike detected",
		Value:       85.5,
		Timestamp:   now,
		Severity:    "high",
	}

	// Test that all fields are properly set
	if anomaly.ID != "device123_temperature_1609459200" {
		t.Errorf("Anomaly.ID = %v, want device123_temperature_1609459200", anomaly.ID)
	}
	if anomaly.DeviceID != "device123" {
		t.Errorf("Anomaly.DeviceID = %v, want device123", anomaly.DeviceID)
	}
	if anomaly.Type != AnomalyTemperature {
		t.Errorf("Anomaly.Type = %v, want %v", anomaly.Type, AnomalyTemperature)
	}
	if anomaly.Description != "Temperature spike detected" {
		t.Errorf("Anomaly.Description = %v, want 'Temperature spike detected'", anomaly.Description)
	}
	if anomaly.Value != 85.5 {
		t.Errorf("Anomaly.Value = %v, want 85.5", anomaly.Value)
	}
	if !anomaly.Timestamp.Equal(now) {
		t.Errorf("Anomaly.Timestamp = %v, want %v", anomaly.Timestamp, now)
	}
	if anomaly.Severity != "high" {
		t.Errorf("Anomaly.Severity = %v, want high", anomaly.Severity)
	}
}

func TestNewAnomaly_DifferentValueTypes(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
	}{
		{"float64 value", 25.5},
		{"int value", 100},
		{"string value", "test_value"},
		{"bool value", true},
		{"nil value", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			anomaly := NewAnomaly("device123", AnomalyTemperature, "test", tt.value)
			if anomaly.Value != tt.value {
				t.Errorf("NewAnomaly().Value = %v, want %v", anomaly.Value, tt.value)
			}
		})
	}
}

func TestNewAnomaly_IDUniqueness(t *testing.T) {
	// Test that creating anomalies at different times generates different IDs
	anomaly1 := NewAnomaly("device123", AnomalyTemperature, "test1", 25.0)
	time.Sleep(2 * time.Second) // Longer delay to ensure different timestamp (unix seconds)
	anomaly2 := NewAnomaly("device123", AnomalyTemperature, "test2", 30.0)

	if anomaly1.ID == anomaly2.ID {
		t.Errorf("NewAnomaly() should generate unique IDs, got same ID: %v", anomaly1.ID)
	}

	// Test different device IDs generate different IDs
	anomaly3 := NewAnomaly("device456", AnomalyTemperature, "test", 25.0)
	if anomaly1.ID == anomaly3.ID {
		t.Errorf("NewAnomaly() with different device IDs should generate different IDs")
	}

	// Test different anomaly types generate different IDs
	anomaly4 := NewAnomaly("device123", AnomalyBattery, "test", 25.0)
	if anomaly1.ID == anomaly4.ID {
		t.Errorf("NewAnomaly() with different anomaly types should generate different IDs")
	}
}