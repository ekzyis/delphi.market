<template>
  <ul>
    <li class="my-3" v-for="market in markets" :key="market.id">
      <router-link :to="'/market/' + market.id + '/form'">{{ market.description }}</router-link>
    </li>
  </ul>
  <button v-if="!showForm" @click.prevent="toggleForm">+ create market</button>
  <div v-else class="flex flex-col justify-center">
    <MarketForm :onCancel="toggleForm" />
  </div>
</template>

<script setup>
import MarketForm from './MarketForm'
import { ref } from 'vue'
import { useSession } from '@/stores/session'
import { useRouter } from 'vue-router'

const session = useSession()
const router = useRouter()
const markets = ref([])
const showForm = ref(false)

// TODO only load markets once per session
const url = window.origin + '/api/markets'
await fetch(url).then(async r => {
  const body = await r.json()
  markets.value = body
})

const toggleForm = () => {
  if (!session.isAuthenticated) {
    return router.push('/login')
  }
  showForm.value = !showForm.value
}

</script>

<style scoped>
a {
  padding: 0 1em;
}
</style>
