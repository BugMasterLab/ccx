<template>
  <v-dialog
    :model-value="modelValue"
    max-width="1200"
    scrollable
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <v-card rounded="xl">
      <v-card-title class="d-flex align-center justify-space-between pa-4">
        <div class="d-flex align-center ga-2">
          <v-icon color="primary">mdi-flask-outline</v-icon>
          <span>{{ t('batchTest.title') }}</span>
        </div>
        <v-btn icon variant="text" @click="$emit('update:modelValue', false)">
          <v-icon>mdi-close</v-icon>
        </v-btn>
      </v-card-title>

      <v-divider />

      <v-card-text class="pa-4">
        <div class="control-panel mb-4">
          <div class="control-section">
            <div class="section-title">{{ t('batchTest.selectChannels') }}</div>
            <div class="d-flex flex-wrap align-center ga-2">
              <v-btn size="small" variant="tonal" @click="selectAll">{{ t('capability.selectAll') }}</v-btn>
              <v-btn size="small" variant="text" @click="clearSelection">{{ t('capability.deselectAll') }}</v-btn>
              <v-chip size="small" color="primary" variant="tonal">
                {{ t('batchTest.selectionSummary', { selected: selectedIds.length, total: channels.length }) }}
              </v-chip>
            </div>
          </div>

          <div class="control-section">
            <div class="section-title">{{ t('batchTest.selectProtocols') }}</div>
            <div class="d-flex flex-wrap ga-2">
              <v-chip
                v-for="protocol in protocolOptions"
                :key="protocol.value"
                :color="selectedProtocols.includes(protocol.value) ? getProtocolColor(protocol.value) : 'default'"
                :variant="selectedProtocols.includes(protocol.value) ? 'flat' : 'tonal'"
                @click="toggleProtocol(protocol.value)"
              >
                {{ protocol.label }}
              </v-chip>
            </div>
          </div>

          <div class="control-section">
            <div class="d-flex align-center justify-space-between mb-2">
              <div class="section-title mb-0">{{ t('batchTest.selectModels') }}</div>
              <div class="d-flex align-center ga-2">
                <v-progress-circular v-if="fetchingModels" indeterminate size="16" width="2" color="primary" />
                <span v-if="fetchingModels" class="text-caption text-medium-emphasis">{{ t('capability.fetchingModels') }}</span>
              </div>
            </div>
            <v-combobox
              v-model="selectedModels"
              :items="availableModels"
              :label="t('capability.selectModels')"
              multiple
              chips
              closable-chips
              clearable
              variant="outlined"
              density="comfortable"
              hide-details="auto"
            />
            <div v-if="!fetchingModels && availableModels.length === 0" class="text-caption text-medium-emphasis mt-2">
              {{ t('batchTest.noModelsHint') }}
            </div>
          </div>

          <div class="d-flex flex-wrap justify-end ga-2">
            <v-btn
              color="info"
              variant="tonal"
              prepend-icon="mdi-speedometer"
              :loading="runningLatency"
              :disabled="selectedIds.length === 0 || runningCapability"
              @click="runLatencyTests"
            >
              {{ t('batchTest.batchLatency') }}
            </v-btn>
            <v-btn
              color="success"
              variant="elevated"
              prepend-icon="mdi-test-tube"
              :loading="runningCapability"
              :disabled="selectedIds.length === 0 || runningLatency || selectedProtocols.length === 0"
              @click="runCapabilityTests"
            >
              {{ t('batchTest.batchCapability') }}
            </v-btn>
          </div>
        </div>

        <v-alert v-if="errorMessage" type="error" variant="tonal" rounded="lg" class="mb-4">
          {{ errorMessage }}
        </v-alert>

        <div class="channel-grid mb-6">
          <v-card
            v-for="channel in channels"
            :key="channel.index"
            :variant="selectedIds.includes(channel.index) ? 'flat' : 'outlined'"
            :color="selectedIds.includes(channel.index) ? 'primary' : undefined"
            class="channel-select-card"
            @click="toggleChannel(channel.index)"
          >
            <v-card-text class="pa-3">
              <div class="d-flex align-start ga-3">
                <v-checkbox-btn
                  :model-value="selectedIds.includes(channel.index)"
                  color="primary"
                  @click.stop="toggleChannel(channel.index)"
                />
                <div class="flex-grow-1 min-w-0">
                  <div class="text-body-1 font-weight-medium text-truncate">{{ channel.name }}</div>
                  <div class="text-caption text-medium-emphasis">
                    #{{ channel.index }} ? {{ getServiceLabel(channel.serviceType) }}
                  </div>
                  <div class="text-caption mt-1">{{ t('batchTest.latencyLabel') }}{{ getLatencyLabel(channel.index) }}</div>
                  <div class="text-caption mt-1">{{ t('batchTest.capabilityLabel') }}{{ getCapabilitySummary(channel.index) }}</div>
                </div>
              </div>
            </v-card-text>
          </v-card>
        </div>

        <div class="d-flex align-center justify-space-between mb-3">
          <div class="d-flex align-center ga-2 flex-wrap">
            <div class="text-subtitle-1 font-weight-bold">{{ t('batchTest.results') }}</div>
            <v-chip size="small" variant="tonal">
              {{ finishedCount }}/{{ activeResultCount }}
            </v-chip>
          </div>
          <v-btn
            v-if="hasRunningJobs"
            size="small"
            color="error"
            variant="tonal"
            :loading="cancellingAll"
            @click="cancelAllJobs"
          >
            {{ cancellingAll ? t('capability.cancelling') : t('batchTest.cancelAll') }}
          </v-btn>
        </div>

        <div v-if="orderedResults.length === 0" class="text-body-2 text-medium-emphasis py-6 text-center">
          {{ t('batchTest.empty') }}
        </div>

        <div v-else class="result-list compact-result-list">
          <v-card v-for="result in orderedResults" :key="result.channelId" variant="outlined" rounded="lg">
            <v-card-text class="pa-3">
              <div class="result-header mb-2">
                <div class="d-flex align-center ga-2 flex-wrap min-w-0">
                  <div class="text-body-2 font-weight-bold text-truncate">{{ result.channelName }}</div>
                  <v-chip size="x-small" variant="tonal" :color="getResultColor(result)">
                    {{ getResultStatus(result) }}
                  </v-chip>
                  <v-chip v-if="result.latency !== null" size="x-small" color="info" variant="tonal">
                    {{ result.latency }}ms
                  </v-chip>
                  <v-chip
                    v-for="protocol in result.compatibleProtocols"
                    :key="protocol"
                    size="x-small"
                    :color="getProtocolColor(protocol)"
                    variant="tonal"
                  >
                    {{ getProtocolDisplayName(protocol) }}
                  </v-chip>
                </div>
                <v-btn
                  v-if="isJobRunning(result)"
                  size="x-small"
                  color="error"
                  variant="text"
                  :loading="cancellingChannelIds.includes(result.channelId)"
                  @click="cancelSingleJob(result.channelId)"
                >
                  {{ t('batchTest.cancelChannel') }}
                </v-btn>
              </div>

              <div v-if="result.progressText" class="text-caption text-medium-emphasis mb-2">
                {{ result.progressText }}
              </div>
              <div v-if="result.error" class="text-caption text-error mb-3">
                {{ result.error }}
              </div>

              <div v-if="result.job?.tests?.length" class="protocol-result-list compact-protocol-list">
                <div v-for="test in sortedTests(result.job.tests)" :key="`${result.channelId}-${test.protocol}`" class="protocol-result-card compact-protocol-card">
                  <div class="protocol-result-header">
                    <v-chip :color="getProtocolColor(test.protocol)" size="small" variant="tonal">
                      {{ getProtocolDisplayName(test.protocol) }}
                    </v-chip>
                    <div class="d-flex align-center ga-2 flex-wrap">
                      <span class="text-caption" :class="getProtocolStatusClass(test.status)">
                        {{ getProtocolStatusLabel(test.status) }}
                      </span>
                      <span class="text-caption text-medium-emphasis">{{ formatSuccessRatio(test) }}</span>
                      <span v-if="hasProtocolLatency(test)" class="text-caption text-medium-emphasis">{{ getAverageLatency(test) }}ms</span>
                      <span v-if="test.success" class="text-caption" :class="test.streamingSupported ? 'text-success' : 'text-warning'">
                        {{ test.streamingSupported ? t('capability.supported') : t('capability.unsupported') }}
                      </span>
                    </div>
                  </div>

                  <div v-if="getModelResults(test).length" class="model-results-flow mt-2">
                    <v-tooltip
                      v-for="modelResult in getModelResults(test)"
                      :key="`${result.channelId}-${test.protocol}-${modelResult.model}`"
                      location="top"
                      content-class="ccx-tooltip"
                    >
                      <template #activator="{ props: tooltipProps }">
                        <div
                          v-bind="tooltipProps"
                          :class="['model-result-badge', getModelBadgeClass(modelResult.status)]"
                        >
                          <span class="model-name">{{ modelResult.model }}</span>
                          <v-icon size="16">{{ getModelStatusIcon(modelResult.status) }}</v-icon>
                        </div>
                      </template>
                      <div class="tooltip-content">
                        <div class="tooltip-title">{{ modelResult.model }}</div>
                        <div class="tooltip-row">
                          <span class="tooltip-label">{{ t('capability.modelStatus') }}</span>
                          <span class="tooltip-value">{{ getModelStatusLabel(modelResult.status) }}</span>
                        </div>
                        <div v-if="modelResult.success" class="tooltip-row">
                          <span class="tooltip-label">{{ t('capability.tooltipLatency') }}</span>
                          <span class="tooltip-value">{{ modelResult.latency }}ms</span>
                        </div>
                        <div v-if="modelResult.success" class="tooltip-row">
                          <span class="tooltip-label">{{ t('capability.tooltipStreaming') }}</span>
                          <span class="tooltip-value">{{ modelResult.streamingSupported ? t('capability.supported') : t('capability.unsupported') }}</span>
                        </div>
                        <div v-if="modelResult.error" class="tooltip-error">{{ modelResult.error }}</div>
                      </div>
                    </v-tooltip>
                  </div>
                </div>
              </div>
            </v-card-text>
          </v-card>
        </div>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { api, type CapabilityModelJobResult, type CapabilityModelJobStatus, type CapabilityProtocolJobResult, type CapabilityTestJob, type Channel, type PingResult } from '../services/api'
