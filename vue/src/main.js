import { createApp } from 'vue'
import { createPinia } from 'pinia'
import * as VueRouter from 'vue-router'
import App from './App.vue'
import './registerServiceWorker'
import './index.css'

import HomeView from '@/components/HomeView'
import LoginView from '@/components/LoginView'
import UserView from '@/components/UserView'

const routes = [
  {
    path: '/', component: HomeView
  },
  {
    path: '/login', component: LoginView
  },
  {
    path: '/user', component: UserView
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