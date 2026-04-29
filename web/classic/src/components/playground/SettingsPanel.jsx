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

import React from 'react';
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
} from '@douyinfe/semi-ui';
import {
  Sparkles,
  Users,
  ToggleLeft,
  X,
  Settings,
  MessageSquare,
  Wand2,
} from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { renderGroupOption, selectFilter } from '../../helpers';
import ParameterControl from './ParameterControl';
import ImageUrlInput from './ImageUrlInput';
import ConfigManager from './ConfigManager';
import CustomRequestEditor from './CustomRequestEditor';
import { PLAYGROUND_REQUEST_TYPES } from '../../constants/playground.constants';

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

  const currentConfig = {
    inputs,
    parameterEnabled,
    showDebugPanel,
    customRequestMode,
    customRequestBody,
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
            onChange={(event) =>
              onInputChange('requestType', event.target.value)
            }
            disabled={customRequestMode}
            className='w-full'
          >
            <Radio value={PLAYGROUND_REQUEST_TYPES.CHAT}>{t('对话')}</Radio>
            <Radio value={PLAYGROUND_REQUEST_TYPES.IMAGE_GENERATION}>
              {t('图片生成')}
            </Radio>
            <Radio value={PLAYGROUND_REQUEST_TYPES.IMAGE_EDIT}>
              {t('图片编辑')}
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
