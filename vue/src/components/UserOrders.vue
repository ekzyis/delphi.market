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
        <tr v-for="o in orders " :key="o.id">
          <td><router-link :to="/market/ + o.MarketId">{{ o.MarketId }}</router-link></td>
          <td>{{ o.side }} {{ o.quantity }} {{ o.ShareDescription }} @ {{ o.price }} sats</td>
          <td class="hidden-sm">{{ ago(new Date(o.CreatedAt)) }}</td>
          <td></td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import ago from 's-ago'

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
