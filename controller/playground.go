package controller

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
)

func Playground(c *gin.Context) {
	playgroundRelay(c, types.RelayFormatOpenAI)
}

func PlaygroundImage(c *gin.Context) {
	playgroundRelay(c, types.RelayFormatOpenAIImage)
}

func PlaygroundAudio(c *gin.Context) {
	playgroundRelay(c, types.RelayFormatOpenAIAudio)
}

func PlaygroundMusic(c *gin.Context) {
	playgroundRelay(c, types.RelayFormatMiniMaxMusic)
}

type playgroundImageStreamResult struct {
	Status   int             `json:"status"`
	Body     json.RawMessage `json:"body,omitempty"`
	Text     string          `json:"text,omitempty"`
	JobID    string          `json:"job_id,omitempty"`
	Deferred bool            `json:"deferred,omitempty"`
}

var (
	playgroundImageStreamPingInterval = 10 * time.Second
	playgroundImageJobTTL             = 15 * time.Minute
	capturePlaygroundImageRelayFunc   = capturePlaygroundImageRelay
	playgroundImageJobs               = struct {
		sync.Mutex
		items map[string]*playgroundImageJob
	}{items: make(map[string]*playgroundImageJob)}
)

type playgroundImageJob struct {
	UserID    int
	Result    playgroundImageStreamResult
	Done      bool
	ExpiresAt time.Time
}

func PlaygroundImageStream(c *gin.Context) {
	targetPath, ok := playgroundImageStreamTargetPath(c.Request.URL.Path)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"message": "unsupported playground image stream path",
				"type":    "invalid_request_error",
			},
		})
		return
	}

	bodyStorage, err := common.GetBodyStorage(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": err.Error(),
				"type":    "invalid_request_error",
			},
		})
		return
	}
	bodyBytes, err := bodyStorage.Bytes()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": err.Error(),
				"type":    "invalid_request_error",
			},
		})
		return
	}

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "streaming is not supported",
				"type":    "server_error",
			},
		})
		return
	}

	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)
	_, _ = c.Writer.Write([]byte(": started\n\n"))
	flusher.Flush()

	jobID := createPlaygroundImageJob(c.GetInt("id"))
	writePlaygroundImageStreamEvent(c, "job", gin.H{"id": jobID})
	flusher.Flush()

	sourceContext := c.Copy()
	resultCh := make(chan playgroundImageStreamResult, 1)
	go func() {
		jobContext, cancel := context.WithTimeout(context.Background(), playgroundImageJobTTL)
		defer cancel()
		result := capturePlaygroundImageRelayFunc(sourceContext, targetPath, bodyBytes, jobContext)
		finishPlaygroundImageJob(jobID, result)
		resultCh <- result
	}()

	ticker := time.NewTicker(playgroundImageStreamPingInterval)
	defer ticker.Stop()

	for {
		select {
		case result := <-resultCh:
			writePlaygroundImageStreamEvent(c, "result", playgroundImageStreamResult{
				Status:   result.Status,
				JobID:    jobID,
				Deferred: true,
			})
			flusher.Flush()
			return
		case <-ticker.C:
			_, _ = c.Writer.Write([]byte(": ping\n\n"))
			flusher.Flush()
		case <-c.Request.Context().Done():
			return
		}
	}
}

