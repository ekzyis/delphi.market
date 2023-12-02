<template>
  <div>
    <div>
      <button type="button" :class="yesClass" class="label success font-mono mx-1 my-3"
        @click.prevent="toggleYes">YES</button>
      <button type="button" :class="noClass" class="label error font-mono mx-1 my-3" @click.prevent="toggleNo">NO</button>
    </div>
    <div v-if="userShares > 0 && showForm" class="flex justify-center items-center">
      <label for="side" class="m-1"><i>sell?</i></label>
      <input v-model="side" true-value="SELL" false-value="BUY" type="checkbox" name="side" class="m-1"/>
    </div>
    <form v-if="side === 'BUY'" v-show="showForm" @submit.prevent="submitBuyForm">
      <label for="stake">how much?</label>
      <input name="stake" v-model="stake" type="number" min="0" placeholder="sats" required />
      <label for="certainty">how sure?</label>
      <input name="certainty" v-model="certainty" type="number" min="0.01" max="0.99" step="0.01" required />
      <label>you receive:</label>
      <label>{{ format(shares) }} {{ selected }} shares @ {{ format(price) }} sats</label>
      <label>you pay:</label>
      <label>{{ format(cost) }} sats</label>
      <label>if you win:</label>
      <label>+{{ format(profit) }} sats</label>
      <button class="col-span-2" type="submit">submit buy order</button>
    </form>
    <form v-else v-show="showForm" @submit.prevent="submitSellForm">
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

// YES NO button logic
// -- which button was pressed?
const selected = ref(route.query.share || null)
// -- button css
const yesClass = computed(() => selected.value === 'YES' ? ['active'] : [])
const noClass = computed(() => selected.value === 'NO' ? ['active'] : [])
// -- show form if any button was pressed
const showForm = computed(() => selected.value !== null)
const toggleYes = () => {
  selected.value = selected.value === 'YES' ? null : 'YES'
}
const toggleNo = () => {
  selected.value = selected.value === 'NO' ? null : 'NO'
}
// show errors below form
const err = ref(null)
// BUY or SELL?
const side = ref(route.query.side || 'BUY')

// -- BUY params
// how much wants the user bet?
const stake = ref(route.query.stake || 100)
// how sure is the user he will win?
const certainty = ref(route.query.certainty || 0.5)
// price per share: more risk, lower price, higher reward
const price = computed(() => certainty.value * 100)
// how many (full) shares can be bought?
const shares = computed(() => {
  const val = price.value > 0 ? stake.value / price.value : null
  // only full shares can be bought
  return Math.round(val)
})
// how much does this order cost?
const cost = computed(() => {
  return shares.value * price.value
})
// how high is the potential reward?
const profit = computed(() => {
  // shares expire at 10 or 0 sats
  const val = (100 * shares.value) - cost.value
  return isNaN(val) ? 0 : val
})

// -- SELL params
// how many shares does the user own?
const userShares = computed(() => (((selected.value === 'YES' ? market.value.user?.YES : market.value.user?.NO) || 0) - sold.value))
// how many share did the user sell since we refreshed our data?
const sold = ref(0)

const disabled = computed(() => userShares.value === 0)

const format = (x, i = 3) => x === null ? null : x >= 1 ? Math.round(x) : x === 0 ? x : x.toFixed(i)

// Fetch market data
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
const yesShareId = computed(() => {
  return market?.value.Shares.find(s => s.Description === 'YES').Id
})
const noShareId = computed(() => {
  return market?.value.Shares.find(s => s.Description === 'NO').Id
})
const shareId = computed(() => {
  return selected.value === 'YES' ? yesShareId.value : noShareId.value
})

const submitBuyForm = async () => {
  if (!session.isAuthenticated) return router.push('/login')
  // TODO validate form
  const url = window.origin + '/api/order'
  const body = JSON.stringify({
    sid: shareId.value,
    quantity: shares.value,
    price: price.value,
    side: 'BUY'
  })
  const res = await fetch(url, { method: 'POST', headers: { 'Content-type': 'application/json' }, body })
  const resBody = await res.json()
  if (res.status !== 402) {
    err.value = `error: server responded with HTTP ${resBody.status}`
    return
  }
  const invoiceId = resBody.id
  router.push('/invoice/' + invoiceId)
}

const submitSellForm = async () => {
  if (!session.isAuthenticated) return router.push('/login')
  // TODO validate form
  const url = window.origin + '/api/order'
  const body = JSON.stringify({
    sid: shareId.value,
    quantity: shares.value,
    price: price.value,
    side: 'SELL'
  })
  const res = await fetch(url, { method: 'POST', headers: { 'Content-type': 'application/json' }, body })
  const resBody = await res.json()
  if (res.status !== 402) {
    err.value = `error: server responded with HTTP ${resBody.status}`
    return
  }
  const invoiceId = resBody.id
  router.push('/invoice/' + invoiceId)
}

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
