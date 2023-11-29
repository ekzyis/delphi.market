import { createApp } from 'vue'
import { createPinia } from 'pinia'
import * as VueRouter from 'vue-router'
import App from './App.vue'
import './registerServiceWorker'
import './index.css'

import HomeView from '@/views/HomeView'
import LoginView from '@/views/LoginView'
import UserView from '@/views/UserView'
import MarketView from '@/views/MarketView'
import InvoiceView from '@/views/InvoiceView'
import UserSettings from '@/components/UserSettings'
import UserInvoices from '@/components/UserInvoices'
import UserOrders from '@/components/UserOrders'
import OrderForm from '@/components/OrderForm'
import MarketOrders from '@/components/MarketOrders'
import MarketStats from '@/components/MarketStats'
import BuyOrderForm from '@/components/BuyOrderForm'
import SellOrderForm from '@/components/SellOrderForm'

const routes = [
  {
    path: '/', component: HomeView
  },
  {
    path: '/login', component: LoginView
  },
  {
    path: '/user',
    component: UserView,
    children: [
      { path: 'settings', name: 'user', component: UserSettings },
      { path: 'invoices', name: 'invoices', component: UserInvoices },
      { path: 'orders', name: 'orders', component: UserOrders }
    ]
  },
  {
    path: '/market/:id',
    component: MarketView,
    children: [
      {
        path: 'form',
        name: 'form',
        component: OrderForm,
        children: [
          {
            path: 'buy', name: 'form-buy', component: BuyOrderForm
          },
          {
            path: 'sell', name: 'form-sell', component: SellOrderForm
          }
        ]
      },
      { path: 'orders', name: 'market-orders', component: MarketOrders },
      { path: 'stats', name: 'market-stats', component: MarketStats }
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
