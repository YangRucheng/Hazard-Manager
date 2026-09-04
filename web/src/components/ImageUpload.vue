<script setup lang="ts">
import { ref } from 'vue'
import {
  NUpload,
  NUploadDragger,
  NImage,
  NButton,
  NIcon,
  useMessage,
  type UploadFileInfo,
  type UploadCustomRequestOptions,
} from 'naive-ui'
import { CloseOutline, CloudUploadOutline } from '@vicons/ionicons5'

import { thumbnailUrl, imageUrl } from '@/api/client'
import { uploadImage } from '@/api/upload'

const props = withDefaults(
  defineProps<{
    /** 已选图片 uuid 列表（v-model） */
    modelValue: string[]
    /** 上限数量 */
    max?: number
    /** 上传按钮文字 */
    placeholder?: string
  }>(),
  { max: 20, placeholder: '点击或拖拽上传图片' },
)

const emit = defineEmits<{ 'update:modelValue': [value: string[]] }>()

const message = useMessage()
const pendingCount = ref(0)

const remaining = (): number => props.max - props.modelValue.length

async function handleCustomRequest(options: UploadCustomRequestOptions): Promise<void> {
  const raw = options.file.file
  if (!raw) {
    message.error('无法读取文件')
    options.onError()
    return
  }
  pendingCount.value += 1
  try {
    const info = await uploadImage(raw)
    if (info && remaining() > 0) {
      emit('update:modelValue', [...props.modelValue, info.id])
      options.onFinish()
    } else {
      message.warning('已达到图片数量上限')
      options.onError()
    }
  } catch (err) {
    message.error(err instanceof Error ? err.message : '上传失败')
    options.onError()
  } finally {
    pendingCount.value -= 1
  }
}

function removeImage(id: string): void {
  emit('update:modelValue', props.modelValue.filter((v) => v !== id))
}

function handleBeforeUpload(data: { file: UploadFileInfo }): boolean {
  const raw = data.file.file
  if (raw && raw.size > 10 * 1024 * 1024) {
    message.error('图片不能超过 10MB')
    return false
  }
  return true
}
</script>

<template>
  <div>
    <div v-if="modelValue.length > 0" class="uploaded-list">
      <div v-for="id in modelValue" :key="id" class="uploaded-item">
        <n-image
          :src="thumbnailUrl(id)"
          :preview-src="imageUrl(id)"
          width="64"
          height="64"
          object-fit="cover"
        />
        <n-button
          class="remove-btn"
          size="tiny"
          quaternary
          circle
          @click="removeImage(id)"
        >
          <template #icon><n-icon><close-outline /></n-icon></template>
        </n-button>
      </div>
    </div>
    <n-upload
      :show-file-list="false"
      :disabled="pendingCount > 0 || remaining() <= 0"
      :max="props.max"
      :custom-request="handleCustomRequest"
      :default-upload="false"
      multiple
      accept="image/jpeg,image/png,image/webp,image/gif"
      :before-upload="handleBeforeUpload"
    >
      <n-upload-dragger class="upload-dragger">
        <div class="upload-inner">
          <n-icon size="28" color="#1668dc"><cloud-upload-outline /></n-icon>
          <span>{{ placeholder }}</span>
        </div>
      </n-upload-dragger>
    </n-upload>
  </div>
</template>

<style scoped>
.uploaded-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 10px;
}

.uploaded-item {
  position: relative;
  border-radius: 6px;
  overflow: hidden;
  border: 1px solid #dbe5f1;
}

.remove-btn {
  position: absolute;
  top: 2px;
  right: 2px;
  background: rgba(15, 23, 42, 0.55);
  color: #fff;
}

.upload-dragger {
  border-style: dashed;
  border-radius: 6px;
}

.upload-inner {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 8px 0;
  color: #17233d;
}
</style>