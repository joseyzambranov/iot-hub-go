package services

import (
	"sync"
	"time"
)

type RateLimiter struct {
	requests    map[string][]time.Time
	maxRequests int
	window      time.Duration
	mutex       sync.RWMutex
}

func NewRateLimiter(maxRequests int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests:    make(map[string][]time.Time),
		maxRequests: maxRequests,
		window:      window,
		mutex:       sync.RWMutex{},
	}
}

func (rl *RateLimiter) IsAllowed(deviceID string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	
	// Obtener o crear la lista de requests para este dispositivo
	requests, exists := rl.requests[deviceID]
	if !exists {
		requests = make([]time.Time, 0)
	}

	// Filtrar requests antiguos fuera de la ventana de tiempo
	validRequests := make([]time.Time, 0)
	cutoff := now.Add(-rl.window)
	
	for _, requestTime := range requests {
		if requestTime.After(cutoff) {
			validRequests = append(validRequests, requestTime)
		}
	}

	// Verificar si podemos agregar una nueva request
	if len(validRequests) >= rl.maxRequests {
		rl.requests[deviceID] = validRequests
		return false
	}

	// Agregar la nueva request
	validRequests = append(validRequests, now)
	rl.requests[deviceID] = validRequests
	
	return true
}

func (rl *RateLimiter) GetRequestCount(deviceID string) int {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	requests, exists := rl.requests[deviceID]
	if !exists {
		return 0
	}

	now := time.Now()
	cutoff := now.Add(-rl.window)
	count := 0

	for _, requestTime := range requests {
		if requestTime.After(cutoff) {
			count++
		}
	}

	return count
}

func (rl *RateLimiter) Reset(deviceID string) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	delete(rl.requests, deviceID)
}

func (rl *RateLimiter) CleanupOldRequests() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	for deviceID, requests := range rl.requests {
		validRequests := make([]time.Time, 0)
		
		for _, requestTime := range requests {
			if requestTime.After(cutoff) {
				validRequests = append(validRequests, requestTime)
			}
		}

		if len(validRequests) == 0 {
			delete(rl.requests, deviceID)
		} else {
			rl.requests[deviceID] = validRequests
		}
	}
}