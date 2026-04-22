import mermaid from 'https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.esm.min.mjs';

mermaid.initialize({ startOnLoad: false, securityLevel: 'loose' });

// ── Elements ──────────────────────────────────────────────────────────────
const fileInput  = document.getElementById('mmd-file-input');
const dropZone   = document.getElementById('viewer-drop-zone');
const outputEl   = document.getElementById('viewer-output');
const filenameEl = document.getElementById('viewer-filename');
const controls   = document.getElementById('viewer-controls');
const clearBtn   = document.getElementById('viewer-clear');
const zoomInBtn  = document.getElementById('zoom-in');
const zoomOutBtn = document.getElementById('zoom-out');
const zoomFitBtn = document.getElementById('zoom-fit');
const errorEl    = document.getElementById('viewer-error');

let renderCount = 0;
let panZoom = null;

// ── Helpers ───────────────────────────────────────────────────────────────
function showError(msg) {
  errorEl.textContent = msg;
  errorEl.hidden = false;
}

function clearError() {
  errorEl.hidden = true;
  errorEl.textContent = '';
}

function destroyPanZoom() {
  if (panZoom) { try { panZoom.destroy(); } catch (_) {} panZoom = null; }
}

// ── Render ────────────────────────────────────────────────────────────────
async function render(source, filename) {
  clearError();
  destroyPanZoom();

  const id = `mmd-${++renderCount}`;
  try {
    const { svg } = await mermaid.render(id, source);

    outputEl.innerHTML = svg;
    dropZone.classList.add('is-loaded');
    controls.hidden = false;
    filenameEl.textContent = filename;

    const svgEl = outputEl.querySelector('svg');
    if (svgEl && window.svgPanZoom) {
      requestAnimationFrame(() => {
        svgEl.setAttribute('width',  outputEl.clientWidth  || 800);
        svgEl.setAttribute('height', outputEl.clientHeight || 500);
        svgEl.style.maxWidth = 'none';

        panZoom = window.svgPanZoom(svgEl, {
          zoomEnabled:           true,
          panEnabled:            true,
          controlIconsEnabled:   false,
          dblClickZoomEnabled:   true,
          mouseWheelZoomEnabled: true,
          fit:    true,
          center: true,
          minZoom: 0.05,
          maxZoom: 20,
        });
      });
    }
  } catch (err) {
    showError(`Render error: ${err.message ?? String(err)}`);
  }
}

// ── Reset ─────────────────────────────────────────────────────────────────
function reset() {
  destroyPanZoom();
  outputEl.innerHTML = '';
  dropZone.classList.remove('is-loaded');
  controls.hidden = true;
  filenameEl.textContent = '';
  fileInput.value = '';
  clearError();
}

// ── File loading ──────────────────────────────────────────────────────────
function readFile(file) {
  if (!file) return;
  const reader = new FileReader();
  reader.onload  = e => render(e.target.result, file.name);
  reader.onerror = () => showError('Could not read file.');
  reader.readAsText(file);
}

// ── Zoom controls ─────────────────────────────────────────────────────────
zoomInBtn.addEventListener('click',  () => panZoom?.zoomIn());
zoomOutBtn.addEventListener('click', () => panZoom?.zoomOut());
zoomFitBtn.addEventListener('click', () => { panZoom?.fit(); panZoom?.center(); });
clearBtn.addEventListener('click', reset);

// ── File / drag events ────────────────────────────────────────────────────
fileInput.addEventListener('change', () => readFile(fileInput.files[0]));

dropZone.addEventListener('dragover', e => {
  e.preventDefault();
  dropZone.classList.add('drag-over');
});

dropZone.addEventListener('dragleave', e => {
  if (!dropZone.contains(e.relatedTarget)) dropZone.classList.remove('drag-over');
});

dropZone.addEventListener('drop', e => {
  e.preventDefault();
  dropZone.classList.remove('drag-over');
  const file = e.dataTransfer.files[0];
  if (file) readFile(file);
});

// ── Resize ────────────────────────────────────────────────────────────────
let resizeTimer;
window.addEventListener('resize', () => {
  clearTimeout(resizeTimer);
  resizeTimer = setTimeout(() => {
    if (!panZoom) return;
    const svgEl = outputEl.querySelector('svg');
    if (!svgEl) return;
    svgEl.setAttribute('width',  outputEl.clientWidth);
    svgEl.setAttribute('height', outputEl.clientHeight);
    panZoom.resize();
    panZoom.fit();
    panZoom.center();
  }, 150);
});
