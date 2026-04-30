package minimax

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
)

type MiniMaxMusicResponse struct {
	Data         MiniMaxMusicData      `json:"data"`
	TraceID      string                `json:"trace_id"`
	ExtraInfo    MiniMaxMusicExtraInfo `json:"extra_info"`
	AnalysisInfo any                   `json:"analysis_info"`
	BaseResp     MiniMaxBaseResp       `json:"base_resp"`
}

type MiniMaxMusicData struct {
	Audio  string `json:"audio"`
	Status int    `json:"status"`
}

type MiniMaxMusicExtraInfo struct {
	MusicDuration   int `json:"music_duration"`
	MusicSampleRate int `json:"music_sample_rate"`
	MusicChannel    int `json:"music_channel"`
	Bitrate         int `json:"bitrate"`
	MusicSize       int `json:"music_size"`
}

func (a *Adaptor) ConvertMiniMaxMusicRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.MiniMaxMusicRequest) (io.Reader, error) {
	if request.OutputFormat == "" {
		request.OutputFormat = "hex"
	}
	if request.AudioSetting == nil {
		request.AudioSetting = &dto.MiniMaxMusicAudioSetting{}
	}
	if request.AudioSetting.Format == "" {
		request.AudioSetting.Format = "mp3"
	}
	if request.AudioSetting.SampleRate == 0 {
		request.AudioSetting.SampleRate = 44100
	}
	if request.AudioSetting.Bitrate == 0 {
		request.AudioSetting.Bitrate = 256000
	}

	if request.Model == "" {
		request.Model = info.OriginModelName
	}
	request.Stream = false
	c.Set("minimax_music_request_model", request.Model)
	c.Set("music_audio_format", request.AudioSetting.Format)

	jsonData, err := common.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error marshalling minimax music request: %w", err)
	}
	return bytes.NewReader(jsonData), nil
}

func fallbackMiniMaxMusicModel(model string) (string, bool) {
	switch model {
	case "music-2.6":
		return "music-2.6-free", true
	case "music-cover":
		return "music-cover-free", true
	default:
		return "", false
	}
}

func miniMaxMusicErrorMessage(model string, baseResp MiniMaxBaseResp) string {
	message := fmt.Sprintf("minimax music error: %d - %s", baseResp.StatusCode, baseResp.StatusMsg)
	if baseResp.StatusCode != 2013 {
		return message
	}
	if fallbackModel, ok := fallbackMiniMaxMusicModel(model); ok {
		return fmt.Sprintf("%s; MiniMax restricts %s to Token Plan or paid accounts, try %s for regular API keys", message, model, fallbackModel)
	}
	return message
}

func handleMusicResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (usage any, err *types.NewAPIError) {
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, types.NewErrorWithStatusCode(
			fmt.Errorf("failed to read minimax music response: %w", readErr),
			types.ErrorCodeReadResponseBodyFailed,
			http.StatusInternalServerError,
		)
	}
	defer resp.Body.Close()

	var minimaxResp MiniMaxMusicResponse
	if unmarshalErr := common.Unmarshal(body, &minimaxResp); unmarshalErr != nil {
		return nil, types.NewErrorWithStatusCode(
			fmt.Errorf("failed to unmarshal minimax music response: %w", unmarshalErr),
			types.ErrorCodeBadResponseBody,
			http.StatusInternalServerError,
		)
	}

	if minimaxResp.BaseResp.StatusCode != 0 {
		c.Set("minimax_music_status_code", int(minimaxResp.BaseResp.StatusCode))
		c.Set("minimax_music_status_msg", minimaxResp.BaseResp.StatusMsg)
		model := c.GetString("minimax_music_request_model")
		if model == "" {
			model = info.OriginModelName
		}
		return nil, types.NewErrorWithStatusCode(
			fmt.Errorf("%s", miniMaxMusicErrorMessage(model, minimaxResp.BaseResp)),
			types.ErrorCodeBadResponse,
			http.StatusBadRequest,
		)
	}
	if minimaxResp.Data.Audio == "" {
		return nil, types.NewErrorWithStatusCode(
			fmt.Errorf("no audio data in minimax music response"),
			types.ErrorCodeBadResponse,
			http.StatusBadRequest,
		)
	}

	if strings.HasPrefix(minimaxResp.Data.Audio, "http") {
		c.Redirect(http.StatusFound, minimaxResp.Data.Audio)
	} else {
		audioData, decodeErr := hex.DecodeString(minimaxResp.Data.Audio)
		if decodeErr != nil {
			return nil, types.NewErrorWithStatusCode(
				fmt.Errorf("failed to decode hex music data: %w", decodeErr),
				types.ErrorCodeBadResponse,
				http.StatusInternalServerError,
			)
		}

		contentType := getContentTypeByFormat(c.GetString("music_audio_format"))
		c.Data(http.StatusOK, contentType, audioData)
	}

	promptTokens := info.GetEstimatePromptTokens()
	if promptTokens <= 0 {
		promptTokens = 1
	}
	usage = &dto.Usage{
		PromptTokens:     promptTokens,
		CompletionTokens: 0,
		TotalTokens:      promptTokens,
	}

	return usage, nil
}
