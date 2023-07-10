import * as Turbo from 'https://cdn.skypack.dev/@hotwired/turbo'
window.Turbo = Turbo

import {Application, Controller} from 'https://unpkg.com/@hotwired/stimulus/dist/stimulus.js'
window.Stimulus = Application.start()

customElements.define('mvp-stream-source', class extends HTMLElement {
  connectedCallback() {
    // console.log('<mvp-stream-source> connect')
    if (!this.streamSource) {
      this.streamSource = new EventSource(this.src)
      Turbo.connectStreamSource(this.streamSource)
      console.log("connecting to stream source ", this.src)
    }
  }

  disconnectedCallback() {
    // console.log('<mvp-stream-source> disconnect')
    requestAnimationFrame(() => {
      if (this.streamSource && !this.isConnected) {
        console.log("disconnecting from stream source ", this.src)
        Turbo.disconnectStreamSource(this.streamSource)
        this.streamSource.close()
        this.streamSource = null
      }
    })
  }

  get src() { return this.getAttribute("src") || "" }
})

Stimulus.register('menu', class extends Controller {
  static values = {
    open: { type: Boolean, default: false },
  };

  connect() {
    this.render()
  }

  toggle() {
    this.openValue = !this.openValue
  }

  openValueChanged() {
    this.render()
  }

  render() {
    let attr = this.element.attributes["aria-controls"]
    if (!attr) return;
    let popup = document.getElementById(attr.value)
    if (!popup) return;
    console.log('popup', popup)
    popup.style.display = (this.openValue ? '' : 'none')
  }

  closeIfOutside(ev) {

  }
});
