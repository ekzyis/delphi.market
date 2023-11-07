<template>
  <a v-if="lnurl" :href="'lightning:' + lnurl">
    <img v-if="qr" :src="'data:image/png;base64,' + qr" />
  </a>
</template>

<script setup>
import { ref } from 'vue'
import { useSession } from '@/stores/session'

let qr = ref(null)
let lnurl = ref(null)

const session = useSession()
await (async () => {
  try {
    if (session.isAuthenticated) return
    const s = await session.login()
    qr = s.qr
    lnurl = s.lnurl
  } catch (err) {
    console.error('error:', err.reason || err)
  }
})()
</script>
