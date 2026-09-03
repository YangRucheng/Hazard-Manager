<script setup lang="ts">
import { computed, h, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  NButton,
  NCard,
  NCascader,
  NDataTable,
  NDatePicker,
  NInput,
  NPagination,
  NSelect,
  NIcon,
  useDialog,
  useMessage,
  type DataTableColumns,
  type CascaderOption,
  type SelectOption,
} from 'naive-ui'
import { AddOutline, RefreshOutline } from '@vicons/ionicons5'

import { client, errorMessage } from '@/api/client'
import StatusTag from '@/components/StatusTag.vue'
import LevelTag from '@/components/LevelTag.vue'
import ImagePreview from '@/components/ImagePreview.vue'
import { formatDate, isOverdue } from '@/utils/date'

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
  categoryId: number | null
  unitId: number | null
  area: string
  keyword: string
  dateRange: [number, number] | null
}

const router = useRouter()
const message = useMessage()
const dialog = useDialog()

const loading = ref(false)
const items = ref<Hazard[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const units = ref<ResponsibleUnit[]>([])
const types = ref<HazardType[]>([])
const typePath = ref<number[]>([])

const filters = ref<Filters>({
  status: null,
  level: null,
  typeId: null,
  categoryId: null,
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

/** 将扁平类型数据组为级联选项（大类 -> 小类）。 */
const typeOptions = computed<CascaderOption[]>(() =>
  types.value
    .filter((t) => t.parentId === 0)
    .map((root) => {
      const children = types.value
        .filter((t) => t.parentId === root.id)
        .map((c) => ({ label: c.name, value: c.id }))
      return children.length > 0 ? { label: root.name, value: root.id, children } : { label: root.name, value: root.id }
    }),
)

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
          categoryId: f.categoryId ?? undefined,
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
  filters.value = { status: null, level: null, typeId: null, categoryId: null, unitId: null, area: '', keyword: '', dateRange: null }
  typePath.value = []
  page.value = 1
  void loadList()
}

/** 级联选中值（兼容 naive-ui：number[] | number | null 等）。 */
type CascaderValue = number[] | string[] | number | string | null

function handleTypeCascade(value: CascaderValue): void {
  typePath.value = Array.isArray(value) ? (value as number[]) : []
  const first = typePath.value[0]
  const second = typePath.value[1]
  filters.value.typeId = typeof first === 'number' ? first : null
  filters.value.categoryId = typeof second === 'number' ? second : null
  page.value = 1
  void loadList()
}

function goCreate(): void {
  void router.push({ name: 'hazard-create' })
}

function goEdit(id: number): void {
  void router.push({ name: 'hazard-edit', params: { id: String(id) } })
}

function confirmDelete(row: Hazard): void {
  dialog.warning({
    title: '删除隐患',
    content: `确定删除「${row.description}」吗？删除后不可恢复。`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      const { error } = await client.DELETE('/hazards/{id}', { params: { path: { id: row.id } } })
      if (error) {
        message.error(errorMessage(error))
        return
      }
      message.success('已删除')
      void loadList()
    },
  })
}

const columns: DataTableColumns<Hazard> = [
  { title: 'ID', key: 'id', width: 70 },
  { title: '检查日期', key: 'inspectionDate', width: 105 },
  { title: '区域', key: 'inspectionArea', width: 110, ellipsis: { tooltip: true } },
  { title: '隐患描述', key: 'description', minWidth: 200, ellipsis: { tooltip: true } },
  {
    title: '类型 / 分类',
    key: 'categoryName',
    width: 130,
    render: (row) => h('span', { class: 'type-cell' }, `${row.typeName} / ${row.categoryName}`),
  },
  { title: '责任单位', key: 'unitName', width: 110, ellipsis: { tooltip: true } },
  { title: '责任人', key: 'person', width: 80 },
  {
    title: '要求完成',
    key: 'dueDate',
    width: 105,
    render: (row) => {
      const overdue = isOverdue(row.dueDate, row.status)
      return h('span', { style: overdue ? 'color:#d03050;font-weight:600' : undefined }, row.dueDate)
    },
  },
  {
    title: '状态',
    key: 'status',
    width: 96,
    render: (row) => {
      const overdue = isOverdue(row.dueDate, row.status)
      return h('div', { style: 'display:flex;align-items:center;gap:6px' }, [
        h(StatusTag, { status: row.status }),
        overdue ? h('span', { style: 'color:#d03050;font-size:12px' }, '逾期') : null,
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
    render: (row) => h(ImagePreview, { imageIds: row.beforeImageIds, size: 36 }),
  },
  {
    title: '操作',
    key: 'actions',
    width: 110,
    fixed: 'right',
    render: (row) =>
      h('div', { style: 'display:flex;gap:4px' }, [
        h(NButton, { size: 'small', text: true, type: 'primary', onClick: () => goEdit(row.id) }, { default: () => '编辑' }),
        h(NButton, { size: 'small', text: true, type: 'error', onClick: () => confirmDelete(row) }, { default: () => '删除' }),
      ]),
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
    <n-card :bordered="true" class="filter-card">
      <div class="page-toolbar">
        <n-cascader
          v-model:value="typePath"
          :options="typeOptions"
          placeholder="隐患类型 / 分类"
          clearable
          style="width: 190px"
          @update:value="handleTypeCascade"
        />
        <n-select
          v-model:value="filters.status"
          :options="statusOptions"
          placeholder="整改状态"
          clearable
          style="width: 130px"
          @update:value="handleSearch"
        />
        <n-select
          v-model:value="filters.level"
          :options="levelOptions"
          placeholder="隐患等级"
          clearable
          style="width: 130px"
          @update:value="handleSearch"
        />
        <n-select
          v-model:value="filters.unitId"
          :options="unitOptions"
          placeholder="责任单位"
          clearable
          filterable
          style="width: 150px"
          @update:value="handleSearch"
        />
        <n-input
          v-model:value="filters.area"
          placeholder="检查区域"
          clearable
          style="width: 130px"
          @keyup.enter="handleSearch"
        />
        <n-input
          v-model:value="filters.keyword"
          placeholder="描述 / 人员 / 单位 关键字"
          clearable
          style="width: 180px"
          @keyup.enter="handleSearch"
        />
        <n-date-picker
          v-model:value="filters.dateRange"
          type="daterange"
          placeholder="检查日期范围"
          clearable
          style="width: 240px"
        />
        <div class="spacer" />
        <n-button type="primary" @click="handleSearch">
          <template #icon><n-icon><refresh-outline /></n-icon></template>
          查询
        </n-button>
        <n-button @click="handleReset">重置</n-button>
        <n-button type="primary" secondary @click="goCreate">
          <template #icon><n-icon><add-outline /></n-icon></template>
          新增隐患
        </n-button>
      </div>

      <n-data-table
        :columns="columns"
        :data="items"
        :loading="loading"
        :bordered="false"
        :row-key="(row: Hazard) => row.id"
        :scroll-x="1400"
        size="small"
      />
      <div class="pager">
        <n-pagination
          :page="page"
          :page-size="pageSize"
          :item-count="total"
          :page-sizes="[10, 20, 50]"
          show-size-picker
          @update:page="handlePageChange"
          @update:page-size="(size: number) => { pageSize = size; page = 1; void loadList() }"
        />
        <span class="pager-total">共 {{ total }} 条</span>
      </div>
    </n-card>
  </div>
</template>

<style scoped>
.pager {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 14px;
}

.pager-total {
  font-size: 13px;
  color: #6b7a90;
}

.type-cell {
  color: #17233d;
}
</style>