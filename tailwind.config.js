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
  plugins: [
    function ({ addComponents }) {
      addComponents({
        '.container': {
          '@screen lg': {
            maxWidth: '768px',
          },
          '@screen xl': {
            maxWidth: '768px',
          },
        }
      })
    }
  ],
}