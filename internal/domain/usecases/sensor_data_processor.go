package usecases

import (
	"context"
	"fmt"
	"log"
	"time"
	
	"iot-hub-go/internal/domain/entities"
	"iot-hub-go/internal/domain/ports"
	"iot-hub-go/internal/domain/repositories"
	"iot-hub-go/internal/domain/services"
)

type SensorDataProcessor struct {
	deviceRepo       repositories.DeviceRepository
	anomalyRepo      repositories.AnomalyRepository
	notificationSvc  ports.NotificationService
	rateLimiter      *services.RateLimiter
}

func NewSensorDataProcessor(deviceRepo repositories.DeviceRepository, anomalyRepo repositories.AnomalyRepository, notificationSvc ports.NotificationService) *SensorDataProcessor {
	// Rate limiter: máximo 10 mensajes por minuto por dispositivo
	rateLimiter := services.NewRateLimiter(10, 1*time.Minute)
	
	return &SensorDataProcessor{
		deviceRepo:      deviceRepo,
		anomalyRepo:     anomalyRepo,
		notificationSvc: notificationSvc,
		rateLimiter:     rateLimiter,
	}
}

func (s *SensorDataProcessor) ProcessSensorData(ctx context.Context, data *entities.SensorData) error {
	// Verificar rate limiting
	if !s.rateLimiter.IsAllowed(data.DeviceID) {
		log.Printf("🚫 RATE LIMIT EXCEDIDO para %s - Rechazando mensaje", data.DeviceID)
		
		// Crear anomalía por rate limiting
		anomaly := entities.NewAnomaly(
			data.DeviceID,
			entities.AnomalyBehaviorPattern,
			fmt.Sprintf("rate limit excedido: más de 10 mensajes/minuto (actual: %d)", s.rateLimiter.GetRequestCount(data.DeviceID)),
			s.rateLimiter.GetRequestCount(data.DeviceID),
		)
		
		if err := s.anomalyRepo.SaveAnomaly(ctx, anomaly); err != nil {
			log.Printf("Error guardando anomalía de rate limit: %v", err)
		}
		
		if s.notificationSvc != nil {
			s.notificationSvc.SendAnomalyAlert(ctx, anomaly)
		}
		
		// Cuarentena automática por rate limit abuse
		reason := fmt.Sprintf("rate limit abuse: %d mensajes/minuto", s.rateLimiter.GetRequestCount(data.DeviceID))
		s.deviceRepo.QuarantineDevice(ctx, data.DeviceID, reason)
		if s.notificationSvc != nil {
			s.notificationSvc.SendQuarantineAlert(ctx, data.DeviceID, reason)
		}
		
		return fmt.Errorf("rate limit exceeded for device %s", data.DeviceID)
	}

	if err := data.Validate(); err != nil {
		log.Printf("⚠️ DATO INVÁLIDO de %s: %v", data.DeviceID, err)
		s.deviceRepo.QuarantineDevice(ctx, data.DeviceID, "datos inválidos")
		if s.notificationSvc != nil {
			s.notificationSvc.SendQuarantineAlert(ctx, data.DeviceID, "datos inválidos")
		}
		return err
	}
	
	device, err := s.deviceRepo.GetDevice(ctx, data.DeviceID)
	if err != nil {
		device = entities.NewDevice(data.DeviceID, data.DeviceType)
	} else if device.Type == "" && data.DeviceType != "" {
		device.Type = data.DeviceType
	}
	
	isQuarantined, err := s.deviceRepo.IsDeviceQuarantined(ctx, data.DeviceID)
	if err != nil {
		return fmt.Errorf("error checking quarantine status: %w", err)
	}
	
	if isQuarantined {
		log.Printf("🔒 MENSAJE RECHAZADO: Dispositivo %s está en cuarentena", data.DeviceID)
		return fmt.Errorf("device is quarantined")
	}
	
	anomalies := s.detectAnomalies(data)
	for _, anomaly := range anomalies {
		if err := s.anomalyRepo.SaveAnomaly(ctx, anomaly); err != nil {
			log.Printf("Error guardando anomalía: %v", err)
		}
		log.Printf("🚨 ANOMALÍA en %s: %s", data.DeviceID, anomaly.Description)
		if s.notificationSvc != nil {
			s.notificationSvc.SendAnomalyAlert(ctx, anomaly)
		}
	}
	
	behaviorAnomalies := s.analyzeBehavior(ctx, device, data)
	for _, anomaly := range behaviorAnomalies {
		if err := s.anomalyRepo.SaveAnomaly(ctx, anomaly); err != nil {
			log.Printf("Error guardando anomalía de comportamiento: %v", err)
		}
		log.Printf("🚨 PATRÓN SOSPECHOSO en %s: %s", data.DeviceID, anomaly.Description)
		if s.notificationSvc != nil {
			s.notificationSvc.SendAnomalyAlert(ctx, anomaly)
		}
	}
	
	if len(anomalies)+len(behaviorAnomalies) == 0 {
		log.Printf("✅ Datos de %s procesados y validados", data.DeviceID)
	}
	
	return s.deviceRepo.UpdateDevice(ctx, device)
}

