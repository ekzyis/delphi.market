<template>
  <div>
    <div>
      <button type="button" :class="yesClass" class="label success font-mono mx-1 my-3"
        @click.prevent="toggleYes">YES</button>
      <button type="button" :class="noClass" class="label error font-mono mx-1 my-3" @click.prevent="toggleNo">NO</button>
    </div>
    <form v-if="side === 'BUY'" v-show="showForm" @submit.prevent="submitBuyForm">
      <label v-if="session.isAuthenticated">you have:</label>
      <label v-if="session.isAuthenticated">{{ session.msats / 1000 }} sats and {{ userShares }} shares</label>
      <label v-if="session.isAuthenticated">sell?</label>
      <input v-if="session.isAuthenticated" v-model="side" true-value="SELL" false-value="BUY" type="checkbox" class="m-1"
        :disabled="userShares === 0" />
      <label for="stake">how much?</label>
      <input id="stake" v-model="stake" type="number" min="0" placeholder="sats" required />
      <label for="certainty">how sure?</label>
      <input id="certainty" v-model="certainty" type="number" min="0.01" max="0.99" step="0.01" required />
      <label>you receive:</label>
      <label>{{ format(buyShares) }} {{ selected }} shares @ {{ format(buyPrice) }} sats</label>
      <label>you pay:</label>
      <label>{{ format(buyCost) }} sats</label>
      <label>if you win:</label>
      <label>+{{ format(buyProfit) }} sats</label>
      <button class="col-span-2" type="submit" :disabled="!!market.SettledAt">submit buy order</button>
    </form>
    <form v-else v-show="showForm" @submit.prevent="submitSellForm">
      <label v-if="session.isAuthenticated">you have:</label>
      <label v-if="session.isAuthenticated">{{ session.msats / 1000 }} sats and {{ userShares }} shares</label>
      <label v-if="session.isAuthenticated">sell?</label>
      <input v-if="session.isAuthenticated" v-model="side" true-value="SELL" false-value="BUY" type="checkbox" class="m-1"
        :disabled="userShares === 0" />
      <label for="shares">how many?</label>
      <input id="shares" v-model="sellShares" type="number" min="1" :max="userShares" placeholder="shares" required />
      <label for="price">price?</label>
      <input id="price" v-model="sellPrice" type="number" min="1" max="99" step="1" required />
      <label>you sell:</label>
      <label>{{ sellShares }} {{ selected }} shares @ {{ format(sellPrice) }} sats</label>
      <label>you make:</label>
      <label>+{{ format(sellProfit) }} sats</label>
      <button class="col-span-2" type="submit" :disabled="userShares === 0 || !!market.SettledAt">
        submit sell order
      </button>
    </form>
    <div v-if="err" class="red text-center">{{ err }}</div>
    <div v-if="success" class="green text-center">{{ success }}</div>
  </div>
</template>

<script setup>
import { useSession } from '@/stores/session'
import { ref, computed, defineProps } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const props = defineProps(['market'])
const market = ref(props.market)
const session = useSession()
const router = useRouter()
const route = useRoute()

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
// show error and success below form
const err = ref(null)
const success = ref(null)
// BUY or SELL?
const side = ref(route.query.side || 'BUY')

// -- BUY params
// how much wants the user bet?
const stake = ref(route.query.stake || 100)
// how sure is the user he will win?
const certainty = ref(route.query.certainty || 0.5)
const buyPrice = computed(() => Math.round(certainty.value * 100))
const buyShares = computed(() => {
  const val = buyPrice.value > 0 ? stake.value / buyPrice.value : null
  // only full shares can be bought
  return Math.round(val)
})
// how much does this order cost?
const buyCost = computed(() => {
  return buyShares.value * buyPrice.value
})
// how high is the potential reward?
const buyProfit = computed(() => {
  // shares expire at 10 or 0 sats
  const val = (100 * buyShares.value) - buyCost.value
  return isNaN(val) ? 0 : val
})
// -- SELL params
// how many shares does the user own?
const userShares = computed(() => (((selected.value === 'YES' ? market.value.user?.YES : market.value.user?.NO) || 0) - sold.value))
// how many shares does the user want to sell?
const sellShares = ref(2)
// at which price wants the user to sell each share?
const sellPrice = ref(50)
const sellProfit = computed(() => sellShares.value * sellPrice.value)
// how many share did the user sell since we refreshed our data?
const sold = ref(0)

const format = (x, i = 3) => x === null ? null : x >= 1 ? Math.round(x) : x === 0 ? x : x.toFixed(i)

// Currently, we only support binary markets.
// (only events which can be answered with YES and NO)
const yesShareId = computed(() => {
  return market?.value.Shares.find(s => s.Description === 'YES').sid
})
const noShareId = computed(() => {
  return market?.value.Shares.find(s => s.Description === 'NO').sid
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
    quantity: buyShares.value,
    price: buyPrice.value,
    side: 'BUY'
  })
  const res = await fetch(url, { method: 'POST', headers: { 'Content-type': 'application/json' }, body })
  const resBody = await res.json()
  if (res.status !== 402) {
    err.value = `error: server responded with HTTP ${res.status}`
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
    quantity: sellShares.value,
    price: sellPrice.value,
    side: 'SELL'
  })
  const res = await fetch(url, { method: 'POST', headers: { 'Content-type': 'application/json' }, body })
  const resBody = await res.json()
  if (res.status === 201) {
    success.value = 'Order created'
    return
  }
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

label {
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
