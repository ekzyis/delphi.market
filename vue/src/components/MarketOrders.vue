<template>
  <div class="text w-auto">
    <table>
      <thead>
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
import { useRoute } from 'vue-router'
import OrderRow from './OrderRow.vue'

const route = useRoute()
const marketId = route.params.id

const orders = ref([])
const url = `/api/market/${marketId}/orders`
await fetch(url)
  .then(r => r.json())
  .then(body => {
    orders.value = body.map(o => {
      // remove market column
      delete o.MarketId
      return o
    })
  })
  .catch(console.error)
</script>

<style>
table {
  width: 100%;
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
