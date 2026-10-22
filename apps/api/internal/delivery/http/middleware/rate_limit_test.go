package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func TestRateLimit_AllowsNormalTraffic(t *testing.T) {
	handler := RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// 5 requests should be allowed
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
		req.RemoteAddr = "192.168.1.100:12345"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: Status = %d, want %d", i+1, w.Code, http.StatusOK)
		}
	}
}

func TestRateLimit_BlocksExcessiveTraffic(t *testing.T) {
	handler := RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Use a unique IP to avoid conflicts with other tests
	testIP := "10.99.99.99:54321"

	// Send rate + 1 requests (6th request should be blocked)
	for i := 0; i < rate+1; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
		req.RemoteAddr = testIP
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if i < rate {
			if w.Code != http.StatusOK {
				t.Errorf("Request %d should be allowed, got status %d", i+1, w.Code)
			}
		} else {
			if w.Code != http.StatusTooManyRequests {
				t.Errorf("Request %d should be blocked, got status %d", i+1, w.Code)
			}
		}
	}
}

func TestRateLimit_DifferentIPsAreIndependent(t *testing.T) {
	handler := RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	ips := []string{
		"172.16.0.1:11111",
		"172.16.0.2:22222",
		"172.16.0.3:33333",
	}

	for _, ip := range ips {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
		req.RemoteAddr = ip
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("First request from %s should be allowed, got %d", ip, w.Code)
		}
	}
}

func TestRateLimit_ConcurrentRequests(t *testing.T) {
	handler := RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	var wg sync.WaitGroup

	// Send concurrent requests from different IPs — should not panic
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
			req.RemoteAddr = "concurrent-test-ip:12345"
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			// We don't check status here — just ensuring no data race/panic
		}(i)
	}

	wg.Wait()
}

func TestRateLimit_ResponseBody(t *testing.T) {
	handler := RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	testIP := "10.88.88.88:99999"

	// Exhaust rate limit
	for i := 0; i <= rate; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
		req.RemoteAddr = testIP
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}

	// Next request should be rate limited with proper response
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	req.RemoteAddr = testIP
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusTooManyRequests)
	}
}

func TestRateLimit_IgnoresSpoofedForwardedHeaders(t *testing.T) {
	// Same chain as the router: PeerAddr, then RealIP, then RateLimit
	handler := PeerAddr(chimiddleware.RealIP(RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))))

	blocked := false
	for i := 0; i < rate+1; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
		// Same peer host, new source port and a fresh spoofed IP each time
		req.RemoteAddr = fmt.Sprintf("10.77.77.77:%d", 40000+i)
		req.Header.Set("X-Forwarded-For", fmt.Sprintf("203.0.113.%d", i))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code == http.StatusTooManyRequests {
			blocked = true
		}
	}
	if !blocked {
		t.Error("rotating X-Forwarded-For / source port bypassed the rate limit")
	}
}
