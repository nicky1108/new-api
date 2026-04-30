/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import React, { useCallback, useEffect, useState } from 'react';
import {
  Card,
  Select,
  Typography,
  Button,
  Switch,
  RadioGroup,
  Radio,
  Input,
  InputNumber,
  Toast,
} from '@douyinfe/semi-ui';
import {
  Sparkles,
  Users,
  ToggleLeft,
  X,
  Settings,
  MessageSquare,
  Wand2,
  Volume2,
  Music,
  RefreshCw,
} from 'lucide-react';
import { useTranslation } from 'react-i18next';
import {
  getUserIdFromLocalStorage,
  renderGroupOption,
  selectFilter,
} from '../../helpers';
import ParameterControl from './ParameterControl';
import ImageUrlInput from './ImageUrlInput';
import ConfigManager from './ConfigManager';
import CustomRequestEditor from './CustomRequestEditor';
import {
  API_ENDPOINTS,
  PLAYGROUND_REQUEST_TYPES,
} from '../../constants/playground.constants';

const IMAGE_QUALITY_OPTIONS = [
  { label: 'auto', value: '' },
  { label: 'low', value: 'low' },
  { label: 'medium', value: 'medium' },
  { label: 'high', value: 'high' },
  { label: 'standard', value: 'standard' },
  { label: 'hd', value: 'hd' },
];

const IMAGE_RESPONSE_FORMAT_OPTIONS = [
  { label: 'auto', value: '' },
  { label: 'url', value: 'url' },
  { label: 'b64_json', value: 'b64_json' },
];

const AUDIO_FORMAT_OPTIONS = [
  { label: 'mp3', value: 'mp3' },
  { label: 'wav', value: 'wav' },
  { label: 'flac', value: 'flac' },
];

const SPEECH_EMOTION_OPTIONS = [
  { label: 'auto', value: '' },
  { label: 'happy', value: 'happy' },
  { label: 'sad', value: 'sad' },
  { label: 'angry', value: 'angry' },
  { label: 'fearful', value: 'fearful' },
  { label: 'disgusted', value: 'disgusted' },
  { label: 'surprised', value: 'surprised' },
  { label: 'neutral', value: 'neutral' },
];

const normalizeSpeechVoiceOptions = (voices = []) =>
  voices
    .filter((voice) => voice?.voice_id)
    .map((voice) => {
      const label = voice.label || voice.voice_name || voice.voice_id;
      const category = voice.category_name || voice.category || '';
      return {
        label: category ? `${label} · ${category}` : label,
        value: voice.voice_id,
      };
    });

