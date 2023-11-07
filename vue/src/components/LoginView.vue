<template>
  <!-- eslint-disable -->
  <div class="my-3">
    <pre>
 _             _       
| | ___   __ _(_)_ __  
| |/ _ \ / _` | | '_ \ 
| | (_) | (_| | | | | |
|_|\___/ \__, |_|_| |_|
         |___/         </pre>
  </div>
  <!-- eslint-enable -->
  <Suspense>
    <LoginQRCode class="flex justify-center m-3" />
  </Suspense>
</template>

<script setup>
import LoginQRCode from './LoginQRCode.vue'
</script>
