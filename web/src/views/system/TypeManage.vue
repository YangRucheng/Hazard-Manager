<script setup lang="ts">
import { computed, h, onMounted, ref } from 'vue'
import {
  NButton,
  NDataTable,
  NForm,
  NFormItem,
  NInput,
  NModal,
  NTag,
  NAutoComplete,
  useDialog,
  useMessage,
  type DataTableColumns,
  type FormInst,
  type FormRules,
  type AutoCompleteOption,
} from 'naive-ui'

import { client, errorMessage } from '@/api/client'

import type { components } from '@/api/schema'

type HazardType = components['schemas']['HazardType']

interface TypeFormModel {
  major: string
  minor: string
}

const message = useMessage()
const dialog = useDialog()

const loading = ref(false)
const list = ref<HazardType[]>([])

const showModal = ref(false)
const editingId = ref<number | null>(null)
const saving = ref(false)
const formRef = ref<FormInst | null>(null)
const form = ref<TypeFormModel>({ major: '', minor: '' })

/** 已有大类（去重，作为大类下拉输入框的候选项）。 */
const majorOptions = computed<AutoCompleteOption[]>(() => {
  const seen = new Set<string>()
  const options: AutoCompleteOption[] = []
  for (const t of list.value) {
    if (t.major && !seen.has(t.major)) {
      seen.add(t.major)
      options.push({ label: t.major, value: t.major })
    }
  }
  return options
})

const rules: FormRules = {
  major: {
    required: true,
    validator: () => (form.value.major.trim() ? true : new Error('请输入或选择大类')),
    trigger: ['input', 'blur'],
  },
  minor: { required: true, message: '请输入小类', trigger: ['input', 'blur'] },
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

function openCreate(): void {
  editingId.value = null
  form.value = { major: '', minor: '' }
  showModal.value = true
}

function openEdit(row: HazardType): void {
  editingId.value = row.id
  form.value = { major: row.major, minor: row.minor }
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
    const body = { major: form.value.major.trim(), minor: form.value.minor.trim() }
    if (editingId.value === null) {
      const { error } = await client.POST('/hazard-types', { body })
      if (error) {
        message.error(errorMessage(error))
        return
      }
      message.success('新增成功')
    } else {
      const { error } = await client.PUT('/hazard-types/{id}', {
        params: { path: { id: editingId.value } },
        body,
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

/** 删除按钮位于编辑弹窗内：被隐患引用的类型会被后端拒绝（只能修改）。 */
function confirmDelete(): void {
  const row = list.value.find((t) => t.id === editingId.value)
  if (!row) {
    return
  }
  dialog.warning({
    title: '删除隐患类型',
    content: `确定删除「${row.major} / ${row.minor}」吗？已被隐患记录引用的类型无法删除，只能修改。`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      const { error } = await client.DELETE('/hazard-types/{id}', {
        params: { path: { id: row.id } },
      })
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

const columns: DataTableColumns<HazardType> = [
  { title: 'ID', key: 'id', width: 70 },
  {
    title: '大类',
    key: 'major',
    minWidth: 180,
    render: (row) =>
      h('span', { style: 'display:inline-flex;align-items:center;gap:8px' }, [
        h(NTag, { size: 'small', type: 'info', bordered: false }, { default: () => row.major }),
      ]),
  },
  { title: '小类', key: 'minor', minWidth: 200 },
  {
    title: '操作',
    key: 'actions',
    width: 110,
    render: (row) =>
      h('div', { style: 'display:flex;gap:8px' }, [
        h(NButton, { size: 'small', onClick: () => openEdit(row) }, { default: () => '编辑' }),
      ]),
  },
]

onMounted(loadList)
</script>

<template>
  <div>
    <div class="page-toolbar">
      <div class="spacer" />
      <n-button type="primary" @click="openCreate">新增隐患类型</n-button>
    </div>
    <n-data-table
      :columns="columns"
      :data="list"
      :loading="loading"
      :bordered="false"
      :scroll-x="640"
      size="small"
      :row-key="(row: HazardType) => row.id"
    />

    <n-modal
      v-model:show="showModal"
      preset="card"
      draggable
      :title="editingId === null ? '新增隐患类型' : '编辑隐患类型'"
      style="width: 520px; max-width: 94vw"
      :mask-closable="false"
    >
      <n-form ref="formRef" :model="form" :rules="rules" label-placement="left" label-width="72">
        <n-form-item label="大类" path="major">
          <n-auto-complete
            v-model:value="form.major"
            :options="majorOptions"
            placeholder="可下拉选择已有大类，也可直接输入新大类"
            clearable
          />
        </n-form-item>
        <n-form-item label="小类" path="minor">
          <n-input v-model:value="form.minor" placeholder="如：线路老化" />
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
