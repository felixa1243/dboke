package middleware

import (
	"net/http"
	"sync"
	"time"

	"dboke-api/internal/pkg/response"
)

type clientVisitor struct {
	lastSeen time.Time
	count    int
}

var (
	visitors = make(map[string]*clientVisitor)
	mu       sync.Mutex
	rate     = 5 // 5 requests per 10 seconds for login
)

func init() {
	go cleanupVisitors()
}

func cleanupVisitors() {
	for {
		time.Sleep(1 * time.Minute)
		mu.Lock()
		for ip, v := range visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(visitors, ip)
			}
		}
		mu.Unlock()
	}
}

// RateLimit is a simple IP-based rate limiter middleware
func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		mu.Lock()
		v, exists := visitors[ip]
		if !exists {
			visitors[ip] = &clientVisitor{lastSeen: time.Now(), count: 1}
			mu.Unlock()
			next.ServeHTTP(w, r)
			return
		}

		if time.Since(v.lastSeen) > 10*time.Second {
			v.count = 0
		}
		v.lastSeen = time.Now()
		v.count++

		if v.count > rate {
			mu.Unlock()
			response.WriteError(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Too many requests. Please try again later.")
			return
		}
		mu.Unlock()

		next.ServeHTTP(w, r)
	})
}
