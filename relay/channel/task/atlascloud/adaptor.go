package atlascloud

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

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

const (
	ChannelName       = "atlascloud"
	defaultDuration   = 5
	defaultResolution = "480p"
	defaultSeed       = -1
)

var ModelList = []string{
	"atlascloud/wan-2.2-turbo-spicy/image-to-video-lora",
}

var validDurations = map[int]bool{
	5: true,
	8: true,
}

var validResolutions = map[string]bool{
	"480p":  true,
	"720p":  true,
	"1080p": true,
}

type TaskAdaptor struct {
	taskcommon.BaseBilling
	apiKey  string
	baseURL string
}

type generateVideoRequest struct {
	Model          string `json:"model"`
	Duration       *int   `json:"duration"`
	HighNoiseLoras []any  `json:"high_noise_loras"`
	Image          string `json:"image"`
	LowNoiseLoras  []any  `json:"low_noise_loras"`
	Prompt         string `json:"prompt"`
	Resolution     string `json:"resolution"`
	Seed           *int   `json:"seed"`
}

type submitResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
	Data    struct {
		ID string `json:"id,omitempty"`
	} `json:"data"`
	ID string `json:"id,omitempty"`
}

type predictionResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
	Data    struct {
		ID      string   `json:"id,omitempty"`
		Status  string   `json:"status,omitempty"`
		Outputs []string `json:"outputs,omitempty"`
		Error   any      `json:"error,omitempty"`
	} `json:"data"`
	ID      string   `json:"id,omitempty"`
	Status  string   `json:"status,omitempty"`
	Outputs []string `json:"outputs,omitempty"`
	Error   any      `json:"error,omitempty"`
}

func (a *TaskAdaptor) Init(info *relaycommon.RelayInfo) {
	if info == nil {
		a.baseURL = constant.ChannelBaseURLs[constant.ChannelTypeAtlasCloud]
		return
	}
	a.baseURL = strings.TrimRight(info.ChannelBaseUrl, "/")
	if a.baseURL == "" {
		a.baseURL = constant.ChannelBaseURLs[constant.ChannelTypeAtlasCloud]
	}
	a.apiKey = info.ApiKey
}

func (a *TaskAdaptor) ValidateRequestAndSetAction(c *gin.Context, info *relaycommon.RelayInfo) *dto.TaskError {
	var req generateVideoRequest
	if err := common.UnmarshalBodyReusable(c, &req); err != nil {
		return service.TaskErrorWrapperLocal(err, "invalid_request", http.StatusBadRequest)
	}
	req.applyDefaults()
	if taskErr := validateGenerateVideoRequest(req); taskErr != nil {
		return taskErr
	}
	if info != nil {
		info.Action = constant.TaskActionGenerate
	}
	c.Set("atlascloud_request", req)
	c.Set("task_request", relaycommon.TaskSubmitReq{
		Prompt:   req.Prompt,
		Model:    req.Model,
		Image:    req.Image,
		Images:   []string{req.Image},
		Duration: *req.Duration,
		Size:     req.Resolution,
	})
	return nil
}

func (a *TaskAdaptor) EstimateBilling(c *gin.Context, info *relaycommon.RelayInfo) map[string]float64 {
	req, err := getAtlasCloudRequest(c)
	if err != nil {
		return nil
	}
	return map[string]float64{
		"seconds":    float64(*req.Duration),
		"resolution": 1,
	}
}

func (a *TaskAdaptor) BuildRequestURL(info *relaycommon.RelayInfo) (string, error) {
	return fmt.Sprintf("%s/api/v1/model/generateVideo", a.baseURL), nil
}

func (a *TaskAdaptor) BuildRequestHeader(c *gin.Context, req *http.Request, info *relaycommon.RelayInfo) error {
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	return nil
}

