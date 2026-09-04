<script setup lang="ts">
import { computed, h, onMounted, ref } from 'vue'
import {
  NButton,
  NCard,
  NDataTable,
  NDatePicker,
  NInput,
  NPagination,
  NSelect,
  NTag,
  useMessage,
  type DataTableColumns,
  type SelectOption,
  type SelectGroupOption,
} from 'naive-ui'

import { client, errorMessage } from '@/api/client'
import StatusTag from '@/components/StatusTag.vue'
import LevelTag from '@/components/LevelTag.vue'
import ImagePreview from '@/components/ImagePreview.vue'
import HazardFormModal from '@/components/HazardFormModal.vue'
import { formatDate, isOverdue } from '@/utils/date'
import { useIsMobile } from '@/utils/media'

import type { components } from '@/api/schema'

type Hazard = components['schemas']['Hazard']
type ResponsibleUnit = components['schemas']['ResponsibleUnit']
type HazardType = components['schemas']['HazardType']
type HazardStatus = components['schemas']['HazardStatus']
type HazardLevel = components['schemas']['HazardLevel']

interface Filters {
  status: HazardStatus | null
  level: HazardLevel | null
  typeId: number | null
  unitId: number | null
  area: string
  keyword: string
  dateRange: [number, number] | null
}

const message = useMessage()
const isMobile = useIsMobile()

