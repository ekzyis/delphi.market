<template>
  <div v-if="session.pubkey">
    <div>authenticated as {{  session.pubkey.slice(0,8) }}</div>
    <button class="my-1" @click="logout">logout</button>
  </div>
</template>

<script setup>
import { useSession } from '@/stores/session'
import { useRouter } from 'vue-router'
const session = useSession()
const router = useRouter()

const logout = async () => {
  await session.logout()
  router.push('/')
}

</script>

<style scoped>
button {
  color: #ffffff;
  border: solid 1px #8787A4;
  padding: 0 1em;
}

button:hover {
  color: #ffffff;
  background: #8787A4;
}
</style>
