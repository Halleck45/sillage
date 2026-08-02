// Sillage : frontend SPA vanilla (zéro framework, zéro dépendance).
// Contrat d'API : voir docs/SPEC-API.md + docs/SPEC-V2.md à la racine du dépôt.
(function () {
  'use strict';

  // ---------------------------------------------------------------------
  // i18n
  // ---------------------------------------------------------------------

  var I18N = {
    fr: {
      'nav.inbox': 'Boîte de réception',
      'nav.allProjects': 'Tous les projets',
      'nav.logout': 'Déconnexion',
      'nav.back': 'Retour',
      'common.projects': 'Projets',
      'common.tasksWord': 'Tâches',
      'common.close': 'Fermer',
      'common.cancel': 'Annuler',
      'common.save': 'Enregistrer',
      'common.create': 'Créer',
      'common.loading': 'Chargement…',
      'common.networkError': 'Erreur réseau.',
      'search.buttonLabel': 'Rechercher',
      'search.placeholder': 'Rechercher tâches et projets…',
      'search.typeToSearch': 'Tapez pour rechercher…',
      'search.noResults': 'Aucun résultat.',
      'sidebar.projectsHeading': 'Projets',
      'sidebar.newProjectTooltip': 'Nouveau projet',
      'sidebar.noProjects': 'Aucun projet',
      'sidebar.agentsHeading': 'Agents',
      'sidebar.newAgentTooltip': 'Nouvel agent',
      'sidebar.noAgents': 'Aucun agent',
      'aria.menu': 'Menu',
      'header.newTask': 'Nouvelle tâche',
      'header.newProject': '+ Nouveau projet',
      'allProjects.emptyTitle': 'Aucun projet pour l\'instant',
      'allProjects.emptySub': 'Créez votre premier projet pour commencer.',
      'allProjects.cardCount.one': '{n} carte',
      'allProjects.cardCount.other': '{n} cartes',
      'inbox.empty': 'Boîte de réception vide. Tout est à jour !',
      'kanban.addCard': '+ Ajouter',
      'kanban.empty': 'Aucune carte pour l\'instant.',
      'kanban.emptyAction': 'Créer la première carte',
      'kanban.card.tasksLabel': 'tâches',
      'kanban.card.reviewCount': '{n} à relire',
      'column.soon': 'Bientôt',
      'column.doing': 'En cours',
      'column.done': 'Terminé',
      'cardMenu.moveTo': 'Déplacer vers {column}',
      'project.editTooltip': 'Modifier le projet',
      'project.editTitle': 'Projet',
      'project.name': 'Nom',
      'project.checkCmd': 'Commande de vérification',
      'project.reposLabel': 'Dépôts',
      'project.repoNamePlaceholder': 'Nom du dépôt',
      'project.addRepo': '+ dépôt',
      'project.removeRepo': 'Retirer',
      'project.errorNameRequired': 'Le nom est requis.',
      'project.errorReposRequired': 'Au moins un dépôt est requis.',
      'project.errorSaveFailed': 'Erreur lors de l\'enregistrement.',
      'project.description': 'Description',
      'project.descriptionPlaceholder': 'Une phrase pour situer le projet',
      'project.contextPrompt': 'Contexte pour les agents',
      'project.contextPromptPlaceholder': 'Conventions, architecture, contraintes à connaître…',
      'work.emptyCard': 'Aucune tâche pour l\'instant.',
      'work.emptyCardAction': 'Créer la première tâche',
      'work.emptyFiltered': 'Aucune tâche ne correspond à ce filtre.',
      'filter.all': 'Toutes {n}',
      'filter.review': 'À relire {n}',
      'filter.ready': 'Prêt à livrer {n}',
      'filter.finished': 'Terminées {n}',
      'badge.new': 'NOUVEAU',
      'workflow.step.review': 'À relire',
      'workflow.step.accepted': 'Accepté',
      'workflow.step.shipped': 'Livré',
      'workflow.doneBanner': 'Tâche terminée',
      'workflow.cancelledBanner': 'Tâche annulée',
      'action.interrupt': 'Interrompre l\'agent',
      'action.accept': 'Accepter le diff et les livrables',
      'action.ship': 'Pousser et livrer',
      'action.shipConfirm': 'Confirmer le push ?',
      'action.reopen': 'Rouvrir la tâche',
      'action.pr': 'Ouvrir la PR',
      'action.prConfirm': 'Confirmer la PR ?',
      'action.finish': 'Marquer comme terminé',
      'action.cancelTask': 'Annuler la tâche',
      'action.cancelTaskConfirm': 'Confirmer l\'annulation ?',
      'tabs.conversation': 'Conversation',
      'tabs.diff': 'Diff',
      'tabs.deliverables': 'Livrables',
      'chat.you': 'Vous',
      'chat.placeholder': 'Répondre à {name}…',
      'chat.send': 'Envoyer ⏎',
      'conversation.empty': 'Aucun message pour l\'instant.',
      'diff.empty': 'Aucune modification.',
      'detail.diff.pushButton': 'Push',
      'detail.diff.pushDisabledTooltip': 'Disponible une fois prêt à livrer',
      'detail.diff.prDisabledTooltip': 'Disponible une fois livré',
      'deliverables.code': 'Code',
      'deliverables.docs': 'Documents',
      'deliverables.images': 'Captures',
      'deliverables.empty': 'Aucun élément.',
      'newTask.title': 'Nouvelle tâche',
      'newTask.titlePlaceholder': 'Que doit faire l\'agent ?',
      'newTask.promptPlaceholder': 'Description ou instructions détaillées (optionnel)',
      'newTask.agentLabel': 'Agent',
      'newTask.repoLabel': 'Dépôt',
      'newTask.projectContextNote': '+ contexte du projet',
      'newTask.hint': 'La conversation s\'ouvre après la création',
      'newTask.submit': 'Créer et discuter',
      'newTask.errorTitleRequired': 'Le titre est requis.',
      'newTask.errorAgentRequired': 'Choisissez un agent.',
      'newTask.errorCreateFailed': 'Erreur lors de la création.',
      'newProject.title': 'Nouveau projet',
      'newProject.nameLabel': 'Nom',
      'newProject.namePlaceholder': 'mon-projet',
      'newProject.pathPlaceholder': '/home/utilisateur/projets/mon-projet',
      'newProject.errorRequired': 'Nom et au moins un dépôt sont requis.',
      'newProject.errorCreateFailed': 'Erreur lors de la création du projet.',
      'newCard.title': 'Nouvelle carte',
      'newCard.titleLabel': 'Titre',
      'newCard.titlePlaceholder': 'Titre de la carte',
      'newCard.errorTitleRequired': 'Le titre est requis.',
      'newCard.errorCreateFailed': 'Erreur lors de la création de la carte.',
      'agent.newTitle': 'Nouvel agent',
      'agent.editTitle': 'Modifier l\'agent',
      'agent.name': 'Nom',
      'agent.emoji': 'Emoji',
      'agent.color': 'Couleur',
      'agent.cli': 'CLI',
      'agent.model': 'Modèle',
      'agent.contextPrompt': 'Prompt de contexte',
      'agent.delete': 'Supprimer',
      'agent.deleteConfirm': 'Confirmer la suppression ?',
      'agent.errorNameRequired': 'Le nom est requis.',
      'agent.errorSaveFailed': 'Erreur lors de l\'enregistrement.',
      'agent.errorDeleteFailed': 'Erreur lors de la suppression.',
      'login.passwordPlaceholder': 'Mot de passe',
      'login.submit': 'Se connecter',
      'login.error': 'Mot de passe incorrect.',
      'time.now': 'à l\'instant',
      'time.min': 'il y a {n} min',
      'time.hour': 'il y a {n} h',
      'time.yesterday': 'hier',
      'time.day': 'il y a {n} j',
      'time.week': 'il y a {n} sem.',
      'time.month': 'il y a {n} mois',
      'time.year.one': 'il y a {n} an',
      'time.year.other': 'il y a {n} ans',
      'tokens.unit': 'tokens',
      'errors.interruptFailed': 'Échec de l\'interruption.',
      'errors.genericFailed': 'Échec.',
      'errors.shipFailed': 'Échec du push.',
      'errors.prFailed': 'Échec de l\'ouverture de la PR.',
      'errors.finishFailed': 'Échec.',
      'errors.cancelFailed': 'Échec de l\'annulation.',
      'onboarding.title': 'Bienvenue dans Sillage',
      'onboarding.intro': 'Choisissez comment sauvegarder votre espace de travail.',
      'onboarding.local.title': 'Travailler en local',
      'onboarding.local.desc': 'Aucune sauvegarde distante. Vous pourrez l\'activer plus tard.',
      'onboarding.local.submit': 'Confirmer',
      'onboarding.init.title': 'Initialiser un dépôt de sauvegarde',
      'onboarding.init.desc': 'Un dépôt git local, avec un remote optionnel.',
      'onboarding.init.submit': 'Initialiser',
      'onboarding.clone.title': 'Rapatrier un espace existant',
      'onboarding.clone.desc': 'Cloner un espace de travail Sillage déjà sauvegardé.',
      'onboarding.clone.warning': 'Le mot de passe de connexion deviendra celui de l\'espace rapatrié.',
      'onboarding.clone.submit': 'Rapatrier',
      'onboarding.errorRemoteRequired': 'L\'URL du dépôt est requise.',
      'onboarding.errorFailed': 'Échec de la configuration.',
      'workspace.tooltip': 'Espace de travail',
      'workspace.title': 'Espace de travail',
      'workspace.state.local': 'Local seul, pas de sauvegarde git',
      'workspace.state.gitNoRemote': 'Git activé, sans remote',
      'workspace.state.gitRemote': 'Git activé, avec remote',
      'workspace.remoteLabel': 'Remote',
      'workspace.privateWarning': 'Dépôt privé recommandé : l\'espace contient vos conversations.',
      'workspace.sync': 'Synchroniser',
      'workspace.syncConfirm': 'Confirmer la synchronisation ?',
      'workspace.activate': 'Activer la sauvegarde git',
      'workspace.lastCommit': 'Dernier commit : {time}',
      'workspace.lastSync': 'Dernière sync : {time}',
      'workspace.never': 'jamais',
      'workspace.dirtyNote': 'Modifications en attente',
      'workspace.syncedJustNow': 'Synchronisé il y a un instant',
      'workspace.errorSaveFailed': 'Échec de l\'enregistrement du remote.',
      'workspace.errorActivateFailed': 'Échec de l\'activation.',
      'workspace.errorSyncFailed': 'Échec de la synchronisation.',
      'preferences.title': 'Préférences',
      'preferences.displayNamePlaceholder': 'Prénom',
      'preferences.langLabel': 'Langue',
      'preferences.errorSaveFailed': 'Échec de l\'enregistrement.'
    },
    en: {
      'nav.inbox': 'Inbox',
      'nav.allProjects': 'All projects',
      'nav.logout': 'Log out',
      'nav.back': 'Back',
      'common.projects': 'Projects',
      'common.tasksWord': 'Tasks',
      'common.close': 'Close',
      'common.cancel': 'Cancel',
      'common.save': 'Save',
      'common.create': 'Create',
      'common.loading': 'Loading…',
      'common.networkError': 'Network error.',
      'search.buttonLabel': 'Search',
      'search.placeholder': 'Search tasks and projects…',
      'search.typeToSearch': 'Type to search…',
      'search.noResults': 'No results.',
      'sidebar.projectsHeading': 'Projects',
      'sidebar.newProjectTooltip': 'New project',
      'sidebar.noProjects': 'No projects',
      'sidebar.agentsHeading': 'Agents',
      'sidebar.newAgentTooltip': 'New agent',
      'sidebar.noAgents': 'No agents',
      'aria.menu': 'Menu',
      'header.newTask': 'New task',
      'header.newProject': '+ New project',
      'allProjects.emptyTitle': 'No projects yet',
      'allProjects.emptySub': 'Create your first project to get started.',
      'allProjects.cardCount.one': '{n} card',
      'allProjects.cardCount.other': '{n} cards',
      'inbox.empty': 'Inbox is empty. All caught up!',
      'kanban.addCard': '+ Add',
      'kanban.empty': 'No cards yet.',
      'kanban.emptyAction': 'Create the first card',
      'kanban.card.tasksLabel': 'tasks',
      'kanban.card.reviewCount': '{n} to review',
      'column.soon': 'Soon',
      'column.doing': 'In progress',
      'column.done': 'Done',
      'cardMenu.moveTo': 'Move to {column}',
      'project.editTooltip': 'Edit project',
      'project.editTitle': 'Project',
      'project.name': 'Name',
      'project.checkCmd': 'Check command',
      'project.reposLabel': 'Repositories',
      'project.repoNamePlaceholder': 'Repository name',
      'project.addRepo': '+ repo',
      'project.removeRepo': 'Remove',
      'project.errorNameRequired': 'Name is required.',
      'project.errorReposRequired': 'At least one repository is required.',
      'project.errorSaveFailed': 'Failed to save.',
      'project.description': 'Description',
      'project.descriptionPlaceholder': 'A sentence describing the project',
      'project.contextPrompt': 'Context for agents',
      'project.contextPromptPlaceholder': 'Conventions, architecture, constraints to know…',
      'work.emptyCard': 'No tasks yet.',
      'work.emptyCardAction': 'Create the first task',
      'work.emptyFiltered': 'No tasks match this filter.',
      'filter.all': 'All {n}',
      'filter.review': 'To review {n}',
      'filter.ready': 'Ready to ship {n}',
      'filter.finished': 'Completed {n}',
      'badge.new': 'NEW',
      'workflow.step.review': 'To review',
      'workflow.step.accepted': 'Accepted',
      'workflow.step.shipped': 'Shipped',
      'workflow.doneBanner': 'Task completed',
      'workflow.cancelledBanner': 'Task cancelled',
      'action.interrupt': 'Stop the agent',
      'action.accept': 'Accept the diff and deliverables',
      'action.ship': 'Push and ship',
      'action.shipConfirm': 'Confirm push?',
      'action.reopen': 'Reopen the task',
      'action.pr': 'Open the PR',
      'action.prConfirm': 'Confirm PR?',
      'action.finish': 'Mark as completed',
      'action.cancelTask': 'Cancel the task',
      'action.cancelTaskConfirm': 'Confirm cancellation?',
      'tabs.conversation': 'Conversation',
      'tabs.diff': 'Diff',
      'tabs.deliverables': 'Deliverables',
      'chat.you': 'You',
      'chat.placeholder': 'Reply to {name}…',
      'chat.send': 'Send ⏎',
      'conversation.empty': 'No messages yet.',
      'diff.empty': 'No changes.',
      'detail.diff.pushButton': 'Push',
      'detail.diff.pushDisabledTooltip': 'Available once ready to ship',
      'detail.diff.prDisabledTooltip': 'Available once shipped',
      'deliverables.code': 'Code',
      'deliverables.docs': 'Documents',
      'deliverables.images': 'Screenshots',
      'deliverables.empty': 'No items.',
      'newTask.title': 'New task',
      'newTask.titlePlaceholder': 'What should the agent do?',
      'newTask.promptPlaceholder': 'Description or detailed instructions (optional)',
      'newTask.agentLabel': 'Agent',
      'newTask.repoLabel': 'Repository',
      'newTask.projectContextNote': '+ project context',
      'newTask.hint': 'The conversation opens after creation',
      'newTask.submit': 'Create and chat',
      'newTask.errorTitleRequired': 'A title is required.',
      'newTask.errorAgentRequired': 'Choose an agent.',
      'newTask.errorCreateFailed': 'Failed to create.',
      'newProject.title': 'New project',
      'newProject.nameLabel': 'Name',
      'newProject.namePlaceholder': 'my-project',
      'newProject.pathPlaceholder': '/home/user/projects/my-project',
      'newProject.errorRequired': 'Name and at least one repository are required.',
      'newProject.errorCreateFailed': 'Failed to create the project.',
      'newCard.title': 'New card',
      'newCard.titleLabel': 'Title',
      'newCard.titlePlaceholder': 'Card title',
      'newCard.errorTitleRequired': 'A title is required.',
      'newCard.errorCreateFailed': 'Failed to create the card.',
      'agent.newTitle': 'New agent',
      'agent.editTitle': 'Edit agent',
      'agent.name': 'Name',
      'agent.emoji': 'Emoji',
      'agent.color': 'Color',
      'agent.cli': 'CLI',
      'agent.model': 'Model',
      'agent.contextPrompt': 'Context prompt',
      'agent.delete': 'Delete',
      'agent.deleteConfirm': 'Confirm deletion?',
      'agent.errorNameRequired': 'Name is required.',
      'agent.errorSaveFailed': 'Failed to save.',
      'agent.errorDeleteFailed': 'Failed to delete.',
      'login.passwordPlaceholder': 'Password',
      'login.submit': 'Log in',
      'login.error': 'Incorrect password.',
      'time.now': 'just now',
      'time.min': '{n} min ago',
      'time.hour': '{n} h ago',
      'time.yesterday': 'yesterday',
      'time.day': '{n} d ago',
      'time.week': '{n} wk ago',
      'time.month': '{n} mo ago',
      'time.year.one': '{n} yr ago',
      'time.year.other': '{n} yrs ago',
      'tokens.unit': 'tokens',
      'errors.interruptFailed': 'Failed to stop the agent.',
      'errors.genericFailed': 'Failed.',
      'errors.shipFailed': 'Failed to push.',
      'errors.prFailed': 'Failed to open the PR.',
      'errors.finishFailed': 'Failed.',
      'errors.cancelFailed': 'Failed to cancel.',
      'onboarding.title': 'Welcome to Sillage',
      'onboarding.intro': 'Choose how to back up your workspace.',
      'onboarding.local.title': 'Work locally',
      'onboarding.local.desc': 'No remote backup. You can enable it later.',
      'onboarding.local.submit': 'Confirm',
      'onboarding.init.title': 'Initialize a backup repository',
      'onboarding.init.desc': 'A local git repository, with an optional remote.',
      'onboarding.init.submit': 'Initialize',
      'onboarding.clone.title': 'Restore an existing workspace',
      'onboarding.clone.desc': 'Clone an already backed-up Sillage workspace.',
      'onboarding.clone.warning': 'The login password will become that of the restored workspace.',
      'onboarding.clone.submit': 'Restore',
      'onboarding.errorRemoteRequired': 'The repository URL is required.',
      'onboarding.errorFailed': 'Setup failed.',
      'workspace.tooltip': 'Workspace',
      'workspace.title': 'Workspace',
      'workspace.state.local': 'Local only, no git backup',
      'workspace.state.gitNoRemote': 'Git enabled, no remote',
      'workspace.state.gitRemote': 'Git enabled, with remote',
      'workspace.remoteLabel': 'Remote',
      'workspace.privateWarning': 'Private repository recommended: the workspace contains your conversations.',
      'workspace.sync': 'Sync',
      'workspace.syncConfirm': 'Confirm sync?',
      'workspace.activate': 'Enable git backup',
      'workspace.lastCommit': 'Last commit: {time}',
      'workspace.lastSync': 'Last sync: {time}',
      'workspace.never': 'never',
      'workspace.dirtyNote': 'Pending changes',
      'workspace.syncedJustNow': 'Synced just now',
      'workspace.errorSaveFailed': 'Failed to save the remote.',
      'workspace.errorActivateFailed': 'Failed to activate.',
      'workspace.errorSyncFailed': 'Failed to sync.',
      'preferences.title': 'Preferences',
      'preferences.displayNamePlaceholder': 'First name',
      'preferences.langLabel': 'Language',
      'preferences.errorSaveFailed': 'Failed to save.'
    }
  };

  function t(key, vars) {
    var dict = I18N[state.lang] || I18N.fr;
    var str = dict[key];
    if (str === undefined) str = (I18N.fr[key] !== undefined ? I18N.fr[key] : key);
    if (vars) {
      Object.keys(vars).forEach(function (k) {
        str = str.replace(new RegExp('\\{' + k + '\\}', 'g'), vars[k]);
      });
    }
    return str;
  }

  function tCount(baseKey, n, vars) {
    var suffix = (n === 1) ? '.one' : '.other';
    var allVars = Object.assign({ n: n }, vars || {});
    return t(baseKey + suffix, allVars);
  }

  function detectInitialLang() {
    var saved = null;
    try { saved = localStorage.getItem('sillage.lang'); } catch (e) {}
    if (saved === 'fr' || saved === 'en') return saved;
    var nav = ((navigator.language || navigator.userLanguage || '') + '').toLowerCase();
    return nav.indexOf('fr') === 0 ? 'fr' : 'en';
  }

  // ---------------------------------------------------------------------
  // Constantes
  // ---------------------------------------------------------------------

  var STATUS_GLYPH = {
    running: { icon: '◐', color: '#8b8982' },
    review: { icon: '◍', color: '#9a6b0d' },
    ready: { icon: '⬆', color: '#2f5fb0' },
    shipped: { icon: '✓', color: '#2f7d54' },
    done: { icon: '✓', color: '#2f7d54' },
    cancelled: { icon: '⊘', color: '#8b8982' }
  };
  var COLUMN_ORDER = ['soon', 'doing', 'done'];
  function columnLabel(key) { return t('column.' + key); }

  // ---------------------------------------------------------------------
  // État en mémoire
  // ---------------------------------------------------------------------

  var state = {
    lang: detectInitialLang(),
    workspace: null,
    settings: { displayName: '', lang: '' },
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
    pendingConfirm: null, // { key, timer }
    sidebarOpen: false // tiroir mobile (< 860px)
  };

  var modalAgentId = null;
  var modalRepos = []; // [{name, path}] pour la modale projet (création/édition)
  var modalRepoCreateMode = true;
  var onboardingExpanded = null;
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

  function localeDecimalSep() { return state.lang === 'en' ? '.' : ','; }

  function formatTokens(n) {
    n = n || 0;
    if (n >= 1000) return (n / 1000).toFixed(1).replace('.', localeDecimalSep()) + 'k';
    return String(n);
  }
  function formatCostNumber(n) {
    return (n || 0).toFixed(2).replace('.', localeDecimalSep());
  }
  function formatMoney(n) {
    var num = formatCostNumber(n);
    return state.lang === 'en' ? ('$' + num) : (num + ' $');
  }
  function tokenSummary(tokens) {
    tokens = tokens || {};
    var total = (tokens.input || 0) + (tokens.output || 0);
    return 'Σ ' + formatTokens(total) + ' ' + t('tokens.unit') + ' · ' + formatMoney(tokens.costUsd);
  }

  function timeAgo(iso) {
    if (!iso) return '';
    var then = new Date(iso).getTime();
    if (isNaN(then)) return '';
    var diff = Math.max(0, Math.round((Date.now() - then) / 1000));
    if (diff < 45) return t('time.now');
    var min = Math.round(diff / 60);
    if (min < 60) return t('time.min', { n: min });
    var hr = Math.round(min / 60);
    if (hr < 24) return t('time.hour', { n: hr });
    var day = Math.round(hr / 24);
    if (day === 1) return t('time.yesterday');
    if (day < 7) return t('time.day', { n: day });
    var week = Math.round(day / 7);
    if (week < 5) return t('time.week', { n: week });
    var month = Math.round(day / 30);
    if (month < 12) return t('time.month', { n: month });
    var year = Math.round(day / 365);
    return tCount('time.year', year);
  }

  function formatTime(iso) {
    var d = new Date(iso);
    if (isNaN(d.getTime())) return '';
    return d.toLocaleTimeString(state.lang === 'en' ? 'en-US' : 'fr-FR', { hour: '2-digit', minute: '2-digit' });
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
    state.workspace = data.workspace || state.workspace || null;
    state.settings = data.settings || state.settings || { displayName: '', lang: '' };
    if (state.settings.lang) {
      state.lang = state.settings.lang;
      try { localStorage.setItem('sillage.lang', state.lang); } catch (e) {}
    }
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
  // Confirmation en deux temps (générique : ship, PR, suppressions)
  // ---------------------------------------------------------------------

  function isPendingConfirm(key) {
    return !!(state.pendingConfirm && state.pendingConfirm.key === key);
  }
  function clearPendingConfirm() {
    if (state.pendingConfirm) clearTimeout(state.pendingConfirm.timer);
    state.pendingConfirm = null;
  }
  function patchConfirmButtons(key) {
    document.querySelectorAll('[data-confirm-key="' + key + '"]').forEach(function (btn) {
      var pending = isPendingConfirm(key);
      btn.textContent = pending ? btn.getAttribute('data-confirm-label') : btn.getAttribute('data-default-label');
    });
  }
  function handleConfirmClick(key, action) {
    if (isPendingConfirm(key)) {
      clearPendingConfirm();
      action();
      return;
    }
    clearPendingConfirm();
    var timer = setTimeout(function () { state.pendingConfirm = null; patchConfirmButtons(key); }, 5000);
    state.pendingConfirm = { key: key, timer: timer };
    patchConfirmButtons(key);
  }
  function handleConfirmClickDispatch(el) {
    var key = el.getAttribute('data-confirm-key');
    var kind = el.getAttribute('data-confirm-action');
    var id = el.getAttribute('data-confirm-id');
    handleConfirmClick(key, function () {
      if (kind === 'ship') doShip(id);
      else if (kind === 'pr') doPr(id);
      else if (kind === 'agent-delete') doDeleteAgent(id);
      else if (kind === 'workspace-sync') doWorkspaceSync();
      else if (kind === 'task-cancel') doCancelTask(id);
    });
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
  // Routage par hash
  // ---------------------------------------------------------------------
  //
  // Formats supportés : #/inbox, #/projects, #/p/{projectId},
  // #/p/{projectId}/c/{cardId}, #/p/{projectId}/c/{cardId}/t/{taskId}
  // (+ ?tab=diff|deliverables optionnel). Le hash est la seule source de
  // vérité pour l'écran affiché ; toute navigation interne appelle
  // navigateTo(), qui met à jour location.hash ; l'écouteur hashchange
  // (et le premier appel après /api/state au boot) appliquent la route
  // via applyRoute().

  function parseHash(hash) {
    hash = String(hash || '').replace(/^#/, '');
    var query = '';
    var qIndex = hash.indexOf('?');
    if (qIndex !== -1) { query = hash.slice(qIndex + 1); hash = hash.slice(0, qIndex); }
    var tab = null;
    if (query) {
      var m = query.match(/(?:^|&)tab=([^&]+)/);
      if (m) tab = decodeURIComponent(m[1]);
    }
    var parts = hash.split('/').filter(function (p) { return p.length > 0; }).map(decodeURIComponent);
    if (parts[0] === 'inbox') return { screen: 'inbox' };
    if (parts[0] === 'projects') return { screen: 'projects' };
    if (parts[0] === 'p' && parts[1]) {
      var projectId = parts[1];
      if (parts[2] === 'c' && parts[3]) {
        var cardId = parts[3];
        if (parts[4] === 't' && parts[5]) {
          return { screen: 'work', projectId: projectId, cardId: cardId, taskId: parts[5], tab: tab };
        }
        return { screen: 'work', projectId: projectId, cardId: cardId };
      }
      return { screen: 'kanban', projectId: projectId };
    }
    return null;
  }

  function navigateTo(hash) {
    if (location.hash === hash) { applyRoute(); return; }
    location.hash = hash;
  }

  function applyRoute() {
    var route = parseHash(location.hash);
    if (!route) { location.hash = '#/projects'; return; }
    if (route.projectId && !state.projectsById[route.projectId]) { location.hash = '#/projects'; return; }
    if (route.cardId) {
      var card = state.cardsById[route.cardId];
      if (!card || card.projectId !== route.projectId) { location.hash = '#/projects'; return; }
    }
    if (route.taskId) {
      var task = state.tasksById[route.taskId];
      if (!task || task.cardId !== route.cardId) { location.hash = '#/projects'; return; }
    }
    var prevCardId = state.cardId;
    var prevTaskId = state.taskId;
    state.screen = route.screen;
    state.projectId = route.projectId || null;
    state.cardId = route.cardId || null;
    state.taskId = route.taskId || null;
    if (state.cardId !== prevCardId) state.taskFilter = 'all';
    if (state.taskId) {
      state.panelTab = route.tab === 'diff' ? 'diff' : (route.tab === 'deliverables' ? 'files' : 'chat');
      if (state.taskId !== prevTaskId) {
        var t = state.tasksById[state.taskId];
        if (t) t.unread = false;
        api('/api/tasks/' + state.taskId + '/read', { method: 'POST' }).catch(function () {});
      }
    }
    closeSidebarDrawer();
    render();
  }

  // ---------------------------------------------------------------------
  // Navigation (construit le hash cible ; l'application de l'état se fait
  // dans applyRoute(), déclenché par le hashchange que provoque navigateTo)
  // ---------------------------------------------------------------------

  function goInbox() { closeSidebarDrawer(); navigateTo('#/inbox'); }
  function goAllProjects() { closeSidebarDrawer(); navigateTo('#/projects'); }
  function goProject(id) { closeSidebarDrawer(); navigateTo('#/p/' + encodeURIComponent(id)); }
  function goCard(id) {
    var c = state.cardsById[id];
    if (!c) return;
    navigateTo('#/p/' + encodeURIComponent(c.projectId) + '/c/' + encodeURIComponent(id));
  }
  function goBack() { if (state.projectId) goProject(state.projectId); else goAllProjects(); }
  function closePanel() {
    if (state.cardId) navigateTo('#/p/' + encodeURIComponent(state.projectId) + '/c/' + encodeURIComponent(state.cardId));
    else goInbox();
  }
  function setFilter(f) { state.taskFilter = f; renderMain(); }
  function setTab(tabKey) {
    if (!state.taskId) return;
    var h = '#/p/' + encodeURIComponent(state.projectId) + '/c/' + encodeURIComponent(state.cardId) + '/t/' + encodeURIComponent(state.taskId);
    if (tabKey === 'diff') h += '?tab=diff';
    else if (tabKey === 'files') h += '?tab=deliverables';
    navigateTo(h);
  }

  function render() { renderSidebar(); renderMain(); }

  function setLang(lang) {
    if (lang !== 'fr' && lang !== 'en') return;
    state.lang = lang;
    state.settings = state.settings || { displayName: '', lang: '' };
    state.settings.lang = lang;
    try { localStorage.setItem('sillage.lang', lang); } catch (e) {}
    api('/api/settings', { method: 'PATCH', body: { lang: lang } }).catch(function () {});
    closeSearch();
    applyStaticTranslations();
    render();
    refreshWorkspaceModalBody();
  }

  function openTask(taskId) {
    var task = state.tasksById[taskId];
    if (!task) return;
    navigateTo('#/p/' + encodeURIComponent(task.projectId) + '/c/' + encodeURIComponent(task.cardId) + '/t/' + encodeURIComponent(taskId));
  }

  function openTaskFromSearch(taskId) {
    openTask(taskId);
  }

  // ---------------------------------------------------------------------
  // Rendu : sidebar
  // ---------------------------------------------------------------------

  function buildLangSwitchHTML() {
    return '<div class="lang-switch">' +
      '<button class="lang-btn ' + (state.lang === 'fr' ? 'active' : '') + '" data-action="set-lang" data-lang="fr">FR</button>' +
      '<span class="lang-sep">·</span>' +
      '<button class="lang-btn ' + (state.lang === 'en' ? 'active' : '') + '" data-action="set-lang" data-lang="en">EN</button>' +
      '</div>';
  }

  function buildSidebarHTML() {
    var navItems = [
      { key: 'inbox', icon: '⌂', label: t('nav.inbox'), action: 'nav-inbox' },
      { key: 'projects', icon: '◫', label: t('nav.allProjects'), action: 'nav-projects' }
    ];
    var navHTML = navItems.map(function (n) {
      var active = state.screen === n.key;
      return '<button class="nav-item ' + (active ? 'active' : '') + '" data-action="' + n.action + '">' +
        '<span class="nav-icon">' + n.icon + '</span>' + escapeHtml(n.label) + '</button>';
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
      return '<div class="agent-item" data-action="edit-agent" data-agent-id="' + a.id + '">' +
        '<span class="agent-avatar" style="background:' + softColor(a.color) + '">' + a.emoji + '</span>' +
        '<span class="agent-name">' + escapeHtml(a.name) + '</span>' +
        (a.active ? '<span class="agent-dot"></span>' : '') +
        '</div>';
    }).join('');

    return '' +
      '<div class="sidebar-brand"><span class="brand-mark"></span><span class="brand-name">Sillage</span></div>' +
      '<div class="sidebar-search-wrap">' +
        '<button class="search-btn" data-action="open-search"><span>⌕</span><span class="search-btn-label">' + escapeHtml(t('search.buttonLabel')) + '</span>' +
        '<span class="kbd">' + (isMac() ? '⌘K' : 'Ctrl+K') + '</span></button>' +
      '</div>' +
      '<div class="nav-list">' + navHTML + '</div>' +
      '<div class="sidebar-section-head"><span>' + escapeHtml(t('sidebar.projectsHeading')) + '</span>' +
      '<button class="icon-btn-sm" data-action="open-new-project" title="' + escapeHtml(t('sidebar.newProjectTooltip')) + '" aria-label="' + escapeHtml(t('sidebar.newProjectTooltip')) + '">+</button></div>' +
      '<nav class="project-list">' + (projectsHTML || '<div class="empty-note-sm">' + escapeHtml(t('sidebar.noProjects')) + '</div>') + '</nav>' +
      '<div class="sidebar-section-head"><span>' + escapeHtml(t('sidebar.agentsHeading')) + '</span>' +
      '<button class="icon-btn-sm" data-action="open-new-agent" title="' + escapeHtml(t('sidebar.newAgentTooltip')) + '" aria-label="' + escapeHtml(t('sidebar.newAgentTooltip')) + '">+</button></div>' +
      '<div class="agent-list">' + (agentsHTML || '<div class="empty-note-sm">' + escapeHtml(t('sidebar.noAgents')) + '</div>') + '</div>' +
      '<div class="sidebar-footer">' +
        '<span class="sidebar-tokens" id="sidebar-tokens">' + tokenSummary(state.tokens.global) + '</span>' +
        '<div class="sidebar-footer-actions">' +
          '<button class="icon-btn-sm" data-action="open-workspace-modal" title="' + escapeHtml(t('workspace.tooltip')) + '" aria-label="' + escapeHtml(t('workspace.tooltip')) + '">⚙</button>' +
          '<button class="logout-link" data-action="logout">' + escapeHtml(t('nav.logout')) + '</button>' +
        '</div>' +
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
      title = t('nav.inbox');
    } else if (state.screen === 'projects') {
      title = t('nav.allProjects');
      actions = '<button class="btn-outline" data-action="open-new-project">' + escapeHtml(t('header.newProject')) + '</button>';
    } else if (state.screen === 'kanban') {
      var p = state.projectsById[state.projectId];
      title = p ? p.name : '';
    } else if (state.screen === 'work') {
      back = '<button class="icon-btn" data-action="go-back" aria-label="' + escapeHtml(t('nav.back')) + '">←</button>';
      var c = state.cardsById[state.cardId];
      var pr = state.projectsById[state.projectId];
      title = c ? c.title : '';
      sub = pr ? '<span class="crumb-sub">' + escapeHtml(pr.name) + '</span>' : '';
      actions = '<button class="btn-primary" data-action="open-new-task">' + escapeHtml(t('header.newTask')) + '</button>';
    }
    var hamburger = '<button class="icon-btn hamburger-btn" data-action="toggle-sidebar" aria-label="' + escapeHtml(t('aria.menu')) + '">☰</button>';
    return '<header class="topbar">' + hamburger + back + '<span class="crumb-title">' + escapeHtml(title) + '</span>' + sub +
      '<div class="topbar-actions">' + actions + '</div></header>';
  }

  // ---------------------------------------------------------------------
  // Rendu : Tous les projets
  // ---------------------------------------------------------------------

  function buildAllProjectsHTML() {
    if (state.projects.length === 0) {
      return buildHeaderHTML() + '<div class="view-body"><div class="empty-state big">' +
        '<div class="empty-title">' + escapeHtml(t('allProjects.emptyTitle')) + '</div>' +
        '<div class="empty-sub">' + escapeHtml(t('allProjects.emptySub')) + '</div>' +
        '<button class="btn-primary" data-action="open-new-project">' + escapeHtml(t('header.newProject')) + '</button>' +
        '</div></div>';
    }
    var tiles = state.projects.map(function (p) {
      var cardCount = state.cards.filter(function (c) { return c.projectId === p.id; }).length;
      var unread = projectUnread(p.id);
      return '<article class="project-tile" data-action="nav-project" data-project-id="' + p.id + '">' +
        '<div class="project-tile-top"><span class="project-hash">#</span><h3>' + escapeHtml(p.name) + '</h3>' +
        (unread ? '<span class="badge-unread">' + unread + '</span>' : '') + '</div>' +
        '<div class="project-tile-meta"><span>' + escapeHtml(tCount('allProjects.cardCount', cardCount)) + '</span>' +
        '<span>' + tokenSummary(p.tokens) + '</span></div>' +
        '</article>';
    }).join('');
    return buildHeaderHTML() + '<div class="view-body all-projects-body"><div class="project-grid">' + tiles +
      '<button class="project-tile project-tile-add" data-action="open-new-project">' + escapeHtml(t('header.newProject')) + '</button>' +
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
    var attention = c.reviewCount ? '<span class="card-attention">' + escapeHtml(t('kanban.card.reviewCount', { n: c.reviewCount })) + '</span>' : '';
    var others = COLUMN_ORDER.filter(function (k) { return k !== colKey; });
    var menuItemsHTML = others.map(function (k) {
      return '<button class="card-menu-item" data-action="move-card" data-card-id="' + c.id + '" data-column="' + k + '">' +
        escapeHtml(t('cardMenu.moveTo', { column: columnLabel(k) })) + '</button>';
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
        '<span>' + (c.tasksDone || 0) + '/' + (c.tasksTotal || 0) + ' ' + escapeHtml(t('kanban.card.tasksLabel')) + '</span>' +
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
    var tokenTxt = project ? tokenSummary(project.tokens) : '';

    var colsHTML = COLUMN_ORDER.map(function (key) {
      var label = columnLabel(key);
      var color = key === 'soon' ? '#d3d0c8' : (key === 'doing' ? '#2f7d54' : '#c3c0b8');
      var list = cards.filter(function (c) { return c.column === key; });
      var addBtn = key === 'soon'
        ? '<button class="add-card-btn" data-action="open-new-card">' + escapeHtml(t('kanban.addCard')) + '</button>'
        : '';
      return '<section class="kanban-col">' +
        '<div class="col-head"><span class="col-dot" style="background:' + color + '"></span>' +
        '<span class="col-label">' + escapeHtml(label) + '</span><span class="col-count">' + list.length + '</span></div>' +
        list.map(function (c) { return buildKanbanCardHTML(c, key); }).join('') +
        addBtn +
        '</section>';
    }).join('');

    var head = '<div class="kanban-head"><div class="kanban-title-col">' +
      '<div class="kanban-title-row"><h1>' + escapeHtml(project ? project.name : '') + '</h1>';
    if (project) {
      head += '<button class="icon-btn" data-action="open-edit-project" title="' + escapeHtml(t('project.editTooltip')) + '" aria-label="' + escapeHtml(t('project.editTooltip')) + '">✎</button>';
    }
    head += '</div>';
    if (project && project.description) {
      head += '<div class="kanban-description">' + escapeHtml(project.description) + '</div>';
    }
    head += '</div>';
    head += '<div class="kanban-stats"><span id="kanban-token-stat" class="token-stat">' + tokenTxt + '</span></div>';
    head += '</div>';

    var emptyNote = cards.length === 0
      ? '<button class="empty-cta" data-action="open-new-card">' +
          '<span class="empty-cta-title">' + escapeHtml(t('kanban.empty')) + '</span>' +
          '<span class="empty-cta-sub">' + escapeHtml(t('kanban.emptyAction')) + '</span>' +
        '</button>'
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
    var titleClass = t.status === 'cancelled' ? 'task-title task-title-cancelled' : 'task-title';
    return '<div class="task-row ' + (selected ? 'selected' : '') + '" data-action="open-task" data-task-id="' + t.id + '">' +
      '<span class="task-glyph" style="color:' + glyph.color + '">' + glyph.icon + '</span>' +
      '<div class="task-main">' +
        '<div class="task-title-line"><span class="' + titleClass + '">' + escapeHtml(t.title) + '</span>' +
        (isNew ? '<span class="badge-new">' + escapeHtml(t2('badge.new')) + '</span>' : '') + '</div>' +
        '<div class="task-meta-line">' + projectTag +
        '<span class="mono">#' + t.ref + '</span>' +
        '<span class="agent-chip"><span class="agent-avatar-sm" style="background:' + softColor(agent.color) + '">' + agent.emoji + '</span>' + escapeHtml(agent.name) + '</span>' +
        '<span>' + escapeHtml(timeAgo(t.updatedAt)) + '</span>' + liveLine +
        '</div>' +
      '</div>' +
      '<div class="task-counts"><span>◆ ' + (t.filesCount || 0) + '</span><span>💬 ' + (t.messagesCount || 0) + '</span></div>' +
      '</div>';
  }
  // Alias local pour éviter un conflit de nom avec le paramètre `t` (Task) ci-dessus.
  function t2(key, vars) { return t(key, vars); }

  function taskMatchesFilter(t, filter) {
    if (filter === 'all') return true;
    if (filter === 'finished') return t.status === 'shipped' || t.status === 'done' || t.status === 'cancelled';
    return t.status === filter;
  }

  function buildTaskListShell(tasksAll, opts) {
    var counts = tasksAll.reduce(function (acc, t) { acc[t.status] = (acc[t.status] || 0) + 1; return acc; }, {});
    var finishedCount = (counts.shipped || 0) + (counts.done || 0) + (counts.cancelled || 0);
    var filters = [
      { key: 'all', label: t('filter.all', { n: tasksAll.length }) },
      { key: 'review', label: t('filter.review', { n: counts.review || 0 }) },
      { key: 'ready', label: t('filter.ready', { n: counts.ready || 0 }) },
      { key: 'finished', label: t('filter.finished', { n: finishedCount }) }
    ];
    var filterHTML = filters.map(function (f) {
      return '<button class="pill ' + (state.taskFilter === f.key ? 'pill-active' : '') + '" data-action="set-filter" data-filter="' + f.key + '">' + escapeHtml(f.label) + '</button>';
    }).join('');
    var visible = tasksAll.filter(function (t) { return taskMatchesFilter(t, state.taskFilter); })
      .sort(function (a, b) { return new Date(b.updatedAt) - new Date(a.updatedAt); });
    var rowsHTML;
    if (visible.length) {
      rowsHTML = visible.map(function (t) { return buildTaskRowHTML(t, opts.showProject); }).join('');
    } else if (tasksAll.length === 0 && opts.emptyCta) {
      rowsHTML = '<button class="empty-cta" data-action="' + opts.emptyCta.action + '">' +
        '<span class="empty-cta-title">' + escapeHtml(opts.emptyMsg) + '</span>' +
        '<span class="empty-cta-sub">' + escapeHtml(opts.emptyCta.label) + '</span>' +
        '</button>';
    } else if (tasksAll.length === 0) {
      rowsHTML = '<div class="empty-state">' + escapeHtml(opts.emptyMsg) + '</div>';
    } else {
      rowsHTML = '<div class="empty-state">' + escapeHtml(t('work.emptyFiltered')) + '</div>';
    }
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
    return buildTaskListShell(tasksAll, {
      showProject: false,
      emptyMsg: t('work.emptyCard'),
      emptyCta: { action: 'open-new-task', label: t('work.emptyCardAction') }
    });
  }

  function buildInboxHTML() {
    var tasksAll = state.tasks.filter(function (t) { return t.unread || t.status === 'review'; });
    return buildTaskListShell(tasksAll, { showProject: true, emptyMsg: t('inbox.empty') });
  }

  // ---------------------------------------------------------------------
  // Rendu : panneau de détail de tâche
  // ---------------------------------------------------------------------

  function primaryActionInfo(task) {
    switch (task.status) {
      case 'running':
        return { label: t('action.interrupt'), cls: 'btn-neutral', action: 'interrupt', kind: 'plain' };
      case 'review':
        return { label: t('action.accept'), cls: 'btn-green', action: 'accept', kind: 'plain' };
      case 'ready': {
        var key = 'ship:' + task.id;
        var pending = isPendingConfirm(key);
        return {
          label: pending ? t('action.shipConfirm') : t('action.ship'),
          cls: 'btn-green', kind: 'confirm', confirmKey: key, confirmAction: 'ship',
          defaultLabel: t('action.ship'), confirmLabel: t('action.shipConfirm')
        };
      }
      case 'shipped':
        return { label: t('action.reopen'), cls: 'btn-neutral', action: 'reopen', kind: 'plain' };
      case 'done':
      case 'cancelled':
        return { label: t('action.reopen'), cls: 'btn-neutral', action: 'reopen', kind: 'plain' };
      default:
        return { label: '', cls: 'btn-neutral', action: '', kind: 'plain' };
    }
  }

  function buildWorkflowHTML(status) {
    if (status === 'done' || status === 'cancelled') {
      var bannerLabel = status === 'done' ? t('workflow.doneBanner') : t('workflow.cancelledBanner');
      return '<div class="workflow-banner">' + escapeHtml(bannerLabel) + '</div>';
    }
    var order = { running: 0, review: 0, ready: 1, shipped: 2 };
    var steps = [
      { label: t('workflow.step.review') },
      { label: t('workflow.step.accepted') },
      { label: t('workflow.step.shipped') }
    ];
    var cur = order[status] !== undefined ? order[status] : 0;
    var html = '<div class="workflow">';
    steps.forEach(function (s, i) {
      var done = i <= cur;
      var isCur = i === cur;
      var barClass = !done ? 'wf-bar-todo' : (isCur ? 'wf-bar-current' : 'wf-bar-done');
      var lblClass = done ? 'wf-label-done' : '';
      html += '<div class="wf-step"><span class="wf-bar ' + barClass + '"></span>' +
        '<span class="wf-label ' + lblClass + ' ' + (isCur ? 'wf-current' : '') + '">' + escapeHtml(s.label) + '</span></div>';
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
    var taskProject = state.projectsById[task.projectId];
    var multiRepo = !!(taskProject && taskProject.repos && taskProject.repos.length > 1);
    var action = primaryActionInfo(task);
    var tabs = ['chat', 'diff', 'files'];
    var tabLabels = { chat: t('tabs.conversation'), diff: t('tabs.diff'), files: t('tabs.deliverables') };
    var tabCounts = { chat: task.messagesCount || 0, diff: task.filesCount || 0, files: (task.docsCount || 0) + (task.filesCount || 0) };
    var tabDataAttr = { chat: 'conversation', diff: 'diff', files: 'deliverables' };

    var tabsHTML = tabs.map(function (tk) {
      var active = state.panelTab === tk;
      var isNew = tk === 'chat' && task.unread && !active;
      return '<button class="tab ' + (active ? 'tab-active' : '') + '" role="tab" data-tab="' + tabDataAttr[tk] + '" data-action="set-tab" data-panel-tab="' + tk + '">' +
        escapeHtml(tabLabels[tk]) + '<span class="tab-count">' + tabCounts[tk] + '</span>' +
        (isNew ? '<span class="tab-dot"></span>' : '') + '</button>';
    }).join('');

    var bodyHTML = '';
    if (state.panelTab === 'chat') bodyHTML = buildConversationHTML(task, agent);
    else if (state.panelTab === 'diff') bodyHTML = buildDiffHTML(task);
    else if (state.panelTab === 'files') bodyHTML = buildDeliverablesHTML(task);

    var primaryBtnHTML;
    if (action.kind === 'confirm') {
      primaryBtnHTML = '<button id="task-primary-action" class="btn-action ' + action.cls + '" data-action="confirm-click" data-confirm-key="' + action.confirmKey + '" data-confirm-action="' + action.confirmAction + '" data-confirm-id="' + task.id + '" data-default-label="' + escapeHtml(action.defaultLabel) + '" data-confirm-label="' + escapeHtml(action.confirmLabel) + '">' + escapeHtml(action.label) + '</button>';
    } else {
      primaryBtnHTML = '<button id="task-primary-action" class="btn-action ' + action.cls + '" data-action="' + action.action + '" data-task-id="' + task.id + '">' + escapeHtml(action.label) + '</button>';
    }

    var showFinishLink = task.status === 'review' || task.status === 'ready' || task.status === 'shipped';
    var showCancelLink = task.status === 'running' || task.status === 'review' || task.status === 'ready';
    var linksRow = '';
    if (showFinishLink || showCancelLink) {
      var cancelKey = 'task-cancel:' + task.id;
      var cancelPending = isPendingConfirm(cancelKey);
      var cancelLabel = cancelPending ? t('action.cancelTaskConfirm') : t('action.cancelTask');
      linksRow = '<div class="detail-link-row">' +
        (showFinishLink ? '<button class="detail-link" data-action="finish-task" data-task-id="' + task.id + '">' + escapeHtml(t('action.finish')) + '</button>' : '') +
        (showCancelLink ? '<button class="detail-link" data-action="confirm-click" data-confirm-key="' + cancelKey + '" data-confirm-action="task-cancel" data-confirm-id="' + task.id + '" data-default-label="' + escapeHtml(t('action.cancelTask')) + '" data-confirm-label="' + escapeHtml(t('action.cancelTaskConfirm')) + '">' + escapeHtml(cancelLabel) + '</button>' : '') +
        '</div>';
    }

    var prRow = '';
    if (task.status === 'shipped') {
      var prKey = 'pr:' + task.id;
      var prPending = isPendingConfirm(prKey);
      var prLabel = prPending ? t('action.prConfirm') : t('action.pr');
      prRow = '<div class="secondary-row">' +
        '<button class="btn-outline btn-block" data-action="confirm-click" data-confirm-key="' + prKey + '" data-confirm-action="pr" data-confirm-id="' + task.id + '" data-default-label="' + escapeHtml(t('action.pr')) + '" data-confirm-label="' + escapeHtml(t('action.prConfirm')) + '">' + escapeHtml(prLabel) + '</button>' +
        '</div>';
    }

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
              (multiRepo && task.repoName ? '<span class="repo-chip">' + escapeHtml(task.repoName) + '</span>' : '') +
            '</div>' +
          '</div>' +
          '<button class="icon-btn" data-action="close-panel" aria-label="' + escapeHtml(t('common.close')) + '">✕</button>' +
        '</div>' +
        '<div class="detail-tokens" id="detail-token-line">' + tokenSummary(task.tokens) + '</div>' +
        buildWorkflowHTML(task.status) +
        '<div class="action-row">' + primaryBtnHTML +
          '<span class="checks">' + renderChecks(task.checks) + '</span>' +
        '</div>' +
        linksRow +
        prRow +
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
    var name = m.authorName || (isUser ? t('chat.you') : (agent.name || ''));
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
      return '<div class="conv-loading">' + escapeHtml(t('common.loading')) + '</div>';
    }
    var items = msgs.length ? msgs.map(function (m) { return buildMessageHTML(m, agent); }).join('') : '<div class="empty-note">' + escapeHtml(t('conversation.empty')) + '</div>';
    return '<div class="conversation" id="conversation-list">' + items + '</div>' +
      '<div class="composer-wrap"><div class="composer">' +
        '<textarea id="composer-input" rows="2" placeholder="' + escapeHtml(t('chat.placeholder', { name: agent.name })) + '"></textarea>' +
        '<div class="composer-row">' +
          '<span class="composer-model"><span class="agent-avatar-sm" style="background:' + softColor(agent.color) + '">' + agent.emoji + '</span>' + escapeHtml(agent.model || '') + '</span>' +
          '<button class="btn-send" data-action="send-message" data-task-id="' + task.id + '">' + escapeHtml(t('chat.send')) + '</button>' +
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
      return '<div class="conv-loading">' + escapeHtml(t('common.loading')) + '</div>';
    }
    if (!diff.files || diff.files.length === 0) {
      return '<div class="empty-note">' + escapeHtml(t('diff.empty')) + '</div>';
    }
    var activePath = state.activeDiffFile[task.id] || diff.files[0].path;
    var activeFile = diff.files.filter(function (f) { return f.path === activePath; })[0] || diff.files[0];
    var fileTabs = diff.files.map(function (f) {
      return '<button class="diff-file-tab ' + (f.path === activeFile.path ? 'active' : '') + '" data-action="select-diff-file" data-task-id="' + task.id + '" data-path="' + escapeHtml(f.path) + '">' +
        escapeHtml(f.path) + ' <span class="diff-add">+' + f.additions + '</span> <span class="diff-del">-' + f.deletions + '</span></button>';
    }).join('');
    var hunks = (activeFile.hunks || []).map(function (h) {
      var lines = (h.lines || []).map(function (l) {
        var mark = l.type === 'add' ? '+' : (l.type === 'del' ? '-' : ' ');
        return '<div class="diff-line diff-' + l.type + '"><span class="diff-mark">' + mark + '</span><span class="diff-text mono">' + escapeHtml(l.text) + '</span></div>';
      }).join('');
      return '<div><div class="diff-hunk-header mono">' + escapeHtml(h.header) + '</div>' + lines + '</div>';
    }).join('');

    var shipKey = 'ship:' + task.id;
    var canShip = task.status === 'ready';
    var shipPending = isPendingConfirm(shipKey);
    var pushDefaultLabel = t('detail.diff.pushButton');
    var pushLabel = (shipPending && canShip) ? t('action.shipConfirm') : pushDefaultLabel;

    var prKey = 'pr:' + task.id;
    var canPr = task.status === 'shipped';
    var prPending = isPendingConfirm(prKey);
    var prDefaultLabel = t('action.pr');
    var prLabel = (prPending && canPr) ? t('action.prConfirm') : prDefaultLabel;

    return '<div class="diff-subtabs">' + fileTabs + '</div>' +
      '<div class="diff-hunks">' + hunks + '</div>' +
      '<div class="diff-footer">' +
        '<span class="mono diff-branch">' + escapeHtml(diff.branch || '') + ' → ' + escapeHtml(diff.base || '') + '</span>' +
        '<button class="btn-outline" ' + (canPr ? '' : ('disabled title="' + escapeHtml(t('detail.diff.prDisabledTooltip')) + '"')) +
        ' data-action="confirm-click" data-confirm-key="' + prKey + '" data-confirm-action="pr" data-confirm-id="' + task.id + '" data-default-label="' + escapeHtml(prDefaultLabel) + '" data-confirm-label="' + escapeHtml(t('action.prConfirm')) + '">' + escapeHtml(prLabel) + '</button>' +
        '<button class="btn-green" ' + (canShip ? '' : ('disabled title="' + escapeHtml(t('detail.diff.pushDisabledTooltip')) + '"')) +
        ' data-action="confirm-click" data-confirm-key="' + shipKey + '" data-confirm-action="ship" data-confirm-id="' + task.id + '" data-default-label="' + escapeHtml(pushDefaultLabel) + '" data-confirm-label="' + escapeHtml(t('action.shipConfirm')) + '">' + escapeHtml(pushLabel) + '</button>' +
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
      return '<div class="conv-loading">' + escapeHtml(t('common.loading')) + '</div>';
    }
    var groups = [
      { key: 'code', label: t('deliverables.code'), icon: '◆' },
      { key: 'docs', label: t('deliverables.docs'), icon: '▤' },
      { key: 'images', label: t('deliverables.images'), icon: '▣' }
    ];
    var html = groups.map(function (g) {
      var items = d[g.key] || [];
      var itemsHTML = items.length ? items.map(function (it) {
        return '<div class="deliv-item"><span class="deliv-icon">' + g.icon + '</span>' +
          '<div class="deliv-main"><div class="deliv-title">' + escapeHtml(it.title) + '</div>' +
          '<div class="deliv-meta mono">' + escapeHtml(it.meta || '') + '</div></div></div>';
      }).join('') : '<div class="empty-note">' + escapeHtml(t('deliverables.empty')) + '</div>';
      return '<section class="deliv-group"><div class="deliv-group-label">' + escapeHtml(g.label) + '</div>' + itemsHTML + '</section>';
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
    }).catch(function (e) { if (e instanceof ApiError) showDetailError(taskId, e.message || t('errors.interruptFailed')); });
  }
  function doAccept(taskId) {
    api('/api/tasks/' + taskId + '/accept', { method: 'POST' }).then(function (task) {
      upsertTask(task); renderMain();
    }).catch(function (e) { if (e instanceof ApiError) showDetailError(taskId, e.message || t('errors.genericFailed')); });
  }
  function doReopen(taskId) {
    api('/api/tasks/' + taskId + '/reopen', { method: 'POST' }).then(function (task) {
      upsertTask(task); renderMain();
    }).catch(function (e) { if (e instanceof ApiError) showDetailError(taskId, e.message || t('errors.genericFailed')); });
  }
  function doShip(taskId) {
    api('/api/tasks/' + taskId + '/ship', { method: 'POST', body: { confirm: true } }).then(function (res) {
      if (res && res.task) upsertTask(res.task);
      renderMain();
    }).catch(function (e) { if (e instanceof ApiError) showDetailError(taskId, e.message || t('errors.shipFailed')); });
  }
  function doPr(taskId) {
    api('/api/tasks/' + taskId + '/pr', { method: 'POST', body: { confirm: true } }).then(function (res) {
      if (res && res.task) upsertTask(res.task);
      if (res && res.url) { try { window.open(res.url, '_blank', 'noopener'); } catch (e) {} }
      renderMain();
    }).catch(function (e) { if (e instanceof ApiError) showDetailError(taskId, e.message || t('errors.prFailed')); });
  }
  function doFinish(taskId) {
    api('/api/tasks/' + taskId + '/finish', { method: 'POST' }).then(function (task) {
      upsertTask(task); renderMain();
    }).catch(function (e) { if (e instanceof ApiError) showDetailError(taskId, e.message || t('errors.finishFailed')); });
  }
  function doCancelTask(taskId) {
    api('/api/tasks/' + taskId + '/cancel', { method: 'POST' }).then(function (task) {
      upsertTask(task); renderMain();
    }).catch(function (e) { if (e instanceof ApiError) showDetailError(taskId, e.message || t('errors.cancelFailed')); });
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
    if (!query) return '<div class="empty-note">' + escapeHtml(t('search.typeToSearch')) + '</div>';
    var projects = state.projects.filter(function (p) { return p.name.toLowerCase().indexOf(query) !== -1; });
    var tasks = state.tasks.filter(function (t) {
      return t.title.toLowerCase().indexOf(query) !== -1 || String(t.ref).indexOf(query) !== -1;
    }).slice(0, 20);
    if (!projects.length && !tasks.length) return '<div class="empty-note">' + escapeHtml(t('search.noResults')) + '</div>';
    var html = '';
    if (projects.length) {
      html += '<div class="search-group-label">' + escapeHtml(t('common.projects')) + '</div>' + projects.map(function (p) {
        return '<button class="search-result" data-action="search-goto-project" data-project-id="' + p.id + '"><span class="hash">#</span>' + escapeHtml(p.name) + '</button>';
      }).join('');
    }
    if (tasks.length) {
      html += '<div class="search-group-label">' + escapeHtml(t('common.tasksWord')) + '</div>' + tasks.map(function (t) {
        var p = state.projectsById[t.projectId];
        return '<button class="search-result" data-action="search-goto-task" data-task-id="' + t.id + '"><span class="mono">#' + t.ref + '</span> ' +
          escapeHtml(t.title) + '<span class="muted-sm">' + (p ? escapeHtml(p.name) : '') + '</span></button>';
      }).join('');
    }
    return html;
  }

  function buildSearchHTML(q) {
    return '<div class="search-box"><input id="search-input" class="search-input" placeholder="' + escapeHtml(t('search.placeholder')) + '" value="' + escapeHtml(q) + '">' +
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
  // Modales génériques
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
    var project = state.projectsById[card.projectId];
    var repos = (project && project.repos) || [];
    var repoSelectHTML = '';
    if (repos.length > 1) {
      var repoOptions = repos.map(function (r) {
        return '<option value="' + escapeHtml(r.name) + '">' + escapeHtml(r.name) + '</option>';
      }).join('');
      repoSelectHTML = '<div class="modal-label">' + escapeHtml(t('newTask.repoLabel')) + '</div>' +
        '<select id="new-task-repo" class="modal-input">' + repoOptions + '</select>';
    }
    return '<div class="modal">' +
      '<div class="modal-head"><span class="modal-title">' + escapeHtml(t('newTask.title')) + '</span><span class="modal-sub">' + escapeHtml(card.title) + '</span>' +
      '<button class="icon-btn" data-action="close-modal" aria-label="' + escapeHtml(t('common.close')) + '">✕</button></div>' +
      '<input id="new-task-title" class="modal-input" placeholder="' + escapeHtml(t('newTask.titlePlaceholder')) + '">' +
      '<textarea id="new-task-prompt" class="modal-textarea" placeholder="' + escapeHtml(t('newTask.promptPlaceholder')) + '" rows="3"></textarea>' +
      repoSelectHTML +
      '<div class="modal-label">' + escapeHtml(t('newTask.agentLabel')) + '</div>' +
      '<div class="agent-choices" id="agent-choices">' + agentChoices + '</div>' +
      '<div class="agent-context-preview" id="agent-context-preview">' + (selected ? escapeHtml(selected.contextPrompt || '') : '') + '</div>' +
      (project && project.contextPrompt ? '<div class="project-context-note">' + escapeHtml(t('newTask.projectContextNote')) + '</div>' : '') +
      '<div id="new-task-error" class="modal-error hidden"></div>' +
      '<div class="modal-foot"><span class="modal-hint">' + escapeHtml(t('newTask.hint')) + '</span>' +
      '<button class="btn-outline" data-action="close-modal">' + escapeHtml(t('common.cancel')) + '</button>' +
      '<button class="btn-green" data-action="submit-new-task" data-card-id="' + card.id + '">' + escapeHtml(t('newTask.submit')) + '</button></div>' +
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
    if (!title) { errEl.textContent = t('newTask.errorTitleRequired'); errEl.classList.remove('hidden'); return; }
    if (!modalAgentId) { errEl.textContent = t('newTask.errorAgentRequired'); errEl.classList.remove('hidden'); return; }
    var body = { cardId: cardId, title: title, agentId: modalAgentId };
    if (prompt) body.prompt = prompt;
    var repoEl = document.getElementById('new-task-repo');
    if (repoEl && repoEl.value) body.repoName = repoEl.value;
    api('/api/tasks', { method: 'POST', body: body }).then(function (task) {
      upsertTask(task);
      closeModal();
      openTask(task.id);
    }).catch(function (e) {
      errEl.textContent = (e instanceof ApiError && e.message) || t('newTask.errorCreateFailed');
      errEl.classList.remove('hidden');
    });
  }

  // Dépôts de projet (lignes éditables, réutilisées par les modales
  // Nouveau projet et Projet). En création avec un seul dépôt, seul le champ
  // chemin est visible (le nom est déduit du basename côté serveur) ; dès
  // qu'il y a plusieurs dépôts (ou en édition), chaque ligne affiche nom + chemin.

  function buildRepoRowsHTML() {
    var simple = modalRepoCreateMode && modalRepos.length === 1;
    if (simple) {
      return '<div class="repo-row repo-row-simple">' +
        '<input class="modal-input mono repo-row-path" placeholder="' + escapeHtml(t('newProject.pathPlaceholder')) + '" value="' + escapeHtml(modalRepos[0].path) + '">' +
        '</div>';
    }
    return modalRepos.map(function (r, i) {
      var canRemove = modalRepos.length > 1;
      return '<div class="repo-row">' +
        '<input class="modal-input repo-row-name" placeholder="' + escapeHtml(t('project.repoNamePlaceholder')) + '" value="' + escapeHtml(r.name) + '">' +
        '<input class="modal-input mono repo-row-path" placeholder="' + escapeHtml(t('newProject.pathPlaceholder')) + '" value="' + escapeHtml(r.path) + '">' +
        (canRemove ? '<button class="icon-btn repo-row-remove" data-action="remove-repo-row" data-index="' + i + '" aria-label="' + escapeHtml(t('project.removeRepo')) + '">✕</button>' : '') +
        '</div>';
    }).join('');
  }
  function captureRepoRowsFromDOM() {
    var nameInputs = document.querySelectorAll('.repo-row-name');
    var pathInputs = document.querySelectorAll('.repo-row-path');
    if (pathInputs.length === 0) return;
    if (nameInputs.length === pathInputs.length) {
      modalRepos = Array.prototype.map.call(pathInputs, function (input, i) {
        return { name: nameInputs[i].value, path: input.value };
      });
    } else {
      modalRepos = [{ name: modalRepos[0] ? modalRepos[0].name : '', path: pathInputs[0].value }];
    }
  }
  function refreshRepoRowsUI() {
    var container = document.getElementById('repo-rows');
    if (container) container.innerHTML = buildRepoRowsHTML();
  }
  function addRepoRow() {
    captureRepoRowsFromDOM();
    modalRepos.push({ name: '', path: '' });
    refreshRepoRowsUI();
  }
  function removeRepoRow(index) {
    captureRepoRowsFromDOM();
    if (modalRepos.length <= 1) return;
    modalRepos.splice(index, 1);
    refreshRepoRowsUI();
  }
  function buildRepoSectionHTML() {
    return '<div class="modal-label">' + escapeHtml(t('project.reposLabel')) + '</div>' +
      '<div id="repo-rows">' + buildRepoRowsHTML() + '</div>' +
      '<button class="add-repo-link" data-action="add-repo-row">' + escapeHtml(t('project.addRepo')) + '</button>';
  }
  function collectReposForSubmit() {
    captureRepoRowsFromDOM();
    return modalRepos.map(function (r) {
      return { name: (r.name || '').trim(), path: (r.path || '').trim() };
    }).filter(function (r) { return r.path; });
  }
  function reposToBody(repos) {
    return repos.map(function (r) {
      var o = { path: r.path };
      if (r.name) o.name = r.name;
      return o;
    });
  }
  function buildProjectExtraFieldsHTML(project) {
    return '<div class="modal-label">' + escapeHtml(t('project.description')) + '</div>' +
      '<input id="project-description" class="modal-input" placeholder="' + escapeHtml(t('project.descriptionPlaceholder')) + '" value="' + escapeHtml(project ? (project.description || '') : '') + '">' +
      '<div class="modal-label">' + escapeHtml(t('project.contextPrompt')) + '</div>' +
      '<textarea id="project-context-prompt" class="modal-textarea" rows="3" placeholder="' + escapeHtml(t('project.contextPromptPlaceholder')) + '">' + (project ? escapeHtml(project.contextPrompt || '') : '') + '</textarea>';
  }

  // Nouveau projet

  function buildNewProjectModalHTML() {
    return '<div class="modal modal-sm">' +
      '<div class="modal-head"><span class="modal-title">' + escapeHtml(t('newProject.title')) + '</span><button class="icon-btn" data-action="close-modal" aria-label="' + escapeHtml(t('common.close')) + '">✕</button></div>' +
      '<div class="modal-label">' + escapeHtml(t('newProject.nameLabel')) + '</div><input id="new-project-name" class="modal-input" placeholder="' + escapeHtml(t('newProject.namePlaceholder')) + '">' +
      buildRepoSectionHTML() +
      buildProjectExtraFieldsHTML(null) +
      '<div id="new-project-error" class="modal-error hidden"></div>' +
      '<div class="modal-foot"><button class="btn-outline" data-action="close-modal">' + escapeHtml(t('common.cancel')) + '</button>' +
      '<button class="btn-green" data-action="submit-new-project">' + escapeHtml(t('common.create')) + '</button></div>' +
      '</div>';
  }
  function openNewProjectModal() {
    modalRepoCreateMode = true;
    modalRepos = [{ name: '', path: '' }];
    openModal(buildNewProjectModalHTML());
    setTimeout(function () { var el = document.getElementById('new-project-name'); if (el) el.focus(); }, 0);
  }
  function submitNewProject() {
    var nameEl = document.getElementById('new-project-name');
    var errEl = document.getElementById('new-project-error');
    var name = nameEl.value.trim();
    var description = document.getElementById('project-description').value.trim();
    var contextPrompt = document.getElementById('project-context-prompt').value.trim();
    var repos = collectReposForSubmit();
    if (!name || repos.length === 0) { errEl.textContent = t('newProject.errorRequired'); errEl.classList.remove('hidden'); return; }
    api('/api/projects', { method: 'POST', body: { name: name, repos: reposToBody(repos), description: description, contextPrompt: contextPrompt } }).then(function (project) {
      upsertProject(project);
      closeModal();
      goProject(project.id);
    }).catch(function (e) {
      errEl.textContent = (e instanceof ApiError && e.message) || t('newProject.errorCreateFailed');
      errEl.classList.remove('hidden');
    });
  }

  // Nouvelle carte

  function buildNewCardModalHTML() {
    return '<div class="modal modal-sm">' +
      '<div class="modal-head"><span class="modal-title">' + escapeHtml(t('newCard.title')) + '</span><button class="icon-btn" data-action="close-modal" aria-label="' + escapeHtml(t('common.close')) + '">✕</button></div>' +
      '<div class="modal-label">' + escapeHtml(t('newCard.titleLabel')) + '</div><input id="new-card-title" class="modal-input" placeholder="' + escapeHtml(t('newCard.titlePlaceholder')) + '">' +
      '<div id="new-card-error" class="modal-error hidden"></div>' +
      '<div class="modal-foot"><button class="btn-outline" data-action="close-modal">' + escapeHtml(t('common.cancel')) + '</button>' +
      '<button class="btn-green" data-action="submit-new-card">' + escapeHtml(t('common.create')) + '</button></div>' +
      '</div>';
  }
  function openNewCardModal() {
    if (!state.projectId) return;
    openModal(buildNewCardModalHTML());
    setTimeout(function () { var el = document.getElementById('new-card-title'); if (el) el.focus(); }, 0);
  }
  function submitNewCard() {
    var titleEl = document.getElementById('new-card-title');
    var errEl = document.getElementById('new-card-error');
    var title = titleEl.value.trim();
    if (!title) { errEl.textContent = t('newCard.errorTitleRequired'); errEl.classList.remove('hidden'); return; }
    api('/api/cards', { method: 'POST', body: { projectId: state.projectId, title: title } }).then(function (card) {
      upsertCard(card);
      closeModal();
      renderMain();
    }).catch(function (e) {
      errEl.textContent = (e instanceof ApiError && e.message) || t('newCard.errorCreateFailed');
      errEl.classList.remove('hidden');
    });
  }

  // Nouvel agent / édition d'agent

  function buildAgentModalHTML(agent) {
    var isEdit = !!agent;
    var title = isEdit ? t('agent.editTitle') : t('agent.newTitle');
    var cliValues = ['claude', 'codex', 'fake'];
    var cliOptions = cliValues.map(function (c) {
      return '<option value="' + c + '"' + (agent && agent.cli === c ? ' selected' : '') + '>' + c + '</option>';
    }).join('');
    var deleteRow = '';
    if (isEdit) {
      var delKey = 'agent-delete:' + agent.id;
      var delPending = isPendingConfirm(delKey);
      var delLabel = delPending ? t('agent.deleteConfirm') : t('agent.delete');
      deleteRow = '<div class="modal-delete-row">' +
        '<button class="delete-link" data-action="confirm-click" data-confirm-key="' + delKey + '" data-confirm-action="agent-delete" data-confirm-id="' + agent.id + '" data-default-label="' + escapeHtml(t('agent.delete')) + '" data-confirm-label="' + escapeHtml(t('agent.deleteConfirm')) + '">' + escapeHtml(delLabel) + '</button>' +
        '</div>';
    }
    return '<div class="modal">' +
      '<div class="modal-head"><span class="modal-title">' + escapeHtml(title) + '</span><button class="icon-btn" data-action="close-modal" aria-label="' + escapeHtml(t('common.close')) + '">✕</button></div>' +
      '<div class="modal-label">' + escapeHtml(t('agent.name')) + '</div><input id="agent-name" class="modal-input" value="' + (agent ? escapeHtml(agent.name) : '') + '">' +
      '<div class="agent-form-row">' +
        '<div><div class="modal-label">' + escapeHtml(t('agent.emoji')) + '</div><input id="agent-emoji" class="modal-input agent-emoji-input" maxlength="4" value="' + (agent ? escapeHtml(agent.emoji || '') : '') + '"></div>' +
        '<div><div class="modal-label">' + escapeHtml(t('agent.color')) + '</div><input type="color" id="agent-color" class="color-input" value="' + (agent && agent.color ? agent.color : '#2f66d0') + '"></div>' +
        '<div><div class="modal-label">' + escapeHtml(t('agent.cli')) + '</div><select id="agent-cli" class="modal-input">' + cliOptions + '</select></div>' +
      '</div>' +
      '<div class="modal-label">' + escapeHtml(t('agent.model')) + '</div><input id="agent-model" class="modal-input mono" value="' + (agent ? escapeHtml(agent.model || '') : '') + '">' +
      '<div class="modal-label">' + escapeHtml(t('agent.contextPrompt')) + '</div><textarea id="agent-context" class="modal-textarea" rows="4">' + (agent ? escapeHtml(agent.contextPrompt || '') : '') + '</textarea>' +
      '<div id="agent-modal-error" class="modal-error hidden"></div>' +
      deleteRow +
      '<div class="modal-foot">' +
        '<button class="btn-outline" data-action="close-modal">' + escapeHtml(t('common.cancel')) + '</button>' +
        '<button class="btn-green" data-action="submit-agent" data-agent-id="' + (agent ? agent.id : '') + '">' + escapeHtml(isEdit ? t('common.save') : t('common.create')) + '</button>' +
      '</div>' +
      '</div>';
  }
  function openNewAgentModal() {
    openModal(buildAgentModalHTML(null));
    setTimeout(function () { var el = document.getElementById('agent-name'); if (el) el.focus(); }, 0);
  }
  function openEditAgentModal(agentId) {
    var agent = state.agentsById[agentId];
    if (!agent) return;
    openModal(buildAgentModalHTML(agent));
  }
  function showAgentModalError(msg) {
    var el = document.getElementById('agent-modal-error');
    if (el) { el.textContent = msg; el.classList.remove('hidden'); }
  }
  function submitAgent(agentId) {
    var name = document.getElementById('agent-name').value.trim();
    var emoji = document.getElementById('agent-emoji').value.trim();
    var color = document.getElementById('agent-color').value;
    var cli = document.getElementById('agent-cli').value;
    var model = document.getElementById('agent-model').value.trim();
    var contextPrompt = document.getElementById('agent-context').value.trim();
    if (!name) { showAgentModalError(t('agent.errorNameRequired')); return; }
    var body = { name: name, emoji: emoji, color: color, cli: cli, model: model, contextPrompt: contextPrompt };
    var isEdit = !!agentId;
    var req = isEdit
      ? api('/api/agents/' + agentId, { method: 'PATCH', body: body })
      : api('/api/agents', { method: 'POST', body: body });
    req.then(function (agent) {
      var i = state.agents.findIndex(function (a) { return a.id === agent.id; });
      if (i >= 0) state.agents[i] = agent; else state.agents.push(agent);
      reindex();
      closeModal();
      renderSidebar();
    }).catch(function (e) {
      showAgentModalError((e instanceof ApiError && e.message) || t('agent.errorSaveFailed'));
    });
  }
  function doDeleteAgent(agentId) {
    api('/api/agents/' + agentId, { method: 'DELETE' }).then(function () {
      state.agents = state.agents.filter(function (a) { return a.id !== agentId; });
      reindex();
      closeModal();
      renderSidebar();
    }).catch(function (e) {
      showAgentModalError((e instanceof ApiError && e.message) || t('agent.errorDeleteFailed'));
      patchConfirmButtons('agent-delete:' + agentId);
    });
  }

  // Édition de projet

  function buildProjectModalHTML(project) {
    return '<div class="modal modal-sm">' +
      '<div class="modal-head"><span class="modal-title">' + escapeHtml(t('project.editTitle')) + '</span><button class="icon-btn" data-action="close-modal" aria-label="' + escapeHtml(t('common.close')) + '">✕</button></div>' +
      '<div class="modal-label">' + escapeHtml(t('project.name')) + '</div><input id="project-edit-name" class="modal-input" value="' + escapeHtml(project.name) + '">' +
      buildRepoSectionHTML() +
      buildProjectExtraFieldsHTML(project) +
      '<div class="modal-label">' + escapeHtml(t('project.checkCmd')) + '</div><input id="project-edit-checkcmd" class="modal-input mono" placeholder="go test ./..." value="' + escapeHtml(project.checkCmd || '') + '">' +
      '<div id="project-modal-error" class="modal-error hidden"></div>' +
      '<div class="modal-foot"><button class="btn-outline" data-action="close-modal">' + escapeHtml(t('common.cancel')) + '</button>' +
      '<button class="btn-green" data-action="submit-project-edit" data-project-id="' + project.id + '">' + escapeHtml(t('common.save')) + '</button></div>' +
      '</div>';
  }
  function openEditProjectModal() {
    var project = state.projectsById[state.projectId];
    if (!project) return;
    modalRepoCreateMode = false;
    var repos = (project.repos && project.repos.length) ? project.repos : [{ name: '', path: '' }];
    modalRepos = repos.map(function (r) { return { name: r.name || '', path: r.path || '' }; });
    openModal(buildProjectModalHTML(project));
    setTimeout(function () { var el = document.getElementById('project-edit-name'); if (el) el.focus(); }, 0);
  }
  function submitProjectEdit(projectId) {
    var name = document.getElementById('project-edit-name').value.trim();
    var checkCmd = document.getElementById('project-edit-checkcmd').value.trim();
    var description = document.getElementById('project-description').value.trim();
    var contextPrompt = document.getElementById('project-context-prompt').value.trim();
    var errEl = document.getElementById('project-modal-error');
    if (!name) { errEl.textContent = t('project.errorNameRequired'); errEl.classList.remove('hidden'); return; }
    var repos = collectReposForSubmit();
    if (repos.length === 0) { errEl.textContent = t('project.errorReposRequired'); errEl.classList.remove('hidden'); return; }
    api('/api/projects/' + projectId, { method: 'PATCH', body: { name: name, checkCmd: checkCmd, repos: reposToBody(repos), description: description, contextPrompt: contextPrompt } }).then(function (project) {
      upsertProject(project);
      closeModal();
      renderSidebar(); renderMain();
    }).catch(function (e) {
      errEl.textContent = (e instanceof ApiError && e.message) || t('project.errorSaveFailed');
      errEl.classList.remove('hidden');
    });
  }

  // Onboarding (modale de bienvenue, une fois par installation)

  function buildOnboardingModalHTML() {
    var cards = [
      { key: 'local', title: t('onboarding.local.title'), desc: t('onboarding.local.desc') },
      { key: 'init', title: t('onboarding.init.title'), desc: t('onboarding.init.desc') },
      { key: 'clone', title: t('onboarding.clone.title'), desc: t('onboarding.clone.desc') }
    ];
    var cardsHTML = cards.map(function (c) {
      var expanded = onboardingExpanded === c.key;
      var body = '';
      if (expanded) {
        if (c.key === 'local') {
          body = '<div class="onboarding-body">' +
            '<button class="btn-green btn-block" data-action="submit-onboarding" data-mode="local">' + escapeHtml(t('onboarding.local.submit')) + '</button>' +
            '</div>';
        } else if (c.key === 'init') {
          body = '<div class="onboarding-body">' +
            '<input id="onboarding-init-remote" class="modal-input mono" placeholder="git@github.com:vous/sillage-workspace.git">' +
            '<button class="btn-green btn-block" data-action="submit-onboarding" data-mode="init">' + escapeHtml(t('onboarding.init.submit')) + '</button>' +
            '</div>';
        } else if (c.key === 'clone') {
          body = '<div class="onboarding-body">' +
            '<input id="onboarding-clone-remote" class="modal-input mono" placeholder="git@github.com:vous/sillage-workspace.git">' +
            '<div class="workspace-warning">' + escapeHtml(t('onboarding.clone.warning')) + '</div>' +
            '<button class="btn-green btn-block" data-action="submit-onboarding" data-mode="clone">' + escapeHtml(t('onboarding.clone.submit')) + '</button>' +
            '</div>';
        }
      }
      return '<div class="onboarding-card ' + (expanded ? 'expanded' : '') + '">' +
        '<button class="onboarding-card-head" data-action="toggle-onboarding-card" data-key="' + c.key + '">' +
          '<span class="onboarding-card-title">' + escapeHtml(c.title) + '</span>' +
          '<span class="onboarding-card-desc">' + escapeHtml(c.desc) + '</span>' +
        '</button>' +
        body +
        '</div>';
    }).join('');
    return '<div class="modal">' +
      '<div class="modal-head"><span class="modal-title">' + escapeHtml(t('onboarding.title')) + '</span></div>' +
      '<div class="onboarding-intro">' + escapeHtml(t('onboarding.intro')) + '</div>' +
      '<div id="onboarding-error" class="modal-error hidden"></div>' +
      '<div class="onboarding-cards" id="onboarding-cards">' + cardsHTML + '</div>' +
      '</div>';
  }
  function openOnboardingModal() {
    onboardingExpanded = null;
    openModal(buildOnboardingModalHTML());
  }
  function toggleOnboardingCard(key) {
    onboardingExpanded = (onboardingExpanded === key) ? null : key;
    openModal(buildOnboardingModalHTML());
  }
  function submitOnboarding(mode) {
    var errEl = document.getElementById('onboarding-error');
    var body = { mode: mode };
    if (mode === 'init') {
      var initEl = document.getElementById('onboarding-init-remote');
      var initRemote = initEl ? initEl.value.trim() : '';
      if (initRemote) body.remote = initRemote;
    } else if (mode === 'clone') {
      var cloneEl = document.getElementById('onboarding-clone-remote');
      var cloneRemote = cloneEl ? cloneEl.value.trim() : '';
      if (!cloneRemote) {
        if (errEl) { errEl.textContent = t('onboarding.errorRemoteRequired'); errEl.classList.remove('hidden'); }
        return;
      }
      body.remote = cloneRemote;
    }
    api('/api/workspace/setup', { method: 'POST', body: body }).then(function (workspace) {
      state.workspace = workspace || state.workspace;
      closeModal();
      if (mode === 'clone') return fetchStateSilently();
    }).catch(function (e) {
      if (errEl) {
        errEl.textContent = (e instanceof ApiError && e.message) || t('onboarding.errorFailed');
        errEl.classList.remove('hidden');
      }
    });
  }

  // Espace de travail (réglages + synchronisation git)

  function buildPreferencesSectionHTML() {
    var settings = state.settings || {};
    return '<div class="modal-label">' + escapeHtml(t('preferences.title')) + '</div>' +
      '<div class="workspace-remote-row">' +
        '<input id="settings-display-name" class="modal-input" placeholder="' + escapeHtml(t('preferences.displayNamePlaceholder')) + '" value="' + escapeHtml(settings.displayName || '') + '">' +
        '<button class="link-save-btn" data-action="save-display-name">' + escapeHtml(t('common.save')) + '</button>' +
      '</div>' +
      '<div class="preferences-lang-row">' +
        '<span class="preferences-lang-label">' + escapeHtml(t('preferences.langLabel')) + '</span>' +
        buildLangSwitchHTML() +
      '</div>' +
      '<div class="preferences-divider"></div>';
  }

  function buildWorkspaceModalBodyHTML() {
    var ws = state.workspace || {};
    var gitEnabled = !!ws.gitEnabled;
    var hasRemote = !!ws.remote;
    var stateLabel = !gitEnabled ? t('workspace.state.local') : (hasRemote ? t('workspace.state.gitRemote') : t('workspace.state.gitNoRemote'));
    var lastCommit = ws.lastCommitAt ? timeAgo(ws.lastCommitAt) : t('workspace.never');
    var lastSync = ws.lastSyncAt ? timeAgo(ws.lastSyncAt) : t('workspace.never');
    var dirtyNote = ws.dirty ? '<div class="workspace-dirty-note">' + escapeHtml(t('workspace.dirtyNote')) + '</div>' : '';

    var syncKey = 'workspace-sync';
    var syncPending = isPendingConfirm(syncKey);
    var syncLabel = syncPending ? t('workspace.syncConfirm') : t('workspace.sync');

    var primaryHTML;
    if (gitEnabled) {
      primaryHTML = '<button class="btn-green btn-block" data-action="confirm-click" data-confirm-key="' + syncKey + '" data-confirm-action="workspace-sync" data-default-label="' + escapeHtml(t('workspace.sync')) + '" data-confirm-label="' + escapeHtml(t('workspace.syncConfirm')) + '">' + escapeHtml(syncLabel) + '</button>';
    } else {
      primaryHTML = '<button class="btn-green btn-block" data-action="activate-workspace-git">' + escapeHtml(t('workspace.activate')) + '</button>';
    }

    return buildPreferencesSectionHTML() +
      '<div class="workspace-state">' + escapeHtml(stateLabel) + '</div>' +
      '<div class="modal-label">' + escapeHtml(t('workspace.remoteLabel')) + '</div>' +
      '<div class="workspace-remote-row">' +
        '<input id="workspace-remote-input" class="modal-input mono" placeholder="git@github.com:vous/sillage-workspace.git" value="' + escapeHtml(ws.remote || '') + '">' +
        '<button class="link-save-btn" data-action="save-workspace-remote">' + escapeHtml(t('common.save')) + '</button>' +
      '</div>' +
      '<div class="workspace-warning">' + escapeHtml(t('workspace.privateWarning')) + '</div>' +
      '<div id="workspace-modal-error" class="modal-error hidden"></div>' +
      '<div id="workspace-sync-message" class="workspace-sync-message hidden"></div>' +
      '<div class="secondary-row">' + primaryHTML + '</div>' +
      '<div class="workspace-meta">' +
        '<span>' + escapeHtml(t('workspace.lastCommit', { time: lastCommit })) + '</span>' +
        '<span>' + escapeHtml(t('workspace.lastSync', { time: lastSync })) + '</span>' +
      '</div>' +
      dirtyNote;
  }
  function buildWorkspaceModalHTML() {
    return '<div class="modal modal-sm">' +
      '<div class="modal-head"><span class="modal-title">' + escapeHtml(t('workspace.title')) + '</span><button class="icon-btn" data-action="close-modal" aria-label="' + escapeHtml(t('common.close')) + '">✕</button></div>' +
      '<div id="workspace-modal-body">' + buildWorkspaceModalBodyHTML() + '</div>' +
      '</div>';
  }
  function openWorkspaceModal() {
    openModal(buildWorkspaceModalHTML());
    setTimeout(function () { var el = document.getElementById('workspace-remote-input'); if (el) el.focus(); }, 0);
  }
  function refreshWorkspaceModalBody() {
    var body = document.getElementById('workspace-modal-body');
    if (body) body.innerHTML = buildWorkspaceModalBodyHTML();
  }
  function saveDisplayName() {
    var el = document.getElementById('settings-display-name');
    var name = el ? el.value.trim() : '';
    var errEl = document.getElementById('workspace-modal-error');
    api('/api/settings', { method: 'PATCH', body: { displayName: name } }).then(function (settings) {
      state.settings = settings || Object.assign({}, state.settings, { displayName: name });
      refreshWorkspaceModalBody();
    }).catch(function (e) {
      if (errEl) {
        errEl.textContent = (e instanceof ApiError && e.message) || t('preferences.errorSaveFailed');
        errEl.classList.remove('hidden');
      }
    });
  }
  function saveWorkspaceRemote() {
    var el = document.getElementById('workspace-remote-input');
    var remote = el ? el.value.trim() : '';
    var errEl = document.getElementById('workspace-modal-error');
    api('/api/workspace', { method: 'PATCH', body: { remote: remote } }).then(function (ws) {
      state.workspace = ws || state.workspace;
      refreshWorkspaceModalBody();
    }).catch(function (e) {
      if (errEl) {
        errEl.textContent = (e instanceof ApiError && e.message) || t('workspace.errorSaveFailed');
        errEl.classList.remove('hidden');
      }
    });
  }
  function activateWorkspaceGit() {
    var el = document.getElementById('workspace-remote-input');
    var remote = el ? el.value.trim() : '';
    var body = { mode: 'init' };
    if (remote) body.remote = remote;
    var errEl = document.getElementById('workspace-modal-error');
    api('/api/workspace/setup', { method: 'POST', body: body }).then(function (ws) {
      state.workspace = ws || state.workspace;
      refreshWorkspaceModalBody();
    }).catch(function (e) {
      if (errEl) {
        errEl.textContent = (e instanceof ApiError && e.message) || t('workspace.errorActivateFailed');
        errEl.classList.remove('hidden');
      }
    });
  }
  function showWorkspaceSyncMessage() {
    var el = document.getElementById('workspace-sync-message');
    if (!el) return;
    el.textContent = t('workspace.syncedJustNow');
    el.classList.remove('hidden');
    setTimeout(function () {
      var el2 = document.getElementById('workspace-sync-message');
      if (el2) el2.classList.add('hidden');
    }, 5000);
  }
  function doWorkspaceSync() {
    var errEl = document.getElementById('workspace-modal-error');
    api('/api/workspace/sync', { method: 'POST', body: { confirm: true } }).then(function (res) {
      state.workspace = state.workspace || {};
      if (res && res.lastSyncAt) state.workspace.lastSyncAt = res.lastSyncAt;
      refreshWorkspaceModalBody();
      showWorkspaceSyncMessage();
    }).catch(function (e) {
      if (errEl) {
        errEl.textContent = (e instanceof ApiError && e.message) || t('workspace.errorSyncFailed');
        errEl.classList.remove('hidden');
      }
    });
  }

  // ---------------------------------------------------------------------
  // Authentification
  // ---------------------------------------------------------------------

  function applyStaticTranslations() {
    document.documentElement.setAttribute('lang', state.lang);
    var p = document.getElementById('login-password');
    var s = document.getElementById('login-submit');
    if (p) p.placeholder = t('login.passwordPlaceholder');
    if (s) s.textContent = t('login.submit');
  }

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
        var msg = t('login.error');
        if (text) { try { var j = JSON.parse(text); if (j && j.error) msg = j.error; } catch (er) {} }
        errEl.textContent = msg; errEl.classList.remove('hidden');
      });
    }).catch(function () {
      errEl.textContent = t('common.networkError'); errEl.classList.remove('hidden');
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

  function onProjectEvent(project) {
    if (!project || !project.id) return;
    upsertProject(project);
    renderSidebar();
    renderMain();
  }

  function onWorkspaceEvent(workspace) {
    state.workspace = workspace;
    refreshWorkspaceModalBody();
    if (workspace && workspace.setupDone && document.getElementById('onboarding-cards')) {
      closeModal();
    }
  }

  function onSettingsEvent(settings) {
    state.settings = settings || state.settings;
    refreshWorkspaceModalBody();
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
    es.addEventListener('project', function (e) { try { onProjectEvent(JSON.parse(e.data)); } catch (er) {} });
    es.addEventListener('workspace', function (e) { try { onWorkspaceEvent(JSON.parse(e.data)); } catch (er) {} });
    es.addEventListener('settings', function (e) { try { onSettingsEvent(JSON.parse(e.data)); } catch (er) {} });
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
      applyRoute();
      connectSSE();
      if (state.workspace && state.workspace.setupDone === false) {
        openOnboardingModal();
      }
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
      case 'open-new-card': openNewCardModal(); break;
      case 'open-new-task': if (state.cardId) openNewTaskModal(state.cardId); break;
      case 'open-new-project': openNewProjectModal(); break;
      case 'open-edit-project': openEditProjectModal(); break;
      case 'submit-project-edit': submitProjectEdit(el.getAttribute('data-project-id')); break;
      case 'add-repo-row': addRepoRow(); break;
      case 'remove-repo-row': removeRepoRow(parseInt(el.getAttribute('data-index'), 10)); break;
      case 'open-new-agent': openNewAgentModal(); break;
      case 'edit-agent': openEditAgentModal(el.getAttribute('data-agent-id')); break;
      case 'submit-agent': submitAgent(el.getAttribute('data-agent-id') || null); break;
      case 'close-modal': closeModal(); break;
      case 'pick-agent': pickAgentInModal(el.getAttribute('data-agent-id')); break;
      case 'submit-new-task': submitNewTask(el.getAttribute('data-card-id')); break;
      case 'submit-new-project': submitNewProject(); break;
      case 'submit-new-card': submitNewCard(); break;
      case 'interrupt': doInterrupt(el.getAttribute('data-task-id')); break;
      case 'accept': doAccept(el.getAttribute('data-task-id')); break;
      case 'reopen': doReopen(el.getAttribute('data-task-id')); break;
      case 'finish-task': doFinish(el.getAttribute('data-task-id')); break;
      case 'confirm-click': handleConfirmClickDispatch(el); break;
      case 'select-diff-file': selectDiffFile(el.getAttribute('data-task-id'), el.getAttribute('data-path')); break;
      case 'send-message': sendMessage(el.getAttribute('data-task-id')); break;
      case 'open-search': openSearch(); break;
      case 'search-goto-project': closeSearch(); goProject(el.getAttribute('data-project-id')); break;
      case 'search-goto-task': closeSearch(); openTaskFromSearch(el.getAttribute('data-task-id')); break;
      case 'logout': doLogout(); break;
      case 'toggle-sidebar': toggleSidebarDrawer(); break;
      case 'close-sidebar-drawer': closeSidebarDrawer(); break;
      case 'set-lang': setLang(el.getAttribute('data-lang')); break;
      case 'open-workspace-modal': openWorkspaceModal(); break;
      case 'save-workspace-remote': saveWorkspaceRemote(); break;
      case 'save-display-name': saveDisplayName(); break;
      case 'activate-workspace-git': activateWorkspaceGit(); break;
      case 'toggle-onboarding-card': toggleOnboardingCard(el.getAttribute('data-key')); break;
      case 'submit-onboarding': submitOnboarding(el.getAttribute('data-mode')); break;
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
    applyStaticTranslations();
    var form = document.getElementById('login-form');
    if (form) form.addEventListener('submit', onLoginSubmit);
    document.addEventListener('click', onGlobalClick);
    document.addEventListener('keydown', onGlobalKeydown);
    window.addEventListener('hashchange', applyRoute);
    boot();
  });
})();
