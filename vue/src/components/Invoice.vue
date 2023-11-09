<template>
  <div class="flex flex-col">
    <router-link v-if="success" :to="callbackUrl" class="label success font-mono">
      <div>Paid</div>
      <small v-if="redirectTimeout > 0">Redirecting in {{ redirectTimeout }} ...</small>
    </router-link>
    <div class="font-mono my-3">
      Payment Required
    </div>
    <div v-if="error" class="label error font-mono">
      <div>Error</div>
      <small>{{ error }}</small>
    </div>
    <div v-if="invoice">
      <figure class="flex flex-col m-auto">
        <a class="m-auto" :href="'lightning:' + invoice.PaymentRequest">
          <img :src="'data:image/png;base64,' + invoice.Qr" />
        </a>
        <figcaption class="flex flex-row my-3 font-mono text-xs">
          <span class="w-[80%] text-ellipsis overflow-hidden">{{ invoice.PaymentRequest }}</span>
          <button @click.prevent="copy">{{ label }}</button>
        </figcaption>
      </figure>
      <div class="grid text-muted text-xs">
        <span class="mx-3 my-1">payment hash</span><span class="text-ellipsis overflow-hidden font-mono me-3 my-1">
          {{ invoice.Hash }}
        </span>
        <span class="mx-3 my-1">created at</span><span class="text-ellipsis overflow-hidden font-mono me-3 my-1">
          {{ invoice.CreatedAt }}
        </span>
        <span class="mx-3 my-1">expires at</span><span class="text-ellipsis overflow-hidden font-mono me-3 my-1">
          {{ invoice.ExpiresAt }}
        </span>
        <span class="mx-3 my-1">sats</span><span class="text-ellipsis overflow-hidden font-mono me-3 my-1">
          {{ invoice.Msats / 1000 }}
        </span>
        <span class="mx-3 my-1">description</span>
        <span class="text-ellipsis overflow-hidden font-mono me-3 my-1">
          <span v-if="invoice.DescriptionMarketId">
            <span v-if="invoice.Description">
              <span>{{ invoice.Description }}</span>
              <router-link :to="'/market/' + invoice.DescriptionMarketId">[market]</router-link>
            </span>
            <span v-else>&lt;empty&gt;</span>
          </span>
          <span v-else>
            <span v-if="invoice.Description">{{ invoice.Description }}</span>
            <span v-else>&lt;empty&gt;</span>
          </span>
        </span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const router = useRouter()
const route = useRoute()
// TODO validate callback url
const callbackUrl = ref('/')
let pollCount = 0
const INVOICE_POLL = 2000
const poll = async () => {
  pollCount++
  const url = window.origin + '/api/invoice/' + route.params.id
  const res = await fetch(url)
  const body = await res.json()
  if (body.ConfirmedAt) {
    success.value = true
    clearInterval(interval)
    if (pollCount > 1) {
      // only redirect if the invoice was not immediately paid
      setInterval(() => {
        if (--redirectTimeout.value === 0) {
          router.push(callbackUrl.value)
        }
      }, 1000)
    } else {
      redirectTimeout.value = -1
    }
  }
}

let interval
const invoice = ref(undefined)
const redirectTimeout = ref(3)
const success = ref(null)
const error = ref(null)
const label = ref('copy')
let copyTimeout = null

const copy = () => {
  navigator.clipboard?.writeText(invoice.value.PaymentRequest)
  label.value = 'copied'
  if (copyTimeout) clearTimeout(copyTimeout)
  copyTimeout = setTimeout(() => { label.value = 'copy' }, 1500)
}

await (async () => {
  const url = window.origin + '/api/invoice/' + route.params.id
  const res = await fetch(url)
  if (res.status === 404) {
    error.value = 'invoice not found'
    return
  }
  const body = await res.json()
  if (body.Description) {
    const regexp = /\[market:(?<id>[0-9]+)\]/
    const m = body.Description.match(regexp)
    const marketId = m.groups?.id
    if (marketId) {
      body.DescriptionMarketId = marketId
      body.Description = body.Description.replace(regexp, '')
      callbackUrl.value = '/market/' + marketId
    }
  }
  invoice.value = body
  interval = setInterval(poll, INVOICE_POLL)
})()
</script>

<style scoped>
img {
  width: 256px;
  height: auto;
}

figcaption {
  margin: 0.75em auto;
  width: 256px;
}

.label {
  margin: 1em auto;
}

a.label {
  text-decoration: none;
}

div.grid {
  grid-template-columns: auto auto;
}
</style>
