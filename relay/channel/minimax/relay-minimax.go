package minimax

import (
	"fmt"
	"strings"

	channelconstant "github.com/QuantumNous/new-api/constant"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relay/constant"
	"github.com/QuantumNous/new-api/types"
)

const miniMaxOfficialBaseURL = "https://api.minimaxi.com"

func ResolveMiniMaxNewAPIBaseURL(baseURL string) string {
	baseURL = strings.TrimRight(baseURL, "/")
	if baseURL == "" || baseURL == channelconstant.ChannelBaseURLs[channelconstant.ChannelTypeMiniMax] {
		return miniMaxOfficialBaseURL
	}
	return baseURL
}

func GetRequestURL(info *relaycommon.RelayInfo) (string, error) {
	baseURL := strings.TrimRight(info.ChannelBaseUrl, "/")
	if baseURL == "" {
		baseURL = channelconstant.ChannelBaseURLs[channelconstant.ChannelTypeMiniMax]
	}
	switch info.RelayFormat {
	case types.RelayFormatClaude:
		return fmt.Sprintf("%s/anthropic/v1/messages", info.ChannelBaseUrl), nil
	default:
		switch info.RelayMode {
		case constant.RelayModeChatCompletions:
			return fmt.Sprintf("%s/v1/text/chatcompletion_v2", baseURL), nil
		case constant.RelayModeImagesGenerations:
			return fmt.Sprintf("%s/v1/image_generation", baseURL), nil
		case constant.RelayModeAudioSpeech:
			return fmt.Sprintf("%s/v1/t2a_v2", ResolveMiniMaxNewAPIBaseURL(baseURL)), nil
		case constant.RelayModeMiniMaxMusic:
			return fmt.Sprintf("%s/v1/music_generation", ResolveMiniMaxNewAPIBaseURL(baseURL)), nil
		default:
			return "", fmt.Errorf("unsupported relay mode: %d", info.RelayMode)
		}
	}
}
