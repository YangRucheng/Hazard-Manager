<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  NButton,
  NCard,
  NDatePicker,
  NForm,
  NFormItem,
  NGrid,
  NGridItem,
  NInput,
  NSelect,
  useMessage,
  type FormInst,
  type FormRules,
  type SelectOption,
  type SelectGroupOption,
} from 'naive-ui'

import { client, errorMessage } from '@/api/client'
import ImageUpload from '@/components/ImageUpload.vue'
import { formatDate } from '@/utils/date'

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
  beforeImageIds: string[]
  status: HazardStatus
  afterImageIds: string[]
  typeId: number | null
  level: HazardLevel
  remark: string
}

const route = useRoute()
const router = useRouter()
const message = useMessage()

const isEdit = computed<boolean>(() => route.name === 'hazard-edit')
const hazardId = computed<number>(() => Number(route.params.id))

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

async function loadDetail(): Promise<void> {
  if (!isEdit.value) {
    return
  }
  loading.value = true
  try {
    const { data, error } = await client.GET('/hazards/{id}', {
      params: { path: { id: hazardId.value } },
    })
    if (error || !data) {
      message.error(errorMessage(error))
      void router.replace({ name: 'hazards' })
      return
    }
    applyDetail(data)
  } finally {
    loading.value = false
  }
}

function applyDetail(h: Hazard): void {
  form.value = {
    inspectionArea: h.inspectionArea,
    inspector: h.inspector,
    description: h.description,
    suggestion: h.suggestion ?? '',
    unitId: h.unitId,
    recheckPerson: h.recheckPerson ?? '',
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
      beforeImageIds: form.value.beforeImageIds,
      status: form.value.status,
      afterImageIds: form.value.afterImageIds,
      typeId: form.value.typeId,
      level: form.value.level,
      remark: form.value.remark.trim() || null,
    }

    if (isEdit.value) {
      const { data, error } = await client.PUT('/hazards/{id}', {
        params: { path: { id: hazardId.value } },
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
    void router.push({ name: 'hazards' })
  } finally {
    saving.value = false
  }
}

function handleCancel(): void {
  void router.push({ name: 'hazards' })
}

function startOfToday(): number {
  const d = new Date()
  d.setHours(0, 0, 0, 0)
  return d.getTime()
}

function addDaysTimestamp(ts: number, days: number): number {
  return ts + days * 24 * 60 * 60 * 1000
}

onMounted(() => {
  void loadEnums()
  void loadDetail()
})
</script>

<template>
  <div class="page">
    <n-card :bordered="true" :title="isEdit ? '编辑隐患' : '新增隐患'">
      <n-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-placement="left"
        label-width="110"
        :disabled="loading"
        style="max-width: 1080px"
      >
        <n-grid :cols="2" :x-gap="24">
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
          <n-grid-item :span="2">
            <n-form-item label="隐患描述" path="description">
              <n-input
                v-model:value="form.description"
                type="textarea"
                :rows="3"
                placeholder="请详细描述隐患情况"
              />
            </n-form-item>
          </n-grid-item>
          <n-grid-item :span="2">
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
              <n-input
                v-model:value="form.recheckPerson"
                placeholder="留空则默认同检查人员"
              />
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
          <n-grid-item :span="2">
            <n-form-item label="整改前图片">
              <ImageUpload v-model="form.beforeImageIds" placeholder="上传整改前图片" />
            </n-form-item>
          </n-grid-item>
          <n-grid-item :span="2">
            <n-form-item label="整改后图片">
              <ImageUpload v-model="form.afterImageIds" placeholder="上传整改后图片" />
            </n-form-item>
          </n-grid-item>
          <n-grid-item :span="2">
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

        <div class="form-actions">
          <n-button type="primary" :loading="saving" @click="handleSubmit">保存</n-button>
          <n-button @click="handleCancel">取消</n-button>
        </div>
      </n-form>
    </n-card>
  </div>
</template>

<style scoped>
.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 8px;
  padding-top: 16px;
  border-top: 1px solid #eef3fa;
}
</style>