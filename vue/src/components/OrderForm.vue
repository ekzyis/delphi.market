<template>
  <div>
    <header class="flex flex-row text-center justify-center pt-1">
      <nav>
        <StyledLink :to="'/market/' + marketId + '/form/buy'">buy</StyledLink>
        <StyledLink :to="'/market/' + marketId + '/form/sell'">sell</StyledLink>
      </nav>
    </header>
    <router-view />
  </div>
</template>

<script setup>
import { useRoute } from 'vue-router'
import StyledLink from '@/components/StyledLink'

const route = useRoute()
const marketId = route.params.id

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