const SettingsPanel = ({
  inputs,
  parameterEnabled,
  models,
  groups,
  styleState,
  showDebugPanel,
  customRequestMode,
  customRequestBody,
  onInputChange,
  onParameterToggle,
  onCloseSettings,
  onConfigImport,
  onConfigReset,
  onCustomRequestModeChange,
  onCustomRequestBodyChange,
  previewPayload,
  messages,
}) => {
  const { t } = useTranslation();
  const requestType = inputs.requestType || PLAYGROUND_REQUEST_TYPES.CHAT;
  const isChatMode = requestType === PLAYGROUND_REQUEST_TYPES.CHAT;
  const isImageGenerationMode =
    requestType === PLAYGROUND_REQUEST_TYPES.IMAGE_GENERATION;
  const isImageEditMode = requestType === PLAYGROUND_REQUEST_TYPES.IMAGE_EDIT;
  const isSpeechSynthesisMode =
    requestType === PLAYGROUND_REQUEST_TYPES.SPEECH_SYNTHESIS;
  const isMusicGenerationMode =
    requestType === PLAYGROUND_REQUEST_TYPES.MUSIC_GENERATION;
  const [speechVoiceOptions, setSpeechVoiceOptions] = useState([]);
  const [speechVoiceLoading, setSpeechVoiceLoading] = useState(false);
  const [speechVoiceLoadError, setSpeechVoiceLoadError] = useState('');

  const currentConfig = {
    inputs,
    parameterEnabled,
    showDebugPanel,
    customRequestMode,
    customRequestBody,
  };

  const hasModelOption = (modelName) =>
    (models || []).some(
      (option) => option?.value === modelName || option?.label === modelName,
    );

  const pickAvailableModel = (candidates) =>
    candidates.find((modelName) => hasModelOption(modelName));

  const fetchSpeechVoices = useCallback(
    async ({ silent = false, signal } = {}) => {
      if (!isSpeechSynthesisMode || customRequestMode || !inputs.model) {
        setSpeechVoiceOptions([]);
        setSpeechVoiceLoadError('');
        setSpeechVoiceLoading(false);
        return;
      }

      setSpeechVoiceLoading(true);
      setSpeechVoiceLoadError('');

      try {
        const response = await fetch(API_ENDPOINTS.AUDIO_VOICES, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'New-Api-User': getUserIdFromLocalStorage(),
          },
          body: JSON.stringify({
            model: inputs.model,
            group: inputs.group,
            voice_type: 'all',
          }),
          signal,
        });

        const data = await response.json().catch(() => ({}));
        if (!response.ok) {
          throw new Error(data?.error?.message || t('音色列表加载失败'));
        }

        setSpeechVoiceOptions(normalizeSpeechVoiceOptions(data?.voices));
      } catch (error) {
        if (error?.name === 'AbortError') {
          return;
        }
        const message = error?.message || t('音色列表加载失败');
        setSpeechVoiceOptions([]);
        setSpeechVoiceLoadError(message);
        if (!silent) {
          Toast.error({ content: message });
        }
      } finally {
        if (!signal?.aborted) {
          setSpeechVoiceLoading(false);
        }
      }
    },
    [
      customRequestMode,
      inputs.group,
      inputs.model,
      isSpeechSynthesisMode,
      t,
    ],
  );

  useEffect(() => {
    if (!isSpeechSynthesisMode) {
      return undefined;
    }
    const controller = new AbortController();
    fetchSpeechVoices({ silent: true, signal: controller.signal });
    return () => controller.abort();
  }, [fetchSpeechVoices, isSpeechSynthesisMode]);

  const handleRequestTypeChange = (value) => {
    onInputChange('requestType', value);

    if (value === PLAYGROUND_REQUEST_TYPES.SPEECH_SYNTHESIS) {
      const currentModel = inputs.model || '';
      if (!currentModel.startsWith('speech-')) {
        const speechModel = pickAvailableModel([
          'speech-2.8-turbo',
          'speech-2.6-turbo',
          'speech-02-turbo',
          'speech-01-turbo',
        ]);
        if (speechModel) {
          onInputChange('model', speechModel);
        }
      }
    }

    if (value === PLAYGROUND_REQUEST_TYPES.MUSIC_GENERATION) {
      if (!inputs.model?.startsWith('music-')) {
        const musicModel = pickAvailableModel([
          'music-2.6-free',
          'music-2.6',
          'music-cover',
          'music-cover-free',
        ]);
        if (musicModel) {
          onInputChange('model', musicModel);
        }
      }
    }
  };

  return (
    <Card
      className='h-full flex flex-col'
      bordered={false}
      bodyStyle={{
        padding: styleState.isMobile ? '16px' : '24px',
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
      }}
    >
      {/* 标题区域 - 与调试面板保持一致 */}
      <div className='flex items-center justify-between mb-6 flex-shrink-0'>
        <div className='flex items-center'>
          <div className='w-10 h-10 rounded-full bg-gradient-to-r from-purple-500 to-pink-500 flex items-center justify-center mr-3'>
            <Settings size={20} className='text-white' />
          </div>
          <Typography.Title heading={5} className='mb-0'>
            {t('模型配置')}
          </Typography.Title>
        </div>

        {styleState.isMobile && onCloseSettings && (
          <Button
            icon={<X size={16} />}
            onClick={onCloseSettings}
            theme='borderless'
            type='tertiary'
            size='small'
            className='!rounded-lg'
          />
        )}
      </div>

      {/* 移动端配置管理 */}
      {styleState.isMobile && (
        <div className='mb-4 flex-shrink-0'>
          <ConfigManager
            currentConfig={currentConfig}
            onConfigImport={onConfigImport}
            onConfigReset={onConfigReset}
            styleState={{ ...styleState, isMobile: false }}
            messages={messages}
          />
        </div>
      )}

      <div className='space-y-6 overflow-y-auto flex-1 pr-2 model-settings-scroll'>
        {/* 自定义请求体编辑器 */}
        <CustomRequestEditor
          customRequestMode={customRequestMode}
          customRequestBody={customRequestBody}
          onCustomRequestModeChange={onCustomRequestModeChange}
          onCustomRequestBodyChange={onCustomRequestBodyChange}
          defaultPayload={previewPayload}
        />

        {/* 调用类型 */}
        <div className={customRequestMode ? 'opacity-50' : ''}>
          <div className='flex items-center gap-2 mb-2'>
            <MessageSquare size={16} className='text-gray-500' />
            <Typography.Text strong className='text-sm'>
              {t('调用类型')}
            </Typography.Text>
            {customRequestMode && (
              <Typography.Text className='text-xs text-orange-600'>
                ({t('已在自定义模式中忽略')})
              </Typography.Text>
            )}
          </div>
          <RadioGroup
            type='button'
            buttonSize='small'
            value={inputs.requestType || PLAYGROUND_REQUEST_TYPES.CHAT}
            onChange={(event) => handleRequestTypeChange(event.target.value)}
            disabled={customRequestMode}
            className='w-full'
            style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}
          >
            <Radio value={PLAYGROUND_REQUEST_TYPES.CHAT}>{t('对话')}</Radio>
            <Radio value={PLAYGROUND_REQUEST_TYPES.IMAGE_GENERATION}>
              {t('图片生成')}
            </Radio>
            <Radio value={PLAYGROUND_REQUEST_TYPES.IMAGE_EDIT}>
              {t('图片编辑')}
            </Radio>
            <Radio value={PLAYGROUND_REQUEST_TYPES.SPEECH_SYNTHESIS}>
              {t('语音合成')}
            </Radio>
            <Radio value={PLAYGROUND_REQUEST_TYPES.MUSIC_GENERATION}>
              {t('音乐生成')}
            </Radio>
          </RadioGroup>
        </div>

        {/* 分组选择 */}
        <div className={customRequestMode ? 'opacity-50' : ''}>
          <div className='flex items-center gap-2 mb-2'>
            <Users size={16} className='text-gray-500' />
            <Typography.Text strong className='text-sm'>
              {t('分组')}
            </Typography.Text>
            {customRequestMode && (
              <Typography.Text className='text-xs text-orange-600'>
                ({t('已在自定义模式中忽略')})
              </Typography.Text>
            )}
          </div>
          <Select
            placeholder={t('请选择分组')}
            name='group'
            required
            selection
            filter={selectFilter}
            autoClearSearchValue={false}
            onChange={(value) => onInputChange('group', value)}
            value={inputs.group}
            autoComplete='new-password'
            optionList={groups}
            renderOptionItem={renderGroupOption}
            style={{ width: '100%' }}
            dropdownStyle={{ width: '100%', maxWidth: '100%' }}
            className='!rounded-lg'
            disabled={customRequestMode}
          />
        </div>

        {/* 模型选择 */}
        <div className={customRequestMode ? 'opacity-50' : ''}>
          <div className='flex items-center gap-2 mb-2'>
            <Sparkles size={16} className='text-gray-500' />
            <Typography.Text strong className='text-sm'>
              {t('模型')}
            </Typography.Text>
            {customRequestMode && (
              <Typography.Text className='text-xs text-orange-600'>
                ({t('已在自定义模式中忽略')})
              </Typography.Text>
            )}
          </div>
          <Select
            placeholder={t('请选择模型')}
            name='model'
            required
            selection
            filter={selectFilter}
            autoClearSearchValue={false}
            onChange={(value) => onInputChange('model', value)}
            value={inputs.model}
            autoComplete='new-password'
            optionList={models}
            style={{ width: '100%' }}
            dropdownStyle={{ width: '100%', maxWidth: '100%' }}
            className='!rounded-lg'
            disabled={customRequestMode}
          />
        </div>

        {isChatMode && (
          <div className={customRequestMode ? 'opacity-50' : ''}>
            <ImageUrlInput
              imageUrls={inputs.imageUrls}
              imageEnabled={inputs.imageEnabled}
              onImageUrlsChange={(urls) => onInputChange('imageUrls', urls)}
              onImageEnabledChange={(enabled) =>
                onInputChange('imageEnabled', enabled)
              }
              disabled={customRequestMode}
            />
          </div>
        )}

        {(isImageGenerationMode || isImageEditMode) && (
          <div className={customRequestMode ? 'opacity-50' : ''}>
            <div className='flex items-center gap-2 mb-3'>
              <Wand2 size={16} className='text-gray-500' />
              <Typography.Text strong className='text-sm'>
                {t('图片参数')}
              </Typography.Text>
            </div>
            <div className='space-y-3'>
              <div>
                <Typography.Text className='text-xs text-gray-500 mb-1 block'>
                  {t('尺寸')}
                </Typography.Text>
                <Input
                  value={inputs.imageSize}
                  onChange={(value) => onInputChange('imageSize', value)}
                  placeholder='1024x1024'
                  disabled={customRequestMode}
                  className='!rounded-lg'
                />
              </div>
              <div className='grid grid-cols-2 gap-3'>
                <div>
                  <Typography.Text className='text-xs text-gray-500 mb-1 block'>
                    {t('质量')}
                  </Typography.Text>
                  <Select
                    value={inputs.imageQuality}
                    optionList={IMAGE_QUALITY_OPTIONS}
                    onChange={(value) => onInputChange('imageQuality', value)}
                    style={{ width: '100%' }}
                    disabled={customRequestMode}
                  />
                </div>
                <div>
                  <Typography.Text className='text-xs text-gray-500 mb-1 block'>
                    {t('数量')}
                  </Typography.Text>
                  <InputNumber
                    value={inputs.imageN}
                    min={1}
                    max={10}
                    precision={0}
                    onNumberChange={(value) =>
                      onInputChange('imageN', value || 1)
                    }
                    style={{ width: '100%' }}
                    disabled={customRequestMode}
                  />
                </div>
              </div>
              <div>
                <Typography.Text className='text-xs text-gray-500 mb-1 block'>
                  {t('响应格式')}
                </Typography.Text>
                <Select
                  value={inputs.imageResponseFormat}
                  optionList={IMAGE_RESPONSE_FORMAT_OPTIONS}
                  onChange={(value) =>
                    onInputChange('imageResponseFormat', value)
                  }
                  style={{ width: '100%' }}
                  disabled={customRequestMode}
                />
              </div>
            </div>
          </div>
        )}

        {isImageEditMode && (
          <div className={customRequestMode ? 'opacity-50' : ''}>
            <ImageUrlInput
              title={t('编辑图片')}
              description={t('启用后可粘贴、上传或填写公开图片URL作为编辑输入')}
              imageUrls={inputs.imageUrls}
              imageEnabled={inputs.imageEnabled}
              onImageUrlsChange={(urls) => onInputChange('imageUrls', urls)}
              onImageEnabledChange={(enabled) =>
                onInputChange('imageEnabled', enabled)
              }
              disabled={customRequestMode}
            />
          </div>
        )}

        {isSpeechSynthesisMode && (
          <div className={customRequestMode ? 'opacity-50' : ''}>
            <div className='flex items-center gap-2 mb-3'>
              <Volume2 size={16} className='text-gray-500' />
              <Typography.Text strong className='text-sm'>
                {t('语音参数')}
              </Typography.Text>
            </div>
            <div className='space-y-3'>
              <div>
                <div className='flex items-center justify-between mb-1'>
                  <Typography.Text className='text-xs text-gray-500 block'>
                    {t('音色 ID')}
                  </Typography.Text>
                  <Button
                    icon={<RefreshCw size={14} />}
                    theme='borderless'
                    type='tertiary'
                    size='small'
                    onClick={() => fetchSpeechVoices({ silent: false })}
                    disabled={customRequestMode || speechVoiceLoading}
                    loading={speechVoiceLoading}
                    aria-label={t('刷新音色')}
                  />
                </div>
                <Select
                  value={inputs.speechVoice || undefined}
                  optionList={speechVoiceOptions}
                  onChange={(value) => onInputChange('speechVoice', value)}
                  placeholder={
                    speechVoiceLoading
                      ? t('正在加载音色')
                      : t('请选择或输入音色 ID')
                  }
                  filter={selectFilter}
                  allowCreate
                  autoClearSearchValue={false}
                  loading={speechVoiceLoading}
                  disabled={customRequestMode}
                  emptyContent={
                    speechVoiceLoadError
                      ? t('加载失败，可直接输入音色 ID')
                      : t('暂无音色')
                  }
                  style={{ width: '100%' }}
                  dropdownStyle={{ width: '100%', maxWidth: '100%' }}
                  className='!rounded-lg'
                />
                {speechVoiceLoadError && (
                  <Typography.Text className='text-xs text-red-500 mt-1 block'>
                    {speechVoiceLoadError}
                  </Typography.Text>
                )}
              </div>
              <div className='grid grid-cols-2 gap-3'>
                <div>
                  <Typography.Text className='text-xs text-gray-500 mb-1 block'>
                    {t('语速')}
                  </Typography.Text>
                  <InputNumber
                    value={inputs.speechSpeed}
                    min={0.5}
                    max={2}
                    step={0.1}
                    precision={1}
                    onNumberChange={(value) =>
                      onInputChange('speechSpeed', value || 1)
                    }
                    style={{ width: '100%' }}
                    disabled={customRequestMode}
                  />
                </div>
                <div>
                  <Typography.Text className='text-xs text-gray-500 mb-1 block'>
                    {t('音量')}
                  </Typography.Text>
                  <InputNumber
                    value={inputs.speechVolume}
                    min={0.1}
                    max={10}
                    step={0.1}
                    precision={1}
                    onNumberChange={(value) =>
                      onInputChange('speechVolume', value || 1)
                    }
                    style={{ width: '100%' }}
                    disabled={customRequestMode}
                  />
                </div>
              </div>
              <div className='grid grid-cols-2 gap-3'>
                <div>
                  <Typography.Text className='text-xs text-gray-500 mb-1 block'>
                    {t('语调')}
                  </Typography.Text>
                  <InputNumber
                    value={inputs.speechPitch}
                    min={-12}
                    max={12}
                    precision={0}
                    onNumberChange={(value) =>
                      onInputChange('speechPitch', value || 0)
                    }
                    style={{ width: '100%' }}
                    disabled={customRequestMode}
                  />
                </div>
                <div>
                  <Typography.Text className='text-xs text-gray-500 mb-1 block'>
                    {t('情绪')}
                  </Typography.Text>
                  <Select
                    value={inputs.speechEmotion}
                    optionList={SPEECH_EMOTION_OPTIONS}
                    onChange={(value) => onInputChange('speechEmotion', value)}
                    style={{ width: '100%' }}
                    disabled={customRequestMode}
                  />
                </div>
              </div>
              <div className='grid grid-cols-3 gap-3'>
                <div>
                  <Typography.Text className='text-xs text-gray-500 mb-1 block'>
                    {t('格式')}
                  </Typography.Text>
                  <Select
                    value={inputs.speechAudioFormat}
                    optionList={AUDIO_FORMAT_OPTIONS}
                    onChange={(value) =>
                      onInputChange('speechAudioFormat', value)
                    }
                    style={{ width: '100%' }}
                    disabled={customRequestMode}
                  />
                </div>
                <div>
                  <Typography.Text className='text-xs text-gray-500 mb-1 block'>
                    {t('采样率')}
                  </Typography.Text>
                  <InputNumber
                    value={inputs.speechSampleRate}
                    min={8000}
                    max={48000}
                    step={1000}
                    precision={0}
                    onNumberChange={(value) =>
                      onInputChange('speechSampleRate', value || 32000)
                    }
                    style={{ width: '100%' }}
                    disabled={customRequestMode}
                  />
                </div>
                <div>
                  <Typography.Text className='text-xs text-gray-500 mb-1 block'>
                    {t('码率')}
                  </Typography.Text>
                  <InputNumber
                    value={inputs.speechBitrate}
                    min={32000}
                    max={320000}
                    step={16000}
                    precision={0}
                    onNumberChange={(value) =>
                      onInputChange('speechBitrate', value || 128000)
                    }
                    style={{ width: '100%' }}
                    disabled={customRequestMode}
                  />
                </div>
              </div>
            </div>
          </div>
        )}

        {isMusicGenerationMode && (
          <div className={customRequestMode ? 'opacity-50' : ''}>
            <div className='flex items-center gap-2 mb-3'>
              <Music size={16} className='text-gray-500' />
              <Typography.Text strong className='text-sm'>
                {t('音乐参数')}
              </Typography.Text>
            </div>
            <div className='space-y-3'>
              <div>
                <Typography.Text className='text-xs text-gray-500 mb-1 block'>
                  {t('歌曲描述')}
                </Typography.Text>
                <Input
                  value={inputs.musicPrompt}
                  onChange={(value) => onInputChange('musicPrompt', value)}
                  placeholder={t('例如：独立民谣, 忧郁, 夜晚咖啡馆')}
                  disabled={customRequestMode}
                  className='!rounded-lg'
                />
              </div>
              <div className='grid grid-cols-3 gap-3'>
                <div>
                  <Typography.Text className='text-xs text-gray-500 mb-1 block'>
                    {t('格式')}
                  </Typography.Text>
                  <Select
                    value={inputs.musicAudioFormat}
                    optionList={AUDIO_FORMAT_OPTIONS}
                    onChange={(value) =>
                      onInputChange('musicAudioFormat', value)
                    }
                    style={{ width: '100%' }}
                    disabled={customRequestMode}
                  />
                </div>
                <div>
                  <Typography.Text className='text-xs text-gray-500 mb-1 block'>
                    {t('采样率')}
                  </Typography.Text>
                  <InputNumber
                    value={inputs.musicSampleRate}
                    min={8000}
                    max={48000}
                    step={1000}
                    precision={0}
                    onNumberChange={(value) =>
                      onInputChange('musicSampleRate', value || 44100)
                    }
                    style={{ width: '100%' }}
                    disabled={customRequestMode}
                  />
                </div>
                <div>
                  <Typography.Text className='text-xs text-gray-500 mb-1 block'>
                    {t('码率')}
                  </Typography.Text>
                  <InputNumber
                    value={inputs.musicBitrate}
                    min={32000}
                    max={320000}
                    step={16000}
                    precision={0}
                    onNumberChange={(value) =>
                      onInputChange('musicBitrate', value || 256000)
                    }
                    style={{ width: '100%' }}
                    disabled={customRequestMode}
                  />
                </div>
              </div>
            </div>
          </div>
        )}

        {/* 参数控制组件 */}
        {isChatMode && (
          <div className={customRequestMode ? 'opacity-50' : ''}>
            <ParameterControl
              inputs={inputs}
              parameterEnabled={parameterEnabled}
              onInputChange={onInputChange}
              onParameterToggle={onParameterToggle}
              disabled={customRequestMode}
            />
          </div>
        )}

        {/* 流式输出开关 */}
        {isChatMode && (
          <div className={customRequestMode ? 'opacity-50' : ''}>
            <div className='flex items-center justify-between'>
              <div className='flex items-center gap-2'>
                <ToggleLeft size={16} className='text-gray-500' />
                <Typography.Text strong className='text-sm'>
                  {t('流式输出')}
                </Typography.Text>
                {customRequestMode && (
                  <Typography.Text className='text-xs text-orange-600'>
                    ({t('已在自定义模式中忽略')})
                  </Typography.Text>
                )}
              </div>
              <Switch
                checked={inputs.stream}
                onChange={(checked) => onInputChange('stream', checked)}
                checkedText={t('开')}
                uncheckedText={t('关')}
                size='small'
                disabled={customRequestMode}
              />
            </div>
          </div>
        )}
      </div>

      {/* 桌面端的配置管理放在底部 */}
      {!styleState.isMobile && (
        <div className='flex-shrink-0 pt-3'>
          <ConfigManager
            currentConfig={currentConfig}
            onConfigImport={onConfigImport}
            onConfigReset={onConfigReset}
            styleState={styleState}
            messages={messages}
          />
        </div>
      )}
    </Card>
  );
};

export default SettingsPanel;