import { useI18n } from '../i18n'

type ChannelType = 'messages' | 'chat' | 'responses' | 'gemini'
type ProtocolType = 'messages' | 'chat' | 'responses' | 'gemini'

interface Props {
  modelValue: boolean
  channels: Channel[]
  channelType: ChannelType
}

interface BatchResult {
  channelId: number
  channelName: string
  latency: number | null
  capabilityStatus: CapabilityTestJob['status'] | 'idle'
  compatibleProtocols: string[]
  progressText: string
  error: string
  job: CapabilityTestJob | null
  updatedAt: number
}

const props = defineProps<Props>()
const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'latency-updated': []
}>()

const { t } = useI18n()

const protocolOrder: ProtocolType[] = ['messages', 'chat', 'responses', 'gemini']
const protocolOptions = computed(() => protocolOrder.map(value => ({ value, label: getProtocolDisplayName(value) })))

const selectedIds = ref<number[]>([])
const selectedProtocols = ref<ProtocolType[]>([...protocolOrder])
const selectedModels = ref<string[]>([])
const availableModels = ref<string[]>([])
const fetchingModels = ref(false)
const runningLatency = ref(false)
const runningCapability = ref(false)
const cancellingAll = ref(false)
const cancellingChannelIds = ref<number[]>([])
const errorMessage = ref('')
const results = ref<Record<number, BatchResult>>({})
const activeJobIds = ref<Record<number, string>>({})
let fetchModelsSeq = 0

