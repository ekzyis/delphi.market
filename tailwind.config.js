/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./server/router/pages/**/*.templ"],
  theme: {
    container: {
      center: true,
      padding: '1rem'
    },
    extend: {
      colors: {
        'background': '191d21',
        'muted': '#6c757d',
      },
    },
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