<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NCard, NTabs, NTabPane } from 'naive-ui'

const route = useRoute()
const router = useRouter()

const activeTab = computed<'units' | 'types' | 'attachments'>(() => {
  if (route.name === 'system-types') {
    return 'types'
  }
  if (route.name === 'system-attachments') {
    return 'attachments'
  }
  return 'units'
})

const tabRouteNames = {
  units: 'system-units',
  types: 'system-types',
  attachments: 'system-attachments',
} as const

function handleTabChange(value: string): void {
  const name = tabRouteNames[value as keyof typeof tabRouteNames] ?? 'system-units'
  void router.push({ name })
}
</script>

<template>
  <div class="page">
    <n-card :bordered="true" class="system-card">
      <n-tabs :value="activeTab" type="line" animated @update:value="handleTabChange">
        <n-tab-pane name="units" tab="责任单位" />
        <n-tab-pane name="types" tab="隐患类型" />
        <n-tab-pane name="attachments" tab="附件管理" />
      </n-tabs>
      <div class="tab-body">
        <router-view />
      </div>
    </n-card>
  </div>
</template>

<style scoped>
.system-card :deep(.n-tabs-nav) {
  box-shadow: none;
}

.tab-body {
  margin-top: 14px;
}
</style>
