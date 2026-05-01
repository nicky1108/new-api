package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestIsLocalModelRateLimitRequestAllowsLoopbackClientIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/v1/chat/completions", nil)
	ctx.Request.RemoteAddr = "127.0.0.1:12345"

	if !isLocalModelRateLimitRequest(ctx) {
		t.Fatal("expected loopback request to bypass model rate limit")
	}
}

func TestIsLocalModelRateLimitRequestRejectsForwardedExternalIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/v1/chat/completions", nil)
	ctx.Request.RemoteAddr = "127.0.0.1:12345"
	ctx.Request.Header.Set("X-Forwarded-For", "203.0.113.10")

	if isLocalModelRateLimitRequest(ctx) {
		t.Fatal("expected forwarded external request to keep model rate limit")
	}
}

func TestIsLocalModelRateLimitRequestAllowsForwardedLoopbackIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/v1/chat/completions", nil)
	ctx.Request.RemoteAddr = "127.0.0.1:12345"
	ctx.Request.Header.Set("X-Forwarded-For", "127.0.0.1, ::1")

	if !isLocalModelRateLimitRequest(ctx) {
		t.Fatal("expected forwarded loopback request to bypass model rate limit")
	}
}
