<template>
  <ul>
    <li class="my-3" v-for="market in markets" :key="market.Id">
      <router-link :to="'/market/' + market.Id">{{ market.Description }}</router-link>
    </li>
    <button>+ create market</button>
  </ul>
</template>

<script setup>
import { ref } from 'vue'

const markets = ref([])

const url = window.origin + '/api/markets'

// TODO only load markets once per session
await fetch(url).then(async r => {
  const body = await r.json()
  markets.value = body
})

</script>

<style scoped>
a {
  padding: 0 1em;
}
</style>
