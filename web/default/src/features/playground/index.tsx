import { useCallback, useEffect, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import {
  getUserModels,
  getUserGroups,
  getVideoGeneration,
  sendImageGeneration,
  sendVideoGeneration,
} from './api'
import { PlaygroundChat } from './components/playground-chat'
import { PlaygroundInput } from './components/playground-input'
import { DEFAULT_GROUP, MESSAGE_STATUS } from './constants'
import { usePlaygroundState, useChatHandler } from './hooks'
import {
  createUserMessage,
  createLoadingAssistantMessage,
  updateAssistantMessageWithError,
  updateLastAssistantMessage,
} from './lib'
import type {
  Message as MessageType,
  MessageMedia,
  VideoGenerationResponse,
} from './types'

const VIDEO_POLL_INTERVAL_MS = 3000
const VIDEO_POLL_MAX_ATTEMPTS = 80

export function Playground() {
  const { t } = useTranslation()
  const {
    config,
    parameterEnabled,
    messages,
    models,
    groups,
    updateMessages,
    setModels,
    setGroups,
    updateConfig,
  } = usePlaygroundState()

  const {
    sendChat,
    stopGeneration,
    isGenerating: isChatGenerating,
  } = useChatHandler({
    config,
    parameterEnabled,
    onMessageUpdate: updateMessages,
  })
  const [isMediaGenerating, setIsMediaGenerating] = useState(false)
  const isGenerating = isChatGenerating || isMediaGenerating

  // Edit dialog state
  const [editingMessageKey, setEditingMessageKey] = useState<string | null>(
    null
  )

  // Load models
  const { data: modelsData, isLoading: isLoadingModels } = useQuery({
    queryKey: ['playground-models'],
    queryFn: getUserModels,
  })

  // Load groups
  const { data: groupsData } = useQuery({
    queryKey: ['playground-groups'],
    queryFn: getUserGroups,
  })

  // Update models when data changes
  useEffect(() => {
    if (!modelsData) return

    setModels(modelsData)

    // Set default model if current model is not available
    const isCurrentModelValid = modelsData.some((m) => m.value === config.model)
    if (modelsData.length > 0 && !isCurrentModelValid) {
      updateConfig('model', modelsData[0].value)
    }
  }, [modelsData, config.model, setModels, updateConfig])

  // Update groups when data changes
  useEffect(() => {
    if (!groupsData) return

    // Add auto group if not present
    const hasAutoGroup = groupsData.some((g) => g.value === DEFAULT_GROUP)
    const processedGroups = hasAutoGroup
      ? groupsData
      : [
          {
            value: DEFAULT_GROUP,
            label: 'Auto',
            ratio: 1,
            desc: 'Circuit Breaker',
          },
          ...groupsData,
        ]

    setGroups(processedGroups)
  }, [groupsData, setGroups])

  const completeAssistantMessage = useCallback(
    (content: string, media?: MessageMedia) => {
      updateMessages((prev) =>
        updateLastAssistantMessage(prev, (message) => ({
          ...message,
          versions: [{ ...message.versions[0], content }],
          media,
          status: MESSAGE_STATUS.COMPLETE,
          isReasoningStreaming: false,
        }))
      )
    },
    [updateMessages]
  )

  const updateAssistantMessageContent = useCallback(
    (content: string) => {
      updateMessages((prev) =>
        updateLastAssistantMessage(prev, (message) => ({
          ...message,
          versions: [{ ...message.versions[0], content }],
          status: MESSAGE_STATUS.STREAMING,
          isReasoningStreaming: false,
        }))
      )
    },
    [updateMessages]
  )

  const failAssistantMessage = useCallback(
    (error: unknown) => {
      updateMessages((prev) =>
        updateAssistantMessageWithError(prev, getErrorMessage(error))
      )
    },
    [updateMessages]
  )

  const handleImageGeneration = useCallback(
    async (text: string) => {
      setIsMediaGenerating(true)
      try {
        const response = await sendImageGeneration({
          model: config.model,
          group: config.group,
          prompt: text,
          response_format: 'url',
        })
        const firstImage = response.data?.[0]
        const url =
          firstImage?.url ||
          (firstImage?.b64_json
            ? `data:image/png;base64,${firstImage.b64_json}`
            : '')
        if (!url) {
          throw new Error('image output is empty')
        }
        completeAssistantMessage(`${t('Image')} ${t('Completed')}`, {
          type: 'image',
          url,
        })
      } catch (error) {
        failAssistantMessage(error)
      } finally {
        setIsMediaGenerating(false)
      }
    },
    [
      completeAssistantMessage,
      config.group,
      config.model,
      failAssistantMessage,
      t,
    ]
  )

  const handleVideoGeneration = useCallback(
    async (text: string) => {
      setIsMediaGenerating(true)
      try {
        const payload = {
          model: config.model,
          group: config.group,
          prompt: text,
          duration: config.videoDuration,
          resolution: config.videoResolution,
          ...(config.videoImage.trim()
            ? { image: config.videoImage.trim() }
            : {}),
          ...(parameterEnabled.seed && config.seed !== null
            ? { seed: config.seed }
            : {}),
        }
        let response = await sendVideoGeneration(payload)
        let taskID = response.id || response.task_id
        if (!taskID) {
          throw new Error('video task id is empty')
        }
        updateAssistantMessageContent(renderVideoStatus(response, t))

        for (let attempt = 0; attempt < VIDEO_POLL_MAX_ATTEMPTS; attempt++) {
          const videoURL = getVideoURL(response)
          if (response.status === 'completed' && videoURL) {
            completeAssistantMessage(renderVideoStatus(response, t), {
              type: 'video',
              url: videoURL,
            })
            return
          }
          if (response.status === 'failed') {
            throw new Error(
              response.error?.message || 'video generation failed'
            )
          }

          await sleep(VIDEO_POLL_INTERVAL_MS)
          response = await getVideoGeneration(taskID)
          taskID = response.id || response.task_id || taskID
          updateAssistantMessageContent(renderVideoStatus(response, t))
        }

        throw new Error('video generation timed out')
      } catch (error) {
        failAssistantMessage(error)
      } finally {
        setIsMediaGenerating(false)
      }
    },
    [
      completeAssistantMessage,
      config.group,
      config.model,
      config.seed,
      config.videoDuration,
      config.videoImage,
      config.videoResolution,
      failAssistantMessage,
      parameterEnabled.seed,
      t,
      updateAssistantMessageContent,
    ]
  )

  const handleSendMessage = (text: string) => {
    const userMessage = createUserMessage(text)
    const assistantMessage = createLoadingAssistantMessage()

    const newMessages = [...messages, userMessage, assistantMessage]
    updateMessages(newMessages)

    if (config.mode === 'image') {
      void handleImageGeneration(text)
      return
    }
    if (config.mode === 'video') {
      void handleVideoGeneration(text)
      return
    }
    sendChat(newMessages)
  }

  const handleCopyMessage = (message: MessageType) => {
    // Copy is handled in MessageActions component
    // eslint-disable-next-line no-console
    console.log('Message copied:', message.key)
  }

  const handleRegenerateMessage = (message: MessageType) => {
    // Find the message index and regenerate from there
    const messageIndex = messages.findIndex((m) => m.key === message.key)
    if (messageIndex === -1) return

    // Remove messages after this one and regenerate
    const messagesUpToHere = messages.slice(0, messageIndex)
    const loadingMessage = createLoadingAssistantMessage()
    const newMessages = [...messagesUpToHere, loadingMessage]

    updateMessages(newMessages)
    sendChat(newMessages)
  }

  const handleEditMessage = useCallback((message: MessageType) => {
    setEditingMessageKey(message.key)
  }, [])

  const handleEditOpenChange = useCallback((open: boolean) => {
    if (!open) setEditingMessageKey(null)
  }, [])

  // Apply edit and optionally re-submit from the edited user message
  const applyEdit = useCallback(
    (newContent: string, submit: boolean) => {
      if (!editingMessageKey) return
      const index = messages.findIndex((m) => m.key === editingMessageKey)
      if (index === -1) return

      const updated = messages.map((m) =>
        m.key === editingMessageKey
          ? { ...m, versions: [{ ...m.versions[0], content: newContent }] }
          : m
      )

      setEditingMessageKey(null)

      if (!submit || updated[index].from !== 'user') {
        updateMessages(updated)
        return
      }

      const toSubmit = [
        ...updated.slice(0, index + 1),
        createLoadingAssistantMessage(),
      ]
      updateMessages(toSubmit)
      sendChat(toSubmit)
    },
    [editingMessageKey, messages, updateMessages, sendChat]
  )

  const handleDeleteMessage = (message: MessageType) => {
    const newMessages = messages.filter((m) => m.key !== message.key)
    updateMessages(newMessages)
  }

  return (
    <div className='relative flex size-full flex-col overflow-hidden'>
      {/* Full-width scroll container: scrolling works even over side whitespace */}
      <div className='flex flex-1 flex-col overflow-hidden'>
        <PlaygroundChat
          messages={messages}
          onCopyMessage={handleCopyMessage}
          onRegenerateMessage={handleRegenerateMessage}
          onEditMessage={handleEditMessage}
          onDeleteMessage={handleDeleteMessage}
          isGenerating={isGenerating}
          editingKey={editingMessageKey}
          onCancelEdit={handleEditOpenChange}
          onSaveEdit={(newContent) => applyEdit(newContent, false)}
          onSaveEditAndSubmit={(newContent) => applyEdit(newContent, true)}
        />
      </div>

      {/* Input area: center content and constrain to the same container width */}
      <div className='mx-auto w-full max-w-4xl'>
        <PlaygroundInput
          disabled={isGenerating}
          groups={groups}
          groupValue={config.group}
          isGenerating={isGenerating}
          isModelLoading={isLoadingModels}
          modelValue={config.model}
          models={models}
          onGroupChange={(value) => updateConfig('group', value)}
          mode={config.mode}
          onModeChange={(value) => updateConfig('mode', value)}
          onModelChange={(value) => updateConfig('model', value)}
          onStop={config.mode === 'chat' ? stopGeneration : undefined}
          onSubmit={handleSendMessage}
          videoDuration={config.videoDuration}
          videoImage={config.videoImage}
          videoResolution={config.videoResolution}
          onVideoDurationChange={(value) =>
            updateConfig('videoDuration', value)
          }
          onVideoImageChange={(value) => updateConfig('videoImage', value)}
          onVideoResolutionChange={(value) =>
            updateConfig('videoResolution', value)
          }
        />
      </div>
    </div>
  )
}

function sleep(ms: number) {
  return new Promise((resolve) => window.setTimeout(resolve, ms))
}

function getVideoURL(response: VideoGenerationResponse) {
  return response.metadata?.url || response.metadata?.video_url || ''
}

function renderVideoStatus(
  response: VideoGenerationResponse,
  t: (key: string) => string
) {
  const taskID = response.id || response.task_id || ''
  const lines = [
    `${t('Video')} ${t('Task')}: ${taskID}`,
    `${t('Status')}: ${response.status}`,
  ]
  if (typeof response.progress === 'number') {
    lines.push(`${t('Progress')}: ${response.progress}%`)
  }
  return lines.join('\n')
}

function getErrorMessage(error: unknown) {
  const err = error as {
    response?: {
      data?: {
        message?: string
        error?: {
          message?: string
        }
      }
    }
    message?: string
  }
  return (
    err?.response?.data?.error?.message ||
    err?.response?.data?.message ||
    err?.message ||
    'Request error occurred'
  )
}
