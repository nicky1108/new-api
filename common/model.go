package common

import "strings"

var (
	// OpenAIResponseOnlyModels is a list of models that are only available for OpenAI responses.
	OpenAIResponseOnlyModels = []string{
		"o3-pro",
		"o3-deep-research",
		"o4-mini-deep-research",
	}
	ImageGenerationModels = []string{
		"dall-e-3",
		"dall-e-2",
		"gpt-image-1",
		"prefix:imagen-",
		"flux-",
		"flux.1-",
		"grok-imagine-image",
		"grok-2-image",
	}
	VideoGenerationModels = []string{
		"grok-imagine-video",
		"sora-",
		"atlascloud/wan-2.2-turbo-spicy/image-to-video-lora",
	}
	OpenAITextModels = []string{
		"gpt-",
		"o1",
		"o3",
		"o4",
		"chatgpt",
	}
)

func IsOpenAIResponseOnlyModel(modelName string) bool {
	for _, m := range OpenAIResponseOnlyModels {
		if strings.Contains(modelName, m) {
			return true
		}
	}
	return false
}

func IsImageGenerationModel(modelName string) bool {
	modelName = strings.ToLower(modelName)
	for _, m := range ImageGenerationModels {
		if strings.Contains(modelName, m) {
			return true
		}
		if strings.HasPrefix(m, "prefix:") && strings.HasPrefix(modelName, strings.TrimPrefix(m, "prefix:")) {
			return true
		}
	}
	return false
}

func IsVideoGenerationModel(modelName string) bool {
	modelName = strings.ToLower(modelName)
	for _, m := range VideoGenerationModels {
		if strings.Contains(modelName, m) {
			return true
		}
		if strings.HasPrefix(m, "prefix:") && strings.HasPrefix(modelName, strings.TrimPrefix(m, "prefix:")) {
			return true
		}
	}
	return false
}

func IsOpenAITextModel(modelName string) bool {
	modelName = strings.ToLower(modelName)
	for _, m := range OpenAITextModels {
		if strings.Contains(modelName, m) {
			return true
		}
	}
	return false
}

func IsAtlasCloudImageModel(modelName string) bool {
	modelName = strings.ToLower(strings.TrimSpace(modelName))
	for _, marker := range []string{
		"seedream-",
		"atlascloud/ghibli",
		"atlascloud/hidream",
		"atlascloud/hunyuan-image",
		"atlascloud/imagen",
		"atlascloud/image",
		"atlascloud/neta-lumina",
		"atlascloud/qwen-image",
		"atlascloud/real-esrgan",
		"atlascloud/step1x",
	} {
		if strings.Contains(modelName, marker) {
			return true
		}
	}
	return IsImageGenerationModel(modelName)
}

func IsAtlasCloudVideoModel(modelName string) bool {
	modelName = strings.ToLower(strings.TrimSpace(modelName))
	for _, marker := range []string{
		"kling-",
		"kling/",
		"hailuo",
		"minimax/video",
		"runway",
		"vidu",
		"image-to-video",
		"text-to-video",
		"video-to-video",
		"i2v",
		"t2v",
		"v2v",
		"-video",
		"/video",
	} {
		if strings.Contains(modelName, marker) {
			return true
		}
	}
	return IsVideoGenerationModel(modelName)
}