func (s *SensorDataProcessor) detectAnomalies(data *entities.SensorData) []*entities.Anomaly {
	var anomalies []*entities.Anomaly
	
	if data.Temperature != 0 {
		if data.Temperature > 50 || data.Temperature < -10 {
			anomaly := entities.NewAnomaly(
				data.DeviceID,
				entities.AnomalyTemperature,
				fmt.Sprintf("temperatura extrema: %.2f°C", data.Temperature),
				data.Temperature,
			)
			anomalies = append(anomalies, anomaly)
		}
	}
	
	if data.BatteryLevel > 0 && data.BatteryLevel < 10 {
		anomaly := entities.NewAnomaly(
			data.DeviceID,
			entities.AnomalyBattery,
			fmt.Sprintf("batería crítica: %.1f%%", data.BatteryLevel),
			data.BatteryLevel,
		)
		anomalies = append(anomalies, anomaly)
	}
	
	if data.AccessAttempts > 5 {
		anomaly := entities.NewAnomaly(
			data.DeviceID,
			entities.AnomalyAccessAttempts,
			fmt.Sprintf("múltiples intentos de acceso: %d", data.AccessAttempts),
			data.AccessAttempts,
		)
		anomalies = append(anomalies, anomaly)
	}
	
	if data.SignalStrength > 0 && data.SignalStrength < 20 {
		anomaly := entities.NewAnomaly(
			data.DeviceID,
			entities.AnomalySignalStrength,
			fmt.Sprintf("señal débil: %.1f%%", data.SignalStrength),
			data.SignalStrength,
		)
		anomalies = append(anomalies, anomaly)
	}
	
	return anomalies
}

func (s *SensorDataProcessor) analyzeBehavior(ctx context.Context, device *entities.Device, data *entities.SensorData) []*entities.Anomaly {
	var anomalies []*entities.Anomaly
	
	behavior := device.Behavior
	behavior.MessageCount++
	
	if data.Temperature != 0 {
		if behavior.AvgTemperature == 0 {
			behavior.AvgTemperature = data.Temperature
		} else {
			oldAvg := behavior.AvgTemperature
			behavior.AvgTemperature = (behavior.AvgTemperature + data.Temperature) / 2
			
			tempDiff := data.Temperature - oldAvg
			if tempDiff > 20 || tempDiff < -20 {
				anomaly := entities.NewAnomaly(
					data.DeviceID,
					entities.AnomalyBehaviorPattern,
					fmt.Sprintf("cambio drástico temperatura: %.1f°C (promedio: %.1f°C)", data.Temperature, oldAvg),
					tempDiff,
				)
				anomalies = append(anomalies, anomaly)
				behavior.AnomalyCount++
			}
		}
	}
	
	if data.BatteryLevel > 0 {
		if behavior.AvgBattery == 0 {
			behavior.AvgBattery = data.BatteryLevel
		} else {
			oldAvg := behavior.AvgBattery
			behavior.AvgBattery = (behavior.AvgBattery + data.BatteryLevel) / 2
			
			batteryDiff := oldAvg - data.BatteryLevel
			if batteryDiff > 50 {
				anomaly := entities.NewAnomaly(
					data.DeviceID,
					entities.AnomalyBehaviorPattern,
					fmt.Sprintf("caída súbita batería: %.1f%% (promedio: %.1f%%)", data.BatteryLevel, oldAvg),
					batteryDiff,
				)
				anomalies = append(anomalies, anomaly)
				behavior.AnomalyCount++
			}
		}
	}
	
	if data.AccessAttempts > 0 {
		behavior.AccessAttempts = append(behavior.AccessAttempts, data.AccessAttempts)
		
		if len(behavior.AccessAttempts) > 10 {
			behavior.AccessAttempts = behavior.AccessAttempts[1:]
		}
		
		if len(behavior.AccessAttempts) >= 3 {
			recentAttempts := 0
			for _, attempts := range behavior.AccessAttempts[len(behavior.AccessAttempts)-3:] {
				recentAttempts += attempts
			}
			
			if recentAttempts > 20 {
				anomaly := entities.NewAnomaly(
					data.DeviceID,
					entities.AnomalyBehaviorPattern,
					fmt.Sprintf("posible ataque fuerza bruta: %d intentos en últimos 3 mensajes", recentAttempts),
					recentAttempts,
				)
				anomalies = append(anomalies, anomaly)
				behavior.AnomalyCount++
			}
		}
	}
	
	const ANOMALY_THRESHOLD = 3
	if behavior.AnomalyCount >= ANOMALY_THRESHOLD {
		reason := fmt.Sprintf("múltiples anomalías detectadas (%d)", behavior.AnomalyCount)
		s.deviceRepo.QuarantineDevice(ctx, data.DeviceID, reason)
		if s.notificationSvc != nil {
			s.notificationSvc.SendQuarantineAlert(ctx, data.DeviceID, reason)
		}
		behavior.AnomalyCount = 0
	}
	
	return anomalies
}