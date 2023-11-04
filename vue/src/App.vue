<template>
  <NavBar />
  <div id="container">
    <router-view />
  </div>
</template>

<script>
import NavBar from './components/NavBar'
export default {
  name: 'App',
  components: { NavBar }
}
</script>

<script setup>
import { useSession } from './stores/session';
const session = useSession()
session.init()
</script>

<style>
html,
body {
  background-color: #091833;
  color: #ffffff;
}

#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #ffffff;
  margin-top: 1em;
}

#container {
  margin: 1em;
}
</style>
