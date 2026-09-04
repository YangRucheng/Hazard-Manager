<script setup lang="ts">
import { computed, h, ref, watch, type Component } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  NLayout,
  NLayoutSider,
  NLayoutHeader,
  NLayoutContent,
  NIcon,
  NButton,
  NDrawer,
  NDrawerContent,
  NDropdown,
  NAvatar,
  useDialog,
} from 'naive-ui'
import {
  LogOutOutline,
  PersonCircleOutline,
  MenuOutline,
  ChevronBackOutline,
  ChevronForwardOutline,
} from '@vicons/ionicons5'

import { useAuth } from '@/stores/auth'
import { useIsMobile } from '@/utils/media'
import SiderContent from '@/layouts/SiderContent.vue'

const route = useRoute()
const router = useRouter()
const dialog = useDialog()
const { username, clearAuth } = useAuth()

const isMobile = useIsMobile()
/** 桌面端侧栏折叠状态。 */
const collapsed = ref(false)
/** 移动端抽屉式侧栏开关。 */
const drawerOpen = ref(false)

function renderIcon(icon: Component) {
  return () => h(NIcon, null, { default: () => h(icon) })
}

const pageTitle = computed<string>(() => String(route.meta.title ?? ''))

function handleMenuSelect(key: string): void {
  drawerOpen.value = false
  if (key !== route.path) {
    void router.push(key)
  }
}

// 视口从移动端切回桌面端时，收起抽屉。
watch(isMobile, (mobile) => {
  if (!mobile) {
    drawerOpen.value = false
  }
})

function toggleSider(): void {
  if (isMobile.value) {
    drawerOpen.value = !drawerOpen.value
  } else {
    collapsed.value = !collapsed.value
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
    <n-layout-sider
      v-if="!isMobile"
      bordered
      :width="220"
      :collapsed-width="64"
      :native-scrollbar="false"
      collapse-mode="width"
      show-trigger="bar"
      :collapsed="collapsed"
      @update:collapsed="(v: boolean) => (collapsed = v)"
    >
      <SiderContent :collapsed="collapsed" @select="handleMenuSelect" />
    </n-layout-sider>

    <n-layout>
      <n-layout-header bordered class="topbar">
        <div class="topbar-left">
          <n-button quaternary circle class="topbar-toggle" @click="toggleSider">
            <template #icon>
              <n-icon>
                <menu-outline v-if="isMobile" />
                <chevron-back-outline v-else-if="collapsed" />
                <chevron-forward-outline v-else />
              </n-icon>
            </template>
          </n-button>
          <div class="topbar-title">{{ pageTitle }}</div>
        </div>
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

    <n-drawer v-model:show="drawerOpen" :width="260" placement="left">
      <n-drawer-content :body-content-style="{ padding: '0' }" closable>
        <SiderContent @select="handleMenuSelect" />
      </n-drawer-content>
    </n-drawer>
  </n-layout>
</template>

<style scoped>
.main-layout {
  height: 100vh;
}

.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 56px;
  padding: 0 12px 0 8px;
}

.topbar-left {
  display: flex;
  align-items: center;
  gap: 4px;
  min-width: 0;
}

.topbar-title {
  font-size: 16px;
  font-weight: 600;
  color: #17233d;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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
  padding: 12px;
}

@media (max-width: 820px) {
  .topbar {
    padding: 0 8px;
  }

  .user-name {
    display: none;
  }

  .content {
    padding: 0;
  }
}
</style>