const orderedResults = computed(() => Object.values(results.value).sort((a, b) => a.channelId - b.channelId))
const activeResultCount = computed(() => orderedResults.value.length)
const hasRunningJobs = computed(() => Object.keys(activeJobIds.value).length > 0)
const finishedCount = computed(() =>
  orderedResults.value.filter(result =>
    result.capabilityStatus === 'completed' ||
    result.capabilityStatus === 'failed' ||
    result.capabilityStatus === 'cancelled' ||
    result.capabilityStatus === 'idle'
  ).length
)

watch(() => props.modelValue, (open) => {
  if (open) {
    errorMessage.value = ''
    selectedIds.value = props.channels.map(channel => channel.index)
    selectedProtocols.value = [...protocolOrder]
    void fetchModelsForSelection()
  }
})

watch(() => [...selectedIds.value].sort((a, b) => a - b).join(','), () => {
  if (props.modelValue) {
    void fetchModelsForSelection()
  }
})

const selectAll = () => {
  selectedIds.value = props.channels.map(channel => channel.index)
}

const clearSelection = () => {
  selectedIds.value = []
}

const toggleChannel = (channelId: number) => {
  selectedIds.value = selectedIds.value.includes(channelId)
    ? selectedIds.value.filter(id => id !== channelId)
    : [...selectedIds.value, channelId]
}

