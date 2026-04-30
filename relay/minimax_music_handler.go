package relay

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relay/helper"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
)

type miniMaxMusicAdaptor interface {
	ConvertMiniMaxMusicRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.MiniMaxMusicRequest) (io.Reader, error)
}

func MiniMaxMusicHelper(c *gin.Context, info *relaycommon.RelayInfo) (newAPIError *types.NewAPIError) {
	info.InitChannelMeta(c)

	musicReq, ok := info.Request.(*dto.MiniMaxMusicRequest)
	if !ok {
		return types.NewError(errors.New("invalid request type"), types.ErrorCodeInvalidRequest, types.ErrOptionWithSkipRetry())
	}

	request, err := common.DeepCopy(musicReq)
	if err != nil {
		return types.NewError(fmt.Errorf("failed to copy request to MiniMaxMusicRequest: %w", err), types.ErrorCodeInvalidRequest, types.ErrOptionWithSkipRetry())
	}

	if err := helper.ModelMappedHelper(c, info, request); err != nil {
		return types.NewError(err, types.ErrorCodeChannelModelMappedError, types.ErrOptionWithSkipRetry())
	}

	adaptor := GetAdaptor(info.ApiType)
	if adaptor == nil {
		return types.NewError(fmt.Errorf("invalid api type: %d", info.ApiType), types.ErrorCodeInvalidApiType, types.ErrOptionWithSkipRetry())
	}
	musicAdaptor, ok := adaptor.(miniMaxMusicAdaptor)
	if !ok {
		return types.NewError(errors.New("selected channel does not support MiniMax music generation"), types.ErrorCodeInvalidApiType, types.ErrOptionWithSkipRetry())
	}
	adaptor.Init(info)

	ioReader, err := musicAdaptor.ConvertMiniMaxMusicRequest(c, info, *request)
	if err != nil {
		return types.NewError(err, types.ErrorCodeConvertRequestFailed, types.ErrOptionWithSkipRetry())
	}

	resp, err := adaptor.DoRequest(c, info, ioReader)
	if err != nil {
		return types.NewOpenAIError(err, types.ErrorCodeDoRequestFailed, http.StatusInternalServerError)
	}
	statusCodeMappingStr := c.GetString("status_code_mapping")

	var httpResp *http.Response
	if resp != nil {
		httpResp = resp.(*http.Response)
		if httpResp.StatusCode != http.StatusOK {
			newAPIError = service.RelayErrorHandler(c.Request.Context(), httpResp, false)
			service.ResetStatusCode(newAPIError, statusCodeMappingStr)
			return newAPIError
		}
	}

	usage, newAPIError := adaptor.DoResponse(c, httpResp, info)
	if newAPIError != nil && shouldFallbackMiniMaxMusicModel(c, request.Model) {
		fallbackModel, _ := fallbackPlaygroundMiniMaxMusicModel(request.Model)
		request.Model = fallbackModel
		c.Set("minimax_music_fallback_model", fallbackModel)

		ioReader, err = musicAdaptor.ConvertMiniMaxMusicRequest(c, info, *request)
		if err != nil {
			return types.NewError(err, types.ErrorCodeConvertRequestFailed, types.ErrOptionWithSkipRetry())
		}
		resp, err = adaptor.DoRequest(c, info, ioReader)
		if err != nil {
			return types.NewOpenAIError(err, types.ErrorCodeDoRequestFailed, http.StatusInternalServerError)
		}
		if resp != nil {
			httpResp = resp.(*http.Response)
			if httpResp.StatusCode != http.StatusOK {
				newAPIError = service.RelayErrorHandler(c.Request.Context(), httpResp, false)
				service.ResetStatusCode(newAPIError, statusCodeMappingStr)
				return newAPIError
			}
		}
		usage, newAPIError = adaptor.DoResponse(c, httpResp, info)
	}
	if newAPIError != nil {
		service.ResetStatusCode(newAPIError, statusCodeMappingStr)
		return newAPIError
	}
	service.PostTextConsumeQuota(c, info, usage.(*dto.Usage), nil)

	return nil
}

func fallbackPlaygroundMiniMaxMusicModel(model string) (string, bool) {
	switch model {
	case "music-2.6":
		return "music-2.6-free", true
	case "music-cover":
		return "music-cover-free", true
	default:
		return "", false
	}
}

func shouldFallbackMiniMaxMusicModel(c *gin.Context, model string) bool {
	if c.GetInt("minimax_music_status_code") != 2013 {
		return false
	}
	_, ok := fallbackPlaygroundMiniMaxMusicModel(model)
	return ok
}
