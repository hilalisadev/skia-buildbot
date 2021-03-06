/**
 * @module particles-sk
 * @description <h2><code>particles-sk</code></h2>
 *
 * <p>
 *   The main application element for particles in Skia.
 * </p>
 *
 */
import { $$ } from 'common-sk/modules/dom'
import { define } from 'elements-sk/define'
import { html, render } from 'lit-html'

import 'elements-sk/spinner-sk'

const CanvasKitInit = require('../../build/canvaskit/canvaskit.js');

const ZOOM_IN_FACTOR = 1.1; // 10%
const ZOOM_OUT_FACTOR = 1/ZOOM_IN_FACTOR;

// This element might be loaded from a different site, and that means we need
// to be careful about how we construct the URL back to the canvas.wasm file.
// Start by recording the script origin.
const scriptOrigin = new URL(document.currentScript.src).origin;
const kitReady = CanvasKitInit({
  locateFile: (file) => {
    return `${scriptOrigin}/static/${file}`;
  },
}).ready();

const loadingTemplate = (ele) => html`
<div class=player-loading title="Loading particles and engine."
     style='width: ${ele._config.width}px; height: ${ele._config.height}px;'>
  <div>Loading</div>
  <spinner-sk active></spinner-sk>
</div>`;

const runningTemplate = (ele) => html`
<div class=container>
  <!-- It would be more mobile friendly to use pointermove, but Safari doesn't support it-->
  <canvas id=player
          @wheel=${ele._wheel}
          @mousemove=${ele._drag}
          width=${ele._config.width * window.devicePixelRatio}
          height=${ele._config.height * window.devicePixelRatio}
          style='width: ${ele._config.width}px; height: ${ele._config.height}px;'>
    Your browser does not support the canvas tag.
  </canvas>
</div>`;

define('particles-player-sk', class extends HTMLElement {
  constructor() {
    super();

    this._engine = {
      kit:       null, // CanvasKit instance
      context:   null, // CK context.
      animation: null, // Particles instance
      surface:   null, // SkSurface
      canvas:    null, // Cached SkCanvas (surface.getCanvas()).
    };

    this._state = {
      loading:        true,
      paused:         this.hasAttribute('paused'),
      time:           0, // a monotonically increasing amount of ms
      lastTs:         0, // last time stamp we had a frame
    };


    this._lastDrag = null;
    this._zoomLevel = 1.0;
  }

  connectedCallback() {
    this._config = {
      width:      this.hasAttribute('width')  ? +this.getAttribute('width')  : 256,
      height:     this.hasAttribute('height') ? +this.getAttribute('height') : 256,
      bgcolor:    this.hasAttribute('bgcolor') ? +this.getAttribute('bgcolor') : 4278190080, // black
    };

    this.render();
  }

  _drag(e) {
    if (!e.buttons || !e.shiftKey) { // ignore movements unless shift is held
      this._lastDrag = null;
      return;
    }
    if (this._lastDrag) {
      const dx = e.clientX - this._lastDrag[0];
      const dy = e.clientY - this._lastDrag[1];

      const canvas = this._engine.canvas.translate(dx / this._zoomLevel,
                                                   dy / this._zoomLevel);
    }
    this._lastDrag = [e.clientX, e.clientY];

  }

  _drawFrame() {
    if (!this._engine.animation || !this._engine.canvas) {
      return;
    }
    window.requestAnimationFrame(this._drawFrame.bind(this));
    if (!this._state.lastTs) {
      this._engine.animation.start(0, true);
      this._state.lastTs = Date.now();
    }

    if (this.isPlaying()) {
      this._state.time += (Date.now() - this._state.lastTs);
    }
    this._state.lastTs = Date.now();

    this._engine.kit.setCurrentContext(this._engine.context);
    this._engine.canvas.clear(this._config.bgcolor);

    this._engine.animation.update(this._state.time / 1000.0);
    this._engine.animation.draw(this._engine.canvas);
    this._engine.surface.flush();
  }

  initialize(config) {
    this._config.width = config.width;
    this._config.height = config.height;

    this.render();
    return kitReady.then((ck) => {
      this._engine.kit = ck;
      this._initializeParticles(config.json);
      this.render();
    });
  }

  _initializeParticles(particlesJSON, assets) {
    this._state.loading = false;

    // Rebuild the surface only if needed.
    if (!this._engine.surface ||
        this._engine.surface.width  != this._config.width ||
        this._engine.surface.height != this._config.height) {

      this.render();

      this._engine.surface && this._engine.surface.delete();
      let canvasEle = $$('#player', this);
      this._engine.surface = this._engine.kit.MakeCanvasSurface(canvasEle);
      if (!this._engine.surface) {
        throw new Error('Could not make SkSurface.');
      }
      // We don't need to call .delete() on the canvas because
      // the parent surface will do that for us.
      this._engine.canvas = this._engine.surface.getCanvas();
      this._engine.context = this._engine.kit.currentContext();
    }

    this._engine.animation && this._engine.animation.delete();

    this._engine.animation = this._engine.kit.MakeParticles(
                                          JSON.stringify(particlesJSON));
    if (!this._engine.animation) {
      throw new Error('Could not parse Particles JSON.');
    }

    this._engine.canvas.clear(this._config.bgcolor);
    // Center the animation
    this.resetView();

    this.restartAnimation();

    this._drawFrame();
  }

  isPlaying() {
    return !this._state.paused;
  }

  play() {
    if (!this.isPlaying()) {
      this._state.paused = false;
    }
    this.render();
  }

  pause() {
    if (this.isPlaying()) {
      this._state.paused = true;
    }
  }

  render() {
    render(this._state.loading
               ? loadingTemplate(this)
               : runningTemplate(this),
           this, {eventContext: this});
  }

  resetView() {
    const ck = this._engine.kit;
    const canvas = this._engine.canvas;
    // Reset to identity
    const tt = canvas.getTotalMatrix();
    const itt = ck.SkMatrix.invert(tt);
    canvas.concat(itt);
    // Zoom to the middle of the animation
    canvas.translate(this._config.width/2, this._config.height/2);
    this._zoomLevel = 1.0;
  }

  restartAnimation() {
    this._state.time = 0;
    this._state.lastTs = 0;
  }

  _wheel(e) {
    e.preventDefault();
    e.stopPropagation();

    let zoom = 0;
    if (e.deltaY < 0) {
      zoom = ZOOM_IN_FACTOR;
    } else {
      zoom = ZOOM_OUT_FACTOR;
    }
    this._zoomLevel *= zoom;
    const ck = this._engine.kit;
    const canvas = this._engine.canvas;

    const tt = canvas.getTotalMatrix();
    const itt = ck.SkMatrix.invert(tt);
    const pts = [e.clientX, e.clientY];
    ck.SkMatrix.mapPoints(itt, pts); // Transform DOM pts into canvas space

    let matr = ck.SkMatrix.scaled(zoom, zoom, pts[0], pts[1]);
    canvas.concat(matr);

  }
});
