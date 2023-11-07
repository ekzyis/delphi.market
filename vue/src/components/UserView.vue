<template>
  <div v-if="session.pubkey">authenticated as {{  session.pubkey.slice(0,8) }}</div>
</template>

<script setup>
import { useSession } from '@/stores/session'
const session = useSession()
</script>
