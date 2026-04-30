package minimax

import (
	"fmt"
	"strings"
)

type VoiceListResponse struct {
	SystemVoice     []VoiceItem     `json:"system_voice,omitempty"`
	VoiceCloning    []VoiceItem     `json:"voice_cloning,omitempty"`
	VoiceGeneration []VoiceItem     `json:"voice_generation,omitempty"`
	BaseResp        MiniMaxBaseResp `json:"base_resp"`
}

type VoiceItem struct {
	VoiceID     string   `json:"voice_id"`
	VoiceName   string   `json:"voice_name,omitempty"`
	Description []string `json:"description,omitempty"`
	CreatedTime string   `json:"created_time,omitempty"`
}

type VoiceOption struct {
	VoiceID      string   `json:"voice_id"`
	VoiceName    string   `json:"voice_name,omitempty"`
	Description  []string `json:"description,omitempty"`
	CreatedTime  string   `json:"created_time,omitempty"`
	Category     string   `json:"category"`
	CategoryName string   `json:"category_name"`
	Label        string   `json:"label"`
}

func NormalizeVoiceType(voiceType string) (string, bool) {
	voiceType = strings.TrimSpace(voiceType)
	if voiceType == "" {
		return "all", true
	}
	switch voiceType {
	case "system", "voice_cloning", "voice_generation", "all":
		return voiceType, true
	default:
		return "", false
	}
}

func FlattenVoiceOptions(response VoiceListResponse) []VoiceOption {
	voices := make([]VoiceOption, 0, len(response.SystemVoice)+len(response.VoiceCloning)+len(response.VoiceGeneration))
	voices = appendVoiceOptions(voices, response.SystemVoice, "system_voice", "系统音色")
	voices = appendVoiceOptions(voices, response.VoiceCloning, "voice_cloning", "复刻音色")
	voices = appendVoiceOptions(voices, response.VoiceGeneration, "voice_generation", "生成音色")
	return voices
}

func appendVoiceOptions(voices []VoiceOption, items []VoiceItem, category string, categoryName string) []VoiceOption {
	for _, item := range items {
		voiceID := strings.TrimSpace(item.VoiceID)
		if voiceID == "" {
			continue
		}
		voiceName := strings.TrimSpace(item.VoiceName)
		label := voiceID
		if voiceName != "" {
			label = fmt.Sprintf("%s (%s)", voiceName, voiceID)
		}
		voices = append(voices, VoiceOption{
			VoiceID:      voiceID,
			VoiceName:    voiceName,
			Description:  item.Description,
			CreatedTime:  item.CreatedTime,
			Category:     category,
			CategoryName: categoryName,
			Label:        label,
		})
	}
	return voices
}
