import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useToast } from '@/composables/useToast'

declare module 'vue-router' {
  interface RouteMeta {
    title?: string
    tab?: boolean
    auth?: boolean
  }
}

const routes: RouteRecordRaw[] = [
  { path: '/', name: 'home', component: () => import('@/pages/Home.vue'), meta: { title: '首页', tab: true } },
  { path: '/search', name: 'search', component: () => import('@/pages/Search.vue'), meta: { title: '找民宿' } },
  { path: '/detail/:id', name: 'detail', component: () => import('@/pages/Detail.vue'), meta: { title: '民宿详情' } },
  {
    path: '/confirm/:id',
    name: 'confirm',
    component: () => import('@/pages/Confirm.vue'),
    meta: { title: '确认订单', auth: true },
  },
  { path: '/pay/:sn', name: 'pay', component: () => import('@/pages/Pay.vue'), meta: { title: '收银台', auth: true } },
  { path: '/orders', name: 'orders', component: () => import('@/pages/Orders.vue'), meta: { title: '订单', tab: true, auth: true } },
  {
    path: '/orders/:sn',
    name: 'orderDetail',
    component: () => import('@/pages/OrderDetail.vue'),
    meta: { title: '订单详情', auth: true },
  },
  { path: '/mine', name: 'mine', component: () => import('@/pages/Mine.vue'), meta: { title: '我的', tab: true } },
  { path: '/login', name: 'login', component: () => import('@/pages/Login.vue'), meta: { title: '登录' } },
  { path: '/:pathMatch(.*)*', redirect: '/' },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior: () => ({ top: 0 }),
})

router.beforeEach((to) => {
  document.title = to.meta.title ? `${to.meta.title} · 栖居` : '栖居 · 民宿预订'

  if (to.meta.auth) {
    const user = useUserStore()
    if (!user.token) {
      useToast().info('请先登录')
      return { name: 'login', query: { redirect: to.fullPath } }
    }
  }
  return true
})

export default router
