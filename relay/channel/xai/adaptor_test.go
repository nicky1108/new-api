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

func TestConvertOpenAIRequestSanitizesReasoningModelPlaygroundDefaults(t *testing.T) {
	req := &dto.GeneralOpenAIRequest{
		Model:            "xai/grok-4.3",
		FrequencyPenalty: floatPtr(0),
		PresencePenalty:  floatPtr(0),
		Stop:             []string{"done"},
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

	encoded, err := common.Marshal(converted)
	if err != nil {
		t.Fatalf("marshal converted request: %v", err)
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
