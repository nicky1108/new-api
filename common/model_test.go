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

func TestAtlasCloudModelsAreClassifiedByGenerationType(t *testing.T) {
	if !IsAtlasCloudImageModel("seedream-3.0") {
		t.Fatalf("seedream-3.0 should be detected as an AtlasCloud image model")
	}
	if !IsAtlasCloudImageModel("atlascloud/qwen-image-edit") {
		t.Fatalf("atlascloud/qwen-image-edit should be detected as an AtlasCloud image model")
	}
	if !IsAtlasCloudVideoModel("kling-v2.0") {
		t.Fatalf("kling-v2.0 should be detected as an AtlasCloud video model")
	}
	if !IsAtlasCloudVideoModel("atlascloud/wan-2.2-turbo-spicy/image-to-video-lora") {
		t.Fatalf("atlascloud wan image-to-video model should be detected as an AtlasCloud video model")
	}
	if IsVideoGenerationModel("atlascloud/qwen-image-edit") {
		t.Fatalf("atlascloud/qwen-image-edit should not be classified as a generic video model")
	}
}
