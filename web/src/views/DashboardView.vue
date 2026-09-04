<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { NCard, NGrid, NGridItem, useMessage } from 'naive-ui'

import { client, errorMessage } from '@/api/client'

import type { components } from '@/api/schema'

type HazardStats = components['schemas']['HazardStats']

const message = useMessage()

const stats = ref<HazardStats>({ pending: 0, blocked: 0, done: 0, overdue: 0 })
const loading = ref(false)

const statCards = [
  { key: 'pending' as const, label: '待整改', color: '#e6a23c' },
  { key: 'blocked' as const, label: '整改受阻', color: '#d03050' },
  { key: 'done' as const, label: '已整改', color: '#18a058' },
  { key: 'overdue' as const, label: '逾期未整改', color: '#1668dc' },
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
    <n-grid :x-gap="14" :y-gap="14" cols="1 s:2 m:4" responsive="screen">
      <n-grid-item v-for="card in statCards" :key="card.key">
        <n-card class="stat-card" :bordered="true" :loading="loading">
          <div class="stat-label" :style="{ color: card.color }">{{ card.label }}</div>
          <div class="stat-value">{{ stats[card.key] }}</div>
        </n-card>
      </n-grid-item>
    </n-grid>
  </div>
</template>

<style scoped>
.stat-card {
  text-align: left;
}

.stat-label {
  font-size: 13px;
  font-weight: 600;
}

.stat-value {
  font-size: 30px;
  font-weight: 650;
  margin-top: 4px;
  color: #17233d;
}
</style>
