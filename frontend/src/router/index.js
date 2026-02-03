import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from '../views/dashboard.vue'
import Login from '../views/Login.vue'
import Account from '../views/account.vue'
import SharedElements from '../components/sharedElements.vue'
import FileBrowser from '../components/FileBrowser.vue'
import PublicShare from '../views/PublicShare.vue'
import PublicBrowse from '../views/PublicBrowse.vue'
import TermsOfService from '../views/TermsOfService.vue'
import PrivacyPolicy from '../views/PrivacyPolicy.vue'
import FriendsView from '../views/FriendsView.vue'
import P2PView from '../views/P2PView.vue'
import HomeView from '../views/HomeView.vue'
import BillingDashboard from '../components/billing/BillingDashboard.vue'
import { useAuthStore } from '../stores/auth'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: Login,
  },
  {
    path: '/cgu',
    name: 'TermsOfService',
    component: TermsOfService,
  },
  {
    path: '/privacy',
    name: 'PrivacyPolicy',
    component: PrivacyPolicy,
  },
  {
    path: '/dashboard',
    component: Dashboard,
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        redirect: '/dashboard/home'
      },
      {
        path: 'home',
        name: 'Home',
        component: HomeView
      },
      {
        path: 'files',
        name: 'MyFiles',
        component: FileBrowser
      },
      {
        path: 'shares',
        name: 'SharedElements',
        component: SharedElements
      },
      {
        path: 'friends',
        name: 'Friends',
        component: FriendsView
      },
      {
        path: 'billing',
        name: 'Billing',
        component: BillingDashboard
      }
    ]
  },
  {
    path: '/p2p',
    name: 'P2P',
    component: P2PView,
    meta: { requiresAuth: true }
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
    redirect: '/dashboard/home',
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
    next('/dashboard/home')
  } else {
    next()
  }
})

export default router
