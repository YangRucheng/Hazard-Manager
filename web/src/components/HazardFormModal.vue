<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import {
  NButton,
  NDatePicker,
  NForm,
  NFormItem,
  NGrid,
  NGridItem,
  NInput,
  NModal,
  NSelect,
  useDialog,
  useMessage,
  type FormInst,
  type FormRules,
  type SelectOption,
  type SelectGroupOption,
} from 'naive-ui'

import { client, errorMessage } from '@/api/client'
import ImageUpload from '@/components/ImageUpload.vue'
import { formatDate } from '@/utils/date'
import { useIsMobile } from '@/utils/media'

import type { components } from '@/api/schema'

type Hazard = components['schemas']['Hazard']
type ResponsibleUnit = components['schemas']['ResponsibleUnit']
type HazardType = components['schemas']['HazardType']
type HazardStatus = components['schemas']['HazardStatus']
type HazardLevel = components['schemas']['HazardLevel']

interface FormModel {
  inspectionArea: string
  inspector: string
  description: string
  suggestion: string
  unitId: number | null
  recheckPerson: string
  rectifyPerson: string
  beforeImageIds: string[]
  status: HazardStatus
  afterImageIds: string[]
  typeId: number | null
  level: HazardLevel
  remark: string
}

const props = withDefaults(defineProps<{ show: boolean; hazardId?: number | null }>(), {
  hazardId: null,
})

const emit = defineEmits<{
  'update:show': [value: boolean]
  saved: []
  deleted: []
}>()

const message = useMessage()
const dialog = useDialog()
const isMobile = useIsMobile()

const isEdit = computed<boolean>(() => props.hazardId != null)

const formRef = ref<FormInst | null>(null)
const saving = ref(false)
const loading = ref(false)
const units = ref<ResponsibleUnit[]>([])
const types = ref<HazardType[]>([])

const inspectionTs = ref<number | null>(startOfToday())
const dueTs = ref<number | null>(addDaysTimestamp(startOfToday(), 7))
let dueTouched = false

const form = ref<FormModel>({
  inspectionArea: '华星现场',
  inspector: '电气自查',
  description: '',
  suggestion: '',
  unitId: null,
  recheckPerson: '',
  rectifyPerson: '',
  beforeImageIds: [],
  status: '待整改',
  afterImageIds: [],
  typeId: null,
  level: '一般隐患',
  remark: '',
})

/** 类型选项：按大类分组展示小类，选项值即「大类+小类」组合行 id。 */
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

const unitOptions = computed<SelectOption[]>(() =>
  units.value.map((u) => ({ label: u.name, value: u.id })),
)

