<script setup lang="ts">
import { computed, h, onMounted, ref } from 'vue'
import {
  NButton,
  NDataTable,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NModal,
  NSelect,
  NSwitch,
  NTag,
  useDialog,
  useMessage,
  type DataTableColumns,
  type FormInst,
  type FormRules,
  type SelectOption,
} from 'naive-ui'

import { client, errorMessage } from '@/api/client'

import type { components } from '@/api/schema'

type HazardType = components['schemas']['HazardType']

interface TypeRow {
  id: number
  parentId: number
  name: string
  sort: number
  status: 0 | 1
  kind: 'type' | 'category'
  parentName: string
}

interface TypeFormModel {
  name: string
  parentId: number
  sort: number
  status: 0 | 1
}

const message = useMessage()
const dialog = useDialog()

const loading = ref(false)
const list = ref<HazardType[]>([])

const showModal = ref(false)
const editingId = ref<number | null>(null)
const saving = ref(false)
const formRef = ref<FormInst | null>(null)
const form = ref<TypeFormModel>({ name: '', parentId: 0, sort: 0, status: 1 })

const parentOptions = computed<SelectOption[]>(() => [
  { label: '（顶级）隐患类型', value: 0 },
  ...list.value
    .filter((t) => t.parentId === 0)
    .sort((a, b) => a.sort - b.sort || a.id - b.id)
    .map((t) => ({ label: `作为分类，归属：${t.name}`, value: t.id })),
])

const rows = computed<TypeRow[]>(() => {
  const out: TypeRow[] = []
  const roots = list.value
    .filter((t) => t.parentId === 0)
    .sort((a, b) => a.sort - b.sort || a.id - b.id)
  for (const root of roots) {
    out.push({ id: root.id, parentId: 0, name: root.name, sort: root.sort, status: root.status, kind: 'type', parentName: '' })
    const children = list.value
      .filter((t) => t.parentId === root.id)
      .sort((a, b) => a.sort - b.sort || a.id - b.id)
    for (const c of children) {
      out.push({ id: c.id, parentId: c.parentId, name: c.name, sort: c.sort, status: c.status, kind: 'category', parentName: root.name })
    }
  }
  return out
})

const rules: FormRules = {
  name: { required: true, message: '请输入名称', trigger: ['input', 'blur'] },
  parentId: {
    required: true,
    validator: () => (form.value.parentId >= 0 ? true : new Error('请选择归属')),
    trigger: ['change'],
  },
}

async function loadList(): Promise<void> {
  loading.value = true
  try {
    const { data, error } = await client.GET('/hazard-types')
    if (error || !data) {
      message.error(errorMessage(error))
      return
    }
    list.value = data
  } finally {
    loading.value = false
  }
}

function openCreateType(): void {
  editingId.value = null
  form.value = { name: '', parentId: 0, sort: list.value.length, status: 1 }
  showModal.value = true
}

function openCreateCategory(root: TypeRow): void {
  editingId.value = null
  form.value = { name: '', parentId: root.id, sort: list.value.length, status: 1 }
  showModal.value = true
}

function openEdit(row: TypeRow): void {
  editingId.value = row.id
  form.value = { name: row.name, parentId: row.parentId, sort: row.sort, status: row.status }
  showModal.value = true
}

async function handleSave(): Promise<void> {
  try {
    await formRef.value?.validate()
  } catch {
    return
  }
  saving.value = true
  try {
    if (editingId.value === null) {
      const { error } = await client.POST('/hazard-types', {
        body: {
          name: form.value.name.trim(),
          parentId: form.value.parentId,
          sort: form.value.sort,
          status: form.value.status,
        },
      })
      if (error) {
        message.error(errorMessage(error))
        return
      }
      message.success('新增成功')
    } else {
      const { error } = await client.PUT('/hazard-types/{id}', {
        params: { path: { id: editingId.value } },
        body: {
          name: form.value.name.trim(),
          parentId: form.value.parentId,
          sort: form.value.sort,
          status: form.value.status,
        },
      })
      if (error) {
        message.error(errorMessage(error))
        return
      }
      message.success('保存成功')
    }
    showModal.value = false
    void loadList()
  } finally {
    saving.value = false
  }
}

