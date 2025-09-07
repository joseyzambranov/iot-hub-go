package repositories

import (
	"context"
	"testing"
	"time"

	"iot-hub-go/internal/domain/entities"
)

func TestNewMemoryDeviceRepository(t *testing.T) {
	repo := NewMemoryDeviceRepository()

	if repo == nil {
		t.Fatal("NewMemoryDeviceRepository() returned nil")
	}

	// Check that it's actually the correct type
	memRepo, ok := repo.(*MemoryDeviceRepository)
	if !ok {
		t.Fatal("NewMemoryDeviceRepository() did not return *MemoryDeviceRepository")
	}

	if memRepo.devices == nil {
		t.Error("NewMemoryDeviceRepository() devices map is nil")
	}
	if memRepo.quarantinedDevices == nil {
		t.Error("NewMemoryDeviceRepository() quarantinedDevices map is nil")
	}
}

func TestMemoryDeviceRepository_SaveAndGetDevice(t *testing.T) {
	repo := NewMemoryDeviceRepository()
	ctx := context.Background()

	device := entities.NewDevice("device123", "sensor")

	// Test SaveDevice
	err := repo.SaveDevice(ctx, device)
	if err != nil {
		t.Errorf("SaveDevice() error = %v, want nil", err)
	}

	// Test GetDevice
	retrievedDevice, err := repo.GetDevice(ctx, "device123")
	if err != nil {
		t.Errorf("GetDevice() error = %v, want nil", err)
	}
	if retrievedDevice == nil {
		t.Fatal("GetDevice() returned nil device")
	}
	if retrievedDevice.ID != device.ID {
		t.Errorf("GetDevice() ID = %v, want %v", retrievedDevice.ID, device.ID)
	}
	if retrievedDevice.Type != device.Type {
		t.Errorf("GetDevice() Type = %v, want %v", retrievedDevice.Type, device.Type)
	}
}

func TestMemoryDeviceRepository_GetDevice_NotFound(t *testing.T) {
	repo := NewMemoryDeviceRepository()
	ctx := context.Background()

	device, err := repo.GetDevice(ctx, "nonexistent")
	if err == nil {
		t.Error("GetDevice() with nonexistent ID should return error")
	}
	if device != nil {
		t.Error("GetDevice() with nonexistent ID should return nil device")
	}
	if err.Error() != "device not found" {
		t.Errorf("GetDevice() error message = %v, want 'device not found'", err.Error())
	}
}

func TestMemoryDeviceRepository_UpdateDevice(t *testing.T) {
	repo := NewMemoryDeviceRepository()
	ctx := context.Background()

	// Create and save initial device
	device := entities.NewDevice("device123", "sensor")
	device.Behavior.MessageCount = 5

	err := repo.SaveDevice(ctx, device)
	if err != nil {
		t.Fatalf("SaveDevice() error = %v", err)
	}

	// Update device
	device.Behavior.MessageCount = 10
	err = repo.UpdateDevice(ctx, device)
	if err != nil {
		t.Errorf("UpdateDevice() error = %v, want nil", err)
	}

	// Retrieve and verify update
	updatedDevice, err := repo.GetDevice(ctx, "device123")
	if err != nil {
		t.Fatalf("GetDevice() error = %v", err)
	}
	if updatedDevice.Behavior.MessageCount != 10 {
		t.Errorf("UpdateDevice() MessageCount = %v, want 10", updatedDevice.Behavior.MessageCount)
	}
}

func TestMemoryDeviceRepository_QuarantineDevice(t *testing.T) {
	repo := NewMemoryDeviceRepository()
	ctx := context.Background()
	deviceID := "device123"

	// Initially device should not be quarantined
	quarantined, err := repo.IsDeviceQuarantined(ctx, deviceID)
	if err != nil {
		t.Errorf("IsDeviceQuarantined() error = %v, want nil", err)
	}
	if quarantined {
		t.Error("IsDeviceQuarantined() = true, want false for new device")
	}

	// Quarantine device
	err = repo.QuarantineDevice(ctx, deviceID, "test reason")
	if err != nil {
		t.Errorf("QuarantineDevice() error = %v, want nil", err)
	}

	// Device should now be quarantined
	quarantined, err = repo.IsDeviceQuarantined(ctx, deviceID)
	if err != nil {
		t.Errorf("IsDeviceQuarantined() error = %v, want nil", err)
	}
	if !quarantined {
		t.Error("IsDeviceQuarantined() = false, want true after quarantine")
	}
}

func TestMemoryDeviceRepository_ReleaseFromQuarantine(t *testing.T) {
	repo := NewMemoryDeviceRepository()
	ctx := context.Background()
	deviceID := "device123"

	// Quarantine device
	err := repo.QuarantineDevice(ctx, deviceID, "test reason")
	if err != nil {
		t.Fatalf("QuarantineDevice() error = %v", err)
	}

	// Verify quarantined
	quarantined, err := repo.IsDeviceQuarantined(ctx, deviceID)
	if err != nil {
		t.Fatalf("IsDeviceQuarantined() error = %v", err)
	}
	if !quarantined {
		t.Fatal("Device should be quarantined")
	}

	// Release from quarantine
	err = repo.ReleaseFromQuarantine(ctx, deviceID)
	if err != nil {
		t.Errorf("ReleaseFromQuarantine() error = %v, want nil", err)
	}

	// Verify not quarantined
	quarantined, err = repo.IsDeviceQuarantined(ctx, deviceID)
	if err != nil {
		t.Errorf("IsDeviceQuarantined() error = %v, want nil", err)
	}
	if quarantined {
		t.Error("IsDeviceQuarantined() = true, want false after release")
	}
}

