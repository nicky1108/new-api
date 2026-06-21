package common

import (
	"testing"

	"github.com/QuantumNous/new-api/constant"
)

func TestAtlasCloudEndpointTypesByModel(t *testing.T) {
	tests := []struct {
		name string
		model string
		want constant.EndpointType
	}{
		{
			name:  "llm defaults to OpenAI compatible endpoint",
			model: "deepseek-v3",
			want:  constant.EndpointTypeOpenAI,
		},
		{
			name:  "image model prefers image generation endpoint",
			model: "seedream-3.0",
			want:  constant.EndpointTypeImageGeneration,
		},
		{
			name:  "video model prefers OpenAI video endpoint",
			model: "kling-v2.0",
			want:  constant.EndpointTypeOpenAIVideo,
		},
		{
			name:  "atlascloud image namespace is not mistaken for video",
			model: "atlascloud/qwen-image-edit",
			want:  constant.EndpointTypeImageGeneration,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetEndpointTypesByChannelType(constant.ChannelTypeAtlasCloud, tt.model)
			if len(got) == 0 {
				t.Fatal("got no endpoint types")
			}
			if got[0] != tt.want {
				t.Fatalf("first endpoint type = %q, want %q; all = %#v", got[0], tt.want, got)
			}
		})
	}
}
