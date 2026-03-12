<template>
  <v-dialog
    :model-value="modelValue"
    max-width="960"
    scrollable
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <v-card rounded="xl">
      <v-card-title class="d-flex align-center justify-space-between pa-4">
        <div class="d-flex align-center ga-2">
          <v-icon color="success">mdi-test-tube</v-icon>
          <span>{{ t('capability.title', { channel: channelName }) }}</span>
        </div>
        <v-btn icon variant="text" @click="$emit('update:modelValue', false)">
          <v-icon>mdi-close</v-icon>
        </v-btn>
      </v-card-title>

      <v-divider />

      <v-card-text class="pa-4">
        <div v-if="state === 'loading'" class="d-flex flex-column align-center py-8">
          <v-progress-circular indeterminate size="48" color="primary" />
          <p class="text-body-1 mt-4 text-medium-emphasis">{{ t('capability.loadingTitle') }}</p>
          <p class="text-caption text-medium-emphasis">{{ t('capability.loadingBody') }}</p>
        </div>

        <div v-else-if="state === 'error'" class="py-4">
          <v-alert type="error" variant="tonal" rounded="lg">
            {{ errorMessage }}
          </v-alert>
        </div>

        <div v-else-if="state === 'result' && result">
          <div class="mb-4">
            <div class="text-body-2 font-weight-medium mb-2">{{ t('capability.compatibleProtocols') }}</div>
            <div class="d-flex flex-wrap ga-2">
              <v-chip
                v-for="proto in result.compatibleProtocols"
                :key="proto"
                :color="getProtocolColor(proto)"
                size="small"
                variant="tonal"
              >
                <v-icon start size="small">{{ getProtocolIcon(proto) }}</v-icon>
                {{ getProtocolDisplayName(proto) }}
              </v-chip>
              <v-chip v-if="result.compatibleProtocols.length === 0" color="grey" size="small" variant="tonal">
                {{ t('capability.noCompatibleProtocols') }}
              </v-chip>
            </div>
          </div>

          <v-table density="comfortable" class="rounded-lg capability-table">
            <thead>
              <tr>
                <th>{{ t('capability.table.protocol') }}</th>
                <th>{{ t('capability.table.status') }}</th>
                <th>{{ t('capability.table.testModel') }}</th>
                <th>{{ t('capability.table.successCount') }}</th>
                <th>{{ t('capability.table.latency') }}</th>
                <th>{{ t('capability.table.streaming') }}</th>
                <th>{{ t('capability.table.actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <template v-for="test in sortedTests" :key="test.protocol">
                <tr>
                  <td>
                    <v-chip :color="getProtocolColor(test.protocol)" size="small" variant="tonal">
                      {{ getProtocolDisplayName(test.protocol) }}
                    </v-chip>
                  </td>
                  <td>
                    <div v-if="test.success" class="d-flex align-center ga-1">
                      <v-icon color="success" size="small">mdi-check-circle</v-icon>
                      <span class="text-body-2 text-success">{{ t('capability.success') }}</span>
                    </div>
                    <v-tooltip v-else :text="test.error || t('capability.failedTooltip')" location="top" content-class="error-tooltip">
                      <template #activator="{ props }">
                        <div v-bind="props" class="d-flex align-center ga-1">
                          <v-icon color="error" size="small">mdi-close-circle</v-icon>
                          <span class="text-body-2 text-error">{{ t('capability.failed') }}</span>
                        </div>
                      </template>
                    </v-tooltip>
                  </td>
                  <td>
                    <span class="text-body-2 text-medium-emphasis">{{ getRecommendedModel(test) }}</span>
                  </td>
                  <td>
                    <span class="text-body-2">{{ formatSuccessRatio(test) }}</span>
                  </td>
                  <td>
                    <span v-if="hasProtocolLatency(test)" class="text-body-2">{{ test.latency }}ms</span>
                    <span v-else class="text-body-2 text-medium-emphasis">-</span>
                  </td>
                  <td>
                    <div v-if="test.success && test.streamingSupported" class="d-flex align-center ga-1">
                      <v-icon color="success" size="small">mdi-check-circle</v-icon>
                      <span class="text-body-2 text-success">{{ t('capability.supported') }}</span>
                    </div>
                    <div v-else-if="test.success" class="d-flex align-center ga-1">
                      <v-icon color="warning" size="small">mdi-minus-circle</v-icon>
                      <span class="text-body-2 text-warning">{{ t('capability.unsupported') }}</span>
                    </div>
                    <span v-else class="text-body-2 text-medium-emphasis">-</span>
                  </td>
                  <td>
                    <v-btn
                      v-if="test.success && test.protocol !== currentTab"
                      size="x-small"
                      color="primary"
                      variant="tonal"
                      rounded="lg"
                      @click="$emit('copyToTab', test.protocol)"
                    >
                      {{ t('capability.copyToTab') }}
                    </v-btn>
                    <v-chip v-else-if="test.protocol === currentTab" size="x-small" color="grey" variant="tonal">
                      {{ t('capability.currentTab') }}
                    </v-chip>
                    <div v-else-if="!test.success && test.protocol !== currentTab" class="d-flex flex-wrap ga-1">
                      <v-btn
                        v-for="successProto in getSuccessfulProtocols()"
                        :key="successProto"
                        size="x-small"
                        :color="getProtocolColor(successProto)"
                        variant="outlined"
                        rounded="lg"
                        @click="$emit('copyToTab', test.protocol)"
                      >
                        {{ t('capability.convert', { protocol: getProtocolDisplayName(successProto) }) }}
                      </v-btn>
                    </div>
                  </td>
                </tr>
                <tr>
                  <td colspan="7" class="model-results-cell">
                    <div class="model-results-wrapper">
                      <div class="d-flex align-center justify-space-between flex-wrap ga-2 mb-2">
                        <div class="text-caption font-weight-medium text-medium-emphasis">
                          {{ t('capability.modelDetails') }}
                        </div>
                        <div class="text-caption text-medium-emphasis">
                          {{ t('capability.attemptedModels', { count: getAttemptedModels(test) }) }}
                        </div>
                      </div>

                      <div v-if="getModelResults(test).length === 0" class="text-body-2 text-medium-emphasis py-2">
                        {{ t('capability.modelDetailsUnavailable') }}
                      </div>

                      <div v-else class="model-results-list">
                        <div
                          v-for="modelResult in getModelResults(test)"
                          :key="`${test.protocol}-${modelResult.model}`"
                          class="model-result-item"
                        >
                          <div class="model-result-main">
                            <div class="d-flex align-center ga-2 flex-wrap">
                              <span class="text-body-2 font-weight-medium">{{ modelResult.model }}</span>
                              <v-chip
                                size="x-small"
                                :color="modelResult.success ? 'success' : 'error'"
                                variant="tonal"
                              >
                                {{ modelResult.success ? t('capability.success') : t('capability.failed') }}
                              </v-chip>
                            </div>
                            <div v-if="modelResult.error" class="text-caption text-error mt-1">
                              {{ modelResult.error }}
                            </div>
                          </div>
                          <div class="model-result-meta text-caption text-medium-emphasis">
                            <span>{{ formatLatency(modelResult.latency) }}</span>
                            <span>{{ formatStreaming(modelResult) }}</span>
                            <span v-if="modelResult.startedAt">{{ t('capability.startedAt') }}: {{ formatTime(modelResult.startedAt) }}</span>
                            <span v-if="modelResult.testedAt">{{ t('capability.testedAt') }}: {{ formatTime(modelResult.testedAt) }}</span>
                          </div>
                        </div>
                      </div>
                    </div>
                  </td>
                </tr>
              </template>
            </tbody>
          </v-table>

          <div class="text-caption text-medium-emphasis mt-3 text-right">
            {{ t('capability.totalDuration', { duration: result.totalDuration }) }}
          </div>
        </div>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { CapabilityTestResult, ModelTestResult, ProtocolTestResult } from '../services/api'
import { useI18n } from '../i18n'

interface Props {
  modelValue: boolean
  channelName: string
  currentTab: string
}

defineProps<Props>()
defineEmits<{
  'update:modelValue': [value: boolean]
  'copyToTab': [protocol: string]
}>()

const { t } = useI18n()

const state = ref<'loading' | 'error' | 'result'>('loading')
const result = ref<CapabilityTestResult | null>(null)
const errorMessage = ref('')

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

const getProtocolIcon = (protocol: string) => {
  const map: Record<string, string> = {
    messages: 'mdi-message-processing',
    chat: 'mdi-robot',
    gemini: 'mdi-diamond-stone',
    responses: 'mdi-code-braces'
  }
  return map[protocol] || 'mdi-api'
}

const getSuccessfulProtocols = () => {
  if (!result.value) return []
  return result.value.tests
    .filter(t => t.success)
    .map(t => t.protocol)
}

const protocolOrder = ['messages', 'chat', 'responses', 'gemini']

const sortedTests = computed(() => {
  if (!result.value) return []
  return [...result.value.tests].sort((a, b) => {
    const indexA = protocolOrder.indexOf(a.protocol)
    const indexB = protocolOrder.indexOf(b.protocol)
    return (indexA === -1 ? 999 : indexA) - (indexB === -1 ? 999 : indexB)
  })
})

const getModelResults = (test: ProtocolTestResult): ModelTestResult[] => {
  return Array.isArray(test.modelResults) ? test.modelResults : []
}

const getAttemptedModels = (test: ProtocolTestResult): number => {
  if (typeof test.attemptedModels === 'number') return test.attemptedModels
  const modelResults = getModelResults(test)
  return modelResults.length
}

const getSuccessCount = (test: ProtocolTestResult): number => {
  if (typeof test.successCount === 'number') return test.successCount
  return getModelResults(test).filter(modelResult => modelResult.success).length
}

const getRecommendedModel = (test: ProtocolTestResult): string => {
  if (test.testedModel) return test.testedModel
  const firstSuccessfulModel = getModelResults(test).find(modelResult => modelResult.success)
  if (firstSuccessfulModel?.model) return firstSuccessfulModel.model
  return '-'
}

const formatSuccessRatio = (test: ProtocolTestResult): string => {
  const attemptedModels = getAttemptedModels(test)
  if (attemptedModels <= 0) return '-'
  return `${getSuccessCount(test)}/${attemptedModels}`
}

const hasProtocolLatency = (test: ProtocolTestResult): boolean => {
  return typeof test.latency === 'number' && test.latency >= 0
}

const formatLatency = (latency: number): string => {
  return latency >= 0 ? `${latency}ms` : '-'
}

const formatStreaming = (modelResult: ModelTestResult): string => {
  if (!modelResult.success) return '-'
  return modelResult.streamingSupported ? t('capability.supported') : t('capability.unsupported')
}

const formatTime = (value: string): string => {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleTimeString()
}

const setLoading = () => {
  state.value = 'loading'
  result.value = null
  errorMessage.value = ''
}

const startTest = (testResult: CapabilityTestResult) => {
  result.value = testResult
  state.value = 'result'
}

const setError = (error: string) => {
  errorMessage.value = error
  state.value = 'error'
}

defineExpose({ startTest, setLoading, setError })
</script>

<style scoped>
:deep(.error-tooltip) {
  color: rgba(var(--v-theme-on-surface), 0.92);
  background-color: rgba(var(--v-theme-surface), 0.98);
  border: 1px solid rgba(var(--v-theme-error), 0.45);
  font-weight: 600;
  letter-spacing: 0.2px;
  max-width: 400px;
  word-break: break-word;
}

.capability-table :deep(th) {
  white-space: nowrap;
}

.model-results-cell {
  padding: 0 !important;
  background: rgba(var(--v-theme-surface-variant), 0.16);
}

.model-results-wrapper {
  padding: 12px 16px;
}

.model-results-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.model-result-item {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 12px;
  border-radius: 12px;
  background: rgba(var(--v-theme-surface), 0.92);
}

.model-result-main {
  min-width: 0;
  flex: 1;
}

.model-result-meta {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 12px;
  white-space: nowrap;
}

@media (max-width: 720px) {
  .model-result-item {
    flex-direction: column;
  }

  .model-result-meta {
    justify-content: flex-start;
    white-space: normal;
  }
}
</style>
