package ports

import (
	"context"
	
	"iot-hub-go/internal/domain/entities"
)

type NotificationService interface {
	SendAnomalyAlert(ctx context.Context, anomaly *entities.Anomaly) error
	SendQuarantineAlert(ctx context.Context, deviceID, reason string) error
}