import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api/auth'
import { http } from '@/api/http'
import type { UserInfo } from '@/types/auth'

const ACCESS_TOKEN_KEY = 'crm_access_token'
const REFRESH_TOKEN_KEY = 'crm_refresh_token'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<UserInfo | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const isAuthenticated = computed(() => !!getAccessToken())
  const displayName = computed(() => {
    if (!user.value) return ''
    if (user.value.firstName || user.value.lastName) {
      return `${user.value.firstName} ${user.value.lastName}`.trim()
    }
    return user.value.username
  })

  function getAccessToken(): string | null {
    return localStorage.getItem(ACCESS_TOKEN_KEY)
  }

  function getRefreshToken(): string | null {
    return localStorage.getItem(REFRESH_TOKEN_KEY)
  }

  function setTokens(accessToken: string, refreshToken: string) {
    localStorage.setItem(ACCESS_TOKEN_KEY, accessToken)
    localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken)
    http.setToken(accessToken)
  }

  function clearTokens() {
    localStorage.removeItem(ACCESS_TOKEN_KEY)
    localStorage.removeItem(REFRESH_TOKEN_KEY)
    http.setToken(null)
    user.value = null
  }

  function initialize() {
    const token = getAccessToken()
    if (token) {
      http.setToken(token)
    }
  }

  async function login(username: string, password: string) {
    loading.value = true
    error.value = null
    try {
      const response = await authApi.login({ username, password })
      setTokens(response.data.accessToken, response.data.refreshToken)
      await fetchMe()
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Login error'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function fetchMe() {
    try {
      const response = await authApi.me()
      user.value = response.data
    } catch {
      user.value = null
    }
  }

  async function refresh(): Promise<boolean> {
    const refreshToken = getRefreshToken()
    if (!refreshToken) return false

    try {
      const response = await authApi.refresh({ refreshToken })
      setTokens(response.data.accessToken, response.data.refreshToken)
      return true
    } catch {
      clearTokens()
      return false
    }
  }

  async function logout() {
    const refreshToken = getRefreshToken()
    if (refreshToken) {
      try {
        await authApi.logout(refreshToken)
      } catch {
        // Ignore logout errors
      }
    }
    clearTokens()
  }

  return {
    user,
    loading,
    error,
    isAuthenticated,
    displayName,
    initialize,
    login,
    fetchMe,
    refresh,
    logout,
    getAccessToken,
    clearTokens,
  }
})
