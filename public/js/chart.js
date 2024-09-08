import {
  Chart,
  LineController,
  TimeScale,
  LinearScale,
  PointElement,
  LineElement,
} from 'chart.js'
import 'chartjs-adapter-dayjs-4/dist/chartjs-adapter-dayjs-4.esm';

Chart.register(
  LineController,
  TimeScale,
  LinearScale,
  PointElement,
  LineElement,
);

const element = document.getElementById('chart')

// TODO: get data from server
const yes = [
  { x: new Date('2024-01-01'), y: 20 },
  { x: new Date('2024-02-01'), y: 41 },
  { x: new Date('2024-03-01'), y: 53 }
]

const no = yes.map(({ x, y }) => ({ x, y: 100 - y }))

const config = {
  type: 'line',
  data: {
    datasets: [
      {
        data: yes,
        backgroundColor: '#149e613d',
        borderColor: '#149e613d',
        borderWidth: 5,
        tension: 0, // draw straight lines instead of bezier curves
        pointStyle: false
      },
      {
        data: no,
        backgroundColor: '#f5395e3d',
        borderColor: '#f5395e3d',
        borderWidth: 5,
        tension: 0,
        pointStyle: false
      }
    ]
  },
  options: {
    scales: {
      x: {
        type: 'time',
        time: { unit: 'month' }
      },
      y: {
        min: 0,
        max: 100
      }
    }
  }
}

new Chart(element, config);