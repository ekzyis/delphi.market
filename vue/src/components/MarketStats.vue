<template>
  <div class="mt-3">
    <div class="mb-1">
      <span>YES: {{  (currentYes * 100).toFixed(2) }}%</span>
      <span>NO: {{  (currentNo * 100).toFixed(2) }}%</span>
    </div>
    <div class="mb-2">
      <span>Volume: {{ volume }} sats</span>
    </div>
    <Line :data="chartData" :options="chartOptions" :plugins="chartPlugins" />
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRoute } from 'vue-router'
import { Line } from 'vue-chartjs'
import { Chart as ChartJS, Title, Tooltip, Legend, BarElement, CategoryScale, LinearScale, TimeScale, LineElement, PointElement } from 'chart.js'
import ago from 's-ago'

ChartJS.register(Title, Tooltip, Legend, BarElement, CategoryScale, LinearScale, TimeScale, LineElement, PointElement)

const route = useRoute()
const marketId = route.params.id

const stats = ref([])
const url = '/api/market/' + marketId + '/stats'
await fetch(url)
  .then(r => r.json())
  .then(body => {
    stats.value = body
  })
  .catch(console.error)

const getFilterData = key => {
  const y = []
  for (let i = 0; i < stats.value?.length - 1; i += 2) {
    const s1 = stats.value[i]
    const s2 = stats.value[i + 1]
    const yes = 'YES' in s1.y ? s1.y.YES : s2.y.YES
    const no = 'NO' in s1.y ? s1.y.NO : s2.y.NO
    const sum = yes + no
    volume = sum
    key === 'YES' ? y.push(yes / sum) : y.push(no / sum)
  }
  return y
}

let volume = 0
const yesData = getFilterData('YES')
const noData = getFilterData('NO')

const currentYes = yesData.at(-1)
const currentNo = noData.at(-1)

const chartData = {
  labels: stats.value ? stats.value.map(({ x }) => ago(new Date(x))) : [],
  datasets: [
    {
      label: 'YES',
      borderColor: '#35df8d',
      backgroundColor: '#35df8d',
      data: yesData,
      fill: 'origin'
    },
    {
      label: 'NO',
      borderColor: '#ff7386',
      backgroundColor: '#ff7386',
      data: noData,
      fill: 'origin'
    }
  ]
}
const chartOptions = {
  responsive: true,
  plugins: {
    background: {
      color: 'white'
    }
  }
}
const chartPlugins = [{
  id: 'background',
  beforeDraw: (chart, args, options) => {
    const { ctx } = chart
    ctx.save()
    ctx.globalCompositeOperation = 'destination-over'
    ctx.fillStyle = options.color
    ctx.fillRect(0, 0, chart.width, chart.height)
    ctx.restore()
  }
}]

</script>

<style scoped>
span {
  margin: 0 0.5em;
}

canvas {
  width: auto;
}
</style>
