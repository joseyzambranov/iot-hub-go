package entities

import (
	"testing"
	"time"
)

func TestSensorData_Validate(t *testing.T) {
	now := time.Now().Unix()

	tests := []struct {
		name    string
		data    *SensorData
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid sensor data",
			data: &SensorData{
				DeviceID:    "device123",
				Timestamp:   now,
				Temperature: 25.0,
				Humidity:    60.0,
				BatteryLevel: 80.0,
				SignalStrength: 75.0,
				AccessAttempts: 0,
			},
			wantErr: false,
		},
		{
			name: "empty device ID",
			data: &SensorData{
				DeviceID:  "",
				Timestamp: now,
			},
			wantErr: true,
			errMsg:  "device_id inválido",
		},
		{
			name: "device ID too long",
			data: &SensorData{
				DeviceID:  "this_is_a_very_long_device_id_that_exceeds_fifty_characters_limit",
				Timestamp: now,
			},
			wantErr: true,
			errMsg:  "device_id inválido",
		},
		{
			name: "timestamp too old",
			data: &SensorData{
				DeviceID:  "device123",
				Timestamp: now - 3700, // more than 1 hour ago
			},
			wantErr: true,
			errMsg:  "timestamp inválido",
		},
		{
			name: "timestamp too future",
			data: &SensorData{
				DeviceID:  "device123",
				Timestamp: now + 3700, // more than 1 hour in future
			},
			wantErr: true,
			errMsg:  "timestamp inválido",
		},
		{
			name: "temperature too low",
			data: &SensorData{
				DeviceID:    "device123",
				Timestamp:   now,
				Temperature: -60.0,
			},
			wantErr: true,
			errMsg:  "temperatura inválida",
		},
		{
			name: "temperature too high",
			data: &SensorData{
				DeviceID:    "device123",
				Timestamp:   now,
				Temperature: 120.0,
			},
			wantErr: true,
			errMsg:  "temperatura inválida",
		},
		{
			name: "humidity too low",
			data: &SensorData{
				DeviceID:  "device123",
				Timestamp: now,
				Humidity:  -10.0,
			},
			wantErr: true,
			errMsg:  "humedad inválida",
		},
		{
			name: "humidity too high",
			data: &SensorData{
				DeviceID:  "device123",
				Timestamp: now,
				Humidity:  110.0,
			},
			wantErr: true,
			errMsg:  "humedad inválida",
		},
		{
			name: "battery level too low",
			data: &SensorData{
				DeviceID:     "device123",
				Timestamp:    now,
				BatteryLevel: -10.0,
			},
			wantErr: true,
			errMsg:  "nivel de batería inválido",
		},
		{
			name: "battery level too high",
			data: &SensorData{
				DeviceID:     "device123",
				Timestamp:    now,
				BatteryLevel: 110.0,
			},
			wantErr: true,
			errMsg:  "nivel de batería inválido",
		},
		{
			name: "signal strength too low",
			data: &SensorData{
				DeviceID:       "device123",
				Timestamp:      now,
				SignalStrength: -10.0,
			},
			wantErr: true,
			errMsg:  "intensidad de señal inválida",
		},
		{
			name: "signal strength too high",
			data: &SensorData{
				DeviceID:       "device123",
				Timestamp:      now,
				SignalStrength: 110.0,
			},
			wantErr: true,
			errMsg:  "intensidad de señal inválida",
		},
		{
			name: "access attempts too low",
			data: &SensorData{
				DeviceID:       "device123",
				Timestamp:      now,
				AccessAttempts: -1,
			},
			wantErr: true,
			errMsg:  "intentos de acceso inválidos",
		},
		{
			name: "access attempts too high",
			data: &SensorData{
				DeviceID:       "device123",
				Timestamp:      now,
				AccessAttempts: 1001,
			},
			wantErr: true,
			errMsg:  "intentos de acceso inválidos",
		},
		{
			name: "zero values are valid",
			data: &SensorData{
				DeviceID:       "device123",
				Timestamp:      now,
				Temperature:    0,
				Humidity:       0,
				BatteryLevel:   0,
				SignalStrength: 0,
				AccessAttempts: 0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.data.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("SensorData.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if len(err.Error()) == 0 || err.Error()[:len(tt.errMsg)] != tt.errMsg {
					t.Errorf("SensorData.Validate() error message = %v, want to contain %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestSensorData_ValidateEdgeCases(t *testing.T) {
	now := time.Now().Unix()

	// Test boundary values
	boundaryTests := []struct {
		name string
		data *SensorData
		want bool
	}{
		{
			name: "temperature at lower boundary",
			data: &SensorData{
				DeviceID:    "device123",
				Timestamp:   now,
				Temperature: -50.0,
			},
			want: false, // should be valid
		},
		{
			name: "temperature at upper boundary",
			data: &SensorData{
				DeviceID:    "device123",
				Timestamp:   now,
				Temperature: 100.0,
			},
			want: false, // should be valid
		},
		{
			name: "humidity at boundaries",
			data: &SensorData{
				DeviceID:  "device123",
				Timestamp: now,
				Humidity:  100.0,
			},
			want: false, // should be valid
		},
		{
			name: "maximum access attempts",
			data: &SensorData{
				DeviceID:       "device123",
				Timestamp:      now,
				AccessAttempts: 1000,
			},
			want: false, // should be valid
		},
	}

	for _, tt := range boundaryTests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.data.Validate()
			if (err != nil) != tt.want {
				t.Errorf("SensorData.Validate() boundary test %s: error = %v, wantErr %v", tt.name, err, tt.want)
			}
		})
	}
}