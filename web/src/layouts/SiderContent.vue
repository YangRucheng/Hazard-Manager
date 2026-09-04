<script setup lang="ts">
import { computed, h, type Component } from 'vue'
import { useRoute } from 'vue-router'
import { NMenu, NIcon, type MenuOption } from 'naive-ui'
import { HomeOutline, WarningOutline, OptionsOutline } from '@vicons/ionicons5'

const props = defineProps<{
  /** 桌面端折叠状态（抽屉中恒为展开）。 */
  collapsed?: boolean
}>()

defineEmits<{ select: [key: string] }>()

const route = useRoute()

function renderIcon(icon: Component) {
  return () => h(NIcon, null, { default: () => h(icon) })
}

const menuOptions: MenuOption[] = [
  { label: '工作台', key: '/dashboard', icon: renderIcon(HomeOutline) },
  { label: '隐患管理', key: '/hazards', icon: renderIcon(WarningOutline) },
  {
    label: '系统管理',
    key: '/system',
    icon: renderIcon(OptionsOutline),
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
    <div class="brand" :class="{ 'brand--collapsed': isCollapsed }">
      <img class="brand-logo" src="/logo.png" alt="电气车间隐患闭环" />
      <div v-show="!isCollapsed" class="brand-text">
        <div class="brand-title">电气车间隐患闭环</div>
        <div class="brand-sub">Hazard Closed-Loop</div>
      </div>
    </div>
    <n-menu
      :value="activeKey"
      :options="menuOptions"
      :root-indent="16"
      :indent="24"
      :collapsed="isCollapsed"
      :collapsed-width="64"
      :collapsed-icon-size="22"
      @update:value="(key: string) => $emit('select', key)"
    />
  </div>
</template>

<style scoped>
.sider-content {
  height: 100%;
}

.brand {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 16px 14px;
}

.brand--collapsed {
  justify-content: center;
  padding: 16px 8px;
}

.brand-logo {
  width: 36px;
  height: 36px;
  object-fit: contain;
  flex-shrink: 0;
}

.brand-text {
  min-width: 0;
}

.brand-title {
  font-size: 15px;
  font-weight: 600;
  color: #17233d;
  white-space: nowrap;
}

.brand-sub {
  font-size: 12px;
  color: #5a7fd0;
  letter-spacing: 0.4px;
  white-space: nowrap;
}
</style>
