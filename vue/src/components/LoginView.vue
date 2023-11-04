<template>
    <!-- eslint-disable -->
    <pre>
 _             _       
| | ___   __ _(_)_ __  
| |/ _ \ / _` | | '_ \ 
| | (_) | (_| | | | | |
|_|\___/ \__, |_|_| |_|
         |___/         </pre>
    <!-- eslint-enable -->
    <Suspense>
        <LoginQRCode class="flex justify-center m-3" />
    </Suspense>
</template>

<script setup>
import LoginQRCode from './LoginQRCode.vue';
</script>

<style scoped>
pre {
    font-family: monospace;
    font-weight: bold;
}
</style>
