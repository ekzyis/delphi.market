<template>
  <tr @mouseover="mouseover" @mouseleave="mouseleave">
    <td v-if="order.MarketId"><router-link :to="/market/ + order.MarketId">{{ order.MarketId }}</router-link></td>
    <td>{{ order.side }} {{ order.quantity }} {{ order.ShareDescription }} @ {{ order.price }} sats
    </td>
    <td :title="order.CreatedAt" class="hidden-sm">{{ ago(new Date(order.CreatedAt)) }}</td>
    <td :class="'font-mono ' + statusClassName + ' ' + selectedClassName">{{ order.Status }}</td>
    <td v-if="showContextMenu && !!session.pubkey">
      <button @click="() => click(order)" :disabled="mine">match</button>
    </td>
  </tr>
</template>

<script setup>
import { ref, defineProps, computed } from 'vue'
import ago from 's-ago'
import { useSession } from '@/stores/session'

const session = useSession()
const props = defineProps(['order', 'selected', 'click'])

const order = ref(props.order)
const showContextMenu = ref(false)
const click = ref(props.click)

const mine = order.value.Pubkey === session?.pubkey

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
  if (!!props.click && order.value.Status === 'PENDING') {
    showContextMenu.value = true
  }
}

const mouseleave = () => {
  showContextMenu.value = false
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