/** 当前选中单位的责任人（只读联动展示）。 */
const linkedPerson = computed<string>(() => {
  const unit = units.value.find((u) => u.id === form.value.unitId)
  return unit ? unit.person : ''
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

const rules: FormRules = {
  description: [
    { required: true, message: '请填写隐患描述', trigger: ['input', 'blur'] },
    { min: 2, message: '隐患描述至少 2 个字符', trigger: ['input', 'blur'] },
  ],
  unitId: { required: true, type: 'number', message: '请选择责任单位', trigger: ['change'] },
  typeId: { required: true, type: 'number', message: '请选择隐患类型', trigger: ['change'] },
}

function defaultForm(): FormModel {
  return {
    inspectionArea: '华星现场',
    inspector: '电气自查',
    description: '',
    suggestion: '',
    unitId: null,
    recheckPerson: '',
    rectifyPerson: '',
    beforeImageIds: [],
    status: '待整改',
    afterImageIds: [],
    typeId: null,
    level: '一般隐患',
    remark: '',
  }
}

function resetDates(): void {
  const now = startOfToday()
  inspectionTs.value = now
  dueTs.value = addDaysTimestamp(now, 7)
  dueTouched = false
}

async function loadEnums(): Promise<void> {
  const [unitRes, typeRes] = await Promise.all([client.GET('/units'), client.GET('/hazard-types')])
  if (unitRes.error || !unitRes.data) {
    message.error(errorMessage(unitRes.error))
    return
  }
  units.value = unitRes.data
  if (typeRes.error || !typeRes.data) {
    message.error(errorMessage(typeRes.error))
    return
  }
  types.value = typeRes.data
}

async function loadDetail(id: number): Promise<void> {
  const { data, error } = await client.GET('/hazards/{id}', { params: { path: { id } } })
  if (error || !data) {
    message.error(errorMessage(error))
    return
  }
  applyDetail(data)
}

function applyDetail(h: Hazard): void {
  form.value = {
    inspectionArea: h.inspectionArea,
    inspector: h.inspector,
    description: h.description,
    suggestion: h.suggestion ?? '',
    unitId: h.unitId,
    recheckPerson: h.recheckPerson ?? '',
    rectifyPerson: h.rectifyPerson ?? '',
    beforeImageIds: h.beforeImageIds ?? [],
    status: h.status,
    afterImageIds: h.afterImageIds ?? [],
    typeId: h.typeId,
    level: h.level,
    remark: h.remark ?? '',
  }
  inspectionTs.value = Date.parse(`${h.inspectionDate}T00:00:00`)
  dueTs.value = Date.parse(`${h.dueDate}T00:00:00`)
  dueTouched = true
}

/** 打开时重置并（编辑态）拉取最新详情。 */
async function prepare(): Promise<void> {
  form.value = defaultForm()
  resetDates()
  formRef.value?.restoreValidation()
  loading.value = true
  try {
    await loadEnums()
    if (props.hazardId != null) {
      await loadDetail(props.hazardId)
    }
  } finally {
    loading.value = false
  }
}

watch(
  () => props.show,
  (open) => {
    if (open) {
      void prepare()
    }
  },
)

function handleInspectionDateChange(ts: number | null): void {
  inspectionTs.value = ts
  if (!dueTouched && ts !== null) {
    dueTs.value = ts + 7 * 24 * 60 * 60 * 1000
  }
}

function handleDueDateChange(ts: number | null): void {
  dueTs.value = ts
  dueTouched = true
}

function close(): void {
  emit('update:show', false)
}

async function handleSubmit(): Promise<void> {
  try {
    await formRef.value?.validate()
  } catch {
    message.warning('请完善表单必填项')
    return
  }
  if (form.value.unitId === null || dueTs.value === null || inspectionTs.value === null) {
    message.warning('请完善表单必填项')
    return
  }
  if (form.value.typeId === null) {
    message.warning('请选择隐患类型')
    return
  }

  saving.value = true
  try {
    const common = {
      inspectionArea: form.value.inspectionArea.trim() || '华星现场',
      inspectionDate: formatDate(new Date(inspectionTs.value)),
      inspector: form.value.inspector.trim() || '电气自查',
      description: form.value.description.trim(),
      suggestion: form.value.suggestion.trim() || null,
      unitId: form.value.unitId,
      dueDate: formatDate(new Date(dueTs.value)),
      recheckPerson: form.value.recheckPerson.trim() || null,
      rectifyPerson: form.value.rectifyPerson.trim() || null,
      beforeImageIds: form.value.beforeImageIds,
      status: form.value.status,
      afterImageIds: form.value.afterImageIds,
      typeId: form.value.typeId,
      level: form.value.level,
      remark: form.value.remark.trim() || null,
    }

    if (props.hazardId != null) {
      const { data, error } = await client.PUT('/hazards/{id}', {
        params: { path: { id: props.hazardId } },
        body: common,
      })
      if (error || !data) {
        message.error(errorMessage(error))
        return
      }
      message.success('保存成功')
    } else {
      const { data, error } = await client.POST('/hazards', { body: common })
      if (error || !data) {
        message.error(errorMessage(error))
        return
      }
      message.success('新增成功')
    }
    emit('saved')
    close()
  } finally {
    saving.value = false
  }
}

/** 删除按钮位于弹窗底部左侧：二次确认后软删除。 */
function confirmDelete(): void {
  if (props.hazardId == null) {
    return
  }
  dialog.warning({
    title: '删除隐患',
    content: '确定删除该隐患记录吗？删除后不可恢复。',
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      const { error } = await client.DELETE('/hazards/{id}', {
        params: { path: { id: props.hazardId as number } },
      })
      if (error) {
        message.error(errorMessage(error))
        return
      }
      message.success('已删除')
      emit('deleted')
      close()
    },
  })
}

function startOfToday(): number {
  const d = new Date()
  d.setHours(0, 0, 0, 0)
  return d.getTime()
}

function addDaysTimestamp(ts: number, days: number): number {
  return ts + days * 24 * 60 * 60 * 1000
}
</script>

<template>
  <n-modal
    :show="show"
    preset="card"
    draggable
    :title="isEdit ? '编辑隐患' : '新增隐患'"
    style="width: min(820px, calc(100vw - 24px))"
    :mask-closable="false"
    @update:show="close"
  >
    <div class="modal-body">
      <n-form
        ref="formRef"
        :model="form"
        :rules="rules"
        :label-placement="isMobile ? 'top' : 'left'"
        :label-width="isMobile ? undefined : 100"
        :disabled="loading"
      >
        <n-grid :cols="isMobile ? 1 : 2" :x-gap="24">
          <n-grid-item>
            <n-form-item label="检查区域" path="inspectionArea">
              <n-input v-model:value="form.inspectionArea" placeholder="默认：华星现场" />
            </n-form-item>
          </n-grid-item>
          <n-grid-item>
            <n-form-item label="检查日期" path="inspectionDate">
              <n-date-picker
                :value="inspectionTs"
                type="date"
                style="width: 100%"
                @update:value="handleInspectionDateChange"
              />
            </n-form-item>
          </n-grid-item>
          <n-grid-item>
            <n-form-item label="检查人员" path="inspector">
              <n-input v-model:value="form.inspector" placeholder="默认：电气自查" />
            </n-form-item>
          </n-grid-item>
          <n-grid-item>
            <n-form-item label="隐患等级" path="level">
              <n-select v-model:value="form.level" :options="levelOptions" />
            </n-form-item>
          </n-grid-item>
          <n-grid-item :span="isMobile ? 1 : 2">
            <n-form-item label="隐患描述" path="description">
              <n-input
                v-model:value="form.description"
                type="textarea"
                :rows="3"
                placeholder="请详细描述隐患情况"
              />
            </n-form-item>
          </n-grid-item>
          <n-grid-item :span="isMobile ? 1 : 2">
            <n-form-item label="建议整改方案">
              <n-input
                v-model:value="form.suggestion"
                type="textarea"
                :rows="2"
                placeholder="建议的整改措施（可选）"
              />
            </n-form-item>
          </n-grid-item>
          <n-grid-item>
            <n-form-item label="责任单位" path="unitId">
              <n-select
                v-model:value="form.unitId"
                :options="unitOptions"
                placeholder="选择后自动关联责任人"
                filterable
                clearable
              />
            </n-form-item>
          </n-grid-item>
          <n-grid-item>
            <n-form-item label="责任人">
              <n-input :value="linkedPerson" readonly placeholder="由责任单位自动带出" />
            </n-form-item>
          </n-grid-item>
          <n-grid-item>
            <n-form-item label="要求完成时间" path="dueDate">
              <n-date-picker
                :value="dueTs"
                type="date"
                style="width: 100%"
                @update:value="handleDueDateChange"
              />
            </n-form-item>
          </n-grid-item>
          <n-grid-item>
            <n-form-item label="复查人员">
              <n-input v-model:value="form.recheckPerson" placeholder="留空则默认同检查人员" />
            </n-form-item>
          </n-grid-item>
          <n-grid-item>
            <n-form-item label="整改员工">
              <n-input v-model:value="form.rectifyPerson" placeholder="负责整改的员工（可选）" />
            </n-form-item>
          </n-grid-item>
          <n-grid-item>
            <n-form-item label="整改状态" path="status">
              <n-select v-model:value="form.status" :options="statusOptions" />
            </n-form-item>
          </n-grid-item>
          <n-grid-item>
            <n-form-item label="隐患类型" path="typeId">
              <n-select
                v-model:value="form.typeId"
                :options="typeOptions"
                placeholder="先选大类，再选小类"
                filterable
                clearable
                style="width: 100%"
              />
            </n-form-item>
          </n-grid-item>
          <n-grid-item :span="isMobile ? 1 : 2">
            <n-form-item label="整改前图片">
              <ImageUpload v-model="form.beforeImageIds" placeholder="上传整改前图片" />
            </n-form-item>
          </n-grid-item>
          <n-grid-item :span="isMobile ? 1 : 2">
            <n-form-item label="整改后图片">
              <ImageUpload v-model="form.afterImageIds" placeholder="上传整改后图片" />
            </n-form-item>
          </n-grid-item>
          <n-grid-item :span="isMobile ? 1 : 2">
            <n-form-item label="备注">
              <n-input
                v-model:value="form.remark"
                type="textarea"
                :rows="2"
                placeholder="其他说明（可选）"
              />
            </n-form-item>
          </n-grid-item>
        </n-grid>
      </n-form>
    </div>

    <template #footer>
      <div class="modal-footer">
        <div class="modal-footer-left">
          <n-button v-if="isEdit" type="error" secondary @click="confirmDelete">删除</n-button>
        </div>
        <div class="modal-footer-right">
          <n-button @click="close">取消</n-button>
          <n-button type="primary" :loading="saving" @click="handleSubmit">保存</n-button>
        </div>
      </div>
    </template>
  </n-modal>
</template>

<style scoped>
.modal-body {
  max-height: min(70vh, 640px);
  overflow-y: auto;
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
