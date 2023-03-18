/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    'views/*.html',
    'static/*.html'
  ],
  theme: {
    extend: {},
  },
  plugins: [
    // 'postcss-import',
    require('@tailwindcss/forms'),
    require('@tailwindcss/typography'),
    // require('@tailwindcss/line-clamp'),
    // require('@tailwindcss/aspect-ratio'),
  ],
}
