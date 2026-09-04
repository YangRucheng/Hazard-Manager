<script setup lang="ts">
import { h, onMounted, ref } from 'vue'
import {
  NButton,
  NCard,
  NDataTable,
  NForm,
  NFormItem,
  NInput,
  NModal,
  NSwitch,
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
const form = ref<UnitFormModel>({ name: '', person: '', remark: '', status: 1 })

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
  form.value = { name: '', person: '', remark: '', status: 1 }
  showModal.value = true
}

function openEdit(row: ResponsibleUnit): void {
  editingId.value = row.id
  form.value = {
    name: row.name,
    person: row.person,
    remark: row.remark ?? '',
    status: row.status,
  }
  showModal.value = true
}

/** 整行可点击打开编辑弹窗（操作列的编辑按钮已移除）。 */
function rowProps(row: ResponsibleUnit) {
  return { onClick: () => openEdit(row) }
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

/** 删除按钮位于编辑弹窗内：已被隐患引用的单位会被后端拒绝删除。 */
function confirmDelete(): void {
  const row = items.value.find((u) => u.id === editingId.value)
  if (!row) {
    return
  }
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
      showModal.value = false
      void loadList()
    },
  })
}

const columns: DataTableColumns<ResponsibleUnit> = [
  { title: '单位名称', key: 'name', minWidth: 180 },
  { title: '责任人', key: 'person', width: 140 },
  { title: '备注', key: 'remark', ellipsis: { tooltip: true }, minWidth: 180 },
  {
    title: '状态',
    key: 'status',
    width: 90,
    render: (row) =>
      h(
        'span',
        { onClick: (e: MouseEvent) => e.stopPropagation() },
        [
          h(NSwitch, {
            value: row.status === 1,
            size: 'small',
            onUpdateValue: () => toggleStatus(row),
          }),
        ],
      ),
  },
]

onMounted(loadList)
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">责任单位</h1>
      </div>
      <n-button type="primary" @click="openCreate">新增单位</n-button>
    </div>

    <n-card class="data-card">
      <n-data-table
        class="row-clickable"
        :columns="columns"
        :data="items"
        :loading="loading"
        :bordered="false"
        :scroll-x="720"
        size="small"
        :row-key="(row: ResponsibleUnit) => row.id"
        :row-props="rowProps"
      />
    </n-card>

    <n-modal
      v-model:show="showModal"
      preset="card"
      draggable
      :title="editingId === null ? '新增责任单位' : '编辑责任单位'"
      style="width: min(560px, calc(100vw - 32px))"
      :mask-closable="false"
    >
      <n-form ref="formRef" :model="form" :rules="rules" label-placement="top">
        <n-form-item label="单位名称" path="name">
          <n-input v-model:value="form.name" placeholder="如：电气车间" />
        </n-form-item>
        <n-form-item label="责任人" path="person">
          <n-input v-model:value="form.person" placeholder="与该单位一一对应" />
        </n-form-item>
        <n-form-item label="备注">
          <n-input
            v-model:value="form.remark"
            type="textarea"
            :autosize="{ minRows: 2, maxRows: 5 }"
            placeholder="可选，补充单位职责、联系方式等说明"
          />
        </n-form-item>
        <n-form-item label="启用">
          <n-switch v-model:value="form.status" :checked-value="1" :unchecked-value="0" />
          <span class="switch-hint">
            {{ form.status === 1 ? '启用：新增隐患时可选择该单位' : '停用：下拉选择时不可用' }}
          </span>
        </n-form-item>
      </n-form>
      <template #footer>
        <div class="modal-footer">
          <div class="modal-footer-left">
            <n-button v-if="editingId !== null" type="error" secondary @click="confirmDelete">
              删除
            </n-button>
          </div>
          <div class="modal-footer-right">
            <n-button @click="showModal = false">取消</n-button>
            <n-button type="primary" :loading="saving" @click="handleSave">保存</n-button>
          </div>
        </div>
      </template>
    </n-modal>
  </div>
</template>

<style scoped>
.row-clickable :deep(.n-data-table-tr) {
  cursor: pointer;
}

.row-clickable :deep(.n-data-table-tr:hover) {
  background: rgba(63, 99, 216, 0.05);
}

.switch-hint {
  margin-left: 10px;
  font-size: 12px;
  color: var(--color-text-muted);
}

.modal-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.modal-footer-right {
  display: flex;
  gap: 8px;
}
</style>
