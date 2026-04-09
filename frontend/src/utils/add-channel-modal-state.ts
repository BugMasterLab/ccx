import type { Channel } from '@/services/api'

export type ChannelWatcherAction = 'load-edit-channel' | 'reset-new-form' | 'noop'

export function resolveChannelWatcherAction(params: {
  show: boolean
  newChannel: Channel | null | undefined
  oldChannel: Channel | null | undefined
}): ChannelWatcherAction {
  const { show, newChannel, oldChannel } = params

  if (!show) {
    return 'noop'
  }

  if (newChannel) {
    return 'load-edit-channel'
  }

  if (oldChannel) {
    return 'noop'
  }

  return 'reset-new-form'
}
