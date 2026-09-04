import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

import { useAuth } from '@/stores/auth'

import MainLayout from '@/layouts/MainLayout.vue'
import LoginView from '@/views/LoginView.vue'
import DashboardView from '@/views/DashboardView.vue'
import HazardListView from '@/views/HazardListView.vue'
import HazardFormView from '@/views/HazardFormView.vue'
import UnitManage from '@/views/system/UnitManage.vue'
import TypeManage from '@/views/system/TypeManage.vue'
import AttachmentManage from '@/views/system/AttachmentManage.vue'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    component: LoginView,
    meta: { public: true, title: '登录' },
  },
  {
    path: '/',
    component: MainLayout,
    children: [
      { path: '', redirect: '/dashboard' },
      { path: 'dashboard', name: 'dashboard', component: DashboardView, meta: { title: '工作台' } },
      { path: 'hazards', name: 'hazards', component: HazardListView, meta: { title: '隐患管理' } },
      { path: 'hazards/new', name: 'hazard-create', component: HazardFormView, meta: { title: '新增隐患' } },
      { path: 'hazards/:id/edit', name: 'hazard-edit', component: HazardFormView, meta: { title: '编辑隐患' } },
      { path: 'system/units', name: 'system-units', component: UnitManage, meta: { title: '责任单位' } },
      { path: 'system/types', name: 'system-types', component: TypeManage, meta: { title: '隐患类型' } },
      {
        path: 'system/attachments',
        name: 'system-attachments',
        component: AttachmentManage,
        meta: { title: '附件管理' },
      },
      { path: 'system', redirect: '/system/units' },
      // 旧「枚举值管理」路径重定向到「系统管理」。
      { path: 'enums', redirect: '/system/units' },
      { path: 'enums/units', redirect: '/system/units' },
      { path: 'enums/types', redirect: '/system/types' },
    ],
  },
  { path: '/:pathMatch(.*)*', redirect: '/dashboard' },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// 登录守卫：未登录跳转登录页（保留回跳地址）。
router.beforeEach((to) => {
  const { isLoggedIn } = useAuth()
  const requiresAuth = to.meta.public !== true
  if (requiresAuth && !isLoggedIn.value) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }
  if (to.name === 'login' && isLoggedIn.value) {
    return { name: 'dashboard' }
  }
  return true
})

router.afterEach((to) => {
  const title = to.meta.title
  document.title = title ? `${String(title)} · 电气车间隐患闭环` : '电气车间隐患闭环'
})

export default router