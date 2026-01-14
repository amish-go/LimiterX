1package limiter

import (
	"sync"
	"time"

	"api-rate-limiter/utils"
)

type RateLimiter interface {
	Allow(key string) (allowed bool, remaining int, resetTime time.Time)
}

type FixedWindowLimiter struct {
	cache    *utils.Cache
	limit    int
	window   time.Duration
	mu       sync.RWMutex
}

type FixedWindowData struct {
	Count     int
	WindowStart time.Time
}

func NewFixedWindowLimiter(cache *utils.Cache, limit int, window time.Duration) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		cache:  cache,
		limit:  limit,
		window: window,
	}
}

func (f *FixedWindowLimiter) Allow(key string) (bool, int, time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()

	now := time.Now()
	val, exists := f.cache.Get(key)

	var data FixedWindowData
	if exists {
		data = val.(FixedWindowData)
		if now.Sub(data.WindowStart) >= f.window {
			data = FixedWindowData{
				Count:     0,
				WindowStart: now,
			}
		}
	} else {
		data = FixedWindowData{
			Count:     0,
			WindowStart: now,
		}
	}

	resetTime := data.WindowStart.Add(f.window)

	if data.Count >= f.limit {
		return false, 0, resetTime
	}

	data.Count++
	f.cache.Set(key, data)

	remaining := f.limit - data.Count
	return true, remaining, resetTime
}

type SlidingWindowLimiter struct {
	cache    *utils.Cache
	limit    int
	window   time.Duration
	mu       sync.RWMutex
}

type SlidingWindowData struct {
	Timestamps []time.Time
}

func NewSlidingWindowLimiter(cache *utils.Cache, limit int, window time.Duration) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		cache:  cache,
		limit:  limit,
		window: window,
	}
}

func (s *SlidingWindowLimiter) Allow(key string) (bool, int, time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-s.window)

	val, exists := s.cache.Get(key)
	var data SlidingWindowData

	if exists {
		data = val.(SlidingWindowData)
		validTimestamps := make([]time.Time, 0)
		for _, ts := range data.Timestamps {
			if ts.After(cutoff) {
				validTimestamps = append(validTimestamps, ts)
			}
		}
		data.Timestamps = validTimestamps
	} else {
		data = SlidingWindowData{
			Timestamps: make([]time.Time, 0),
		}
	}

	if len(data.Timestamps) >= s.limit {
		oldestTimestamp := data.Timestamps[0]
		resetTime := oldestTimestamp.Add(s.window)
		return false, 0, resetTime
	}

	data.Timestamps = append(data.Timestamps, now)
	s.cache.Set(key, data)

	remaining := s.limit - len(data.Timestamps)
	
	var resetTime time.Time
	if len(data.Timestamps) > 0 {
		resetTime = data.Timestamps[0].Add(s.window)
	} else {
		resetTime = now.Add(s.window)
	}

	return true, remaining, resetTime
}