function confirmDelete(row: TypeRow): void {
  dialog.warning({
    title: '删除类型/分类',
    content: `确定删除「${row.name}」吗？存在子分类或被隐患引用的无法删除。`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      const { error } = await client.DELETE('/hazard-types/{id}', { params: { path: { id: row.id } } })
      if (error) {
        message.error(errorMessage(error))
        return
      }
      message.success('已删除')
      void loadList()
    },
  })
}

const columns: DataTableColumns<TypeRow> = [
  { title: 'ID', key: 'id', width: 70 },
  {
    title: '名称',
    key: 'name',
    minWidth: 200,
    render: (row) =>
      h(
        'div',
        { style: `display:flex;align-items:center;gap:8px;${row.kind === 'category' ? 'padding-left:28px;' : ''}` },
        [
          h(
            NTag,
            { size: 'small', type: row.kind === 'type' ? 'info' : 'default', bordered: false },
            { default: () => (row.kind === 'type' ? '类型' : '分类') },
          ),
          h('span', row.name),
        ],
      ),
  },
  {
    title: '归属',
    key: 'parentName',
    width: 130,
    render: (row) => (row.kind === 'category' ? row.parentName : '—'),
  },
  { title: '排序', key: 'sort', width: 80 },
  {
    title: '状态',
    key: 'status',
    width: 90,
    render: (row) =>
      h(
        NTag,
        { size: 'small', type: row.status === 1 ? 'success' : 'default', bordered: false },
        { default: () => (row.status === 1 ? '启用' : '停用') },
      ),
  },
  {
    title: '操作',
    key: 'actions',
    width: 190,
    render: (row) =>
      h('div', { style: 'display:flex;gap:4px' }, [
        row.kind === 'type'
          ? h(NButton, { size: 'small', text: true, type: 'primary', onClick: () => openCreateCategory(row) }, { default: () => '新增分类' })
          : null,
        h(NButton, { size: 'small', text: true, type: 'primary', onClick: () => openEdit(row) }, { default: () => '编辑' }),
        h(NButton, { size: 'small', text: true, type: 'error', onClick: () => confirmDelete(row) }, { default: () => '删除' }),
      ]),
  },
]

onMounted(loadList)
</script>

<template>
  <div>
    <div class="page-toolbar">
      <span class="hint">大类为「隐患类型」，其下为「隐患分类」，两级共用一张表</span>
      <div class="spacer" />
      <n-button type="primary" @click="openCreateType">新增类型</n-button>
    </div>
    <n-data-table
      :columns="columns"
      :data="rows"
      :loading="loading"
      :bordered="false"
      :row-key="(row: TypeRow) => row.id"
    />

    <n-modal
      v-model:show="showModal"
      preset="card"
      :title="editingId === null ? '新增类型/分类' : '编辑类型/分类'"
      style="width: 480px"
      :mask-closable="false"
    >
      <n-form ref="formRef" :model="form" :rules="rules" label-placement="left" label-width="86">
        <n-form-item label="名称" path="name">
          <n-input v-model:value="form.name" placeholder="如：线路老化" />
        </n-form-item>
        <n-form-item label="归属" path="parentId">
          <n-select v-model:value="form.parentId" :options="parentOptions" />
        </n-form-item>
        <n-form-item label="排序">
          <n-input-number v-model:value="form.sort" :min="0" style="width: 100%" />
        </n-form-item>
        <n-form-item label="启用">
          <n-switch v-model:value="form.status" :checked-value="1" :unchecked-value="0" />
        </n-form-item>
      </n-form>
      <template #footer>
        <div style="display: flex; justify-content: flex-end; gap: 8px">
          <n-button @click="showModal = false">取消</n-button>
          <n-button type="primary" :loading="saving" @click="handleSave">保存</n-button>
        </div>
      </template>
    </n-modal>
  </div>
</template>

<style scoped>
.hint {
  font-size: 13px;
  color: #6b7a90;
}
</style>