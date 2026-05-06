package api

const managePageHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Proxy Lite TLS</title>
  <style>
    :root {
      --panel: rgba(255,255,255,.86);
      --ink: #162122;
      --muted: #5c6a6d;
      --accent: #0d7a66;
      --accent-strong: #0a5a4c;
      --line: rgba(22,33,34,.12);
      --warn: #b04942;
      --shadow: 0 18px 48px rgba(9,26,25,.12);
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      min-height: 100vh;
      color: var(--ink);
      font-family: "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif;
      background:
        radial-gradient(circle at top left, rgba(13,122,102,.18), transparent 28%),
        radial-gradient(circle at bottom right, rgba(176,73,66,.1), transparent 22%),
        linear-gradient(160deg, #efe7d8 0%, #f8f5ee 48%, #e7f0eb 100%);
    }
    .shell {
      width: min(1260px, calc(100vw - 32px));
      margin: 24px auto 40px;
      display: grid;
      gap: 18px;
    }
    .hero, .panel {
      background: var(--panel);
      border: 1px solid var(--line);
      border-radius: 24px;
      box-shadow: var(--shadow);
      backdrop-filter: blur(14px);
    }
    .hero {
      padding: 28px;
      display: grid;
      gap: 16px;
    }
    .hero h1 {
      margin: 0;
      font-size: clamp(30px, 5vw, 48px);
      line-height: 1;
      letter-spacing: -.04em;
    }
    .hero p {
      margin: 0;
      max-width: 780px;
      color: var(--muted);
      line-height: 1.6;
    }
    .hero-top {
      display: flex;
      gap: 14px;
      align-items: start;
      justify-content: space-between;
      flex-wrap: wrap;
    }
    .pill {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      padding: 7px 11px;
      border-radius: 999px;
      background: rgba(13,122,102,.12);
      color: var(--accent-strong);
      font-size: 12px;
    }
    .pill.warn {
      background: rgba(176,73,66,.12);
      color: var(--warn);
    }
    .stats {
      display: grid;
      grid-template-columns: repeat(3, minmax(0, 1fr));
      gap: 12px;
    }
    .stat {
      padding: 14px 16px;
      border-radius: 18px;
      background: rgba(255,255,255,.7);
      border: 1px solid var(--line);
    }
    .stat strong {
      display: block;
      font-size: 22px;
      margin-bottom: 4px;
    }
    .grid {
      display: grid;
      grid-template-columns: 340px minmax(0, 1fr);
      gap: 18px;
    }
    .stack {
      display: grid;
      gap: 18px;
      align-content: start;
    }
    .panel {
      padding: 22px;
    }
    .panel h2 {
      margin: 0 0 14px;
      font-size: 18px;
    }
    form {
      display: grid;
      gap: 12px;
    }
    label {
      display: grid;
      gap: 6px;
      color: var(--muted);
      font-size: 13px;
    }
    input, select, button {
      font: inherit;
    }
    input, select {
      width: 100%;
      padding: 12px 14px;
      border-radius: 14px;
      border: 1px solid var(--line);
      background: rgba(255,255,255,.92);
      color: var(--ink);
    }
    button {
      border: 0;
      border-radius: 999px;
      padding: 11px 16px;
      background: var(--accent);
      color: #fff;
      cursor: pointer;
      transition: transform .15s ease, background .15s ease;
    }
    button:hover {
      transform: translateY(-1px);
      background: var(--accent-strong);
    }
    button.secondary {
      background: rgba(22,33,34,.08);
      color: var(--ink);
    }
    button.warn {
      background: var(--warn);
    }
    .toolbar {
      display: flex;
      gap: 12px;
      align-items: center;
      justify-content: space-between;
      flex-wrap: wrap;
      margin-bottom: 14px;
    }
    .toolbar-controls {
      display: flex;
      gap: 10px;
      flex-wrap: wrap;
      align-items: center;
      margin-bottom: 14px;
    }
    .toolbar-controls input {
      min-width: 240px;
    }
    .toolbar-controls select {
      min-width: 140px;
      width: auto;
    }
    .status {
      min-height: 22px;
      font-size: 13px;
      color: var(--muted);
    }
    .status.error {
      color: var(--warn);
    }
    .cards {
      display: grid;
      gap: 14px;
    }
    .card {
      border: 1px solid var(--line);
      border-radius: 18px;
      padding: 16px;
      background: rgba(255,255,255,.75);
      display: grid;
      gap: 10px;
    }
    .card-header {
      display: flex;
      gap: 12px;
      align-items: start;
      justify-content: space-between;
    }
    .meta {
      color: var(--muted);
      font-size: 13px;
      line-height: 1.5;
    }
    .actions {
      display: flex;
      gap: 8px;
      flex-wrap: wrap;
    }
    .mono {
      font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
      font-size: 12px;
      word-break: break-all;
    }
    .hint {
      padding: 12px 14px;
      border-radius: 16px;
      background: rgba(22,33,34,.05);
      color: var(--muted);
      line-height: 1.5;
      font-size: 13px;
    }
    .hidden {
      display: none !important;
    }
    dialog {
      width: min(92vw, 580px);
      border: 0;
      border-radius: 26px;
      padding: 0;
      box-shadow: var(--shadow);
    }
    dialog::backdrop {
      background: rgba(13, 24, 18, .45);
      backdrop-filter: blur(6px);
    }
    .dialog-body {
      padding: 24px;
      display: grid;
      gap: 16px;
    }
    .qr-wrap {
      display: grid;
      justify-items: center;
      gap: 12px;
    }
    .qr-wrap img {
      width: min(100%, 360px);
      border-radius: 18px;
      background: #fff;
      padding: 12px;
      border: 1px solid var(--line);
    }
    @media (max-width: 980px) {
      .grid, .stats { grid-template-columns: 1fr; }
      .toolbar-controls { width: 100%; }
      .toolbar-controls input, .toolbar-controls select { width: 100%; min-width: 0; }
    }
  </style>
</head>
<body>
  <div class="shell">
    <section class="hero">
      <div class="hero-top">
        <div>
          <h1>Proxy Lite TLS</h1>
          <p>Manage VLESS over TLS clients on the shared HTTPS edge. Create per-device identities, revoke them instantly, and hand out ready-to-import links or QR codes without touching the Xray runtime by hand.</p>
        </div>
        <div id="sessionPill" class="pill hidden"></div>
      </div>
      <div class="stats">
        <div class="stat">
          <strong id="activeCount">0</strong>
          <div class="meta">Active clients</div>
        </div>
        <div class="stat">
          <strong id="disabledCount">0</strong>
          <div class="meta">Disabled clients</div>
        </div>
        <div class="stat">
          <strong id="userCount">0</strong>
          <div class="meta">Users</div>
        </div>
      </div>
      <div id="endpointHint" class="hint hidden"></div>
      <div id="status" class="status"></div>
    </section>

    <section id="loginPanel" class="panel">
      <h2>Admin Login</h2>
      <form id="loginForm">
        <label>Username
          <input id="loginUsername" name="username" autocomplete="username" required placeholder="admin">
        </label>
        <label>Password
          <input id="loginPassword" name="password" type="password" autocomplete="current-password" required placeholder="Enter admin password">
        </label>
        <button type="submit">Sign In</button>
      </form>
    </section>

    <section id="dashboard" class="grid hidden">
      <div class="stack">
        <section class="panel">
          <h2>Create User</h2>
          <form id="userForm">
            <label>Name
              <input id="userName" name="name" required placeholder="alice">
            </label>
            <label>Email
              <input id="userEmail" name="email" required placeholder="alice@example.com">
            </label>
            <button type="submit">Create User</button>
          </form>
        </section>

        <section class="panel">
          <div class="toolbar" style="margin-bottom:12px;">
            <div>
              <h2 style="margin-bottom:4px;">Users</h2>
              <div class="meta">Delete is allowed only when the user has no linked clients.</div>
            </div>
          </div>
          <div id="userCards" class="cards"></div>
        </section>

        <section class="panel">
          <h2>Create Client</h2>
          <form id="clientForm">
            <label>User
              <select id="userSelect" name="user_id" required></select>
            </label>
            <label>Device Label
              <input id="clientLabel" name="label" required placeholder="alice-laptop">
            </label>
            <button type="submit">Issue Client</button>
          </form>
        </section>

        <section class="panel">
          <h2>Session</h2>
          <div class="hint">Only authenticated admins can manage users and clients. Logging out clears the browser session cookie.</div>
          <div class="actions" style="margin-top:12px;">
            <button id="refreshBtn" type="button" class="secondary">Refresh</button>
            <button id="logoutBtn" type="button" class="warn">Log Out</button>
          </div>
        </section>
      </div>

      <section class="panel">
        <div class="toolbar">
          <div>
            <h2 style="margin-bottom:4px;">Issued Clients</h2>
            <div class="meta">Search by label, user, email, client UUID, or state.</div>
          </div>
          <button id="refreshListBtn" class="secondary" type="button">Reload List</button>
        </div>
        <div class="toolbar-controls">
          <input id="searchInput" type="search" placeholder="Search client or user">
          <select id="stateFilter">
            <option value="all">All states</option>
            <option value="active">Active only</option>
            <option value="disabled">Disabled only</option>
          </select>
        </div>
        <div id="cards" class="cards"></div>
      </section>
    </section>
  </div>

  <dialog id="qrDialog">
    <div class="dialog-body">
      <div class="card-header">
        <div>
          <h2 id="qrTitle" style="margin:0;">Client QR</h2>
          <div id="qrSubtitle" class="meta"></div>
        </div>
        <button class="secondary" type="button" onclick="closeQrDialog()">Close</button>
      </div>
      <div class="qr-wrap">
        <img id="qrImage" alt="Client QR code">
        <div class="meta">Import this QR code with a VLESS-compatible client on mobile.</div>
      </div>
    </div>
  </dialog>

  <script>
    const state = {
      session: null,
      users: [],
      clients: [],
      search: '',
      filterState: 'all',
    };

    const statusEl = document.getElementById('status');
    const loginPanel = document.getElementById('loginPanel');
    const dashboard = document.getElementById('dashboard');
    const cardsEl = document.getElementById('cards');
    const userCardsEl = document.getElementById('userCards');
    const userSelect = document.getElementById('userSelect');
    const sessionPill = document.getElementById('sessionPill');
    const endpointHint = document.getElementById('endpointHint');
    const activeCount = document.getElementById('activeCount');
    const disabledCount = document.getElementById('disabledCount');
    const userCount = document.getElementById('userCount');
    const qrDialog = document.getElementById('qrDialog');
    const qrImage = document.getElementById('qrImage');
    const qrTitle = document.getElementById('qrTitle');
    const qrSubtitle = document.getElementById('qrSubtitle');
    const searchInput = document.getElementById('searchInput');
    const stateFilter = document.getElementById('stateFilter');

    function setStatus(message, isError = false) {
      statusEl.textContent = message;
      statusEl.className = 'status' + (isError ? ' error' : '');
    }

    function delay(ms) {
      return new Promise((resolve) => setTimeout(resolve, ms));
    }

    async function request(path, options = {}, retries = 0) {
      try {
        return await fetch(path, {
          credentials: 'same-origin',
          cache: 'no-store',
          ...options,
          headers: {
            ...(options.headers || {}),
          },
        });
      } catch (error) {
        if (retries > 0) {
          await delay(250);
          return request(path, options, retries - 1);
        }
        throw error;
      }
    }

    async function api(path, options = {}) {
      const method = (options.method || 'GET').toUpperCase();
      const retries = method === 'GET' ? 1 : 0;
      const response = await request(path, {
        ...options,
      }, retries);
      if (response.status === 401) {
        state.session = null;
        renderSession();
        throw new Error('Please log in again');
      }
      if (!response.ok) {
        let message = 'Request failed';
        try {
          const data = await response.json();
          message = data.error || message;
        } catch (_) {}
        throw new Error(message);
      }
      return response;
    }

    function closeQrDialog() {
      if (qrImage.dataset.url) {
        URL.revokeObjectURL(qrImage.dataset.url);
        qrImage.dataset.url = '';
      }
      qrDialog.close();
    }

    function renderSession() {
      const loggedIn = !!state.session;
      loginPanel.classList.toggle('hidden', loggedIn);
      dashboard.classList.toggle('hidden', !loggedIn);
      sessionPill.classList.toggle('hidden', !loggedIn);
      endpointHint.classList.toggle('hidden', !loggedIn);

      if (loggedIn) {
        sessionPill.textContent = 'Signed in as ' + state.session.username;
        endpointHint.textContent = 'Shared edge: ' + state.session.public_host + ':' + state.session.public_port + '  |  WS path: ' + state.session.ws_path;
      } else {
        sessionPill.textContent = '';
        endpointHint.textContent = '';
      }
    }

    function renderStats() {
      userCount.textContent = String(state.users.length);
      activeCount.textContent = String(state.clients.filter((item) => item.state === 'active').length);
      disabledCount.textContent = String(state.clients.filter((item) => item.state === 'disabled').length);
    }

    function renderUsers() {
      userSelect.innerHTML = '';
      if (state.users.length === 0) {
        const option = document.createElement('option');
        option.textContent = 'Create a user first';
        option.value = '';
        userSelect.appendChild(option);
        return;
      }
      for (const user of state.users) {
        const option = document.createElement('option');
        option.value = user.id;
        option.textContent = user.name + ' <' + user.email + '>';
        userSelect.appendChild(option);
      }
    }

    function clientsForUser(userId) {
      return state.clients.filter((item) => item.user_id === userId);
    }

    function userCard(user) {
      const wrapper = document.createElement('article');
      wrapper.className = 'card';
      const linkedClients = clientsForUser(user.id);
      wrapper.innerHTML = [
        '<div class="card-header">',
          '<div>',
            '<div style="font-weight:700;">' + escapeHtml(user.name) + '</div>',
            '<div class="meta">' + escapeHtml(user.email) + '</div>',
          '</div>',
          '<span class="pill">' + String(linkedClients.length) + ' client' + (linkedClients.length === 1 ? '' : 's') + '</span>',
        '</div>',
        '<div class="meta">user_id: <span class="mono">' + escapeHtml(user.id) + '</span></div>',
        '<div class="actions">',
          '<button type="button" class="secondary" data-action="delete">Delete User</button>',
        '</div>',
      ].join('');

      wrapper.querySelector('button[data-action="delete"]').addEventListener('click', () => handleUserDelete(user));
      return wrapper;
    }

    function renderUserDirectory() {
      userCardsEl.innerHTML = '';
      if (state.users.length === 0) {
        userCardsEl.innerHTML = '<div class="meta">No users yet.</div>';
        return;
      }
      for (const user of state.users) {
        userCardsEl.appendChild(userCard(user));
      }
    }

    function userById(id) {
      return state.users.find((user) => user.id === id) || null;
    }

    function filteredClients() {
      const search = state.search.trim().toLowerCase();
      return state.clients.filter((item) => {
        if (state.filterState !== 'all' && item.state !== state.filterState) {
          return false;
        }

        if (!search) {
          return true;
        }

        const user = userById(item.user_id);
        const haystack = [
          item.label,
          item.id,
          item.user_id,
          item.client_uuid,
          item.state,
          user ? user.name : '',
          user ? user.email : '',
        ].join(' ').toLowerCase();
        return haystack.includes(search);
      });
    }

    function escapeHtml(value) {
      return String(value)
        .replaceAll('&', '&amp;')
        .replaceAll('<', '&lt;')
        .replaceAll('>', '&gt;')
        .replaceAll('"', '&quot;')
        .replaceAll("'", '&#39;');
    }

    function clientCard(item) {
      const wrapper = document.createElement('article');
      wrapper.className = 'card';
      const disabled = item.state !== 'active';
      const user = userById(item.user_id);
      const userLabel = user ? user.name + ' <' + user.email + '>' : item.user_id;

      wrapper.innerHTML = [
        '<div class="card-header">',
          '<div>',
            '<div style="font-weight:700;">' + escapeHtml(item.label) + '</div>',
            '<div class="meta">user: <span class="mono">' + escapeHtml(userLabel) + '</span></div>',
          '</div>',
          '<span class="pill' + (disabled ? ' warn' : '') + '">' + escapeHtml(item.state) + '</span>',
        '</div>',
        '<div class="meta">client_id: <span class="mono">' + escapeHtml(item.id) + '</span></div>',
        '<div class="meta">uuid: <span class="mono">' + escapeHtml(item.client_uuid) + '</span></div>',
        '<div class="meta">created: ' + new Date(item.created_at).toLocaleString() + '</div>',
        '<div class="actions">',
          '<button type="button" data-action="copy">Copy Link</button>',
          '<button type="button" class="secondary" data-action="download">Download Link</button>',
          '<button type="button" class="secondary" data-action="qr">Show QR</button>',
          (disabled
            ? '<button type="button" class="secondary" data-action="enable">Enable</button>'
            : '<button type="button" class="warn" data-action="disable">Disable</button>'),
          '<button type="button" class="secondary" data-action="delete">Delete</button>',
        '</div>',
      ].join('');

      wrapper.querySelectorAll('button[data-action]').forEach((button) => {
        button.addEventListener('click', () => handleClientAction(item, button.dataset.action));
      });

      return wrapper;
    }

    function renderClients() {
      cardsEl.innerHTML = '';
      const items = filteredClients();
      if (items.length === 0) {
        cardsEl.innerHTML = '<div class="meta">No clients match the current filters.</div>';
        return;
      }
      for (const item of items) {
        cardsEl.appendChild(clientCard(item));
      }
    }

    async function copyText(value) {
      if (navigator.clipboard && navigator.clipboard.writeText) {
        await navigator.clipboard.writeText(value);
        return;
      }
      const textarea = document.createElement('textarea');
      textarea.value = value;
      textarea.style.position = 'fixed';
      textarea.style.opacity = '0';
      document.body.appendChild(textarea);
      textarea.select();
      document.execCommand('copy');
      textarea.remove();
    }

    async function loadSession() {
      try {
        const response = await request('/api/v1/auth/me', {}, 1);
        if (!response.ok) {
          state.session = null;
          renderSession();
          return false;
        }
        const session = await response.json();
        if (!session.authenticated) {
          state.session = null;
          renderSession();
          return false;
        }
        state.session = session;
        renderSession();
        return true;
      } catch (_) {
        state.session = null;
        renderSession();
        return false;
      }
    }

    async function loadData() {
      if (!state.session) {
        return;
      }
      setStatus('Loading users and clients...');
      const [usersResult, clientsResult] = await Promise.allSettled([
        api('/api/v1/users'),
        api('/api/v1/clients'),
      ]);

      let loadError = null;

      if (usersResult.status === 'fulfilled') {
        state.users = (await usersResult.value.json()).items || [];
      } else {
        state.users = [];
        loadError = usersResult.reason;
      }

      if (clientsResult.status === 'fulfilled') {
        state.clients = (await clientsResult.value.json()).items || [];
      } else {
        state.clients = [];
        loadError = clientsResult.reason;
      }

      renderUsers();
      renderUserDirectory();
      renderStats();
      renderClients();
      if (loadError) {
        throw loadError;
      }
      setStatus('Ready');
    }

    async function handleClientAction(item, action) {
      try {
        if (action === 'copy') {
          const res = await api('/api/v1/clients/' + item.id + '/share-link');
          const text = await res.text();
          await copyText(text);
          setStatus('Share link copied for ' + item.label);
          return;
        }

        if (action === 'download') {
          const res = await api('/api/v1/clients/' + item.id + '/share-link?download=1');
          const blob = await res.blob();
          const url = URL.createObjectURL(blob);
          const link = document.createElement('a');
          link.href = url;
          link.download = item.label + '-' + item.id + '.txt';
          link.click();
          URL.revokeObjectURL(url);
          setStatus('Share link downloaded for ' + item.label);
          return;
        }

        if (action === 'qr') {
          const res = await api('/api/v1/clients/' + item.id + '/share-link-qr.png');
          const blob = await res.blob();
          if (qrImage.dataset.url) {
            URL.revokeObjectURL(qrImage.dataset.url);
          }
          const url = URL.createObjectURL(blob);
          qrImage.dataset.url = url;
          qrImage.src = url;
          qrTitle.textContent = item.label;
          qrSubtitle.textContent = 'UUID: ' + item.client_uuid;
          qrDialog.showModal();
          setStatus('QR code ready for ' + item.label);
          return;
        }

        if (action === 'disable') {
          if (!confirm('Disable ' + item.label + '?')) {
            return;
          }
          await api('/api/v1/clients/' + item.id + '/disable', { method: 'POST' });
          await loadData();
          setStatus('Disabled ' + item.label);
          return;
        }

        if (action === 'enable') {
          await api('/api/v1/clients/' + item.id + '/enable', { method: 'POST' });
          await loadData();
          setStatus('Enabled ' + item.label);
          return;
        }

        if (action === 'delete') {
          if (!confirm('Delete ' + item.label + '?')) {
            return;
          }
          await api('/api/v1/clients/' + item.id, { method: 'DELETE' });
          await loadData();
          setStatus('Deleted ' + item.label);
        }
      } catch (error) {
        setStatus(error.message, true);
      }
    }

    async function handleUserDelete(user) {
      try {
        const linkedClients = clientsForUser(user.id);
        if (linkedClients.length > 0) {
          throw new Error('Delete this user\'s clients first');
        }
        if (!confirm('Delete user ' + user.name + '?')) {
          return;
        }
        await api('/api/v1/users/' + user.id, { method: 'DELETE' });
        await loadData();
        setStatus('Deleted user ' + user.name);
      } catch (error) {
        setStatus(error.message, true);
      }
    }

    document.getElementById('loginForm').addEventListener('submit', async (event) => {
      event.preventDefault();
      try {
        const payload = {
          username: document.getElementById('loginUsername').value.trim(),
          password: document.getElementById('loginPassword').value,
        };
        await api('/api/v1/auth/login', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload),
        });
        document.getElementById('loginPassword').value = '';
        await loadSession();
        await loadData();
        setStatus('Signed in');
      } catch (error) {
        setStatus(error.message, true);
      }
    });

    document.getElementById('logoutBtn').addEventListener('click', async () => {
      try {
        await api('/api/v1/auth/logout', { method: 'POST' });
      } catch (_) {}
      state.session = null;
      state.users = [];
      state.clients = [];
      renderSession();
      renderUsers();
      renderUserDirectory();
      renderStats();
      renderClients();
      setStatus('Signed out');
    });

    document.getElementById('refreshBtn').addEventListener('click', async () => {
      try {
        await loadSession();
        await loadData();
      } catch (error) {
        setStatus(error.message, true);
      }
    });

    document.getElementById('refreshListBtn').addEventListener('click', async () => {
      try {
        await loadData();
      } catch (error) {
        setStatus(error.message, true);
      }
    });

    searchInput.addEventListener('input', (event) => {
      state.search = event.target.value;
      renderClients();
    });

    stateFilter.addEventListener('change', (event) => {
      state.filterState = event.target.value;
      renderClients();
    });

    document.getElementById('userForm').addEventListener('submit', async (event) => {
      event.preventDefault();
      try {
        const payload = {
          name: document.getElementById('userName').value.trim(),
          email: document.getElementById('userEmail').value.trim(),
        };
        await api('/api/v1/users', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload),
        });
        event.target.reset();
        await loadData();
        setStatus('User created');
      } catch (error) {
        setStatus(error.message, true);
      }
    });

    document.getElementById('clientForm').addEventListener('submit', async (event) => {
      event.preventDefault();
      try {
        const payload = {
          user_id: document.getElementById('userSelect').value,
          label: document.getElementById('clientLabel').value.trim(),
        };
        await api('/api/v1/clients', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload),
        });
        event.target.reset();
        await loadData();
        setStatus('Client issued');
      } catch (error) {
        setStatus(error.message, true);
      }
    });

    renderSession();
    renderUsers();
    renderUserDirectory();
    renderStats();
    renderClients();
    loadSession()
      .then((loggedIn) => loggedIn ? loadData() : setStatus('Sign in to manage clients.'))
      .catch((error) => setStatus(error.message, true));
  </script>
</body>
</html>
`
