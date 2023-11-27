<template>
  <div class="text w-auto">
    <table>
      <thead>
        <th>market</th>
        <th>description</th>
        <th class="hidden-sm">created at</th>
        <th>status</th>
      </thead>
      <tbody>
        <OrderRow :order="o" v-for="o in orders" :key="o.id" />
      </tbody>
    </table>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import OrderRow from './OrderRow.vue'

const orders = ref(null)

const url = '/api/orders'
await fetch(url)
  .then(r => r.json())
  .then(body => {
    orders.value = body
  })
  .catch(console.error)
</script>

<style scoped>
table {
  width: fit-content;
  align-items: center;
}

th {
  padding: 0 2rem;
}

@media only screen and (max-width: 600px) {
  th {
    padding: 0 1rem;
  }

  .hidden-sm {
    display: none
  }
}
</style>
