const { addDynamicIconSelectors } = require('@iconify/tailwind')

/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./**/*.templ', './**/*.ts'],
  theme: {
    extend: {},
  },
  plugins: [
    addDynamicIconSelectors(),
  ],
}