func (a *TaskAdaptor) BuildRequestBody(c *gin.Context, info *relaycommon.RelayInfo) (io.Reader, error) {
	req, err := getAtlasCloudRequest(c)
	if err != nil {
		return nil, err
	}
	if info != nil && info.ChannelMeta != nil && info.UpstreamModelName != "" {
		req.Model = info.UpstreamModelName
	}
	data, err := common.Marshal(req)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
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

	var upstreamResp submitResponse
	if err := common.Unmarshal(responseBody, &upstreamResp); err != nil {
		taskErr = service.TaskErrorWrapper(errors.Wrapf(err, "body: %s", responseBody), "unmarshal_response_body_failed", http.StatusInternalServerError)
		return
	}
	if upstreamResp.Code != 0 && upstreamResp.Code != http.StatusOK {
		message := upstreamResp.Message
		if message == "" {
			message = "atlascloud submit failed"
		}
		taskErr = service.TaskErrorWrapperLocal(fmt.Errorf("%s", message), "task_failed", http.StatusBadRequest)
		return
	}

	upstreamID := firstNonEmpty(upstreamResp.Data.ID, upstreamResp.ID)
	if upstreamID == "" {
		taskErr = service.TaskErrorWrapper(fmt.Errorf("task_id is empty"), "invalid_response", http.StatusInternalServerError)
		return
	}

	publicID := upstreamID
	if info != nil && info.TaskRelayInfo != nil && info.PublicTaskID != "" {
		publicID = info.PublicTaskID
	}
	ov := dto.NewOpenAIVideo()
	ov.ID = publicID
	ov.TaskID = publicID
	ov.CreatedAt = time.Now().Unix()
	if info != nil {
		ov.Model = info.OriginModelName
	}
	ov.Progress = 10
	c.JSON(http.StatusOK, ov)

	return upstreamID, responseBody, nil
}

func (a *TaskAdaptor) FetchTask(baseUrl, key string, body map[string]any, proxy string) (*http.Response, error) {
	taskID, ok := body["task_id"].(string)
	if !ok || strings.TrimSpace(taskID) == "" {
		return nil, fmt.Errorf("invalid task_id")
	}
	if strings.TrimSpace(baseUrl) == "" {
		baseUrl = constant.ChannelBaseURLs[constant.ChannelTypeAtlasCloud]
	}
	uri := fmt.Sprintf("%s/api/v1/model/prediction/%s", strings.TrimRight(baseUrl, "/"), taskID)

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Accept", "application/json")

	client, err := service.GetHttpClientWithProxy(proxy)
	if err != nil {
		return nil, fmt.Errorf("new proxy http client failed: %w", err)
	}
	return client.Do(req)
}

