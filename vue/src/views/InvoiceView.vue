<template>
  <!-- eslint-disable -->
  <div class="my-3">
    <pre>
 _  _    ___ ____  
| || |  / _ \___ \ 
| || |_| | | |__) |
|__   _| |_| / __/ 
   |_|  \___/_____|</pre>
  </div>
  <!-- eslint-enable -->
  <Suspense>
    <Invoice />
  </Suspense>
</template>

<script setup>
import Invoice from '@/components/Invoice'
</script>
