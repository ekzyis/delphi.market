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
      <small>{{ error  }}</small>
    </div>
    <figure class="flex flex-col m-auto">
      <a class="m-auto" v-if="lnurl" :href="'lightning:' + lnurl">
        <img v-if="qr" :src="'data:image/png;base64,' + qr" />
      </a>
      <figcaption class="my-3 font-mono text-xs text-ellipsis overflow-hidden">{{ lnurl }}</figcaption>
    </figure>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useSession } from '@/stores/session'
import { useRouter } from 'vue-router'

const router = useRouter()
const session = useSession()

const qr = ref(null)
const lnurl = ref(null)
let interval = null
const LOGIN_POLL = 2000
const redirectTimeout = ref(3)
const success = ref(null)
const error = ref(null)

const poll = async () => {
  try {
    await session.checkSession()
    if (session.isAuthenticated) {
      success.value = true
      clearInterval(interval)
      setInterval(() => {
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
  width: fit-content;
  margin: 1em auto;
  padding: 0.5em 3em;
  cursor: pointer;
}
.label:hover {
  color: white;
}
.success {
  background-color: rgba(20,158,97,.24);
  color: #35df8d;
}
.success:hover {
  background-color: #35df8d;
}
.error {
    background-color: rgba(245,57,94,.24);
    color: #ff7386;
}
.error:hover {
  background-color: #ff7386;
}
</style>
