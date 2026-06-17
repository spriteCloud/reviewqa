// reviewqa serve — Phase A frontend (read-only viewer).
// Plain vanilla JS so the project keeps its no-build posture.

const $featureList = document.querySelector('[data-feature-list]')
const $docList = document.querySelector('[data-doc-list]')
const $content = document.querySelector('[data-content]')
const $projectName = document.querySelector('[data-project-name]')

let activeFeature = null
let activeDoc = null
let viewMode = 'pretty' // 'pretty' | 'raw'

async function fetchJSON (url) {
  const r = await fetch(url, { headers: { 'Accept': 'application/json' } })
  if (!r.ok) throw new Error(`${url} → ${r.status}`)
  return r.json()
}

function el (tag, attrs = {}, ...children) {
  const node = document.createElement(tag)
  for (const [k, v] of Object.entries(attrs)) {
    if (k === 'class') node.className = v
    else if (k === 'html') node.innerHTML = v
    else if (k.startsWith('on') && typeof v === 'function') node.addEventListener(k.slice(2), v)
    else node.setAttribute(k, v)
  }
  for (const c of children) {
    if (c == null) continue
    if (typeof c === 'string') node.appendChild(document.createTextNode(c))
    else node.appendChild(c)
  }
  return node
}

function highlightPlaceholders (text) {
  // Wrap "<thing>" placeholders in a span so the UI can colorize them.
  // Returns an HTML string — caller passes via html: so we never inject
  // unsanitized user content (this is read-only viewer of files the user
  // already has on disk; XSS surface is theoretical here, but escape
  // anyway for correctness).
  const escaped = text.replace(/[&<>]/g, ch => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;' }[ch]))
  return escaped.replace(/&quot;[^&]*?&quot;|"[^"]*"/g, m => `<span class="placeholder-arg">${m}</span>`)
}

function renderTags (tags) {
  if (!tags || !tags.length) return null
  return el('div', { class: 'tag-row' }, ...tags.map(t => el('span', { class: 'tag' }, t)))
}

function renderFeature (feature, gherkin) {
  $content.replaceChildren()
  const header = el('div', { class: 'feature-header' },
    el('h1', {}, feature.name || '(unnamed feature)'),
    el('div', { class: 'path' }, feature.path),
    renderTags(feature.tags),
    feature.narrative ? el('div', { class: 'narrative' }, feature.narrative) : null,
  )
  const tabs = el('div', { class: 'tab-row' },
    el('button', { class: 'tab ' + (viewMode === 'pretty' ? 'active' : ''), onclick: () => { viewMode = 'pretty'; renderFeature(feature, gherkin) } }, 'Pretty'),
    el('button', { class: 'tab ' + (viewMode === 'raw' ? 'active' : ''), onclick: () => { viewMode = 'raw'; renderFeature(feature, gherkin) } }, 'Raw .feature'),
  )
  $content.appendChild(header)
  $content.appendChild(tabs)

  if (viewMode === 'raw') {
    $content.appendChild(el('pre', { class: 'raw-block' }, gherkin))
    return
  }

  for (const sc of feature.scenarios) {
    const node = el('div', { class: 'scenario' },
      el('div', { class: 'scenario-name' }, sc.name || '(unnamed scenario)'),
      renderTags(sc.tags),
      el('ol', { class: 'steps' }, ...sc.steps.map(st =>
        el('li', {},
          el('span', { class: 'kw kw-' + st.keyword }, st.keyword),
          el('span', { class: 'text', html: highlightPlaceholders(st.text) }),
        ),
      )),
    )
    $content.appendChild(node)
  }
}

async function openFeature (path) {
  activeFeature = path
  activeDoc = null
  refreshSidebarSelection()
  try {
    const data = await fetchJSON('/api/feature?path=' + encodeURIComponent(path))
    renderFeature(data.feature, data.gherkin)
  } catch (err) {
    $content.replaceChildren(el('div', { class: 'error' }, 'Failed to load feature: ' + err.message))
  }
}

async function openDoc (path, kind) {
  activeDoc = path
  activeFeature = null
  refreshSidebarSelection()
  try {
    const data = await fetchJSON('/api/doc?path=' + encodeURIComponent(path))
    const escaped = data.body.replace(/[&<>]/g, ch => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;' }[ch]))
    const isHTML = path.endsWith('.html')
    const inner = isHTML
      ? el('div', { class: 'doc-content', html: data.body })
      : el('pre', { class: 'raw-block', html: escaped })
    $content.replaceChildren(
      el('div', { class: 'feature-header' },
        el('h1', {}, kind === 'catalogue' ? 'Test catalogue' : kind === 'summary' ? 'Stakeholder summary' : kind === 'findings' ? 'Bug-discovery ledger' : 'Doc'),
        el('div', { class: 'path' }, path),
      ),
      inner,
    )
  } catch (err) {
    $content.replaceChildren(el('div', { class: 'error' }, 'Failed to load doc: ' + err.message))
  }
}

function refreshSidebarSelection () {
  for (const li of $featureList.querySelectorAll('li')) {
    li.classList.toggle('active', li.dataset.path === activeFeature)
  }
  for (const li of $docList.querySelectorAll('li')) {
    li.classList.toggle('active', li.dataset.path === activeDoc)
  }
}

async function init () {
  try {
    const project = await fetchJSON('/api/project')
    $projectName.textContent = project.name
    document.title = `reviewqa · ${project.name}`

    $featureList.replaceChildren()
    if (!project.features.length) {
      $featureList.appendChild(el('li', { class: 'empty' }, 'No .feature files found in tests/e2e/features/'))
    } else {
      for (const f of project.features) {
        const meta = `${f.scenarios} scenario${f.scenarios === 1 ? '' : 's'}${f.tags && f.tags.length ? ' · ' + f.tags.join(' ') : ''}`
        $featureList.appendChild(el('li', { 'data-path': f.path, onclick: () => openFeature(f.path) },
          el('div', {}, f.name || f.path.split('/').pop()),
          el('span', { class: 'meta' }, meta),
        ))
      }
    }

    $docList.replaceChildren()
    if (!project.docs || !project.docs.length) {
      $docList.appendChild(el('li', { class: 'empty' }, '—'))
    } else {
      for (const d of project.docs) {
        $docList.appendChild(el('li', { 'data-path': d.path, onclick: () => openDoc(d.path, d.kind) },
          el('div', {}, d.title),
          el('span', { class: 'meta' }, d.path),
        ))
      }
    }
  } catch (err) {
    $content.replaceChildren(el('div', { class: 'error' }, 'Failed to load project: ' + err.message))
  }
}

init()
