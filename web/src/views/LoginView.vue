<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NForm, NFormItem, NInput, NButton, useMessage, type FormInst, type FormRules } from 'naive-ui'

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
  <div class="login-page">
    <div class="login-card">
      <div class="login-head">
        <img class="login-logo" src="/logo.png" alt="电气车间隐患闭环系统" />
        <h1 class="login-title">电气车间隐患闭环系统</h1>
        <p class="login-sub">Hazard Closed-Loop Management</p>
      </div>
      <n-form ref="formRef" :model="form" :rules="rules" size="large" @keyup.enter="handleSubmit">
        <n-form-item label="账号" path="username">
          <n-input v-model:value="form.username" placeholder="请输入账号" />
        </n-form-item>
        <n-form-item label="密码" path="password">
          <n-input
            v-model:value="form.password"
            type="password"
            show-password-on="mousedown"
            placeholder="请输入密码"
          />
        </n-form-item>
        <n-button class="login-btn" type="primary" block :loading="loading" @click="handleSubmit">
          登 录
        </n-button>
      </n-form>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background:
    radial-gradient(1200px 500px at 20% -10%, rgba(64, 152, 252, 0.35), transparent 60%),
    radial-gradient(1000px 500px at 90% 110%, rgba(22, 104, 220, 0.3), transparent 55%),
    #eef4fb;
}

.login-card {
  width: 380px;
  padding: 40px 36px 32px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 12px 40px rgba(22, 104, 220, 0.12);
  border: 1px solid #dbe5f1;
}

.login-head {
  text-align: center;
  margin-bottom: 28px;
}

.login-logo {
  width: 52px;
  height: 52px;
  display: block;
  margin: 0 auto 12px;
  object-fit: contain;
}

.login-title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #17233d;
}

.login-sub {
  margin: 6px 0 0;
  font-size: 12px;
  color: #8a97ab;
  letter-spacing: 1px;
}

.login-btn {
  margin-top: 8px;
}
</style>