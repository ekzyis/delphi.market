<template>
  <div v-if="session.pubkey" class="flex flex-row items-center">
    <div>authenticated as {{ session.pubkey.slice(0, 8) }}</div>
    <button class="ms-2 my-3" @click="logout">logout</button>
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
