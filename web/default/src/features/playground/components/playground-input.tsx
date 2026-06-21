import { useState } from 'react'
import {
  PaperclipIcon,
  FileIcon,
  ImageIcon,
  ScreenShareIcon,
  CameraIcon,
  GlobeIcon,
  SendIcon,
  SquareIcon,
  MessageSquareIcon,
  VideoIcon,
  BarChartIcon,
  BoxIcon,
  NotepadTextIcon,
  CodeSquareIcon,
  GraduationCapIcon,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { cn } from '@/lib/utils'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  PromptInput,
  PromptInputButton,
  PromptInputFooter,
  PromptInputTextarea,
  PromptInputTools,
  type PromptInputMessage,
} from '@/components/ai-elements/prompt-input'
import { Suggestion, Suggestions } from '@/components/ai-elements/suggestion'
import { ModelGroupSelector } from '@/components/model-group-selector'
import type { ModelOption, GroupOption, PlaygroundMode } from '../types'

interface PlaygroundInputProps {
  onSubmit: (text: string) => void
  onStop?: () => void
  disabled?: boolean
  isGenerating?: boolean
  models: ModelOption[]
  modelValue: string
  onModelChange: (value: string) => void
  isModelLoading?: boolean
  groups: GroupOption[]
  groupValue: string
  onGroupChange: (value: string) => void
  mode: PlaygroundMode
  onModeChange: (value: PlaygroundMode) => void
  videoImage: string
  onVideoImageChange: (value: string) => void
  videoDuration: 5 | 8
  onVideoDurationChange: (value: 5 | 8) => void
  videoResolution: '480p' | '720p' | '1080p'
  onVideoResolutionChange: (value: '480p' | '720p' | '1080p') => void
}

const suggestions = [
  { icon: BarChartIcon, text: 'Analyze data', color: '#76d0eb' },
  { icon: BoxIcon, text: 'Surprise me', color: '#76d0eb' },
  { icon: NotepadTextIcon, text: 'Summarize text', color: '#ea8444' },
  { icon: CodeSquareIcon, text: 'Code', color: '#6c71ff' },
  { icon: GraduationCapIcon, text: 'Get advice', color: '#76d0eb' },
  { icon: null, text: 'More' },
]

