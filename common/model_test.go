package common

import "testing"

func TestXAIImagineModelsAreClassifiedByGenerationType(t *testing.T) {
	if !IsImageGenerationModel("grok-imagine-image") {
		t.Fatalf("grok-imagine-image should be detected as an image generation model")
	}
	if !IsImageGenerationModel("grok-imagine-image-pro") {
		t.Fatalf("grok-imagine-image-pro should be detected as an image generation model")
	}
	if !IsImageGenerationModel("grok-2-image-1212") {
		t.Fatalf("grok-2-image-1212 should be detected as an image generation model")
	}
	if !IsVideoGenerationModel("grok-imagine-video") {
		t.Fatalf("grok-imagine-video should be detected as a video generation model")
	}
}
