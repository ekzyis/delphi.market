<template>
  <tr @mouseleave="mouseleave">
    <td v-if="order.MarketId"><router-link :to="/market/ + order.MarketId">{{ order.MarketId }}</router-link></td>
    <td>{{ order.side }} {{ order.quantity }} {{ order.ShareDescription }} @ {{ order.price }} sats
    </td>
    <td :title="order.CreatedAt" class="hidden-sm">{{ ago(new Date(order.CreatedAt)) }}</td>
    <td :class="'font-mono ' + statusClassName + ' ' + selectedClassName" @mouseover="mouseover">{{ order.Status }}</td>
    <td v-if="showContextMenu">
      <button @click="() => onMatchClick?.(order)" v-if="showMatch">match</button>
      <button @click="() => cancelOrder(order)" v-if="showCancel">cancel</button>
    </td>
  </tr>
</template>

<script setup>
import { ref, defineProps, computed } from 'vue'
import ago from 's-ago'
import { useSession } from '@/stores/session'

const session = useSession()
const props = defineProps(['order', 'selected', 'onMatchClick'])

const order = ref(props.order)
const showContextMenu = ref(false)
const onMatchClick = ref(props.onMatchClick)
const mine = computed(() => order.value.Pubkey === session?.pubkey)
const showMatch = computed(() => !mine.value && order.value.Status === 'PENDING')
const showCancel = computed(() => mine.value && order.value.Status === 'PENDING')

const statusClassName = computed(() => {
  const status = order.value.Status
  if (status === 'EXECUTED') return 'success'
  if (status === 'PENDING') return 'info'
  return 'error'
})

const selectedClassName = computed(() => {
  if (typeof props.selected === 'boolean') {
    return props.selected ? 'selected' : ''
  }
  if (Array.isArray(props.selected)) {
    return props.selected.some(id => id === order.value.Id) ? 'selected' : ''
  }
  return ''
})

const mouseover = () => {
  showContextMenu.value = true && !!session.pubkey
}

const mouseleave = () => {
  showContextMenu.value = false
}

const cancelOrder = async () => {
  const url = '/api/order/' + order.value.Id
  await fetch(url, { method: 'DELETE' }).catch(console.error)
}

</script>

<style scoped>
td {
  padding: 0 0.5em;
}

.selected {
  background-color: #35df8d;
  color: white;
}
</style>
