package repositories

import (
	"context"
	"sync"
	"time"
	
	"iot-hub-go/internal/domain/entities"
	"iot-hub-go/internal/domain/repositories"
)

type MemoryAnomalyRepository struct {
	anomalies map[string]*entities.Anomaly
	mutex     sync.RWMutex
}

func NewMemoryAnomalyRepository() repositories.AnomalyRepository {
	return &MemoryAnomalyRepository{
		anomalies: make(map[string]*entities.Anomaly),
	}
}

func (r *MemoryAnomalyRepository) SaveAnomaly(ctx context.Context, anomaly *entities.Anomaly) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	r.anomalies[anomaly.ID] = anomaly
	return nil
}

func (r *MemoryAnomalyRepository) GetAnomaliesByDevice(ctx context.Context, deviceID string, since time.Time) ([]*entities.Anomaly, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	var result []*entities.Anomaly
	for _, anomaly := range r.anomalies {
		if anomaly.DeviceID == deviceID && anomaly.Timestamp.After(since) {
			result = append(result, anomaly)
		}
	}
	
	return result, nil
}

func (r *MemoryAnomalyRepository) GetAnomaliesByType(ctx context.Context, anomalyType entities.AnomalyType, since time.Time) ([]*entities.Anomaly, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	var result []*entities.Anomaly
	for _, anomaly := range r.anomalies {
		if anomaly.Type == anomalyType && anomaly.Timestamp.After(since) {
			result = append(result, anomaly)
		}
	}
	
	return result, nil
}

func (r *MemoryAnomalyRepository) CountAnomaliesByDevice(ctx context.Context, deviceID string, since time.Time) (int, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	count := 0
	for _, anomaly := range r.anomalies {
		if anomaly.DeviceID == deviceID && anomaly.Timestamp.After(since) {
			count++
		}
	}
	
	return count, nil
}