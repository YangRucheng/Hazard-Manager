<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { NButton, NCard, NInput } from 'naive-ui'

import { client } from '@/api/client'
import { BUILD_TIME } from '@/api/config'
import { formatIsoDateTime } from '@/utils/date'

import type { components } from '@/api/schema'

type SystemInfo = components['schemas']['SystemInfo']

/** 开源仓库地址（About 页展示与打开）。 */
const REPO_URL = 'https://github.com/YangRucheng/Hazard-Manager'

const info = ref<SystemInfo | null>(null)

const frontTime = computed<string>(() => formatIsoDateTime(BUILD_TIME) || '—')
/** 后端编译时间：CI 注入后才有效；unknown / 接口不可用时不展示假时间。 */
const backBuildTime = computed<string>(() => {
  const t = info.value?.buildTime
  if (!t || t === 'unknown') {
    return '—'
  }
  return formatIsoDateTime(t) || '—'
})
const backStartTime = computed<string>(() => formatIsoDateTime(info.value?.startTime) || '—')

function openRepo(): void {
  window.open(REPO_URL, '_blank', 'noopener')
}

onMounted(async () => {
  // 旧后端 / 未登录态下接口不可用时不阻塞页面，回退为「—」。
  const { data } = await client.GET('/system/info')
  info.value = data ?? null
})
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">关于</h1>
      </div>
    </div>

    <n-card class="data-card">
      <div class="about-hero">
        <img class="about-logo" src="/logo.png" alt="电气车间隐患闭环系统" />
        <div>
          <div class="about-name">电气车间隐患闭环系统</div>
          <div class="about-tag">隐患登记 · 责任整改 · 复查验收，全流程闭环跟进</div>
        </div>
      </div>

      <div class="about-section">
        <h2 class="about-heading">构建信息</h2>
        <div class="about-row">
          <span class="about-label">前端打包时间</span>
          <span class="about-value">{{ frontTime }}</span>
        </div>
        <div class="about-row">
          <span class="about-label">后端编译时间</span>
          <span class="about-value">{{ backBuildTime }}</span>
        </div>
        <div class="about-row">
          <span class="about-label">后端启动时间</span>
          <span class="about-value">{{ backStartTime }}</span>
        </div>
      </div>

      <div class="about-section">
        <h2 class="about-heading">开源仓库</h2>
        <div class="about-row">
          <span class="about-label">源码地址</span>
          <div class="repo-box">
            <n-input :value="REPO_URL" readonly />
            <n-button type="primary" secondary @click="openRepo">在 GitHub 打开</n-button>
          </div>
        </div>
      </div>
    </n-card>
  </div>
</template>

<style scoped>
.about-hero {
  display: flex;
  align-items: center;
  gap: 16px;
}

.about-logo {
  width: 56px;
  height: 56px;
  object-fit: contain;
  flex: none;
}

.about-name {
  color: var(--color-text-strong);
  font-size: 20px;
  font-weight: 650;
}

.about-tag {
  margin-top: 4px;
  color: var(--color-text);
  font-size: 14px;
}

.about-section {
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid var(--color-border-subtle);
}

.about-heading {
  margin: 0 0 8px;
  color: var(--color-text-strong);
  font-size: 15px;
  font-weight: 600;
}

.about-row {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 8px 0;
}

.about-label {
  width: 128px;
  flex: none;
  color: var(--color-text);
}

.about-value {
  color: var(--color-text-strong);
}

.repo-box {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
  flex: 1;
}

.repo-box .n-input {
  flex: 1;
  min-width: 0;
}
</style>
