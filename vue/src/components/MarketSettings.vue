<template>
  <div id="container2" class="mt-3 mx-3">
    <div class="grid grid-cols-2 my-3 items-center">
      <label>market settlement</label>
      <div class="grid grid-cols-2 my-3">
        <button :class="yesClass" class="label success font-mono mx-1" @click.prevent="() => click('YES')">YES</button>
        <button :class="noClass" class="label error font-mono mx-1" @click.prevent="() => click('NO')">NO</button>
      </div>
      <div class="col-span-2 mb-3" v-if="selected">
        <p><b>Are you sure you want to settle this market?</b></p>
        <p>
          This will cancel all pending orders and halt trading indefinitely.
          Users with winning shares will receive 100 sats per winning share. Users with losing shares receive nothing.
        </p>
        <p class="red"><b>You cannot undo this action.</b></p>
      </div>
      <button class="col-span-2" v-if="selected" @click.prevent="confirm">confirm</button>
    </div>
    <div v-if="err" class="red text-center">{{ err }}</div>
    <div v-if="success" class="green text-center">{{ success }}</div>
  </div>
</template>

<script setup>

import { computed, defineProps, ref } from 'vue'

const props = defineProps(['market'])
const market = ref(props.market)

const err = ref(null)
const success = ref(null)

const selected = ref(null)
const yesClass = computed(() => selected.value === 'YES' ? ['active'] : [])
const noClass = computed(() => selected.value === 'NO' ? ['active'] : [])
const click = (sel) => {
  selected.value = selected.value === sel ? null : sel
}

const confirm = async () => {
  const sid = market.value.Shares.find(s => s.Description === selected.value).Id
  const url = '/api/market/' + market.value.Id + '/settle'
  const body = JSON.stringify({ sid })
  try {
    await fetch(url, { method: 'POST', headers: { 'Content-type': 'application/json' }, body })
  } catch (err) {
    console.error(err)
  }
}

</script>

<style scoped>
textarea {
  padding: 0 0.25em;
}

.label {
  width: auto;
  padding: 0.2em 1em
}

.success.active {
  background-color: #35df8d;
  color: white;
}

.error.active {
  background-color: #ff7386;
  color: white;
}

#container2 {
  max-width: 33vw;
}

@media only screen and (max-width: 1024px) {
  #container2 {
    max-width: 80vw;
  }
}

@media only screen and (max-width: 768px) {
  #container2 {
    max-width: 90vw;
  }
}

@media only screen and (max-width: 640px) {
  #container2 {
    max-width: 100vw;
  }
}
</style>
