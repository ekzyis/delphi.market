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
      <router-link to="/user">settings</router-link>
      <router-link to="/user/invoices">invoices</router-link>
    </nav>
  </header>
  <Suspense>
    <router-view />
  </Suspense>
</template>

<style scoped>
nav {
  display: flex;
  justify-content: center;
}

nav>a {
  margin: 0 3px;
}
</style>
