<template>
  <div class="flex flex-col">
    <div class="font-mono mb-3">
      Payment Required
    </div>
    <router-link v-if="invoice?.ConfirmedAt" :to="callbackUrl" class="label success font-mono">
      <div>Paid</div>
      <small v-if="redirectTimeout < 4">Redirecting in {{ redirectTimeout }} ...</small>
    </router-link>
    <div v-if="invoice && !invoice.ConfirmedAt && new Date(invoice.ExpiresAt) < new Date()" class="label error font-mono">
      <div>Expired</div>
    </div>
    <div v-if="notFound" class="label error font-mono">
      <div>Not Found</div>
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
        <span v-if="faucet" class="mx-3 my-1">faucet</span>
        <span v-if="faucet" class="text-ellipsis overflow-hidden font-mono me-3 my-1">
          <a href="https://faucet.mutinynet.com/" target="_blank">faucet.mutinynet.com</a>
        </span>
        <span class="mx-3 my-1">payment hash</span><span class="text-ellipsis overflow-hidden font-mono me-3 my-1">
          {{ invoice.Hash }}
        </span>
        <span class="mx-3 my-1">created at</span><span class="text-ellipsis overflow-hidden font-mono me-3 my-1">
          {{ invoice.CreatedAt }} ({{ ago(new Date(invoice.CreatedAt)) }})
        </span>
        <span class="mx-3 my-1">expires at</span><span class="text-ellipsis overflow-hidden font-mono me-3 my-1">
          {{ invoice.ExpiresAt }} ({{ ago(new Date(invoice.ExpiresAt)) }})
        </span>
        <span class="mx-3 my-1">sats</span><span class="text-ellipsis overflow-hidden font-mono me-3 my-1">
          {{ invoice.Msats / 1000 }}
        </span>
        <span class="mx-3 my-1">description</span>
        <span class="text-ellipsis overflow-hidden font-mono me-3 my-1">
          <span v-if="invoice.DescriptionMarketId">
            <span v-if="invoice.Description">
              <span>{{ invoice.Description }}</span>
              <router-link :to="'/market/' + invoice.DescriptionMarketId + '/orders'">[market]</router-link>
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
import { onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import ago from 's-ago'

const router = useRouter()
const route = useRoute()

const invoice = ref(undefined)
const redirectTimeout = ref(4)
const notFound = ref(undefined)
const label = ref('copy')
let copyTimeout = null

const copy = () => {
  navigator.clipboard?.writeText(invoice.value.PaymentRequest)
  label.value = 'copied'
  if (copyTimeout) clearTimeout(copyTimeout)
  copyTimeout = setTimeout(() => { label.value = 'copy' }, 1500)
}

const callbackUrl = ref('/')
let pollCount = 0
let pollTimeout
let redirectInterval
const INVOICE_POLL = 2000
const fetchInvoice = async () => {
  const url = window.origin + '/api/invoice/' + route.params.id
  const res = await fetch(url)
  notFound.value = res.status === 404
  if (res.status === 404) {
    return
  }
  const body = await res.json()
  if (body.Description) {
    // parse invoice description to show links
    const regexp = /\[market:(?<id>[0-9]+)\]/
    const m = body.Description.match(regexp)
    const marketId = m?.groups?.id
    if (marketId) {
      body.DescriptionMarketId = marketId
      body.Description = body.Description.replace(regexp, '')
      callbackUrl.value = '/market/' + marketId + '/orders'
    }
  }
  invoice.value = body
  if (new Date(invoice.value.ExpiresAt) < new Date()) {
    // invoice expired
    return
  }
  if (!invoice.value.ConfirmedAt) {
    // invoice not pad (yet?)
    pollTimeout = setTimeout(() => {
      pollCount++
      fetchInvoice()
    }, INVOICE_POLL)
  } else {
    // invoice paid
    // we check for pollCount > 0 to only redirect if invoice wasn't already paid when we visited the page
    if (pollCount > 0) {
      redirectInterval = setInterval(() => {
        redirectTimeout.value--
        if (redirectTimeout.value === 0) {
          clearInterval(redirectInterval)
          return router.push(callbackUrl.value)
        }
      }, 1000)
    }
  }
}
await fetchInvoice()

onUnmounted(() => { clearTimeout(pollTimeout); clearInterval(redirectInterval) })

const faucet = window.location.hostname === 'delphi.market' ? 'https://faucet.mutinynet.com' : ''

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
  margin: auto;
  margin-bottom: 0.75em
}

a.label {
  text-decoration: none;
}

div.grid {
  grid-template-columns: auto auto;
}
</style>
