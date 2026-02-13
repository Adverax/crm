import './assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
import { useAuthStore } from './stores/auth'
import { http } from './api/http'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)

// Initialize auth: restore token from localStorage and set refresh handler
const authStore = useAuthStore()
authStore.initialize()
http.setRefreshHandler(() => authStore.refresh())

app.mount('#app')