const loading = ref(false)
const items = ref<Hazard[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const units = ref<ResponsibleUnit[]>([])
const types = ref<HazardType[]>([])

/** 新增/编辑弹窗（hazardId 为空表示新增）。 */
const showModal = ref(false)
const editingId = ref<number | null>(null)

const filters = ref<Filters>({
  status: null,
  level: null,
  typeId: null,
  unitId: null,
  area: '',
  keyword: '',
  dateRange: null,
})

const statusOptions: SelectOption[] = [
  { label: '待整改', value: '待整改' },
  { label: '整改受阻', value: '整改受阻' },
  { label: '已整改', value: '已整改' },
]

const levelOptions: SelectOption[] = [
  { label: '一般隐患', value: '一般隐患' },
  { label: '重大隐患', value: '重大隐患' },
]

const unitOptions = computed<SelectOption[]>(() =>
  units.value.map((u) => ({ label: u.name, value: u.id })),
)

/** 类型筛选项：按大类分组，选择小类（= 组合行）即按该类型过滤。 */
const typeOptions = computed<SelectGroupOption[]>(() => {
  const groupMap = new Map<string, SelectOption[]>()
  for (const t of types.value) {
    if (!groupMap.has(t.major)) {
      groupMap.set(t.major, [])
    }
    groupMap.get(t.major)!.push({ label: t.minor, value: t.id })
  }
  return [...groupMap.entries()].map(([label, options]) => ({
    type: 'group',
    label,
    key: label,
    options,
  }))
})

async function loadUnits(): Promise<void> {
  const { data, error } = await client.GET('/units')
  if (error || !data) {
    message.error(errorMessage(error))
    return
  }
  units.value = data
}

async function loadTypes(): Promise<void> {
  const { data, error } = await client.GET('/hazard-types')
  if (error || !data) {
    message.error(errorMessage(error))
    return
  }
  types.value = data
}

async function loadList(): Promise<void> {
  loading.value = true
  try {
    const f = filters.value
    const { data, error } = await client.GET('/hazards', {
      params: {
        query: {
          page: page.value,
          pageSize: pageSize.value,
          status: f.status ?? undefined,
          level: f.level ?? undefined,
          typeId: f.typeId ?? undefined,
          unitId: f.unitId ?? undefined,
          area: f.area || undefined,
          keyword: f.keyword || undefined,
          dateFrom: f.dateRange ? formatDate(new Date(f.dateRange[0])) : undefined,
          dateTo: f.dateRange ? formatDate(new Date(f.dateRange[1])) : undefined,
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

function handleSearch(): void {
  page.value = 1
  void loadList()
}

function handleReset(): void {
  filters.value = { status: null, level: null, typeId: null, unitId: null, area: '', keyword: '', dateRange: null }
  page.value = 1
  void loadList()
}

function openCreate(): void {
  editingId.value = null
  showModal.value = true
}

function openEdit(row: Hazard): void {
  editingId.value = row.id
  showModal.value = true
}

/** 整行可点击弹出编辑弹窗（操作列已移除，删除在弹窗内）。 */
function rowProps(row: Hazard) {
  return { onClick: () => openEdit(row) }
}

function handleAfterSaved(): void {
  void loadList()
}

const columns: DataTableColumns<Hazard> = [
  { title: 'ID', key: 'id', width: 70 },
  { title: '检查日期', key: 'inspectionDate', width: 105 },
  { title: '区域', key: 'inspectionArea', width: 110, ellipsis: { tooltip: true } },
  { title: '隐患描述', key: 'description', minWidth: 200, ellipsis: { tooltip: true } },
  {
    title: '类型',
    key: 'type',
    width: 170,
    render: (row) =>
      h('span', { class: 'type-cell' }, `${row.typeMajor} / ${row.typeMinor}`),
  },
  { title: '责任单位', key: 'unitName', width: 110, ellipsis: { tooltip: true } },
  { title: '责任人', key: 'person', width: 80 },
  {
    title: '整改员工',
    key: 'rectifyPerson',
    width: 100,
    render: (row) => row.rectifyPerson || '—',
  },
  { title: '要求完成', key: 'dueDate', width: 105 },
  {
    title: '状态',
    key: 'status',
    width: 130,
    render: (row) => {
      const overdue = isOverdue(row.dueDate, row.status)
      return h('div', { style: 'display:flex;align-items:center;gap:6px' }, [
        h(StatusTag, { status: row.status }),
        overdue
          ? h(NTag, { type: 'error', size: 'small', bordered: false }, { default: () => '逾期' })
          : null,
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
    title: '整改前图片',
    key: 'beforeImageIds',
    width: 150,
    render: (row) =>
      h(
        'span',
        // 点缩略图仅预览，不冒泡触发整行编辑弹窗。
        { onClick: (e: MouseEvent) => e.stopPropagation() },
        [h(ImagePreview, { imageIds: row.beforeImageIds, size: 36 })],
      ),
  },
]

onMounted(() => {
  void loadUnits()
  void loadTypes()
  void loadList()
})
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">隐患台账</h1>
      </div>
      <n-button type="primary" @click="openCreate">新增隐患</n-button>
    </div>

    <n-card class="filter-card">
      <div class="filter-heading">
        <span class="filter-title">筛选条件</span>
      </div>
      <div class="filter-fields">
        <label class="filter-field">
          <span>隐患类型</span>
          <n-select
            v-model:value="filters.typeId"
            :options="typeOptions"
            placeholder="按大类分组选择"
            clearable
            filterable
            class="full-width"
            @update:value="handleSearch"
          />
        </label>
        <label class="filter-field">
          <span>整改状态</span>
          <n-select
            v-model:value="filters.status"
            :options="statusOptions"
            placeholder="全部"
            clearable
            class="full-width"
            @update:value="handleSearch"
          />
        </label>
        <label class="filter-field">
          <span>隐患等级</span>
          <n-select
            v-model:value="filters.level"
            :options="levelOptions"
            placeholder="全部"
            clearable
            class="full-width"
            @update:value="handleSearch"
          />
        </label>
        <label class="filter-field">
          <span>责任单位</span>
          <n-select
            v-model:value="filters.unitId"
            :options="unitOptions"
            placeholder="全部"
            clearable
            filterable
            class="full-width"
            @update:value="handleSearch"
          />
        </label>
        <label class="filter-field">
          <span>检查区域</span>
          <n-input
            v-model:value="filters.area"
            placeholder="如：华星现场"
            clearable
            class="full-width"
            @keyup.enter="handleSearch"
          />
        </label>
        <label class="filter-field">
          <span>描述 / 人员 / 单位关键字</span>
          <n-input
            v-model:value="filters.keyword"
            placeholder="模糊搜索"
            clearable
            class="full-width"
            @keyup.enter="handleSearch"
          />
        </label>
        <label class="filter-field">
          <span>检查日期范围</span>
          <n-date-picker
            v-model:value="filters.dateRange"
            type="daterange"
            placeholder="起始 — 截止"
            clearable
            class="full-width"
          />
        </label>
      </div>
      <div class="filter-actions">
        <n-button @click="handleReset">重置</n-button>
        <n-button type="primary" :loading="loading" @click="handleSearch">查询</n-button>
      </div>
    </n-card>

    <n-card class="data-card">
      <n-data-table
        class="row-clickable"
        :columns="columns"
        :data="items"
        :loading="loading"
        :bordered="false"
        :row-key="(row: Hazard) => row.id"
        :row-props="rowProps"
        :scroll-x="1500"
        size="small"
      />
      <div class="pagination-bar">
        <span class="muted pager-total">共 {{ total }} 条</span>
        <n-pagination
          :page="page"
          :page-size="pageSize"
          :item-count="total"
          :page-sizes="[10, 20, 50]"
          :show-size-picker="!isMobile"
          :simple="isMobile"
          @update:page="handlePageChange"
          @update:page-size="(size: number) => { pageSize = size; page = 1; void loadList() }"
        />
      </div>
    </n-card>

    <HazardFormModal
      v-model:show="showModal"
      :hazard-id="editingId"
      @saved="handleAfterSaved"
      @deleted="handleAfterSaved"
    />
  </div>
</template>

<style scoped>
.filter-fields {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 12px 16px;
}

.filter-fields .filter-field {
  min-width: 0;
}

.filter-fields .filter-field > span {
  line-height: 1.5;
}

.type-cell {
  color: var(--color-text-strong);
}

.pager-total {
  margin-right: auto;
  font-size: 13px;
}

.row-clickable :deep(.n-data-table-tr) {
  cursor: pointer;
}

.row-clickable :deep(.n-data-table-tr:hover) {
  background: rgba(63, 99, 216, 0.05);
}

@media (max-width: 640px) {
  .filter-fields {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
