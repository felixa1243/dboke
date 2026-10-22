package middleware

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"

	"dboke-api/internal/pkg/contextkeys"
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

// PeerAddr records the TCP peer address before middleware.RealIP rewrites
// r.RemoteAddr from client-controlled X-Forwarded-For / X-Real-IP headers.
// It must be registered ahead of RealIP.
func PeerAddr(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), contextkeys.PeerAddrKey, r.RemoteAddr)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// clientIP returns the rate-limit key for r: the peer's host without the port,
// so neither spoofed proxy headers nor new source ports start a fresh bucket.
func clientIP(r *http.Request) string {
	addr, ok := r.Context().Value(contextkeys.PeerAddrKey).(string)
	if !ok || addr == "" {
		addr = r.RemoteAddr
	}
	if host, _, err := net.SplitHostPort(addr); err == nil {
		return host
	}
	return addr
}

// RateLimit is a simple IP-based rate limiter middleware
func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
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
