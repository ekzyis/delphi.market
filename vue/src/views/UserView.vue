<template>
  <!-- eslint-disable -->
  <div class="my-3">
    <pre>

 _   _ ___  ___ _ __ 
| | | / __|/ _ \ '__|
| |_| \__ \  __/ |   
 \__,_|___/\___|_|   </pre>
  </div>
  <!-- eslint-enable -->
  <div v-if="session.pubkey">
    <div>authenticated as {{  session.pubkey.slice(0,8) }}</div>
    <button class="my-3" @click="logout">logout</button>
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
