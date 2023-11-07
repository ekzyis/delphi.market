import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

export const useSession = defineStore('session', () => {
  const pubkey = ref(null)
  const isAuthenticated = computed(() => !!pubkey.value)
  const initialized = ref(false)

  async function init () {
    try {
      const { pubkey } = await checkSession()
      if (pubkey) {
        console.log('authenticated as', pubkey)
      } else console.log('unauthenticated')
    } catch (err) {
      console.error('error:', err.reason || err)
    }
    initialized.value = true
  }

  function checkSession () {
    const url = window.origin + '/api/session'
    return fetch(url, {
      credentials: 'include'
    })
      .then(async r => {
        const body = await r.json()
        pubkey.value = body.pubkey
        return body
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

  return { pubkey, isAuthenticated, initialized, init, checkSession, login, logout }
})
