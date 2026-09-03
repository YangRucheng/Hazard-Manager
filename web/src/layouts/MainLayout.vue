<script setup lang="ts">
import { computed, h, type Component } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  NLayout,
  NLayoutSider,
  NLayoutHeader,
  NLayoutContent,
  NMenu,
  NIcon,
  NDropdown,
  NAvatar,
  useDialog,
  type MenuOption,
} from 'naive-ui'
import { HomeOutline, WarningOutline, ListOutline, LogOutOutline, PersonCircleOutline } from '@vicons/ionicons5'

import { useAuth } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const dialog = useDialog()
const { username, clearAuth } = useAuth()

function renderIcon(icon: Component) {
  return () => h(NIcon, null, { default: () => h(icon) })
}

const menuOptions: MenuOption[] = [
  { label: '工作台', key: '/dashboard', icon: renderIcon(HomeOutline) },
  { label: '隐患管理', key: '/hazards', icon: renderIcon(WarningOutline) },
  {
    label: '枚举值管理',
    key: '/enums',
    icon: renderIcon(ListOutline),
    children: [
      { label: '责任单位', key: '/enums/units' },
      { label: '隐患类型', key: '/enums/types' },
    ],
  },
]

/** 当前菜单选中项：枚举子路由回退到父级。 */
const activeKey = computed<string>(() => {
  const path = route.path
  if (path.startsWith('/enums/')) {
    return path
  }
  return path
})

const pageTitle = computed<string>(() => String(route.meta.title ?? ''))

function handleMenuSelect(key: string): void {
  if (key !== route.path) {
    void router.push(key)
  }
}

function handleLogout(): void {
  dialog.warning({
    title: '退出登录',
    content: '确定要退出当前账号吗？',
    positiveText: '退出',
    negativeText: '取消',
    onPositiveClick: () => {
      clearAuth()
      void router.push({ name: 'login' })
    },
  })
}
</script>

<template>
  <n-layout class="main-layout" has-sider position="absolute">
    <n-layout-sider bordered :width="220" :native-scrollbar="false">
      <div class="brand">
        <div class="brand-mark">危</div>
        <div>
          <div class="brand-title">电气车间隐患闭环</div>
          <div class="brand-sub">Hazard Closed-Loop</div>
        </div>
      </div>
      <n-menu
        :value="activeKey"
        :options="menuOptions"
        :root-indent="16"
        :indent="24"
        @update:value="handleMenuSelect"
      />
    </n-layout-sider>

    <n-layout>
      <n-layout-header bordered class="topbar">
        <div class="topbar-title">{{ pageTitle }}</div>
        <n-dropdown
          trigger="click"
          placement="bottom-end"
          :options="[{ label: '退出登录', key: 'logout', icon: renderIcon(LogOutOutline) }]"
          @select="handleLogout"
        >
          <div class="user-entry">
            <n-avatar :size="30" round :style="{ background: '#1668dc' }">
              <n-icon><person-circle-outline /></n-icon>
            </n-avatar>
            <span class="user-name">{{ username ?? 'admin' }}</span>
          </div>
        </n-dropdown>
      </n-layout-header>

      <n-layout-content class="content" :native-scrollbar="false">
        <router-view />
      </n-layout-content>
    </n-layout>
  </n-layout>
</template>

<style scoped>
.main-layout {
  height: 100vh;
}

.brand {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 16px 14px;
}

.brand-mark {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  background: linear-gradient(135deg, #1668dc, #4098fc);
  color: #fff;
  font-size: 18px;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
}

.brand-title {
  font-size: 15px;
  font-weight: 600;
  color: #17233d;
}

.brand-sub {
  font-size: 11px;
  color: #8a97ab;
  letter-spacing: 0.4px;
}

.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 56px;
  padding: 0 20px;
}

.topbar-title {
  font-size: 16px;
  font-weight: 600;
  color: #17233d;
}

.user-entry {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
  transition: background 0.2s;
}

.user-entry:hover {
  background: rgba(22, 104, 220, 0.08);
}

.user-name {
  font-size: 14px;
  color: #1f2937;
}

.content {
  padding: 16px;
}
</style>