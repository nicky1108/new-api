package atlascloud

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
	"github.com/gin-gonic/gin"
)

func newAtlasCloudImageContext(body string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/images/generations", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c, w
}

func TestGetRequestURL(t *testing.T) {
	adaptor := &Adaptor{}

	tests := []struct {
		name string
		info *relaycommon.RelayInfo
		want string
	}{
		{
			name: "llm appends v1 without duplicating it",
			info: &relaycommon.RelayInfo{
				RequestURLPath: "/v1/chat/completions",
				ChannelMeta: &relaycommon.ChannelMeta{
					ChannelBaseUrl: "https://api.atlascloud.ai/v1",
				},
			},
			want: "https://api.atlascloud.ai/v1/chat/completions",
		},
		{
			name: "image uses AtlasCloud generateImage endpoint",
			info: &relaycommon.RelayInfo{
				RelayMode: relayconstant.RelayModeImagesGenerations,
				ChannelMeta: &relaycommon.ChannelMeta{
					ChannelBaseUrl: "https://api.atlascloud.ai",
				},
			},
			want: "https://api.atlascloud.ai/api/v1/model/generateImage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := adaptor.GetRequestURL(tt.info)
			if err != nil {
				t.Fatalf("GetRequestURL returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("GetRequestURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestConvertImageRequestPreservesAtlasCloudFields(t *testing.T) {
	c, _ := newAtlasCloudImageContext(`{
		"model": "seedream-3.0",
		"prompt": "draw a city",
		"seed": 0,
		"size": "1024x1024",
		"response_format": "url",
		"group": "default"
	}`)
	info := &relaycommon.RelayInfo{
		OriginModelName: "seedream-3.0",
		ChannelMeta: &relaycommon.ChannelMeta{
			UpstreamModelName: "mapped-seedream",
		},
	}
	req := dto.ImageRequest{
		Model:  "seedream-3.0",
		Prompt: "draw a city",
	}

	got, err := (&Adaptor{}).ConvertImageRequest(c, info, req)
	if err != nil {
		t.Fatalf("ConvertImageRequest returned error: %v", err)
	}
	payload, ok := got.(map[string]any)
	if !ok {
		t.Fatalf("converted request type = %T, want map[string]any", got)
	}
	if payload["model"] != "mapped-seedream" {
		t.Fatalf("model = %#v, want mapped-seedream", payload["model"])
	}
	if payload["prompt"] != "draw a city" {
		t.Fatalf("prompt = %#v, want draw a city", payload["prompt"])
	}
	if payload["seed"] != float64(0) {
		t.Fatalf("seed = %#v, want explicit 0", payload["seed"])
	}
	if _, exists := payload["group"]; exists {
		t.Fatalf("group should not be forwarded to AtlasCloud")
	}
	if _, exists := payload["response_format"]; exists {
		t.Fatalf("response_format should not be forwarded to AtlasCloud")
	}
}

func TestDoResponseForImageGenerationPollsPrediction(t *testing.T) {
	var pollCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/model/prediction/prediction_id":
			pollCount++
			_, _ = w.Write([]byte(`{"code":200,"data":{"id":"prediction_id","status":"completed","outputs":["https://example.com/output.png"]}}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	originalInterval := imagePollInterval
	originalTimeout := imagePollTimeout
	imagePollInterval = time.Millisecond
	imagePollTimeout = time.Second
	defer func() {
		imagePollInterval = originalInterval
		imagePollTimeout = originalTimeout
	}()

	c, w := newAtlasCloudImageContext(`{}`)
	info := &relaycommon.RelayInfo{
		RelayMode: relayconstant.RelayModeImagesGenerations,
		ChannelMeta: &relaycommon.ChannelMeta{
			ChannelBaseUrl: server.URL,
			ApiKey:         "test-key",
		},
	}
	adaptor := &Adaptor{}
	adaptor.Init(info)

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"code":200,"data":{"id":"prediction_id"}}`)),
	}
	usage, apiErr := adaptor.DoResponse(c, resp, info)
	if apiErr != nil {
		t.Fatalf("DoResponse returned error: %v", apiErr)
	}
	if usage == nil {
		t.Fatal("usage is nil")
	}
	if pollCount != 1 {
		t.Fatalf("pollCount = %d, want 1", pollCount)
	}

	var imageResp dto.ImageResponse
	if err := common.Unmarshal(w.Body.Bytes(), &imageResp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(imageResp.Data) != 1 || imageResp.Data[0].Url != "https://example.com/output.png" {
		t.Fatalf("image response data = %#v, want output URL", imageResp.Data)
	}
}
