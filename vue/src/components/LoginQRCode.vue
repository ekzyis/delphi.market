<template>
  <a v-if="lnurl" :href="'lightning:' + lnurl">
    <img v-if="qr" :src="'data:image/png;base64,' + qr" />
  </a>
</template>

<script setup>
import { ref } from 'vue'
import { useSession } from '@/stores/session'
import { useRouter } from 'vue-router'

const qr = ref(null)
const lnurl = ref(null)

const login = async () => {
  const s = await session.login()
  qr.value = s.qr
  lnurl.value = s.lnurl
}

const router = useRouter()
const session = useSession()

await (async () => {
  // redirect to / if session already exists
  if (session.initialized) {
    if (session.isAuthenticated) return router.push('/')
    return login()
  }
  // else subscribe to changes
  return session.$subscribe(() => {
    if (session.initialized) {
      // for some reason, computed property is only updated when accessing the store directly
      // it is not updated inside the second argument
      if (session.isAuthenticated) return router.push('/')
      return login()
    }
  })
})()

</script>
