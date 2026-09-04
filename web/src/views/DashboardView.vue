<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { NButton, NCard, useMessage } from 'naive-ui'

import { client, errorMessage } from '@/api/client'

import type { components } from '@/api/schema'

type HazardStats = components['schemas']['HazardStats']

const message = useMessage()

const stats = ref<HazardStats>({ pending: 0, blocked: 0, done: 0, overdue: 0 })
const loading = ref(false)

const statCards = [
  { key: 'pending' as const, label: '待整改', hint: '已登记、待安排整改', color: '#d99020' },
  { key: 'blocked' as const, label: '整改受阻', hint: '受阻或超期未完成', color: '#d94b64' },
  { key: 'done' as const, label: '已整改', hint: '已验收完成闭环', color: '#229b6b' },
  { key: 'overdue' as const, label: '逾期未整改', hint: '超过要求完成时间', color: '#3f63d8' },
]

async function loadData(): Promise<void> {
  loading.value = true
  try {
    const { data, error } = await client.GET('/hazards/stats')
    if (error || !data) {
      message.error(errorMessage(error))
      return
    }
    stats.value = data
  } finally {
    loading.value = false
  }
}

onMounted(loadData)
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">工作台</h1>
      </div>
      <n-button :loading="loading" @click="loadData">刷新数据</n-button>
    </div>

    <div class="stat-grid">
      <n-card v-for="card in statCards" :key="card.key" :loading="loading">
        <div class="stat">
          <span class="stat-label">{{ card.label }}</span>
          <strong :style="{ color: card.color }">{{ stats[card.key] }}</strong>
          <span class="muted stat-hint">{{ card.hint }}</span>
        </div>
      </n-card>
    </div>
  </div>
</template>

<style scoped>
.stat {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.stat-label {
  color: var(--color-text);
  font-weight: 500;
}

.stat strong {
  font-size: 30px;
  font-weight: 650;
  line-height: 1.2;
}

.stat-hint {
  font-size: 12px;
}
</style>
