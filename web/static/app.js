'use strict';

/* ─── State ──────────────────────────────────────────────────────────────── */
const S = {
  user:        null,
  bookmarks:   [],
  bmTotal:     0,
  bmOffset:    0,
  bmLimit:     50,
  bmFilter:    '',
  bmFolder:    '',
  wallpapers:  [],
  sessions:    [],
  editingBmId: null,
};

/* ─── API client ─────────────────────────────────────────────────────────── */
async function api(method, path, body) {
  const opts = {
    method,
    credentials: 'same-origin',
    headers: body ? { 'Content-Type': 'application/json' } : {},
    body: body ? JSON.stringify(body) : undefined,
  };
  const r = await fetch('/api' + path, opts);
  if (r.status === 204) return null;
  const json = await r.json().catch(() => ({}));
  if (!r.ok) {
    const err = new Error(json.error || r.statusText || 'Erreur réseau');
    err.code   = json.code;
    err.status = r.status;
    throw err;
  }
  return json;
}

const GET  = path       => api('GET',    path);
const POST = (path, b)  => api('POST',   path, b);
const PUT  = (path, b)  => api('PUT',    path, b);
const DEL  = (path, b)  => api('DELETE', path, b);

/* ─── DOM helpers ────────────────────────────────────────────────────────── */
const $  = id => document.getElementById(id);
const el = (tag, cls, text) => {
  const e = document.createElement(tag);
  if (cls)  e.className   = cls;
  if (text) e.textContent = text;
  return e;
};

function show(id)  { $(id).classList.remove('hidden'); }
function hide(id)  { $(id).classList.add('hidden'); }
function toggle(id, on) { $(id).classList.toggle('hidden', !on); }

function setError(id, msg) { $(id).textContent = msg || ''; }

/* ─── Rain canvas ────────────────────────────────────────────────────────── */
function initRain() {
  const canvas = $('rain-canvas');
  const ctx    = canvas.getContext('2d');
  let drops    = [];

  function resize() {
    canvas.width  = window.innerWidth;
    canvas.height = window.innerHeight;
    const count = Math.floor(canvas.width / 14);
    drops = Array.from({ length: count }, () => ({
      x:       Math.random() * canvas.width,
      y:       Math.random() * canvas.height,
      speed:   7 + Math.random() * 10,
      length:  12 + Math.random() * 18,
      opacity: 0.04 + Math.random() * 0.12,
    }));
  }

  function frame() {
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    for (const d of drops) {
      ctx.beginPath();
      ctx.strokeStyle = `rgba(190,215,255,${d.opacity})`;
      ctx.lineWidth   = 1;
      ctx.moveTo(d.x, d.y);
      ctx.lineTo(d.x - 0.8, d.y + d.length);
      ctx.stroke();
      d.y += d.speed * 0.016 * 60; // ~60fps target
      if (d.y - d.length > canvas.height) {
        d.y = -d.length;
        d.x = Math.random() * canvas.width;
      }
    }
    requestAnimationFrame(frame);
  }

  window.addEventListener('resize', resize);
  resize();
  requestAnimationFrame(frame);
}

/* ─── Clock ──────────────────────────────────────────────────────────────── */
function startClock() {
  const DAYS   = ['Dimanche','Lundi','Mardi','Mercredi','Jeudi','Vendredi','Samedi'];
  const MONTHS = ['janvier','février','mars','avril','mai','juin',
                  'juillet','août','septembre','octobre','novembre','décembre'];

  function isoWeek(d) {
    const date = new Date(Date.UTC(d.getFullYear(), d.getMonth(), d.getDate()));
    date.setUTCDate(date.getUTCDate() + 4 - (date.getUTCDay() || 7));
    const y0 = new Date(Date.UTC(date.getUTCFullYear(), 0, 1));
    return Math.ceil((((date - y0) / 86400000) + 1) / 7);
  }

  function tick() {
    const now = new Date();
    const h = String(now.getHours()).padStart(2, '0');
    const m = String(now.getMinutes()).padStart(2, '0');
    $('clock').textContent    = `${h}:${m}`;
    $('date-line').textContent =
      `${DAYS[now.getDay()]} ${now.getDate()} ${MONTHS[now.getMonth()]} ${now.getFullYear()} · S${isoWeek(now)}`;
  }

  tick();
  setInterval(tick, 1000);
}

/* ─── Wallpaper ──────────────────────────────────────────────────────────── */
async function loadWallpaper() {
  const wps = S.wallpapers;
  if (!wps.length) return; // gradient fallback stays

  const pinned = wps.filter(w => w.is_pinned);
  const pool   = pinned.length ? pinned : wps;
  const wp     = pool[Math.floor(Math.random() * pool.length)];
  const url    = `/media/${S.user.id}/${wp.filename}`;
  const ext    = wp.filename.split('.').pop().toLowerCase();
  const isVid  = ['mp4','webm'].includes(ext);

  if (isVid) {
    const v   = $('bg-video');
    v.src     = url;
    v.classList.remove('bg-hidden');
    $('bg-gradient').classList.add('bg-hidden');
    v.addEventListener('loadeddata', () => sampleLuminance(v), { once: true });
  } else {
    const img = $('bg-image');
    img.src   = url;
    img.classList.remove('bg-hidden');
    $('bg-gradient').classList.add('bg-hidden');
    img.addEventListener('load', () => sampleLuminance(img), { once: true });
  }
}

