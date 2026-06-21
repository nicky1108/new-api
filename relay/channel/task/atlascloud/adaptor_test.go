package atlascloud

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/gin-gonic/gin"
)

func newJSONTaskContext(t *testing.T, body string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/videos", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c, w
}

func TestBuildRequestBodyPreservesAtlasCloudFields(t *testing.T) {
	c, _ := newJSONTaskContext(t, `{
		"model": "atlascloud/wan-2.2-turbo-spicy/image-to-video-lora",
		"duration": 8,
		"high_noise_loras": ["high"],
		"image": "https://example.com/input.png",
		"low_noise_loras": ["low"],
		"prompt": "make it move",
		"resolution": "1080p",
		"seed": 0
	}`)
	info := &relaycommon.RelayInfo{
		OriginModelName: "atlascloud/wan-2.2-turbo-spicy/image-to-video-lora",
		ChannelMeta: &relaycommon.ChannelMeta{
			UpstreamModelName: "mapped/atlascloud-video",
		},
	}
	adaptor := &TaskAdaptor{}

	if taskErr := adaptor.ValidateRequestAndSetAction(c, info); taskErr != nil {
		t.Fatalf("ValidateRequestAndSetAction returned error: %v", taskErr)
	}
	reader, err := adaptor.BuildRequestBody(c, info)
	if err != nil {
		t.Fatalf("BuildRequestBody returned error: %v", err)
	}
	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	var got generateVideoRequest
	if err := common.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal request body: %v", err)
	}
	if got.Model != "mapped/atlascloud-video" {
		t.Fatalf("model = %q, want mapped model", got.Model)
	}
	if got.Duration == nil || *got.Duration != 8 {
		t.Fatalf("duration = %v, want 8", got.Duration)
	}
	if got.Seed == nil || *got.Seed != 0 {
		t.Fatalf("seed = %v, want explicit 0", got.Seed)
	}
	if got.Resolution != "1080p" {
		t.Fatalf("resolution = %q, want 1080p", got.Resolution)
	}
	if len(got.HighNoiseLoras) != 1 || got.HighNoiseLoras[0] != "high" {
		t.Fatalf("high_noise_loras = %#v, want [high]", got.HighNoiseLoras)
	}
	if len(got.LowNoiseLoras) != 1 || got.LowNoiseLoras[0] != "low" {
		t.Fatalf("low_noise_loras = %#v, want [low]", got.LowNoiseLoras)
	}
}

func TestValidateRejectsInvalidAtlasCloudOptions(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			name: "invalid duration",
			body: `{"model":"atlascloud/model","image":"https://example.com/a.png","prompt":"p","duration":6}`,
		},
		{
			name: "invalid resolution",
			body: `{"model":"atlascloud/model","image":"https://example.com/a.png","prompt":"p","resolution":"4k"}`,
		},
		{
			name: "too many high noise loras",
			body: `{"model":"atlascloud/model","image":"https://example.com/a.png","prompt":"p","high_noise_loras":["a","b","c","d"]}`,
		},
		{
			name: "missing image",
			body: `{"model":"atlascloud/model","prompt":"p"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := newJSONTaskContext(t, tt.body)
			if taskErr := (&TaskAdaptor{}).ValidateRequestAndSetAction(c, &relaycommon.RelayInfo{}); taskErr == nil {
				t.Fatal("ValidateRequestAndSetAction returned nil, want error")
			}
		})
	}
}

func TestDoResponseReturnsPublicTaskID(t *testing.T) {
	c, w := newJSONTaskContext(t, `{}`)
	info := &relaycommon.RelayInfo{
		OriginModelName: "atlascloud/wan-2.2-turbo-spicy/image-to-video-lora",
		TaskRelayInfo: &relaycommon.TaskRelayInfo{
			PublicTaskID: "task_public",
		},
	}
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"code":200,"data":{"id":"prediction_id"}}`)),
	}

	taskID, taskData, taskErr := (&TaskAdaptor{}).DoResponse(c, resp, info)
	if taskErr != nil {
		t.Fatalf("DoResponse returned error: %v", taskErr)
	}
	if taskID != "prediction_id" {
		t.Fatalf("taskID = %q, want upstream prediction id", taskID)
	}
	if len(taskData) == 0 {
		t.Fatal("taskData is empty")
	}

	var public dto.OpenAIVideo
	if err := common.Unmarshal(w.Body.Bytes(), &public); err != nil {
		t.Fatalf("unmarshal public response: %v", err)
	}
	if public.ID != "task_public" || public.TaskID != "task_public" {
		t.Fatalf("public ids = %q/%q, want task_public", public.ID, public.TaskID)
	}
	if public.Status != dto.VideoStatusQueued {
		t.Fatalf("status = %q, want queued", public.Status)
	}
}

func TestParseTaskResultMapsCompletedOutput(t *testing.T) {
	body := []byte(`{"code":200,"data":{"id":"prediction_id","status":"completed","outputs":["https://example.com/video.mp4"]}}`)

	taskInfo, err := (&TaskAdaptor{}).ParseTaskResult(body)
	if err != nil {
		t.Fatalf("ParseTaskResult returned error: %v", err)
	}
	if taskInfo.Status != string(model.TaskStatusSuccess) {
		t.Fatalf("status = %q, want success", taskInfo.Status)
	}
	if taskInfo.Url != "https://example.com/video.mp4" {
		t.Fatalf("url = %q, want output url", taskInfo.Url)
	}
	if taskInfo.Progress != "100%" {
		t.Fatalf("progress = %q, want 100%%", taskInfo.Progress)
	}
}

func TestConvertToOpenAIVideoIncludesResultURL(t *testing.T) {
	task := &model.Task{
		TaskID:   "task_public",
		Status:   model.TaskStatusSuccess,
		Progress: "100%",
		Properties: model.Properties{
			OriginModelName: "atlascloud/wan-2.2-turbo-spicy/image-to-video-lora",
		},
		PrivateData: model.TaskPrivateData{
			ResultURL: "https://example.com/video.mp4",
		},
	}

	body, err := (&TaskAdaptor{}).ConvertToOpenAIVideo(task)
	if err != nil {
		t.Fatalf("ConvertToOpenAIVideo returned error: %v", err)
	}
	var video dto.OpenAIVideo
	if err := common.Unmarshal(body, &video); err != nil {
		t.Fatalf("unmarshal video: %v", err)
	}
	if video.ID != "task_public" {
		t.Fatalf("id = %q, want task_public", video.ID)
	}
	if video.Status != dto.VideoStatusCompleted {
		t.Fatalf("status = %q, want completed", video.Status)
	}
	if video.Metadata["url"] != "https://example.com/video.mp4" {
		t.Fatalf("metadata url = %#v, want result url", video.Metadata["url"])
	}
}
