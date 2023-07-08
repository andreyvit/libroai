const plugin = require('tailwindcss/plugin')

module.exports = plugin(function({addUtilities, theme, e}) {
  const values = theme('gridTemplate')
  const c = className => '.'+e(className)

  var utilities = {}
  for (let [key, value] of Object.entries(values)) {
    utilities[c(`grid-cols-autofill-${key}`)] = {
      'grid-template-columns': `repeat(auto-fill, ${value})`
    }
    utilities[c(`grid-cols-autofit-${key}`)] = {
      'grid-template-columns': `repeat(auto-fit, ${value})`
    }
    if (!/minmax/.test(value)) {
      utilities[c(`grid-cols-autofill-flex-${key}`)] = {
        'grid-template-columns': `repeat(auto-fill, minmax(${value}, 1fr))`
      }
    }
  }
  addUtilities(utilities, ['responsive'])
},
{
  theme: {
    gridTemplate: theme => ({
      ...theme('spacing'),
      'auto': 'auto',
      'max-content': 'max-content',
    })
  },
})