func (a *TaskAdaptor) ParseTaskResult(respBody []byte) (*relaycommon.TaskInfo, error) {
	var resp predictionResponse
	if err := common.Unmarshal(respBody, &resp); err != nil {
		return nil, errors.Wrap(err, "unmarshal task result failed")
	}
	if resp.Code != 0 && resp.Code != http.StatusOK {
		return &relaycommon.TaskInfo{
			Code:     resp.Code,
			Status:   string(model.TaskStatusFailure),
			Reason:   firstNonEmpty(resp.Message, "atlascloud prediction failed"),
			Progress: taskcommon.ProgressComplete,
		}, nil
	}

	status := firstNonEmpty(resp.Data.Status, resp.Status)
	taskInfo := &relaycommon.TaskInfo{
		Code:     0,
		TaskID:   firstNonEmpty(resp.Data.ID, resp.ID),
		Status:   atlasCloudStatusToTaskStatus(status),
		Progress: atlasCloudProgress(status),
		Reason:   atlasCloudReason(resp),
	}
	if outputs := firstNonEmptyStringSlice(resp.Data.Outputs, resp.Outputs); len(outputs) > 0 {
		taskInfo.Url = outputs[0]
	}
	if taskInfo.Status == string(model.TaskStatusSuccess) && taskInfo.Url == "" {
		taskInfo.Status = string(model.TaskStatusFailure)
		taskInfo.Reason = firstNonEmpty(taskInfo.Reason, "atlascloud completed without output")
		taskInfo.Progress = taskcommon.ProgressComplete
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
	if task.Status == model.TaskStatusFailure && task.FailReason != "" {
		openAIVideo.Error = &dto.OpenAIVideoError{Message: task.FailReason}
	}
	return common.Marshal(openAIVideo)
}

func (r *generateVideoRequest) applyDefaults() {
	if r.Duration == nil {
		duration := defaultDuration
		r.Duration = &duration
	}
	if r.Resolution == "" {
		r.Resolution = defaultResolution
	}
	if r.Seed == nil {
		seed := defaultSeed
		r.Seed = &seed
	}
	if r.HighNoiseLoras == nil {
		r.HighNoiseLoras = []any{}
	}
	if r.LowNoiseLoras == nil {
		r.LowNoiseLoras = []any{}
	}
}

func validateGenerateVideoRequest(req generateVideoRequest) *dto.TaskError {
	if strings.TrimSpace(req.Model) == "" {
		return service.TaskErrorWrapperLocal(fmt.Errorf("model field is required"), "missing_model", http.StatusBadRequest)
	}
	if strings.TrimSpace(req.Image) == "" {
		return service.TaskErrorWrapperLocal(fmt.Errorf("field image is required"), "invalid_request", http.StatusBadRequest)
	}
	if strings.TrimSpace(req.Prompt) == "" {
		return service.TaskErrorWrapperLocal(fmt.Errorf("field prompt is required"), "invalid_request", http.StatusBadRequest)
	}
	if req.Duration == nil || !validDurations[*req.Duration] {
		return service.TaskErrorWrapperLocal(fmt.Errorf("duration must be one of: 5, 8"), "invalid_request", http.StatusBadRequest)
	}
	if !validResolutions[req.Resolution] {
		return service.TaskErrorWrapperLocal(fmt.Errorf("resolution must be one of: 480p, 720p, 1080p"), "invalid_request", http.StatusBadRequest)
	}
	if len(req.HighNoiseLoras) > 3 {
		return service.TaskErrorWrapperLocal(fmt.Errorf("high_noise_loras supports at most 3 items"), "invalid_request", http.StatusBadRequest)
	}
	if len(req.LowNoiseLoras) > 3 {
		return service.TaskErrorWrapperLocal(fmt.Errorf("low_noise_loras supports at most 3 items"), "invalid_request", http.StatusBadRequest)
	}
	return nil
}

func getAtlasCloudRequest(c *gin.Context) (generateVideoRequest, error) {
	v, ok := c.Get("atlascloud_request")
	if !ok {
		return generateVideoRequest{}, fmt.Errorf("request not found in context")
	}
	req, ok := v.(generateVideoRequest)
	if !ok {
		return generateVideoRequest{}, fmt.Errorf("invalid atlascloud request type")
	}
	return req, nil
}

func atlasCloudStatusToTaskStatus(status string) string {
	switch strings.ToLower(status) {
	case "queued", "pending", "submitted", "starting":
		return string(model.TaskStatusQueued)
	case "processing", "running", "in_progress", "generating":
		return string(model.TaskStatusInProgress)
	case "completed", "succeeded", "success", "done":
		return string(model.TaskStatusSuccess)
	case "failed", "failure", "error", "cancelled", "canceled":
		return string(model.TaskStatusFailure)
	default:
		return string(model.TaskStatusUnknown)
	}
}

func atlasCloudProgress(status string) string {
	switch strings.ToLower(status) {
	case "queued", "pending", "submitted", "starting":
		return taskcommon.ProgressQueued
	case "processing", "running", "in_progress", "generating":
		return taskcommon.ProgressInProgress
	case "completed", "succeeded", "success", "done":
		return taskcommon.ProgressComplete
	case "failed", "failure", "error", "cancelled", "canceled":
		return taskcommon.ProgressComplete
	default:
		return ""
	}
}

func atlasCloudReason(resp predictionResponse) string {
	if resp.Message != "" {
		return resp.Message
	}
	if resp.Data.Error != nil {
		return fmt.Sprintf("%v", resp.Data.Error)
	}
	if resp.Error != nil {
		return fmt.Sprintf("%v", resp.Error)
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func firstNonEmptyStringSlice(values ...[]string) []string {
	for _, value := range values {
		if len(value) > 0 {
			return value
		}
	}
	return nil
}
