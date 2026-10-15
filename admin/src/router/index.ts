import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useToast } from '@/composables/useToast'

declare module 'vue-router' {
  interface RouteMeta {
    title?: string
    subtitle?: string
    auth?: boolean
  }
}

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/Login.vue'),
    meta: { title: '登录' },
  },
  {
    path: '/',
    component: () => import('@/components/AppLayout.vue'),
    meta: { auth: true },
    children: [
      {
        path: '',
        name: 'dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '数据看板', subtitle: '今日经营概览' },
      },
      {
        path: 'homestays',
        name: 'homestays',
        component: () => import('@/views/Homestays.vue'),
        meta: { title: '民宿管理', subtitle: '上下架、价格与库存' },
      },
      {
        path: 'orders',
        name: 'orders',
        component: () => import('@/views/Orders.vue'),
        meta: { title: '订单管理', subtitle: '查询、详情与强制取消' },
      },
      {
        path: 'users',
        name: 'users',
        component: () => import('@/views/Users.vue'),
        meta: { title: '用户管理', subtitle: '用户列表与禁用' },
      },
      {
        path: 'comments',
        name: 'comments',
        component: () => import('@/views/Comments.vue'),
        meta: { title: '评价管理', subtitle: '评价审核与删除' },
      },
      {
        path: 'agent',
        name: 'agent',
        component: () => import('@/views/Agent.vue'),
        meta: { title: 'AI 助手', subtitle: '用自然语言查数据、办业务' },
      },
      {
        path: 'knowledge',
        name: 'knowledge',
        component: () => import('@/views/Knowledge.vue'),
        meta: { title: '知识库', subtitle: 'AI 助手的回答依据' },
      },
      {
        path: 'monitor',
        name: 'monitor',
        component: () => import('@/views/Monitor.vue'),
        meta: { title: 'AI 值守', subtitle: '异常巡检与经营日报' },
      },
    ],
  },
  { path: '/:pathMatch(.*)*', redirect: '/' },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior: () => ({ top: 0 }),
})

router.beforeEach((to) => {
  document.title = to.meta.title ? `${to.meta.title} · 栖居后台` : '栖居 · 后台管理'

  if (to.meta.auth) {
    const auth = useAuthStore()
    if (!auth.token) {
      useToast().info('请先登录后台')
      return { name: 'login', query: { redirect: to.fullPath } }
    }
  }
  if (to.name === 'login') {
    const auth = useAuthStore()
    if (auth.token) return { name: 'dashboard' }
  }
  return true
})

export default router
