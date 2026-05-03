package xai

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relay/channel"
	taskcommon "github.com/QuantumNous/new-api/relay/channel/task/taskcommon"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

const ChannelName = "xai"

var ModelList = []string{
	"grok-imagine-video",
}

type TaskAdaptor struct {
	taskcommon.BaseBilling
	apiKey  string
	baseURL string
}

type videoTaskResponse struct {
	ID        string `json:"id"`
	RequestID string `json:"request_id,omitempty"`
	TaskID    string `json:"task_id,omitempty"`
	Object    string `json:"object,omitempty"`
	Model     string `json:"model,omitempty"`
	Status    string `json:"status,omitempty"`
	Progress  int    `json:"progress,omitempty"`
	CreatedAt int64  `json:"created_at,omitempty"`
	URL       string `json:"url,omitempty"`
	VideoURL  string `json:"video_url,omitempty"`
	Video     *struct {
		URL      string `json:"url,omitempty"`
		Duration any    `json:"duration,omitempty"`
	} `json:"video,omitempty"`
	Error *dto.OpenAIVideoError `json:"error,omitempty"`
}

func (a *TaskAdaptor) Init(info *relaycommon.RelayInfo) {
	a.baseURL = strings.TrimRight(info.ChannelBaseUrl, "/")
	a.apiKey = info.ApiKey
}

func (a *TaskAdaptor) ValidateRequestAndSetAction(c *gin.Context, info *relaycommon.RelayInfo) *dto.TaskError {
	return relaycommon.ValidateMultipartDirect(c, info)
}

func (a *TaskAdaptor) BuildRequestURL(info *relaycommon.RelayInfo) (string, error) {
	if info.Action == constant.TaskActionRemix {
		return "", fmt.Errorf("xAI video remix is not supported")
	}
	return fmt.Sprintf("%s/v1/videos/generations", a.baseURL), nil
}

func (a *TaskAdaptor) BuildRequestHeader(c *gin.Context, req *http.Request, info *relaycommon.RelayInfo) error {
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Content-Type", "application/json")
	return nil
}

func (a *TaskAdaptor) BuildRequestBody(c *gin.Context, info *relaycommon.RelayInfo) (io.Reader, error) {
	req, err := relaycommon.GetTaskRequest(c)
	if err != nil {
		return nil, err
	}

	body := map[string]any{
		"model":  info.UpstreamModelName,
		"prompt": req.Prompt,
	}
	var images []string
	if req.Image != "" {
		images = append(images, req.Image)
	}
	if len(req.Images) > 0 {
		images = append(images, req.Images...)
	}
	if len(images) == 1 {
		body["image"] = map[string]string{"url": images[0]}
	} else if len(images) > 1 {
		referenceImages := make([]map[string]string, 0, len(images))
		for _, image := range images {
			referenceImages = append(referenceImages, map[string]string{"url": image})
		}
		body["reference_images"] = referenceImages
	}
	if req.Size != "" {
		body["size"] = req.Size
	}
	if req.Duration > 0 {
		body["duration"] = req.Duration
	}
	if req.Seconds != "" {
		if seconds, err := strconv.Atoi(req.Seconds); err == nil {
			body["duration"] = seconds
		}
	}
	for k, v := range req.Metadata {
		if k == "model" || k == "prompt" {
			continue
		}
		body[k] = v
	}

	bodyBytes, err := common.Marshal(body)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(bodyBytes), nil
}

