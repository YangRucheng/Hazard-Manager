<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NCard, NForm, NFormItem, NInput, NButton, useMessage, type FormInst, type FormRules } from 'naive-ui'

import { client, errorMessage } from '@/api/client'
import { useAuth } from '@/stores/auth'

interface LoginForm {
  username: string
  password: string
}

const router = useRouter()
const route = useRoute()
const message = useMessage()
const { setAuth } = useAuth()

const formRef = ref<FormInst | null>(null)
const loading = ref(false)
const form = ref<LoginForm>({ username: 'admin', password: '' })

const rules: FormRules = {
  username: { required: true, message: '请输入账号', trigger: ['input', 'blur'] },
  password: { required: true, message: '请输入密码', trigger: ['input', 'blur'] },
}

async function handleSubmit(): Promise<void> {
  loading.value = true
  try {
    try {
      await formRef.value?.validate()
    } catch {
      return // 校验不通过
    }
    const { data, error } = await client.POST('/auth/login', {
      body: { username: form.value.username.trim(), password: form.value.password },
    })
    if (error || !data) {
      message.error(errorMessage(error))
      return
    }
    setAuth(data.token, data.user.username)
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/dashboard'
    void router.replace(redirect)
    message.success('登录成功')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <main class="login-page">
    <section class="login-intro">
      <div class="intro-content">
        <img class="login-logo" src="/logo.png" alt="电气车间隐患闭环系统" />
        <span class="eyebrow">ELECTRICAL WORKSHOP</span>
        <h1>电气车间<br />隐患闭环管理</h1>
      </div>
    </section>
    <section class="login-panel">
      <n-card class="login-card" :bordered="false">
        <h2>欢迎登录</h2>
        <n-form ref="formRef" :model="form" :rules="rules" size="large" @submit.prevent="handleSubmit">
          <n-form-item label="账号" path="username">
            <n-input v-model:value="form.username" placeholder="请输入账号" />
          </n-form-item>
          <n-form-item label="密码" path="password">
            <n-input
              v-model:value="form.password"
              type="password"
              show-password-on="click"
              placeholder="请输入密码"
              @keyup.enter="handleSubmit"
            />
          </n-form-item>
          <n-button type="primary" block size="large" :loading="loading" @click="handleSubmit">
            登录
          </n-button>
        </n-form>
      </n-card>
    </section>
  </main>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  min-height: 100dvh;
  display: grid;
  grid-template-columns: 1.15fr 1fr;
  background: var(--color-bg);
}

.login-intro {
  background:
    radial-gradient(circle at 18% 20%, rgb(255 255 255 / 14%), transparent 24%),
    linear-gradient(145deg, #2947a6, #5579e7);
  color: #fff;
  display: grid;
  place-items: center;
  position: relative;
  overflow: hidden;
}

.login-intro::after {
  content: '';
  position: absolute;
  width: 480px;
  height: 480px;
  border: 1px solid rgb(255 255 255 / 18%);
  border-radius: 50%;
  right: -120px;
  bottom: -160px;
  box-shadow:
    0 0 0 80px rgb(255 255 255 / 4%),
    0 0 0 160px rgb(255 255 255 / 3%);
}

.intro-content {
  z-index: 1;
  max-width: 500px;
  padding: 48px;
}

.eyebrow {
  letter-spacing: 3px;
  opacity: 0.75;
}

.login-logo {
  display: block;
  width: 88px;
  height: 88px;
  margin-bottom: 24px;
  object-fit: contain;
}

h1 {
  font-size: 48px;
  line-height: 1.25;
  margin: 20px 0 0;
}

.login-panel {
  display: grid;
  place-items: center;
  padding: 48px;
  background:
    radial-gradient(circle at 100% 0%, rgb(63 99 216 / 7%), transparent 32%), var(--color-bg);
}

.login-card {
  width: 420px;
  border: 1px solid var(--color-border-subtle);
  box-shadow: 0 18px 44px rgb(15 23 42 / 10%);
}

.login-card h2 {
  font-size: 28px;
  margin: 0 0 22px;
  color: var(--color-text-strong);
}

@media (max-width: 900px) {
  .login-page {
    grid-template-columns: 1fr;
  }

  .login-intro {
    min-height: 220px;
  }

  h1 {
    font-size: 32px;
  }

  .intro-content {
    padding: 24px;
    text-align: center;
  }

  .login-logo {
    width: 64px;
    height: 64px;
    margin: 0 auto 16px;
  }

  .login-panel {
    padding: 24px 16px;
  }

  .login-card {
    width: min(420px, 100%);
  }
}
</style>
