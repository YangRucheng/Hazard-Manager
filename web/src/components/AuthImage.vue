<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { NImage } from 'naive-ui'

import { authedBlobUrl } from '@/api/client'

const props = withDefaults(
  defineProps<{
    /** 缩略图资源 URL（imageUrl / thumbnailUrl 生成） */
    url: string
    /** 原图资源 URL（点开放大预览用；缺省或加载失败时回退缩略图） */
    previewUrl?: string
    width?: number | string
    height?: number | string
    objectFit?: 'fill' | 'contain' | 'cover' | 'none' | 'scale-down'
  }>(),
  { previewUrl: undefined, width: undefined, height: undefined, objectFit: 'cover' },
)

const src = ref<string | null>(null)
const previewSrc = ref<string | null>(null)

/** 数值宽高转 px，供根元素占位布局用。 */
function cssSize(v: number | string | undefined): string | undefined {
  if (v === undefined) {
    return undefined
  }
  return typeof v === 'number' ? `${v}px` : v
}

const rootStyle = computed(() => ({
  width: cssSize(props.width),
  height: cssSize(props.height),
}))

watch(
  () => [props.url, props.previewUrl] as const,
  ([url, previewUrl]) => {
    src.value = null
    previewSrc.value = null
    void authedBlobUrl(url)
      .then((u) => {
        if (props.url === url) {
          src.value = u
        }
      })
      .catch(() => {
        // 加载失败：保持占位块（鉴权失效等场景由列表接口的 401 中间件统一处理跳转）。
      })
    if (previewUrl) {
      void authedBlobUrl(previewUrl)
        .then((u) => {
          if (props.previewUrl === previewUrl) {
            previewSrc.value = u
          }
        })
        .catch(() => {
          // 原图失败时 preview-src 回退缩略图。
        })
    }
  },
  { immediate: true },
)
</script>

<template>
  <div class="auth-image" :style="rootStyle">
    <n-image
      v-if="src"
      :src="src"
      :preview-src="previewSrc ?? src"
      :width="width"
      :height="height"
      :object-fit="objectFit"
    />
    <div v-else class="auth-image-placeholder" />
  </div>
</template>

<style scoped>
.auth-image {
  display: inline-block;
  line-height: 0;
}

.auth-image-placeholder {
  width: 100%;
  height: 100%;
  background: #f0f4f8;
}
</style>
