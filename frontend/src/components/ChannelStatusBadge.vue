<template>
  <div class="status-badge" :class="[statusClass, { 'has-metrics': showMetrics }]">
    <v-tooltip location="top" content-class="ccx-tooltip">
      <template #activator="{ props: tooltipProps }">
        <div class="badge-content" v-bind="tooltipProps">
          <v-icon :size="iconSize" class="status-icon">{{ statusIcon }}</v-icon>
          <span v-if="showLabel" class="status-label">{{ statusLabel }}</span>
        </div>
      </template>
      <div class="tooltip-content">
        <div class="font-weight-bold mb-1">{{ statusLabel }}</div>
        <template v-if="metrics">
          <div class="text-caption">
            <div>{{ t('status.metrics.requests') }}: {{ metrics.requestCount }}</div>
            <div>{{ t('status.metrics.successRate') }}: {{ metrics.successRate?.toFixed(1) || 0 }}%</div>
            <div>{{ t('status.metrics.consecutiveFailures') }}: {{ metrics.consecutiveFailures }}</div>
            <div v-if="metrics.circuitState">{{ t('status.metrics.circuitState') }}: {{ t(`status.circuit.${metrics.circuitState}` as Parameters<typeof t>[0]) }}</div>
            <div v-if="metrics.breakerFailureRate !== undefined">{{ t('status.metrics.breakerFailureRate') }}: {{ metrics.breakerFailureRate?.toFixed(1) || 0 }}%</div>
            <div v-if="metrics.halfOpenSuccesses">{{ t('status.metrics.halfOpenSuccesses') }}: {{ metrics.halfOpenSuccesses }}</div>
            <div v-if="metrics.nextRetryAt">{{ t('status.metrics.nextRetry') }}: {{ formatTime(metrics.nextRetryAt) }}</div>
            <div v-if="metrics.lastSuccessAt">{{ t('status.metrics.lastSuccess') }}: {{ formatTime(metrics.lastSuccessAt) }}</div>
            <div v-if="metrics.lastFailureAt">{{ t('status.metrics.lastFailure') }}: {{ formatTime(metrics.lastFailureAt) }}</div>
          </div>
        </template>
        <div v-else class="text-caption text-medium-emphasis">{{ t('status.metrics.noData') }}</div>
      </div>
    </v-tooltip>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ChannelStatus, ChannelMetrics } from '../services/api'
import { useI18n } from '../i18n'

const props = withDefaults(defineProps<{
  status: ChannelStatus | 'healthy' | 'error' | 'unknown'
  metrics?: ChannelMetrics
  showLabel?: boolean
  size?: 'small' | 'default' | 'large'
}>(), {
  showLabel: true,
  size: 'default'
})

const { t } = useI18n()

const effectiveStatus = computed(() => {
  if (props.status === 'disabled') return 'disabled'
  if (props.status === 'suspended') return 'suspended'
  if (props.metrics?.circuitState === 'open') return 'breaker-open'
  if (props.metrics?.circuitState === 'half_open') return 'breaker-half-open'
  return props.status
})

// 状态配置映射
const STATUS_CONFIG: Record<string, { icon: string; color: string; label: string; class: string }> = {
  active: {
    icon: 'mdi-check-circle',
    color: 'success',
    label: 'status.active',
    class: 'status-active'
  },
  healthy: {
    icon: 'mdi-check-circle',
    color: 'success',
    label: 'status.healthy',
    class: 'status-active'
  },
  suspended: {
    icon: 'mdi-pause-circle',
    color: 'warning',
    label: 'status.suspended',
    class: 'status-suspended'
  },
  disabled: {
    icon: 'mdi-close-circle',
    color: 'error',
    label: 'status.disabled',
    class: 'status-disabled'
  },
  'breaker-open': {
    icon: 'mdi-alert-octagon',
    color: 'error',
    label: 'status.breakerOpen',
    class: 'status-breaker-open'
  },
  'breaker-half-open': {
    icon: 'mdi-progress-clock',
    color: 'info',
    label: 'status.breakerHalfOpen',
    class: 'status-breaker-half-open'
  },
  error: {
    icon: 'mdi-alert-circle',
    color: 'error',
    label: 'status.error',
    class: 'status-error'
  },
  unknown: {
    icon: 'mdi-help-circle',
    color: 'grey',
    label: 'status.unknown',
    class: 'status-unknown'
  }
}

