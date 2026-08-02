// Atelier : frontend SPA vanilla (zéro framework, zéro dépendance).
// Contrat d'API : voir SPEC-API.md à la racine du dépôt.
(function () {
  'use strict';

  // ---------------------------------------------------------------------
  // Constantes
  // ---------------------------------------------------------------------

  var STATUS_GLYPH = {
    running: { icon: '◐', color: '#8b8982' },
    review: { icon: '◍', color: '#9a6b0d' },
    ready: { icon: '⬆', color: '#2f5fb0' },
    shipped: { icon: '✓', color: '#2f7d54' }
  };
  var COLUMN_LABELS = { soon: 'Bientôt', doing: 'En cours', done: 'Terminé' };
  var COLUMN_ORDER = ['soon', 'doing', 'done'];

  // ---------------------------------------------------------------------
  // État en mémoire
  // ---------------------------------------------------------------------

  var state = {
    projects: [], cards: [], tasks: [], agents: [],
    projectsById: {}, cardsById: {}, tasksById: {}, agentsById: {},
    tokens: { global: { input: 0, output: 0, costUsd: 0 } },
    messagesByTask: {}, diffByTask: {}, deliverablesByTask: {},
    activeDiffFile: {},
    detailErrorByTask: {},
    loading: {},
    screen: 'inbox', // 'inbox' | 'projects' | 'kanban' | 'work'
    projectId: null, cardId: null, taskId: null,
    panelTab: 'chat', taskFilter: 'all',
    searchOpen: false, modal: null,
    pendingShip: null,
    sidebarOpen: false // tiroir mobile (< 860px)
  };

  var modalAgentId = null;
  var modalColumn = 'soon';
  var sseOpenedOnce = false;

  // ---------------------------------------------------------------------
  // Utilitaires
  // ---------------------------------------------------------------------

  function escapeHtml(str) {
    return String(str === null || str === undefined ? '' : str)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#39;');
  }

  function isMac() {
    return /Mac|iPod|iPhone|iPad/.test(navigator.platform || '');
  }

  function softColor(hex, alpha) {
    if (!hex) return '#eeece6';
    var h = String(hex).replace('#', '');
    if (h.length !== 6) return hex;
    var r = parseInt(h.slice(0, 2), 16), g = parseInt(h.slice(2, 4), 16), b = parseInt(h.slice(4, 6), 16);
    if (isNaN(r) || isNaN(g) || isNaN(b)) return '#eeece6';
    return 'rgba(' + r + ',' + g + ',' + b + ',' + (alpha === undefined ? 0.22 : alpha) + ')';
  }

  function formatTokens(n) {
    n = n || 0;
    if (n >= 1000) return (n / 1000).toFixed(1).replace('.', ',') + 'k';
    return String(n);
  }
  function formatCost(n) {
    return (n || 0).toFixed(2).replace('.', ',');
  }
  function tokenSummary(tokens) {
    tokens = tokens || {};
    var total = (tokens.input || 0) + (tokens.output || 0);
    return 'Σ ' + formatTokens(total) + ' tokens · ' + formatCost(tokens.costUsd) + ' $';
  }

  function timeAgo(iso) {
    if (!iso) return '';
    var then = new Date(iso).getTime();
    if (isNaN(then)) return '';
    var diff = Math.max(0, Math.round((Date.now() - then) / 1000));
    if (diff < 45) return 'à l\'instant';
    var min = Math.round(diff / 60);
    if (min < 60) return 'il y a ' + min + ' min';
    var hr = Math.round(min / 60);
    if (hr < 24) return 'il y a ' + hr + ' h';
    var day = Math.round(hr / 24);
    if (day === 1) return 'hier';
    if (day < 7) return 'il y a ' + day + ' j';
    var week = Math.round(day / 7);
    if (week < 5) return 'il y a ' + week + ' sem.';
    var month = Math.round(day / 30);
    if (month < 12) return 'il y a ' + month + ' mois';
    var year = Math.round(day / 365);
    return 'il y a ' + year + ' an' + (year > 1 ? 's' : '');
  }

  function formatTime(iso) {
    var d = new Date(iso);
    if (isNaN(d.getTime())) return '';
    return d.toLocaleTimeString('fr-FR', { hour: '2-digit', minute: '2-digit' });
  }

  function renderMarkdown(raw) {
    var text = String(raw === null || raw === undefined ? '' : raw);
    var blocks = [];
    var withPlaceholders = text.replace(/```([\s\S]*?)```/g, function (m, code) {
      var idx = blocks.length;
      blocks.push(code.replace(/^\n/, '').replace(/\n$/, ''));
      return '@@CODE' + idx + '@@';
    });
    var escaped = escapeHtml(withPlaceholders);
    var html = escaped.replace(/`([^`]+)`/g, function (m, c) { return '<code>' + c + '</code>'; });
    html = html.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');
    var lines = html.split('\n');
    var out = [];
    var inList = false;
    for (var i = 0; i < lines.length; i++) {
      var line = lines[i];
      var m = line.match(/^\s*[-*]\s+(.*)$/);
      if (m) {
        if (!inList) { out.push('<ul>'); inList = true; }
        out.push('<li>' + m[1] + '</li>');
      } else {
        if (inList) { out.push('</ul>'); inList = false; }
        if (line.trim() === '') out.push('');
        else out.push('<p>' + line + '</p>');
      }
    }
    if (inList) out.push('</ul>');
    html = out.join('');
    html = html.replace(/@@CODE(\d+)@@/g, function (m, idx) {
      return '<pre class="code-block"><code>' + escapeHtml(blocks[Number(idx)]) + '</code></pre>';
    });
    return html;
  }

  // ---------------------------------------------------------------------
  // Couche API
  // ---------------------------------------------------------------------

  function ApiError(status, data) {
    this.status = status;
    this.data = data;
    this.message = (data && (data.error || data.output)) || ('HTTP ' + status);
  }
  ApiError.prototype = Object.create(Error.prototype);

  function api(path, opts) {
    opts = opts || {};
    var headers = Object.assign({ 'Content-Type': 'application/json' }, opts.headers || {});
    return fetch(path, {
      method: opts.method || 'GET',
      credentials: 'same-origin',
      headers: headers,
      body: opts.body !== undefined ? JSON.stringify(opts.body) : undefined
    }).then(function (res) {
      if (res.status === 401) {
        showLogin();
        throw new ApiError(401, null);
      }
      if (res.status === 204 || res.status === 202) return null;
      return res.text().then(function (text) {
        var data = null;
        if (text) { try { data = JSON.parse(text); } catch (e) { data = null; } }
        if (!res.ok) throw new ApiError(res.status, data);
        return data;
      });
    });
  }

  // ---------------------------------------------------------------------
  // Gestion de l'état : hydratation, index, upserts
  // ---------------------------------------------------------------------

  function reindex() {
    state.projectsById = {}; state.projects.forEach(function (p) { state.projectsById[p.id] = p; });
    state.cardsById = {}; state.cards.forEach(function (c) { state.cardsById[c.id] = c; });
    state.tasksById = {}; state.tasks.forEach(function (t) { state.tasksById[t.id] = t; });
    state.agentsById = {}; state.agents.forEach(function (a) { state.agentsById[a.id] = a; });
  }

  function hydrateState(data) {
    state.projects = data.projects || [];
    state.cards = data.cards || [];
    state.tasks = data.tasks || [];
    state.agents = data.agents || [];
    state.tokens = data.tokens || { global: { input: 0, output: 0, costUsd: 0 } };
    reindex();
  }

  function upsertProject(p) {
    var i = state.projects.findIndex(function (x) { return x.id === p.id; });
    if (i >= 0) state.projects[i] = p; else state.projects.push(p);
    state.projectsById[p.id] = p;
  }
  function upsertCard(c) {
    var i = state.cards.findIndex(function (x) { return x.id === c.id; });
    if (i >= 0) state.cards[i] = c; else state.cards.push(c);
    state.cardsById[c.id] = c;
  }
  function upsertTask(t) {
    if (!t) return;
    var i = state.tasks.findIndex(function (x) { return x.id === t.id; });
    if (i >= 0) state.tasks[i] = t; else state.tasks.push(t);
    state.tasksById[t.id] = t;
  }

  function projectUnread(pid) {
    // Project.unread n'a pas d'événement SSE dédié : recalculé en direct
    // à partir des tâches connues pour rester exact en temps réel.
    return state.tasks.filter(function (t) { return t.projectId === pid && t.unread; }).length;
  }

  function activeAgentsForCard(cardId) {
    var ids = {};
    state.tasks.forEach(function (t) { if (t.cardId === cardId && t.status === 'running') ids[t.agentId] = true; });
    return Object.keys(ids).map(function (id) { return state.agentsById[id]; }).filter(Boolean);
  }

  // ---------------------------------------------------------------------
  // Tiroir sidebar (mobile < 860px)
  // ---------------------------------------------------------------------

  function syncSidebarDrawer() {
    var sidebar = document.getElementById('sidebar');
    var backdrop = document.getElementById('sidebar-backdrop');
    if (sidebar) sidebar.classList.toggle('open', state.sidebarOpen);
    if (backdrop) backdrop.classList.toggle('hidden', !state.sidebarOpen);
  }
  function closeSidebarDrawer() { state.sidebarOpen = false; syncSidebarDrawer(); }
  function toggleSidebarDrawer() { state.sidebarOpen = !state.sidebarOpen; syncSidebarDrawer(); }

  // ---------------------------------------------------------------------
  // Navigation
  // ---------------------------------------------------------------------

  function goInbox() { closeSidebarDrawer(); state.screen = 'inbox'; state.projectId = null; state.cardId = null; state.taskId = null; render(); }
  function goAllProjects() { closeSidebarDrawer(); state.screen = 'projects'; state.projectId = null; state.cardId = null; state.taskId = null; render(); }
  function goProject(id) { closeSidebarDrawer(); state.screen = 'kanban'; state.projectId = id; state.cardId = null; state.taskId = null; render(); }
  function goCard(id) {
    var c = state.cardsById[id];
    if (!c) return;
    state.screen = 'work'; state.cardId = id; state.projectId = c.projectId; state.taskId = null; state.taskFilter = 'all';
    render();
  }
  function goBack() { if (state.projectId) goProject(state.projectId); else goAllProjects(); }
  function closePanel() { state.taskId = null; render(); }
  function setFilter(f) { state.taskFilter = f; renderMain(); }
  function setTab(t) { state.panelTab = t; renderMain(); }

  function render() { renderSidebar(); renderMain(); }

  function openTask(taskId) {
    state.taskId = taskId;
    state.panelTab = 'chat';
    var t = state.tasksById[taskId];
    if (t) t.unread = false;
    // Le chargement des messages est délégué à loadMessages(), appelé
    // paresseusement par buildConversationHTML() au premier rendu de l'onglet.
    render();
    api('/api/tasks/' + taskId + '/read', { method: 'POST' }).catch(function () {});
  }

  function openTaskFromSearch(taskId) {
    var t = state.tasksById[taskId];
    if (!t) return;
    goCard(t.cardId);
    openTask(taskId);
  }

  // ---------------------------------------------------------------------
  // Rendu : sidebar
  // ---------------------------------------------------------------------

  function buildSidebarHTML() {
    var navItems = [
      { key: 'inbox', icon: '⌂', label: 'Boîte de réception', action: 'nav-inbox' },
      { key: 'projects', icon: '◫', label: 'Tous les projets', action: 'nav-projects' }
    ];
    var navHTML = navItems.map(function (n) {
      var active = state.screen === n.key;
      return '<button class="nav-item ' + (active ? 'active' : '') + '" data-action="' + n.action + '">' +
        '<span class="nav-icon">' + n.icon + '</span>' + n.label + '</button>';
    }).join('');

    var projectsHTML = state.projects.map(function (p) {
      var active = (state.screen === 'kanban' || state.screen === 'work') && state.projectId === p.id;
      var unread = projectUnread(p.id);
      return '<button class="project-item ' + (active ? 'active' : '') + '" data-action="nav-project" data-project-id="' + p.id + '">' +
        '<span class="hash">#</span><span class="project-name">' + escapeHtml(p.name) + '</span>' +
        (unread ? '<span class="badge-unread">' + unread + '</span>' : '') +
        '</button>';
    }).join('');

    var agentsHTML = state.agents.map(function (a) {
      return '<div class="agent-item" data-agent-id="' + a.id + '">' +
        '<span class="agent-avatar" style="background:' + softColor(a.color) + '">' + a.emoji + '</span>' +
        '<span class="agent-name">' + escapeHtml(a.name) + '</span>' +
        (a.active ? '<span class="agent-dot"></span>' : '') +
        '</div>';
    }).join('');

    return '' +
      '<div class="sidebar-brand"><span class="brand-mark"></span><span class="brand-name">Atelier</span></div>' +
      '<div class="sidebar-search-wrap">' +
        '<button class="search-btn" data-action="open-search"><span>⌕</span><span class="search-btn-label">Rechercher</span>' +
        '<span class="kbd">' + (isMac() ? '⌘K' : 'Ctrl+K') + '</span></button>' +
      '</div>' +
      '<div class="nav-list">' + navHTML + '</div>' +
      '<div class="sidebar-section-head"><span>Projets</span>' +
      '<button class="icon-btn-sm" data-action="open-new-project" title="Nouveau projet">+</button></div>' +
      '<nav class="project-list">' + (projectsHTML || '<div class="empty-note-sm">Aucun projet</div>') + '</nav>' +
      '<div class="sidebar-section-head"><span>Agents</span></div>' +
      '<div class="agent-list">' + (agentsHTML || '<div class="empty-note-sm">Aucun agent</div>') + '</div>' +
      '<div class="sidebar-footer">' +
        '<span class="sidebar-tokens" id="sidebar-tokens">' + tokenSummary(state.tokens.global) + '</span>' +
        '<button class="logout-link" data-action="logout">Déconnexion</button>' +
      '</div>';
  }

  function renderSidebar() {
    var el = document.getElementById('sidebar');
    if (el) el.innerHTML = buildSidebarHTML();
  }

  // ---------------------------------------------------------------------
  // Rendu : en-tête commun
  // ---------------------------------------------------------------------

  function buildHeaderHTML() {
    var back = '', title = '', sub = '', actions = '';
    if (state.screen === 'inbox') {
      title = 'Boîte de réception';
    } else if (state.screen === 'projects') {
      title = 'Tous les projets';
      actions = '<button class="btn-outline" data-action="open-new-project">+ Nouveau projet</button>';
    } else if (state.screen === 'kanban') {
      var p = state.projectsById[state.projectId];
      title = p ? p.name : '';
    } else if (state.screen === 'work') {
      back = '<button class="icon-btn" data-action="go-back">←</button>';
      var c = state.cardsById[state.cardId];
      var pr = state.projectsById[state.projectId];
      title = c ? c.title : '';
      sub = pr ? '<span class="crumb-sub">' + escapeHtml(pr.name) + '</span>' : '';
      actions = '<button class="btn-primary" data-action="open-new-task">Nouvelle tâche</button>';
    }
    var hamburger = '<button class="icon-btn hamburger-btn" data-action="toggle-sidebar" aria-label="Menu">☰</button>';
    return '<header class="topbar">' + hamburger + back + '<span class="crumb-title">' + escapeHtml(title) + '</span>' + sub +
      '<div class="topbar-actions">' + actions + '</div></header>';
  }

  // ---------------------------------------------------------------------
  // Rendu : Tous les projets
  // ---------------------------------------------------------------------

  function buildAllProjectsHTML() {
    if (state.projects.length === 0) {
      return buildHeaderHTML() + '<div class="view-body"><div class="empty-state big">' +
        '<div class="empty-title">Aucun projet pour l\'instant</div>' +
        '<div class="empty-sub">Créez votre premier projet pour commencer.</div>' +
        '<button class="btn-primary" data-action="open-new-project">Nouveau projet</button>' +
        '</div></div>';
    }
    var tiles = state.projects.map(function (p) {
      var cardCount = state.cards.filter(function (c) { return c.projectId === p.id; }).length;
      var unread = projectUnread(p.id);
      return '<article class="project-tile" data-action="nav-project" data-project-id="' + p.id + '">' +
        '<div class="project-tile-top"><span class="project-hash">#</span><h3>' + escapeHtml(p.name) + '</h3>' +
        (unread ? '<span class="badge-unread">' + unread + '</span>' : '') + '</div>' +
        '<div class="project-tile-meta"><span>' + cardCount + ' carte' + (cardCount > 1 ? 's' : '') + '</span>' +
        '<span>' + tokenSummary(p.tokens) + '</span></div>' +
        '</article>';
    }).join('');
    return buildHeaderHTML() + '<div class="view-body all-projects-body"><div class="project-grid">' + tiles +
      '<button class="project-tile project-tile-add" data-action="open-new-project">+ Nouveau projet</button>' +
      '</div></div>';
  }

  // ---------------------------------------------------------------------
  // Rendu : Kanban
  // ---------------------------------------------------------------------

  function buildKanbanCardHTML(c, colKey) {
    var agents = activeAgentsForCard(c.id);
    var avatarsHTML = agents.map(function (a) {
      return '<span class="card-avatar" style="background:' + softColor(a.color) + '">' + a.emoji + '</span>';
    }).join('');
    var barColor = colKey === 'done' ? 'var(--green-live)' : 'var(--accent)';
    var liveHTML = c.liveActivity ? '<div class="card-live"><span class="live-dot"></span><span class="live-text mono">' +
      escapeHtml(c.liveActivity) + '</span></div>' : '';
    var attention = c.reviewCount ? '<span class="card-attention">' + c.reviewCount + ' à relire</span>' : '';
    var others = COLUMN_ORDER.filter(function (k) { return k !== colKey; });
    var menuItemsHTML = others.map(function (k) {
      return '<button class="card-menu-item" data-action="move-card" data-card-id="' + c.id + '" data-column="' + k + '">' +
        'Déplacer vers ' + COLUMN_LABELS[k] + '</button>';
    }).join('');
    return '<article class="kanban-card" data-action="open-card" data-card-id="' + c.id + '">' +
      '<div class="card-top">' +
        '<h3 class="card-title ' + (colKey === 'done' ? 'card-title-done' : '') + '">' + escapeHtml(c.title) + '</h3>' +
        '<div class="card-avatars">' + avatarsHTML + '</div>' +
        '<button class="card-menu-btn" data-action="toggle-card-menu" data-card-id="' + c.id + '">⋯</button>' +
        '<div class="card-menu hidden" data-card-menu="' + c.id + '">' + menuItemsHTML + '</div>' +
      '</div>' +
      liveHTML +
      '<div class="card-meta">' +
        '<span>' + (c.tasksDone || 0) + '/' + (c.tasksTotal || 0) + ' tâches</span>' +
        '<span>▤ ' + (c.docsCount || 0) + '</span>' +
        '<span>💬 ' + (c.messagesCount || 0) + '</span>' +
        attention +
      '</div>' +
      '<div class="progress-track"><div class="progress-fill" style="width:' + (c.progress || 0) + '%; background:' + barColor + '"></div></div>' +
      '</article>';
  }

  function buildKanbanHTML() {
    var project = state.projectsById[state.projectId];
    var cards = state.cards.filter(function (c) { return c.projectId === state.projectId; });
    var doingCount = cards.filter(function (c) { return c.column === 'doing'; }).length;
    var totalReview = cards.reduce(function (n, c) { return n + (c.reviewCount || 0); }, 0);
    var totalTasks = cards.reduce(function (n, c) { return n + (c.tasksTotal || 0); }, 0);
    var tokenTxt = project ? tokenSummary(project.tokens) : '';

    var colsHTML = COLUMN_ORDER.map(function (key) {
      var label = COLUMN_LABELS[key];
      var color = key === 'soon' ? '#d3d0c8' : (key === 'doing' ? '#2f7d54' : '#c3c0b8');
      var list = cards.filter(function (c) { return c.column === key; });
      return '<section class="kanban-col">' +
        '<div class="col-head"><span class="col-dot" style="background:' + color + '"></span>' +
        '<span class="col-label">' + label + '</span><span class="col-count">' + list.length + '</span></div>' +
        list.map(function (c) { return buildKanbanCardHTML(c, key); }).join('') +
        '<button class="add-card-btn" data-action="open-new-card" data-column="' + key + '">+ Ajouter</button>' +
        '</section>';
    }).join('');

    var head = '<div class="kanban-head"><h1>' + escapeHtml(project ? project.name : '') + '</h1>';
    if (cards.length > 0) {
      head += '<div class="kanban-stats">' +
        '<span><b>' + doingCount + '</b> en cours</span>' +
        '<span><b class="' + (totalReview ? 'amber' : '') + '">' + totalReview + '</b> à relire</span>' +
        '<span><b>' + totalTasks + '</b> tâches</span>' +
        '<span id="kanban-token-stat" class="token-stat">' + tokenTxt + '</span>' +
        '</div>';
    }
    head += '</div>';

    var emptyNote = cards.length === 0
      ? '<div class="empty-state">Aucune carte pour l\'instant. Utilisez « + Ajouter » dans une colonne pour commencer.</div>'
      : '';

    return buildHeaderHTML() + '<div class="view-body kanban-body">' + head + emptyNote +
      '<div class="kanban-columns">' + colsHTML + '</div></div>';
  }

  // ---------------------------------------------------------------------
  // Rendu : liste de tâches (partagé par Carte et Boîte de réception)
  // ---------------------------------------------------------------------

  function buildTaskRowHTML(t, showProject) {
    var glyph = STATUS_GLYPH[t.status] || STATUS_GLYPH.running;
    var agent = state.agentsById[t.agentId] || { emoji: '', name: '', color: '#ccc' };
    var selected = state.taskId === t.id;
    var isNew = t.unread && !selected;
    var projectTag = '';
    if (showProject) {
      var p = state.projectsById[t.projectId];
      var c = state.cardsById[t.cardId];
      projectTag = '<span class="task-project-tag">' + (p ? escapeHtml(p.name) : '') +
        (c ? ' · ' + escapeHtml(c.title) : '') + '</span>';
    }
    var liveLine = (t.status === 'running' && t.liveActivity)
      ? '<span class="live-line mono"><span class="live-dot"></span>' + escapeHtml(t.liveActivity) + '</span>' : '';
    return '<div class="task-row ' + (selected ? 'selected' : '') + '" data-action="open-task" data-task-id="' + t.id + '">' +
      '<span class="task-glyph" style="color:' + glyph.color + '">' + glyph.icon + '</span>' +
      '<div class="task-main">' +
        '<div class="task-title-line"><span class="task-title">' + escapeHtml(t.title) + '</span>' +
        (isNew ? '<span class="badge-new">NEW</span>' : '') + '</div>' +
        '<div class="task-meta-line">' + projectTag +
        '<span class="mono">#' + t.ref + '</span>' +
        '<span class="agent-chip"><span class="agent-avatar-sm" style="background:' + softColor(agent.color) + '">' + agent.emoji + '</span>' + escapeHtml(agent.name) + '</span>' +
        '<span>' + timeAgo(t.updatedAt) + '</span>' + liveLine +
        '</div>' +
      '</div>' +
      '<div class="task-counts"><span>◆ ' + (t.filesCount || 0) + '</span><span>💬 ' + (t.messagesCount || 0) + '</span></div>' +
      '</div>';
  }

  function buildTaskListShell(tasksAll, opts) {
    var counts = tasksAll.reduce(function (acc, t) { acc[t.status] = (acc[t.status] || 0) + 1; return acc; }, {});
    var filters = [
      { key: 'all', label: 'Toutes ' + tasksAll.length },
      { key: 'review', label: 'À relire ' + (counts.review || 0) },
      { key: 'ready', label: 'Prêt à livrer ' + (counts.ready || 0) },
      { key: 'shipped', label: 'Livré ' + (counts.shipped || 0) }
    ];
    var filterHTML = filters.map(function (f) {
      return '<button class="pill ' + (state.taskFilter === f.key ? 'pill-active' : '') + '" data-action="set-filter" data-filter="' + f.key + '">' + f.label + '</button>';
    }).join('');
    var visible = tasksAll.filter(function (t) { return state.taskFilter === 'all' || t.status === state.taskFilter; })
      .sort(function (a, b) { return new Date(b.updatedAt) - new Date(a.updatedAt); });
    var rowsHTML = visible.length
      ? visible.map(function (t) { return buildTaskRowHTML(t, opts.showProject); }).join('')
      : '<div class="empty-state">' + opts.emptyMsg + '</div>';
    var task = state.taskId ? state.tasksById[state.taskId] : null;
    var panelHTML = task ? buildDetailPanelHTML(task) : '';
    return buildHeaderHTML() + '<div class="view-body work-body ' + (task ? 'has-panel' : '') + '" style="padding:0;">' +
      '<div class="task-list-pane"><div class="filter-pills">' + filterHTML + '</div>' +
      '<div class="task-list">' + rowsHTML + '</div></div>' +
      panelHTML +
      '</div>';
  }

  function buildWorkHTML() {
    var tasksAll = state.tasks.filter(function (t) { return t.cardId === state.cardId; });
    return buildTaskListShell(tasksAll, { showProject: false, emptyMsg: 'Aucune tâche pour l\'instant.' });
  }

  function buildInboxHTML() {
    var tasksAll = state.tasks.filter(function (t) { return t.unread || t.status === 'review'; });
    return buildTaskListShell(tasksAll, { showProject: true, emptyMsg: 'Boîte de réception vide. Tout est à jour !' });
  }

  // ---------------------------------------------------------------------
  // Rendu : panneau de détail de tâche
  // ---------------------------------------------------------------------

  function primaryActionInfo(task) {
    switch (task.status) {
      case 'running':
        return { label: 'Interrompre l\'agent', cls: 'btn-neutral', action: 'interrupt', defaultLabel: 'Interrompre l\'agent' };
      case 'review':
        return { label: 'Accepter le diff et les livrables', cls: 'btn-green', action: 'accept', defaultLabel: 'Accepter le diff et les livrables' };
      case 'ready': {
        var pending = state.pendingShip && state.pendingShip.taskId === task.id;
        return { label: pending ? 'Confirmer le push ?' : 'Pousser et livrer', cls: 'btn-green', action: 'ship-click', defaultLabel: 'Pousser et livrer' };
      }
      case 'shipped':
        return { label: 'Rouvrir la tâche', cls: 'btn-neutral', action: 'reopen', defaultLabel: 'Rouvrir la tâche' };
      default:
        return { label: '', cls: 'btn-neutral', action: '', defaultLabel: '' };
    }
  }

  function buildWorkflowHTML(status) {
    var order = { running: 0, review: 0, ready: 1, shipped: 2 };
    var steps = [{ label: 'À relire' }, { label: 'Accepté' }, { label: 'Livré' }];
    var cur = order[status] !== undefined ? order[status] : 0;
    var html = '<div class="workflow">';
    steps.forEach(function (s, i) {
      var done = i <= cur;
      var isCur = i === cur;
      var barClass = !done ? 'wf-bar-todo' : (isCur ? 'wf-bar-current' : 'wf-bar-done');
      var lblClass = done ? 'wf-label-done' : '';
      html += '<div class="wf-step"><span class="wf-bar ' + barClass + '"></span>' +
        '<span class="wf-label ' + lblClass + ' ' + (isCur ? 'wf-current' : '') + '">' + s.label + '</span></div>';
    });
    html += '</div>';
    return html;
  }

  function renderChecks(checks) {
    if (!checks || checks.length === 0) return '';
    return checks.map(function (c) {
      return '<span class="check ' + (c.ok ? 'check-ok' : 'check-fail') + '">' + (c.ok ? '✓' : '✕') + ' ' + escapeHtml(c.label) + '</span>';
    }).join('');
  }

  function buildDetailPanelHTML(task) {
    var agent = state.agentsById[task.agentId] || { emoji: '?', name: '?', model: '?', color: '#ccc' };
    var glyph = STATUS_GLYPH[task.status] || STATUS_GLYPH.running;
    var soft = softColor(agent.color);
    var err = state.detailErrorByTask[task.id];
    var action = primaryActionInfo(task);
    var tabs = ['chat', 'diff', 'files'];
    var tabLabels = { chat: 'Conversation', diff: 'Diff', files: 'Livrables' };
    var tabCounts = { chat: task.messagesCount || 0, diff: task.filesCount || 0, files: (task.docsCount || 0) + (task.filesCount || 0) };
    var tabDataAttr = { chat: 'conversation', diff: 'diff', files: 'deliverables' };

    var tabsHTML = tabs.map(function (t) {
      var active = state.panelTab === t;
      var isNew = t === 'chat' && task.unread && !active;
      return '<button class="tab ' + (active ? 'tab-active' : '') + '" role="tab" data-tab="' + tabDataAttr[t] + '" data-action="set-tab" data-panel-tab="' + t + '">' +
        tabLabels[t] + '<span class="tab-count">' + tabCounts[t] + '</span>' +
        (isNew ? '<span class="tab-dot"></span>' : '') + '</button>';
    }).join('');

    var bodyHTML = '';
    if (state.panelTab === 'chat') bodyHTML = buildConversationHTML(task, agent);
    else if (state.panelTab === 'diff') bodyHTML = buildDiffHTML(task);
    else if (state.panelTab === 'files') bodyHTML = buildDeliverablesHTML(task);

    return '<aside class="detail-panel">' +
      (err ? '<div class="detail-error">' + escapeHtml(err) + '</div>' : '') +
      '<div class="detail-head">' +
        '<div class="detail-head-row">' +
          '<span class="task-glyph" style="color:' + glyph.color + '">' + glyph.icon + '</span>' +
          '<div class="detail-head-main">' +
            '<div class="detail-title">' + escapeHtml(task.title) + '</div>' +
            '<div class="detail-meta">' +
              '<span class="agent-chip"><span class="agent-avatar-sm" style="background:' + soft + '">' + agent.emoji + '</span>' + escapeHtml(agent.name) + '</span>' +
              '<span class="mono">' + escapeHtml(agent.model || '') + '</span>' +
              '<span class="mono">' + escapeHtml(task.branch || '') + '</span>' +
            '</div>' +
          '</div>' +
          '<button class="icon-btn" data-action="close-panel">✕</button>' +
        '</div>' +
        '<div class="detail-tokens" id="detail-token-line">' + tokenSummary(task.tokens) + '</div>' +
        buildWorkflowHTML(task.status) +
        '<div class="action-row">' +
          '<button id="task-primary-action" class="btn-action ' + action.cls + '" data-action="' + action.action + '" data-task-id="' + task.id + '" data-default-label="' + escapeHtml(action.defaultLabel) + '">' + action.label + '</button>' +
          '<span class="checks">' + renderChecks(task.checks) + '</span>' +
        '</div>' +
        '<div class="tabs">' + tabsHTML + '</div>' +
      '</div>' +
      '<div class="tab-body">' + bodyHTML + '</div>' +
      '</aside>';
  }

  // ---------------------------------------------------------------------
  // Onglet Conversation
  // ---------------------------------------------------------------------

  function buildMessageHTML(m, agent) {
    var isUser = m.author === 'user';
    var emoji = isUser ? '🙂' : (agent.emoji || '');
    var bg = isUser ? '#eeece6' : softColor(agent.color);
    var name = isUser ? 'Vous' : (agent.name || '');
    return '<div class="message">' +
      '<span class="msg-avatar" style="background:' + bg + '">' + emoji + '</span>' +
      '<div class="msg-body">' +
        '<div class="msg-head"><span class="msg-name">' + escapeHtml(name) + '</span><span class="msg-time">' + formatTime(m.createdAt) + '</span></div>' +
        '<div class="msg-text">' + renderMarkdown(m.text) + '</div>' +
      '</div></div>';
  }

  function buildConversationHTML(task, agent) {
    var msgs = state.messagesByTask[task.id];
    if (!msgs) {
      loadMessages(task.id);
      return '<div class="conv-loading">Chargement…</div>';
    }
    var items = msgs.length ? msgs.map(function (m) { return buildMessageHTML(m, agent); }).join('') : '<div class="empty-note">Aucun message pour l\'instant.</div>';
    return '<div class="conversation" id="conversation-list">' + items + '</div>' +
      '<div class="composer-wrap"><div class="composer">' +
        '<textarea id="composer-input" rows="2" placeholder="Répondre à ' + escapeHtml(agent.name) + '…"></textarea>' +
        '<div class="composer-row">' +
          '<span class="composer-model"><span class="agent-avatar-sm" style="background:' + softColor(agent.color) + '">' + agent.emoji + '</span>' + escapeHtml(agent.model || '') + '</span>' +
          '<button class="btn-send" data-action="send-message" data-task-id="' + task.id + '">Envoyer ⏎</button>' +
        '</div>' +
      '</div></div>';
  }

  function loadMessages(taskId) {
    var key = 'msg:' + taskId;
    if (state.loading[key]) return;
    state.loading[key] = true;
    api('/api/tasks/' + taskId).then(function (data) {
      state.messagesByTask[taskId] = (data && data.messages) || [];
      if (data && data.task) upsertTask(data.task);
    }).catch(function () {
      state.messagesByTask[taskId] = [];
    }).then(function () {
      delete state.loading[key];
      renderMain();
    });
  }

  function scrollConversationToBottom() {
    var el = document.getElementById('conversation-list');
    if (el) el.scrollTop = el.scrollHeight;
  }

  function sendMessage(taskId) {
    var el = document.getElementById('composer-input');
    if (!el) return;
    var text = el.value.trim();
    if (!text) return;
    el.value = '';
    el.disabled = true;
    api('/api/tasks/' + taskId + '/messages', { method: 'POST', body: { text: text } })
      .catch(function () {})
      .then(function () {
        var el2 = document.getElementById('composer-input');
        if (el2) el2.disabled = false;
      });
  }

  // ---------------------------------------------------------------------
  // Onglet Diff
  // ---------------------------------------------------------------------

  function loadDiff(taskId) {
    var key = 'diff:' + taskId;
    if (state.loading[key]) return;
    state.loading[key] = true;
    api('/api/tasks/' + taskId + '/diff').then(function (data) {
      state.diffByTask[taskId] = data || { files: [] };
    }).catch(function () {
      state.diffByTask[taskId] = { files: [] };
    }).then(function () {
      delete state.loading[key];
      renderMain();
    });
  }

  function buildDiffHTML(task) {
    var diff = state.diffByTask[task.id];
    if (!diff) {
      loadDiff(task.id);
      return '<div class="conv-loading">Chargement…</div>';
    }
    if (!diff.files || diff.files.length === 0) {
      return '<div class="empty-note">Aucune modification.</div>';
    }
    var activePath = state.activeDiffFile[task.id] || diff.files[0].path;
    var activeFile = diff.files.filter(function (f) { return f.path === activePath; })[0] || diff.files[0];
    var fileTabs = diff.files.map(function (f) {
      return '<button class="diff-file-tab ' + (f.path === activeFile.path ? 'active' : '') + '" data-action="select-diff-file" data-task-id="' + task.id + '" data-path="' + escapeHtml(f.path) + '">' +
        escapeHtml(f.path) + ' <span class="diff-add">+' + f.additions + '</span> <span class="diff-del">-' + f.deletions + '</span></button>';
    }).join('');
    var hunks = (activeFile.hunks || []).map(function (h) {
      var lines = (h.lines || []).map(function (l) {
        var mark = l.type === 'add' ? '+' : (l.type === 'del' ? '-' : ' ');
        return '<div class="diff-line diff-' + l.type + '"><span class="diff-mark">' + mark + '</span><span class="diff-text mono">' + escapeHtml(l.text) + '</span></div>';
      }).join('');
      return '<div><div class="diff-hunk-header mono">' + escapeHtml(h.header) + '</div>' + lines + '</div>';
    }).join('');
    var canShip = task.status === 'ready';
    var pending = state.pendingShip && state.pendingShip.taskId === task.id;
    var pushLabel = pending && canShip ? 'Confirmer le push ?' : 'Push';
    return '<div class="diff-subtabs">' + fileTabs + '</div>' +
      '<div class="diff-hunks">' + hunks + '</div>' +
      '<div class="diff-footer">' +
        '<span class="mono diff-branch">' + escapeHtml(diff.branch || '') + ' → ' + escapeHtml(diff.base || '') + '</span>' +
        '<button class="btn-outline" disabled title="bientôt">Ouvrir la PR</button>' +
        '<button class="btn-green" ' + (canShip ? '' : 'disabled title="Disponible une fois prêt à livrer"') +
        ' data-action="ship-click" data-task-id="' + task.id + '" data-default-label="Push">' + pushLabel + '</button>' +
      '</div>';
  }

  function selectDiffFile(taskId, path) {
    state.activeDiffFile[taskId] = path;
    renderMain();
  }

  // ---------------------------------------------------------------------
  // Onglet Livrables
  // ---------------------------------------------------------------------

  function loadDeliverables(taskId) {
    var key = 'deliv:' + taskId;
    if (state.loading[key]) return;
    state.loading[key] = true;
    api('/api/tasks/' + taskId + '/deliverables').then(function (data) {
      state.deliverablesByTask[taskId] = data || { code: [], docs: [], images: [] };
    }).catch(function () {
      state.deliverablesByTask[taskId] = { code: [], docs: [], images: [] };
    }).then(function () {
      delete state.loading[key];
      renderMain();
    });
  }

  function buildDeliverablesHTML(task) {
    var d = state.deliverablesByTask[task.id];
    if (!d) {
      loadDeliverables(task.id);
      return '<div class="conv-loading">Chargement…</div>';
    }
    var groups = [
      { key: 'code', label: 'Code', icon: '◆' },
      { key: 'docs', label: 'Documents', icon: '▤' },
      { key: 'images', label: 'Captures', icon: '▣' }
    ];
    var html = groups.map(function (g) {
      var items = d[g.key] || [];
      var itemsHTML = items.length ? items.map(function (it) {
        return '<div class="deliv-item"><span class="deliv-icon">' + g.icon + '</span>' +
          '<div class="deliv-main"><div class="deliv-title">' + escapeHtml(it.title) + '</div>' +
          '<div class="deliv-meta mono">' + escapeHtml(it.meta || '') + '</div></div></div>';
      }).join('') : '<div class="empty-note">Aucun élément.</div>';
      return '<section class="deliv-group"><div class="deliv-group-label">' + g.label + '</div>' + itemsHTML + '</section>';
    }).join('');
    return '<div class="deliverables">' + html + '</div>';
  }

  // ---------------------------------------------------------------------
  // Actions de tâche
  // ---------------------------------------------------------------------

  function showDetailError(taskId, msg) {
    state.detailErrorByTask[taskId] = msg;
    renderMain();
    setTimeout(function () {
      delete state.detailErrorByTask[taskId];
      var el = document.querySelector('.detail-error');
      if (el) el.remove();
    }, 6000);
  }

  function doInterrupt(taskId) {
    api('/api/tasks/' + taskId + '/interrupt', { method: 'POST' }).then(function (task) {
      upsertTask(task); renderMain();
    }).catch(function (e) { if (e instanceof ApiError) showDetailError(taskId, e.message || 'Échec de l\'interruption.'); });
  }
  function doAccept(taskId) {
    api('/api/tasks/' + taskId + '/accept', { method: 'POST' }).then(function (task) {
      upsertTask(task); renderMain();
    }).catch(function (e) { if (e instanceof ApiError) showDetailError(taskId, e.message || 'Échec.'); });
  }
  function doReopen(taskId) {
    api('/api/tasks/' + taskId + '/reopen', { method: 'POST' }).then(function (task) {
      upsertTask(task); renderMain();
    }).catch(function (e) { if (e instanceof ApiError) showDetailError(taskId, e.message || 'Échec.'); });
  }
  function doShip(taskId) {
    api('/api/tasks/' + taskId + '/ship', { method: 'POST', body: { confirm: true } }).then(function (res) {
      if (res && res.task) upsertTask(res.task);
      renderMain();
    }).catch(function (e) { if (e instanceof ApiError) showDetailError(taskId, e.message || 'Échec du push.'); });
  }
  function patchShipButtons(taskId) {
    var els = document.querySelectorAll('[data-action="ship-click"][data-task-id="' + taskId + '"]');
    els.forEach(function (btn) {
      var pending = state.pendingShip && state.pendingShip.taskId === taskId;
      btn.textContent = pending ? 'Confirmer le push ?' : btn.getAttribute('data-default-label');
    });
  }
  function handleShipClick(taskId) {
    var now = Date.now();
    var pending = state.pendingShip;
    if (pending && pending.taskId === taskId) {
      clearTimeout(pending.timer);
      state.pendingShip = null;
      doShip(taskId);
      return;
    }
    if (pending) clearTimeout(pending.timer);
    var timer = setTimeout(function () { state.pendingShip = null; patchShipButtons(taskId); }, 5000);
    state.pendingShip = { taskId: taskId, timer: timer };
    patchShipButtons(taskId);
  }

  function moveCard(cardId, column) {
    closeAllCardMenus();
    api('/api/cards/' + cardId, { method: 'PATCH', body: { column: column } }).then(function (updated) {
      upsertCard(updated); renderMain();
    }).catch(function () {});
  }

  function closeAllCardMenus() {
    document.querySelectorAll('.card-menu').forEach(function (m) { m.classList.add('hidden'); });
  }
  function toggleCardMenu(cardId) {
    var el = document.querySelector('.card-menu[data-card-menu="' + cardId + '"]');
    if (!el) return;
    var willOpen = el.classList.contains('hidden');
    closeAllCardMenus();
    if (willOpen) el.classList.remove('hidden');
  }

  // ---------------------------------------------------------------------
  // Rendu principal
  // ---------------------------------------------------------------------

  function buildMainHTML() {
    if (state.cardId) return buildWorkHTML();
    if (state.projectId) return buildKanbanHTML();
    if (state.screen === 'inbox') return buildInboxHTML();
    return buildAllProjectsHTML();
  }

  function renderMain() {
    var mainEl = document.getElementById('main');
    if (!mainEl) return;
    var prevComposer = document.getElementById('composer-input');
    var hadFocus = prevComposer && document.activeElement === prevComposer;
    var draft = prevComposer ? prevComposer.value : '';
    var selStart = prevComposer ? prevComposer.selectionStart : null;
    var prevConv = document.getElementById('conversation-list');
    var prevScroll = prevConv ? prevConv.scrollTop : null;

    mainEl.innerHTML = buildMainHTML();

    var newComposer = document.getElementById('composer-input');
    if (newComposer && draft) {
      newComposer.value = draft;
      if (hadFocus) {
        newComposer.focus();
        if (selStart !== null) { try { newComposer.setSelectionRange(selStart, selStart); } catch (e) {} }
      }
    }
    var newConv = document.getElementById('conversation-list');
    if (newConv && prevScroll !== null) newConv.scrollTop = prevScroll;
  }

  // ---------------------------------------------------------------------
  // Recherche (Ctrl/Cmd+K)
  // ---------------------------------------------------------------------

  function overlayBgCloseSearch(e) { if (e.target === e.currentTarget) closeSearch(); }

  function buildSearchResultsHTML(q) {
    var query = q.trim().toLowerCase();
    if (!query) return '<div class="empty-note">Tapez pour rechercher…</div>';
    var projects = state.projects.filter(function (p) { return p.name.toLowerCase().indexOf(query) !== -1; });
    var tasks = state.tasks.filter(function (t) {
      return t.title.toLowerCase().indexOf(query) !== -1 || String(t.ref).indexOf(query) !== -1;
    }).slice(0, 20);
    if (!projects.length && !tasks.length) return '<div class="empty-note">Aucun résultat.</div>';
    var html = '';
    if (projects.length) {
      html += '<div class="search-group-label">Projets</div>' + projects.map(function (p) {
        return '<button class="search-result" data-action="search-goto-project" data-project-id="' + p.id + '"><span class="hash">#</span>' + escapeHtml(p.name) + '</button>';
      }).join('');
    }
    if (tasks.length) {
      html += '<div class="search-group-label">Tâches</div>' + tasks.map(function (t) {
        var p = state.projectsById[t.projectId];
        return '<button class="search-result" data-action="search-goto-task" data-task-id="' + t.id + '"><span class="mono">#' + t.ref + '</span> ' +
          escapeHtml(t.title) + '<span class="muted-sm">' + (p ? escapeHtml(p.name) : '') + '</span></button>';
      }).join('');
    }
    return html;
  }

  function buildSearchHTML(q) {
    return '<div class="search-box"><input id="search-input" class="search-input" placeholder="Rechercher tâches et projets…" value="' + escapeHtml(q) + '">' +
      '<div class="search-results" id="search-results">' + buildSearchResultsHTML(q) + '</div></div>';
  }

  function openSearch() {
    state.searchOpen = true;
    var root = document.getElementById('search-overlay');
    root.classList.remove('hidden');
    root.innerHTML = buildSearchHTML('');
    root.addEventListener('click', overlayBgCloseSearch);
    var input = document.getElementById('search-input');
    input.focus();
    input.addEventListener('input', function () {
      var results = document.getElementById('search-results');
      if (results) results.innerHTML = buildSearchResultsHTML(input.value);
    });
  }
  function closeSearch() {
    state.searchOpen = false;
    var root = document.getElementById('search-overlay');
    root.classList.add('hidden');
    root.innerHTML = '';
    root.removeEventListener('click', overlayBgCloseSearch);
  }

  // ---------------------------------------------------------------------
  // Modales
  // ---------------------------------------------------------------------

  function overlayBgCloseModal(e) { if (e.target === e.currentTarget) closeModal(); }

  function openModal(html) {
    state.modal = true;
    var root = document.getElementById('modal-root');
    root.innerHTML = '<div class="modal-overlay" id="modal-overlay">' + html + '</div>';
    document.getElementById('modal-overlay').addEventListener('click', overlayBgCloseModal);
  }
  function closeModal() {
    state.modal = null;
    var root = document.getElementById('modal-root');
    root.innerHTML = '';
  }

  // Nouvelle tâche

  function buildNewTaskModalHTML(card) {
    var agentChoices = state.agents.map(function (a) {
      return '<button class="agent-choice ' + (a.id === modalAgentId ? 'selected' : '') + '" data-action="pick-agent" data-agent-id="' + a.id + '">' +
        '<span class="agent-choice-avatar" style="background:' + softColor(a.color) + '">' + a.emoji + '</span>' +
        '<span class="agent-choice-info"><span class="agent-choice-name">' + escapeHtml(a.name) + '</span>' +
        '<span class="agent-choice-model mono">' + escapeHtml(a.model || '') + '</span></span></button>';
    }).join('');
    var selected = state.agentsById[modalAgentId];
    return '<div class="modal">' +
      '<div class="modal-head"><span class="modal-title">Nouvelle tâche</span><span class="modal-sub">' + escapeHtml(card.title) + '</span>' +
      '<button class="icon-btn" data-action="close-modal">✕</button></div>' +
      '<input id="new-task-title" class="modal-input" placeholder="Que doit faire l\'agent ?">' +
      '<textarea id="new-task-prompt" class="modal-textarea" placeholder="Description ou instructions détaillées (optionnel)" rows="3"></textarea>' +
      '<div class="modal-label">Agent</div>' +
      '<div class="agent-choices" id="agent-choices">' + agentChoices + '</div>' +
      '<div class="agent-context-preview" id="agent-context-preview">' + (selected ? escapeHtml(selected.contextPrompt || '') : '') + '</div>' +
      '<div id="new-task-error" class="modal-error hidden"></div>' +
      '<div class="modal-foot"><span class="modal-hint">La conversation s\'ouvre après la création</span>' +
      '<button class="btn-outline" data-action="close-modal">Annuler</button>' +
      '<button class="btn-green" data-action="submit-new-task" data-card-id="' + card.id + '">Créer et discuter</button></div>' +
      '</div>';
  }

  function openNewTaskModal(cardId) {
    var card = state.cardsById[cardId];
    if (!card) return;
    modalAgentId = state.agents[0] ? state.agents[0].id : null;
    openModal(buildNewTaskModalHTML(card));
    setTimeout(function () { var el = document.getElementById('new-task-title'); if (el) el.focus(); }, 0);
  }

  function pickAgentInModal(agentId) {
    modalAgentId = agentId;
    document.querySelectorAll('#agent-choices .agent-choice').forEach(function (el) {
      el.classList.toggle('selected', el.getAttribute('data-agent-id') === agentId);
    });
    var preview = document.getElementById('agent-context-preview');
    var a = state.agentsById[agentId];
    if (preview) preview.textContent = a ? (a.contextPrompt || '') : '';
  }

  function submitNewTask(cardId) {
    var titleEl = document.getElementById('new-task-title');
    var promptEl = document.getElementById('new-task-prompt');
    var errEl = document.getElementById('new-task-error');
    var title = titleEl.value.trim();
    var prompt = promptEl.value.trim();
    if (!title) { errEl.textContent = 'Le titre est requis.'; errEl.classList.remove('hidden'); return; }
    if (!modalAgentId) { errEl.textContent = 'Choisissez un agent.'; errEl.classList.remove('hidden'); return; }
    var body = { cardId: cardId, title: title, agentId: modalAgentId };
    if (prompt) body.prompt = prompt;
    api('/api/tasks', { method: 'POST', body: body }).then(function (task) {
      upsertTask(task);
      closeModal();
      goCard(cardId);
      openTask(task.id);
    }).catch(function (e) {
      errEl.textContent = (e instanceof ApiError && e.message) || 'Erreur lors de la création.';
      errEl.classList.remove('hidden');
    });
  }

  // Nouveau projet

  function buildNewProjectModalHTML() {
    return '<div class="modal modal-sm">' +
      '<div class="modal-head"><span class="modal-title">Nouveau projet</span><button class="icon-btn" data-action="close-modal">✕</button></div>' +
      '<div class="modal-label">Nom</div><input id="new-project-name" class="modal-input" placeholder="mon-projet">' +
      '<div class="modal-label">Chemin absolu du dépôt git</div>' +
      '<input id="new-project-path" class="modal-input mono" placeholder="/home/utilisateur/projets/mon-projet">' +
      '<div id="new-project-error" class="modal-error hidden"></div>' +
      '<div class="modal-foot"><button class="btn-outline" data-action="close-modal">Annuler</button>' +
      '<button class="btn-green" data-action="submit-new-project">Créer</button></div>' +
      '</div>';
  }
  function openNewProjectModal() {
    openModal(buildNewProjectModalHTML());
    setTimeout(function () { var el = document.getElementById('new-project-name'); if (el) el.focus(); }, 0);
  }
  function submitNewProject() {
    var nameEl = document.getElementById('new-project-name');
    var pathEl = document.getElementById('new-project-path');
    var errEl = document.getElementById('new-project-error');
    var name = nameEl.value.trim();
    var path = pathEl.value.trim();
    if (!name || !path) { errEl.textContent = 'Nom et chemin sont requis.'; errEl.classList.remove('hidden'); return; }
    api('/api/projects', { method: 'POST', body: { name: name, path: path } }).then(function (project) {
      upsertProject(project);
      closeModal();
      goProject(project.id);
    }).catch(function (e) {
      errEl.textContent = (e instanceof ApiError && e.message) || 'Erreur lors de la création du projet.';
      errEl.classList.remove('hidden');
    });
  }

  // Nouvelle carte

  function buildNewCardModalHTML() {
    var pills = COLUMN_ORDER.map(function (k) {
      return '<button class="col-pill ' + (k === modalColumn ? 'selected' : '') + '" data-action="pick-column" data-column="' + k + '">' + COLUMN_LABELS[k] + '</button>';
    }).join('');
    return '<div class="modal modal-sm">' +
      '<div class="modal-head"><span class="modal-title">Nouvelle carte</span><button class="icon-btn" data-action="close-modal">✕</button></div>' +
      '<div class="modal-label">Titre</div><input id="new-card-title" class="modal-input" placeholder="Titre de la carte">' +
      '<div class="modal-label">Colonne</div><div class="col-pills" id="col-pills">' + pills + '</div>' +
      '<div id="new-card-error" class="modal-error hidden"></div>' +
      '<div class="modal-foot"><button class="btn-outline" data-action="close-modal">Annuler</button>' +
      '<button class="btn-green" data-action="submit-new-card">Créer</button></div>' +
      '</div>';
  }
  function openNewCardModal(column) {
    if (!state.projectId) return;
    modalColumn = column || 'soon';
    openModal(buildNewCardModalHTML());
    setTimeout(function () { var el = document.getElementById('new-card-title'); if (el) el.focus(); }, 0);
  }
  function pickColumnInModal(col) {
    modalColumn = col;
    document.querySelectorAll('#col-pills .col-pill').forEach(function (el) {
      el.classList.toggle('selected', el.getAttribute('data-column') === col);
    });
  }
  function submitNewCard() {
    var titleEl = document.getElementById('new-card-title');
    var errEl = document.getElementById('new-card-error');
    var title = titleEl.value.trim();
    if (!title) { errEl.textContent = 'Le titre est requis.'; errEl.classList.remove('hidden'); return; }
    api('/api/cards', { method: 'POST', body: { projectId: state.projectId, title: title, column: modalColumn } }).then(function (card) {
      upsertCard(card);
      closeModal();
      renderMain();
    }).catch(function (e) {
      errEl.textContent = (e instanceof ApiError && e.message) || 'Erreur lors de la création de la carte.';
      errEl.classList.remove('hidden');
    });
  }

  // ---------------------------------------------------------------------
  // Authentification
  // ---------------------------------------------------------------------

  function showLogin() {
    var shell = document.getElementById('shell');
    var login = document.getElementById('login-screen');
    if (shell) shell.classList.add('hidden');
    if (login) login.classList.remove('hidden');
    var pwd = document.getElementById('login-password');
    if (pwd) { pwd.value = ''; pwd.focus(); }
  }
  function hideLogin() {
    var shell = document.getElementById('shell');
    var login = document.getElementById('login-screen');
    if (login) login.classList.add('hidden');
    if (shell) shell.classList.remove('hidden');
  }
  function doLogout() {
    api('/api/logout', { method: 'POST' }).catch(function () {}).then(function () { showLogin(); });
  }

  function onLoginSubmit(e) {
    e.preventDefault();
    var pwdEl = document.getElementById('login-password');
    var errEl = document.getElementById('login-error');
    errEl.classList.add('hidden');
    fetch('/api/login', {
      method: 'POST', credentials: 'same-origin', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ password: pwdEl.value })
    }).then(function (res) {
      if (res.status === 204) { return boot(); }
      return res.text().then(function (text) {
        var msg = 'Mot de passe incorrect.';
        if (text) { try { var j = JSON.parse(text); if (j && j.error) msg = j.error; } catch (er) {} }
        errEl.textContent = msg; errEl.classList.remove('hidden');
      });
    }).catch(function () {
      errEl.textContent = 'Erreur réseau.'; errEl.classList.remove('hidden');
    });
  }

  // ---------------------------------------------------------------------
  // Temps réel (SSE)
  // ---------------------------------------------------------------------

  function onTaskEvent(task) { upsertTask(task); render(); }

  function onMessageEvent(m) {
    if (state.messagesByTask[m.taskId]) {
      var exists = state.messagesByTask[m.taskId].some(function (x) { return x.id === m.id; });
      if (!exists) state.messagesByTask[m.taskId].push(m);
    }
    if (state.taskId === m.taskId && state.panelTab === 'chat') {
      renderMain();
      scrollConversationToBottom();
    }
  }

  function onActivityEvent(payload) {
    var t = state.tasksById[payload.taskId];
    if (t) t.liveActivity = payload.line;
    var visibleInList = t && ((state.cardId && t.cardId === state.cardId) || (state.screen === 'inbox' && (t.unread || t.status === 'review')));
    var visibleInPanel = state.taskId === payload.taskId;
    if (visibleInList || visibleInPanel) renderMain();
  }

  function onTokensEvent(payload) {
    if (payload.global) state.tokens.global = payload.global;
    if (payload.projects) {
      Object.keys(payload.projects).forEach(function (pid) {
        if (state.projectsById[pid]) state.projectsById[pid].tokens = payload.projects[pid];
      });
    }
    if (payload.tasks) {
      Object.keys(payload.tasks).forEach(function (tid) {
        if (state.tasksById[tid]) state.tasksById[tid].tokens = payload.tasks[tid];
      });
    }
    var footer = document.getElementById('sidebar-tokens');
    if (footer) footer.textContent = tokenSummary(state.tokens.global);
    var kanbanTok = document.getElementById('kanban-token-stat');
    if (kanbanTok && state.projectId && state.projectsById[state.projectId]) kanbanTok.textContent = tokenSummary(state.projectsById[state.projectId].tokens);
    var detailTok = document.getElementById('detail-token-line');
    if (detailTok && state.taskId && state.tasksById[state.taskId]) detailTok.textContent = tokenSummary(state.tasksById[state.taskId].tokens);
  }

  function onCardsEvent(list) {
    if (!Array.isArray(list) || list.length === 0) { renderSidebar(); renderMain(); return; }
    var pid = list[0].projectId;
    state.cards = state.cards.filter(function (c) { return c.projectId !== pid; }).concat(list);
    reindex();
    renderSidebar(); renderMain();
  }

  function onAgentsEvent(list) {
    state.agents = list || [];
    reindex();
    renderSidebar();
  }

  function fetchStateSilently() {
    return api('/api/state').then(function (data) {
      if (data) { hydrateState(data); render(); }
    }).catch(function () {});
  }

  function connectSSE() {
    var es = new EventSource('/api/events');
    es.addEventListener('task', function (e) { try { onTaskEvent(JSON.parse(e.data)); } catch (er) {} });
    es.addEventListener('message', function (e) { try { onMessageEvent(JSON.parse(e.data)); } catch (er) {} });
    es.addEventListener('activity', function (e) { try { onActivityEvent(JSON.parse(e.data)); } catch (er) {} });
    es.addEventListener('tokens', function (e) { try { onTokensEvent(JSON.parse(e.data)); } catch (er) {} });
    es.addEventListener('cards', function (e) { try { onCardsEvent(JSON.parse(e.data)); } catch (er) {} });
    es.addEventListener('agents', function (e) { try { onAgentsEvent(JSON.parse(e.data)); } catch (er) {} });
    es.onopen = function () {
      if (sseOpenedOnce) fetchStateSilently();
      sseOpenedOnce = true;
    };
  }

  // ---------------------------------------------------------------------
  // Amorçage
  // ---------------------------------------------------------------------

  function boot() {
    return api('/api/state').then(function (data) {
      if (!data) return;
      hydrateState(data);
      hideLogin();
      render();
      connectSSE();
    }).catch(function () { /* showLogin déjà déclenché par api() sur 401 */ });
  }

  // ---------------------------------------------------------------------
  // Délégation d'événements globale
  // ---------------------------------------------------------------------

  function onGlobalClick(e) {
    var el = e.target.closest('[data-action]');
    if (!el) { closeAllCardMenus(); return; }
    var action = el.getAttribute('data-action');
    if (action !== 'toggle-card-menu') closeAllCardMenus();
    switch (action) {
      case 'nav-inbox': goInbox(); break;
      case 'nav-projects': goAllProjects(); break;
      case 'nav-project': goProject(el.getAttribute('data-project-id')); break;
      case 'open-card': goCard(el.getAttribute('data-card-id')); break;
      case 'go-back': goBack(); break;
      case 'open-task': openTask(el.getAttribute('data-task-id')); break;
      case 'close-panel': closePanel(); break;
      case 'set-filter': setFilter(el.getAttribute('data-filter')); break;
      case 'set-tab': setTab(el.getAttribute('data-panel-tab')); break;
      case 'toggle-card-menu': toggleCardMenu(el.getAttribute('data-card-id')); break;
      case 'move-card': moveCard(el.getAttribute('data-card-id'), el.getAttribute('data-column')); break;
      case 'open-new-card': openNewCardModal(el.getAttribute('data-column')); break;
      case 'open-new-task': if (state.cardId) openNewTaskModal(state.cardId); break;
      case 'open-new-project': openNewProjectModal(); break;
      case 'close-modal': closeModal(); break;
      case 'pick-agent': pickAgentInModal(el.getAttribute('data-agent-id')); break;
      case 'pick-column': pickColumnInModal(el.getAttribute('data-column')); break;
      case 'submit-new-task': submitNewTask(el.getAttribute('data-card-id')); break;
      case 'submit-new-project': submitNewProject(); break;
      case 'submit-new-card': submitNewCard(); break;
      case 'interrupt': doInterrupt(el.getAttribute('data-task-id')); break;
      case 'accept': doAccept(el.getAttribute('data-task-id')); break;
      case 'ship-click': handleShipClick(el.getAttribute('data-task-id')); break;
      case 'reopen': doReopen(el.getAttribute('data-task-id')); break;
      case 'select-diff-file': selectDiffFile(el.getAttribute('data-task-id'), el.getAttribute('data-path')); break;
      case 'send-message': sendMessage(el.getAttribute('data-task-id')); break;
      case 'open-search': openSearch(); break;
      case 'search-goto-project': closeSearch(); goProject(el.getAttribute('data-project-id')); break;
      case 'search-goto-task': closeSearch(); openTaskFromSearch(el.getAttribute('data-task-id')); break;
      case 'logout': doLogout(); break;
      case 'toggle-sidebar': toggleSidebarDrawer(); break;
      case 'close-sidebar-drawer': closeSidebarDrawer(); break;
    }
  }

  function onGlobalKeydown(e) {
    var mod = e.ctrlKey || e.metaKey;
    if (mod && (e.key === 'k' || e.key === 'K')) {
      e.preventDefault();
      if (state.searchOpen) closeSearch(); else openSearch();
      return;
    }
    if (e.key === 'Escape') {
      if (state.searchOpen) { closeSearch(); return; }
      if (state.modal) { closeModal(); return; }
      if (state.sidebarOpen) { closeSidebarDrawer(); return; }
    }
    if (e.target && e.target.id === 'composer-input' && e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      var wrap = e.target.closest('.composer');
      var btn = wrap ? wrap.querySelector('[data-action="send-message"]') : null;
      if (btn) sendMessage(btn.getAttribute('data-task-id'));
    }
  }

  document.addEventListener('DOMContentLoaded', function () {
    var form = document.getElementById('login-form');
    if (form) form.addEventListener('submit', onLoginSubmit);
    document.addEventListener('click', onGlobalClick);
    document.addEventListener('keydown', onGlobalKeydown);
    boot();
  });
})();
