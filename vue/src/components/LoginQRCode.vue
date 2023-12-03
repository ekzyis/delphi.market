<template>
  <div class="flex flex-col">
    <router-link v-if="success" to="/" class="label success font-mono">
      <div>Authenticated</div>
      <small>Redirecting in {{ redirectTimeout }} ...</small>
    </router-link>
    <div class="font-mono my-3">
      LNURL-auth
    </div>
    <div v-if="error" class="label error font-mono">
      <div>Authentication error</div>
      <small>{{ error }}</small>
    </div>
    <figure v-if="lnurl && qr" class="flex flex-col m-auto">
      <a class="m-auto" :href="'lightning:' + lnurl">
        <img :src="'data:image/png;base64,' + qr" />
      </a>
      <figcaption class="flex flex-row my-3 font-mono text-xs">
        <span class="w-[80%] text-ellipsis overflow-hidden">{{ lnurl }}</span>
        <button @click.prevent="copy">{{ label }}</button>
      </figcaption>
    </figure>
  </div>
</template>

<script setup>
import { onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useSession } from '@/stores/session'

const router = useRouter()
const session = useSession()

const qr = ref(null)
const lnurl = ref(null)
let interval = null
const LOGIN_POLL = 2000
const redirectTimeout = ref(3)
const success = ref(null)
const error = ref(null)
const label = ref('copy')
let copyTimeout = null

const copy = () => {
  navigator.clipboard?.writeText(lnurl.value)
  label.value = 'copied'
  if (copyTimeout) clearTimeout(copyTimeout)
  copyTimeout = setTimeout(() => { label.value = 'copy' }, 1500)
}

const poll = async () => {
  try {
    await session.checkSession()
    if (session.isAuthenticated) {
      success.value = true
      clearInterval(interval)
      interval = setInterval(() => {
        if (--redirectTimeout.value === 0) {
          router.push('/')
        }
      }, 1000)
    }
  } catch (err) {
    // ignore 404 errors
    if (err.reason !== 'session not found') {
      console.error(err)
      error.value = err.reason
    }
  }
}

const login = async () => {
  const s = await session.login()
  qr.value = s.qr
  lnurl.value = s.lnurl
  interval = setInterval(poll, LOGIN_POLL)
}

await (async () => {
  // redirect to / if session already exists
  if (session.initialized) {
    if (session.isAuthenticated) return router.push('/')
    return login()
  }
  // else subscribe to changes
  return session.$subscribe(() => {
    if (session.initialized) {
      // for some reason, computed property is only updated when accessing the store directly
      // it is not updated inside the second argument
      if (session.isAuthenticated) return router.push('/')
      return login()
    }
  })
})()

onUnmounted(() => { clearInterval(interval) })

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
</style>
