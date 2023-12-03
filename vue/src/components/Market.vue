<template>
  <!-- eslint-disable -->
  <div class="my-3">
    <pre>
                      _        _   
 _ __ ___   __ _ _ __| | _____| |_ 
| '_ ` _ \ / _` | '__| |/ / _ \ __|
| | | | | | (_| | |  |   ( __/ |_  
|_| |_| |_|\__,_|_|  |_|\_\___|\__|</pre>
  </div>
  <div class="font-mono">{{ market.Description }}</div>
  <!-- eslint-enable -->
  <header class="flex flex-row text-center justify-center pt-1">
    <nav>
      <StyledLink :to="'/market/' + marketId + '/form'">form</StyledLink>
      <StyledLink :to="'/market/' + marketId + '/orders'">orders</StyledLink>
      <StyledLink :to="'/market/' + marketId + '/stats'">stats</StyledLink>
      <StyledLink v-if="mine" :to="'/market/' + marketId + '/settings'"><i>settings</i></StyledLink>
    </nav>
  </header>
  <Suspense>
    <router-view :market="market" />
  </Suspense>
</template>

<script setup>
import { ref } from 'vue'
import { useRoute } from 'vue-router'
import { useSession } from '@/stores/session'
import StyledLink from '@/components/StyledLink'

const session = useSession()
const route = useRoute()
const marketId = route.params.id

const market = ref(null)
const mine = ref(false)
const url = '/api/market/' + marketId
await fetch(url)
  .then(r => r.json())
  .then(body => {
    market.value = body
    mine.value = market.value.Pubkey === session.pubkey
  })
  .catch(console.error)

</script>

<style scoped>
nav {
  display: flex;
  justify-content: center;
}

nav>a {
  margin: 0 3px;
}
</style>