const toggleProtocol = (protocol: ProtocolType) => {
  selectedProtocols.value = selectedProtocols.value.includes(protocol)
    ? selectedProtocols.value.filter(item => item !== protocol)
    : [...selectedProtocols.value, protocol].sort((a, b) => protocolOrder.indexOf(a) - protocolOrder.indexOf(b))
}

const ensureResult = (channel: Channel): BatchResult => {
  const existing = results.value[channel.index]
  if (existing) return existing

  const created: BatchResult = {
    channelId: channel.index,
    channelName: channel.name,
    latency: typeof channel.latency === 'number' ? channel.latency : null,
    capabilityStatus: 'idle',
    compatibleProtocols: [],
    progressText: '',
    error: '',
    job: null,
    updatedAt: Date.now()
  }
  results.value = { ...results.value, [channel.index]: created }
  return created
}

const updateResult = (channelId: number, patch: Partial<BatchResult>) => {
  const channel = props.channels.find(item => item.index === channelId)
  if (!channel) return
  const current = ensureResult(channel)
  results.value = {
    ...results.value,
    [channelId]: {
      ...current,
      ...patch,
      updatedAt: Date.now()
    }
  }
}

const getServiceLabel = (serviceType: Channel['serviceType']) => {
  switch (serviceType) {
    case 'claude': return 'Claude'
    case 'gemini': return 'Gemini'
    case 'responses': return 'Responses'
    default: return 'OpenAI'
  }
}

const getProtocolDisplayName = (protocol: string) => {
  const map: Record<string, string> = {
    messages: 'Claude',
    chat: 'OpenAI Chat',
    gemini: 'Gemini',
    responses: 'Codex'
  }
  return map[protocol] || protocol
}

const getProtocolColor = (protocol: string) => {
  const map: Record<string, string> = {
    messages: 'orange',
    chat: 'primary',
    gemini: 'deep-purple',
    responses: 'teal'
  }
  return map[protocol] || 'grey'
}

const getLatencyLabel = (channelId: number) => {
  const result = results.value[channelId]
  if (typeof result?.latency === 'number') {
    return `${result.latency}ms`
  }
  const channel = props.channels.find(item => item.index === channelId)
  return typeof channel?.latency === 'number' ? `${channel.latency}ms` : t('batchTest.notTested')
}

const getCapabilitySummary = (channelId: number) => {
  const result = results.value[channelId]
  if (!result || result.capabilityStatus === 'idle') return t('batchTest.notTested')
  if (result.compatibleProtocols.length > 0) return result.compatibleProtocols.map(getProtocolDisplayName).join(' / ')
  if (result.progressText) return result.progressText
  if (result.error) return t('capability.failed')
  return getResultStatus(result)
}

const getResultColor = (result: BatchResult) => {
  if (result.capabilityStatus === 'completed') return 'success'
  if (result.capabilityStatus === 'failed' || result.capabilityStatus === 'cancelled') return 'error'
  if (result.capabilityStatus === 'running' || result.capabilityStatus === 'queued') return 'info'
  return 'grey'
}

