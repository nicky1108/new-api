package xai

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/tidwall/gjson"
)

func floatPtr(v float64) *float64 {
	return &v
}

func uintPtr(v uint) *uint {
	return &v
}

func boolPtr(v bool) *bool {
	return &v
}

func intPtr(v int) *int {
	return &v
}

func TestConvertOpenAIRequestSanitizesReasoningModelPlaygroundDefaults(t *testing.T) {
	req := &dto.GeneralOpenAIRequest{
		Model:            "xai/grok-4.3",
		Stream:           boolPtr(true),
		StreamOptions:    &dto.StreamOptions{IncludeUsage: true},
		MaxTokens:        uintPtr(4096),
		Temperature:      floatPtr(0.7),
		TopP:             floatPtr(1),
		FrequencyPenalty: floatPtr(0),
		PresencePenalty:  floatPtr(0),
		Stop:             []string{"done"},
		LogProbs:         boolPtr(false),
		TopLogProbs:      intPtr(0),
	}
	info := &relaycommon.RelayInfo{
		ChannelMeta: &relaycommon.ChannelMeta{
			UpstreamModelName: "xai/grok-4.3",
		},
	}

	got, err := (&Adaptor{}).ConvertOpenAIRequest(nil, info, req)
	if err != nil {
		t.Fatalf("ConvertOpenAIRequest returned error: %v", err)
	}

	converted, ok := got.(*dto.GeneralOpenAIRequest)
	if !ok {
		t.Fatalf("converted request type = %T, want *dto.GeneralOpenAIRequest", got)
	}
	if converted.Model != "grok-4.3" {
		t.Fatalf("model = %q, want grok-4.3", converted.Model)
	}
	if info.UpstreamModelName != "grok-4.3" {
		t.Fatalf("upstream model = %q, want grok-4.3", info.UpstreamModelName)
	}
	if converted.MaxTokens != nil {
		t.Fatalf("max_tokens should not be forwarded for xAI reasoning models")
	}
	if converted.MaxCompletionTokens == nil || *converted.MaxCompletionTokens != 4096 {
		t.Fatalf("max_completion_tokens = %#v, want 4096", converted.MaxCompletionTokens)
	}

	encoded, err := common.Marshal(converted)
	if err != nil {
		t.Fatalf("marshal converted request: %v", err)
	}
	if gjson.GetBytes(encoded, "temperature").Exists() {
		t.Fatalf("temperature should not be forwarded for xAI reasoning models: %s", encoded)
	}
	if gjson.GetBytes(encoded, "top_p").Exists() {
		t.Fatalf("top_p should not be forwarded for xAI reasoning models: %s", encoded)
	}
	if gjson.GetBytes(encoded, "frequency_penalty").Exists() {
		t.Fatalf("frequency_penalty should not be forwarded for xAI reasoning models: %s", encoded)
	}
	if gjson.GetBytes(encoded, "presence_penalty").Exists() {
		t.Fatalf("presence_penalty should not be forwarded for xAI reasoning models: %s", encoded)
	}
	if gjson.GetBytes(encoded, "stop").Exists() {
		t.Fatalf("stop should not be forwarded for xAI reasoning models: %s", encoded)
	}
	if gjson.GetBytes(encoded, "logprobs").Exists() {
		t.Fatalf("logprobs should not be forwarded for xAI reasoning models: %s", encoded)
	}
	if gjson.GetBytes(encoded, "top_logprobs").Exists() {
		t.Fatalf("top_logprobs should not be forwarded for xAI reasoning models: %s", encoded)
	}
	if gjson.GetBytes(encoded, "stream_options").Exists() {
		t.Fatalf("stream_options should not be forwarded for xAI reasoning models: %s", encoded)
	}
}

func TestConvertOpenAIRequestSanitizesReasoningSearchVariant(t *testing.T) {
	req := &dto.GeneralOpenAIRequest{
		Model:            "xai/grok-4.3-search",
		FrequencyPenalty: floatPtr(0),
		PresencePenalty:  floatPtr(0),
		Stop:             []string{"done"},
	}
	info := &relaycommon.RelayInfo{
		ChannelMeta: &relaycommon.ChannelMeta{
			UpstreamModelName: "xai/grok-4.3-search",
		},
	}

	got, err := (&Adaptor{}).ConvertOpenAIRequest(nil, info, req)
	if err != nil {
		t.Fatalf("ConvertOpenAIRequest returned error: %v", err)
	}

	payload, ok := got.(map[string]any)
	if !ok {
		t.Fatalf("converted request type = %T, want map[string]any", got)
	}
	if payload["model"] != "grok-4.3" {
		t.Fatalf("model = %#v, want grok-4.3", payload["model"])
	}
	if _, exists := payload["frequency_penalty"]; exists {
		t.Fatalf("frequency_penalty should not be forwarded for xAI reasoning search models")
	}
	if _, exists := payload["presence_penalty"]; exists {
		t.Fatalf("presence_penalty should not be forwarded for xAI reasoning search models")
	}
	if _, exists := payload["stop"]; exists {
		t.Fatalf("stop should not be forwarded for xAI reasoning search models")
	}

	searchParameters, ok := payload["search_parameters"].(map[string]any)
	if !ok {
		t.Fatalf("search_parameters = %#v, want map[string]any", payload["search_parameters"])
	}
	if searchParameters["mode"] != "on" {
		t.Fatalf("search mode = %#v, want on", searchParameters["mode"])
	}
}
