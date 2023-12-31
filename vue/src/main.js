import { createApp } from 'vue'
import { createPinia } from 'pinia'
import * as VueRouter from 'vue-router'
import App from './App.vue'
import './registerServiceWorker'
import './index.css'

import HomeView from '@/views/HomeView'
import AboutView from '@/views/AboutView'
import LoginView from '@/views/LoginView'
import UserView from '@/views/UserView'
import MarketView from '@/views/MarketView'
import InvoiceView from '@/views/InvoiceView'
import UserWallet from '@/components/UserWallet'
import UserInvoices from '@/components/UserInvoices'
import UserOrders from '@/components/UserOrders'
import OrderForm from '@/components/OrderForm'
import MarketOrders from '@/components/MarketOrders'
import MarketStats from '@/components/MarketStats'
import MarketSettings from '@/components/MarketSettings'

const routes = [
  {
    path: '/', component: HomeView
  },
  {
    path: '/about', component: AboutView
  },
  {
    path: '/login', component: LoginView
  },
  {
    path: '/user',
    component: UserView,
    children: [
      { path: 'wallet', name: 'user', component: UserWallet },
      { path: 'invoices', name: 'invoices', component: UserInvoices },
      { path: 'orders', name: 'orders', component: UserOrders }
    ]
  },
  {
    path: '/market/:id',
    component: MarketView,
    children: [
      { path: 'form', name: 'form', component: OrderForm },
      { path: 'orders', name: 'market-orders', component: MarketOrders },
      { path: 'stats', name: 'market-stats', component: MarketStats },
      { path: 'settings', name: 'market-settings', component: MarketSettings }
    ]
  },
  {
    path: '/invoice/:id', component: InvoiceView
  }
]
const router = VueRouter.createRouter({
  history: VueRouter.createWebHashHistory(),
  routes
})

const pinia = createPinia()

const app = createApp(App)
app.use(router)
app.use(pinia)
app.mount('#app')
