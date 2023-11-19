<template>
  <form ref="form" class="flex flex-col mx-auto text-left" method="post" action="/api/market"
    @submit.prevent="submitForm">
    <label for="desc">event description</label>
    <textarea v-model="description" class="mb-1" id="desc" name="desc" type="text"></textarea>
    <label for="endDate">end date</label>
    <input v-model="endDate" class="mb-3" id="endDate" name="endDate" type="date" />
    <div class="flex flex-row justify-center">
      <button type="button" class="me-1" @click.prevent="$props.onCancel">cancel</button>
      <button type="submit">submit</button>
    </div>
  </form>
</template>

<script setup>
import { ref, defineProps } from 'vue'

defineProps(['onCancel'])

const form = ref(null)
const description = ref(null)
const endDate = ref(null)

const submitForm = async () => {
  const url = window.origin + '/api/market'
  const body = JSON.stringify({ description: description.value, endDate: endDate.value })
  await fetch(url, { method: 'post', headers: { 'Content-type': 'application/json' }, body })
}

</script>

<style scoped>
textarea {
  color: #000;
}
</style>
