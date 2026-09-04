<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { NMenu, type MenuOption } from 'naive-ui'

const props = defineProps<{
  /** 桌面端折叠状态（抽屉中恒为展开）。 */
  collapsed?: boolean
}>()

defineEmits<{ select: [key: string] }>()

const route = useRoute()

const menuOptions: MenuOption[] = [
  { label: '工作台', key: '/dashboard' },
  { label: '隐患管理', key: '/hazards' },
  {
    label: '系统管理',
    key: '/system',
    children: [
      { label: '责任单位', key: '/system/units' },
      { label: '隐患类型', key: '/system/types' },
      { label: '附件管理', key: '/system/attachments' },
    ],
  },
]

/** 当前菜单选中项：与路由 path 对齐。 */
const activeKey = computed<string>(() => route.path)

/** 折叠仅作用于桌面端侧栏（抽屉内 collapsed prop 传 false）。 */
const isCollapsed = computed<boolean>(() => props.collapsed === true)
</script>

<template>
  <div class="sider-content">
    <div class="brand" :class="{ 'brand--compact': isCollapsed }">
      <img class="brand-logo" src="/logo.png" alt="电气车间隐患闭环" />
      <span v-if="!isCollapsed" class="brand-name">电气车间隐患闭环</span>
    </div>
    <n-menu
      :value="activeKey"
      :options="menuOptions"
      :collapsed="isCollapsed"
      :collapsed-width="64"
      :collapsed-icon-size="20"
      @update:value="(key: string) => $emit('select', key)"
    />
  </div>
</template>

<style scoped>
.sider-content {
  height: 100%;
}

.brand {
  height: 68px;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 16px;
  border-bottom: 1px solid var(--color-border-subtle);
  color: var(--color-text-strong);
  font-size: 16px;
  font-weight: 650;
  white-space: nowrap;
}

.brand--compact {
  justify-content: center;
  padding: 0;
}

.brand-logo {
  width: 34px;
  height: 34px;
  object-fit: contain;
  flex: none;
}
</style>
