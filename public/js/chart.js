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


function transformPoint({ X, Y }) {
  return { x: new Date(X), y: Y * 100 }
}

const no = JSON.parse($("#chart-data").getAttribute("chart-data-p0")).map(transformPoint)
const yes = JSON.parse($("#chart-data").getAttribute("chart-data-p1")).map(transformPoint)

const config = {
  type: 'line',
  data: {
    datasets: [
      {
        data: yes,
        backgroundColor: '#149e613d',
        borderColor: '#149e613d',
        borderWidth: 3,
        tension: 0, // draw straight lines instead of bezier curves
        pointStyle: false
      },
      {
        data: no,
        backgroundColor: '#f5395e3d',
        borderColor: '#f5395e3d',
        borderWidth: 3,
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