function sampleLuminance(media) {
  try {
    const c = document.createElement('canvas');
    c.width = c.height = 16;
    const ctx = c.getContext('2d');
    ctx.drawImage(media, 0, 0, 16, 16);
    const d = ctx.getImageData(0, 0, 16, 16).data;
    let sum = 0;
    for (let i = 0; i < d.length; i += 4) {
      sum += 0.299 * d[i] + 0.587 * d[i+1] + 0.114 * d[i+2];
    }
    document.documentElement.dataset.theme = (sum / (d.length / 4)) > 128 ? 'light' : 'dark';
  } catch {
    // cross-origin or security restriction — keep current theme
  }
}

/* ─── Search ─────────────────────────────────────────────────────────────── */
const ENGINES = {
  duckduckgo: 'https://duckduckgo.com/?q=',
  google:     'https://www.google.com/search?q=',
  brave:      'https://search.brave.com/search?q=',
  bing:       'https://www.bing.com/search?q=',
  kagi:       'https://kagi.com/search?q=',
};

function engineURL() {
  const u = S.user;
  if (!u) return ENGINES.duckduckgo;
  if (u.search_engine === 'custom' && u.search_engine_url) return u.search_engine_url;
  return ENGINES[u.search_engine] || ENGINES.duckduckgo;
}

function handleSearch(e) {
  e.preventDefault();
  const q = $('search-input').value.trim();
  if (!q) return;

  const m = q.match(/^!(\w+)\s*(.*)/s);
  if (m) {
    const bang = m[1].toLowerCase();
    const rest = m[2].trim();
    switch (bang) {
      case 'bm': case 'edit':
        openBookmarks();
        $('search-input').value = '';
        return;
      case 'g':
        open(`https://www.google.com/search?q=${encodeURIComponent(rest)}`);
        return;
      case 'yt':
        open(`https://www.youtube.com/results?search_query=${encodeURIComponent(rest)}`);
        return;
      case 'gh':
        open(`https://github.com/search?q=${encodeURIComponent(rest)}`);
        return;
      case 'hub':
        open(`https://hub.docker.com/search?q=${encodeURIComponent(rest)}`);
        return;
      default:
        // pass all other DDG bangs through
        open(`https://duckduckgo.com/?q=${encodeURIComponent(q)}`);
        return;
    }
  }
  open(engineURL() + encodeURIComponent(q));
}

function open(url) { window.location.href = url; }

/* ─── Login ──────────────────────────────────────────────────────────────── */
async function handleLogin(e) {
  e.preventDefault();
  setError('login-error', '');
  const email    = $('login-email').value.trim();
  const password = $('login-password').value;
  const totpCode = $('login-totp').value.trim();

  try {
    await POST('/auth/login', { email, password, totp_code: totpCode || undefined });
    await boot();
  } catch (err) {
    if (err.code === 'TOTP_REQUIRED') {
      show('totp-group');
      $('login-totp').focus();
      setError('login-error', 'Code TOTP requis');
    } else {
      setError('login-error', err.message);
    }
  }
}

