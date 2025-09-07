package repositories

import (
	"context"
	"testing"
	"time"

	"iot-hub-go/internal/domain/entities"
)

func TestNewMemoryAnomalyRepository(t *testing.T) {
	repo := NewMemoryAnomalyRepository()

	if repo == nil {
		t.Fatal("NewMemoryAnomalyRepository() returned nil")
	}

	// Check that it's actually the correct type
	memRepo, ok := repo.(*MemoryAnomalyRepository)
	if !ok {
		t.Fatal("NewMemoryAnomalyRepository() did not return *MemoryAnomalyRepository")
	}

	if memRepo.anomalies == nil {
		t.Error("NewMemoryAnomalyRepository() anomalies map is nil")
	}
}

func TestMemoryAnomalyRepository_SaveAnomaly(t *testing.T) {
	repo := NewMemoryAnomalyRepository()
	ctx := context.Background()

	anomaly := entities.NewAnomaly("device123", entities.AnomalyTemperature, "High temperature", 85.5)

	err := repo.SaveAnomaly(ctx, anomaly)
	if err != nil {
		t.Errorf("SaveAnomaly() error = %v, want nil", err)
	}

	// Verify the anomaly was saved by checking internal state
	memRepo := repo.(*MemoryAnomalyRepository)
	memRepo.mutex.RLock()
	savedAnomaly, exists := memRepo.anomalies[anomaly.ID]
	memRepo.mutex.RUnlock()

	if !exists {
		t.Error("SaveAnomaly() anomaly not found in repository")
	}
	if savedAnomaly.ID != anomaly.ID {
		t.Errorf("SaveAnomaly() saved ID = %v, want %v", savedAnomaly.ID, anomaly.ID)
	}
	if savedAnomaly.DeviceID != anomaly.DeviceID {
		t.Errorf("SaveAnomaly() saved DeviceID = %v, want %v", savedAnomaly.DeviceID, anomaly.DeviceID)
	}
}

func TestMemoryAnomalyRepository_GetAnomaliesByDevice(t *testing.T) {
	repo := NewMemoryAnomalyRepository()
	ctx := context.Background()

	// Create anomalies for different devices and times
	now := time.Now()
	past := now.Add(-2 * time.Hour)

	anomaly1 := entities.NewAnomaly("device123", entities.AnomalyTemperature, "High temp", 85.0)
	anomaly1.Timestamp = now.Add(-1 * time.Hour) // 1 hour ago

	anomaly2 := entities.NewAnomaly("device123", entities.AnomalyBattery, "Low battery", 5.0)
	anomaly2.Timestamp = now.Add(-30 * time.Minute) // 30 minutes ago

	anomaly3 := entities.NewAnomaly("device456", entities.AnomalyTemperature, "High temp", 90.0)
	anomaly3.Timestamp = now.Add(-30 * time.Minute) // 30 minutes ago

	anomaly4 := entities.NewAnomaly("device123", entities.AnomalySignalStrength, "Weak signal", 10.0)
	anomaly4.Timestamp = past.Add(-1 * time.Hour) // 3 hours ago (before 'since' time)

	// Save all anomalies
	repo.SaveAnomaly(ctx, anomaly1)
	repo.SaveAnomaly(ctx, anomaly2)
	repo.SaveAnomaly(ctx, anomaly3)
	repo.SaveAnomaly(ctx, anomaly4)

	// Get anomalies for device123 since 2 hours ago
	anomalies, err := repo.GetAnomaliesByDevice(ctx, "device123", past)
	if err != nil {
		t.Errorf("GetAnomaliesByDevice() error = %v, want nil", err)
	}

	// Should return anomaly1 and anomaly2 (not anomaly3 which is different device, not anomaly4 which is too old)
	if len(anomalies) != 2 {
		t.Errorf("GetAnomaliesByDevice() count = %v, want 2", len(anomalies))
	}

	// Verify correct anomalies returned
	foundIDs := make(map[string]bool)
	for _, anomaly := range anomalies {
		foundIDs[anomaly.ID] = true
		if anomaly.DeviceID != "device123" {
			t.Errorf("GetAnomaliesByDevice() returned anomaly for device %v, want device123", anomaly.DeviceID)
		}
	}

	if !foundIDs[anomaly1.ID] {
		t.Error("GetAnomaliesByDevice() missing anomaly1")
	}
	if !foundIDs[anomaly2.ID] {
		t.Error("GetAnomaliesByDevice() missing anomaly2")
	}
}

