package notifications

import (
	"context"
	"log"
	"sync"
	
	"iot-hub-go/internal/domain/entities"
	"iot-hub-go/internal/domain/ports"
)

type NotificationManager struct {
	services []ports.NotificationService
	mu       sync.RWMutex
}

func NewNotificationManager() *NotificationManager {
	return &NotificationManager{
		services: make([]ports.NotificationService, 0),
	}
}

func (nm *NotificationManager) AddService(service ports.NotificationService) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	nm.services = append(nm.services, service)
}

func (nm *NotificationManager) SendAnomalyAlert(ctx context.Context, anomaly *entities.Anomaly) error {
	nm.mu.RLock()
	services := make([]ports.NotificationService, len(nm.services))
	copy(services, nm.services)
	nm.mu.RUnlock()
	
	var wg sync.WaitGroup
	for _, service := range services {
		wg.Add(1)
		go func(svc ports.NotificationService) {
			defer wg.Done()
			if err := svc.SendAnomalyAlert(ctx, anomaly); err != nil {
				log.Printf("Error enviando notificación de anomalía: %v", err)
			}
		}(service)
	}
	
	wg.Wait()
	return nil
}

func (nm *NotificationManager) SendQuarantineAlert(ctx context.Context, deviceID, reason string) error {
	nm.mu.RLock()
	services := make([]ports.NotificationService, len(nm.services))
	copy(services, nm.services)
	nm.mu.RUnlock()
	
	var wg sync.WaitGroup
	for _, service := range services {
		wg.Add(1)
		go func(svc ports.NotificationService) {
			defer wg.Done()
			if err := svc.SendQuarantineAlert(ctx, deviceID, reason); err != nil {
				log.Printf("Error enviando notificación de cuarentena: %v", err)
			}
		}(service)
	}
	
	wg.Wait()
	return nil
}