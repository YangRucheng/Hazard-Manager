<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NCard, NTabs, NTabPane } from 'naive-ui'

const route = useRoute()
const router = useRouter()

const activeTab = computed<'units' | 'types'>(() =>
  route.name === 'enums-types' ? 'types' : 'units',
)

function handleTabChange(value: string): void {
  void router.push(value === 'types' ? { name: 'enums-types' } : { name: 'enums-units' })
}
</script>

<template>
  <div class="page">
    <n-card :bordered="true">
      <n-tabs :value="activeTab" type="line" animated @update:value="handleTabChange">
        <n-tab-pane name="units" tab="责任单位" />
        <n-tab-pane name="types" tab="隐患类型" />
      </n-tabs>
      <router-view />
    </n-card>
  </div>
</template>