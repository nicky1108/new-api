package controller

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/relay/channel/minimax"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
)

type playgroundMiniMaxVoiceListRequest struct {
	Model     string `json:"model"`
	Group     string `json:"group,omitempty"`
	VoiceType string `json:"voice_type,omitempty"`
}

func PlaygroundAudioVoices(c *gin.Context) {
	useAccessToken := c.GetBool("use_access_token")
	if useAccessToken {
		playgroundVoiceOpenAIError(c, http.StatusForbidden, "暂不支持使用 access token")
		return
	}

	var request playgroundMiniMaxVoiceListRequest
	if err := common.UnmarshalBodyReusable(c, &request); err != nil {
		playgroundVoiceOpenAIError(c, http.StatusBadRequest, err.Error())
		return
	}

	voiceType, ok := minimax.NormalizeVoiceType(request.VoiceType)
	if !ok {
		playgroundVoiceOpenAIError(c, http.StatusBadRequest, "invalid voice_type")
		return
	}

	channelType := common.GetContextKeyInt(c, constant.ContextKeyChannelType)
	if channelType != constant.ChannelTypeMiniMax {
		playgroundVoiceOpenAIError(c, http.StatusBadRequest, "current model is not routed to a MiniMax channel")
		return
	}

	apiKey := strings.TrimSpace(common.GetContextKeyString(c, constant.ContextKeyChannelKey))
	if apiKey == "" {
		playgroundVoiceOpenAIError(c, http.StatusBadRequest, "MiniMax API key is empty")
		return
	}

	channelSetting, _ := common.GetContextKeyType[dto.ChannelSettings](c, constant.ContextKeyChannelSetting)
	client, err := service.GetHttpClientWithProxy(channelSetting.Proxy)
	if err != nil {
		playgroundVoiceOpenAIError(c, http.StatusInternalServerError, err.Error())
		return
	}
	if client == nil {
		client = http.DefaultClient
	}

	payload, err := common.Marshal(gin.H{"voice_type": voiceType})
	if err != nil {
		playgroundVoiceOpenAIError(c, http.StatusInternalServerError, err.Error())
		return
	}

	baseURL := minimax.ResolveMiniMaxNewAPIBaseURL(common.GetContextKeyString(c, constant.ContextKeyChannelBaseUrl))
	upstreamURL := fmt.Sprintf("%s/v1/get_voice", baseURL)
	upstreamRequest, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, upstreamURL, bytes.NewReader(payload))
	if err != nil {
		playgroundVoiceOpenAIError(c, http.StatusInternalServerError, err.Error())
		return
	}
	upstreamRequest.Header.Set("Authorization", "Bearer "+apiKey)
	upstreamRequest.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(upstreamRequest)
	if err != nil {
		playgroundVoiceOpenAIError(c, http.StatusBadGateway, err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		playgroundVoiceOpenAIError(c, http.StatusBadGateway, err.Error())
		return
	}

	var miniMaxResponse minimax.VoiceListResponse
	if err := common.Unmarshal(body, &miniMaxResponse); err != nil {
		playgroundVoiceOpenAIError(c, http.StatusBadGateway, fmt.Sprintf("failed to parse MiniMax voice response: %v", err))
		return
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		message := fmt.Sprintf("MiniMax voice endpoint returned status %d", resp.StatusCode)
		if miniMaxResponse.BaseResp.StatusMsg != "" {
			message = fmt.Sprintf("minimax voice error: %d - %s", miniMaxResponse.BaseResp.StatusCode, miniMaxResponse.BaseResp.StatusMsg)
		}
		playgroundVoiceOpenAIError(c, http.StatusBadGateway, message)
		return
	}

	if miniMaxResponse.BaseResp.StatusCode != 0 {
		playgroundVoiceOpenAIError(c, http.StatusBadRequest, fmt.Sprintf("minimax voice error: %d - %s", miniMaxResponse.BaseResp.StatusCode, miniMaxResponse.BaseResp.StatusMsg))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"voice_type": voiceType,
		"voices":     minimax.FlattenVoiceOptions(miniMaxResponse),
	})
}

func playgroundVoiceOpenAIError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"error": gin.H{
			"message": message,
			"type":    "invalid_request_error",
		},
	})
}
