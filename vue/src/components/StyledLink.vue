<template>
  <router-link :to="to" :class="{ selected }">
    <slot></slot>
  </router-link>
</template>

<script setup>

import { computed, defineProps } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const props = defineProps(['to'])
const selected = computed(() => props.to !== '/' ? route.path.startsWith(props.to) : route.path === props.to, [route.path])

</script>
