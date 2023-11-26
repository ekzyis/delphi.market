<template>
  <header class="flex flex-row text-center justify-center pt-1">
    <nav>
      <router-link to="/">market</router-link>
      <router-link to="/user" v-if="session.isAuthenticated">user</router-link>
      <router-link to="/login" v-else-if="session.isAuthenticated === false" href="/login">login</router-link>
      <router-link disabled to="/" v-else>...</router-link>
    </nav>
  </header>
</template>

<script setup>
import { useSession } from '@/stores/session'
const session = useSession()
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
