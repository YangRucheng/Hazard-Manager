<script setup lang="ts">
import { NEmpty, NImageGroup } from 'naive-ui'

import AuthImage from '@/components/AuthImage.vue'
import { thumbnailUrl, imageUrl } from '@/api/client'

const props = withDefaults(defineProps<{ imageIds: string[] | undefined | null; size?: number }>(), {
  size: 40,
})
</script>

<template>
  <n-empty
    v-if="!props.imageIds || props.imageIds.length === 0"
    size="small"
    description="无图片"
    style="padding: 2px 0"
  />
  <n-image-group v-else>
    <div class="image-preview" :style="{ gap: `${Math.max(4, Math.floor(props.size / 8))}px` }">
      <AuthImage
        v-for="id in props.imageIds"
        :key="id"
        :url="thumbnailUrl(id)"
        :preview-url="imageUrl(id)"
        :width="props.size"
        :height="props.size"
        class="preview-item"
      />
    </div>
  </n-image-group>
</template>

<style scoped>
.image-preview {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
}

.preview-item {
  border-radius: 4px;
  border: 1px solid var(--color-border);
  overflow: hidden;
}
</style>