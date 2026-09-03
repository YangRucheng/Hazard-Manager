<script setup lang="ts">
import { onMounted, ref, h } from 'vue'
import { useRouter } from 'vue-router'
import { NCard, NGrid, NGridItem, NDataTable, useMessage, type DataTableColumns, NButton } from 'naive-ui'

import { client, errorMessage } from '@/api/client'
import StatusTag from '@/components/StatusTag.vue'
import LevelTag from '@/components/LevelTag.vue'
import ImagePreview from '@/components/ImagePreview.vue'
import { isOverdue } from '@/utils/date'

import type { components } from '@/api/schema'

type Hazard = components['schemas']['Hazard']
type HazardStats = components['schemas']['HazardStats']

const router = useRouter()
const message = useMessage()

const stats = ref<HazardStats>({ pending: 0, blocked: 0, done: 0, overdue: 0 })
const recent = ref<Hazard[]>([])
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
    const [statsRes, listRes] = await Promise.all([
      client.GET('/hazards/stats'),
      client.GET('/hazards', { params: { query: { page: 1, pageSize: 8 } } }),
    ])
    if (statsRes.error || !statsRes.data) {
      message.error(errorMessage(statsRes.error))
      return
    }
    stats.value = statsRes.data
    if (listRes.error || !listRes.data) {
      message.error(errorMessage(listRes.error))
      return
    }
    recent.value = listRes.data.items
  } finally {
    loading.value = false
  }
}

function editHazard(id: number): void {
  void router.push({ name: 'hazard-edit', params: { id: String(id) } })
}

const columns: DataTableColumns<Hazard> = [
  { title: '检查日期', key: 'inspectionDate', width: 110 },
  {
    title: '隐患描述',
    key: 'description',
    ellipsis: { tooltip: true },
    minWidth: 220,
  },
  { title: '责任单位', key: 'unitName', width: 120 },
  { title: '责任人', key: 'person', width: 90 },
  {
    title: '整改前图片',
    key: 'beforeImageIds',
    width: 150,
    render: (row) =>
      h(ImagePreview, { imageIds: row.beforeImageIds, size: 36 }),
  },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: (row) => {
      const overdue = isOverdue(row.dueDate, row.status)
      return h('div', { style: 'display:flex;align-items:center;gap:6px' }, [
        h(StatusTag, { status: row.status }),
        overdue ? h('span', { style: 'color:#d03050;font-size:12px' }, '逾期') : null,
      ])
    },
  },
  {
    title: '等级',
    key: 'level',
    width: 90,
    render: (row) => h(LevelTag, { level: row.level }),
  },
  {
    title: '操作',
    key: 'actions',
    width: 80,
    render: (row) => h(NButton, { size: 'small', text: true, type: 'primary', onClick: () => editHazard(row.id) }, { default: () => '查看' }),
  },
]

onMounted(loadData)
</script>

<template>
  <div class="page">
    <n-grid :x-gap="14" :y-gap="14" cols="1 s:2 m:4" responsive="screen">
      <n-grid-item v-for="card in statCards" :key="card.key">
        <n-card class="stat-card" :bordered="true">
          <div class="stat-label" :style="{ color: card.color }">{{ card.label }}</div>
          <div class="stat-value">{{ stats[card.key] }}</div>
        </n-card>
      </n-grid-item>
    </n-grid>

    <n-card title="最近隐患" style="margin-top: 14px" :bordered="true">
      <n-data-table
        :columns="columns"
        :data="recent"
        :loading="loading"
        :bordered="false"
        :row-key="(row: Hazard) => row.id"
        size="small"
      />
    </n-card>
  </div>
</template>

<style scoped>
.stat-card {
  text-align: left;
}

.stat-label {
  font-size: 13px;
  color: #6b7a90;
}

.stat-value {
  font-size: 30px;
  font-weight: 650;
  margin-top: 4px;
  color: #17233d;
}
</style>