<script setup lang="ts">
import { h, onMounted, ref } from 'vue'
import {
  NButton,
  NDataTable,
  NImage,
  NPagination,
  NSwitch,
  NTag,
  useDialog,
  useMessage,
  type DataTableColumns,
} from 'naive-ui'

import { client, errorMessage, imageUrl, thumbnailUrl } from '@/api/client'
import { useIsMobile } from '@/utils/media'

import type { components } from '@/api/schema'

type ImageSummary = components['schemas']['ImageSummary']

const message = useMessage()
const dialog = useDialog()
const isMobile = useIsMobile()

const loading = ref(false)
const items = ref<ImageSummary[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const onlyUnused = ref(false)

/** 格式化字节数为可读文本。 */
function formatBytes(bytes: number): string {
  if (bytes < 1024) {
    return `${bytes} B`
  }
  if (bytes < 1024 * 1024) {
    return `${(bytes / 1024).toFixed(1)} KB`
  }
  return `${(bytes / 1024 / 1024).toFixed(2)} MB`
}

function formatDateTime(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) {
    return iso
  }
  const pad = (n: number): string => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

async function loadList(): Promise<void> {
  loading.value = true
  try {
    const { data, error } = await client.GET('/images', {
      params: {
        query: {
          page: page.value,
          pageSize: pageSize.value,
          onlyUnused: onlyUnused.value || undefined,
        },
      },
    })
    if (error || !data) {
      message.error(errorMessage(error))
      return
    }
    items.value = data.items
    total.value = data.pagination.total
  } finally {
    loading.value = false
  }
}

function handlePageChange(p: number): void {
  page.value = p
  void loadList()
}

function handleFilterChange(): void {
  page.value = 1
  void loadList()
}

function confirmDelete(row: ImageSummary): void {
  dialog.warning({
    title: '删除附件',
    content:
      row.refCount > 0
        ? `该附件被 ${row.refCount} 条隐患引用，通常无法删除，仍要尝试吗？`
        : '确定删除该未引用附件吗？将同时移除落盘文件，删除后不可恢复。',
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      const { error } = await client.DELETE('/images/{id}', {
        params: { path: { id: row.id } },
      })
      if (error) {
        message.error(errorMessage(error))
        return
      }
      message.success('已删除')
      void loadList()
    },
  })
}

function openOriginal(row: ImageSummary): void {
  window.open(imageUrl(row.id), '_blank', 'noopener')
}

const columns: DataTableColumns<ImageSummary> = [
  {
    title: '预览',
    key: 'thumbnail',
    width: 92,
    render: (row) =>
      h(NImage, {
        src: thumbnailUrl(row.id),
        previewSrc: imageUrl(row.id),
        width: 64,
        height: 64,
        objectFit: 'cover',
        style: 'border-radius:4px;border:1px solid #dbe5f1;overflow:hidden',
      }),
  },
  { title: 'UUID', key: 'id', minWidth: 220, ellipsis: { tooltip: true } },
  { title: '格式', key: 'mimeType', width: 110 },
  { title: '大小', key: 'sizeBytes', width: 100, render: (row) => formatBytes(row.sizeBytes) },
  {
    title: '尺寸',
    key: 'dimension',
    width: 110,
    render: (row) => (row.width && row.height ? `${row.width}×${row.height}` : '—'),
  },
  {
    title: '引用数',
    key: 'refCount',
    width: 100,
    render: (row) =>
      h(
        NTag,
        { type: row.refCount > 0 ? 'info' : 'default', size: 'small', bordered: false },
        { default: () => String(row.refCount) },
      ),
  },
  {
    title: '上传时间',
    key: 'createdAt',
    width: 150,
    render: (row) => formatDateTime(row.createdAt),
  },
  {
    title: '操作',
    key: 'actions',
    width: 150,
    fixed: 'right',
    render: (row) =>
      h('div', { style: 'display:flex;gap:8px' }, [
        h(NButton, { size: 'small', onClick: () => openOriginal(row) }, { default: () => '原图' }),
        h(
          NButton,
          { size: 'small', type: 'error', secondary: true, onClick: () => confirmDelete(row) },
          { default: () => '删除' },
        ),
      ]),
  },
]

onMounted(loadList)
</script>

<template>
  <div>
    <div class="page-toolbar">
      <div class="filter-inline">
        <n-switch v-model:value="onlyUnused" @update:value="handleFilterChange" />
        <span class="filter-label">仅看未引用</span>
      </div>
      <div class="spacer" />
      <n-button :loading="loading" @click="loadList">刷新</n-button>
    </div>

    <n-data-table
      :columns="columns"
      :data="items"
      :loading="loading"
      :bordered="false"
      :row-key="(row: ImageSummary) => row.id"
      :scroll-x="960"
      size="small"
    />
    <div class="pager">
      <n-pagination
        :page="page"
        :page-size="pageSize"
        :item-count="total"
        :simple="isMobile"
        @update:page="handlePageChange"
      />
      <span class="pager-total">共 {{ total }} 条</span>
    </div>
  </div>
</template>

<style scoped>
.filter-inline {
  display: flex;
  align-items: center;
  gap: 8px;
}

.filter-label {
  font-size: 14px;
  color: #1f2937;
}

.pager {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 14px;
}

.pager-total {
  font-size: 13px;
  color: #1f2937;
}
</style>
