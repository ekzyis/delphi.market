<template>
  <div id="container">
    <NavBar />
    <router-view />
  </div>
</template>

<script setup>
import { useSession } from './stores/session'
const session = useSession()
session.checkSession()
</script>

<!-- eslint-disable -->
<!-- eslint wants to combine this <script> and <script setup> which breaks the code ... -->
<script>
import NavBar from './components/NavBar'
export default {
  name: 'App',
  components: { NavBar }
}
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
  margin: 1em auto;
  width: fit-content;
}
</style>
