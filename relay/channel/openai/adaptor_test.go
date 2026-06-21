package openai

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
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

func TestConvertOpenAIRequestSanitizesXAIReasoningWhenUsingOpenAICompatibleChannel(t *testing.T) {
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
			ChannelType:       constant.ChannelTypeOpenAI,
			ChannelBaseUrl:    "https://api.x.ai",
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
	for _, field := range []string{
		"temperature",
		"top_p",
		"frequency_penalty",
		"presence_penalty",
		"stop",
		"logprobs",
		"top_logprobs",
		"stream_options",
	} {
		if gjson.GetBytes(encoded, field).Exists() {
			t.Fatalf("%s should not be forwarded for xAI reasoning models: %s", field, encoded)
		}
	}
}
