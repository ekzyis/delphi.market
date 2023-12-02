<template>
  <div class="text w-auto">
    <table>
      <thead>
        <th>sats</th>
        <th>created at</th>
        <th class="hidden-sm">expires at</th>
        <th>status</th>
        <th>details</th>
      </thead>
      <tbody>
        <tr v-for="i in invoices" :key="i.id">
          <td>{{ i.Msats / 1000 }}</td>
          <td :title="i.CreatedAt">{{ ago(new Date(i.CreatedAt)) }}</td>
          <td :title="i.ExpiresAt" class="hidden-sm">{{ ago(new Date(i.ExpiresAt)) }}</td>
          <td :class="'font-mono ' + classFromStatus(i.Status)">{{ i.Status }}</td>
          <td>
            <router-link :to="/invoice/ + i.Id">open</router-link>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import ago from 's-ago'

const classFromStatus = (status) => status === 'PAID' ? 'success' : status === 'PENDING' ? 'info' : 'error'

const invoices = ref(null)

const url = '/api/invoices'
await fetch(url)
  .then(r => r.json())
  .then(body => {
    invoices.value = body
  })
  .catch(console.error)
</script>

<style scoped>
table {
  width: fit-content;
  align-items: center;
}

td {
  padding: 0 0.5em;
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
