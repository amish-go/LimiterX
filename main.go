ackage main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"api-rate-limiter/limiter"
	"api-rate-limiter/utils"
)

type Config struct {
	Port          string
	Algorithm     string
	RequestsLimit int
	WindowSeconds int
}

func loadConfig() Config {
	config := Config{
		Port:          getEnv("PORT", "8080"),
		Algorithm:     getEnv("ALGORITHM", "sliding"),
		RequestsLimit: getEnvInt("REQUESTS_LIMIT", 100),
		WindowSeconds: getEnvInt("WINDOW_SECONDS", 60),
	}
	return config
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func getClientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		return xff
	}
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}
	return r.RemoteAddr
}

func main() {
	config := loadConfig()
	cache := utils.NewCache(time.Minute * 10)
	
	var rateLimiter limiter.RateLimiter
	windowDuration := time.Duration(config.WindowSeconds) * time.Second
	
	if config.Algorithm == "fixed" {
		rateLimiter = limiter.NewFixedWindowLimiter(cache, config.RequestsLimit, windowDuration)
	} else {
		rateLimiter = limiter.NewSlidingWindowLimiter(cache, config.RequestsLimit, windowDuration)
	}

	http.HandleFunc("/api/resource", func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)
		
		allowed, remaining, resetTime := rateLimiter.Allow(clientIP)
		
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(config.RequestsLimit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))
		
		if !allowed {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "rate limit exceeded",
				"reset":   resetTime.Unix(),
			})
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "request successful",
			"ip":      clientIP,
		})
	})

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		stats := cache.GetStats()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total_keys":    stats.TotalKeys,
			"total_requests": stats.TotalRequests,
			"algorithm":     config.Algorithm,
			"limit":         config.RequestsLimit,
			"window":        config.WindowSeconds,
		})
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("Starting rate limiter server on port %s", config.Port)
	log.Printf("Algorithm: %s, Limit: %d requests per %d seconds", 
		config.Algorithm, config.RequestsLimit, config.WindowSeconds)
	
	if err := http.ListenAndServe(":"+config.Port, nil); err != nil {
		log.Fatal(err)
	}
}