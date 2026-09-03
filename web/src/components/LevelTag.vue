<script setup lang="ts">
import { computed } from 'vue'
import { NTag } from 'naive-ui'

import type { components } from '@/api/schema'

type HazardLevel = components['schemas']['HazardLevel']

const props = defineProps<{ level: HazardLevel }>()

const config = computed<{ type: 'info' | 'error'; label: string }>(() => {
  if (props.level === '重大隐患') {
    return { type: 'error', label: '重大隐患' }
  }
  return { type: 'info', label: '一般隐患' }
})
</script>

<template>
  <n-tag :type="config.type" size="small" :bordered="false">{{ config.label }}</n-tag>
</template>