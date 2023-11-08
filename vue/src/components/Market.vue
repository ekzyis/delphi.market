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
</template>

<script setup>
import { ref } from 'vue'
import { useRoute } from 'vue-router'

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