async function handleForgot(e) {
  e.preventDefault();
  setError('forgot-msg', '');
  const email = $('forgot-email').value.trim();
  try {
    await POST('/auth/forgot-password', { email });
    setError('forgot-msg', 'Si l'adresse existe, un email a été envoyé.');
  } catch {
    setError('forgot-msg', 'Si l'adresse existe, un email a été envoyé.');
  }
}

function showForgotForm() {
  hide('login-form');
  hide('forgot-link');
  show('forgot-form');
}

function showLoginForm() {
  hide('forgot-form');
  show('forgot-link');
  show('login-form');
  setError('forgot-msg', '');
}

/* ─── Logout ─────────────────────────────────────────────────────────────── */
async function logout() {
  try { await POST('/auth/logout'); } catch {}
  S.user = null;
  hide('view-home');
  hide('overlay-bookmarks');
  hide('overlay-settings');
  hide('overlay-admin');
  show('view-login');
  $('login-email').focus();
}

/* ─── Bookmark panel ─────────────────────────────────────────────────────── */
function openBookmarks() {
  show('overlay-bookmarks');
  loadBookmarks();
}

function closeBookmarks() {
  hide('overlay-bookmarks');
}

async function loadBookmarks() {
  const params = new URLSearchParams({
    offset: S.bmOffset,
    limit:  S.bmLimit,
  });
  if (S.bmFilter) params.set('search', S.bmFilter);
  if (S.bmFolder) params.set('folder', S.bmFolder);

  try {
    const data = await GET(`/bookmarks?${params}`);
    S.bookmarks = data.bookmarks || data.items || data || [];
    S.bmTotal   = data.total ?? S.bookmarks.length;
    renderBookmarks();
  } catch (err) {
    $('bm-list').textContent = 'Erreur : ' + err.message;
  }
}

function renderBookmarks() {
  const list = $('bm-list');
  list.innerHTML = '';

  if (!S.bookmarks.length) {
    const empty = el('p', 'bm-empty', 'Aucun marque-page. Utilisez + Ajouter ou Importer.');
    list.appendChild(empty);
    return;
  }

  // Group by folder
  const byFolder = new Map();
  const noFolder = [];
  for (const bm of S.bookmarks) {
    if (bm.folder) {
      if (!byFolder.has(bm.folder)) byFolder.set(bm.folder, []);
      byFolder.get(bm.folder).push(bm);
    } else {
      noFolder.push(bm);
    }
  }

  // Render folders
  for (const [folder, bms] of [...byFolder.entries()].sort()) {
    const section = el('div', 'bm-section');
    section.appendChild(el('div', 'bm-section-name', folder));
    const count = bms.length;
    bms.forEach((bm, i) => {
      section.appendChild(makeBmItem(bm, i < count - 1 ? '├' : '└'));
    });
    list.appendChild(section);
  }

  // Render unfoldered
  if (noFolder.length) {
    const section = el('div', 'bm-section');
    section.appendChild(el('div', 'bm-section-name', 'Sans dossier'));
    noFolder.forEach(bm => section.appendChild(makeBmItem(bm, '·')));
    list.appendChild(section);
  }

  // Pagination
  if (S.bmTotal > S.bmLimit) {
    const pag = el('div', 'pagination');
    const prev = el('button', 'page-btn', '← Précédent');
    prev.disabled = S.bmOffset === 0;
    prev.onclick  = () => { S.bmOffset = Math.max(0, S.bmOffset - S.bmLimit); loadBookmarks(); };

    const info = el('span', 'page-info',
      `${S.bmOffset + 1}–${Math.min(S.bmOffset + S.bmLimit, S.bmTotal)} / ${S.bmTotal}`);

    const next = el('button', 'page-btn', 'Suivant →');
    next.disabled = S.bmOffset + S.bmLimit >= S.bmTotal;
    next.onclick  = () => { S.bmOffset += S.bmLimit; loadBookmarks(); };

    pag.append(prev, info, next);
    list.appendChild(pag);
  }

  // Populate folder filter
  const sel = $('bm-folder-filter');
  const current = sel.value;
  sel.innerHTML = '<option value="">Tous les dossiers</option>';
  for (const f of [...byFolder.keys()].sort()) {
    const opt = document.createElement('option');
    opt.value       = f;
    opt.textContent = f;
    if (f === current) opt.selected = true;
    sel.appendChild(opt);
  }
}

function makeBmItem(bm, glyph) {
  const wrap = el('div', 'bm-item');

  const g = el('span', 'bm-tree-glyph', glyph);

  const link = el('a', 'bm-link', bm.title || bm.url);
  link.href   = bm.url;
  link.target = '_blank';
  link.rel    = 'noopener noreferrer';

  const urlSpan = el('span', 'bm-url', new URL(bm.url).hostname);

  const actions = el('div', 'bm-actions');

  const editBtn = el('button', 'icon-btn', '✎');
  editBtn.title = 'Modifier';
  editBtn.onclick = () => openEditBookmark(bm);

  const delBtn = el('button', 'icon-btn danger', '✕');
  delBtn.title = 'Supprimer';
  delBtn.onclick = () => deleteBookmark(bm.id);

  actions.append(editBtn, delBtn);
  wrap.append(g, link, urlSpan, actions);

  // Tags row
  if (bm.tags && bm.tags.length) {
    const tagsRow = el('div', 'bm-tags-row');
    for (const tag of bm.tags) {
      tagsRow.appendChild(el('span', 'tag-chip', '#' + tag.name));
    }
    // Return a fragment with both rows
    const frag = document.createDocumentFragment();
    frag.append(wrap, tagsRow);
    return frag;
  }

  return wrap;
}

function openAddBookmark() {
  S.editingBmId = null;
  $('modal-bm-title').textContent   = 'Ajouter un marque-page';
  $('modal-bm-url').value           = '';
  $('modal-bm-title-input').value   = '';
  $('modal-bm-folder').value        = '';
  $('modal-bm-tags').value          = '';
  setError('modal-bm-error', '');
  show('modal-bookmark');
  $('modal-bm-url').focus();
}

function openEditBookmark(bm) {
  S.editingBmId = bm.id;
  $('modal-bm-title').textContent   = 'Modifier le marque-page';
  $('modal-bm-url').value           = bm.url;
  $('modal-bm-title-input').value   = bm.title;
  $('modal-bm-folder').value        = bm.folder || '';
  $('modal-bm-tags').value          = (bm.tags || []).map(t => t.name).join(', ');
  setError('modal-bm-error', '');
  show('modal-bookmark');
  $('modal-bm-title-input').focus();
}

async function saveBookmark() {
  setError('modal-bm-error', '');
  const url    = $('modal-bm-url').value.trim();
  const title  = $('modal-bm-title-input').value.trim();
  const folder = $('modal-bm-folder').value.trim() || null;
  const tags   = $('modal-bm-tags').value.split(',').map(t => t.trim()).filter(Boolean);

  if (!url || !title) {
    setError('modal-bm-error', 'URL et titre requis.');
    return;
  }

  try {
    if (S.editingBmId) {
      await PUT(`/bookmarks/${S.editingBmId}`, { url, title, folder, tags });
    } else {
      await POST('/bookmarks', { url, title, folder, tags });
    }
    hide('modal-bookmark');
    S.bmOffset = 0;
    loadBookmarks();
  } catch (err) {
    setError('modal-bm-error', err.message);
  }
}

async function deleteBookmark(id) {
  if (!confirm('Supprimer ce marque-page ?')) return;
  try {
    await DEL(`/bookmarks/${id}`);
    loadBookmarks();
  } catch (err) {
    alert(err.message);
  }
}

async function exportBookmarks() {
  try {
    const r = await fetch('/api/bookmarks/export', { credentials: 'same-origin' });
    if (!r.ok) throw new Error('Échec export');
    const blob = await r.blob();
    const a    = document.createElement('a');
    a.href     = URL.createObjectURL(blob);
    a.download = 'bookmarks.html';
    a.click();
    URL.revokeObjectURL(a.href);
  } catch (err) {
    alert(err.message);
  }
}

async function importBookmarks(file) {
  if (!file) return;
  const form = new FormData();
  form.append('file', file);
  try {
    const r = await fetch('/api/bookmarks/import', {
      method: 'POST',
      credentials: 'same-origin',
      body: form,
    });
    const json = await r.json().catch(() => ({}));
    if (!r.ok) throw new Error(json.error || 'Erreur import');
    alert(`Import terminé : ${json.imported ?? '?'} ajoutés, ${json.skipped ?? '?'} ignorés.`);
    S.bmOffset = 0;
    loadBookmarks();
  } catch (err) {
    alert(err.message);
  }
}

/* ─── Settings panel ─────────────────────────────────────────────────────── */
function openSettings() {
  renderSettingsTab('profile');
  show('overlay-settings');
}

function closeSettings() {
  hide('overlay-settings');
}

function renderSettingsTab(tabName) {
  // Update tab highlights
  document.querySelectorAll('#settings-tabs .tab').forEach(t => {
    t.classList.toggle('active', t.dataset.tab === tabName);
  });

  const body = $('settings-body');
  body.innerHTML = '';

  switch (tabName) {
    case 'profile':     body.appendChild(buildProfileTab());     break;
    case 'security':    body.appendChild(buildSecurityTab());    break;
    case 'engine':      body.appendChild(buildEngineTab());      break;
    case 'sessions':    body.appendChild(buildSessionsTab());    break;
    case 'wallpapers':  body.appendChild(buildWallpapersTab());  break;
    case 'bookmarklet': body.appendChild(buildBookmarkletTab()); break;
  }
}

function buildProfileTab() {
  const frag = document.createDocumentFragment();
  const u = S.user;

  const sec = el('div', 'settings-section');
  sec.appendChild(el('div', 'settings-section-title', 'Profil'));

  const uRow = el('div', 'form-group mt-1');
  const uLbl = el('label', 'form-label', 'Nom d\'utilisateur');
  const uIn  = el('input', 'form-input');
  uIn.type   = 'text'; uIn.value = u.username; uIn.id = 'prof-username';
  uRow.append(uLbl, uIn);

  const eRow = el('div', 'form-group mt-1');
  const eLbl = el('label', 'form-label', 'Email');
  const eIn  = el('input', 'form-input');
  eIn.type   = 'email'; eIn.value = u.email; eIn.id = 'prof-email';
  eRow.append(eLbl, eIn);

  const err  = el('div', 'error-msg'); err.id = 'prof-error';
  const btn  = el('button', 'btn btn-primary mt-1', 'Enregistrer');
  btn.onclick = async () => {
    setError('prof-error', '');
    try {
      await PUT('/me', {
        username: $('prof-username').value.trim(),
        email:    $('prof-email').value.trim(),
      });
      S.user = await GET('/me');
    } catch (e) { setError('prof-error', e.message); }
  };

  sec.append(uRow, eRow, err, btn);
  frag.appendChild(sec);
  return frag;
}

function buildSecurityTab() {
  const frag = document.createDocumentFragment();

  // Change password
  const pwSec = el('div', 'settings-section');
  pwSec.appendChild(el('div', 'settings-section-title', 'Changer le mot de passe'));

  const curRow = el('div', 'form-group mt-1');
  const curLbl = el('label', 'form-label', 'Mot de passe actuel');
  const curIn  = el('input', 'form-input');
  curIn.type = 'password'; curIn.id = 'pw-current';
  curRow.append(curLbl, curIn);

  const newRow = el('div', 'form-group mt-1');
  const newLbl = el('label', 'form-label', 'Nouveau mot de passe');
  const newIn  = el('input', 'form-input');
  newIn.type = 'password'; newIn.id = 'pw-new';
  newRow.append(newLbl, newIn);

  const pwErr = el('div', 'error-msg'); pwErr.id = 'pw-error';
  const pwBtn = el('button', 'btn btn-primary mt-1', 'Modifier');
  pwBtn.onclick = async () => {
    setError('pw-error', '');
    try {
      await PUT('/me/password', {
        current_password: $('pw-current').value,
        new_password:     $('pw-new').value,
      });
      $('pw-current').value = '';
      $('pw-new').value     = '';
      setError('pw-error', '✓ Mot de passe modifié');
    } catch (e) { setError('pw-error', e.message); }
  };

  pwSec.append(curRow, newRow, pwErr, pwBtn);

  // TOTP
  const totpSec = el('div', 'settings-section');
  totpSec.appendChild(el('div', 'settings-section-title', 'Authentification à deux facteurs (TOTP)'));
  totpSec.appendChild(buildTOTPSection());

  frag.append(pwSec, totpSec);
  return frag;
}

function buildTOTPSection() {
  const wrap = el('div');

  const note = el('p', 'text-sm text-dim mb-1',
    'Le TOTP ajoute un code à 6 chiffres lors de la connexion (Google Authenticator, etc.)');
  wrap.appendChild(note);

  // "Configure" flow: POST /me/totp → show QR → PUT /me/totp with code
  const configBtn = el('button', 'btn btn-secondary', 'Configurer / Réinitialiser le TOTP');
  configBtn.onclick = async () => {
    configBtn.disabled = true;
    try {
      const res = await POST('/me/totp');
      wrap.innerHTML = '';

      const qrWrap = el('div', 'totp-qr-wrap');
      const info   = el('p', 'text-sm text-dim', 'Scannez ce lien avec votre app authenticator :');
      const qrLink = el('div', 'totp-qr-link', res.qr_url);
      const code   = el('input', 'form-input');
      code.placeholder = 'Code à 6 chiffres pour confirmer';
      code.inputMode   = 'numeric';
      code.maxLength   = 6;
      const errEl   = el('div', 'error-msg');
      const confBtn = el('button', 'btn btn-primary mt-1', 'Confirmer');
      confBtn.onclick = async () => {
        try {
          await PUT('/me/totp', { code: code.value.trim() });
          wrap.innerHTML = '';
          wrap.appendChild(el('p', 'text-sm text-dim', 'TOTP activé ✓'));
          wrap.appendChild(buildDisableTOTPBtn());
        } catch (e) { errEl.textContent = e.message; }
      };
      qrWrap.append(info, qrLink, code, errEl, confBtn);
      wrap.append(qrWrap, buildDisableTOTPBtn());
    } catch (e) {
      configBtn.disabled = false;
      alert(e.message);
    }
  };

  wrap.append(configBtn, buildDisableTOTPBtn());
  return wrap;
}

function buildDisableTOTPBtn() {
  const disBtn = el('button', 'btn btn-danger mt-1', 'Désactiver le TOTP');
  disBtn.onclick = async () => {
    const pw = prompt('Confirmez avec votre mot de passe actuel :');
    if (!pw) return;
    try {
      await DEL('/me/totp', { password: pw });
      renderSettingsTab('security');
    } catch (e) { alert(e.message); }
  };
  return disBtn;
}

function buildEngineTab() {
  const frag = document.createDocumentFragment();
  const sec  = el('div', 'settings-section');
  sec.appendChild(el('div', 'settings-section-title', 'Moteur de recherche'));

  const engines = [
    { id: 'duckduckgo', label: 'DuckDuckGo' },
    { id: 'google',     label: 'Google' },
    { id: 'brave',      label: 'Brave' },
    { id: 'bing',       label: 'Bing' },
    { id: 'kagi',       label: 'Kagi' },
    { id: 'custom',     label: 'Personnalisé' },
  ];

  const grid = el('div', 'engine-grid');
  for (const eng of engines) {
    const btn = el('button', 'engine-btn', eng.label);
    btn.dataset.engine = eng.id;
    if (S.user.search_engine === eng.id) btn.classList.add('active');
    btn.onclick = async () => {
      let customURL = undefined;
      if (eng.id === 'custom') {
        customURL = prompt('URL du moteur (doit finir par =) :', S.user.search_engine_url || '');
        if (!customURL) return;
      }
      try {
        await PUT('/me/search-engine', { engine: eng.id, custom_url: customURL || undefined });
        S.user = await GET('/me');
        grid.querySelectorAll('.engine-btn').forEach(b => {
          b.classList.toggle('active', b.dataset.engine === eng.id);
        });
      } catch (e) { alert(e.message); }
    };
    grid.appendChild(btn);
  }

  sec.appendChild(grid);
  frag.appendChild(sec);
  return frag;
}

function buildSessionsTab() {
  const frag = document.createDocumentFragment();
  const sec  = el('div', 'settings-section');
  sec.appendChild(el('div', 'settings-section-title', 'Sessions actives'));

  const list = el('div'); list.id = 'sessions-list';
  list.textContent = 'Chargement…';

  GET('/me/sessions').then(data => {
    S.sessions = Array.isArray(data) ? data : (data.sessions || []);
    list.innerHTML = '';
    if (!S.sessions.length) {
      list.textContent = 'Aucune session.';
      return;
    }
    for (const sess of S.sessions) {
      const row  = el('div', 'session-item');
      const info = el('div', 'session-info');
      const agent = el('div', 'session-agent', sess.user_agent || 'Agent inconnu');
      const meta  = el('div', 'session-meta',
        `${sess.ip || '—'} · Exp. ${new Date(sess.expires_at * 1000).toLocaleDateString('fr-FR')}`);

      if (sess.current)        agent.innerHTML += '<span class="badge badge-current">actuelle</span>';
      if (sess.is_bookmarklet) agent.innerHTML += '<span class="badge badge-bookmarklet">bookmarklet</span>';

      info.append(agent, meta);

      const revokeBtn = el('button', 'btn btn-small btn-danger', 'Révoquer');
      if (sess.current) {
        revokeBtn.disabled = true;
        revokeBtn.title    = 'Session courante';
      } else {
        revokeBtn.onclick = async () => {
          try {
            await DEL(`/me/sessions/${sess.id}`);
            renderSettingsTab('sessions');
          } catch (e) { alert(e.message); }
        };
      }

      row.append(info, revokeBtn);
      list.appendChild(row);
    }

    const revokeAll = el('button', 'btn btn-danger mt-2', 'Révoquer toutes les sessions');
    revokeAll.onclick = async () => {
      if (!confirm('Révoquer toutes les sessions (vous serez déconnecté) ?')) return;
      try { await DEL('/me/sessions'); } catch {}
      await logout();
    };
    list.appendChild(revokeAll);
  }).catch(e => { list.textContent = 'Erreur : ' + e.message; });

  sec.appendChild(list);
  frag.appendChild(sec);
  return frag;
}

function buildWallpapersTab() {
  const frag = document.createDocumentFragment();
  const sec  = el('div', 'settings-section');
  sec.appendChild(el('div', 'settings-section-title', 'Fonds d\'écran'));

  // Upload area
  const uploadLabel = el('label', 'upload-area');
  uploadLabel.textContent = '+ Cliquez ou glissez pour uploader un fond d\'écran';
  const fileIn = el('input', 'sr-only');
  fileIn.type   = 'file';
  fileIn.accept = '.jpg,.jpeg,.png,.webp,.avif,.mp4,.webm';
  fileIn.id     = 'wp-file-input';
  fileIn.setAttribute('for', 'wp-file-input');
  fileIn.onchange = async () => {
    const f = fileIn.files[0];
    if (!f) return;
    const form = new FormData();
    form.append('file', f);
    try {
      await fetch('/api/wallpapers', {
        method:      'POST',
        credentials: 'same-origin',
        body:        form,
      });
      S.wallpapers = (await GET('/wallpapers')) || [];
      renderWallpaperGrid(grid);
    } catch (e) { alert(e.message); }
  };
  uploadLabel.appendChild(fileIn);
  sec.appendChild(uploadLabel);

  const grid = el('div', 'wallpaper-grid'); grid.id = 'wp-grid';
  renderWallpaperGrid(grid);
  sec.appendChild(grid);

  frag.appendChild(sec);
  return frag;
}

function renderWallpaperGrid(grid) {
  grid.innerHTML = '';
  if (!S.wallpapers.length) {
    grid.appendChild(el('p', 'text-sm text-dimmer', 'Aucun fond d\'écran.'));
    return;
  }
  for (const wp of S.wallpapers) {
    const url  = `/media/${S.user.id}/${wp.filename}`;
    const ext  = wp.filename.split('.').pop().toLowerCase();
    const isVid = ['mp4','webm'].includes(ext);

    const thumb    = el('div', 'wp-thumb' + (wp.is_pinned ? ' pinned' : ''));
    const media    = el(isVid ? 'video' : 'img');
    media.src = url;
    if (isVid) { media.muted = true; media.autoplay = true; media.loop = true; }

    const overlay  = el('div', 'wp-thumb-overlay');

    const pinBtn = el('button', 'btn btn-small btn-secondary', wp.is_pinned ? '★' : '☆');
    pinBtn.title = wp.is_pinned ? 'Désépingler' : 'Épingler';
    pinBtn.onclick = async () => {
      try {
        await PUT(`/wallpapers/${wp.id}/pin`, { pinned: !wp.is_pinned });
        S.wallpapers = await GET('/wallpapers');
        const g = $('wp-grid');
        if (g) renderWallpaperGrid(g);
      } catch (e) { alert(e.message); }
    };

    const delBtn = el('button', 'btn btn-small btn-danger', '✕');
    delBtn.title  = 'Supprimer';
    delBtn.onclick = async () => {
      if (!confirm('Supprimer ce fond d\'écran ?')) return;
      try {
        await DEL(`/wallpapers/${wp.id}`);
        S.wallpapers = await GET('/wallpapers');
        const g = $('wp-grid');
        if (g) renderWallpaperGrid(g);
      } catch (e) { alert(e.message); }
    };

    overlay.append(pinBtn, delBtn);
    thumb.append(media, overlay);
    grid.appendChild(thumb);
  }
}

function buildBookmarkletTab() {
  const frag = document.createDocumentFragment();
  const sec  = el('div', 'settings-section');
  sec.appendChild(el('div', 'settings-section-title', 'Bookmarklet'));

  const info = el('p', 'text-sm text-dim mb-1',
    'Glissez le lien ci-dessous dans votre barre de favoris pour sauvegarder des pages en un clic.');
  sec.appendChild(info);

  const code = el('div', 'bookmarklet-code'); code.id = 'bml-code';
  code.textContent = 'Chargement…';

  const copyBtn  = el('button', 'btn btn-secondary mt-1', 'Copier le lien');
  const revokeBtn = el('button', 'btn btn-danger mt-1',   'Révoquer');

  copyBtn.onclick = () => {
    navigator.clipboard.writeText(code.textContent).then(() => {
      copyBtn.textContent = 'Copié ✓';
      setTimeout(() => { copyBtn.textContent = 'Copier le lien'; }, 2000);
    });
  };

  revokeBtn.onclick = async () => {
    if (!confirm('Révoquer le bookmarklet ? Le lien actuel ne fonctionnera plus.')) return;
    try {
      await DEL('/me/bookmarklet');
      code.textContent = 'Révoqué. Générez-en un nouveau.';
    } catch (e) { alert(e.message); }
  };

  // GET /me/bookmarklet generates a new session each time — only call on demand
  code.textContent = 'Cliquez sur « Générer » pour créer un bookmarklet.';

  const genBtn = el('button', 'btn btn-secondary mt-1', 'Générer un bookmarklet');
  genBtn.onclick = async () => {
    try {
      const res = await GET('/me/bookmarklet');
      code.textContent = res.bookmarklet || '—';
    } catch (e) { alert(e.message); }
  };

  sec.append(code, copyBtn, revokeBtn, genBtn);
  frag.appendChild(sec);
  return frag;
}

/* ─── Admin panel ────────────────────────────────────────────────────────── */
function openAdmin() {
  renderAdminTab('stats');
  show('overlay-admin');
}

function closeAdmin() {
  hide('overlay-admin');
}

function renderAdminTab(tabName) {
  document.querySelectorAll('#admin-tabs .tab').forEach(t => {
    t.classList.toggle('active', t.dataset.tab === tabName);
  });

  const body = $('admin-body');
  body.innerHTML = '';

  switch (tabName) {
    case 'stats': body.appendChild(buildAdminStats());   break;
    case 'users': body.appendChild(buildAdminUsers());   break;
    case 'audit': body.appendChild(buildAdminAudit());   break;
  }
}

function buildAdminStats() {
  const frag = document.createDocumentFragment();
  const grid = el('div', 'stat-grid');
  grid.textContent = 'Chargement…';

  GET('/admin/stats').then(s => {
    grid.innerHTML = '';
    const stats = [
      { v: s.total_users,      l: 'Utilisateurs' },
      { v: s.active_users,     l: 'Actifs' },
      { v: s.total_bookmarks,  l: 'Marque-pages' },
      { v: s.total_wallpapers, l: 'Fonds d\'écran' },
      { v: fmtBytes(s.db_size_bytes), l: 'Base de données' },
    ];
    for (const { v, l } of stats) {
      const card = el('div', 'stat-card');
      card.appendChild(el('div', 'stat-value', String(v)));
      card.appendChild(el('div', 'stat-label', l));
      grid.appendChild(card);
    }
  }).catch(e => { grid.textContent = 'Erreur : ' + e.message; });

  frag.appendChild(grid);
  return frag;
}

function fmtBytes(b) {
  if (!b) return '0 B';
  const u = ['B','KB','MB','GB'];
  let i = 0; let n = b;
  while (n >= 1024 && i < u.length - 1) { n /= 1024; i++; }
  return `${n.toFixed(i ? 1 : 0)} ${u[i]}`;
}

function buildAdminUsers() {
  const frag = document.createDocumentFragment();

  const addBtn = el('button', 'btn btn-secondary mb-1', '+ Créer un utilisateur');
  addBtn.onclick = () => {
    $('nu-username').value = '';
    $('nu-email').value    = '';
    $('nu-password').value = '';
    setError('nu-error', '');
    show('modal-new-user');
  };

  const list = el('div'); list.id = 'admin-users-list';
  list.textContent = 'Chargement…';

  GET('/admin/users').then(data => {
    const users = data.users || data.items || data || [];
    list.innerHTML = '';
    for (const u of users) {
      const row = el('div', 'admin-user-row');

      const nameEl = el('div');
      const name   = el('span', 'user-name', u.username);
      const email  = el('span', 'user-email', ' · ' + u.email);
      if (u.role === 'admin') name.innerHTML += '<span class="badge badge-admin">admin</span>';
      if (!u.is_active)       name.innerHTML += '<span class="badge badge-inactive">suspendu</span>';
      nameEl.append(name, email);

      const acts = el('div', 'flex gap-1');

      if (u.id !== S.user.id) {
        const suspBtn = el('button', 'btn btn-small btn-secondary',
          u.is_active ? 'Suspendre' : 'Activer');
        suspBtn.onclick = async () => {
          try {
            if (u.is_active) await PUT(`/admin/users/${u.id}/suspend`);
            else             await PUT(`/admin/users/${u.id}/activate`);
            renderAdminTab('users');
          } catch (e) { alert(e.message); }
        };

        const delBtn = el('button', 'btn btn-small btn-danger', 'Supprimer');
        delBtn.onclick = async () => {
          if (!confirm(`Supprimer l'utilisateur ${u.username} ? Cette action est irréversible.`)) return;
          try {
            await DEL(`/admin/users/${u.id}`);
            renderAdminTab('users');
          } catch (e) { alert(e.message); }
        };

        acts.append(suspBtn, delBtn);
      } else {
        acts.appendChild(el('span', 'text-sm text-dimmer', '(vous)'));
      }

      row.append(nameEl, acts);
      list.appendChild(row);
    }
  }).catch(e => { list.textContent = 'Erreur : ' + e.message; });

  frag.append(addBtn, list);
  return frag;
}

function buildAdminAudit() {
  const frag  = document.createDocumentFragment();
  const list  = el('div'); list.id = 'admin-audit-list';
  list.textContent = 'Chargement…';

  GET('/admin/audit?limit=100').then(data => {
    const entries = data.entries || data.items || data || [];
    list.innerHTML = '';
    if (!entries.length) { list.textContent = 'Aucune entrée.'; return; }
    for (const e of entries) {
      const row    = el('div', 'audit-row');
      const action = el('span', 'audit-action', e.action);
      const ip     = el('span', 'audit-ip', e.ip || '—');
      const user   = el('span', 'audit-user', e.user_id ? e.user_id.slice(0,8) + '…' : 'système');
      const time   = el('span', 'audit-time',
        new Date(e.created_at * 1000).toLocaleString('fr-FR', { dateStyle: 'short', timeStyle: 'short' }));
      row.append(action, ip, user, time);
      list.appendChild(row);
    }
  }).catch(e => { list.textContent = 'Erreur : ' + e.message; });

  frag.appendChild(list);
  return frag;
}

/* ─── New user (admin) ───────────────────────────────────────────────────── */
async function createUser() {
  setError('nu-error', '');
  const username = $('nu-username').value.trim();
  const email    = $('nu-email').value.trim();
  const password = $('nu-password').value;

  try {
    await POST('/auth/register', { username, email, password });
    hide('modal-new-user');
    renderAdminTab('users');
  } catch (e) { setError('nu-error', e.message); }
}

/* ─── Boot ───────────────────────────────────────────────────────────────── */
async function boot() {
  try {
    S.user = await GET('/me');
  } catch {
    hide('view-home');
    show('view-login');
    $('login-email').focus();
    return;
  }

  // Load wallpapers
  try {
    S.wallpapers = (await GET('/wallpapers')) || [];
  } catch { S.wallpapers = []; }

  loadWallpaper();
  startClock();

  // Show admin button if admin
  toggle('btn-admin', S.user.role === 'admin');

  hide('view-login');
  show('view-home');
  $('search-input').focus();
}

/* ─── Event wiring ───────────────────────────────────────────────────────── */
document.addEventListener('DOMContentLoaded', () => {
  initRain();

  // Boot
  boot();

  // Login
  $('login-form').addEventListener('submit', handleLogin);
  $('forgot-form').addEventListener('submit', handleForgot);
  $('forgot-link').addEventListener('click', showForgotForm);
  $('forgot-back').addEventListener('click', showLoginForm);

  // Home nav
  $('search-form').addEventListener('submit', handleSearch);
  $('btn-logout').addEventListener('click', logout);
  $('btn-bookmarks').addEventListener('click', openBookmarks);
  $('btn-settings').addEventListener('click', openSettings);
  $('btn-admin').addEventListener('click', openAdmin);

  // Bookmark overlay
  $('bm-close').addEventListener('click', closeBookmarks);
  $('bm-add-btn').addEventListener('click', openAddBookmark);
  $('bm-export-btn').addEventListener('click', exportBookmarks);
  $('bm-import-file').addEventListener('change', e => importBookmarks(e.target.files[0]));
  $('bm-search-input').addEventListener('input', () => {
    S.bmFilter = $('bm-search-input').value.trim();
    S.bmOffset = 0;
    loadBookmarks();
  });
  $('bm-folder-filter').addEventListener('change', () => {
    S.bmFolder = $('bm-folder-filter').value;
    S.bmOffset = 0;
    loadBookmarks();
  });

  // Bookmark modal
  $('modal-bm-save').addEventListener('click', saveBookmark);
  $('modal-bm-cancel').addEventListener('click', () => hide('modal-bookmark'));

  // Settings overlay
  $('settings-close').addEventListener('click', closeSettings);
  document.querySelectorAll('#settings-tabs .tab').forEach(tab => {
    tab.addEventListener('click', () => renderSettingsTab(tab.dataset.tab));
  });

  // Admin overlay
  $('admin-close').addEventListener('click', closeAdmin);
  document.querySelectorAll('#admin-tabs .tab').forEach(tab => {
    tab.addEventListener('click', () => renderAdminTab(tab.dataset.tab));
  });

  // New user modal
  $('nu-save').addEventListener('click', createUser);
  $('nu-cancel').addEventListener('click', () => hide('modal-new-user'));

  // Close overlays on backdrop click
  ['overlay-bookmarks','overlay-settings','overlay-admin','modal-bookmark','modal-new-user'].forEach(id => {
    $(id).addEventListener('click', e => {
      if (e.target === $(id)) hide(id);
    });
  });

  // Keyboard: Escape closes top-most overlay
  document.addEventListener('keydown', e => {
    if (e.key !== 'Escape') return;
    const overlays = ['modal-bookmark','modal-new-user','overlay-bookmarks','overlay-settings','overlay-admin'];
    for (const id of overlays) {
      const el = $(id);
      if (!el.classList.contains('hidden')) {
        hide(id);
        break;
      }
    }
  });
});