func TestMemoryAnomalyRepository_GetAnomaliesByType(t *testing.T) {
	repo := NewMemoryAnomalyRepository()
	ctx := context.Background()

	now := time.Now()
	past := now.Add(-2 * time.Hour)

	// Create anomalies of different types and times
	tempAnomaly1 := entities.NewAnomaly("device123", entities.AnomalyTemperature, "High temp", 85.0)
	tempAnomaly1.Timestamp = now.Add(-1 * time.Hour)

	tempAnomaly2 := entities.NewAnomaly("device456", entities.AnomalyTemperature, "High temp", 90.0)
	tempAnomaly2.Timestamp = now.Add(-30 * time.Minute)

	batteryAnomaly := entities.NewAnomaly("device123", entities.AnomalyBattery, "Low battery", 5.0)
	batteryAnomaly.Timestamp = now.Add(-30 * time.Minute)

	oldTempAnomaly := entities.NewAnomaly("device789", entities.AnomalyTemperature, "High temp", 95.0)
	oldTempAnomaly.Timestamp = past.Add(-1 * time.Hour) // 3 hours ago (before 'since' time)

	// Save all anomalies
	repo.SaveAnomaly(ctx, tempAnomaly1)
	repo.SaveAnomaly(ctx, tempAnomaly2)
	repo.SaveAnomaly(ctx, batteryAnomaly)
	repo.SaveAnomaly(ctx, oldTempAnomaly)

	// Get temperature anomalies since 2 hours ago
	anomalies, err := repo.GetAnomaliesByType(ctx, entities.AnomalyTemperature, past)
	if err != nil {
		t.Errorf("GetAnomaliesByType() error = %v, want nil", err)
	}

	// Should return tempAnomaly1 and tempAnomaly2 (not batteryAnomaly which is different type, not oldTempAnomaly which is too old)
	if len(anomalies) != 2 {
		t.Errorf("GetAnomaliesByType() count = %v, want 2", len(anomalies))
	}

	// Verify correct anomalies returned
	foundIDs := make(map[string]bool)
	for _, anomaly := range anomalies {
		foundIDs[anomaly.ID] = true
		if anomaly.Type != entities.AnomalyTemperature {
			t.Errorf("GetAnomaliesByType() returned anomaly of type %v, want %v", anomaly.Type, entities.AnomalyTemperature)
		}
	}

	if !foundIDs[tempAnomaly1.ID] {
		t.Error("GetAnomaliesByType() missing tempAnomaly1")
	}
	if !foundIDs[tempAnomaly2.ID] {
		t.Error("GetAnomaliesByType() missing tempAnomaly2")
	}
}

func TestMemoryAnomalyRepository_CountAnomaliesByDevice(t *testing.T) {
	repo := NewMemoryAnomalyRepository()
	ctx := context.Background()

	now := time.Now()
	past := now.Add(-2 * time.Hour)

	// Create anomalies for different devices and times
	anomaly1 := entities.NewAnomaly("device123", entities.AnomalyTemperature, "High temp", 85.0)
	anomaly1.Timestamp = now.Add(-1 * time.Hour)

	anomaly2 := entities.NewAnomaly("device123", entities.AnomalyBattery, "Low battery", 5.0)
	anomaly2.Timestamp = now.Add(-30 * time.Minute)

	anomaly3 := entities.NewAnomaly("device456", entities.AnomalyTemperature, "High temp", 90.0)
	anomaly3.Timestamp = now.Add(-30 * time.Minute)

	oldAnomaly := entities.NewAnomaly("device123", entities.AnomalySignalStrength, "Weak signal", 10.0)
	oldAnomaly.Timestamp = past.Add(-1 * time.Hour) // 3 hours ago (before 'since' time)

	// Save all anomalies
	repo.SaveAnomaly(ctx, anomaly1)
	repo.SaveAnomaly(ctx, anomaly2)
	repo.SaveAnomaly(ctx, anomaly3)
	repo.SaveAnomaly(ctx, oldAnomaly)

	// Count anomalies for device123 since 2 hours ago
	count, err := repo.CountAnomaliesByDevice(ctx, "device123", past)
	if err != nil {
		t.Errorf("CountAnomaliesByDevice() error = %v, want nil", err)
	}

	// Should count anomaly1 and anomaly2 (not anomaly3 which is different device, not oldAnomaly which is too old)
	if count != 2 {
		t.Errorf("CountAnomaliesByDevice() count = %v, want 2", count)
	}

	// Test with device that has no anomalies
	count, err = repo.CountAnomaliesByDevice(ctx, "nonexistent", past)
	if err != nil {
		t.Errorf("CountAnomaliesByDevice() error = %v, want nil", err)
	}
	if count != 0 {
		t.Errorf("CountAnomaliesByDevice() for nonexistent device count = %v, want 0", count)
	}
}

