<template>
  <div>
    <div v-if="session.pubkey" class="grid flex-row items-center">
      <div>authenticated as {{ session.pubkey.slice(0, 8) }}</div>
      <button class="ms-2 my-3" @click="logout">logout</button>
      <div>you have {{ session.msats / 1000 }} sats</div>
      <button class="ms-2 my-3" @click="toggleWithdrawalForm" :disabled="session.msats === 0">
        <span v-if="showWithdrawalForm">cancel</span>
        <span v-else>withdraw</span>
      </button>
    </div>
    <form v-show="showWithdrawalForm" @submit.prevent="submitWithdrawal">
      <label for="bolt11">bolt11</label>
      <input name="bolt11" id="bolt11" type="text" required v-model="bolt11" />
      <button type="submit" class="col-span-2">submit withdrawal</button>
    </form>
    <div v-if="err" class="red text-center">{{ err }}</div>
    <div v-if="success" class="green text-center">{{ success }}</div>
  </div>
</template>

<script setup>
import { useSession } from '@/stores/session'
import { ref } from 'vue'
import { useRouter } from 'vue-router'
const session = useSession()
const router = useRouter()

const logout = async () => {
  await session.logout()
  router.push('/')
}

const showWithdrawalForm = ref(false)
const bolt11 = ref(null)
const toggleWithdrawalForm = () => {
  showWithdrawalForm.value = !showWithdrawalForm.value
}
const err = ref(null)
const success = ref(null)
const submitWithdrawal = async () => {
  success.value = null
  err.value = null
  const url = '/api/withdrawal'
  const body = JSON.stringify({ bolt11: bolt11.value })
  try {
    const res = await fetch(url, { method: 'POST', headers: { 'Content-type': 'application/json' }, body })
    if (res.status === 200) {
      success.value = 'invoice paid'
      return session.checkSession()
    }
    const resBody = await res.json()
    err.value = resBody.reason || `error: server responded with HTTP ${res.status}`
  } catch (err) {
    console.error(err)
  }
}

</script>

<style scoped>
.grid {
  grid-template-columns: auto auto;
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
