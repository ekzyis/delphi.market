<template>
  <div>
    <button type="button" :class="yesClass" class="label success font-mono mx-1 my-3"
      @click.prevent="toggleYes">YES</button>
    <button type="button" :class="noClass" class="label error font-mono mx-1 my-3" @click.prevent="toggleNo">NO</button>
    <form v-show="showForm" @submit.prevent="submitForm">
      <label for="inv">you have:</label>
      <label name="inv">{{ userShares }} shares</label>
      <label for="shares">how many?</label>
      <input name="shares" v-model="shares" type="number" min="1" :max="userShares" placeholder="shares" required />
      <label for="price">price?</label>
      <input name="price" v-model="price" type="number" min="1" max="99" step="1" required />
      <label>you sell:</label>
      <label>{{ shares }} {{ selected }} shares @ {{ price }} sats</label>
      <label>you make:</label>
      <label>{{ format(profit) }} sats</label>
      <button class="col-span-2" type="submit" :disabled="disabled">submit sell order</button>
    </form>
    <div v-if="err" class="red text-center">{{ err }}</div>
  </div>
</template>

<script setup>
import { useSession } from '@/stores/session'
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const session = useSession()
const router = useRouter()
const route = useRoute()
const marketId = route.params.id
const selected = ref(route.query.share || null)
const showForm = computed(() => selected.value !== null)
const err = ref(null)

// how many shares wants the user sell?
const shares = ref(route.query.shares || 0)
// at which price?
const price = ref(route.query.price || 50)
// how high is the potential reward?
const profit = computed(() => {
  const val = shares.value * price.value
  return isNaN(val) ? 0 : val
})

const format = (x, i = 3) => x === null ? null : x >= 1 ? Math.round(x) : x === 0 ? x : x.toFixed(i)

const market = ref(null)

const url = '/api/market/' + marketId
await fetch(url)
  .then(r => r.json())
  .then(body => {
    market.value = body
  })
  .catch(console.error)
// Currently, we only support binary markets.
// (only events which can be answered with YES and NO)
const yesShareId = computed(() => market?.value.Shares.find(s => s.Description === 'YES').Id)
const noShareId = computed(() => market?.value.Shares.find(s => s.Description === 'NO').Id)
const shareId = computed(() => selected.value === 'YES' ? yesShareId.value : noShareId.value)
const sold = ref(0)
const userShares = computed(() => (((selected.value === 'YES' ? market.value.user?.YES : market.value.user?.NO) || 0) - sold.value))

const disabled = computed(() => userShares.value === 0)

const submitForm = async () => {
  if (!session.isAuthenticated) return router.push('/login')
  // TODO validate form
  const url = window.origin + '/api/order'
  const body = JSON.stringify({
    sid: shareId.value,
    quantity: shares.value,
    price: price.value,
    // TODO support selling
    side: 'SELL'
  })
  const res = await fetch(url, { method: 'POST', headers: { 'Content-type': 'application/json' }, body })
  const resBody = await res.json()
  if (res.status === 201) {
    sold.value += shares.value
    return
  }
  if (res.status !== 402) {
    err.value = `error: server responded with HTTP ${resBody.status}`
    return
  }
  const invoiceId = resBody.id
  router.push('/invoice/' + invoiceId)
}

const toggleYes = () => {
  selected.value = selected.value === 'YES' ? null : 'YES'
}

const toggleNo = () => {
  selected.value = selected.value === 'NO' ? null : 'NO'
}

const yesClass = computed(() => selected.value === 'YES' ? ['active'] : [])
const noClass = computed(() => selected.value === 'NO' ? ['active'] : [])

</script>

<style scoped>
.success.active {
  background-color: #35df8d;
  color: white;
}

.error.active {
  background-color: #ff7386;
  color: white;
}

form {
  margin: 0 auto;
  display: grid;
  grid-template-columns: auto auto;
}

form>* {
  margin: 0.5em 1em;
}
</style>
