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
import Credits from '../views/Credits.vue'
import FriendsView from '../views/FriendsView.vue'
import P2PView from '../views/P2PView.vue'
import HomeView from '../views/HomeView.vue'
import UsageDashboard from '../components/usage/UsageDashboard.vue'
import LandingHome from '../views/landing/HomeView.vue'
import LandingPricing from '../views/landing/PricingView.vue'
import LandingTransfer from '../views/landing/TransferView.vue'
import LandingSecurity from '../views/landing/SecurityView.vue'
import { useAuthStore } from '../stores/auth'
import { isP2PSubdomain } from '../composables/useSubdomain'

const isLocalAuthBypassEnabled =
  import.meta.env.DEV && String(import.meta.env.VITE_LOCAL_BYPASS_AUTH).toLowerCase() === 'true'

const routes = [
  {
    path: '/',
    name: 'LandingHome',
    component: isP2PSubdomain
      ? () => import('../views/p2p/P2PSubdomainView.vue')
      : LandingHome,
    meta: isP2PSubdomain ? { requiresAuth: true } : {},
  },
  {
    path: '/pricing',
    name: 'Pricing',
    component: LandingPricing,
  },
  {
    path: '/transfer',
    name: 'Transfer',
    component: LandingTransfer,
  },
  {
    path: '/compare',
    name: 'Compare',
    component: () => import('../views/landing/CompareView.vue'),
  },
  {
    path: '/security',
    name: 'Security',
    component: LandingSecurity,
  },
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
    path: '/credits',
    name: 'Credits',
    component: Credits,
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
      }
    ]
  },
  {
    path: '/usage',
    name: 'Usage',
    component: UsageDashboard,
    meta: { requiresAuth: true }
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
  ...(!isP2PSubdomain ? [{ path: '/', redirect: '/dashboard/home' }] : []),
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach(async (to, from, next) => {
  if (isLocalAuthBypassEnabled) {
    if (to.name === 'Login') {
      next('/dashboard/home')
      return
    }
    next()
    return
  }

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
    next(isP2PSubdomain ? '/' : '/dashboard/home')
  } else {
    next()
  }
})

export default router