func (a *TaskAdaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (*http.Response, error) {
	return channel.DoTaskApiRequest(a, c, info, requestBody)
}

func (a *TaskAdaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (taskID string, taskData []byte, taskErr *dto.TaskError) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		taskErr = service.TaskErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError)
		return
	}
	_ = resp.Body.Close()

	var upstreamResp videoTaskResponse
	if err := common.Unmarshal(responseBody, &upstreamResp); err != nil {
		taskErr = service.TaskErrorWrapper(errors.Wrapf(err, "body: %s", responseBody), "unmarshal_response_body_failed", http.StatusInternalServerError)
		return
	}

	upstreamID := firstNonEmpty(upstreamResp.ID, upstreamResp.RequestID, upstreamResp.TaskID)
	if upstreamID == "" {
		taskErr = service.TaskErrorWrapper(fmt.Errorf("task_id is empty"), "invalid_response", http.StatusInternalServerError)
		return
	}

	publicResp := dto.NewOpenAIVideo()
	publicResp.ID = info.PublicTaskID
	publicResp.TaskID = info.PublicTaskID
	publicResp.Model = firstNonEmpty(upstreamResp.Model, info.OriginModelName)
	publicResp.Status = xaiStatusToOpenAIVideoStatus(firstNonEmpty(upstreamResp.Status, "pending"))
	publicResp.Progress = upstreamResp.Progress
	if publicResp.Progress == 0 {
		publicResp.Progress = 10
	}
	publicResp.CreatedAt = upstreamResp.CreatedAt
	if upstreamResp.URL != "" {
		publicResp.SetMetadata("url", upstreamResp.URL)
	}
	if upstreamResp.VideoURL != "" {
		publicResp.SetMetadata("url", upstreamResp.VideoURL)
	}
	if upstreamResp.Video != nil && upstreamResp.Video.URL != "" {
		publicResp.SetMetadata("url", upstreamResp.Video.URL)
	}
	if upstreamResp.Error != nil {
		publicResp.Error = upstreamResp.Error
	}

	c.JSON(http.StatusOK, publicResp)
	return upstreamID, responseBody, nil
}

func (a *TaskAdaptor) FetchTask(baseUrl, key string, body map[string]any, proxy string) (*http.Response, error) {
	taskID, ok := body["task_id"].(string)
	if !ok || strings.TrimSpace(taskID) == "" {
		return nil, fmt.Errorf("invalid task_id")
	}

	uri := fmt.Sprintf("%s/v1/videos/%s", strings.TrimRight(baseUrl, "/"), taskID)
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+key)

	client, err := service.GetHttpClientWithProxy(proxy)
	if err != nil {
		return nil, fmt.Errorf("new proxy http client failed: %w", err)
	}
	return client.Do(req)
}

func (a *TaskAdaptor) ParseTaskResult(respBody []byte) (*relaycommon.TaskInfo, error) {
	var resp videoTaskResponse
	if err := common.Unmarshal(respBody, &resp); err != nil {
		return nil, errors.Wrap(err, "unmarshal task result failed")
	}

	taskInfo := &relaycommon.TaskInfo{
		Code:     0,
		Status:   xaiStatusToTaskStatus(resp.Status),
		Progress: xaiProgress(resp.Status, resp.Progress),
	}
	if resp.URL != "" {
		taskInfo.Url = resp.URL
	} else if resp.VideoURL != "" {
		taskInfo.Url = resp.VideoURL
	} else if resp.Video != nil && resp.Video.URL != "" {
		taskInfo.Url = resp.Video.URL
	}
	if resp.Error != nil {
		taskInfo.Reason = resp.Error.Message
	} else if strings.EqualFold(resp.Status, "expired") {
		taskInfo.Reason = "task expired"
	}
	return taskInfo, nil
}

func (a *TaskAdaptor) GetModelList() []string {
	return ModelList
}

func (a *TaskAdaptor) GetChannelName() string {
	return ChannelName
}

func (a *TaskAdaptor) ConvertToOpenAIVideo(task *model.Task) ([]byte, error) {
	openAIVideo := task.ToOpenAIVideo()
	return common.Marshal(openAIVideo)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func xaiStatusToTaskStatus(status string) string {
	switch strings.ToLower(status) {
	case "queued", "pending", "submitted":
		return string(model.TaskStatusQueued)
	case "running", "processing", "in_progress", "generating":
		return string(model.TaskStatusInProgress)
	case "done", "completed", "succeeded", "success":
		return string(model.TaskStatusSuccess)
	case "failed", "failure", "cancelled", "canceled", "expired":
		return string(model.TaskStatusFailure)
	default:
		return string(model.TaskStatusUnknown)
	}
}

func xaiStatusToOpenAIVideoStatus(status string) string {
	return model.TaskStatus(xaiStatusToTaskStatus(status)).ToVideoStatus()
}

func xaiProgress(status string, progress int) string {
	if progress > 0 {
		return fmt.Sprintf("%d%%", progress)
	}
	switch strings.ToLower(status) {
	case "queued", "pending", "submitted":
		return taskcommon.ProgressQueued
	case "running", "processing", "in_progress", "generating":
		return taskcommon.ProgressInProgress
	case "done", "completed", "succeeded", "success":
		return taskcommon.ProgressComplete
	case "failed", "failure", "cancelled", "canceled", "expired":
		return taskcommon.ProgressComplete
	default:
		return ""
	}
}
