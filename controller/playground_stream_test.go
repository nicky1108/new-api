package controller

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestPlaygroundImageStreamEmitsKeepaliveBeforeResult(t *testing.T) {
	gin.SetMode(gin.TestMode)

	origInterval := playgroundImageStreamPingInterval
	origCapture := capturePlaygroundImageRelayFunc
	playgroundImageStreamPingInterval = time.Millisecond
	capturePlaygroundImageRelayFunc = func(_ *gin.Context, targetPath string, bodyBytes []byte) playgroundImageStreamResult {
		if targetPath != "/pg/images/generations" {
			t.Fatalf("unexpected target path: %s", targetPath)
		}
		if !strings.Contains(string(bodyBytes), `"model"`) {
			t.Fatalf("request body was not captured: %s", string(bodyBytes))
		}
		time.Sleep(5 * time.Millisecond)
		return playgroundImageStreamResult{
			Status: http.StatusOK,
			Body:   []byte(`{"data":[{"url":"https://example.test/image.png"}]}`),
		}
	}
	t.Cleanup(func() {
		playgroundImageStreamPingInterval = origInterval
		capturePlaygroundImageRelayFunc = origCapture
	})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(
		http.MethodPost,
		"/pg/images/generations/stream",
		strings.NewReader(`{"model":"gpt-image-2","prompt":"test"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	ctx.Request = req

	PlaygroundImageStream(ctx)

	body := recorder.Body.String()
	if !strings.Contains(body, ": started\n\n") {
		t.Fatalf("expected initial stream comment, got: %s", body)
	}
	if !strings.Contains(body, ": ping\n\n") {
		t.Fatalf("expected keepalive ping before final result, got: %s", body)
	}
	if !strings.Contains(body, "event: result") {
		t.Fatalf("expected final result event, got: %s", body)
	}
	if !strings.Contains(body, `"status":200`) {
		t.Fatalf("expected result status, got: %s", body)
	}
}
