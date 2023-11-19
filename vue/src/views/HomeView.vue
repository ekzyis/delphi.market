<template>
  <!-- eslint-disable -->
  <div class="my-3">
    <pre>
     _      _       _     _ 
  __| | ___| |_ __ | |__ (_)
 / _` |/ _ \ | '_ \| '_ \| |
| (_| |  __/ | |_) | | | | |
 \__,_|\___|_| .__/|_| |_|_|
             |_|.market     </pre>

  </div>
  <!-- eslint-enable -->
  <Suspense>
    <MarketList />
  </Suspense>
</template>

<script setup>
import MarketList from '@/components/MarketList'
</script>
