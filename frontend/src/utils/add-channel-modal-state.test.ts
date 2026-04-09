import { describe, expect, it } from 'vitest'

import type { Channel } from '@/services/api'
import { resolveChannelWatcherAction } from './add-channel-modal-state'

const sampleChannel: Channel = {
  index: 1,
  name: 'existing-channel',
  serviceType: 'openai',
  baseUrl: 'https://example.com/v1',
  apiKeys: ['sk-test'],
}

describe('resolveChannelWatcherAction', () => {
  it('新增模式打开时返回重置表单动作', () => {
    expect(resolveChannelWatcherAction({
      show: true,
      newChannel: null,
      oldChannel: null,
    })).toBe('reset-new-form')
  })

  it('编辑模式切入时返回回填动作', () => {
    expect(resolveChannelWatcherAction({
      show: true,
      newChannel: sampleChannel,
      oldChannel: null,
    })).toBe('load-edit-channel')
  })

  it('编辑态 channel 被清空时保持 noop，避免误切快速添加', () => {
    expect(resolveChannelWatcherAction({
      show: true,
      newChannel: null,
      oldChannel: sampleChannel,
    })).toBe('noop')
  })

  it('弹窗关闭时始终忽略 channel 变化', () => {
    expect(resolveChannelWatcherAction({
      show: false,
      newChannel: sampleChannel,
      oldChannel: null,
    })).toBe('noop')
  })
})
