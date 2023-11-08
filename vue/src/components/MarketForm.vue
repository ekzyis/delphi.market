<template>
  <form ref="form" class="mx-auto text-left" method="post" action="/api/market" @submit.prevent="submitForm">
    <label class="mx-1" for="desc">description</label>
    <input v-model="description" class="mx-1" id="desc" name="desc" type="text" placeholder="event description">
    <button class="mx-1" type="submit">submit</button>
  </form>
</template>

<script setup>
import { ref } from 'vue'

const form = ref(null)
const description = ref(null)

const submitForm = async () => {
  const url = window.origin + '/api/market'
  const body = JSON.stringify({ description: description.value })
  await fetch(url, { method: 'post', body })
}

</script>

<style scoped>
input {
  padding: 0 0.3em;
  color: #000;
}
</style>
