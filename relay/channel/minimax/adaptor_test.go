package minimax

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/dto"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"

	"github.com/gin-gonic/gin"
)

func TestGetRequestURLForImageGeneration(t *testing.T) {
	t.Parallel()

	info := &relaycommon.RelayInfo{
		RelayMode: relayconstant.RelayModeImagesGenerations,
		ChannelMeta: &relaycommon.ChannelMeta{
			ChannelBaseUrl: "https://api.minimax.chat",
		},
	}

	got, err := GetRequestURL(info)
	if err != nil {
		t.Fatalf("GetRequestURL returned error: %v", err)
	}

	want := "https://api.minimax.chat/v1/image_generation"
	if got != want {
		t.Fatalf("GetRequestURL() = %q, want %q", got, want)
	}
}

func TestGetRequestURLForMusicGeneration(t *testing.T) {
	t.Parallel()

	info := &relaycommon.RelayInfo{
		RelayMode: relayconstant.RelayModeMiniMaxMusic,
		ChannelMeta: &relaycommon.ChannelMeta{
			ChannelBaseUrl: "https://api.minimax.chat",
		},
	}

	got, err := GetRequestURL(info)
	if err != nil {
		t.Fatalf("GetRequestURL returned error: %v", err)
	}

	want := "https://api.minimax.chat/v1/music_generation"
	if got != want {
		t.Fatalf("GetRequestURL() = %q, want %q", got, want)
	}
}

func TestConvertMiniMaxMusicRequestDefaults(t *testing.T) {
	t.Parallel()

	adaptor := &Adaptor{}
	info := &relaycommon.RelayInfo{
		RelayMode:       relayconstant.RelayModeMiniMaxMusic,
		OriginModelName: "music-2.5",
	}
	request := dto.MiniMaxMusicRequest{
		Model:  "music-2.5",
		Prompt: "独立民谣, 忧郁",
		Lyrics: "[verse]\n晚风吹过旧街口",
	}

	reader, err := adaptor.ConvertMiniMaxMusicRequest(gin.CreateTestContextOnly(httptest.NewRecorder(), gin.New()), info, request)
	if err != nil {
		t.Fatalf("ConvertMiniMaxMusicRequest returned error: %v", err)
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll returned error: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if payload["model"] != "music-2.5" {
		t.Fatalf("model = %#v, want %q", payload["model"], "music-2.5")
	}
	if payload["output_format"] != "hex" {
		t.Fatalf("output_format = %#v, want hex", payload["output_format"])
	}
	audioSetting, ok := payload["audio_setting"].(map[string]any)
	if !ok {
		t.Fatalf("audio_setting = %#v, want object", payload["audio_setting"])
	}
	if audioSetting["format"] != "mp3" {
		t.Fatalf("audio_setting.format = %#v, want mp3", audioSetting["format"])
	}
	if audioSetting["sample_rate"] != float64(44100) {
		t.Fatalf("audio_setting.sample_rate = %#v, want 44100", audioSetting["sample_rate"])
	}
	if audioSetting["bitrate"] != float64(256000) {
		t.Fatalf("audio_setting.bitrate = %#v, want 256000", audioSetting["bitrate"])
	}
}

func TestConvertImageRequest(t *testing.T) {
	t.Parallel()

	adaptor := &Adaptor{}
	info := &relaycommon.RelayInfo{
		RelayMode:       relayconstant.RelayModeImagesGenerations,
		OriginModelName: "image-01",
	}
	request := dto.ImageRequest{
		Model:          "image-01",
		Prompt:         "a red fox in snowfall",
		Size:           "1536x1024",
		ResponseFormat: "url",
		N:              uintPtr(2),
	}

	got, err := adaptor.ConvertImageRequest(gin.CreateTestContextOnly(httptest.NewRecorder(), gin.New()), info, request)
	if err != nil {
		t.Fatalf("ConvertImageRequest returned error: %v", err)
	}

	body, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if payload["model"] != "image-01" {
		t.Fatalf("model = %#v, want %q", payload["model"], "image-01")
	}
	if payload["prompt"] != request.Prompt {
		t.Fatalf("prompt = %#v, want %q", payload["prompt"], request.Prompt)
	}
	if payload["n"] != float64(2) {
		t.Fatalf("n = %#v, want 2", payload["n"])
	}
	if payload["aspect_ratio"] != "3:2" {
		t.Fatalf("aspect_ratio = %#v, want %q", payload["aspect_ratio"], "3:2")
	}
	if payload["response_format"] != "url" {
		t.Fatalf("response_format = %#v, want %q", payload["response_format"], "url")
	}
}

func TestDoResponseForMusicGeneration(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Set("music_audio_format", "mp3")

	info := &relaycommon.RelayInfo{
		RelayMode: relayconstant.RelayModeMiniMaxMusic,
		StartTime: time.Unix(1700000000, 0),
	}
	info.SetEstimatePromptTokens(12)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body: ioNopCloser(`{
			"data":{"audio":"616263","status":2},
			"trace_id":"trace",
			"extra_info":{"music_duration":1000},
			"base_resp":{"status_code":0,"status_msg":"success"}
		}`),
	}

	adaptor := &Adaptor{}
	usage, err := adaptor.DoResponse(c, resp, info)
	if err != nil {
		t.Fatalf("DoResponse returned error: %v", err)
	}
	if usage == nil {
		t.Fatalf("DoResponse returned nil usage")
	}
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if got := recorder.Header().Get("Content-Type"); !strings.Contains(got, "audio/mpeg") {
		t.Fatalf("Content-Type = %q, want audio/mpeg", got)
	}
	if got := recorder.Body.String(); got != "abc" {
		t.Fatalf("body = %q, want abc", got)
	}
}

func TestDoResponseForImageGeneration(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	info := &relaycommon.RelayInfo{
		RelayMode: relayconstant.RelayModeImagesGenerations,
		StartTime: time.Unix(1700000000, 0),
	}
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       httptest.NewRecorder().Result().Body,
	}
	resp.Body = ioNopCloser(`{"data":{"image_urls":["https://example.com/minimax.png"]}}`)

	adaptor := &Adaptor{}
	usage, err := adaptor.DoResponse(c, resp, info)
	if err != nil {
		t.Fatalf("DoResponse returned error: %v", err)
	}
	if usage == nil {
		t.Fatalf("DoResponse returned nil usage")
	}

	body := recorder.Body.String()
	if !strings.Contains(body, `"url":"https://example.com/minimax.png"`) {
		t.Fatalf("response body = %s, want OpenAI image response with image URL", body)
	}
	if strings.Contains(body, `"image_urls"`) {
		t.Fatalf("response body = %s, should not expose raw MiniMax image_urls payload", body)
	}
}

type nopReadCloser struct {
	*strings.Reader
}

func (n nopReadCloser) Close() error {
	return nil
}

func ioNopCloser(body string) nopReadCloser {
	return nopReadCloser{Reader: strings.NewReader(body)}
}

func uintPtr(v uint) *uint {
	return &v
}