func TestMemoryAnomalyRepository_EmptyRepository(t *testing.T) {
	repo := NewMemoryAnomalyRepository()
	ctx := context.Background()
	past := time.Now().Add(-1 * time.Hour)

	// Test operations on empty repository
	anomalies, err := repo.GetAnomaliesByDevice(ctx, "device123", past)
	if err != nil {
		t.Errorf("GetAnomaliesByDevice() on empty repo error = %v, want nil", err)
	}
	if len(anomalies) != 0 {
		t.Errorf("GetAnomaliesByDevice() on empty repo count = %v, want 0", len(anomalies))
	}

	anomalies, err = repo.GetAnomaliesByType(ctx, entities.AnomalyTemperature, past)
	if err != nil {
		t.Errorf("GetAnomaliesByType() on empty repo error = %v, want nil", err)
	}
	if len(anomalies) != 0 {
		t.Errorf("GetAnomaliesByType() on empty repo count = %v, want 0", len(anomalies))
	}

	count, err := repo.CountAnomaliesByDevice(ctx, "device123", past)
	if err != nil {
		t.Errorf("CountAnomaliesByDevice() on empty repo error = %v, want nil", err)
	}
	if count != 0 {
		t.Errorf("CountAnomaliesByDevice() on empty repo count = %v, want 0", count)
	}
}

func TestMemoryAnomalyRepository_TimeFiltering(t *testing.T) {
	repo := NewMemoryAnomalyRepository()
	ctx := context.Background()

	now := time.Now()

	// Create anomaly exactly at the 'since' time
	exactTimeAnomaly := entities.NewAnomaly("device123", entities.AnomalyTemperature, "Exact time", 85.0)
	exactTimeAnomaly.Timestamp = now

	// Create anomaly just before the 'since' time
	beforeTimeAnomaly := entities.NewAnomaly("device123", entities.AnomalyBattery, "Before time", 5.0)
	beforeTimeAnomaly.Timestamp = now.Add(-1 * time.Nanosecond)

	// Create anomaly just after the 'since' time
	afterTimeAnomaly := entities.NewAnomaly("device123", entities.AnomalySignalStrength, "After time", 10.0)
	afterTimeAnomaly.Timestamp = now.Add(1 * time.Nanosecond)

	repo.SaveAnomaly(ctx, exactTimeAnomaly)
	repo.SaveAnomaly(ctx, beforeTimeAnomaly)
	repo.SaveAnomaly(ctx, afterTimeAnomaly)

	// Get anomalies since 'now' (should only include afterTimeAnomaly)
	anomalies, err := repo.GetAnomaliesByDevice(ctx, "device123", now)
	if err != nil {
		t.Errorf("GetAnomaliesByDevice() error = %v, want nil", err)
	}

	if len(anomalies) != 1 {
		t.Errorf("GetAnomaliesByDevice() with exact time filtering count = %v, want 1", len(anomalies))
	}

	if len(anomalies) > 0 && anomalies[0].ID != afterTimeAnomaly.ID {
		t.Errorf("GetAnomaliesByDevice() returned wrong anomaly, got %v, want %v", anomalies[0].ID, afterTimeAnomaly.ID)
	}
}

func TestMemoryAnomalyRepository_ConcurrentAccess(t *testing.T) {
	repo := NewMemoryAnomalyRepository()
	ctx := context.Background()

	done := make(chan bool, 3)
	past := time.Now().Add(-1 * time.Hour)

	// Goroutine 1: Save anomalies
	go func() {
		for i := 0; i < 100; i++ {
			anomaly := entities.NewAnomaly("device123", entities.AnomalyTemperature, "Test", float64(i))
			repo.SaveAnomaly(ctx, anomaly)
		}
		done <- true
	}()

	// Goroutine 2: Read by device
	go func() {
		for i := 0; i < 100; i++ {
			repo.GetAnomaliesByDevice(ctx, "device123", past)
		}
		done <- true
	}()

	// Goroutine 3: Read by type and count
	go func() {
		for i := 0; i < 100; i++ {
			repo.GetAnomaliesByType(ctx, entities.AnomalyTemperature, past)
			repo.CountAnomaliesByDevice(ctx, "device123", past)
		}
		done <- true
	}()

	// Wait for all goroutines
	<-done
	<-done
	<-done

	t.Log("Concurrent access test completed successfully")
}