// 计算属性
const statusConfig = computed(() => {
  return STATUS_CONFIG[effectiveStatus.value] || STATUS_CONFIG.unknown
})

const statusIcon = computed(() => statusConfig.value.icon)
const statusLabel = computed(() => t(statusConfig.value.label as Parameters<typeof t>[0]))
const statusClass = computed(() => statusConfig.value.class)

const iconSize = computed(() => {
  switch (props.size) {
    case 'small': return 16
    case 'large': return 24
    default: return 20
  }
})

const showMetrics = computed(() => !!props.metrics)

// 格式化时间
const formatTime = (dateStr: string): string => {
  const date = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()

  if (diff < 60000) {
    return t('status.metrics.justNow')
  } else if (diff < 3600000) {
    return t('status.metrics.minutesAgo', { count: Math.floor(diff / 60000) })
  } else if (diff < 86400000) {
    return t('status.metrics.hoursAgo', { count: Math.floor(diff / 3600000) })
  } else {
    return date.toLocaleDateString()
  }
}
</script>

<style scoped>
/* =====================================================
   🎮 状态徽章 - 复古像素主题样式
   Neo-Brutalism: 直角、实体边框、高对比度
   ===================================================== */

.status-badge {
  display: inline-flex;
  align-items: center;
  position: relative;
}

.badge-content {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  padding: 5px 9px;
  background: rgb(var(--v-theme-surface));
  border: 1px solid rgb(var(--v-theme-on-surface));
  cursor: help;
  transition: all 0.1s ease;
  line-height: 1;
}

.badge-content :deep(.v-icon) {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  align-self: center;
  line-height: 1;
}

.v-theme--dark .badge-content {
  border-color: rgba(255, 255, 255, 0.6);
}

.badge-content:hover {
  background: rgba(var(--v-theme-surface-variant), 0.8);
}

.status-label {
  display: inline-flex;
  align-items: center;
  line-height: 1;
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0;
}

/* 状态样式 - 高对比度实心边框 */
.status-active .badge-content {
  background: var(--ccx-status-active-bg);
  color: var(--ccx-status-active-fg);
  border-color: var(--ccx-status-active-fg);
}

.status-active .badge-content .status-icon {
  color: var(--ccx-status-active-fg) !important;
}

.status-suspended .badge-content {
  background: var(--ccx-status-suspended-bg);
  color: var(--ccx-status-suspended-fg);
  border-color: var(--ccx-status-suspended-fg);
  animation: pixel-blink 1.5s step-end infinite;
}

.status-suspended .badge-content .status-icon {
  color: var(--ccx-status-suspended-fg) !important;
}

.status-disabled .badge-content {
  background: var(--ccx-status-disabled-bg);
  color: var(--ccx-status-disabled-fg);
  border-color: var(--ccx-status-disabled-fg);
}

.status-disabled .badge-content .status-icon {
  color: var(--ccx-status-disabled-fg) !important;
}

.status-breaker-open .badge-content {
  background: var(--ccx-status-breaker-open-bg);
  color: var(--ccx-status-breaker-open-fg);
  border-color: var(--ccx-status-breaker-open-fg);
  animation: pixel-blink 1.2s step-end infinite;
}

.status-breaker-open .badge-content .status-icon {
  color: var(--ccx-status-breaker-open-fg) !important;
}

.status-breaker-half-open .badge-content {
  background: var(--ccx-status-breaker-half-open-bg);
  color: var(--ccx-status-breaker-half-open-fg);
  border-color: var(--ccx-status-breaker-half-open-fg);
}

