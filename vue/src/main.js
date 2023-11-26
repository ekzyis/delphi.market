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

const routes = [
  {
    path: '/', component: HomeView
  },
  {
    path: '/login', component: LoginView
  },
  {
    path: '/user', component: UserView
  },
  {
    path: '/market/:id', component: MarketView
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
