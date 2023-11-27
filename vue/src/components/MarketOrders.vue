<template>
  <div class="text w-auto">
    <table>
      <thead>
        <th>description</th>
        <th class="hidden-sm">created at</th>
        <th>status</th>
      </thead>
      <tbody>
        <tr v-for="o in orders " :key="o.id" class="success">
          <td>{{ o.side }} {{ o.quantity }} {{ o.ShareDescription }} @ {{ o.price }} sats</td>
          <td class="hidden-sm">{{ ago(new Date(o.CreatedAt)) }}</td>
          <td class="font-mono">PENDING</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRoute } from 'vue-router'
import ago from 's-ago'

const route = useRoute()
const marketId = route.params.id

const orders = ref([])
const url = `/api/market/${marketId}/orders`
await fetch(url)
  .then(r => r.json())
  .then(body => {
    orders.value = body
  })
  .catch(console.error)
</script>

<style scoped>
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
