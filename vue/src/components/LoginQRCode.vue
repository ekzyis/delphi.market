<template>
  <a v-if="lnurl" :href="'lightning:' + lnurl">
    <img v-if="qr" :src="'data:image/png;base64,' + qr" />
  </a>
</template>

<script setup>
import { ref } from 'vue'
import { useSession } from '@/stores/session'

const qr = ref(null)
const lnurl = ref(null)

const login = async () => {
  const s = await session.login()
  qr.value = s.qr
  lnurl.value = s.lnurl
}

const session = useSession()
// wait until session is initialized
if (session.initialized && !session.isAuthenticated) {
  await login()
} else {
  // else subscribe to changes
  session.$subscribe(async (_, s) => {
    if (s.initialized && !s.isAuthenticated) {
      await login()
    }
  })
}

</script>
