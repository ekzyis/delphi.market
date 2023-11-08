<template>
  <ul>
    <li class="my-3" v-for="market in markets" :key="market.Id">
      <router-link :to="'/market/' + market.Id">{{ market.Description }}</router-link>
    </li>
  </ul>
  <button v-if="!showForm" @click.prevent="toggleForm">+ create market</button>
  <div class="flex flex-col justify-center" v-else>
    <div class="my-1 text-center">+ create market</div>
    <form ref="form" class="mx-auto text-left" method="post" action="/api/market" @submit.prevent="submitForm">
      <label class="mx-1" for="desc">description</label>
      <input v-model="description" class="mx-1" id="desc" name="desc" type="text">
      <button class="mx-1" type="submit">submit</button>
    </form>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const markets = ref([])
const form = ref(null)
const description = ref(null)
const showForm = ref(false)

// TODO only load markets once per session
const url = window.origin + '/api/markets'
await fetch(url).then(async r => {
  const body = await r.json()
  markets.value = body
})

const toggleForm = () => {
  description.value = null
  showForm.value = !showForm.value
}

const submitForm = async () => {
  const url = window.origin + '/api/market'
  const body = JSON.stringify({ description: description.value })
  await fetch(url, { method: 'post', body })
  toggleForm()
}

</script>

<style scoped>
input {
  padding: 0 0.3em;
  color: #000;
}

a {
  padding: 0 1em;
}
</style>
