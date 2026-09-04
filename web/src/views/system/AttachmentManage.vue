<script setup lang="ts">
import { h, onMounted, ref } from 'vue'
import {
  NButton,
  NCard,
  NDataTable,
  NPagination,
  NSwitch,
  NTag,
  useDialog,
  useMessage,
  type DataTableColumns,
} from 'naive-ui'

import AuthImage from '@/components/AuthImage.vue'
import { authedBlobUrl, client, errorMessage, imageUrl, thumbnailUrl } from '@/api/client'
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

/** 原图与列表缩略图同样受鉴权保护，需带令牌取回 blob 后再在新标签打开。 */
const openingId = ref<string | null>(null)

async function openOriginal(row: ImageSummary): Promise<void> {
  if (openingId.value !== null) {
    return
  }
  openingId.value = row.id
  try {
    const url = await authedBlobUrl(imageUrl(row.id))
    window.open(url, '_blank', 'noopener')
  } catch (err) {
    message.error(err instanceof Error ? err.message : '原图加载失败')
  } finally {
    openingId.value = null
  }
}

const columns: DataTableColumns<ImageSummary> = [
  {
    title: '预览',
    key: 'thumbnail',
    width: 96,
    render: (row) =>
      h(AuthImage, {
        url: thumbnailUrl(row.id),
        previewUrl: imageUrl(row.id),
        width: 64,
        height: 64,
        style: 'border-radius:6px;border:1px solid #e2e7ef;overflow:hidden',
      }),
  },
  { title: 'UUID', key: 'id', minWidth: 240, ellipsis: { tooltip: true } },
  { title: '格式', key: 'mimeType', width: 120 },
  { title: '大小', key: 'sizeBytes', width: 110, render: (row) => formatBytes(row.sizeBytes) },
  {
    title: '尺寸',
    key: 'dimension',
    width: 120,
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
    width: 160,
    render: (row) => formatDateTime(row.createdAt),
  },
  {
    title: '操作',
    key: 'actions',
    width: 160,
    fixed: 'right',
    render: (row) =>
      h('div', { style: 'display:flex;gap:8px' }, [
        h(
          NButton,
          {
            size: 'small',
            loading: openingId.value === row.id,
            onClick: () => void openOriginal(row),
          },
          { default: () => '原图' },
        ),
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
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">附件管理</h1>
      </div>
      <n-button :loading="loading" @click="loadList">刷新</n-button>
    </div>

    <n-card class="filter-card">
      <div class="filter-heading">
        <span class="filter-title">附件范围</span>
      </div>
      <div class="filter-toggle">
        <n-switch v-model:value="onlyUnused" @update:value="handleFilterChange" />
        <span>仅看未被隐患引用的附件</span>
      </div>
    </n-card>

    <n-card class="data-card">
      <n-data-table
        :columns="columns"
        :data="items"
        :loading="loading"
        :bordered="false"
        :row-key="(row: ImageSummary) => row.id"
        :scroll-x="1100"
        size="small"
      />
      <div class="pagination-bar">
        <span class="muted pager-total">共 {{ total }} 条</span>
        <n-pagination
          :page="page"
          :page-size="pageSize"
          :item-count="total"
          :simple="isMobile"
          @update:page="handlePageChange"
        />
      </div>
    </n-card>
  </div>
</template>

<style scoped>
.filter-toggle {
  display: flex;
  align-items: center;
  gap: 10px;
  color: var(--color-text);
  font-size: 14px;
}

.pager-total {
  margin-right: auto;
  font-size: 13px;
}
</style>
