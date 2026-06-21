package atlascloud

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/relay/channel"
	"github.com/QuantumNous/new-api/relay/channel/openai"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
)

const ChannelName = "AtlasCloud"

var ModelList = []string{
	"deepseek-v3",
	"qwen-turbo",
	"seedream-3.0",
	"kling-v2.0",
	"atlascloud/wan-2.2-turbo-spicy/image-to-video-lora",
}

var (
	imagePollInterval = 2 * time.Second
	imagePollTimeout  = 2 * time.Minute
)

type Adaptor struct {
	openai.Adaptor
	apiKey  string
	baseURL string
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

func (a *Adaptor) Init(info *relaycommon.RelayInfo) {
	a.baseURL = constant.ChannelBaseURLs[constant.ChannelTypeAtlasCloud]
	if info == nil {
		return
	}
	a.Adaptor.Init(info)
	if baseURL := relayInfoBaseURL(info); baseURL != "" {
		a.baseURL = baseURL
	}
	a.apiKey = relayInfoAPIKey(info)
}

func (a *Adaptor) GetRequestURL(info *relaycommon.RelayInfo) (string, error) {
	if info != nil && isImageRelayMode(info.RelayMode) {
		return fmt.Sprintf("%s/api/v1/model/generateImage", atlasCloudRootBaseURL(relayInfoBaseURL(info))), nil
	}

	baseURL := atlasCloudRootBaseURL("")
	requestPath := "/v1/chat/completions"
	if info != nil {
		baseURL = atlasCloudRootBaseURL(relayInfoBaseURL(info))
		if strings.TrimSpace(info.RequestURLPath) != "" {
			requestPath = info.RequestURLPath
		}
	}
	if !strings.HasPrefix(requestPath, "/") {
		requestPath = "/" + requestPath
	}
	if !strings.HasPrefix(requestPath, "/v1/") && requestPath != "/v1" {
		requestPath = "/v1" + requestPath
	}
	return baseURL + requestPath, nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, header *http.Header, info *relaycommon.RelayInfo) error {
	channel.SetupApiRequestHeader(info, c, header)
	header.Set("Authorization", "Bearer "+a.apiKey)
	if header.Get("Accept") == "" {
		header.Set("Accept", "application/json")
	}
	if info != nil && isImageRelayMode(info.RelayMode) {
		header.Set("Content-Type", "application/json")
	}
	return nil
}

func (a *Adaptor) ConvertImageRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.ImageRequest) (any, error) {
	if info == nil || !isImageRelayMode(info.RelayMode) {
		return a.Adaptor.ConvertImageRequest(c, info, request)
	}

	payload := map[string]any{}
	if c != nil && c.Request != nil && strings.HasPrefix(c.Request.Header.Get("Content-Type"), "application/json") {
		storage, err := common.GetBodyStorage(c)
		if err != nil {
			return nil, err
		}
		bodyBytes, err := storage.Bytes()
		if err != nil {
			return nil, err
		}
		if len(bytes.TrimSpace(bodyBytes)) > 0 {
			if err := common.Unmarshal(bodyBytes, &payload); err != nil {
				return nil, err
			}
		}
	}
	if len(payload) == 0 {
		bodyBytes, err := common.Marshal(request)
		if err != nil {
			return nil, err
		}
		if err := common.Unmarshal(bodyBytes, &payload); err != nil {
			return nil, err
		}
		for key, raw := range request.Extra {
			var value any
			if err := common.Unmarshal(raw, &value); err != nil {
				return nil, err
			}
			payload[key] = value
		}
	}

	modelName := request.Model
	if info.ChannelMeta != nil && info.UpstreamModelName != "" {
		modelName = info.UpstreamModelName
	}
	if modelName == "" {
		modelName = info.OriginModelName
	}
	if strings.TrimSpace(modelName) == "" {
		return nil, errors.New("model field is required")
	}
	prompt, _ := payload["prompt"].(string)
	if strings.TrimSpace(prompt) == "" {
		prompt = request.Prompt
	}
	if strings.TrimSpace(prompt) == "" {
		return nil, errors.New("prompt field is required")
	}

	payload["model"] = modelName
	payload["prompt"] = prompt
	for _, key := range []string{
		"group",
		"n",
		"response_format",
		"size",
		"stream",
		"user",
	} {
		delete(payload, key)
	}
	return payload, nil
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (usage any, err *types.NewAPIError) {
	if info == nil || !isImageRelayMode(info.RelayMode) {
		return a.Adaptor.DoResponse(c, resp, info)
	}
	defer service.CloseResponseBodyGracefully(resp)

	responseBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, types.NewOpenAIError(readErr, types.ErrorCodeReadResponseBodyFailed, http.StatusInternalServerError)
	}

	var submitResp submitResponse
	if unmarshalErr := common.Unmarshal(responseBody, &submitResp); unmarshalErr != nil {
		return nil, types.NewOpenAIError(unmarshalErr, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
	}
	if !atlasCloudCodeOK(submitResp.Code) {
		message := firstNonEmpty(submitResp.Message, "atlascloud image submit failed")
		return nil, types.NewOpenAIError(errors.New(message), types.ErrorCodeBadResponse, http.StatusBadGateway)
	}

	predictionID := firstNonEmpty(submitResp.Data.ID, submitResp.ID)
	if predictionID == "" {
		return nil, types.NewOpenAIError(errors.New("atlascloud image prediction id is empty"), types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
	}

	outputURL, pollErr := a.pollImagePrediction(c, info, predictionID)
	if pollErr != nil {
		return nil, types.NewOpenAIError(pollErr, types.ErrorCodeBadResponse, http.StatusBadGateway)
	}
	c.JSON(http.StatusOK, dto.ImageResponse{
		Created: time.Now().Unix(),
		Data: []dto.ImageData{
			{Url: outputURL},
		},
	})
	return &dto.Usage{}, nil
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return ChannelName
}

func (a *Adaptor) pollImagePrediction(c *gin.Context, info *relaycommon.RelayInfo, predictionID string) (string, error) {
	deadline := time.Now().Add(imagePollTimeout)
	for {
		outputURL, done, err := a.fetchImagePrediction(c, info, predictionID)
		if err != nil || done {
			return outputURL, err
		}
		if time.Now().After(deadline) {
			return "", fmt.Errorf("atlascloud image prediction %s timed out", predictionID)
		}
		select {
		case <-time.After(imagePollInterval):
		case <-c.Request.Context().Done():
			return "", c.Request.Context().Err()
		}
	}
}

func (a *Adaptor) fetchImagePrediction(c *gin.Context, info *relaycommon.RelayInfo, predictionID string) (string, bool, error) {
	uri := fmt.Sprintf("%s/api/v1/model/prediction/%s", atlasCloudRootBaseURL(relayInfoBaseURL(info)), predictionID)
	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, uri, nil)
	if err != nil {
		return "", true, err
	}
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Accept", "application/json")

	client, err := service.GetHttpClientWithProxy(relayInfoProxy(info))
	if err != nil {
		return "", true, fmt.Errorf("new proxy http client failed: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", false, err
	}
	defer service.CloseResponseBodyGracefully(resp)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", true, fmt.Errorf("atlascloud prediction status %d: %s", resp.StatusCode, string(body))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", true, err
	}

	var prediction predictionResponse
	if err := common.Unmarshal(body, &prediction); err != nil {
		return "", true, err
	}
	if !atlasCloudCodeOK(prediction.Code) {
		return "", true, errors.New(firstNonEmpty(prediction.Message, "atlascloud prediction failed"))
	}

	status := strings.ToLower(firstNonEmpty(prediction.Data.Status, prediction.Status))
	outputs := firstNonEmptyStringSlice(prediction.Data.Outputs, prediction.Outputs)
	switch status {
	case "completed", "succeeded", "success", "done":
		if len(outputs) == 0 || outputs[0] == "" {
			return "", true, errors.New("atlascloud image prediction completed without output")
		}
		return outputs[0], true, nil
	case "failed", "failure", "error", "cancelled", "canceled":
		return "", true, fmt.Errorf("atlascloud image prediction failed: %s", atlasCloudReason(prediction))
	case "queued", "pending", "submitted", "starting", "processing", "running", "in_progress", "generating", "":
		return "", false, nil
	default:
		return "", false, nil
	}
}

func atlasCloudRootBaseURL(baseURL string) string {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = strings.TrimRight(constant.ChannelBaseURLs[constant.ChannelTypeAtlasCloud], "/")
	}
	baseURL = strings.TrimSuffix(baseURL, "/v1")
	return strings.TrimRight(baseURL, "/")
}

func relayInfoBaseURL(info *relaycommon.RelayInfo) string {
	if info == nil || info.ChannelMeta == nil {
		return ""
	}
	return info.ChannelBaseUrl
}

func relayInfoAPIKey(info *relaycommon.RelayInfo) string {
	if info == nil || info.ChannelMeta == nil {
		return ""
	}
	return info.ApiKey
}

func relayInfoProxy(info *relaycommon.RelayInfo) string {
	if info == nil || info.ChannelMeta == nil {
		return ""
	}
	return info.ChannelSetting.Proxy
}

func isImageRelayMode(relayMode int) bool {
	return relayMode == relayconstant.RelayModeImagesGenerations || relayMode == relayconstant.RelayModeImagesEdits
}

func atlasCloudCodeOK(code int) bool {
	return code == 0 || code == http.StatusOK
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
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
	return "atlascloud prediction failed"
}