func PlaygroundImageJob(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "image job id is required",
				"type":    "invalid_request_error",
			},
		})
		return
	}

	userID := c.GetInt("id")
	result, done, ok := getPlaygroundImageJob(jobID, userID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"message": "image job not found",
				"type":    "not_found_error",
			},
		})
		return
	}
	if !done {
		c.JSON(http.StatusAccepted, gin.H{"status": "pending"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func playgroundImageStreamTargetPath(path string) (string, bool) {
	if !strings.HasSuffix(path, "/stream") {
		return "", false
	}
	targetPath := strings.TrimSuffix(path, "/stream")
	switch targetPath {
	case "/pg/images/generations", "/pg/images/edits":
		return targetPath, true
	default:
		return "", false
	}
}

func createPlaygroundImageJob(userID int) string {
	now := time.Now()
	for {
		jobID := randomPlaygroundImageJobID()
		playgroundImageJobs.Lock()
		cleanupExpiredPlaygroundImageJobsLocked(now)
		if _, exists := playgroundImageJobs.items[jobID]; !exists {
			playgroundImageJobs.items[jobID] = &playgroundImageJob{
				UserID:    userID,
				ExpiresAt: now.Add(playgroundImageJobTTL),
			}
			playgroundImageJobs.Unlock()
			return jobID
		}
		playgroundImageJobs.Unlock()
	}
}

func finishPlaygroundImageJob(jobID string, result playgroundImageStreamResult) {
	playgroundImageJobs.Lock()
	defer playgroundImageJobs.Unlock()

	job, ok := playgroundImageJobs.items[jobID]
	if !ok {
		return
	}
	job.Result = result
	job.Done = true
	job.ExpiresAt = time.Now().Add(playgroundImageJobTTL)
}

func getPlaygroundImageJob(jobID string, userID int) (playgroundImageStreamResult, bool, bool) {
	now := time.Now()
	playgroundImageJobs.Lock()
	defer playgroundImageJobs.Unlock()

	cleanupExpiredPlaygroundImageJobsLocked(now)
	job, ok := playgroundImageJobs.items[jobID]
	if !ok || job.UserID != userID {
		return playgroundImageStreamResult{}, false, false
	}
	if !job.Done {
		return playgroundImageStreamResult{}, false, true
	}

	result := job.Result
	delete(playgroundImageJobs.items, jobID)
	return result, true, true
}

func cleanupExpiredPlaygroundImageJobsLocked(now time.Time) {
	for jobID, job := range playgroundImageJobs.items {
		if now.After(job.ExpiresAt) {
			delete(playgroundImageJobs.items, jobID)
		}
	}
}

func randomPlaygroundImageJobID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err == nil {
		return hex.EncodeToString(b[:])
	}
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func capturePlaygroundImageRelay(original *gin.Context, targetPath string, bodyBytes []byte, requestContext context.Context) (result playgroundImageStreamResult) {
	defer func() {
		if recovered := recover(); recovered != nil {
			body, _ := json.Marshal(gin.H{
				"error": gin.H{
					"message": fmt.Sprintf("playground image relay panic: %v", recovered),
					"type":    "server_error",
				},
			})
			result = playgroundImageStreamResult{
				Status: http.StatusInternalServerError,
				Body:   body,
			}
		}
	}()

	recorder := httptest.NewRecorder()
	captured, _ := gin.CreateTestContext(recorder)
	captured.Keys = cloneGinKeys(original.Keys)
	captured.Params = append(gin.Params(nil), original.Params...)
	captured.Request = clonePlaygroundImageRequest(original.Request, targetPath, bodyBytes, requestContext)

	playgroundRelay(captured, types.RelayFormatOpenAIImage)

	status := recorder.Code
	if status == 0 {
		status = http.StatusOK
	}
	body := bytes.TrimSpace(recorder.Body.Bytes())
	result = playgroundImageStreamResult{Status: status}
	if len(body) == 0 {
		return result
	}
	if json.Valid(body) {
		result.Body = json.RawMessage(append([]byte(nil), body...))
	} else {
		result.Text = string(body)
	}
	return result
}

func cloneGinKeys(keys map[string]any) map[string]any {
	if len(keys) == 0 {
		return nil
	}
	clone := make(map[string]any, len(keys))
	for key, value := range keys {
		clone[key] = value
	}
	delete(clone, common.KeyBodyStorage)
	delete(clone, common.KeyRequestBody)
	return clone
}

func clonePlaygroundImageRequest(original *http.Request, targetPath string, bodyBytes []byte, requestContext context.Context) *http.Request {
	if requestContext == nil {
		requestContext = context.Background()
	}
	req := original.Clone(requestContext)
	req.Header = original.Header.Clone()
	req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	req.ContentLength = int64(len(bodyBytes))

	urlCopy := &url.URL{}
	if original.URL != nil {
		*urlCopy = *original.URL
	}
	urlCopy.Path = targetPath
	urlCopy.RawPath = ""
	req.URL = urlCopy
	req.RequestURI = targetPath
	if urlCopy.RawQuery != "" {
		req.RequestURI += "?" + urlCopy.RawQuery
	}
	return req
}

func writePlaygroundImageStreamEvent(c *gin.Context, event string, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		data, _ = json.Marshal(playgroundImageStreamResult{
			Status: http.StatusInternalServerError,
			Body:   json.RawMessage(`{"error":{"message":"failed to encode stream payload","type":"server_error"}}`),
		})
		event = "result"
	}
	_, _ = fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, data)
}

func playgroundRelay(c *gin.Context, relayFormat types.RelayFormat) {
	var newAPIError *types.NewAPIError

	defer func() {
		if newAPIError != nil {
			c.JSON(newAPIError.StatusCode, gin.H{
				"error": newAPIError.ToOpenAIError(),
			})
		}
	}()

	useAccessToken := c.GetBool("use_access_token")
	if useAccessToken {
		newAPIError = types.NewError(errors.New("暂不支持使用 access token"), types.ErrorCodeAccessDenied, types.ErrOptionWithSkipRetry())
		return
	}

	relayInfo, err := relaycommon.GenRelayInfo(c, relayFormat, nil, nil)
	if err != nil {
		newAPIError = types.NewError(err, types.ErrorCodeInvalidRequest, types.ErrOptionWithSkipRetry())
		return
	}

	userId := c.GetInt("id")

	// Write user context to ensure acceptUnsetRatio is available
	userCache, err := model.GetUserCache(userId)
	if err != nil {
		newAPIError = types.NewError(err, types.ErrorCodeQueryDataError, types.ErrOptionWithSkipRetry())
		return
	}
	userCache.WriteContext(c)

	tempToken := &model.Token{
		UserId: userId,
		Name:   fmt.Sprintf("playground-%s", relayInfo.UsingGroup),
		Group:  relayInfo.UsingGroup,
	}
	_ = middleware.SetupContextForToken(c, tempToken)

	Relay(c, relayFormat)
}
