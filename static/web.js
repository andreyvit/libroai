import * as Turbo from 'https://cdn.skypack.dev/@hotwired/turbo'
window.Turbo = Turbo
import {Application, Controller} from 'https://unpkg.com/@hotwired/stimulus/dist/stimulus.js'
window.Stimulus = Application.start()

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