func TestMemoryDeviceRepository_GetQuarantinedDevices(t *testing.T) {
	repo := NewMemoryDeviceRepository()
	ctx := context.Background()

	// Create and save devices
	device1 := entities.NewDevice("device1", "sensor")
	device2 := entities.NewDevice("device2", "camera")
	device3 := entities.NewDevice("device3", "sensor")

	repo.SaveDevice(ctx, device1)
	repo.SaveDevice(ctx, device2)
	repo.SaveDevice(ctx, device3)

	// Quarantine some devices
	repo.QuarantineDevice(ctx, "device1", "reason1")
	repo.QuarantineDevice(ctx, "device3", "reason3")

	// Get quarantined devices
	quarantined, err := repo.GetQuarantinedDevices(ctx)
	if err != nil {
		t.Errorf("GetQuarantinedDevices() error = %v, want nil", err)
	}

	// Should have 2 quarantined devices
	if len(quarantined) != 2 {
		t.Errorf("GetQuarantinedDevices() count = %v, want 2", len(quarantined))
	}

	// Verify correct devices are returned
	quarantinedIDs := make(map[string]bool)
	for _, device := range quarantined {
		quarantinedIDs[device.ID] = true
	}

	if !quarantinedIDs["device1"] {
		t.Error("GetQuarantinedDevices() missing device1")
	}
	if !quarantinedIDs["device3"] {
		t.Error("GetQuarantinedDevices() missing device3")
	}
	if quarantinedIDs["device2"] {
		t.Error("GetQuarantinedDevices() incorrectly includes device2")
	}
}

func TestMemoryDeviceRepository_QuarantineExpiration(t *testing.T) {
	repo := NewMemoryDeviceRepository().(*MemoryDeviceRepository)
	ctx := context.Background()
	deviceID := "device123"

	// Manually set quarantine time to past to simulate expiration
	pastTime := time.Now().Add(-6 * time.Minute) // More than 5 minute threshold
	repo.quarantinedDevices[deviceID] = pastTime

	// Check if device is quarantined (should be false due to expiration)
	quarantined, err := repo.IsDeviceQuarantined(ctx, deviceID)
	if err != nil {
		t.Errorf("IsDeviceQuarantined() error = %v, want nil", err)
	}
	if quarantined {
		t.Error("IsDeviceQuarantined() = true, want false for expired quarantine")
	}

	// Device should be removed from quarantine map
	if _, exists := repo.quarantinedDevices[deviceID]; exists {
		t.Error("Expired quarantine should be removed from map")
	}
}

func TestMemoryDeviceRepository_CleanExpiredQuarantines(t *testing.T) {
	repo := NewMemoryDeviceRepository().(*MemoryDeviceRepository)
	ctx := context.Background()

	// Add devices with different quarantine times
	now := time.Now()
	repo.quarantinedDevices["device1"] = now.Add(-10 * time.Minute) // Expired
	repo.quarantinedDevices["device2"] = now.Add(-2 * time.Minute)  // Not expired
	repo.quarantinedDevices["device3"] = now.Add(-8 * time.Minute)  // Expired

	// Clean with 5 minute duration
	err := repo.CleanExpiredQuarantines(ctx, 5*time.Minute)
	if err != nil {
		t.Errorf("CleanExpiredQuarantines() error = %v, want nil", err)
	}

	// Check results
	if _, exists := repo.quarantinedDevices["device1"]; exists {
		t.Error("device1 should be removed (expired)")
	}
	if _, exists := repo.quarantinedDevices["device2"]; !exists {
		t.Error("device2 should remain (not expired)")
	}
	if _, exists := repo.quarantinedDevices["device3"]; exists {
		t.Error("device3 should be removed (expired)")
	}
}

func TestMemoryDeviceRepository_ConcurrentAccess(t *testing.T) {
	repo := NewMemoryDeviceRepository()
	ctx := context.Background()

	device := entities.NewDevice("device123", "sensor")

	// Test concurrent reads and writes
	done := make(chan bool, 2)

	// Goroutine 1: Write
	go func() {
		for i := 0; i < 100; i++ {
			device.Behavior.MessageCount = i
			repo.UpdateDevice(ctx, device)
		}
		done <- true
	}()

	// Goroutine 2: Read
	go func() {
		for i := 0; i < 100; i++ {
			repo.GetDevice(ctx, "device123")
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// If we get here without deadlock, the test passes
	t.Log("Concurrent access test completed successfully")
}

func TestMemoryDeviceRepository_QuarantineConcurrentAccess(t *testing.T) {
	repo := NewMemoryDeviceRepository()
	ctx := context.Background()
	deviceID := "device123"

	done := make(chan bool, 3)

	// Goroutine 1: Quarantine/Release
	go func() {
		for i := 0; i < 50; i++ {
			repo.QuarantineDevice(ctx, deviceID, "test")
			repo.ReleaseFromQuarantine(ctx, deviceID)
		}
		done <- true
	}()

	// Goroutine 2: Check quarantine status
	go func() {
		for i := 0; i < 50; i++ {
			repo.IsDeviceQuarantined(ctx, deviceID)
		}
		done <- true
	}()

	// Goroutine 3: Get quarantined devices
	go func() {
		for i := 0; i < 50; i++ {
			repo.GetQuarantinedDevices(ctx)
		}
		done <- true
	}()

	// Wait for all goroutines
	<-done
	<-done
	<-done

	t.Log("Quarantine concurrent access test completed successfully")
}