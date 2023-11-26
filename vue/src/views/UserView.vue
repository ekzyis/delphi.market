<template>
  <!-- eslint-disable -->
  <div class="my-3">
    <pre>

 _   _ ___  ___ _ __ 
| | | / __|/ _ \ '__|
| |_| \__ \  __/ |   
 \__,_|___/\___|_|   </pre>
  </div>
  <!-- eslint-enable -->
  <header class="flex flex-row text-center justify-center pt-1">
    <nav>
      <StyledLink to="/user/settings">settings</StyledLink>
      <StyledLink to="/user/invoices">invoices</StyledLink>
      <StyledLink to="/user/orders">orders</StyledLink>
    </nav>
  </header>
  <Suspense>
    <router-view />
  </Suspense>
</template>

<script setup>
import StyledLink from '@/components/StyledLink'
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
