<script setup lang="ts">
import { h, onMounted, ref } from 'vue'
import {
  NButton,
  NDataTable,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NModal,
  NSwitch,
  NTag,
  useDialog,
  useMessage,
  type DataTableColumns,
  type FormInst,
  type FormRules,
} from 'naive-ui'

import { client, errorMessage } from '@/api/client'

import type { components } from '@/api/schema'

type ResponsibleUnit = components['schemas']['ResponsibleUnit']

interface UnitFormModel {
  name: string
  person: string
  remark: string
  sort: number
  status: 0 | 1
}

const message = useMessage()
const dialog = useDialog()

const loading = ref(false)
const items = ref<ResponsibleUnit[]>([])

const showModal = ref(false)
const editingId = ref<number | null>(null)
const saving = ref(false)
const formRef = ref<FormInst | null>(null)
const form = ref<UnitFormModel>({ name: '', person: '', remark: '', sort: 0, status: 1 })

const rules: FormRules = {
  name: { required: true, message: '请输入单位名称', trigger: ['input', 'blur'] },
  person: { required: true, message: '请输入责任人（与单位一一对应）', trigger: ['input', 'blur'] },
}

async function loadList(): Promise<void> {
  loading.value = true
  try {
    const { data, error } = await client.GET('/units')
    if (error || !data) {
      message.error(errorMessage(error))
      return
    }
    items.value = data
  } finally {
    loading.value = false
  }
}

function openCreate(): void {
  editingId.value = null
  form.value = { name: '', person: '', remark: '', sort: items.value.length, status: 1 }
  showModal.value = true
}

function openEdit(row: ResponsibleUnit): void {
  editingId.value = row.id
  form.value = {
    name: row.name,
    person: row.person,
    remark: row.remark ?? '',
    sort: row.sort,
    status: row.status,
  }
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
      const { error } = await client.POST('/units', {
        body: {
          name: form.value.name.trim(),
          person: form.value.person.trim(),
          remark: form.value.remark.trim() || null,
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
      const { error } = await client.PUT('/units/{id}', {
        params: { path: { id: editingId.value } },
        body: {
          name: form.value.name.trim(),
          person: form.value.person.trim(),
          remark: form.value.remark.trim() || null,
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

function toggleStatus(row: ResponsibleUnit): void {
  void client
    .PUT('/units/{id}', {
      params: { path: { id: row.id } },
      body: { status: row.status === 1 ? 0 : 1 },
    })
    .then(({ error }) => {
      if (error) {
        message.error(errorMessage(error))
        return
      }
      void loadList()
    })
}

function confirmDelete(row: ResponsibleUnit): void {
  dialog.warning({
    title: '删除责任单位',
    content: `确定删除「${row.name}」吗？已被隐患引用的单位无法删除。`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      const { error } = await client.DELETE('/units/{id}', { params: { path: { id: row.id } } })
      if (error) {
        message.error(errorMessage(error))
        return
      }
      message.success('已删除')
      void loadList()
    },
  })
}

const columns: DataTableColumns<ResponsibleUnit> = [
  { title: 'ID', key: 'id', width: 70 },
  { title: '单位名称', key: 'name', minWidth: 160 },
  { title: '责任人', key: 'person', width: 110 },
  { title: '备注', key: 'remark', ellipsis: { tooltip: true }, minWidth: 160 },
  { title: '排序', key: 'sort', width: 80 },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: (row) =>
      h(
        NSwitch,
        { value: row.status === 1, onUpdateValue: () => toggleStatus(row), size: 'small' },
        {
          checked: () => h(NTag, { type: 'success', size: 'small', bordered: false }, { default: () => '启用' }),
          unchecked: () => h(NTag, { type: 'default', size: 'small', bordered: false }, { default: () => '停用' }),
        },
      ),
  },
  {
    title: '操作',
    key: 'actions',
    width: 120,
    render: (row) =>
      h('div', { style: 'display:flex;gap:4px' }, [
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
      <div class="spacer" />
      <n-button type="primary" @click="openCreate">新增单位</n-button>
    </div>
    <n-data-table
      :columns="columns"
      :data="items"
      :loading="loading"
      :bordered="false"
      :row-key="(row: ResponsibleUnit) => row.id"
    />

    <n-modal
      v-model:show="showModal"
      preset="card"
      :title="editingId === null ? '新增责任单位' : '编辑责任单位'"
      style="width: 460px"
      :mask-closable="false"
    >
      <n-form ref="formRef" :model="form" :rules="rules" label-placement="left" label-width="72">
        <n-form-item label="单位名称" path="name">
          <n-input v-model:value="form.name" placeholder="如：电气车间" />
        </n-form-item>
        <n-form-item label="责任人" path="person">
          <n-input v-model:value="form.person" placeholder="与该单位一一对应" />
        </n-form-item>
        <n-form-item label="备注">
          <n-input v-model:value="form.remark" placeholder="可选" />
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