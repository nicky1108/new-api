package dto

import (
	"strings"

	"github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
)

type MiniMaxMusicRequest struct {
	Model         string                    `json:"model"`
	Prompt        string                    `json:"prompt,omitempty"`
	Lyrics        string                    `json:"lyrics"`
	Stream        bool                      `json:"stream,omitempty"`
	OutputFormat  string                    `json:"output_format,omitempty"`
	AudioSetting  *MiniMaxMusicAudioSetting `json:"audio_setting,omitempty"`
	AigcWatermark *bool                     `json:"aigc_watermark,omitempty"`
}

type MiniMaxMusicAudioSetting struct {
	SampleRate int    `json:"sample_rate,omitempty"`
	Bitrate    int    `json:"bitrate,omitempty"`
	Format     string `json:"format,omitempty"`
}

func (r *MiniMaxMusicRequest) GetTokenCountMeta() *types.TokenCountMeta {
	return &types.TokenCountMeta{
		CombineText: strings.TrimSpace(r.Prompt + "\n" + r.Lyrics),
		TokenType:   types.TokenTypeTextNumber,
	}
}

func (r *MiniMaxMusicRequest) IsStream(c *gin.Context) bool {
	return r.Stream
}

func (r *MiniMaxMusicRequest) SetModelName(modelName string) {
	if modelName != "" {
		r.Model = modelName
	}
}
