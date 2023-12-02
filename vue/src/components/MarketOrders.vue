<template>
  <div class="text w-auto mt-3">
    <table>
      <thead>
        <th>description</th>
        <th class="hidden-sm">created at</th>
        <th>status</th>
        <th></th>
      </thead>
      <tbody>
        <OrderRow :order="o" v-for="o in orders" :key="o.Id" @mouseover="() => mouseover(o.Id)"
          :selected="selected" :click="click" />
      </tbody>
    </table>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import OrderRow from './OrderRow.vue'

const router = useRouter()
const route = useRoute()
const marketId = route.params.id

const selected = ref([])

function mouseover (oid) {
  const o2id = orders.value.find(i => i.OrderId === oid)?.Id
  if (o2id) {
    selected.value = [oid, o2id]
  } else {
    // reset selection
    selected.value = []
  }
}

const click = (order) => {
  // redirect to form with prefilled inputs to match order
  if (order.side === 'BUY') {
    // match BUY YES with BUY NO and vice versa
    const stake = order.quantity * (100 - order.price)
    const certainty = (100 - order.price) / 100
    const share = order.ShareDescription === 'YES' ? 'NO' : 'YES'
    router.push(`/market/${marketId}/form?stake=${stake}&certainty=${certainty}&share=${share}&side=BUY`)
  }
  if (order.side === 'SELL') {
    // SELL YES -> BUY YES, SELL NO -> BUY NO
    const stake = order.quantity * order.price
    const certainty = order.price / 100
    const share = order.ShareDescription === 'YES' ? 'YES' : 'NO'
    router.push(`/market/${marketId}/form?stake=${stake}&certainty=${certainty}&share=${share}&side=BUY`)
  }
}

const orders = ref([])
const url = `/api/market/${marketId}/orders`
await fetch(url)
  .then(r => r.json())
  .then(body => {
    orders.value = body?.map(o => {
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
