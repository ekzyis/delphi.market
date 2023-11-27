<template>
  <tr>
    <td v-if="order.MarketId"><router-link :to="/market/ + order.MarketId">{{ order.MarketId }}</router-link></td>
    <td :class="descClassName">{{ order.side }} {{ order.quantity }} {{ order.ShareDescription }} @ {{ order.price }} sats
    </td>
    <td :title="order.CreatedAt" class="hidden-sm">{{ ago(new Date(order.CreatedAt)) }}</td>
    <td :class="'font-mono ' + statusClassName">{{ order.Status }}</td>
  </tr>
</template>

<script setup>
import { ref, defineProps, computed } from 'vue'
import ago from 's-ago'

const props = defineProps(['order'])

const order = ref(props.order)

const descClassName = computed(() => {
  return order.value.side === 'BUY' ? 'success' : 'error'
})

const statusClassName = computed(() => {
  const status = order.value.Status
  if (status === 'PAID') return 'success'
  if (status === 'PENDING') return 'info'
  return 'error'
})

</script>
