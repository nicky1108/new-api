package controller

import (
	"context"
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
	capturePlaygroundImageRelayFunc = func(_ *gin.Context, targetPath string, bodyBytes []byte, _ context.Context) playgroundImageStreamResult {
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

func TestPlaygroundImageStreamDefersLargeResultBodyToJob(t *testing.T) {
	gin.SetMode(gin.TestMode)

	origInterval := playgroundImageStreamPingInterval
	origCapture := capturePlaygroundImageRelayFunc
	playgroundImageStreamPingInterval = time.Millisecond
	largeImageBody := `{"data":[{"b64_json":"` + strings.Repeat("a", 4096) + `"}]}`
	capturePlaygroundImageRelayFunc = func(_ *gin.Context, targetPath string, bodyBytes []byte, _ context.Context) playgroundImageStreamResult {
		if targetPath != "/pg/images/edits" {
			t.Fatalf("unexpected target path: %s", targetPath)
		}
		if !strings.Contains(string(bodyBytes), `name="image"`) {
			t.Fatalf("request body was not captured as multipart: %s", string(bodyBytes))
		}
		time.Sleep(5 * time.Millisecond)
		return playgroundImageStreamResult{
			Status: http.StatusOK,
			Body:   []byte(largeImageBody),
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
		"/pg/images/edits/stream",
		strings.NewReader("--test-boundary\r\nContent-Disposition: form-data; name=\"model\"\r\n\r\ngpt-image-2\r\n--test-boundary\r\nContent-Disposition: form-data; name=\"prompt\"\r\n\r\nedit this\r\n--test-boundary\r\nContent-Disposition: form-data; name=\"image\"\r\n\r\ndata:image/png;base64,aaaa\r\n--test-boundary--\r\n"),
	)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=test-boundary")
	ctx.Request = req

	PlaygroundImageStream(ctx)

	body := recorder.Body.String()
	if !strings.Contains(body, "event: job") {
		t.Fatalf("expected job event before deferred result, got: %s", body)
	}
	if !strings.Contains(body, "event: result") {
		t.Fatalf("expected final result event, got: %s", body)
	}
	if !strings.Contains(body, `"deferred":true`) {
		t.Fatalf("expected result body to be deferred, got: %s", body)
	}
	if strings.Contains(body, largeImageBody) {
		t.Fatalf("stream result should not inline large image payload")
	}
}