.status-breaker-half-open .badge-content .status-icon {
  color: var(--ccx-status-breaker-half-open-fg) !important;
}

.status-error .badge-content {
  background: var(--ccx-status-error-bg);
  color: var(--ccx-status-error-fg);
  border-color: var(--ccx-status-error-fg);
}

.status-error .badge-content .status-icon {
  color: var(--ccx-status-error-fg) !important;
}

.status-unknown .badge-content {
  background: var(--ccx-status-unknown-bg);
  color: var(--ccx-status-unknown-fg);
  border-color: var(--ccx-status-unknown-fg);
}

.status-unknown .badge-content .status-icon {
  color: var(--ccx-status-unknown-fg) !important;
}

/* 手机端隐藏状态文字，改为像素点样式 */
@media (max-width: 600px) {
  .status-label {
    display: none;
  }

  .badge-content {
    padding: 0;
    background: transparent !important;
    border: none !important;
  }

  .badge-content .v-icon {
    font-size: 0 !important;
    width: 10px;
    height: 10px;
    margin-right: 10px;
    position: relative;
  }

  /* 活跃状态 - 绿色像素点 */
  .status-active .badge-content .v-icon {
    background: var(--ccx-status-active-dot-bg);
    border: 2px solid var(--ccx-status-active-dot-border);
  }

  .status-active .badge-content .v-icon::after {
    content: '';
    position: absolute;
    top: -3px;
    left: -3px;
    width: 14px;
    height: 14px;
    background: var(--ccx-status-active-dot-glow);
    animation: pixel-pulse 1s step-end infinite;
  }

  /* 暂停状态 - 橙色像素点 */
  .status-suspended .badge-content .v-icon {
    background: var(--ccx-status-suspended-dot-bg);
    border: 2px solid var(--ccx-status-suspended-dot-border);
  }

  .status-suspended .badge-content .v-icon::after {
    content: '';
    position: absolute;
    top: -3px;
    left: -3px;
    width: 14px;
    height: 14px;
    background: var(--ccx-status-suspended-dot-glow);
    animation: pixel-pulse 0.75s step-end infinite;
  }

  /* 自动熔断状态 - 红色像素点 */
  .status-breaker-open .badge-content .v-icon {
    background: var(--ccx-status-breaker-open-dot-bg);
    border: 2px solid var(--ccx-status-breaker-open-dot-border);
  }

  .status-breaker-open .badge-content .v-icon::after {
    content: '';
    position: absolute;
    top: -3px;
    left: -3px;
    width: 14px;
    height: 14px;
    background: var(--ccx-status-breaker-open-dot-glow);
    animation: pixel-pulse 0.75s step-end infinite;
  }

  /* half-open 状态 - 蓝色像素点 */
  .status-breaker-half-open .badge-content .v-icon {
    background: var(--ccx-status-breaker-half-open-dot-bg);
    border: 2px solid var(--ccx-status-breaker-half-open-dot-border);
  }

  .status-breaker-half-open .badge-content .v-icon::after {
    content: '';
    position: absolute;
    top: -3px;
    left: -3px;
    width: 14px;
    height: 14px;
    background: var(--ccx-status-breaker-half-open-dot-glow);
    animation: pixel-pulse 0.85s step-end infinite;
  }

  /* 禁用状态 - 灰色像素点 */
  .status-disabled .badge-content .v-icon,
  .status-unknown .badge-content .v-icon {
    background: var(--ccx-status-disabled-dot-bg);
    border: 2px solid var(--ccx-status-disabled-dot-border);
  }

  @keyframes pixel-pulse {
    0%, 100% {
      opacity: 1;
    }
    50% {
      opacity: 0.4;
    }
  }
}

/* 像素风格闪烁动画 */
@keyframes pixel-blink {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.6;
  }
}

.tooltip-content {
  max-width: 200px;
}
</style>

<!-- 非 scoped 样式 - 用于 teleport 到 body 的 tooltip -->
<style>
</style>
