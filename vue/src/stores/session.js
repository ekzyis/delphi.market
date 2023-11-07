import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

export const useSession = defineStore('session', () => {
  const pubkey = ref(null)
  const isAuthenticated = computed(() => !!pubkey.value)

  async function init () {
    try {
      const { pubkey } = await checkSession()
      if (pubkey) {
        console.log('authenticated as', pubkey)
        return
      }
      console.log('unauthenticated')
    } catch (err) {
      console.error('error:', err.reason || err)
    }
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

  return { pubkey, isAuthenticated, init, checkSession, login }
})
