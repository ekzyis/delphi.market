import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

export const useSession = defineStore('session', () => {
  const pubkey = ref(null)
  const initialized = ref(false)
  const isAuthenticated = computed(() => initialized.value ? !!pubkey.value : undefined)

  function checkSession () {
    const url = window.origin + '/api/session'
    return fetch(url, {
      credentials: 'include'
    })
      .then(async r => {
        const body = await r.json()
        if (body.pubkey) {
          pubkey.value = body.pubkey
          console.log('authenticated as', body.pubkey)
        } else console.log('unauthenticated')
        initialized.value = true
        return body
      }).catch(err => {
        console.error('error:', err.reason || err)
      })
  }

  function login () {
    const url = window.origin + '/api/login'
    return fetch(url, { credentials: 'include' }).then(r => r.json())
  }

  function logout () {
    const url = window.origin + '/api/logout'
    return fetch(url, { method: 'POST', credentials: 'include' })
      .then(async r => {
        const body = await r.json()
        if (body.status === 'OK') {
          pubkey.value = null
        }
      })
  }

  return { pubkey, isAuthenticated, initialized, checkSession, login, logout }
})
