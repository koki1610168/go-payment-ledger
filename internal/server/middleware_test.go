package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func dummyHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}

func TestRateLimiter_AllowsRequestsWithinLimit(t *testing.T) {
	// rate=10 rps, burst=5 means 5 immediate requests allowed
	limiter := NewRateLimiter(10, 5)
	handler := limiter.Middleware(dummyHandler())

	// First 5 requests should all succeed (within burst)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		res := httptest.NewRecorder()

		handler.ServeHTTP(res, req)

		if res.Code != http.StatusOK {
			t.Errorf("request %d: got status %d, want %d", i+1, res.Code, http.StatusOK)
		}
	}
}

func TestRateLimiter_BlocksExcessRequests(t *testing.T) {
	// Very restrictive: 1 request per second, burst of 2
	limiter := NewRateLimiter(1, 2)
	handler := limiter.Middleware(dummyHandler())

	ip := "192.168.1.1:12345"

	// First 2 requests should succeed (burst)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = ip
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)

		if res.Code != http.StatusOK {
			t.Errorf("request %d: expected OK, got %d", i+1, res.Code)
		}
	}

	// Third request should be rate limited
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = ip
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 TooManyRequests, got %d", res.Code)
	}

	// Check Retry-After header is set
	if res.Header().Get("Retry-After") == "" {
		t.Error("expected Retry-After header to be set")
	}
}

func TestRateLimiter_SeparateLimitsPerIP(t *testing.T) {
	// Burst of 1 so we can easily test isolation
	limiter := NewRateLimiter(1, 1)
	handler := limiter.Middleware(dummyHandler())

	ip1 := "192.168.1.1:12345"
	ip2 := "192.168.1.2:12345"

	// IP1's first request succeeds
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1.RemoteAddr = ip1
	res1 := httptest.NewRecorder()
	handler.ServeHTTP(res1, req1)

	if res1.Code != http.StatusOK {
		t.Errorf("IP1 first request: got %d, want %d", res1.Code, http.StatusOK)
	}

	// IP1's second request should be blocked
	req1b := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1b.RemoteAddr = ip1
	res1b := httptest.NewRecorder()
	handler.ServeHTTP(res1b, req1b)

	if res1b.Code != http.StatusTooManyRequests {
		t.Errorf("IP1 second request: got %d, want %d", res1b.Code, http.StatusTooManyRequests)
	}

	// IP2's first request should still succeed (separate limiter)
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.RemoteAddr = ip2
	res2 := httptest.NewRecorder()
	handler.ServeHTTP(res2, req2)

	if res2.Code != http.StatusOK {
		t.Errorf("IP2 first request: got %d, want %d", res2.Code, http.StatusOK)
	}
}

func TestRateLimiter_GetLimiterCreatesSameLimiterForSameIP(t *testing.T) {
	limiter := NewRateLimiter(10, 5)

	ip := "10.0.0.1:8080"

	l1 := limiter.getLimiter(ip)
	l2 := limiter.getLimiter(ip)

	if l1 != l2 {
		t.Error("expected same limiter instance for same IP")
	}
}

func TestRateLimiter_GetLimiterCreatesDifferentLimitersForDifferentIPs(t *testing.T) {
	limiter := NewRateLimiter(10, 5)

	l1 := limiter.getLimiter("10.0.0.1:8080")
	l2 := limiter.getLimiter("10.0.0.2:8080")

	if l1 == l2 {
		t.Error("expected different limiter instances for different IPs")
	}
}