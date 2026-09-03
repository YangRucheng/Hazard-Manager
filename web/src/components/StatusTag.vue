<script setup lang="ts">
import { computed } from 'vue'
import { NTag } from 'naive-ui'

import type { components } from '@/api/schema'

type HazardStatus = components['schemas']['HazardStatus']

const props = defineProps<{ status: HazardStatus }>()

type TagType = 'warning' | 'error' | 'success' | 'default'

const config = computed<{ type: TagType; label: string }>(() => {
  switch (props.status) {
    case '待整改':
      return { type: 'warning', label: '待整改' }
    case '整改受阻':
      return { type: 'error', label: '整改受阻' }
    case '已整改':
      return { type: 'success', label: '已整改' }
    default:
      return { type: 'default', label: props.status }
  }
})
</script>

<template>
  <n-tag :type="config.type" size="small" round :bordered="false">{{ config.label }}</n-tag>
</template>