const getResultStatus = (result: BatchResult) => {
  switch (result.capabilityStatus) {
    case 'queued': return t('capability.modelQueued')
    case 'running': return t('capability.protocolRunning')
    case 'completed': return t('capability.success')
    case 'failed': return t('capability.failed')
    case 'cancelled': return t('capability.cancelled')
    default: return t('batchTest.idle')
  }
}

const getProtocolStatusLabel = (status: CapabilityTestJob['status'] | CapabilityProtocolJobResult['status']) => {
  switch (status) {
    case 'queued': return t('capability.modelQueued')
    case 'running': return t('capability.protocolRunning')
    case 'completed': return t('capability.success')
    case 'failed': return t('capability.failed')
    case 'cancelled': return t('capability.cancelled')
    default: return status
  }
}

const getProtocolStatusClass = (status: CapabilityProtocolJobResult['status']) => {
  if (status === 'completed') return 'text-success'
  if (status === 'failed') return 'text-error'
  if (status === 'running') return 'text-info'
  return 'text-medium-emphasis'
}

const requestPing = async (channelId: number): Promise<PingResult> => {
  switch (props.channelType) {
    case 'chat':
      return api.pingChatChannel(channelId)
    case 'gemini':
      return api.pingGeminiChannel(channelId)
    case 'responses':
      return api.pingResponsesChannel(channelId)
    default:
      return api.pingChannel(channelId)
  }
}

const fetchModelsForSelection = async () => {
  const seq = ++fetchModelsSeq
  const selectedChannels = props.channels.filter(channel => selectedIds.value.includes(channel.index))
  if (selectedChannels.length === 0) {
    availableModels.value = []
    selectedModels.value = []
    return
  }

  fetchingModels.value = true
  const modelSet = new Set<string>()
  try {
    await Promise.all(selectedChannels.map(async channel => {
      channel.supportedModels?.forEach(model => modelSet.add(model))
      const key = channel.apiKeys?.[0]
      if (!key) return

      try {
        let response: { data?: Array<{ id: string }> } | null = null
        switch (props.channelType) {
          case 'chat':
            response = await api.getChatChannelModels(channel.index, key, channel.baseUrl)
            break
          case 'gemini':
            response = await api.getGeminiChannelModels(channel.index, key, channel.baseUrl)
            break
          case 'responses':
            response = await api.getResponsesChannelModels(channel.index, key, channel.baseUrl)
            break
          default:
            response = await api.getChannelModels(channel.index, key, channel.baseUrl)
            break
        }
        response?.data?.forEach(model => modelSet.add(model.id))
      } catch {
        // ignore per-channel model fetch error and fallback to supportedModels/manual input
      }
    }))
  } finally {
    if (seq === fetchModelsSeq) {
      const models = Array.from(modelSet).sort((a, b) => a.localeCompare(b))
      availableModels.value = models
      selectedModels.value = selectedModels.value.filter(model => modelSet.has(model))
      fetchingModels.value = false
    }
  }
}

const applyCapabilityJob = (channel: Channel, job: CapabilityTestJob) => {
  const completed = job.progress?.completedModels ?? 0
  const total = job.progress?.totalModels ?? 0
  updateResult(channel.index, {
    capabilityStatus: job.status,
    compatibleProtocols: job.compatibleProtocols ?? [],
    progressText: total > 0 ? t('capability.progressSummary', { done: completed, total }) : '',
    error: job.error ?? '',
    job
  })
}

const waitForCapabilityJob = async (channel: Channel, jobId: string) => {
  while (true) {
    if (activeJobIds.value[channel.index] !== jobId) {
      return null
    }
    const latest = await api.getChannelCapabilityTestStatus(props.channelType, channel.index, jobId)
    applyCapabilityJob(channel, latest)
    if (latest.status === 'completed' || latest.status === 'failed' || latest.status === 'cancelled') {
      const next = { ...activeJobIds.value }
      delete next[channel.index]
      activeJobIds.value = next
      return latest
    }
    await new Promise(resolve => window.setTimeout(resolve, 1000))
  }
}

const isJobRunning = (result: BatchResult) => {
  return result.capabilityStatus === 'queued' || result.capabilityStatus === 'running'
}

