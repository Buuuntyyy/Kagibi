import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from '../views/dashboard.vue'
import Login from '../views/Login.vue'
import Account from '../views/account.vue'
import SharedElements from '../components/sharedElements.vue'
import FileBrowser from '../components/FileBrowser.vue'
import PublicShare from '../views/PublicShare.vue'
import PublicBrowse from '../views/PublicBrowse.vue'
import { useAuthStore } from '../stores/auth'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: Login,
  },
  {
    path: '/dashboard',
    component: Dashboard,
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'Dashboard',
        component: FileBrowser
      },
      {
        path: 'shares',
        name: 'SharedElements',
        component: SharedElements
      }
    ]
  },
  {
    path: '/account',
    name: 'Account',
    component: Account,
    meta: { requiresAuth: true },
  },
  {
    path: '/s/:token',
    name: 'PublicShare',
    component: PublicShare,
  },
  {
    path: '/s/:token/browse/:subpath(.*)*',
    name: 'PublicBrowse',
    component: PublicBrowse,
  },
  {
    path: '/',
    redirect: '/dashboard',
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore()

  // Skip auth check for public routes to avoid 401 errors for non-authenticated users
  if (to.name === 'PublicShare' || to.name === 'PublicBrowse') {
    next()
    return
  }

  const isAuthenticated = await authStore.checkAuth()

  if (to.meta.requiresAuth && !isAuthenticated) {
    next('/login')
  } else if (to.name === 'Login' && isAuthenticated) {
    next('/dashboard')
  } else {
    next()
  }
})

export default router
