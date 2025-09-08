package services

import (
	"testing"
	"time"
)

func TestRateLimiter_IsAllowed(t *testing.T) {
	tests := []struct {
		name           string
		maxRequests    int
		window         time.Duration
		deviceID       string
		requestCount   int
		expectedResult bool
	}{
		{
			name:           "First request should be allowed",
			maxRequests:    5,
			window:         1 * time.Minute,
			deviceID:       "device-001",
			requestCount:   1,
			expectedResult: true,
		},
		{
			name:           "Requests within limit should be allowed",
			maxRequests:    5,
			window:         1 * time.Minute,
			deviceID:       "device-002",
			requestCount:   5,
			expectedResult: true,
		},
		{
			name:           "Requests exceeding limit should be denied",
			maxRequests:    5,
			window:         1 * time.Minute,
			deviceID:       "device-003",
			requestCount:   6,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rl := NewRateLimiter(tt.maxRequests, tt.window)
			
			var lastResult bool
			// Make the specified number of requests
			for i := 0; i < tt.requestCount; i++ {
				lastResult = rl.IsAllowed(tt.deviceID)
			}

			if lastResult != tt.expectedResult {
				t.Errorf("Expected %v, but got %v for device %s", tt.expectedResult, lastResult, tt.deviceID)
			}
		})
	}
}

func TestRateLimiter_GetRequestCount(t *testing.T) {
	rl := NewRateLimiter(10, 1*time.Minute)
	deviceID := "test-device"

	// Make 3 requests
	for i := 0; i < 3; i++ {
		rl.IsAllowed(deviceID)
	}

	count := rl.GetRequestCount(deviceID)
	if count != 3 {
		t.Errorf("Expected request count of 3, but got %d", count)
	}
}

func TestRateLimiter_Reset(t *testing.T) {
	rl := NewRateLimiter(5, 1*time.Minute)
	deviceID := "test-device"

	// Make some requests
	for i := 0; i < 3; i++ {
		rl.IsAllowed(deviceID)
	}

	// Verify requests are recorded
	if count := rl.GetRequestCount(deviceID); count != 3 {
		t.Errorf("Expected request count of 3 before reset, but got %d", count)
	}

	// Reset the device
	rl.Reset(deviceID)

	// Verify requests are cleared
	if count := rl.GetRequestCount(deviceID); count != 0 {
		t.Errorf("Expected request count of 0 after reset, but got %d", count)
	}
}

func TestRateLimiter_WindowExpiry(t *testing.T) {
	// Use a very short window for testing
	shortWindow := 50 * time.Millisecond
	rl := NewRateLimiter(2, shortWindow)
	deviceID := "test-device"

	// Make 2 requests (at limit)
	if !rl.IsAllowed(deviceID) {
		t.Error("First request should be allowed")
	}
	if !rl.IsAllowed(deviceID) {
		t.Error("Second request should be allowed")
	}

	// Third request should be denied
	if rl.IsAllowed(deviceID) {
		t.Error("Third request should be denied due to rate limit")
	}

	// Wait for window to expire
	time.Sleep(shortWindow + 10*time.Millisecond)

	// Request should now be allowed again
	if !rl.IsAllowed(deviceID) {
		t.Error("Request should be allowed after window expiry")
	}
}

func TestRateLimiter_MultipleDevices(t *testing.T) {
	rl := NewRateLimiter(2, 1*time.Minute)

	// Device 1 makes 2 requests (at limit)
	device1 := "device-001"
	if !rl.IsAllowed(device1) {
		t.Error("Device 1 first request should be allowed")
	}
	if !rl.IsAllowed(device1) {
		t.Error("Device 1 second request should be allowed")
	}

	// Device 2 should still be able to make requests
	device2 := "device-002"
	if !rl.IsAllowed(device2) {
		t.Error("Device 2 first request should be allowed")
	}
	if !rl.IsAllowed(device2) {
		t.Error("Device 2 second request should be allowed")
	}

	// Both devices should now be at their limit
	if rl.IsAllowed(device1) {
		t.Error("Device 1 third request should be denied")
	}
	if rl.IsAllowed(device2) {
		t.Error("Device 2 third request should be denied")
	}
}

func TestRateLimiter_CleanupOldRequests(t *testing.T) {
	shortWindow := 50 * time.Millisecond
	rl := NewRateLimiter(5, shortWindow)
	deviceID := "test-device"

	// Make some requests
	for i := 0; i < 3; i++ {
		rl.IsAllowed(deviceID)
	}

	// Verify requests are recorded
	if count := rl.GetRequestCount(deviceID); count != 3 {
		t.Errorf("Expected request count of 3, but got %d", count)
	}

	// Wait for window to expire
	time.Sleep(shortWindow + 10*time.Millisecond)

	// Clean up old requests
	rl.CleanupOldRequests()

	// Verify old requests are cleaned up
	if count := rl.GetRequestCount(deviceID); count != 0 {
		t.Errorf("Expected request count of 0 after cleanup, but got %d", count)
	}
}