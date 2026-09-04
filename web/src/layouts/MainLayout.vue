<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  NLayout,
  NLayoutSider,
  NLayoutHeader,
  NLayoutContent,
  NIcon,
  NButton,
  NBreadcrumb,
  NBreadcrumbItem,
  NDrawer,
  NDrawerContent,
  NDropdown,
  useDialog,
} from 'naive-ui'
import { MenuOutline } from '@vicons/ionicons5'

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

const pageTitle = computed<string>(() => String(route.meta.title ?? ''))

// 视口从移动端切回桌面端时，收起抽屉。
watch(isMobile, (mobile) => {
  if (!mobile) {
    drawerOpen.value = false
  }
})

/** 内容区留白：对齐参考系统（桌面 24/28/32，移动端收紧）。 */
const contentStyle = computed<string>(() =>
  isMobile.value ? 'padding: 14px 12px 24px;' : 'padding: 24px 28px 32px;',
)

function handleMenuSelect(key: string): void {
  drawerOpen.value = false
  if (key !== route.path) {
    void router.push(key)
  }
}

function toggleSider(): void {
  if (isMobile.value) {
    drawerOpen.value = true
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
  <n-layout has-sider class="app-shell">
    <n-layout-sider
      v-if="!isMobile"
      bordered
      collapse-mode="width"
      :collapsed-width="64"
      :width="220"
      :collapsed="collapsed"
      show-trigger
      @collapse="collapsed = true"
      @expand="collapsed = false"
    >
      <SiderContent :collapsed="collapsed" @select="handleMenuSelect" />
    </n-layout-sider>

    <n-layout class="app-main">
      <n-layout-header bordered class="topbar">
        <div class="topbar-left">
          <n-button
            v-if="isMobile"
            quaternary
            circle
            class="topbar-toggle"
            aria-label="打开导航菜单"
            @click="toggleSider"
          >
            <template #icon>
              <n-icon><menu-outline /></n-icon>
            </template>
          </n-button>
          <n-breadcrumb v-if="!isMobile">
            <n-breadcrumb-item>隐患闭环</n-breadcrumb-item>
            <n-breadcrumb-item v-if="pageTitle">{{ pageTitle }}</n-breadcrumb-item>
          </n-breadcrumb>
          <div v-if="isMobile && pageTitle" class="topbar-title">{{ pageTitle }}</div>
        </div>

        <n-dropdown
          trigger="click"
          placement="bottom-end"
          :options="[{ label: '退出登录', key: 'logout' }]"
          @select="handleLogout"
        >
          <button type="button" class="user-menu-trigger" aria-label="打开用户菜单">
            <span class="user-name">{{ username ?? 'admin' }}</span>
            <span class="user-menu-caret" aria-hidden="true" />
          </button>
        </n-dropdown>
      </n-layout-header>

      <n-layout-content
        class="app-content"
        :content-style="contentStyle"
        :native-scrollbar="false"
      >
        <router-view />
      </n-layout-content>
    </n-layout>

    <n-drawer v-model:show="drawerOpen" :width="240" placement="left">
      <n-drawer-content :body-content-style="{ padding: '0' }" closable>
        <SiderContent @select="handleMenuSelect" />
      </n-drawer-content>
    </n-drawer>
  </n-layout>
</template>

<style scoped>
.app-shell {
  height: 100vh;
  background: var(--color-bg);
}

.app-main {
  background: var(--color-bg);
}

.topbar {
  height: 68px;
  padding: 0 28px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  background: rgb(255 255 255 / 92%);
  backdrop-filter: blur(12px);
}

.topbar-left {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.topbar-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text-strong);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-menu-trigger {
  min-height: 40px;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 4px 13px 4px 15px;
  border: 1px solid transparent;
  border-radius: 12px;
  color: inherit;
  background: transparent;
  cursor: pointer;
  transition:
    background-color 0.2s ease,
    border-color 0.2s ease;
}

.user-menu-trigger:hover,
.user-menu-trigger:focus-visible {
  border-color: #dce5ff;
  background: var(--color-primary-soft);
  outline: none;
}

.user-name {
  color: var(--color-text-strong);
  font-size: 14px;
  font-weight: 600;
  line-height: 1.25;
  max-width: 160px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-menu-caret {
  width: 7px;
  height: 7px;
  margin-top: -3px;
  flex: none;
  border-right: 1.5px solid var(--color-text-muted);
  border-bottom: 1.5px solid var(--color-text-muted);
  transform: rotate(45deg);
}

.app-content {
  background:
    radial-gradient(circle at 100% 0%, rgb(63 99 216 / 4%), transparent 28%), var(--color-bg);
}

@media (max-width: 820px) {
  .topbar {
    height: 56px;
    padding: 0 12px;
  }

  .app-content {
    background: var(--color-bg);
  }
}
</style>