const cancelSingleJob = async (channelId: number) => {
  const jobId = activeJobIds.value[channelId]
  if (!jobId) return
  cancellingChannelIds.value = [...cancellingChannelIds.value, channelId]
  try {
    await api.cancelCapabilityTest(props.channelType, channelId, jobId)
    const latest = await api.getChannelCapabilityTestStatus(props.channelType, channelId, jobId)
    const channel = props.channels.find(item => item.index === channelId)
    if (channel) {
      applyCapabilityJob(channel, latest)
    }
  } finally {
    const next = { ...activeJobIds.value }
    delete next[channelId]
    activeJobIds.value = next
    cancellingChannelIds.value = cancellingChannelIds.value.filter(id => id !== channelId)
  }
}

const cancelAllJobs = async () => {
  cancellingAll.value = true
  try {
    await Promise.all(Object.entries(activeJobIds.value).map(async ([channelIdStr, jobId]) => {
      const channelId = Number(channelIdStr)
      await cancelSingleJob(channelId)
    }))
  } finally {
    cancellingAll.value = false
  }
}

const runLatencyTests = async () => {
  errorMessage.value = ''
  runningLatency.value = true
  try {
    const targets = props.channels.filter(channel => selectedIds.value.includes(channel.index))
    await Promise.all(targets.map(async channel => {
      const result = await requestPing(channel.index)
      updateResult(channel.index, {
        channelName: channel.name,
        latency: result.latency,
        error: result.success ? '' : (result.error || '')
      })
    }))
    emit('latency-updated')
  } catch (error) {
    errorMessage.value = t('toast.batchLatencyFailed', { message: error instanceof Error ? error.message : t('system.unknown') })
  } finally {
    runningLatency.value = false
  }
}

const runCapabilityTests = async () => {
  errorMessage.value = ''
  if (selectedProtocols.value.length === 0) {
    errorMessage.value = t('batchTest.noProtocolSelected')
    return
  }
  runningCapability.value = true
  try {
    activeJobIds.value = {}
    const targets = props.channels.filter(channel => selectedIds.value.includes(channel.index))
    await Promise.all(targets.map(async channel => {
      updateResult(channel.index, {
        channelName: channel.name,
        capabilityStatus: 'queued',
        compatibleProtocols: [],
        progressText: t('batchTest.creatingJob'),
        error: '',
        job: null
      })
      const started = await api.startChannelCapabilityTest(
        props.channelType,
        channel.index,
        undefined,
        selectedModels.value.length > 0 ? selectedModels.value : undefined,
        selectedProtocols.value
      )
      activeJobIds.value = { ...activeJobIds.value, [channel.index]: started.jobId }
      if (started.job) {
        applyCapabilityJob(channel, started.job)
        if (started.job.status === 'completed' || started.job.status === 'failed' || started.job.status === 'cancelled') {
          const next = { ...activeJobIds.value }
          delete next[channel.index]
          activeJobIds.value = next
          return
        }
      }
      await waitForCapabilityJob(channel, started.jobId)
    }))
  } catch (error) {
    errorMessage.value = t('batchTest.batchCapabilityFailed', { message: error instanceof Error ? error.message : t('system.unknown') })
  } finally {
    runningCapability.value = false
  }
}

const sortedTests = (tests: CapabilityProtocolJobResult[]) => {
  return [...tests].sort((a, b) => protocolOrder.indexOf(a.protocol as ProtocolType) - protocolOrder.indexOf(b.protocol as ProtocolType))
}

const getModelResults = (test: CapabilityProtocolJobResult): CapabilityModelJobResult[] => {
  return Array.isArray(test.modelResults) ? test.modelResults : []
}

const getSuccessCount = (test: CapabilityProtocolJobResult): number => {
  if (typeof test.successCount === 'number') return test.successCount
  return getModelResults(test).filter(item => item.success).length
}

const getAttemptedModels = (test: CapabilityProtocolJobResult): number => {
  if (typeof test.attemptedModels === 'number') return test.attemptedModels
  return getModelResults(test).length
}

