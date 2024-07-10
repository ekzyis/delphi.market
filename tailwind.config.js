/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./server/router/pages/**/*.templ"],
  theme: {
    container: {
      center: true,
      padding: '1rem'
    },
    extend: {},
  },
  plugins: [],
}