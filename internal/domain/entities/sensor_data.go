package entities

import (
	"fmt"
	"time"
)

type SensorData struct {
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

func (s *SensorData) Validate() error {
	if s.DeviceID == "" || len(s.DeviceID) > 50 {
		return fmt.Errorf("device_id inválido: debe tener entre 1-50 caracteres")
	}
	
	now := time.Now().Unix()
	if s.Timestamp < now-3600 || s.Timestamp > now+3600 {
		return fmt.Errorf("timestamp inválido: %d fuera del rango permitido", s.Timestamp)
	}
	
	if s.Temperature != 0 {
		if s.Temperature < -50 || s.Temperature > 100 {
			return fmt.Errorf("temperatura inválida: %.2f fuera del rango -50°C a 100°C", s.Temperature)
		}
	}
	
	if s.Humidity != 0 {
		if s.Humidity < 0 || s.Humidity > 100 {
			return fmt.Errorf("humedad inválida: %.2f fuera del rango 0-100%%", s.Humidity)
		}
	}
	
	if s.BatteryLevel != 0 {
		if s.BatteryLevel < 0 || s.BatteryLevel > 100 {
			return fmt.Errorf("nivel de batería inválido: %.2f fuera del rango 0-100%%", s.BatteryLevel)
		}
	}
	
	if s.SignalStrength != 0 {
		if s.SignalStrength < 0 || s.SignalStrength > 100 {
			return fmt.Errorf("intensidad de señal inválida: %.2f fuera del rango 0-100%%", s.SignalStrength)
		}
	}
	
	if s.AccessAttempts < 0 || s.AccessAttempts > 1000 {
		return fmt.Errorf("intentos de acceso inválidos: %d fuera del rango 0-1000", s.AccessAttempts)
	}
	
	return nil
}