export function PlaygroundInput({
  onSubmit,
  onStop,
  disabled,
  isGenerating,
  models,
  modelValue,
  onModelChange,
  isModelLoading = false,
  groups,
  groupValue,
  onGroupChange,
  mode,
  onModeChange,
  videoImage,
  onVideoImageChange,
  videoDuration,
  onVideoDurationChange,
  videoResolution,
  onVideoResolutionChange,
}: PlaygroundInputProps) {
  const { t } = useTranslation()
  const [text, setText] = useState('')

  const isModelSelectDisabled =
    disabled || isModelLoading || models.length === 0
  const isGroupSelectDisabled = disabled || groups.length === 0

  const handleSubmit = (message: PromptInputMessage) => {
    if (!message.text?.trim() || disabled) return
    onSubmit(message.text)
    setText('')
  }

  const handleFileAction = (action: string) => {
    toast.info(t('Feature in development'), {
      description: action,
    })
  }

  const handleSuggestionClick = (suggestion: string) => {
    onSubmit(suggestion)
  }

  const modeOptions: Array<{
    value: PlaygroundMode
    label: string
    icon: typeof MessageSquareIcon
  }> = [
    { value: 'chat', label: t('Chat'), icon: MessageSquareIcon },
    { value: 'image', label: t('Image'), icon: ImageIcon },
    { value: 'video', label: t('Video'), icon: VideoIcon },
  ]

  return (
    <div className='grid shrink-0 gap-4 px-1 md:pb-4'>
      <PromptInput
        groupClassName='rounded-[20px] [--radius:20px]'
        onSubmit={handleSubmit}
      >
        <PromptInputTextarea
          autoComplete='off'
          autoCorrect='off'
          autoCapitalize='off'
          spellCheck={false}
          className='px-5 md:text-base'
          disabled={disabled}
          onChange={(event) => setText(event.target.value)}
          placeholder={t('Ask anything')}
          value={text}
        />

        <div className='flex flex-wrap items-center gap-2 border-t px-2.5 py-2'>
          <div
            className='border-input flex h-9 items-center gap-1 rounded-full border p-1'
            aria-label={t('Mode')}
          >
            {modeOptions.map(({ value, label, icon: Icon }) => (
              <PromptInputButton
                key={value}
                type='button'
                className={cn(
                  'h-7 rounded-full px-2.5 text-xs font-medium',
                  mode === value
                    ? 'bg-primary text-primary-foreground hover:bg-primary/90'
                    : 'text-muted-foreground'
                )}
                disabled={disabled}
                onClick={() => onModeChange(value)}
                variant={mode === value ? 'default' : 'ghost'}
              >
                <Icon size={14} />
                <span>{label}</span>
              </PromptInputButton>
            ))}
          </div>

          {mode === 'video' && (
            <div className='flex min-w-0 flex-1 flex-wrap items-center gap-2'>
              <Select
                disabled={disabled}
                value={String(videoDuration)}
                onValueChange={(value) =>
                  onVideoDurationChange(Number(value) as 5 | 8)
                }
              >
                <SelectTrigger className='h-8 w-[6.5rem] rounded-full'>
                  <SelectValue placeholder={t('Duration')} />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value='5'>5s</SelectItem>
                  <SelectItem value='8'>8s</SelectItem>
                </SelectContent>
              </Select>

              <Select
                disabled={disabled}
                value={videoResolution}
                onValueChange={(value) =>
                  onVideoResolutionChange(value as '480p' | '720p' | '1080p')
                }
              >
                <SelectTrigger className='h-8 w-[7rem] rounded-full'>
                  <SelectValue placeholder='720p' />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value='480p'>480p</SelectItem>
                  <SelectItem value='720p'>720p</SelectItem>
                  <SelectItem value='1080p'>1080p</SelectItem>
                </SelectContent>
              </Select>

              <Input
                className='h-8 min-w-[12rem] flex-1 rounded-full text-xs'
                disabled={disabled}
                onChange={(event) => onVideoImageChange(event.target.value)}
                placeholder={`${t('Image')} ${t('URL')}`}
                value={videoImage}
              />
            </div>
          )}
        </div>

        <PromptInputFooter className='p-2.5'>
          <PromptInputTools>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <PromptInputButton
                  className='!rounded-full border font-medium'
                  disabled={disabled}
                  variant='outline'
                >
                  <PaperclipIcon size={16} />
                  <span className='hidden sm:inline'>{t('Attach')}</span>
                  <span className='sr-only sm:hidden'>{t('Attach')}</span>
                </PromptInputButton>
              </DropdownMenuTrigger>
              <DropdownMenuContent align='start'>
                <DropdownMenuItem
                  onClick={() => handleFileAction('upload-file')}
                >
                  <FileIcon className='mr-2' size={16} />
                  {t('Upload file')}
                </DropdownMenuItem>
                <DropdownMenuItem
                  onClick={() => handleFileAction('upload-photo')}
                >
                  <ImageIcon className='mr-2' size={16} />
                  {t('Upload photo')}
                </DropdownMenuItem>
                <DropdownMenuItem
                  onClick={() => handleFileAction('take-screenshot')}
                >
                  <ScreenShareIcon className='mr-2' size={16} />
                  {t('Take screenshot')}
                </DropdownMenuItem>
                <DropdownMenuItem
                  onClick={() => handleFileAction('take-photo')}
                >
                  <CameraIcon className='mr-2' size={16} />
                  {t('Take photo')}
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>

            <PromptInputButton
              className='rounded-full border font-medium'
              disabled={disabled}
              onClick={() => toast.info(t('Search feature in development'))}
              variant='outline'
            >
              <GlobeIcon size={16} />
              <span className='hidden sm:inline'>{t('Search')}</span>
              <span className='sr-only sm:hidden'>{t('Search')}</span>
            </PromptInputButton>
          </PromptInputTools>

          <div className='flex items-center gap-1.5 md:gap-2'>
            <ModelGroupSelector
              selectedModel={modelValue}
              models={models}
              onModelChange={onModelChange}
              selectedGroup={groupValue}
              groups={groups}
              onGroupChange={onGroupChange}
              disabled={isModelSelectDisabled || isGroupSelectDisabled}
            />

            {isGenerating && onStop ? (
              <PromptInputButton
                className='text-foreground rounded-full font-medium'
                onClick={onStop}
                variant='secondary'
              >
                <SquareIcon className='fill-current' size={16} />
                <span className='hidden sm:inline'>{t('Stop')}</span>
                <span className='sr-only sm:hidden'>{t('Stop')}</span>
              </PromptInputButton>
            ) : (
              <PromptInputButton
                className='text-foreground rounded-full font-medium'
                disabled={disabled || !text.trim()}
                type='submit'
                variant='secondary'
              >
                <SendIcon size={16} />
                <span className='hidden sm:inline'>{t('Send')}</span>
                <span className='sr-only sm:hidden'>{t('Send')}</span>
              </PromptInputButton>
            )}
          </div>
        </PromptInputFooter>
      </PromptInput>

      <Suggestions>
        {suggestions.map(({ icon: Icon, text, color }) => (
          <Suggestion
            className={`text-xs font-normal sm:text-sm ${
              text === 'More' ? 'hidden sm:flex' : ''
            }`}
            key={text}
            onClick={() => handleSuggestionClick(text)}
            suggestion={text}
          >
            {Icon && <Icon size={16} style={{ color }} />}
            {text}
          </Suggestion>
        ))}
      </Suggestions>
    </div>
  )
}
