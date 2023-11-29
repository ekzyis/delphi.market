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
    </nav>
  </header>
  <Suspense>
    <router-view class="m-3" />
  </Suspense>
</template>

<script setup>
import { ref } from 'vue'
import { useRoute } from 'vue-router'
import StyledLink from '@/components/StyledLink'

const route = useRoute()
const marketId = route.params.id

const market = ref(null)
const url = '/api/market/' + marketId
await fetch(url)
  .then(r => r.json())
  .then(body => {
    market.value = body
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
