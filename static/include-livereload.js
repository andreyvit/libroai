if (/^localdev\.|^localhost:/.test(location.host) && !document.getElementById('livereload-script')) {
  let script = document.createElement('script')
  script.id = 'livereload-script'
  script.src = 'http://' + location.host.split(':')[0] + ':35729/livereload.js?snipver=1'
  document.head.appendChild(script)
}