const formatSuccessRatio = (test: CapabilityProtocolJobResult): string => {
  const attempted = getAttemptedModels(test)
  return attempted > 0 ? `${getSuccessCount(test)}/${attempted}` : '-'
}

const getAverageLatency = (test: CapabilityProtocolJobResult): number => {
  const successModels = getModelResults(test).filter(item => item.success && typeof item.latency === 'number' && item.latency >= 0)
  if (successModels.length === 0) return -1
  return Math.round(successModels.reduce((sum, item) => sum + item.latency, 0) / successModels.length)
}

const hasProtocolLatency = (test: CapabilityProtocolJobResult): boolean => getAverageLatency(test) >= 0

const getModelBadgeClass = (status: CapabilityModelJobStatus) => {
  switch (status) {
    case 'success': return 'success-badge'
    case 'failed': return 'error-badge'
    case 'running': return 'running-badge'
    case 'skipped': return 'skipped-badge'
    default: return 'queued-badge'
  }
}

const getModelStatusIcon = (status: CapabilityModelJobStatus) => {
  switch (status) {
    case 'queued': return 'mdi-timer-sand'
    case 'running': return 'mdi-progress-clock'
    case 'skipped': return 'mdi-skip-next'
    case 'success': return 'mdi-check-circle'
    default: return 'mdi-close-circle'
  }
}

const getModelStatusLabel = (status: CapabilityModelJobStatus) => {
  switch (status) {
    case 'queued': return t('capability.modelQueued')
    case 'running': return t('capability.modelRunning')
    case 'success': return t('capability.modelSuccess')
    case 'failed': return t('capability.modelFailed')
    case 'skipped': return t('capability.modelSkipped')
    default: return status
  }
}
</script>

<style scoped>
.control-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.control-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.section-title {
  font-size: 0.875rem;
  font-weight: 700;
}

.channel-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 12px;
}

.channel-select-card {
  cursor: pointer;
  transition: transform 0.18s ease, box-shadow 0.18s ease;
}

.channel-select-card:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 18px rgba(15, 23, 42, 0.08);
}

.result-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.compact-result-list {
  gap: 10px;
}

.protocol-result-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.compact-protocol-list {
  gap: 8px;
}

.protocol-result-card {
  padding: 12px;
  border-radius: 12px;
  background: rgba(var(--v-theme-surface-variant), 0.12);
  border: 1px solid rgba(var(--v-theme-outline), 0.14);
}

.compact-protocol-card {
  padding: 10px 12px;
  border-radius: 10px;
}

.result-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  flex-wrap: wrap;
}

.protocol-result-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  flex-wrap: wrap;
}

.model-results-flow {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.model-result-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  border-radius: 8px;
  border: 1px solid transparent;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Courier New', monospace;
}

.success-badge {
  background: #f0fdf4;
  color: #16a34a;
  border-color: #dcfce7;
}

.error-badge {
  background: #fef2f2;
  color: #dc2626;
  border-color: #fee2e2;
}

.running-badge {
  background: rgba(var(--v-theme-info), 0.12);
  color: rgb(var(--v-theme-info));
  border-color: rgba(var(--v-theme-info), 0.2);
}

.skipped-badge {
  background: rgba(var(--v-theme-surface-variant), 0.4);
  color: rgba(var(--v-theme-on-surface), 0.55);
  border-color: rgba(var(--v-theme-outline), 0.18);
}

.queued-badge {
  background: rgba(var(--v-theme-surface-variant), 0.6);
  color: rgb(var(--v-theme-on-surface));
  border-color: rgba(var(--v-theme-outline), 0.18);
}

.model-name {
  font-size: 0.8125rem;
}

.tooltip-content {
  padding: 4px 0;
  min-width: 220px;
}

.tooltip-title {
  font-weight: 600;
  margin-bottom: 6px;
}

.tooltip-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  font-size: 0.8125rem;
  margin: 6px 0;
}

.tooltip-label {
  opacity: 0.72;
}

.tooltip-value {
  font-weight: 600;
}

.tooltip-error {
  font-size: 0.8125rem;
  color: inherit;
  margin-top: 4px;
  max-width: 320px;
  word-break: break-word;
}

.min-w-0 {
  min-width: 0;
}
</style>
