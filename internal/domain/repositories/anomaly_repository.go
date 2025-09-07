package repositories

import (
	"context"
	"time"
	
	"iot-hub-go/internal/domain/entities"
)

type AnomalyRepository interface {
	SaveAnomaly(ctx context.Context, anomaly *entities.Anomaly) error
	GetAnomaliesByDevice(ctx context.Context, deviceID string, since time.Time) ([]*entities.Anomaly, error)
	GetAnomaliesByType(ctx context.Context, anomalyType entities.AnomalyType, since time.Time) ([]*entities.Anomaly, error)
	CountAnomaliesByDevice(ctx context.Context, deviceID string, since time.Time) (int, error)
}