// Sillage : frontend SPA vanilla (zéro framework, zéro dépendance).
// Contrat d'API : voir docs/SPEC-API.md à la racine du dépôt.
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
      'common.workstreamsWord': 'Chantiers',
      'common.close': 'Fermer',
      'common.cancel': 'Annuler',
      'panel.expand': 'Agrandir le panneau',
      'panel.collapse': 'Réduire le panneau',
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
      'sidebar.projectMenuTooltip': 'Actions du projet',
      'sidebar.markAllRead': 'Tout marquer comme lu',
      'sidebar.hideProject': 'Masquer de la barre latérale',
      'sidebar.showProject': 'Afficher dans la barre latérale',
      'sidebar.agentsHeading': 'Agents',
      'sidebar.agentsAtWork': 'Au travail',
      'sidebar.newAgentTooltip': 'Nouvel agent',
      'sidebar.noAgents': 'Aucun agent',
      'aria.menu': 'Menu',
      'header.newTask': 'Nouvelle tâche',
      'header.newTaskTooltip': 'Nouvelle tâche (N)',
      'header.newProject': 'Nouveau projet',
      'header.newProjectTooltip': 'Nouveau projet (N)',
      'allProjects.emptyTitle': 'Aucun projet pour l\'instant',
      'allProjects.emptySub': 'Créez votre premier projet pour commencer.',
      'allProjects.cardCount.one': '{n} chantier',
      'allProjects.cardCount.other': '{n} chantiers',
      'allProjects.hiddenBadge': 'Masqué de la barre latérale',
      'inbox.empty': 'Boîte de réception vide. Tout est à jour !',
      'kanban.addCard': '+ Ajouter',
      'kanban.empty': 'Rien en cours pour l\'instant.',
      'kanban.emptyAction': 'Confier une première tâche à un agent',
      'kanban.newWorkstream': 'Nouveau chantier',
      'kanban.newTaskTooltip': 'Confier une tâche à un agent (N). Son chantier est créé pour vous.',
      'kanban.card.tasksLabel': 'tâches',
      'kanban.card.noTask': 'Aucune tâche',
      'kanban.card.reviewCount': '{n} à relire',
      'kanban.card.awaitingShip': 'À livrer',
      'inbox.awaitingShip.title': 'Chantiers à livrer',
      'cardMenu.moveTo': 'Déplacer vers {column}',
      'project.editTooltip': 'Modifier le projet',
      'project.editTitle': 'Réglages du projet',
      'project.tabGeneral': 'Général',
      'project.tabRepos': 'Dépôts',
      'project.tabInstructions': 'Instructions',
      'project.tabPermissions': 'Permissions',
      'project.tabDelivery': 'Livraison',
      'project.tabLinks': 'Liens',
      'project.tabDanger': 'Supprimer',
      'project.name': 'Nom',
      'project.baseBranch': 'Branche de base',
      'project.baseBranchHint': 'Les chantiers partent de cette branche et y retournent à la livraison. Vide : branche par défaut du dépôt.',
      'project.checkCmd': 'Commande de vérification',
      'project.permissionProfilesHint': 'Activez seulement les technologies de confiance pour ce projet. Chaque profil autorise ses commandes et place ses téléchargements dans un cache Sillage inscriptible, hors du dépôt et des dossiers système.',
      'project.profile.ruby': 'Ruby',
      'project.profile.rubyDesc': 'bundle, gem, rake et ruby · les gems ne sont jamais installées dans /usr/local',
      'project.profile.node': 'Node.js',
      'project.profile.nodeDesc': 'node, npm, npx, pnpm et yarn · caches de paquets isolés',
      'project.profile.php': 'PHP',
      'project.profile.phpDesc': 'php et Composer · cache et outils Composer isolés',
      'project.profile.python': 'Python',
      'project.profile.pythonDesc': 'python, pip, pytest, ruff et uv · caches de paquets isolés',
      'project.profile.go': 'Go',
      'project.profile.goDesc': 'go, gofmt et golangci-lint · caches de build et de modules isolés',
      'project.profile.rust': 'Rust',
      'project.profile.rustDesc': 'cargo, rustc, rustfmt et clippy · registre et artefacts isolés',
      'project.allowedTools': 'Règles Claude supplémentaires',
      'project.allowedToolsPlaceholder': 'Bash(go test:*)\nBash(gofmt:*)',
      'project.allowedToolsHint': 'Option avancée, une règle par ligne dans la syntaxe de Claude CLI. Les règles des profils cochés sont appliquées automatiquement sans être recopiées ici. Rien qui puisse pousser ne sera accepté.',
      'project.reposHint': 'Chemins locaux des dépôts git de ce projet.',
      'project.repoNamePlaceholder': 'Nom du dépôt',
      'project.repoName': 'Nom',
      'project.repoPath': 'Chemin',
      'project.addRepo': '+ dépôt',
      'project.removeRepo': 'Retirer',
      'project.previewCmdPlaceholder': 'Commande de recette (optionnelle)',
      'project.previewUrlPlaceholder': 'URL à ouvrir (optionnelle)',
      'project.previewHintIntro': 'Recette manuelle : la commande est lancée dans le worktree du chantier ou de la tâche.',
      'project.previewVarIdDesc': 'noms : base de données, conteneur, répertoire',
      'project.previewVarNDesc': 'arithmétique explicite',
      'project.previewVarPortDesc': 'port prêt à l\'emploi, sans calcul à écrire',
      'project.previewVarDirDesc': 'chemins absolus',
      'project.previewVarBranchDesc': 'affichage, debug',
      'project.previewExampleLabel': 'Par exemple, lancer un serveur Python :',
      'project.errorNameRequired': 'Le nom est requis.',
      'project.errorReposRequired': 'Au moins un dépôt est requis.',
      'project.errorSaveFailed': 'Erreur lors de l\'enregistrement.',
      'project.delete': 'Supprimer le projet',
      'project.deleteConfirm': 'Confirmer la suppression ?',
      'project.deleteSubtext': 'Supprime aussi {cards} chantiers et {tasks} tâches.',
      'project.deleteWarning': 'Les worktrees et les branches déjà créés ne sont pas touchés, mais le projet disparaît de Sillage. Sans retour possible.',
      'project.unsaved': 'Modifications non enregistrées',
      'project.linksHint': 'Raccourcis affichés sur la page du projet, 12 au plus.',
      'project.addLink': '+ lien',
      'project.linksEmpty': 'Aucun lien épinglé.',
      'project.linksInvalidUrl': 'Utilisez une URL http(s).',
      'project.linksMax': 'Maximum 12 liens.',
      'project.description': 'Description',
      'project.descriptionPlaceholder': 'Une phrase pour situer le projet',
      'project.contextPrompt': 'Contexte pour les agents',
      'project.contextPromptPlaceholder': 'Conventions, architecture, contraintes à connaître…',
      'project.instructionsHint': 'Transmis à chaque agent du projet, en plus du contexte du chantier.',
      'work.emptyCard': 'Aucune tâche pour l\'instant.',
      'work.emptyCardAction': 'Créer la première tâche',
      'work.emptyFiltered': 'Aucune tâche ne correspond à ce filtre.',
      'filter.all': 'Toutes {n}',
      'filter.waiting': 'En attente {n}',
      'filter.running': 'En cours {n}',
      'filter.review': 'À relire {n}',
      'filter.finished': 'Traitées {n}',
      'badge.new': 'NOUVEAU',
      'status.waiting': 'En attente',
      'status.running': 'En cours',
      'status.review': 'À relire',
      'status.accepted': 'Acceptée',
      'status.cancelled': 'Refusée',
      'status.rebasing': 'Rebase en cours',
      'status.rebasingNote': 'La tâche est rejouée sur le travail qui vient d\'être accepté.',
      'action.interrupt': 'Interrompre l\'agent',
      'action.startNow': 'Démarrer maintenant',
      'action.accept': 'Accepter',
      'action.accepting': 'Acceptation…',
      'action.acceptTooltip': 'Fusionner dans la branche du chantier',
      'action.refuse': 'Refuser',
      'action.refuseTooltip': 'Écarter cette tâche du chantier',
      'action.reopen': 'Rouvrir la tâche',
      'taskRow.waitingFor': 'En attente de « {title} »',
      'waitingFor.note': 'Démarrera automatiquement quand « {title} » sera acceptée.',
      'waitingFor.unknownTask': 'La tâche attendue n\'existe plus : démarrez-la manuellement.',
      'newTask.waitsForLabel': 'Démarrer après',
      'newTask.waitsForNone': 'Immédiatement',
      'behind.taskTooltip.one': '1 commit de retard sur {base} : à rebaser avant acceptation, sinon conflit.',
      'behind.taskTooltip.other': '{n} commits de retard sur {base} : à rebaser avant acceptation, sinon conflit.',
      'behind.cardTooltip.one': 'Le chantier a 1 commit de retard sur {base}.',
      'behind.cardTooltip.other': 'Le chantier a {n} commits de retard sur {base}.',
      'behind.cardLabel.one': '↓ 1 commit de retard sur {base}',
      'behind.cardLabel.other': '↓ {n} commits de retard sur {base}',
      'behind.askRebase': 'Demander le rebase',
      'behind.askRebaseTooltip': 'Envoyer un message à l\'agent pour qu\'il rebase sa branche sur {base}',
      'behind.rebasePrompt.one': 'Ta branche a 1 commit de retard sur {base}. Rebase-la dessus, règle les conflits éventuels, puis relance les vérifications du projet.',
      'behind.rebasePrompt.other': 'Ta branche a {n} commits de retard sur {base}. Rebase-la dessus, règle les conflits éventuels, puis relance les vérifications du projet.',
      'action.cancelTask': 'Annuler la tâche',
      'action.cancelTaskConfirm': 'Confirmer l\'annulation ?',
      'task.delete': 'Supprimer la tâche',
      'task.deleteConfirm': 'Confirmer la suppression ?',
      'task.moreActions': 'Autres actions',
      'task.filesCount.one': '{n} fichier',
      'task.filesCount.other': '{n} fichiers',
      'ship.button': 'Livrer',
      'ship.subtext.pr': 'Pousse {branch} et ouvre une pull request vers {base}',
      'ship.subtext.mr': 'Pousse {branch} et ouvre une merge request vers {base}',
      'ship.subtext.push': 'Pousse {branch} sur origin, sans ouvrir de pull request',
      'ship.subtext.unknownForge': 'Pousse {branch} (forge inconnue : pas de pull request)',
      'ship.subtext.merge': 'Fusionne {branch} dans {base} en local, sans jamais pousser',
      'ship.subtext.mergePush': 'Fusionne {branch} dans {base}, puis pousse {base}',
      'ship.subtext.repos': '{n} dépôts concernés',
      'ship.subtext.nothing': 'Rien à livrer pour l\'instant',
      'ship.partial.one': 'Livraison partielle : 1 tâche n\'est pas encore acceptée',
      'ship.partial.other': 'Livraison partielle : {n} tâches ne sont pas encore acceptées',
      'ship.partialNote.one': '1 tâche est encore en cours ou à relire : seul le travail accepté part maintenant. Vous pourrez livrer le reste ensuite.',
      'ship.partialNote.other': '{n} tâches sont encore en cours ou à relire : seul le travail accepté part maintenant. Vous pourrez livrer le reste ensuite.',
      'ship.blocked.noTasks': 'Aucune tâche dans ce chantier',
      'ship.blocked.nothingAccepted': 'Aucune tâche acceptée',
      'ship.blocked.nothingToShip': 'Aucune branche à livrer',
      'ship.blocked.behindTarget': '{base} a avancé de son côté : rattrapez-la avant de livrer (la fusion se fait en fast-forward uniquement).',
      'ship.alreadyOnTarget': 'Déjà sur {base}',
      'ship.alreadyOnTargetSub': 'Tout le travail du chantier est arrivé dans {base} : plus rien à livrer.',
      'catchUp.button': 'Rattraper {base}',
      'catchUp.tooltip': 'Fusionne {base} dans la branche du chantier, sans réécrire son historique. C\'est ce qui rend la livraison possible à nouveau.',
      'catchUp.done': 'Chantier remis à jour depuis {base}.',
      'catchUp.upToDate': 'Le chantier était déjà à jour.',
      'catchUp.conflictTitle': 'Conflit avec {base}',
      'catchUp.conflictTitleMulti': 'Conflits de rattrapage',
      'catchUp.conflictBody': 'Fusionner {base} dans le chantier bute sur {files}. Rien n\'a été modifié : le chantier est intact. Un agent peut reprendre la base et régler les conflits dans une tâche dédiée.',
      'catchUp.askAgent': 'Confier à un agent',
      'catchUp.errorFailed': 'Le rattrapage a échoué.',
      'catchUp.taskTitle': 'Rattraper {base}',
      'catchUp.taskPrompt': 'La branche de ce chantier est en retard sur {base} et ne peut plus être fusionnée : {files} entre en conflit. Fusionne {base} dans la branche du chantier, règle les conflits en gardant les deux intentions, puis relance les vérifications du projet.',
      'ship.modalTitle': 'Livrer le chantier',
      'ship.modalConfirm': 'Livrer',
      'ship.shipping': 'Livraison…',
      'ship.repoNothing': 'Rien à livrer',
      'ship.repoShipped': 'Déjà livré',
      'ship.commits.one': '{n} nouveau commit',
      'ship.commits.other': '{n} nouveaux commits',
      'ship.files.one': '{n} fichier',
      'ship.files.other': '{n} fichiers',
      'ship.prLink': 'Voir la pull request',
      'ship.mergedInto': 'Fusionné dans {base}',
      'ship.mergedPushed': 'Fusionné dans {base} et poussé',
      'ship.pushedBranch': 'Branche {branch} poussée',
      'ship.errorFailed': 'Échec de la livraison.',
      'preview.button': 'Recette',
      'preview.buttonCard': 'Recette du chantier',
      'preview.tooltip': 'Lancer ce chantier pour s\'en servir',
      'preview.taskTooltip': 'Lancer cette tâche pour s\'en servir',
      'preview.cardTitle': 'Recette : {title}',
      'preview.taskTitle': 'Recette de la tâche #{ref}',
      'preview.allTitle': 'Recettes en cours',
      'preview.taskScope': 'Vous recettez la tâche #{ref}, pas le chantier.',
      'preview.statusRunning': 'en cours',
      'preview.statusStopped': 'arrêtée',
      'preview.statusExited': 'terminé (code {code})',
      'preview.statusFailed': 'échec du lancement',
      'preview.statusIdle': 'jamais lancée',
      'preview.start': 'Lancer',
      'preview.stop': 'Arrêter',
      'preview.restart': 'Relancer',
      'preview.showLog': 'Journal',
      'preview.noCmd': 'Pas de commande de recette',
      'preview.noCmdHint': 'Ajoutez-la dans les réglages du projet, dépôt par dépôt. En attendant, le chemin du worktree est juste là.',
      'preview.openSettings': 'Ouvrir les réglages du projet',
      'preview.noBranch': 'Aucune branche de chantier pour l\'instant : créez une tâche d\'abord.',
      'preview.worktree': 'Worktree',
      'preview.copyPath': 'Copier le chemin',
      'preview.copied': 'Copié',
      'preview.logEmpty': 'Aucune sortie pour le moment.',
      'preview.errorStartFailed': 'Le lancement a échoué.',
      'preview.errorStopFailed': 'L\'arrêt a échoué.',
      'preview.running.one': '{n} recette en cours',
      'preview.running.other': '{n} recettes en cours',
      'delivery.label': 'Livrer veut dire',
      'delivery.modePr': 'Ouvrir une pull request',
      'delivery.modePush': 'Pousser la branche',
      'delivery.modeMerge': 'Fusionner en local',
      'delivery.modeMergePush': 'Fusionner puis pousser',
      'delivery.targetPlaceholder': 'branche par défaut du dépôt',
      'delivery.mergeNote': 'Fusion locale en fast-forward uniquement : Sillage ne pousse jamais cette branche.',
      'delivery.mergePushNote': 'La branche cible est d\'abord rattrapée depuis origin, fusionnée en fast-forward, puis poussée. Jamais de force.',
      'delivery.prNote': 'La branche du chantier est poussée, puis la pull request (ou merge request) est ouverte.',
      'delivery.pushNote': 'La branche du chantier est poussée sur origin. Rien n\'est ouvert, rien n\'est fusionné.',
      'delivery.warning.ghMissing': 'gh introuvable dans le PATH : repli sur une URL de pull request pré-remplie.',
      'delivery.warning.ghInstallLink': 'Installer le CLI GitHub',
      'delivery.warning.glabMissing': 'glab introuvable dans le PATH : repli sur une URL de merge request pré-remplie.',
      'delivery.warning.glabInstallLink': 'Installer le CLI GitLab',
      'delivery.warning.noRemote': 'Un dépôt du projet n\'a pas de remote « origin » : rien ne peut être poussé.',
      'delivery.warning.unknownForge': 'Forge inconnue : la branche sera poussée sans ouvrir de pull request.',
      'tabs.conversation': 'Conversation',
      'tabs.diff': 'Diff',
      'tabs.deliverables': 'Livrables',
      'tabs.history': 'Historique',
      'chat.you': 'Vous',
      'chat.placeholder': 'Répondre à {name}…',
      'chat.send': 'Envoyer ⏎',
      'chat.working': 'Travaille…',
      'chat.jumpToBottom': 'Revenir en bas',
      'chat.newMessage': 'Nouveau message',
      'chat.copyMessage': 'Copier le message',
      'chat.copyCode': 'Copier le code',
      'chat.copied': 'Copié',
      'chat.attach': 'Joindre une image',
      'chat.attachRemove': 'Retirer l\'image',
      'chat.attachTooMany': 'Six images au maximum par message.',
      'chat.attachTooLarge': '{name} : image trop lourde (8 Mo maximum).',
      'chat.attachUnsupported': '{name} : format non pris en charge (PNG, JPEG, GIF ou WebP).',
      'chat.attachFailed': 'Image illisible : {name}.',
      'chat.imageOpen': 'Voir en grand',
      'chat.accepted': 'Acceptée : fusionnée dans {branch}',
      'chat.autoAccepted': 'Acceptée : son travail était déjà dans {branch}',
      'chat.mergeConflict': 'Conflit avec la branche du chantier sur {files} : demandez à l\'agent de reprendre la base.',
      'chat.rebased': 'Rebasée automatiquement sur {branch} : cette tâche repart du travail accepté.',
      'chat.rebaseConflict': 'Rebase automatique impossible sur {files} (rien n\'a été modifié) : demandez à l\'agent de reprendre la base.',
      'chat.toolDenied': 'Outil refusé à l\'agent : {tool}. Pour l\'autoriser, ajoutez-le aux outils autorisés dans les réglages du projet.',
      'chat.interrupted': 'Agent interrompu par un arrêt de Sillage : la tâche est repassée en revue. Ce qu\'il avait commité est là ; envoyez-lui un message pour qu\'il reprenne.',
      'conversation.empty': 'Aucun message pour l\'instant.',
      'diff.empty': 'Aucune modification.',
      'diff.openDifftool': 'Ouvrir dans l\'outil de diff',
      'diff.openIn': 'Ouvrir dans {tool}',
      'diff.openAllHint': 'Toutes les modifications d\'un coup, côte à côte',
      'diff.noToolHint': 'Aucun outil de diff réglé dans git : git config --global diff.tool meld',
      'diff.openFile': 'Ouvrir ce fichier dans {tool}',
      'diff.files.one': '{n} fichier',
      'diff.files.other': '{n} fichiers',
      'diff.collapseAll': 'Tout replier',
      'diff.expandAll': 'Tout déplier',
      'diff.collapse': 'Replier',
      'diff.expand': 'Déplier',
      'diff.new': 'nouveau',
      'diff.deleted': 'supprimé',
      'diff.generated': 'généré',
      'diff.skipped.one': '{n} ligne inchangée',
      'diff.skipped.other': '{n} lignes inchangées',
      'deliverables.code': 'Code',
      'deliverables.docs': 'Documents',
      'deliverables.images': 'Captures',
      'deliverables.empty': 'Aucun élément.',
      'history.empty': 'Aucune commande pour l\'instant.',
      'newTask.title': 'Nouvelle tâche',
      'newTask.titlePlaceholder': 'Que doit faire l\'agent ?',
      'newTask.promptPlaceholder': 'Description ou instructions détaillées (optionnel)',
      'newTask.attachHint': 'Joindre une image, ou la coller (Ctrl+V)',
      'newTask.agentLabel': 'Agent',
      'newTask.repoLabel': 'Dépôt',
      'newTask.projectContextNote': '+ contexte du projet',
      'newTask.workstreamContextNote': '+ contexte du chantier',
      'newTask.submit': 'Créer et discuter',
      'newTask.submitAnother': 'Créer et enchaîner',
      'newTask.submitTooltip': 'Créer et ouvrir la conversation ({mod}+Entrée)',
      'newTask.submitAnotherTooltip': 'Créer et garder le formulaire ouvert ({mod}+Maj+Entrée)',
      'newTask.created': 'Tâche #{ref} créée : {title}',
      'newTask.agentHint': 'Flèches pour changer d\'agent',
      'newTask.errorTitleRequired': 'Le titre est requis.',
      'newTask.errorAgentRequired': 'Choisissez un agent.',
      'newTask.errorCreateFailed': 'Erreur lors de la création.',
      'shortcuts.title': 'Raccourcis clavier',
      'shortcuts.hint': 'Raccourcis : ?',
      'shortcuts.sectionGlobal': 'Partout',
      'shortcuts.sectionForm': 'Dans un formulaire',
      'shortcuts.sectionSearch': 'Dans la recherche',
      'shortcuts.sectionTask': 'Dans une tâche',
      'shortcuts.search': 'Rechercher',
      'shortcuts.create': 'Créer (tâche, chantier ou projet selon l\'écran)',
      'shortcuts.help': 'Afficher cette aide',
      'shortcuts.escape': 'Fermer / revenir en arrière',
      'shortcuts.submit': 'Valider',
      'shortcuts.submitAnother': 'Valider et enchaîner',
      'shortcuts.pickAgent': 'Changer d\'agent',
      'shortcuts.tab': 'Passer d\'un champ à l\'autre (reste dans le formulaire)',
      'shortcuts.searchNav': 'Parcourir les résultats',
      'shortcuts.searchOpen': 'Ouvrir le résultat',
      'shortcuts.sendMessage': 'Envoyer le message',
      'shortcuts.expandPanel': 'Agrandir ou réduire le panneau',
      'shortcuts.diffNav': 'Fichier suivant / précédent du diff',
      'newProject.title': 'Nouveau projet',
      'newProject.pathLabel': 'Chemin d\'un dépôt git',
      'newProject.pathPlaceholder': '/home/utilisateur/projets/mon-projet',
      'newProject.hint': 'Le nom du projet et le mode de livraison en sont déduits. Tout le reste s\'ajuste ensuite dans les réglages.',
      'newProject.errorPathRequired': 'Le chemin d\'un dépôt git est requis.',
      'newProject.errorCreateFailed': 'Erreur lors de la création du projet.',
      'newCard.title': 'Nouveau chantier',
      'newCard.titleLabel': 'Titre du chantier',
      'newCard.titlePlaceholder': 'Titre du chantier',
      'newCard.errorTitleRequired': 'Le titre est requis.',
      'newCard.errorCreateFailed': 'Erreur lors de la création du chantier.',
      'newCard.submitAndPlan': 'Créer et planifier',
      'newCard.submitAndPlanTooltip': 'Créer le chantier, puis demander un plan de tâches à une IA',
      'plan.button': 'Plan de tâches',
      'plan.buttonTooltip': 'Demander à une IA de découper ce chantier en étapes',
      'plan.title': 'Plan de tâches',
      'plan.intro': 'L\'agent lit le dépôt et propose un découpage en étapes. Il n\'écrit rien : rien n\'est créé avant que vous ne validiez.',
      'plan.extraLabel': 'Précisions (facultatif)',
      'plan.extraPlaceholder': 'Par où commencer, ce que le découpage doit respecter, ce qu\'il ne doit pas toucher…',
      'plan.repoLabel': 'Dépôt à lire',
      'plan.agentLabel': 'Agent',
      'plan.agentHint': 'Il planifiera, puis fera le travail',
      'plan.propose': 'Proposer un plan',
      'plan.proposing': 'Lecture du dépôt…',
      'plan.waitNote': 'L\'agent lit le dépôt. Cela prend souvent une minute ou deux.',
      'plan.errorNoAgent': 'Aucun agent ne sait planifier. Les agents Claude et Codex le savent, ainsi que l\'agent de test.',
      'plan.errorAgentRequired': 'Choisissez un agent.',
      'plan.errorFailed': 'Le plan n\'a pas abouti.',
      'plan.stepsTitle.one': '1 étape proposée',
      'plan.stepsTitle.other': '{n} étapes proposées',
      'plan.stepTitlePlaceholder': 'Titre de la tâche',
      'plan.stepPromptPlaceholder': 'Ce que l\'agent doit faire',
      'plan.removeStep': 'Retirer cette étape',
      'plan.chainLabel': 'Enchaîner : chaque tâche démarre quand la précédente est acceptée',
      'plan.chainHint': 'Décoché, toutes les tâches démarrent en même temps.',
      'plan.again': 'Recommencer',
      'plan.create.one': 'Créer la tâche',
      'plan.create.other': 'Créer les {n} tâches',
      'plan.creating': 'Création…',
      'plan.errorEmpty': 'Il ne reste aucune étape à créer.',
      'plan.errorPartial': '{n} tâche(s) créée(s), puis une erreur : {msg}',
      'issues.import': 'Depuis une issue GitHub',
      'issues.importTooltip': 'Choisir une issue ouverte du dépôt : elle remplit le titre et la consigne',
      'issues.title': 'Issues ouvertes',
      'issues.searchPlaceholder': 'Rechercher : texte libre, label:bug, author:…',
      'issues.repoLabel': 'Dépôt',
      'issues.loading': 'Lecture des issues…',
      'issues.empty': 'Aucune issue ouverte sur ce dépôt.',
      'issues.emptySearch': 'Aucune issue ne correspond à cette recherche.',
      'issues.errorFailed': 'Les issues n\'ont pas pu être lues.',
      'issues.cap': 'Les {n} issues les plus récentes. Affinez la recherche pour aller au-delà.',
      'issues.by': 'par {login}',
      'issues.promptHeader': 'Corrige l\'issue GitHub #{n} : {title}',
      'nav.vigie': 'Vigie',
      'vigie.check': 'Vérifier maintenant',
      'vigie.checking': 'Vérification…',
      'vigie.checkedAgo': 'vérifié {when}',
      'vigie.notify': 'Activer les notifications',
      'vigie.introTitle': 'La Vigie relit à votre place',
      'vigie.introSub': 'Choisissez les projets à surveiller. À chaque nouvelle PR ou issue, un agent la relit en lecture seule et vous prépare un verdict et une revue. Rien n\'est jamais publié sur GitHub.',
      'vigie.noGithub': 'Aucun de vos projets n\'a de dépôt sur GitHub.',
      'vigie.emptyTitle': 'Rien à signaler',
      'vigie.emptySub': 'Les nouvelles PR et issues des projets surveillés arriveront ici, déjà relues.',
      'vigie.watchedHeading': 'Projets sur GitHub',
      'vigie.watch': 'Surveiller',
      'vigie.unwatch': 'Ne plus surveiller',
      'vigie.analyst': 'Relu par',
      'vigie.kind.pr': 'PR',
      'vigie.kind.issue': 'Issue',
      'vigie.kind.upkeep': 'Suggestion',
      'vigie.detail.kindUpkeep': 'Suggestion',
      'vigie.decision.upgrade': 'À considérer',
      'vigie.detail.plan': 'Ce qu\'il faudrait faire',
      'vigie.detail.foundBy': 'Repéré par la veille d\'entretien {when}',
      'vigie.detail.discussion': 'La discussion sur GitHub',
      'vigie.newComments.one': '{n} nouveau commentaire',
      'vigie.newComments.other': '{n} nouveaux commentaires',
      'vigie.upkeep.check': 'Vérifier les mises à jour',
      'vigie.upkeep.last': 'mises à jour vérifiées {when}',
      'vigie.upkeep.running': '{agent} vérifie si {repo} a besoin de mises à jour',
      'vigie.upkeep.promptHeader': 'Entretien du dépôt {repo} : {title}.',
      'vigie.log.upkeep': 'Entretien de {repo} vérifié, nouveaux constats : {n}',
      'vigie.log.upkeepFailed': 'Vérification de l\'entretien de {repo} impossible : {error}',
      'vigie.log.updated': 'relancées par un commentaire : {n}',
      'vigie.by': 'par {login}',
      'vigie.high': 'Prioritaire',
      'vigie.queued': 'En attente d\'analyse',
      'vigie.stopped': 'Relecture arrêtée.',
      'vigie.queuedNext': 'En file : c\'est la prochaine relue',
      'vigie.queuedAhead.one': 'En file : {n} relecture avant elle',
      'vigie.queuedAhead.other': 'En file : {n} relectures avant elle',
      'vigie.queuedPaused': 'En attente : la Vigie est coupée sur ce projet',
      'vigie.queuedHint': 'Les relectures passent une à la fois, pour ne pas lancer dix agents sur votre machine.',
      'vigie.live.title': '{agent} relit',
      'vigie.live.diff': 'Lecture du diff sur GitHub…',
      'vigie.live.start': 'Démarrage de l\'agent…',
      'vigie.live.limit': 'abandon après {n} min',
      'vigie.live.stop': 'Arrêter',
      'sessions.pill.active': 'Une session Claude en cours dans ce projet',
      'sessions.pill.recent.one': '{n} session Claude récente',
      'sessions.pill.recent.other': '{n} sessions Claude récentes',
      'sessions.title': 'Sessions Claude Code',
      'sessions.hint': 'Les sessions lancées à la main dans les dossiers de ce projet, celles que liste « claude --resume ». Sillage les montre sans y toucher.',
      'sessions.active': 'en cours',
      'sessions.prompts.one': '{n} demande',
      'sessions.prompts.other': '{n} demandes',
      'sessions.copy': 'Copier la commande de reprise',
      'sessions.copied': 'Commande copiée',
      'sessions.workstream': 'En faire un chantier',
      'sessions.allPrompts': 'Ses dernières demandes',
      'sessions.empty': 'Aucune session Claude Code récente dans les dossiers de ce projet.',
      'sessions.promptHeader': 'Reprends le travail d\'une session Claude Code menée hors de Sillage, « {title} ».',
      'sessions.promptBranch': 'Elle travaillait dans {cwd}, sur la branche {branch}. Ce qui n\'y a pas été commité n\'est pas dans ce worktree.',
      'sessions.promptAsks': 'Ses dernières demandes, de la plus ancienne à la plus récente :',
      'sessions.shipWarning': 'Une session Claude Code est ouverte dans {repo} (« {title} ») : elle écrit peut-être au même endroit que cette livraison.',
      'autoconf.title': 'Configuration automatique',
      'autoconf.hint': 'Un agent lit le dépôt (fichiers de build, CI, README) et propose la vérification lancée après chaque tâche, la commande de recette et son adresse. Il ne modifie rien, et vous relisez avant d\'enregistrer.',
      'autoconf.start': 'Lancer l\'analyse',
      'autoconf.with': 'Agent',
      'autoconf.running': '{agent} lit {repo}',
      'autoconf.runningShort': '{agent} configure le projet',
      'autoconf.readyShort': 'Configuration proposée, à relire',
      'autoconf.stop': 'Arrêter',
      'autoconf.ready': 'Proposition de {agent}',
      'autoconf.apply': 'Reprendre ces valeurs',
      'autoconf.applied': 'Valeurs reprises dans les onglets. Relisez-les, puis enregistrez.',
      'autoconf.dismiss': 'Oublier',
      'autoconf.retry': 'Relancer',
      'autoconf.error': 'Échec : {error}',
      'autoconf.stopped': 'Analyse arrêtée.',
      'autoconf.check': 'Vérification après chaque tâche',
      'autoconf.preview': 'Recette',
      'autoconf.url': 'Adresse',
      'autoconf.profiles': 'Profils de permissions',
      'autoconf.found': 'Trouvé',
      'autoconf.notes': 'À savoir',
      'autoconf.install': 'installation',
      'autoconf.build': 'build',
      'autoconf.test': 'tests',
      'autoconf.current': 'actuellement : {v}',
      'autoconf.nothing': 'rien de proposé',
      'autoconf.warning': 'Ces commandes s\'exécuteront sur votre machine : relisez-les avant d\'enregistrer.',
      'autoconf.noAgent': 'Il faut un agent capable de lire sans écrire (Claude Code ou Codex).',
      'autoconf.nudge': 'Sillage ne sait pas encore vérifier ni lancer ce projet.',
      'autoconf.nudgeAction': 'Configurer automatiquement',
      'vigie.analyzing': 'Analyse en cours…',
      'vigie.error': 'Analyse impossible : {error}',
      'vigie.copyReview': 'Copier les commentaires',
      'vigie.copied': 'Copiés, à coller sur GitHub',
      'inbox.signals.title': 'Sur GitHub',
      'sidebar.agentsToSetup.one': '1 agent à installer',
      'sidebar.agentsToSetup.other': '{n} agents à installer',
      'vigie.githubAction.merge': 'Merger sur GitHub',
      'vigie.githubAction.close': 'Fermer sur GitHub',
      'vigie.githubAction.changes': 'Voir sur GitHub',
      'vigie.githubAction.rework': 'Voir sur GitHub',
      'vigie.githubAction.needs-info': 'Voir sur GitHub',
      'vigie.githubAction.fix': 'Voir sur GitHub',
      'vigie.githubAction.feature': 'Voir sur GitHub',
      'vigie.githubAction.undefined': 'Voir sur GitHub',
      'vigie.githubAction.': 'Voir sur GitHub',
      'lane.next': 'À lancer',
      'lane.agents': 'Les agents travaillent',
      'lane.you': 'À vous',
      'lane.shipped': 'Livré',
      'vigie.detail.review': 'Commentaires proposés',
      'vigie.detail.reply': 'Réponse proposée',
      'vigie.detail.nothingToAdd': 'Rien à ajouter : la PR peut être mergée telle quelle.',
      'vigie.detail.prBody': 'Description de la PR, par {login}',
      'vigie.detail.issueBody': 'L’issue, par {login}',
      'hello.day': 'Bonjour',
      'hello.evening': 'Bonsoir',
      'hello.and': 'et',
      'hello.waiting': 'Vous avez {list}.',
      'hello.nothing': 'Rien ne vous attend.',
      'hello.review.one': '{n} tâche à relire',
      'hello.review.other': '{n} tâches à relire',
      'hello.signals.one': '{n} sujet GitHub à trancher',
      'hello.signals.other': '{n} sujets GitHub à trancher',
      'hello.ship.one': '{n} chantier à livrer',
      'hello.ship.other': '{n} chantiers à livrer',
      'hello.running.one': 'Un agent travaille pour vous.',
      'hello.running.other': '{n} agents travaillent pour vous.',
      'pulse.you.one': '{n} pour vous',
      'pulse.you.other': '{n} pour vous',
      'pulse.agents.one': '{n} agent au travail',
      'pulse.agents.other': '{n} agents au travail',
      'pulse.calm': 'Tout est calme',
      'pulse.empty': 'Rien pour l’instant',
      'vigie.detail.kindPr': 'Pull request',
      'vigie.detail.kindIssue': 'Issue',
      'vigie.detail.opened': 'ouverte {when}',
      'vigie.detail.copy': 'Copier',
      'vigie.detail.onGithub': 'Ouvrir sur GitHub',
      'vigie.detail.readBy': 'Relu par {agent}',
      'vigie.detail.analyzing': '{agent} relit…',
      'vigie.detail.waiting': 'Le jugement arrive dès que la relecture est finie.',
      'vigie.detail.noBody': 'L’auteur n’a rien écrit.',
      'vigie.openGithub': 'Voir sur GitHub',
      'vigie.workstream': 'Confier à un agent',
      'vigie.toCard': 'En faire un chantier',
      'vigie.toCardHint': 'Un chantier, puis un agent le découpe en tâches que vous relisez',
      'vigie.dismiss': 'Retirer de la liste',
      'vigie.retry': 'Réessayer',
      'vigie.localOnly': 'Sillage ne publie jamais rien sur GitHub. Les commentaires proposés restent ici : copiez-les et collez-les vous-même si vous êtes d\'accord.',
      'vigie.promptAnalysis': 'Analyse préalable de la Vigie :',
      'vigie.log.error': 'Lecture de {repo} impossible : {error}',
      'vigie.log.analysisFailed': 'Échec de l\'analyse de {kind} #{n} ({repo}) : {error}',
      'vigie.log.analysis': '{kind} #{n} ({repo}) analysée en {s} s',
      'vigie.log.older': 'ouvertes avant le {date}, ignorées : {n}',
      'vigie.log.drafts': 'brouillons ignorés : {n}',
      'vigie.log.known': 'déjà vues : {n}',
      'vigie.log.list': '{repo} · {kind} ouvertes : {open}, nouvelles : {added}',
      'vigie.log.empty': 'Aucune vérification depuis le démarrage de Sillage.',
      'vigie.log.hide': 'Masquer le journal',
      'vigie.log.show': 'Voir le journal',
      'vigie.replyNote': 'Réponse rédigée par l\'agent, en anglais, non publiée. « Copier la réponse » la met dans le presse-papier pour la coller sur l\'issue.',
      'vigie.copyReply': 'Copier la réponse',
      'vigie.decision.needs-info': 'Informations manquantes',
      'vigie.decision.feature': 'Demande d\'évolution',
      'vigie.decision.fix': 'Bug à corriger',
      'vigie.decision.close': 'À fermer',
      'vigie.decision.rework': 'À retravailler',
      'vigie.decision.changes': 'Quelques retours à faire',
      'vigie.decision.merge': 'Mergeable en l\'état',
      'vigie.neverChecked': 'jamais vérifié',
      'vigie.reviewNote': 'Commentaires rédigés par l\'agent, non publiés. « Copier les commentaires » les met dans le presse-papier pour les coller dans la revue GitHub.',
      'workstream.editTooltip': 'Modifier le chantier',
      'workstream.editTitle': 'Chantier',
      'workstream.errorSaveFailed': 'Erreur lors de l\'enregistrement.',
      'workstream.delete': 'Supprimer le chantier',
      'workstream.deleteConfirm': 'Confirmer la suppression ?',
      'workstream.deleteSubtext.one': 'Supprime aussi sa tâche.',
      'workstream.deleteSubtext.other': 'Supprime aussi ses {n} tâches.',
      'agent.newTitle': 'Nouvel agent',
      'agent.editTitle': 'Modifier l\'agent',
      'agent.name': 'Nom',
      'agent.emoji': 'Emoji',
      'agent.color': 'Couleur',
      'agent.cli': 'CLI',
      'agent.cli.claude': 'Claude Code (claude)',
      'agent.cli.codex': 'OpenAI Codex (codex)',
      'agent.cli.copilot': 'GitHub Copilot (copilot)',
      'agent.cli.agy': 'Google Antigravity (agy)',
      'agent.cli.kiro': 'Kiro CLI (kiro)',
      'agent.cli.fake': 'Agent de test (fake)',
      'agent.model': 'Modèle',
      'agent.contextPrompt': 'Prompt de contexte',
      'agent.delete': 'Supprimer',
      'agent.deleteConfirm': 'Confirmer la suppression ?',
      'agent.errorNameRequired': 'Le nom est requis.',
      'agent.errorSaveFailed': 'Erreur lors de l\'enregistrement.',
      'agent.errorDeleteFailed': 'Erreur lors de la suppression.',
      'agent.warning.codexSandbox': 'Codex a besoin des espaces de noms utilisateur non privilégiés pour son bac à sable, et AppArmor les bloque sur cette machine. Les tâches confiées à cet agent ne peuvent pas démarrer tant que ce n\'est pas résolu.',
      'agent.warning.copyCmd': 'Copier la commande',
      'agent.warning.codexSandboxFallback': 'Vous pouvez aussi contourner le problème en lançant Sillage avec la variable d\'environnement SILLAGE_CODEX_SANDBOX=danger-full-access ; le confinement restant est alors celui de Sillage (worktree dédié, pas de push par l\'agent).',
      'agent.warning.codexSandboxLink': 'En savoir plus (documentation OpenAI Codex)',
      'agent.warning.agyPolicy': 'Antigravity demande votre accord avant chaque commande et devant certains fichiers, et personne ne peut le donner quand un agent travaille tout seul : l\'accord est refusé d\'office et la tâche se termine sans rien produire. Il lui faut deux réglages : une politique d\'exécution qui fait confiance au bac à sable, et le droit de lire et écrire dans les worktrees.',
      'agent.warning.agyPolicyHint': 'Sillage peut les poser dans ~/.gemini/antigravity-cli/settings.json. Les commandes partiront alors sans question, mais seulement dans le bac à sable, que Sillage impose toujours à cet agent ; les droits sur les fichiers s\'arrêtent au dossier des worktrees, son terrain de travail. Vos autres réglages Antigravity sont conservés.',
      'agent.warning.agyPolicyLink': 'Voir la documentation d’Antigravity CLI',
      'agent.warning.agyPolicyFix': 'Régler pour moi',
      'agent.warning.agyPolicyFixConfirm': 'Écrire dans settings.json ?',
      'agent.warning.agyPolicyFixed': 'Réglé. Antigravity peut travailler dans son bac à sable.',
      'agent.warning.kiroApiKey': 'Kiro CLI est installé, mais son mode headless exige une clé API dans la variable d’environnement KIRO_API_KEY. Les tâches confiées à cet agent ne peuvent pas démarrer sans elle.',
      'agent.warning.kiroApiKeyHint': 'Ajoutez KIRO_API_KEY à l’environnement qui lance Sillage, puis relancez Sillage. La clé n’est ni affichée ni enregistrée dans l’espace de travail.',
      'agent.warning.kiroApiKeyLink': 'Voir l’authentification de Kiro CLI',
      'agent.warning.cliNotFound': 'Agent non connecté : CLI {cli} introuvable dans le PATH.',
      'agent.warning.installHint': 'Installez le CLI, puis relancez Sillage :',
      'agent.warning.installLink': 'Voir la documentation d’installation',
      'agent.quotaTitle': 'Quota',
      'agent.quotaWindow5h': '5 heures',
      'agent.quotaWindowWeek': 'Semaine',
      'agent.quotaUsedPercent': '{percent}% utilisé',
      'agent.quotaResetsIn': 'réinitialisation {time}',
      'agent.quotaUpdatedAt': 'Mis à jour : {time}',
      'agent.quotaUnavailable': 'Quotas non disponibles pour cet agent (le CLI {cli} ne les expose pas).',
      'time.inMin': 'dans {n} min',
      'time.inHour': 'dans {n} h',
      'time.inDay': 'dans {n} j',
      'reassign.tooltip': 'Réassigner la tâche',
      'chat.reassignedTo': 'Tâche réassignée à {name}',
      'errors.reassignFailed': 'Échec de la réassignation.',
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
      'errors.acceptFailed': 'Échec de l\'acceptation.',
      'errors.acceptConflict': 'Conflit avec la branche du chantier : demandez à l\'agent de reprendre la base.',
      'errors.cancelFailed': 'Échec de l\'annulation.',
      'errors.deleteTaskFailed': 'Échec de la suppression.',
      'errors.deleteCardFailed': 'Échec de la suppression.',
      'errors.deleteProjectFailed': 'Échec de la suppression.',
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
      'workspace.autoSync': 'Synchronisation automatique (toutes les 15 min)',
      'workspace.errorSaveFailed': 'Échec de l\'enregistrement du remote.',
      'workspace.errorActivateFailed': 'Échec de l\'activation.',
      'workspace.errorSyncFailed': 'Échec de la synchronisation.',
      'workspace.errorAutoSyncFailed': 'Échec de l\'enregistrement.',
      'preferences.title': 'Préférences',
      'preferences.displayNamePlaceholder': 'Prénom',
      'preferences.langLabel': 'Langue',
      'preferences.errorSaveFailed': 'Échec de l\'enregistrement.',
      'sidebar.settingsButton': 'Réglages',
      'sidebar.repoTooltip': 'Voir le dépôt sur GitHub',
      'sidebar.sponsorTooltip': 'Sponsoriser le projet',
      'settings.tabGeneral': 'Général',
      'settings.tabSync': 'Sauvegarde',
      'settings.tabStats': 'Statistiques',
      'usage.empty': 'Aucun projet.',
      'update.sidebarAvailable': 'Mise à jour disponible',
      'update.title': 'Mises à jour',
      'update.currentVersion': 'Version {version}',
      'update.devBuild': 'Compilation locale : rien à comparer.',
      'update.upToDate': 'Sillage est à jour.',
      'update.availableHeadline': 'Sillage {latest} est disponible (vous avez {current}).',
      'update.releaseNotes': 'Voir les nouveautés',
      'update.checkNow': 'Vérifier maintenant',
      'update.checking': 'Vérification…',
      'update.neverChecked': 'Jamais vérifié',
      'update.lastChecked': 'Vérifié {time}',
      'update.autoCheckLab': 'Vérifier les mises à jour automatiquement',
      'lan.heading': 'Depuis vos autres appareils',
      'lan.copy': 'Copier l’adresse',
      'lan.toggle': 'Ouvrir Sillage au réseau local',
      'lan.offNote': 'Sillage ne répond qu’à cet ordinateur. Activez pour le suivre depuis votre téléphone sur le même wifi.',
      'lan.noPassword': 'Il faut d’abord un mot de passe : lancez Sillage avec SILLAGE_PASSWORD. Sans lui, n’importe qui sur le wifi pourrait faire travailler vos agents.',
      'lan.noAddress': 'Aucune adresse réseau trouvée : cet ordinateur n’est connecté à aucun réseau local.',
      'lan.certNote': 'La connexion est chiffrée par un certificat créé sur cette machine : le navigateur prévient une fois. Pour vérifier, son empreinte doit être celle-ci :',
      'update.autoCheckNote': 'Une requête par jour vers GitHub, pour lire le numéro de la dernière version. Rien de votre machine ne sort.',
      'update.apply': 'Mettre à jour et redémarrer',
      'update.applying': 'Mise à jour en cours…',
      'update.applied': 'Sillage {version} installé, redémarrage…',
      'update.appliedNoRestart': 'Sillage {version} installé. Redémarrez Sillage pour l\'utiliser.',
      'update.reconnecting': 'Reconnexion à Sillage…',
      'update.manualIntro': 'À jouer dans un terminal :',
      'update.method.brew': 'Installé avec Homebrew',
      'update.method.binary': 'Binaire installé dans {dir}',
      'update.method.go': 'Installé avec go install',
      'update.method.unknown': 'Mode d\'installation inconnu',
      'update.blocker.tasksRunning': 'Un agent travaille : interrompez-le avant de mettre à jour.',
      'update.blocker.previewsRunning': 'Une recette tourne : arrêtez-la avant de mettre à jour.',
      'update.blocker.notWritable': 'Le dossier du binaire n\'est pas modifiable par Sillage.',
      'update.blocker.brewMissing': 'La commande brew est introuvable dans le PATH.',
      'update.blocker.goInstall': 'Une installation par go install se met à jour avec go.',
      'update.blocker.unknownMethod': 'Sillage ne sait pas comment ce binaire a été installé.',
      'update.errorCheckFailed': 'Impossible de joindre GitHub.',
      'update.errorApplyFailed': 'Échec de la mise à jour.',
      'update.serviceHeading': 'Démarrage',
      'update.serviceOn': 'Sillage est lancé à l\'ouverture de session.',
      'update.serviceOff': 'Sillage n\'est pas lancé à l\'ouverture de session.',
      'update.serviceFlagsNote': 'L\'instance en cours tourne avec des arguments que le service ne reprendra pas : il lance le binaire sans option.'
    },
    en: {
      'nav.inbox': 'Inbox',
      'nav.allProjects': 'All projects',
      'nav.logout': 'Log out',
      'nav.back': 'Back',
      'common.projects': 'Projects',
      'common.tasksWord': 'Tasks',
      'common.workstreamsWord': 'Workstreams',
      'common.close': 'Close',
      'common.cancel': 'Cancel',
      'panel.expand': 'Expand panel',
      'panel.collapse': 'Collapse panel',
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
      'sidebar.projectMenuTooltip': 'Project actions',
      'sidebar.markAllRead': 'Mark all as read',
      'sidebar.hideProject': 'Hide from sidebar',
      'sidebar.showProject': 'Show in sidebar',
      'sidebar.agentsHeading': 'Agents',
      'sidebar.agentsAtWork': 'At work',
      'sidebar.newAgentTooltip': 'New agent',
      'sidebar.noAgents': 'No agents',
      'aria.menu': 'Menu',
      'header.newTask': 'New task',
      'header.newTaskTooltip': 'New task (N)',
      'header.newProject': 'New project',
      'header.newProjectTooltip': 'New project (N)',
      'allProjects.emptyTitle': 'No projects yet',
      'allProjects.emptySub': 'Create your first project to get started.',
      'allProjects.cardCount.one': '{n} workstream',
      'allProjects.cardCount.other': '{n} workstreams',
      'allProjects.hiddenBadge': 'Hidden from sidebar',
      'inbox.empty': 'Inbox is empty. All caught up!',
      'kanban.addCard': '+ Add',
      'kanban.empty': 'Nothing in progress yet.',
      'kanban.emptyAction': 'Hand a first task to an agent',
      'kanban.newWorkstream': 'New workstream',
      'kanban.newTaskTooltip': 'Hand a task to an agent (N). Its workstream is created for you.',
      'kanban.card.tasksLabel': 'tasks',
      'kanban.card.noTask': 'No task',
      'kanban.card.reviewCount': '{n} to review',
      'kanban.card.awaitingShip': 'Ready to ship',
      'inbox.awaitingShip.title': 'Workstreams ready to ship',
      'cardMenu.moveTo': 'Move to {column}',
      'project.editTooltip': 'Edit project',
      'project.editTitle': 'Project settings',
      'project.tabGeneral': 'General',
      'project.tabRepos': 'Repositories',
      'project.tabInstructions': 'Instructions',
      'project.tabPermissions': 'Permissions',
      'project.tabDelivery': 'Delivery',
      'project.tabLinks': 'Links',
      'project.tabDanger': 'Delete',
      'project.name': 'Name',
      'project.baseBranch': 'Base branch',
      'project.baseBranchHint': 'Workstreams branch off this branch and ship back into it. Empty: the repository default branch.',
      'project.checkCmd': 'Check command',
      'project.permissionProfilesHint': 'Enable only technologies you trust for this project. Each profile allows its commands and moves downloads to a writable Sillage cache, outside the repository and system directories.',
      'project.profile.ruby': 'Ruby',
      'project.profile.rubyDesc': 'bundle, gem, rake, and ruby · gems are never installed in /usr/local',
      'project.profile.node': 'Node.js',
      'project.profile.nodeDesc': 'node, npm, npx, pnpm, and yarn · isolated package caches',
      'project.profile.php': 'PHP',
      'project.profile.phpDesc': 'php and Composer · isolated Composer cache and tools',
      'project.profile.python': 'Python',
      'project.profile.pythonDesc': 'python, pip, pytest, ruff, and uv · isolated package caches',
      'project.profile.go': 'Go',
      'project.profile.goDesc': 'go, gofmt, and golangci-lint · isolated build and module caches',
      'project.profile.rust': 'Rust',
      'project.profile.rustDesc': 'cargo, rustc, rustfmt, and clippy · isolated registry and artifacts',
      'project.allowedTools': 'Additional Claude rules',
      'project.allowedToolsPlaceholder': 'Bash(go test:*)\nBash(gofmt:*)',
      'project.allowedToolsHint': 'Advanced option, one rule per line using Claude CLI syntax. Rules from checked profiles are applied automatically without being copied here. Nothing able to push will ever be accepted.',
      'project.reposHint': 'Local paths of this project\'s git repositories.',
      'project.repoNamePlaceholder': 'Repository name',
      'project.repoName': 'Name',
      'project.repoPath': 'Path',
      'project.addRepo': '+ repo',
      'project.removeRepo': 'Remove',
      'project.previewCmdPlaceholder': 'Preview command (optional)',
      'project.previewUrlPlaceholder': 'URL to open (optional)',
      'project.previewHintIntro': 'Manual preview: the command runs in the workstream or task worktree.',
      'project.previewVarIdDesc': 'names: database, container, directory',
      'project.previewVarNDesc': 'explicit arithmetic',
      'project.previewVarPortDesc': 'ready-to-use port, no arithmetic to write',
      'project.previewVarDirDesc': 'absolute paths',
      'project.previewVarBranchDesc': 'display, debugging',
      'project.previewExampleLabel': 'For example, to launch a Python server:',
      'project.errorNameRequired': 'Name is required.',
      'project.errorReposRequired': 'At least one repository is required.',
      'project.errorSaveFailed': 'Failed to save.',
      'project.delete': 'Delete the project',
      'project.deleteConfirm': 'Confirm deletion?',
      'project.deleteSubtext': 'Also deletes {cards} workstreams and {tasks} tasks.',
      'project.deleteWarning': 'Existing worktrees and branches are left untouched, but the project disappears from Sillage. This cannot be undone.',
      'project.unsaved': 'Unsaved changes',
      'project.linksHint': 'Shortcuts shown on the project page, 12 at most.',
      'project.addLink': '+ link',
      'project.linksEmpty': 'No pinned links.',
      'project.linksInvalidUrl': 'Use an http(s) URL.',
      'project.linksMax': 'Maximum 12 links.',
      'project.description': 'Description',
      'project.descriptionPlaceholder': 'A sentence describing the project',
      'project.contextPrompt': 'Context for agents',
      'project.contextPromptPlaceholder': 'Conventions, architecture, constraints to know…',
      'project.instructionsHint': 'Passed to every agent of the project, on top of the workstream context.',
      'work.emptyCard': 'No tasks yet.',
      'work.emptyCardAction': 'Create the first task',
      'work.emptyFiltered': 'No tasks match this filter.',
      'filter.all': 'All {n}',
      'filter.waiting': 'Waiting {n}',
      'filter.running': 'In progress {n}',
      'filter.review': 'To review {n}',
      'filter.finished': 'Handled {n}',
      'badge.new': 'NEW',
      'status.waiting': 'Waiting',
      'status.running': 'In progress',
      'status.review': 'To review',
      'status.rebasing': 'Rebasing',
      'status.rebasingNote': 'The task is being replayed on top of the work just accepted.',
      'status.accepted': 'Accepted',
      'status.cancelled': 'Refused',
      'action.interrupt': 'Stop the agent',
      'action.startNow': 'Start now',
      'action.accept': 'Accept',
      'action.accepting': 'Accepting…',
      'action.acceptTooltip': 'Merge into the workstream branch',
      'action.refuse': 'Refuse',
      'action.refuseTooltip': 'Leave this task out of the workstream',
      'action.reopen': 'Reopen the task',
      'taskRow.waitingFor': 'Waiting for "{title}"',
      'waitingFor.note': 'Will start automatically once "{title}" is accepted.',
      'waitingFor.unknownTask': 'The task it was waiting for no longer exists: start it manually.',
      'newTask.waitsForLabel': 'Start after',
      'newTask.waitsForNone': 'Immediately',
      'behind.taskTooltip.one': '1 commit behind {base}: rebase before accepting, or it will conflict.',
      'behind.taskTooltip.other': '{n} commits behind {base}: rebase before accepting, or it will conflict.',
      'behind.cardTooltip.one': 'The workstream is 1 commit behind {base}.',
      'behind.cardTooltip.other': 'The workstream is {n} commits behind {base}.',
      'behind.cardLabel.one': '↓ 1 commit behind {base}',
      'behind.cardLabel.other': '↓ {n} commits behind {base}',
      'behind.askRebase': 'Ask for a rebase',
      'behind.askRebaseTooltip': 'Send the agent a message asking it to rebase its branch onto {base}',
      'behind.rebasePrompt.one': 'Your branch is 1 commit behind {base}. Rebase onto it, settle any conflicts, then run the project checks again.',
      'behind.rebasePrompt.other': 'Your branch is {n} commits behind {base}. Rebase onto it, settle any conflicts, then run the project checks again.',
      'action.cancelTask': 'Cancel the task',
      'action.cancelTaskConfirm': 'Confirm cancellation?',
      'task.delete': 'Delete the task',
      'task.deleteConfirm': 'Confirm deletion?',
      'task.moreActions': 'More actions',
      'task.filesCount.one': '{n} file',
      'task.filesCount.other': '{n} files',
      'ship.button': 'Ship',
      'ship.subtext.pr': 'Pushes {branch} and opens a pull request against {base}',
      'ship.subtext.mr': 'Pushes {branch} and opens a merge request against {base}',
      'ship.subtext.push': 'Pushes {branch} to origin, without opening a pull request',
      'ship.subtext.unknownForge': 'Pushes {branch} (unknown forge: no pull request)',
      'ship.subtext.merge': 'Merges {branch} into {base} locally, never pushing',
      'ship.subtext.mergePush': 'Merges {branch} into {base}, then pushes {base}',
      'ship.subtext.repos': '{n} repositories involved',
      'ship.subtext.nothing': 'Nothing to ship yet',
      'ship.partial.one': 'Partial delivery: 1 task is not accepted yet',
      'ship.partial.other': 'Partial delivery: {n} tasks are not accepted yet',
      'ship.partialNote.one': '1 task is still running or waiting for review: only accepted work ships now. You will be able to ship the rest later.',
      'ship.partialNote.other': '{n} tasks are still running or waiting for review: only accepted work ships now. You will be able to ship the rest later.',
      'ship.blocked.noTasks': 'No task in this workstream',
      'ship.blocked.nothingAccepted': 'No accepted task',
      'ship.blocked.nothingToShip': 'No branch to ship',
      'ship.blocked.behindTarget': '{base} has moved on: catch up with it before shipping (merging is fast-forward only).',
      'ship.alreadyOnTarget': 'Already on {base}',
      'ship.alreadyOnTargetSub': 'All of the workstream\'s work has landed in {base}: nothing left to ship.',
      'catchUp.button': 'Catch up with {base}',
      'catchUp.tooltip': 'Merges {base} into the workstream branch, without rewriting its history. That is what makes shipping possible again.',
      'catchUp.done': 'Workstream caught up with {base}.',
      'catchUp.upToDate': 'The workstream was already up to date.',
      'catchUp.conflictTitle': 'Conflict with {base}',
      'catchUp.conflictTitleMulti': 'Catch-up conflicts',
      'catchUp.conflictBody': 'Merging {base} into the workstream clashes on {files}. Nothing was changed: the workstream is intact. An agent can pick up the new base and settle the conflicts in a dedicated task.',
      'catchUp.askAgent': 'Hand it to an agent',
      'catchUp.errorFailed': 'Catching up failed.',
      'catchUp.taskTitle': 'Catch up with {base}',
      'catchUp.taskPrompt': 'This workstream\'s branch is behind {base} and can no longer be merged: {files} conflicts. Merge {base} into the workstream branch, settle the conflicts keeping both intents, then run the project checks again.',
      'ship.modalTitle': 'Ship the workstream',
      'ship.modalConfirm': 'Ship',
      'ship.shipping': 'Shipping…',
      'ship.repoNothing': 'Nothing to ship',
      'ship.repoShipped': 'Already shipped',
      'ship.commits.one': '{n} new commit',
      'ship.commits.other': '{n} new commits',
      'ship.files.one': '{n} file',
      'ship.files.other': '{n} files',
      'ship.prLink': 'View the pull request',
      'ship.mergedInto': 'Merged into {base}',
      'ship.mergedPushed': 'Merged into {base} and pushed',
      'ship.pushedBranch': 'Branch {branch} pushed',
      'ship.errorFailed': 'Failed to ship.',
      'preview.button': 'Preview',
      'preview.buttonCard': 'Workstream preview',
      'preview.tooltip': 'Run this workstream to try it out',
      'preview.taskTooltip': 'Run this task to try it out',
      'preview.cardTitle': 'Preview: {title}',
      'preview.taskTitle': 'Preview of task #{ref}',
      'preview.allTitle': 'Running previews',
      'preview.taskScope': 'You are previewing task #{ref}, not the workstream.',
      'preview.statusRunning': 'running',
      'preview.statusStopped': 'stopped',
      'preview.statusExited': 'finished (exit {code})',
      'preview.statusFailed': 'failed to start',
      'preview.statusIdle': 'never started',
      'preview.start': 'Start',
      'preview.stop': 'Stop',
      'preview.restart': 'Restart',
      'preview.showLog': 'Log',
      'preview.noCmd': 'No preview command',
      'preview.noCmdHint': 'Add one in the project settings, per repository. In the meantime, the worktree path is right here.',
      'preview.openSettings': 'Open project settings',
      'preview.noBranch': 'No workstream branch yet: create a task first.',
      'preview.worktree': 'Worktree',
      'preview.copyPath': 'Copy path',
      'preview.copied': 'Copied',
      'preview.logEmpty': 'No output yet.',
      'preview.errorStartFailed': 'Failed to start.',
      'preview.errorStopFailed': 'Failed to stop.',
      'preview.running.one': '{n} preview running',
      'preview.running.other': '{n} previews running',
      'delivery.label': 'Shipping means',
      'delivery.modePr': 'Open a pull request',
      'delivery.modePush': 'Push the branch',
      'delivery.modeMerge': 'Merge locally',
      'delivery.modeMergePush': 'Merge, then push',
      'delivery.targetPlaceholder': 'repository default branch',
      'delivery.mergeNote': 'Local fast-forward merge only: Sillage never pushes that branch.',
      'delivery.mergePushNote': 'The target branch is first caught up from origin, fast-forwarded, then pushed. Never forced.',
      'delivery.prNote': 'The workstream branch is pushed, then the pull request (or merge request) is opened.',
      'delivery.pushNote': 'The workstream branch is pushed to origin. Nothing is opened, nothing is merged.',
      'delivery.warning.ghMissing': 'gh not found in PATH: falling back to a prefilled pull request URL.',
      'delivery.warning.ghInstallLink': 'Install the GitHub CLI',
      'delivery.warning.glabMissing': 'glab not found in PATH: falling back to a prefilled merge request URL.',
      'delivery.warning.glabInstallLink': 'Install the GitLab CLI',
      'delivery.warning.noRemote': 'One repository has no "origin" remote: nothing can be pushed.',
      'delivery.warning.unknownForge': 'Unknown forge: the branch will be pushed without opening a pull request.',
      'tabs.conversation': 'Conversation',
      'tabs.diff': 'Diff',
      'tabs.deliverables': 'Deliverables',
      'tabs.history': 'History',
      'chat.you': 'You',
      'chat.placeholder': 'Reply to {name}…',
      'chat.send': 'Send ⏎',
      'chat.working': 'Working…',
      'chat.jumpToBottom': 'Jump to latest',
      'chat.newMessage': 'New message',
      'chat.copyMessage': 'Copy message',
      'chat.copyCode': 'Copy code',
      'chat.copied': 'Copied',
      'chat.attach': 'Attach an image',
      'chat.attachRemove': 'Remove image',
      'chat.attachTooMany': 'Six images per message at most.',
      'chat.attachTooLarge': '{name}: image too large (8 MB maximum).',
      'chat.attachUnsupported': '{name}: unsupported format (PNG, JPEG, GIF or WebP).',
      'chat.attachFailed': 'Unreadable image: {name}.',
      'chat.imageOpen': 'View full size',
      'chat.accepted': 'Accepted: merged into {branch}',
      'chat.autoAccepted': 'Accepted: its work was already in {branch}',
      'chat.mergeConflict': 'Conflict with the workstream branch on {files}: ask the agent to rebase on it.',
      'chat.rebased': 'Rebased automatically onto {branch}: this task now starts from the accepted work.',
      'chat.rebaseConflict': 'Automatic rebase not possible on {files} (nothing was changed): ask the agent to rebase on the workstream branch.',
      'chat.toolDenied': 'Tool denied to the agent: {tool}. To allow it, add it to the allowed tools in the project settings.',
      'chat.interrupted': 'Agent interrupted by a Sillage shutdown: the task went back to review. What it had committed is there; send it a message to pick the work back up.',
      'conversation.empty': 'No messages yet.',
      'diff.empty': 'No changes.',
      'diff.openDifftool': 'Open in diff tool',
      'diff.openIn': 'Open in {tool}',
      'diff.openAllHint': 'Every change at once, side by side',
      'diff.noToolHint': 'No diff tool set in git: git config --global diff.tool meld',
      'diff.openFile': 'Open this file in {tool}',
      'diff.files.one': '{n} file',
      'diff.files.other': '{n} files',
      'diff.collapseAll': 'Collapse all',
      'diff.expandAll': 'Expand all',
      'diff.collapse': 'Collapse',
      'diff.expand': 'Expand',
      'diff.new': 'new',
      'diff.deleted': 'deleted',
      'diff.generated': 'generated',
      'diff.skipped.one': '{n} unchanged line',
      'diff.skipped.other': '{n} unchanged lines',
      'deliverables.code': 'Code',
      'deliverables.docs': 'Documents',
      'deliverables.images': 'Screenshots',
      'deliverables.empty': 'No items.',
      'history.empty': 'No commands yet.',
      'newTask.title': 'New task',
      'newTask.titlePlaceholder': 'What should the agent do?',
      'newTask.promptPlaceholder': 'Description or detailed instructions (optional)',
      'newTask.attachHint': 'Attach an image, or paste one (Ctrl+V)',
      'newTask.agentLabel': 'Agent',
      'newTask.repoLabel': 'Repository',
      'newTask.projectContextNote': '+ project context',
      'newTask.workstreamContextNote': '+ workstream context',
      'newTask.submit': 'Create and chat',
      'newTask.submitAnother': 'Create and add another',
      'newTask.submitTooltip': 'Create and open the conversation ({mod}+Enter)',
      'newTask.submitAnotherTooltip': 'Create and keep the form open ({mod}+Shift+Enter)',
      'newTask.created': 'Task #{ref} created: {title}',
      'newTask.agentHint': 'Arrow keys to switch agent',
      'newTask.errorTitleRequired': 'A title is required.',
      'newTask.errorAgentRequired': 'Choose an agent.',
      'newTask.errorCreateFailed': 'Failed to create.',
      'shortcuts.title': 'Keyboard shortcuts',
      'shortcuts.hint': 'Shortcuts: ?',
      'shortcuts.sectionGlobal': 'Anywhere',
      'shortcuts.sectionForm': 'In a form',
      'shortcuts.sectionSearch': 'In search',
      'shortcuts.sectionTask': 'In a task',
      'shortcuts.search': 'Search',
      'shortcuts.create': 'Create (task, workstream or project, depending on the screen)',
      'shortcuts.help': 'Show this help',
      'shortcuts.escape': 'Close / go back',
      'shortcuts.submit': 'Submit',
      'shortcuts.submitAnother': 'Submit and add another',
      'shortcuts.pickAgent': 'Switch agent',
      'shortcuts.tab': 'Move between fields (stays inside the form)',
      'shortcuts.searchNav': 'Move through results',
      'shortcuts.searchOpen': 'Open the result',
      'shortcuts.sendMessage': 'Send the message',
      'shortcuts.expandPanel': 'Expand or collapse the panel',
      'shortcuts.diffNav': 'Next / previous file in the diff',
      'newProject.title': 'New project',
      'newProject.pathLabel': 'Path to a git repository',
      'newProject.pathPlaceholder': '/home/user/projects/my-project',
      'newProject.hint': 'The project name and delivery mode are derived from it. Everything else is adjusted later, in the settings.',
      'newProject.errorPathRequired': 'The path to a git repository is required.',
      'newProject.errorCreateFailed': 'Failed to create the project.',
      'newCard.title': 'New workstream',
      'newCard.titleLabel': 'Workstream title',
      'newCard.titlePlaceholder': 'Workstream title',
      'newCard.errorTitleRequired': 'A title is required.',
      'newCard.errorCreateFailed': 'Failed to create the workstream.',
      'newCard.submitAndPlan': 'Create and plan',
      'newCard.submitAndPlanTooltip': 'Create the workstream, then ask an AI for a task plan',
      'plan.button': 'Task plan',
      'plan.buttonTooltip': 'Ask an AI to split this workstream into steps',
      'plan.title': 'Task plan',
      'plan.intro': 'The agent reads the repository and proposes a split into steps. It writes nothing: nothing is created until you approve it.',
      'plan.extraLabel': 'Details (optional)',
      'plan.extraPlaceholder': 'Where to start, what the split must respect, what it must not touch…',
      'plan.repoLabel': 'Repository to read',
      'plan.agentLabel': 'Agent',
      'plan.agentHint': 'It will plan, then do the work',
      'plan.propose': 'Propose a plan',
      'plan.proposing': 'Reading the repository…',
      'plan.waitNote': 'The agent is reading the repository. This often takes a minute or two.',
      'plan.errorNoAgent': 'No agent can plan. Claude and Codex agents can, and so can the test agent.',
      'plan.errorAgentRequired': 'Pick an agent.',
      'plan.errorFailed': 'The plan did not come through.',
      'plan.stepsTitle.one': '1 step proposed',
      'plan.stepsTitle.other': '{n} steps proposed',
      'plan.stepTitlePlaceholder': 'Task title',
      'plan.stepPromptPlaceholder': 'What the agent has to do',
      'plan.removeStep': 'Remove this step',
      'plan.chainLabel': 'Chain them: each task starts once the previous one is accepted',
      'plan.chainHint': 'Unchecked, every task starts at once.',
      'plan.again': 'Start over',
      'plan.create.one': 'Create the task',
      'plan.create.other': 'Create the {n} tasks',
      'plan.creating': 'Creating…',
      'plan.errorEmpty': 'No step left to create.',
      'plan.errorPartial': '{n} task(s) created, then an error: {msg}',
      'issues.import': 'From a GitHub issue',
      'issues.importTooltip': 'Pick an open issue from the repository: it fills the title and the brief',
      'issues.title': 'Open issues',
      'issues.searchPlaceholder': 'Search: free text, label:bug, author:…',
      'issues.repoLabel': 'Repository',
      'issues.loading': 'Reading the issues…',
      'issues.empty': 'No open issue on this repository.',
      'issues.emptySearch': 'No issue matches this search.',
      'issues.errorFailed': 'The issues could not be read.',
      'issues.cap': 'The {n} most recent issues. Refine the search to go further.',
      'issues.by': 'by {login}',
      'issues.promptHeader': 'Fix GitHub issue #{n}: {title}',
      'nav.vigie': 'Lookout',
      'vigie.check': 'Check now',
      'vigie.checking': 'Checking…',
      'vigie.checkedAgo': 'checked {when}',
      'vigie.notify': 'Turn on notifications',
      'vigie.introTitle': 'The lookout reads for you',
      'vigie.introSub': 'Pick the projects to watch. Every new pull request or issue is read by an agent, in read-only mode, who drafts a verdict and a review for you. Nothing is ever posted to GitHub.',
      'vigie.noGithub': 'None of your projects has a repository on GitHub.',
      'vigie.emptyTitle': 'All quiet',
      'vigie.emptySub': 'New pull requests and issues from watched projects will land here, already read.',
      'vigie.watchedHeading': 'Projects on GitHub',
      'vigie.watch': 'Watch',
      'vigie.unwatch': 'Stop watching',
      'vigie.analyst': 'Read by',
      'vigie.kind.pr': 'PR',
      'vigie.kind.issue': 'Issue',
      'vigie.kind.upkeep': 'Suggestion',
      'vigie.detail.kindUpkeep': 'Suggestion',
      'vigie.decision.upgrade': 'Worth considering',
      'vigie.detail.plan': 'What to do',
      'vigie.detail.foundBy': 'Found by the upkeep check {when}',
      'vigie.detail.discussion': 'Discussion on GitHub',
      'vigie.newComments.one': '{n} new comment',
      'vigie.newComments.other': '{n} new comments',
      'vigie.upkeep.check': 'Check for updates',
      'vigie.upkeep.last': 'updates checked {when}',
      'vigie.upkeep.running': '{agent} is checking whether {repo} needs updates',
      'vigie.upkeep.promptHeader': 'Upkeep of the {repo} repository: {title}.',
      'vigie.log.upkeep': 'Upkeep of {repo} checked, new findings: {n}',
      'vigie.log.upkeepFailed': 'Could not check the upkeep of {repo}: {error}',
      'vigie.log.updated': 'reopened by a comment: {n}',
      'vigie.by': 'by {login}',
      'vigie.high': 'Priority',
      'vigie.queued': 'Waiting for analysis',
      'vigie.stopped': 'Review stopped.',
      'vigie.queuedNext': 'Queued: next in line',
      'vigie.queuedAhead.one': 'Queued: {n} review ahead',
      'vigie.queuedAhead.other': 'Queued: {n} reviews ahead',
      'vigie.queuedPaused': 'Waiting: the watch is off for this project',
      'vigie.queuedHint': 'Reviews run one at a time, so ten agents never start on your machine at once.',
      'vigie.live.title': '{agent} is reading',
      'vigie.live.diff': 'Fetching the diff from GitHub…',
      'vigie.live.start': 'Starting the agent…',
      'vigie.live.limit': 'gives up after {n} min',
      'vigie.live.stop': 'Stop',
      'sessions.pill.active': 'A Claude session is running in this project',
      'sessions.pill.recent.one': '{n} recent Claude session',
      'sessions.pill.recent.other': '{n} recent Claude sessions',
      'sessions.title': 'Claude Code sessions',
      'sessions.hint': 'Sessions started by hand in this project\'s folders, the ones "claude --resume" lists. Sillage shows them without touching them.',
      'sessions.active': 'running',
      'sessions.prompts.one': '{n} request',
      'sessions.prompts.other': '{n} requests',
      'sessions.copy': 'Copy the resume command',
      'sessions.copied': 'Command copied',
      'sessions.workstream': 'Turn into a workstream',
      'sessions.allPrompts': 'Its latest requests',
      'sessions.empty': 'No recent Claude Code session in this project\'s folders.',
      'sessions.promptHeader': 'Pick up the work of a Claude Code session run outside Sillage, "{title}".',
      'sessions.promptBranch': 'It worked in {cwd}, on branch {branch}. Whatever was not committed there is not in this worktree.',
      'sessions.promptAsks': 'Its latest requests, oldest first:',
      'sessions.shipWarning': 'A Claude Code session is open in {repo} ("{title}"): it may be writing where this delivery goes.',
      'autoconf.title': 'Automatic setup',
      'autoconf.hint': 'An agent reads the repository (build files, CI, README) and suggests the check run after every task, the preview command and its address. It changes nothing, and you review before saving.',
      'autoconf.start': 'Start',
      'autoconf.with': 'Agent',
      'autoconf.running': '{agent} is reading {repo}',
      'autoconf.runningShort': '{agent} is setting up the project',
      'autoconf.readyShort': 'Setup suggested, ready to review',
      'autoconf.stop': 'Stop',
      'autoconf.ready': 'Suggested by {agent}',
      'autoconf.apply': 'Use these values',
      'autoconf.applied': 'Values copied into the tabs. Review them, then save.',
      'autoconf.dismiss': 'Discard',
      'autoconf.retry': 'Run again',
      'autoconf.error': 'Failed: {error}',
      'autoconf.stopped': 'Stopped.',
      'autoconf.check': 'Check after every task',
      'autoconf.preview': 'Preview',
      'autoconf.url': 'Address',
      'autoconf.profiles': 'Permission profiles',
      'autoconf.found': 'Found',
      'autoconf.notes': 'Good to know',
      'autoconf.install': 'install',
      'autoconf.build': 'build',
      'autoconf.test': 'tests',
      'autoconf.current': 'currently: {v}',
      'autoconf.nothing': 'nothing suggested',
      'autoconf.warning': 'These commands will run on your machine: review them before saving.',
      'autoconf.noAgent': 'This needs an agent that can read without writing (Claude Code or Codex).',
      'autoconf.nudge': 'Sillage does not know yet how to check or run this project.',
      'autoconf.nudgeAction': 'Set up automatically',
      'vigie.analyzing': 'Analyzing…',
      'vigie.error': 'Analysis failed: {error}',
      'vigie.copyReview': 'Copy comments',
      'vigie.copied': 'Copied, paste them on GitHub',
      'inbox.signals.title': 'On GitHub',
      'sidebar.agentsToSetup.one': '1 agent to install',
      'sidebar.agentsToSetup.other': '{n} agents to install',
      'vigie.githubAction.merge': 'Merge on GitHub',
      'vigie.githubAction.close': 'Close on GitHub',
      'vigie.githubAction.changes': 'Open on GitHub',
      'vigie.githubAction.rework': 'Open on GitHub',
      'vigie.githubAction.needs-info': 'Open on GitHub',
      'vigie.githubAction.fix': 'Open on GitHub',
      'vigie.githubAction.feature': 'Open on GitHub',
      'vigie.githubAction.undefined': 'Open on GitHub',
      'vigie.githubAction.': 'Open on GitHub',
      'lane.next': 'Up next',
      'lane.agents': 'Agents at work',
      'lane.you': 'Your turn',
      'lane.shipped': 'Shipped',
      'vigie.detail.review': 'Suggested review',
      'vigie.detail.reply': 'Suggested reply',
      'vigie.detail.nothingToAdd': 'Nothing to add: the PR can be merged as is.',
      'vigie.detail.prBody': 'PR description, by {login}',
      'vigie.detail.issueBody': 'The issue, by {login}',
      'hello.day': 'Hello',
      'hello.evening': 'Good evening',
      'hello.and': 'and',
      'hello.waiting': 'You have {list}.',
      'hello.nothing': 'Nothing is waiting for you.',
      'hello.review.one': '{n} task to review',
      'hello.review.other': '{n} tasks to review',
      'hello.signals.one': '{n} GitHub item to decide on',
      'hello.signals.other': '{n} GitHub items to decide on',
      'hello.ship.one': '{n} workstream to ship',
      'hello.ship.other': '{n} workstreams to ship',
      'hello.running.one': 'One agent is working for you.',
      'hello.running.other': '{n} agents are working for you.',
      'pulse.you.one': '{n} for you',
      'pulse.you.other': '{n} for you',
      'pulse.agents.one': '{n} agent at work',
      'pulse.agents.other': '{n} agents at work',
      'pulse.calm': 'All quiet',
      'pulse.empty': 'Nothing yet',
      'vigie.detail.kindPr': 'Pull request',
      'vigie.detail.kindIssue': 'Issue',
      'vigie.detail.opened': 'opened {when}',
      'vigie.detail.copy': 'Copy',
      'vigie.detail.onGithub': 'Open on GitHub',
      'vigie.detail.readBy': 'Read by {agent}',
      'vigie.detail.analyzing': '{agent} is reading…',
      'vigie.detail.waiting': 'The verdict arrives as soon as the read is done.',
      'vigie.detail.noBody': 'The author wrote nothing.',
      'vigie.openGithub': 'Open on GitHub',
      'vigie.workstream': 'Hand to an agent',
      'vigie.toCard': 'Turn into a workstream',
      'vigie.toCardHint': 'A workstream, then an agent splits it into tasks you review',
      'vigie.dismiss': 'Remove from list',
      'vigie.retry': 'Retry',
      'vigie.localOnly': 'Sillage never posts anything to GitHub. Suggested comments stay here: copy and paste them yourself if you agree.',
      'vigie.promptAnalysis': 'Prior analysis from the lookout:',
      'vigie.log.error': 'Could not read {repo}: {error}',
      'vigie.log.analysisFailed': 'Analysis of {kind} #{n} ({repo}) failed: {error}',
      'vigie.log.analysis': '{kind} #{n} ({repo}) analyzed in {s} s',
      'vigie.log.older': 'opened before {date}, skipped: {n}',
      'vigie.log.drafts': 'drafts skipped: {n}',
      'vigie.log.known': 'already seen: {n}',
      'vigie.log.list': '{repo} · open {kind}: {open}, new: {added}',
      'vigie.log.empty': 'No check since Sillage started.',
      'vigie.log.hide': 'Hide log',
      'vigie.log.show': 'Show log',
      'vigie.replyNote': 'Reply drafted by the agent, not posted. "Copy reply" puts it on your clipboard so you can paste it on the issue.',
      'vigie.copyReply': 'Copy reply',
      'vigie.decision.needs-info': 'Needs more information',
      'vigie.decision.feature': 'Feature request',
      'vigie.decision.fix': 'Bug to fix',
      'vigie.decision.close': 'Should be closed',
      'vigie.decision.rework': 'Needs rework',
      'vigie.decision.changes': 'Minor changes needed',
      'vigie.decision.merge': 'Ready to merge',
      'vigie.neverChecked': 'never checked',
      'vigie.reviewNote': 'Comments drafted by the agent, not posted. "Copy comments" puts them on your clipboard so you can paste them into the GitHub review.',
      'workstream.editTooltip': 'Edit workstream',
      'workstream.editTitle': 'Workstream',
      'workstream.errorSaveFailed': 'Failed to save.',
      'workstream.delete': 'Delete the workstream',
      'workstream.deleteConfirm': 'Confirm deletion?',
      'workstream.deleteSubtext.one': 'Also deletes its task.',
      'workstream.deleteSubtext.other': 'Also deletes its {n} tasks.',
      'agent.newTitle': 'New agent',
      'agent.editTitle': 'Edit agent',
      'agent.name': 'Name',
      'agent.emoji': 'Emoji',
      'agent.color': 'Color',
      'agent.cli': 'CLI',
      'agent.cli.claude': 'Claude Code (claude)',
      'agent.cli.codex': 'OpenAI Codex (codex)',
      'agent.cli.copilot': 'GitHub Copilot (copilot)',
      'agent.cli.agy': 'Google Antigravity (agy)',
      'agent.cli.kiro': 'Kiro CLI (kiro)',
      'agent.cli.fake': 'Test agent (fake)',
      'agent.model': 'Model',
      'agent.contextPrompt': 'Context prompt',
      'agent.delete': 'Delete',
      'agent.deleteConfirm': 'Confirm deletion?',
      'agent.errorNameRequired': 'Name is required.',
      'agent.errorSaveFailed': 'Failed to save.',
      'agent.errorDeleteFailed': 'Failed to delete.',
      'agent.warning.codexSandbox': 'Codex needs unprivileged user namespaces for its sandbox, and AppArmor blocks them on this machine. Tasks assigned to this agent cannot start until this is resolved.',
      'agent.warning.copyCmd': 'Copy command',
      'agent.warning.codexSandboxFallback': 'You can also work around this by starting Sillage with the SILLAGE_CODEX_SANDBOX=danger-full-access environment variable; the remaining containment is then Sillage\'s own (dedicated worktree, no push from the agent).',
      'agent.warning.codexSandboxLink': 'Learn more (OpenAI Codex documentation)',
      'agent.warning.agyPolicy': 'Antigravity asks for your approval before every command and for some files, and nobody can give it while an agent works on its own: approval is denied outright and the task ends without producing anything. It needs two settings: an execution policy that trusts the sandbox, and permission to read and write inside the worktrees.',
      'agent.warning.agyPolicyHint': 'Sillage can set both in ~/.gemini/antigravity-cli/settings.json. Commands will then run without asking, but only inside the sandbox, which Sillage always forces on this agent; the file permissions stop at the worktrees folder, which is where it works. Your other Antigravity settings are kept.',
      'agent.warning.agyPolicyLink': 'Open the Antigravity CLI documentation',
      'agent.warning.agyPolicyFix': 'Set it for me',
      'agent.warning.agyPolicyFixConfirm': 'Write to settings.json?',
      'agent.warning.agyPolicyFixed': 'Set. Antigravity can work inside its sandbox.',
      'agent.warning.kiroApiKey': 'Kiro CLI is installed, but its headless mode requires an API key in the KIRO_API_KEY environment variable. Tasks assigned to this agent cannot start without it.',
      'agent.warning.kiroApiKeyHint': 'Add KIRO_API_KEY to the environment that starts Sillage, then restart Sillage. The key is neither displayed nor stored in the workspace.',
      'agent.warning.kiroApiKeyLink': 'Open the Kiro CLI authentication guide',
      'agent.warning.cliNotFound': 'Agent not connected: {cli} CLI not found in PATH.',
      'agent.warning.installHint': 'Install the CLI, then restart Sillage:',
      'agent.warning.installLink': 'Open the installation guide',
      'agent.quotaTitle': 'Quota',
      'agent.quotaWindow5h': '5 hours',
      'agent.quotaWindowWeek': 'Week',
      'agent.quotaUsedPercent': '{percent}% used',
      'agent.quotaResetsIn': 'resets {time}',
      'agent.quotaUpdatedAt': 'Updated: {time}',
      'agent.quotaUnavailable': 'Quotas unavailable for this agent (the {cli} CLI doesn\'t expose them).',
      'time.inMin': 'in {n} min',
      'time.inHour': 'in {n} h',
      'time.inDay': 'in {n} d',
      'reassign.tooltip': 'Reassign the task',
      'chat.reassignedTo': 'Task reassigned to {name}',
      'errors.reassignFailed': 'Failed to reassign.',
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
      'errors.acceptFailed': 'Failed to accept.',
      'errors.acceptConflict': 'Conflict with the workstream branch: ask the agent to rebase on it.',
      'errors.cancelFailed': 'Failed to cancel.',
      'errors.deleteTaskFailed': 'Failed to delete.',
      'errors.deleteCardFailed': 'Failed to delete.',
      'errors.deleteProjectFailed': 'Failed to delete.',
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
      'workspace.autoSync': 'Auto-sync every 15 minutes',
      'workspace.errorSaveFailed': 'Failed to save the remote.',
      'workspace.errorActivateFailed': 'Failed to activate.',
      'workspace.errorSyncFailed': 'Failed to sync.',
      'workspace.errorAutoSyncFailed': 'Failed to save.',
      'preferences.title': 'Preferences',
      'preferences.displayNamePlaceholder': 'First name',
      'preferences.langLabel': 'Language',
      'preferences.errorSaveFailed': 'Failed to save.',
      'sidebar.settingsButton': 'Settings',
      'sidebar.repoTooltip': 'View the repository on GitHub',
      'sidebar.sponsorTooltip': 'Sponsor the project',
      'settings.tabGeneral': 'General',
      'settings.tabSync': 'Backup',
      'settings.tabStats': 'Statistics',
      'usage.empty': 'No projects yet.',
      'update.sidebarAvailable': 'Update available',
      'update.title': 'Updates',
      'update.currentVersion': 'Version {version}',
      'update.devBuild': 'Local build: nothing to compare.',
      'update.upToDate': 'Sillage is up to date.',
      'update.availableHeadline': 'Sillage {latest} is available (you run {current}).',
      'update.releaseNotes': 'See what changed',
      'update.checkNow': 'Check now',
      'update.checking': 'Checking…',
      'update.neverChecked': 'Never checked',
      'update.lastChecked': 'Checked {time}',
      'update.autoCheckLab': 'Check for updates automatically',
      'lan.heading': 'From your other devices',
      'lan.copy': 'Copy address',
      'lan.toggle': 'Open Sillage to the local network',
      'lan.offNote': 'Sillage only answers this computer. Turn this on to follow it from your phone on the same wifi.',
      'lan.noPassword': 'A password comes first: start Sillage with SILLAGE_PASSWORD. Without one, anyone on the wifi could put your agents to work.',
      'lan.noAddress': 'No network address found: this computer is not on a local network.',
      'lan.certNote': 'The connection is encrypted with a certificate made on this machine, so the browser warns once. To check it, its fingerprint must be:',
      'update.autoCheckNote': 'One request a day to GitHub, to read the latest version number. Nothing from your machine leaves it.',
      'update.apply': 'Update and restart',
      'update.applying': 'Updating…',
      'update.applied': 'Sillage {version} installed, restarting…',
      'update.appliedNoRestart': 'Sillage {version} installed. Restart Sillage to run it.',
      'update.reconnecting': 'Reconnecting to Sillage…',
      'update.manualIntro': 'Run this in a terminal:',
      'update.method.brew': 'Installed with Homebrew',
      'update.method.binary': 'Binary installed in {dir}',
      'update.method.go': 'Installed with go install',
      'update.method.unknown': 'Unknown installation method',
      'update.blocker.tasksRunning': 'An agent is working: interrupt it before updating.',
      'update.blocker.previewsRunning': 'A preview is running: stop it before updating.',
      'update.blocker.notWritable': 'Sillage cannot write to the binary\'s directory.',
      'update.blocker.brewMissing': 'The brew command is not in the PATH.',
      'update.blocker.goInstall': 'A go install setup is updated with go.',
      'update.blocker.unknownMethod': 'Sillage cannot tell how this binary was installed.',
      'update.errorCheckFailed': 'Could not reach GitHub.',
      'update.errorApplyFailed': 'Update failed.',
      'update.serviceHeading': 'Startup',
      'update.serviceOn': 'Sillage starts when you log in.',
      'update.serviceOff': 'Sillage does not start when you log in.',
      'update.serviceFlagsNote': 'This instance runs with arguments the service will not carry over: it starts the binary with no options.'
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

  // Glyphes d'état, posés dans une pastille teintée (voir buildTaskGlyphHTML et
  // .task-glyph) : un disque plein pour « à relire », qui se repère d'un coup
  // d'œil sans se confondre avec le losange des compteurs de fichiers.
  var STATUS_GLYPH = {
    waiting: { icon: '◌', color: '#8b8982' },
    running: { icon: '◐', color: '#8b8982' },
    review: { icon: '●', color: '#9a6b0d' },
    accepted: { icon: '✓', color: '#2f7d54' },
    cancelled: { icon: '⊘', color: '#8b8982' }
  };

  // ---------------------------------------------------------------------
  // État en mémoire
  // ---------------------------------------------------------------------

  var state = {
    lang: detectInitialLang(),
    workspace: null,
    settings: { displayName: '', lang: '' },
    projects: [], cards: [], tasks: [], agents: [],
    signals: [], watch: {}, // Vigie : signaux actifs et état du veilleur
    autoconfigs: {}, // configurations automatiques en cours ou à relire, par projet
    claudeSessions: {}, // sessions Claude Code hors Sillage, par projet : {at, list}
    vigieError: '',
    vigieLogOpen: false, // journal de la Vigie déplié
    projectsById: {}, cardsById: {}, tasksById: {}, agentsById: {},
    tokens: { global: { input: 0, output: 0, costUsd: 0 } },
    messagesByTask: {}, diffByTask: {}, deliverablesByTask: {},
    diffCollapsed: {}, diffTreeOpen: false,
    detailErrorByTask: {},
    previews: [], // runs de recette en cours ou terminés (jamais persistés côté serveur)
    previewLogByRun: {}, // journal par run : tampon initial (GET) + lignes SSE
    previewScope: null, // { kind: 'card'|'task'|'all', id, runId } quand le panneau est ouvert
    deliveryByCard: {}, // aperçu de livraison (GET /api/cards/{id}/delivery), non persisté
    shipResultByCard: {}, // dernier résultat de livraison, affiché dans la modale
    catchUpErrorByCard: {}, // échec de rattrapage hors conflit, affiché dans la barre
    acceptingByTask: {}, // acceptation en vol : le bouton attend sa fusion, sans deuxième clic possible
    attachDrafts: {}, // images jointes en attente d'envoi, par tâche : [{key, name, mime, size, dataUrl}]
    loading: {},
    screen: 'inbox', // 'inbox' | 'vigie' | 'projects' | 'kanban' | 'work'
    projectId: null, cardId: null, taskId: null,
    panelTab: 'chat', panelExpanded: false, taskFilter: 'all',
    pendingConversationScroll: false, // défiler vers le bas au prochain rendu (ouverture d'une conversation)
    convUnseen: false, // un message est arrivé alors qu'on lisait plus haut dans le fil
    convJumpState: null, // dernier état peint de la pastille de retour en bas ('off' | 'jump' | 'new')
    searchOpen: false, modal: null,
    pendingConfirm: null, // { key, timer }
    sidebarOpen: false // tiroir mobile (< 860px)
  };

  var modalAgentId = null;
  var modalAttachments = []; // images jointes en attente, modale de création de tâche
  var modalRepos = []; // [{name, path}] pour les onglets de la modale de projet
  var modalLinks = []; // [{url, title}] liens épinglés, modale de projet
  // Modale de plan de tâches. planDraft porte la proposition en cours de
  // relecture : elle vit hors du DOM parce que retirer une étape réécrit la
  // liste, et que ce qui a été corrigé à la main ne doit pas repartir.
  var planCardId = null;
  var planDraft = null; // null (formulaire) | { steps: [{title, prompt}], repoName, chain }
  var planBusy = false; // un aller-retour est en cours : les boutons sont figés
  // Ce qui a été saisi dans le formulaire, gardé hors du DOM : la modale se
  // re-rend pendant l'attente puis en cas d'échec, et un plan qui n'aboutit pas
  // ne doit pas effacer les précisions qu'on venait d'écrire.
  var planExtra = '';
  var planRepoName = '';
  // Attente d'un plan : de quoi la rendre visiblement vivante (voir
  // buildPlanThinkingHTML). planRunToken périme la réponse d'un aller-retour
  // dont la modale a été fermée entre-temps : sans lui, elle se rouvrirait
  // toute seule quelques minutes plus tard.
  var planRunToken = 0;
  var planStartedAt = 0;
  var planTicker = null;
  var planWaitKey = 'plan.proposing'; // libellé de l'attente en cours
  // Création de tâche : ce qui a été saisi, gardé hors du DOM. La modale de
  // création et le choix d'une issue GitHub occupent le même emplacement
  // (#modal-root) : passer de l'une à l'autre effacerait la saisie en cours.
  var newTaskDraft = null; // {title, prompt, repoName, waitsForTaskId}
  // Chantier implicite : la modale de création ouverte depuis un projet (ou la
  // Vigie) n'a pas encore de chantier. Le serveur en crée un au nom de la tâche.
  var newTaskProjectId = null;
  var newTaskSignalId = null; // signal de la Vigie auquel la tâche répond
  // Choix d'une issue GitHub. La liste vient de la forge à chaque frappe (voir
  // scheduleIssueSearch) : issueRunToken périme la réponse d'une recherche
  // dépassée par la suivante, ou dont la modale est déjà fermée.
  var issueCardId = null;
  var issueRepoName = '';
  var issueQuery = '';
  var issueList = null; // null = première lecture en cours
  var issueBusy = false;
  var issueError = '';
  var issueRunToken = 0;
  var issueSearchTimer = null;
  // Onglet actif et brouillon des champs simples de la modale de projet. Les
  // panneaux se rendent un par un : sans ce brouillon, changer d'onglet
  // perdrait ce qui vient d'être saisi dans le panneau qu'on quitte.
  var projectModalTab = 'general';
  var projectDraft = null; // {name, description, target, contextPrompt, checkCmd, mode} ou null
  var projectDraftDirty = false;
  var searchIndex = 0; // résultat de recherche actif (navigation aux flèches)
  var onboardingExpanded = null;
  var settingsModalTab = 'general'; // 'general' | 'sync' | 'agents' | 'stats' | 'update'
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

  // Le CSS coupe déjà toutes les animations (voir « Micro-interactions ») ; le
  // défilement animé, lui, se pilote en JS et doit se couper de la même façon.
  function prefersReducedMotion() {
    return !!(window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches);
  }

  // Libellé de la touche de modification (badges et infobulles des raccourcis).
  function modKeyLabel() { return isMac() ? '⌘' : 'Ctrl'; }

  // playWake : l'onde verte qui traverse une fois une surface qui vient de
  // changer (voir « Le sillage » dans style.css). Toujours appelée après le
  // rendu qui a reconstruit la surface, sinon la classe part avec l'ancien
  // nœud. La classe est retirée en fin d'animation pour qu'un second geste la
  // rejoue : reposée sur un nœud qui la porte déjà, elle ne relancerait rien.
  function playWake(el) {
    if (!el || prefersReducedMotion()) return;
    el.classList.add('wake');
    // Le nom de l'animation est vérifié parce que l'événement remonte des
    // enfants : une pastille qui bat dans la surface la retirerait sinon.
    el.addEventListener('animationend', function done(e) {
      if (e.animationName !== 'om-wake') return;
      el.classList.remove('wake');
      el.removeEventListener('animationend', done);
    });
  }

  // Fallback favicon (liens épinglés) : référencé depuis un attribut onerror
  // généré côté chaîne HTML, doit donc être une fonction globale réelle.
  window.__pinnedLinkIconError = function (img) {
    var span = document.createElement('span');
    span.className = 'pinned-link-fallback';
    span.textContent = '🔗';
    if (img && img.parentNode) img.parentNode.replaceChild(span, img);
  };

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
  function tokenTotal(tokens) {
    tokens = tokens || {};
    return (tokens.input || 0) + (tokens.output || 0);
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

  // formatResetCountdown : durée avant iso (futur), pour l'heure de
  // réinitialisation d'une fenêtre de quota. Contrepartie de timeAgo (passé).
  function formatResetCountdown(iso) {
    if (!iso) return '';
    var then = new Date(iso).getTime();
    if (isNaN(then)) return '';
    var diff = Math.max(0, Math.round((then - Date.now()) / 1000));
    var min = Math.max(1, Math.round(diff / 60));
    if (min < 60) return t('time.inMin', { n: min });
    var hr = Math.round(min / 60);
    if (hr < 24) return t('time.inHour', { n: hr });
    var day = Math.round(hr / 24);
    return t('time.inDay', { n: day });
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
    html = html.replace(/(^|[^*\w])\*([^*\s][^*\n]*?)\*(?![*\w])/g, '$1<em>$2</em>');
    // Liens Markdown : http(s) seulement, le texte est déjà échappé (les
    // guillemets aussi, donc l'URL ne peut pas sortir de son attribut).
    html = html.replace(/\[([^\]]+)\]\((https?:\/\/[^\s)]+)\)/g, '<a href="$2" target="_blank" rel="noopener noreferrer" data-action="external">$1</a>');
    var lines = html.split('\n');
    var out = [];
    var inList = false;
    for (var i = 0; i < lines.length; i++) {
      var line = lines[i];
      var m = line.match(/^\s*[-*]\s+(.*)$/);
      var h = !m && line.match(/^\s*#{1,6}\s+(.*)$/);
      if (h) {
        if (inList) { out.push('</ul>'); inList = false; }
        out.push('<p class="md-h">' + h[1] + '</p>');
      } else if (m) {
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
      // Le bouton de copie vit dans le bloc plutôt qu'à côté : c'est du code, on
      // le prend en entier, et il ne se révèle qu'au survol du bloc.
      return '<div class="code-wrap"><pre class="code-block"><code>' + escapeHtml(blocks[Number(idx)]) + '</code></pre>' +
        '<button class="code-copy" data-action="copy-code" title="' + escapeHtml(t('chat.copyCode')) + '" aria-label="' + escapeHtml(t('chat.copyCode')) + '">' +
        copyIconHTML() + '</button></div>';
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
    state.previews = data.previews || [];
    state.signals = data.signals || [];
    state.watch = data.watch || {};
    state.lan = data.lan || {};
    state.autoconfigs = {};
    (data.autoconfigs || []).forEach(function (r) { state.autoconfigs[r.projectId] = r; });
    state.workspace = data.workspace || state.workspace || null;
    state.update = data.update || state.update || null;
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

  // Purge locale d'une tâche supprimée (état + tout le cache associé).
  function removeTaskLocally(taskId) {
    var idx = state.tasks.findIndex(function (t) { return t.id === taskId; });
    if (idx >= 0) state.tasks.splice(idx, 1);
    delete state.tasksById[taskId];
    delete state.messagesByTask[taskId];
    delete state.diffByTask[taskId];
    delete state.deliverablesByTask[taskId];
    delete state.detailErrorByTask[taskId];
  }

  // Purge locale d'un chantier supprimé, avec cascade sur ses tâches (le
  // backend supprime aussi les tâches, mais chaque taskDeleted peut arriver
  // avant ou après ; cette purge est idempotente et sert de filet).
  function removeCardLocally(cardId) {
    var idx = state.cards.findIndex(function (c) { return c.id === cardId; });
    if (idx >= 0) state.cards.splice(idx, 1);
    delete state.cardsById[cardId];
    delete state.deliveryByCard[cardId];
    delete state.shipResultByCard[cardId];
    delete state.catchUpErrorByCard[cardId];
    state.tasks.filter(function (t) { return t.cardId === cardId; }).forEach(function (t) {
      removeTaskLocally(t.id);
    });
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

  function projectHasRunningTask(pid) {
    return state.tasks.some(function (t) { return t.projectId === pid && t.status === 'running'; });
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
      if (kind === 'agent-delete') doDeleteAgent(id);
      else if (kind === 'agent-fix') doFixAgentWarning(id);
      else if (kind === 'workspace-sync') doWorkspaceSync();
      else if (kind === 'task-cancel') doCancelTask(id);
      else if (kind === 'task-delete') doDeleteTask(id);
      else if (kind === 'card-delete') doDeleteCard(id);
      else if (kind === 'project-delete') doDeleteProject(id);
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
    if (parts[0] === 'vigie') return { screen: 'vigie' };
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
      if (parts[2] === 's' && parts[3]) return { screen: 'signal', projectId: projectId, signalId: parts[3] };
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
    // L'accueil est ce qui vous attend, pas la liste des dossiers.
    if (!route) { location.hash = '#/inbox'; return; }
    if (route.projectId && !state.projectsById[route.projectId]) { location.hash = '#/projects'; return; }
    if (route.cardId) {
      var card = state.cardsById[route.cardId];
      if (!card || card.projectId !== route.projectId) { location.hash = '#/projects'; return; }
    }
    if (route.signalId) {
      var sig = findSignal(route.signalId);
      if (!sig || sig.projectId !== route.projectId) { location.hash = '#/p/' + encodeURIComponent(route.projectId); return; }
    }
    if (route.taskId) {
      var task = state.tasksById[route.taskId];
      if (!task || task.cardId !== route.cardId) { location.hash = '#/projects'; return; }
    }
    var prevCardId = state.cardId;
    var prevTaskId = state.taskId;
    var prevPanelTab = state.panelTab;
    state.screen = route.screen;
    state.projectId = route.projectId || null;
    state.cardId = route.cardId || null;
    state.taskId = route.taskId || null;
    state.signalId = route.signalId || null;
    if (state.cardId !== prevCardId) state.taskFilter = 'all';
    if (state.taskId) {
      state.panelTab = route.tab === 'diff' ? 'diff' : (route.tab === 'deliverables' ? 'files' : (route.tab === 'history' ? 'history' : 'chat'));
      if (state.taskId !== prevTaskId) {
        state.panelExpanded = false;
        var t = state.tasksById[state.taskId];
        if (t) t.unread = false;
        api('/api/tasks/' + state.taskId + '/read', { method: 'POST' }).catch(function () {});
      }
      // Ouvrir une conversation (nouvelle tâche, ou retour sur l'onglet chat) doit
      // toujours défiler vers le dernier message, même si celui-ci arrive après
      // un chargement asynchrone (voir loadMessages -> renderMain).
      if (state.panelTab === 'chat' && (state.taskId !== prevTaskId || prevPanelTab !== 'chat')) {
        state.pendingConversationScroll = true;
        // On repart en bas du fil : plus rien n'attend d'être vu plus bas.
        state.convUnseen = false;
      }
    }
    closeSidebarDrawer();
    syncDeliveryPolling();
    render();
  }

  // ---------------------------------------------------------------------
  // Navigation (construit le hash cible ; l'application de l'état se fait
  // dans applyRoute(), déclenché par le hashchange que provoque navigateTo)
  // ---------------------------------------------------------------------

  function goInbox() { closeSidebarDrawer(); navigateTo('#/inbox'); }
  function goVigie() { closeSidebarDrawer(); navigateTo('#/vigie'); }
  // Depuis la boîte de réception : la Vigie, le signal déplié et sous les yeux.
  function openSignal(id) {
    var s = findSignal(id);
    if (!s) return;
    closeSidebarDrawer();
    navigateTo('#/p/' + encodeURIComponent(s.projectId) + '/s/' + encodeURIComponent(s.id));
  }
  function goAllProjects() { closeSidebarDrawer(); navigateTo('#/projects'); }
  function goProject(id) { closeSidebarDrawer(); navigateTo('#/p/' + encodeURIComponent(id)); }
  function goCard(id) {
    var c = state.cardsById[id];
    if (!c) return;
    // Un chantier d'une seule tâche, c'est cette tâche : l'ouvrir directement
    // évite un écran de liste qui ne contient qu'une ligne.
    var tasks = state.tasks.filter(function (tk) { return tk.cardId === id; });
    if (tasks.length === 1) { openTask(tasks[0].id); return; }
    navigateTo('#/p/' + encodeURIComponent(c.projectId) + '/c/' + encodeURIComponent(id));
  }
  function goBack() { if (state.projectId) goProject(state.projectId); else goAllProjects(); }
  function closePanel() {
    if (state.cardId) navigateTo('#/p/' + encodeURIComponent(state.projectId) + '/c/' + encodeURIComponent(state.cardId));
    else goInbox();
  }
  // En place, sans rendu : c'est la classe posée sur le même élément qui laisse
  // la largeur du panneau s'animer. Le contenu s'adapte seul (requêtes de
  // conteneur), seule la poignée change de libellé.
  function togglePanelExpand() {
    state.panelExpanded = !state.panelExpanded;
    var body = document.querySelector('.work-body.has-panel');
    if (!body) { renderMain(); return; }
    body.classList.toggle('panel-expanded', state.panelExpanded);
    var h = body.querySelector('.panel-handle');
    if (h) { h.title = panelHandleLabel(); h.setAttribute('aria-label', panelHandleLabel()); }
  }
  function setFilter(f) { state.taskFilter = f; renderMain(); }
  function setTab(tabKey) {
    if (!state.taskId) return;
    var h = '#/p/' + encodeURIComponent(state.projectId) + '/c/' + encodeURIComponent(state.cardId) + '/t/' + encodeURIComponent(state.taskId);
    if (tabKey === 'diff') h += '?tab=diff';
    else if (tabKey === 'files') h += '?tab=deliverables';
    else if (tabKey === 'history') h += '?tab=history';
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

  // Compteur des recettes en cours : la seule protection contre les serveurs
  // oubliés. Pas de TTL qui coupe un serveur pendant qu'on s'en sert ; on rend
  // visible ce qui tourne, et on laisse l'humain arrêter.
  function buildPreviewCounterHTML() {
    var running = previewRunning();
    if (running.length === 0) return '';
    return '<button class="preview-counter" data-action="open-all-previews">' +
      '<span class="preview-dot preview-dot-on"></span>' +
      escapeHtml(tCount('preview.running', running.length)) + '</button>';
  }

  function buildSidebarHTML() {
    var navItems = [
      { key: 'inbox', icon: '⌂', label: t('nav.inbox'), action: 'nav-inbox' },
      { key: 'vigie', icon: '◎', label: t('nav.vigie'), action: 'nav-vigie', badge: vigieAttentionCount() },
      { key: 'projects', icon: '◫', label: t('nav.allProjects'), action: 'nav-projects' }
    ];
    var navHTML = navItems.map(function (n) {
      var active = state.screen === n.key;
      return '<button class="nav-item ' + (active ? 'active' : '') + '" data-action="' + n.action + '">' +
        '<span class="nav-icon">' + n.icon + '</span><span class="nav-label">' + escapeHtml(n.label) + '</span>' +
        (n.badge ? '<span class="badge-unread">' + n.badge + '</span>' : '') + '</button>';
    }).join('');

    var projectsHTML = state.projects.filter(function (p) { return !p.hiddenFromSidebar; }).map(function (p) {
      var active = (state.screen === 'kanban' || state.screen === 'work') && state.projectId === p.id;
      var unread = projectUnread(p.id);
      // Le fuseau remplace le dièse quand un agent travaille dans le projet :
      // même repère visuel qu'une tâche en cours, vu de la sidebar.
      var hashHTML = projectHasRunningTask(p.id)
        ? '<span class="task-spinner project-item-spinner" title="' + escapeHtml(t('status.running')) + '"></span>'
        : '<span class="hash">#</span>';
      return '<div class="project-item-wrap">' +
        '<button class="project-item ' + (active ? 'active' : '') + '" data-action="nav-project" data-project-id="' + p.id + '">' +
          hashHTML + '<span class="project-name">' + escapeHtml(p.name) + '</span>' +
          (unread ? '<span class="badge-unread">' + unread + '</span>' : '') +
        '</button>' +
        '<button class="project-menu-btn" data-action="toggle-project-menu" data-project-id="' + p.id + '" data-menu-scope="sidebar" title="' + escapeHtml(t('sidebar.projectMenuTooltip')) + '" aria-label="' + escapeHtml(t('sidebar.projectMenuTooltip')) + '">⋯</button>' +
      '</div>';
    }).join('');

    // La barre latérale sert à naviguer, pas à configurer. Elle ne montre donc
    // que les agents dont l'état change quelque chose maintenant : ceux qui
    // travaillent, et ceux dont le CLI est mal installé (un avertissement qu'on
    // ne peut pas se permettre de cacher). Le reste de la liste vit dans les
    // réglages, onglet Agents. Sept lignes d'agents au repos, c'était la moitié
    // de la barre pour une information qu'on ne relit jamais.
    // Un avertissement ne mérite sa ligne que pour un agent qu'on emploie : les
    // CLI jamais installés se regroupent en une seule ligne vers les réglages.
    var usedAgents = {};
    state.tasks.forEach(function (tk) { usedAgents[tk.agentId] = true; });
    var shownAgents = state.agents.filter(function (a) { return a.active || (a.warning && usedAgents[a.id]); });
    var idleBroken = state.agents.filter(function (a) { return !a.active && a.warning && !usedAgents[a.id]; }).length;
    var agentsHTML = shownAgents.map(function (a) {
      var warn = a.warning ? '<span class="agent-warning-sm" title="' + escapeHtml(agentWarningText(a.warning)) + '">⚠</span>' : '';
      return '<div class="agent-item" data-action="edit-agent" data-agent-id="' + a.id + '">' +
        '<span class="agent-avatar" style="background:' + softColor(a.color) + '">' + a.emoji + '</span>' +
        '<span class="agent-name">' + escapeHtml(a.name) + '</span>' +
        warn +
        (a.active ? '<span class="agent-dot"></span>' : '') +
        '</div>';
    }).join('');
    // Quand personne ne travaille, la section entière disparaît plutôt que
    // d'afficher un titre au-dessus du vide.
    var agentsSectionHTML = agentsHTML
      ? '<div class="sidebar-section-head"><span>' + escapeHtml(t('sidebar.agentsAtWork')) + '</span></div>' +
        '<div class="agent-list">' + agentsHTML + '</div>'
      : '';
    if (idleBroken) {
      agentsSectionHTML += '<button class="sidebar-agents-setup" data-action="open-settings-agents">' +
        escapeHtml(tCount('sidebar.agentsToSetup', idleBroken)) + '</button>';
    }

    return '' +
      '<div class="sidebar-brand">' +
        '<span class="brand-mark"></span><span class="brand-name">Sillage</span>' +
      '</div>' +
      '<div class="sidebar-search-wrap">' +
        '<button class="search-btn" data-action="open-search"><span>⌕</span><span class="search-btn-label">' + escapeHtml(t('search.buttonLabel')) + '</span>' +
        '<span class="kbd">' + (isMac() ? '⌘K' : 'Ctrl+K') + '</span></button>' +
      '</div>' +
      '<div class="nav-list">' + navHTML + '</div>' +
      '<div class="sidebar-section-head"><span>' + escapeHtml(t('sidebar.projectsHeading')) + '</span>' +
      '<button class="icon-btn-sm" data-action="open-new-project" title="' + escapeHtml(t('sidebar.newProjectTooltip')) + '" aria-label="' + escapeHtml(t('sidebar.newProjectTooltip')) + '">+</button></div>' +
      '<nav class="project-list">' + (projectsHTML || '<div class="empty-note-sm">' + escapeHtml(t('sidebar.noProjects')) + '</div>') + '</nav>' +
      agentsSectionHTML +
      // Le pied ne garde que ce qui appelle une action : une recette en cours, une
      // mise à jour disponible, et la porte des réglages. Les liens du dépôt et
      // du sponsoring ont rejoint l'onglet Général : on ne les ouvre pas en
      // travaillant, ils n'ont donc pas à être visibles en permanence.
      '<div class="sidebar-footer">' +
        buildPreviewCounterHTML() +
        buildUpdateLineHTML() +
        '<button class="settings-btn" data-action="open-workspace-modal">⚙ ' + escapeHtml(t('sidebar.settingsButton')) + '</button>' +
        '<div class="sidebar-footer-actions">' +
          '<button class="icon-btn-sm" data-action="open-shortcuts" title="' + escapeHtml(t('shortcuts.title')) + '" aria-label="' + escapeHtml(t('shortcuts.title')) + '">?</button>' +
        '</div>' +
      '</div>';
  }

  function renderSidebar() {
    var el = document.getElementById('sidebar');
    if (el) el.innerHTML = buildSidebarHTML();
    updateFaviconBadge();
  }

  // Pastille rouge sur la favicon quand au moins une tâche est non lue :
  // seul signal visible quand l'onglet n'est pas au premier plan. Positionnée
  // dans le coin haut-gauche du logo (20x20), le seul entièrement libre de
  // pixels (le tracé du logo n'occupe que x >= 9 sur les lignes y <= 7).
  var faviconUnread = null;
  var faviconSvgTextPromise = null;

  function faviconSvgText() {
    if (!faviconSvgTextPromise) {
      faviconSvgTextPromise = fetch('favicon.svg').then(function (r) { return r.text(); });
    }
    return faviconSvgTextPromise;
  }

  // Remplacer le <link> (plutôt que muter son href) car certains navigateurs
  // ignorent un changement de href sur un lien déjà chargé.
  function setFaviconHref(href) {
    var link = document.createElement('link');
    link.rel = 'icon';
    link.type = 'image/svg+xml';
    link.href = href;
    document.head.appendChild(link);
    document.querySelectorAll('link[rel="icon"][type="image/svg+xml"]').forEach(function (el) {
      if (el !== link) el.remove();
    });
  }

  function updateFaviconBadge() {
    var hasUnread = state.tasks.some(function (t) { return t.unread; });
    if (hasUnread === faviconUnread) return;
    faviconUnread = hasUnread;
    if (!hasUnread) { setFaviconHref('favicon.svg'); return; }
    faviconSvgText().then(function (svg) {
      if (!faviconUnread) return;
      var badged = svg.replace('</svg>', '<circle cx="4" cy="4" r="3.5" fill="#b0392f"/></svg>');
      setFaviconHref('data:image/svg+xml,' + encodeURIComponent(badged));
    }).catch(function () {});
  }

  // Le contenu du menu « ⋯ » d'un projet : « Tout marquer comme lu » n'a de
  // sens que depuis la barre latérale (scope "sidebar"), la bascule de
  // visibilité est proposée partout, avec un libellé qui dépend de l'état
  // courant du projet.
  function projectMenuItemsHTML(projectId, scope) {
    var p = state.projectsById[projectId];
    var markReadHTML = scope === 'sidebar'
      ? '<button class="project-menu-item" data-action="mark-project-read" data-project-id="' + projectId + '">' + escapeHtml(t('sidebar.markAllRead')) + '</button>'
      : '';
    var visibilityLabel = (p && p.hiddenFromSidebar) ? t('sidebar.showProject') : t('sidebar.hideProject');
    return markReadHTML +
      '<button class="project-menu-item" data-action="toggle-project-sidebar-visibility" data-project-id="' + projectId + '">' + escapeHtml(visibilityLabel) + '</button>';
  }

  // Le menu « ⋯ » d'un projet est construit à la volée dans #menu-root, un
  // conteneur au niveau de la page plutôt qu'un enfant caché du bouton :
  // ouvert depuis .project-list (overflow-y: auto dans la barre latérale) ou
  // depuis la grille de la page "Tous les projets", un menu absolu resterait
  // pris dans le scroll de son conteneur et déborderait, forçant à scroller
  // pour en voir la fin. En position fixe, ancré sur la position réelle du
  // bouton au moment du clic, il flotte par-dessus tout, sans dépendre de la
  // hauteur restante du conteneur qui l'a ouvert.
  function closeAllProjectMenus() {
    var root = document.getElementById('menu-root');
    if (root) { root.innerHTML = ''; delete root.dataset.openProjectMenu; }
    // Le bouton « ⋯ » n'est révélé qu'au survol de son wrap (voir style.css) ;
    // le menu vivant ailleurs dans le DOM, il resterait affiché après que la
    // souris l'ait quitté pour aller cliquer dans le menu sans ce rappel.
    document.querySelectorAll('.project-menu-btn-open').forEach(function (b) { b.classList.remove('project-menu-btn-open'); });
  }
  function toggleProjectMenu(projectId, btnEl, scope) {
    var root = document.getElementById('menu-root');
    if (!root) return;
    var wasOpenForThis = root.dataset.openProjectMenu === projectId;
    closeAllProjectMenus();
    if (wasOpenForThis) return;
    if (!btnEl) btnEl = document.querySelector('.project-menu-btn[data-project-id="' + projectId + '"]');
    if (!btnEl) return;
    var rect = btnEl.getBoundingClientRect();
    root.dataset.openProjectMenu = projectId;
    btnEl.classList.add('project-menu-btn-open');
    root.innerHTML = '<div class="project-menu" style="position:fixed; top:' + (rect.bottom + 4) +
      'px; right:' + (document.documentElement.clientWidth - rect.right) + 'px;">' +
      projectMenuItemsHTML(projectId, scope) + '</div>';
  }
  // Optimiste : les tâches locales passent lues tout de suite, l'appel serveur
  // suit sans bloquer (comme markTaskRead sur l'ouverture d'une tâche).
  function markProjectAllRead(projectId) {
    closeAllProjectMenus();
    state.tasks.forEach(function (t) { if (t.projectId === projectId) t.unread = false; });
    render();
    api('/api/projects/' + projectId + '/mark-all-read', { method: 'POST' }).catch(function () {});
  }

  // Bascule le projet hors ou dans la barre latérale : optimiste, avec retour
  // en arrière silencieux si l'appel échoue (le menu qui l'a déclenché n'existe
  // parfois plus après le re-rendu, ex. depuis la page "Tous les projets").
  function toggleProjectSidebarVisibility(projectId) {
    closeAllProjectMenus();
    var p = state.projectsById[projectId];
    if (!p) return;
    var hidden = !p.hiddenFromSidebar;
    p.hiddenFromSidebar = hidden;
    render();
    api('/api/projects/' + projectId, { method: 'PATCH', body: { hiddenFromSidebar: hidden } }).then(function (project) {
      upsertProject(project);
      render();
    }).catch(function () {
      p.hiddenFromSidebar = !hidden;
      render();
    });
  }

  // ---------------------------------------------------------------------
  // Rendu : en-tête commun
  // ---------------------------------------------------------------------

  // Bouton de création principal d'un écran : pictogramme + libellé + badge du
  // raccourci clavier (« N »), pour que le raccourci s'apprenne sans documentation.
  function buildCreateButtonHTML(action, label, tooltip, cls) {
    return '<button class="' + cls + ' btn-create" data-action="' + action + '" title="' + escapeHtml(tooltip) + '">' +
      '<span class="btn-ico" aria-hidden="true">+</span>' +
      '<span class="btn-create-label">' + escapeHtml(label) + '</span>' +
      '<span class="kbd kbd-in-btn" aria-hidden="true">N</span></button>';
  }

  function buildHeaderHTML() {
    var back = '', title = '', branchHTML = '', actions = '', editBtn = '';
    if (state.screen === 'inbox') {
      title = t('nav.inbox');
    } else if (state.screen === 'vigie') {
      title = t('nav.vigie');
      actions = buildVigieHeaderActionsHTML();
    } else if (state.screen === 'projects') {
      title = t('nav.allProjects');
      actions = buildCreateButtonHTML('open-new-project', t('header.newProject'), t('header.newProjectTooltip'), 'btn-outline');
    } else if (state.screen === 'kanban') {
      // Le nom du projet est déjà affiché en grand titre (avec bouton edit)
      // dans buildKanbanHTML : pas besoin de le répéter dans la barre du haut.
    } else if (state.screen === 'signal') {
      var sp = state.projectsById[state.projectId];
      var sg = findSignal(state.signalId);
      back = '<button class="back-crumb" data-action="go-back" aria-label="' + escapeHtml(t('nav.back')) + (sp ? ' : ' + escapeHtml(sp.name) : '') + '">' +
        '<svg class="back-chevron" viewBox="0 0 24 24" width="18" height="18" aria-hidden="true"><path d="M15 5l-7 7 7 7" fill="none" stroke="currentColor" stroke-width="2.3" stroke-linecap="round" stroke-linejoin="round"/></svg>' +
        (sp ? '<span class="back-crumb-project">' + escapeHtml(sp.name) + '</span>' : '') +
        '</button>';
      title = sg ? signalRefText(sg) : '';
    } else if (state.screen === 'work') {
      var c = state.cardsById[state.cardId];
      var pr = state.projectsById[state.projectId];
      // Le nom du projet vient se poser contre la flèche : cliquer n'importe où
      // dans ce groupe ramène au même endroit (goBack), donc les deux ne font
      // qu'un plutôt que d'écarteler le projet à droite du titre du chantier.
      back = '<button class="back-crumb" data-action="go-back" aria-label="' + escapeHtml(t('nav.back')) + (pr ? ' : ' + escapeHtml(pr.name) : '') + '">' +
        '<svg class="back-chevron" viewBox="0 0 24 24" width="18" height="18" aria-hidden="true"><path d="M15 5l-7 7 7 7" fill="none" stroke="currentColor" stroke-width="2.3" stroke-linecap="round" stroke-linejoin="round"/></svg>' +
        (pr ? '<span class="back-crumb-project">' + escapeHtml(pr.name) + '</span>' : '') +
        '</button>';
      title = c ? c.title : '';
      if (c) {
        // Toutes les CardBranch d'un chantier partagent le même nom de branche
        // (dérivé de card.Ref + card.Title, indépendant du dépôt) : une seule
        // suffit à l'afficher, absente tant qu'aucune tâche n'a créé la branche.
        var branchName = c.branches && c.branches.length ? c.branches[0].branch : '';
        if (branchName) branchHTML = '<span class="crumb-branch mono">' + escapeHtml(branchName) + '</span>';
        editBtn = '<button class="icon-btn" data-action="open-edit-card" title="' + escapeHtml(t('workstream.editTooltip')) + '" aria-label="' + escapeHtml(t('workstream.editTooltip')) + '">✎</button>';
        // Recette avant Livrer : on éprouve, puis on livre. Le bouton est
        // toujours là, même sans commande configurée (le panneau propose alors
        // le chemin du worktree).
        actions = buildPreviewButtonHTML(c) + buildShipButtonHTML(c);
      }
    }
    var hamburger = '<button class="icon-btn hamburger-btn" data-action="toggle-sidebar" aria-label="' + escapeHtml(t('aria.menu')) + '">☰</button>';
    // Sur le kanban, la barre du haut n'a plus rien à montrer (le titre du
    // projet est déjà le gros titre de buildKanbanHTML) : elle ne reste
    // visible que sur mobile, pour le hamburger d'ouverture de la sidebar.
    var topbarClass = state.screen === 'kanban' ? 'topbar topbar-empty-desktop' : 'topbar';
    return '<header class="' + topbarClass + '">' + hamburger + back +
      '<div class="crumb-title-wrap"><span class="crumb-title">' + escapeHtml(title) + '</span>' + branchHTML + '</div>' + editBtn +
      '<div class="topbar-actions">' + actions + '</div></header>';
  }

  // ---------------------------------------------------------------------
  // Rendu : Tous les projets
  // ---------------------------------------------------------------------

  // Une tuile de projet dit où il en est, dans le vocabulaire du flux : ce qui
  // vous attend d'abord, ce que font les agents ensuite.
  function projectPulseText(p, cardCount) {
    var you = 0, agents = 0;
    state.cards.forEach(function (c) {
      if (c.projectId !== p.id) return;
      var lane = cardLane(c);
      if (lane === 'you') you++; else if (lane === 'agents') agents++;
    });
    if (projectWatched(p)) {
      vigieSignals().forEach(function (sg) {
        if (sg.projectId !== p.id) return;
        if (signalLane(sg) === 'you') you++; else agents++;
      });
    }
    var parts = [];
    if (you) parts.push(tCount('pulse.you', you));
    if (agents) parts.push(tCount('pulse.agents', agents));
    if (!parts.length) parts.push(cardCount ? t('pulse.calm') : t('pulse.empty'));
    return { text: parts.join(' · '), you: you > 0 };
  }

  function buildAllProjectsHTML() {
    if (state.projects.length === 0) {
      return buildHeaderHTML() + '<div class="view-body"><div class="empty-state big">' +
        '<div class="empty-title">' + escapeHtml(t('allProjects.emptyTitle')) + '</div>' +
        '<div class="empty-sub">' + escapeHtml(t('allProjects.emptySub')) + '</div>' +
        buildCreateButtonHTML('open-new-project', t('header.newProject'), t('header.newProjectTooltip'), 'btn-primary') +
        '</div></div>';
    }
    var tiles = state.projects.map(function (p) {
      var cardCount = state.cards.filter(function (c) { return c.projectId === p.id; }).length;
      var unread = projectUnread(p.id);
      var pulse = projectPulseText(p, cardCount);
      return '<div class="project-tile-wrap">' +
        '<article class="project-tile" data-action="nav-project" data-project-id="' + p.id + '">' +
          '<div class="project-tile-top"><span class="project-hash">#</span><h3>' + escapeHtml(p.name) + '</h3>' +
          (unread ? '<span class="badge-unread">' + unread + '</span>' : '') + '</div>' +
          '<div class="project-tile-meta"><span class="' + (pulse.you ? 'project-pulse-you' : '') + '">' + escapeHtml(pulse.text) + '</span>' +
          (p.hiddenFromSidebar ? '<span class="project-tile-hidden-note">' + escapeHtml(t('allProjects.hiddenBadge')) + '</span>' : '') +
          '</div>' +
        '</article>' +
        '<button class="project-menu-btn project-tile-menu-btn" data-action="toggle-project-menu" data-project-id="' + p.id + '" data-menu-scope="tile" title="' + escapeHtml(t('sidebar.projectMenuTooltip')) + '" aria-label="' + escapeHtml(t('sidebar.projectMenuTooltip')) + '">⋯</button>' +
      '</div>';
    }).join('');
    return buildHeaderHTML() + '<div class="view-body all-projects-body"><div class="page-col"><div class="project-grid">' + tiles +
      '<button class="project-tile project-tile-add" data-action="open-new-project">' + escapeHtml(t('header.newProject')) + '</button>' +
      '</div></div></div>';
  }

  // ---------------------------------------------------------------------
  // Rendu : Vigie
  // ---------------------------------------------------------------------

  // La Vigie montre ce qui attend une décision, déjà relu par un agent. Ordre :
  // le prioritaire, puis l'ordinaire, puis ce qui a échoué ou attend son
  // analyse, et le bruit (priorité basse) en dernier. Rien n'y est publié sur
  // GitHub : la revue se copie, elle ne s'envoie pas.
  function signalRank(s) {
    if (s.status === 'ready') return s.priority === 'high' ? 0 : (s.priority === 'low' ? 4 : 1);
    if (s.status === 'error') return 2;
    return 3;
  }

  function vigieSignals() {
    return state.signals.filter(function (s) { return s.status !== 'dismissed' && s.status !== 'handled'; })
      .sort(function (a, b) {
        var d = signalRank(a) - signalRank(b);
        if (d) return d;
        return idNumJS(b.id) - idNumJS(a.id);
      });
  }

  function idNumJS(id) { var m = /\d+/.exec(String(id || '')); return m ? parseInt(m[0], 10) : 0; }

  // Le badge de la barre latérale ne compte que ce qui mérite d'être lu : un
  // signal analysé et qui n'est pas du bruit.
  function vigieAttentionCount() {
    return state.signals.filter(function (s) { return s.status === 'ready' && s.priority !== 'low'; }).length;
  }

  function vigieProjects() {
    return state.projects.filter(function (p) { return p.githubRepos && p.githubRepos.length; });
  }
  function projectWatched(p) { return !!(p && p.watch && p.watch.enabled); }

  function canNotify() {
    return typeof window.Notification !== 'undefined' && window.isSecureContext;
  }

  function buildVigieHeaderActionsHTML() {
    var html = '';
    if (canNotify() && Notification.permission === 'default' && state.projects.some(projectWatched)) {
      html += '<button class="btn-outline" data-action="vigie-notify">' + escapeHtml(t('vigie.notify')) + '</button>';
    }
    if (state.projects.some(projectWatched)) {
      var checking = state.watch && state.watch.checking;
      var when = state.watch && state.watch.checkedAt;
      html += '<span class="vigie-checked" title="' + escapeHtml(when ? new Date(when).toLocaleString() : '') + '">' +
        escapeHtml(checking ? '' : (when ? t('vigie.checkedAgo', { when: timeAgo(when) }) : t('vigie.neverChecked'))) + '</span>';
      html += '<button class="btn-outline" data-action="vigie-check"' + (checking ? ' disabled' : '') + '>' +
        escapeHtml(t(checking ? 'vigie.checking' : 'vigie.check')) + '</button>';
    }
    return html;
  }

  function buildVigieHTML() {
    var projects = vigieProjects();
    var body;
    if (!projects.length) {
      body = '<div class="empty-state big"><div class="empty-title">' + escapeHtml(t('vigie.introTitle')) + '</div>' +
        '<div class="empty-sub">' + escapeHtml(t('vigie.noGithub')) + '</div></div>';
    } else if (!projects.some(projectWatched)) {
      body = '<div class="vigie-intro"><div class="empty-title">' + escapeHtml(t('vigie.introTitle')) + '</div>' +
        '<p class="empty-sub">' + escapeHtml(t('vigie.introSub')) + '</p>' +
        buildVigieProjectsHTML(projects) + '</div>';
    } else {
      var signals = vigieSignals();
      var listHTML = signals.length
        ? '<div class="signal-grid">' + signals.map(function (sg) { return buildSignalCardHTML(sg, { showProject: true, actions: true }); }).join('') + '</div>'
        : '<div class="empty-state big"><div class="empty-title">' + escapeHtml(t('vigie.emptyTitle')) + '</div>' +
          '<div class="empty-sub">' + escapeHtml(t('vigie.emptySub')) + '</div></div>';
      body = (state.vigieError ? '<div class="modal-error">' + escapeHtml(state.vigieError) + '</div>' : '') +
        listHTML +
        '<div class="vigie-foot">' +
          '<div class="vigie-foot-head"><span>' + escapeHtml(t('vigie.watchedHeading')) + '</span></div>' +
          buildVigieProjectsHTML(projects) +
          '<div class="modal-note">' + escapeHtml(t('vigie.localOnly')) + '</div>' +
          buildVigieLogHTML(null) +
        '</div>';
    }
    return buildHeaderHTML() + '<div class="view-body"><div class="page-col vigie-body">' + body + '</div></div>';
  }

  // Les analystes possibles : ceux dont Sillage garantit la lecture seule
  // (voir canPlan côté serveur).
  function analystAgents() {
    return state.agents.filter(function (a) { return a.cli === 'claude' || a.cli === 'codex' || a.cli === 'fake'; });
  }

  // La veille d'entretien d'un projet : sa dernière date et de quoi la
  // relancer, ou, pendant qu'elle tourne, ce que fait l'agent.
  function upkeepLiveInnerHTML(p) {
    var c = state.watch && state.watch.current;
    if (c && c.kind === 'upkeep' && c.projectId === p.id) {
      var agent = p.watch && state.agentsById[p.watch.agentId];
      var steps = c.steps || [];
      return '<span class="task-spinner"></span><span class="signal-line-text"><span>' +
        escapeHtml(t('vigie.upkeep.running', { agent: agent ? agent.name : '', repo: c.repoName })) + ' · ' + elapsedHTML(c.startedAt) + '</span>' +
        '<span class="signal-line-now mono">' + escapeHtml(steps.length ? steps[steps.length - 1] : t('vigie.live.start')) + '</span></span>' +
        '<button class="btn-quiet" data-action="vigie-upkeep-stop" data-project-id="' + p.id + '">' + escapeHtml(t('vigie.live.stop')) + '</button>';
    }
    var last = p.watch && p.watch.upkeepAt;
    return (last ? '<span class="vigie-upkeep-last">' + escapeHtml(t('vigie.upkeep.last', { when: timeAgo(last) })) + '</span>' : '') +
      '<button class="btn-quiet" data-action="vigie-upkeep" data-project-id="' + p.id + '">' + escapeHtml(t('vigie.upkeep.check')) + '</button>';
  }

  function buildVigieProjectsHTML(projects) {
    var errors = (state.watch && state.watch.errors) || {};
    return '<div class="vigie-projects">' + projects.map(function (p) {
      var on = projectWatched(p);
      var agentHTML = '';
      if (on) {
        var opts = analystAgents().map(function (a) {
          return '<option value="' + escapeHtml(a.id) + '"' + (a.id === p.watch.agentId ? ' selected' : '') + '>' +
            escapeHtml(a.emoji + ' ' + a.name) + '</option>';
        }).join('');
        agentHTML = '<label class="vigie-agent"><span>' + escapeHtml(t('vigie.analyst')) + '</span>' +
          '<select class="modal-input vigie-agent-select" data-project-id="' + p.id + '">' + opts + '</select></label>';
      }
      var err = errors[p.id] ? '<div class="vigie-project-error">' + escapeHtml(errors[p.id]) + '</div>' : '';
      var upkeep = on ? '<div class="vigie-upkeep" data-upkeep-live="' + p.id + '">' + upkeepLiveInnerHTML(p) + '</div>' : '';
      return '<div class="vigie-project">' +
        '<div class="vigie-project-line"><span class="hash">#</span><span class="vigie-project-name">' + escapeHtml(p.name) + '</span>' +
        agentHTML +
        '<button class="' + (on ? 'btn-quiet' : 'btn-primary') + ' vigie-toggle" data-action="vigie-watch" data-project-id="' + p.id + '" data-enabled="' + (on ? '0' : '1') + '">' +
        escapeHtml(t(on ? 'vigie.unwatch' : 'vigie.watch')) + '</button></div>' + upkeep + err +
      '</div>';
    }).join('') + '</div>';
  }


  // Pictogrammes des deux sortes de signal, dessinés en SVG au trait (même
  // règle que shipIconHTML : jamais d'emoji, jamais de police d'icônes).
  function signalKindIconHTML(kind) {
    var body = kind === 'upkeep'
      ? '<path d="M13.2 6.2A5.4 5.4 0 1 0 13.4 9"/><path d="M13.6 2.8v3.6H10"/><path d="M8 10.6V5.6M5.9 7.6L8 5.5l2.1 2.1"/>'
      : kind === 'pr'
      ? '<circle cx="4.5" cy="3.5" r="1.7"/><circle cx="4.5" cy="12.5" r="1.7"/><circle cx="11.5" cy="12.5" r="1.7"/>' +
        '<path d="M4.5 5.2v5.6M11.5 10.8V7.2a2.2 2.2 0 0 0-2.2-2.2H7"/><path d="M8.6 3.3L6.9 5l1.7 1.7"/>'
      : '<circle cx="8" cy="8" r="6"/><circle cx="8" cy="8" r="1.3" fill="currentColor" stroke="none"/>';
    return '<svg class="signal-kind-icon" viewBox="0 0 16 16" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true" focusable="false">' + body + '</svg>';
  }

  // La référence GitHub est toujours un lien : on doit pouvoir sauter à la PR
  // depuis n'importe où. data-action="external" l'empêche d'ouvrir aussi la
  // carte qui la contient.
  // La référence courte d'un signal : dépôt#numéro, ou le dépôt seul pour un
  // constat d'entretien, qui n'a pas de numéro sur la forge.
  function signalRefText(s) {
    return s.kind === 'upkeep' ? s.repoName : s.repoName + '#' + s.number;
  }

  function signalRefLinkHTML(s) {
    var label = escapeHtml(signalRefText(s));
    if (!s.url) return '<span class="signal-ref mono">' + label + '</span>';
    return '<a class="signal-ref-link mono" href="' + escapeHtml(s.url) + '" target="_blank" rel="noopener noreferrer" data-action="external" title="' + escapeHtml(t('vigie.openGithub')) + '">' +
      label + externalIconHTML() + '</a>';
  }

  function signalDecisionHTML(s) {
    return s.decision ? '<span class="signal-decision signal-decision-' + escapeHtml(s.decision) + '">' + escapeHtml(t('vigie.decision.' + s.decision)) + '</span>' : '';
  }

  // L'état d'un signal en une ligne : le verdict une fois analysé, sinon où en
  // est l'agent.
  function signalStateHTML(s, cls) {
    if (s.status === 'ready') {
      return '<div class="' + cls + '">' + signalDecisionHTML(s) + '<span>' + escapeHtml(s.verdict) + '</span></div>';
    }
    if (s.status === 'error') {
      return '<div class="' + cls + ' signal-verdict-error">' + escapeHtml(signalErrorText(s)) + '</div>';
    }
    return '<div class="' + cls + ' signal-verdict-pending" data-signal-line="' + s.id + '">' + signalLineInnerHTML(s) + '</div>';
  }

  function signalErrorText(s) {
    return s.error === 'stopped' ? t('vigie.stopped') : t('vigie.error', { error: s.error });
  }

  // La relecture en cours (une seule à la fois, voir analyzeQueued) si c'est
  // celle de ce signal.
  function signalProgress(s) {
    var c = state.watch && state.watch.current;
    return c && c.signalId === s.id ? c : null;
  }

  function signalAnalyst(s) {
    var project = state.projectsById[s.projectId];
    return (project && project.watch && state.agentsById[project.watch.agentId]) || null;
  }

  // Où en est un signal en file : combien passent avant lui. Sans ça, « en
  // attente » ne dit pas si quelque chose avance.
  function signalQueueText(s) {
    var project = state.projectsById[s.projectId];
    if (project && project.watch && !project.watch.enabled) return t('vigie.queuedPaused');
    var queued = state.signals.filter(function (x) { return x.status === 'queued'; })
      .sort(function (a, b) { return idNumber(a.id) - idNumber(b.id); });
    var ahead = queued.indexOf(s);
    if (ahead < 0) ahead = 0;
    if (state.watch && state.watch.current) ahead++;
    return ahead === 0 ? t('vigie.queuedNext') : tCount('vigie.queuedAhead', ahead);
  }

  function idNumber(id) { return parseInt(String(id).replace(/\D/g, ''), 10) || 0; }

  // La dernière chose que fait l'agent, ou sa phase quand il n'a encore rien fait.
  function progressNowText(c) {
    if (!c) return '';
    if (c.phase === 'diff') return t('vigie.live.diff');
    var steps = c.steps || [];
    return steps.length ? steps[steps.length - 1] : t('vigie.live.start');
  }

  // Une ligne sur une carte : qui relit, depuis quand, et ce qu'il fait là.
  function signalLineInnerHTML(s) {
    if (s.status !== 'analyzing') return '<span>' + escapeHtml(signalQueueText(s)) + '</span>';
    var c = signalProgress(s);
    var analyst = signalAnalyst(s);
    return '<span class="task-spinner"></span><span class="signal-line-text">' +
      '<span>' + escapeHtml(analyst ? t('vigie.live.title', { agent: analyst.name }) : t('vigie.analyzing')) +
      (c ? ' · ' + elapsedHTML(c.startedAt) : '') + '</span>' +
      (c ? '<span class="signal-line-now mono">' + escapeHtml(progressNowText(c)) + '</span>' : '') +
    '</span>';
  }

  // Le suivi en direct de la fiche : chaque outil employé par l'agent, la durée
  // et le délai d'abandon, et de quoi l'arrêter.
  function signalLiveInnerHTML(s) {
    if (s.status === 'queued') {
      return '<div class="sd-live-head"><span class="sd-live-title">' + escapeHtml(signalQueueText(s)) + '</span></div>' +
        '<div class="sd-note">' + escapeHtml(t('vigie.queuedHint')) + '</div>';
    }
    var c = signalProgress(s);
    var analyst = signalAnalyst(s);
    var steps = (c && c.steps) || [];
    var html = '<div class="sd-live-head"><span class="task-spinner"></span>' +
      '<span class="sd-live-title">' + escapeHtml(t('vigie.live.title', { agent: analyst ? analyst.name : '' })) + '</span>' +
      (c ? '<span class="sd-live-time">' + elapsedHTML(c.startedAt) +
        (c.timeoutSec ? ' · ' + escapeHtml(t('vigie.live.limit', { n: Math.round(c.timeoutSec / 60) })) : '') + '</span>' : '') +
      (c ? '<button class="btn-quiet sd-live-stop" data-action="signal-stop" data-signal-id="' + s.id + '">' + escapeHtml(t('vigie.live.stop')) + '</button>' : '') +
    '</div>';
    if (c && (c.phase === 'diff' || !steps.length)) {
      html += '<div class="sd-live-now">' + escapeHtml(progressNowText(c)) + '</div>';
    }
    if (steps.length) {
      html += '<ol class="live-steps">' + steps.map(function (line, i) {
        return '<li class="mono' + (i === steps.length - 1 ? ' live-step-now' : '') + '">' + escapeHtml(line) + '</li>';
      }).join('') + '</ol>';
    }
    return html;
  }

  // Les événements de progression ne reconstruisent pas l'écran (le défilement
  // de la fiche sauterait à chaque outil) : seules ces zones sont réécrites.
  function patchSignalLive() {
    Array.prototype.forEach.call(document.querySelectorAll('[data-signal-line]'), function (el) {
      var sg = findSignal(el.getAttribute('data-signal-line'));
      if (sg && (sg.status === 'queued' || sg.status === 'analyzing')) el.innerHTML = signalLineInnerHTML(sg);
    });
    Array.prototype.forEach.call(document.querySelectorAll('[data-upkeep-live]'), function (el) {
      var p = state.projectsById[el.getAttribute('data-upkeep-live')];
      if (p) el.innerHTML = upkeepLiveInnerHTML(p);
    });
    Array.prototype.forEach.call(document.querySelectorAll('[data-signal-live]'), function (el) {
      var sg = findSignal(el.getAttribute('data-signal-live'));
      if (sg && (sg.status === 'queued' || sg.status === 'analyzing')) el.innerHTML = signalLiveInnerHTML(sg);
    });
  }

  function elapsedText(iso) {
    var then = new Date(iso).getTime();
    if (isNaN(then)) return '';
    var sec = Math.max(0, Math.round((Date.now() - then) / 1000));
    var m = Math.floor(sec / 60);
    var rest = sec % 60;
    return m + ':' + (rest < 10 ? '0' : '') + rest;
  }

  // Un compteur qui avance, mis à jour chaque seconde par tickElapsed : c'est
  // ce qui distingue un agent qui travaille d'une page figée.
  function elapsedHTML(iso) {
    return '<span class="elapsed mono" data-elapsed-since="' + escapeHtml(iso || '') + '">' + escapeHtml(elapsedText(iso)) + '</span>';
  }

  function tickElapsed() {
    Array.prototype.forEach.call(document.querySelectorAll('[data-elapsed-since]'), function (el) {
      el.textContent = elapsedText(el.getAttribute('data-elapsed-since'));
    });
  }
  setInterval(tickElapsed, 1000);

  // 'github' : la décision se prend là-bas (merger, fermer) ; 'copy' : il y a
  // quelque chose à dire à l'auteur ; 'agent' : il y a du code à écrire.
  function signalPrimaryAction(s) {
    if (s.kind === 'upkeep') return 'agent';
    if (s.decision === 'merge') return 'github';
    if (s.decision === 'close') return s.review ? 'copy' : 'github';
    if (s.kind === 'issue' && (s.decision === 'fix' || s.decision === 'feature' || !s.decision)) return 'agent';
    return s.review ? 'copy' : 'github';
  }

  // Les actions d'un signal, la principale en tête et seule en plein. compact :
  // sur une carte, seulement l'action principale (le reste est dans la fiche).
  function buildSignalActionsHTML(s, compact, opts) {
    var actions = [];
    if (s.status === 'ready') {
      var primary = signalPrimaryAction(s);
      var copy = s.review && s.kind !== 'upkeep' ? '<button class="' + (primary === 'copy' ? 'btn-primary' : 'btn-outline') + '" data-action="signal-copy" data-signal-id="' + s.id + '">' +
        escapeHtml(t(s.kind === 'pr' ? 'vigie.copyReview' : 'vigie.copyReply')) + '</button>' : '';
      var agent = s.kind === 'issue' || s.kind === 'upkeep' ? '<button class="' + (primary === 'agent' ? 'btn-primary' : 'btn-outline') + '" data-action="signal-workstream" data-signal-id="' + s.id + '">' + escapeHtml(t('vigie.workstream')) + '</button>' : '';
      // Un travail plus large qu'une tâche : un chantier, que la planification
      // découpe. Seulement sur la fiche, la carte garde un seul bouton.
      var toCard = (s.kind === 'issue' || s.kind === 'upkeep') && !compact ? '<button class="btn-outline" data-action="signal-card" data-signal-id="' + s.id + '" title="' + escapeHtml(t('vigie.toCardHint')) + '">' + escapeHtml(t('vigie.toCard')) + '</button>' : '';
      var gh = s.url ? '<a class="' + (primary === 'github' ? 'btn-primary' : 'btn-quiet') + '" href="' + escapeHtml(s.url) + '" target="_blank" rel="noopener noreferrer" data-action="external">' +
        escapeHtml(t(primary === 'github' ? 'vigie.githubAction.' + (s.decision || '') : 'vigie.openGithub')) + '</a>' : '';
      var first = primary === 'copy' ? copy : (primary === 'agent' ? agent : gh);
      actions.push(first);
      if (!compact) {
        // La fiche a sa propre carte GitHub : le lien n'y est repris que s'il
        // est l'action principale.
        var rest = compact === false && opts && opts.noGithub ? [agent, toCard, copy] : [agent, toCard, copy, gh];
        rest.forEach(function (h) { if (h && h !== first) actions.push(h); });
      }
    } else if (s.status === 'error') {
      actions.push('<button class="' + (compact ? 'btn-outline' : 'btn-primary') + '" data-action="signal-retry" data-signal-id="' + s.id + '">' + escapeHtml(t('vigie.retry')) + '</button>');
    }
    if (!compact) {
      actions.push('<button class="btn-quiet signal-dismiss" data-action="signal-dismiss" data-signal-id="' + s.id + '">' + escapeHtml(t('vigie.dismiss')) + '</button>');
    }
    actions = actions.filter(Boolean);
    return actions.length ? '<div class="signal-actions">' + actions.join('') + '</div>' : '';
  }

  // Une PR ou une issue se présente comme un chantier : une carte qu'on ouvre
  // pour relire. opts.showProject : le nom du projet (écran Vigie, qui mélange
  // les projets) ; opts.actions : l'action principale sous le verdict.
  function buildSignalCardHTML(s, opts) {
    opts = opts || {};
    var project = state.projectsById[s.projectId];
    var meta = [];
    if (opts.showProject && project) meta.push(project.name);
    if (s.createdAt) meta.push(timeAgo(s.createdAt));
    var cls = 'kanban-card signal-card';
    if (s.status === 'ready' && s.priority === 'low') cls += ' signal-low';
    return '<article class="' + cls + '" data-action="open-signal" data-signal-id="' + s.id + '">' +
      '<div class="signal-card-head">' +
        '<span class="signal-kind-chip">' + signalKindIconHTML(s.kind) + escapeHtml(t('vigie.kind.' + s.kind)) + '</span>' +
        signalRefLinkHTML(s) +
        (s.status === 'ready' && s.priority === 'high' ? '<span class="signal-high">' + escapeHtml(t('vigie.high')) + '</span>' : '') +
        (s.newComments ? '<span class="signal-comments">' + escapeHtml(tCount('vigie.newComments', s.newComments)) + '</span>' : '') +
        '<span class="signal-when">' + escapeHtml(meta.join(' · ')) + '</span>' +
      '</div>' +
      '<h3 class="card-title signal-card-title">' + escapeHtml(s.title) + '</h3>' +
      signalStateHTML(s, 'signal-card-verdict') +
      (opts.actions ? buildSignalActionsHTML(s, true) : '') +
    '</article>';
  }

  // La fiche d'un signal. Deux colonnes, comme une page qu'on lit et une
  // décision qu'on prend : à gauche le texte (la revue proposée, puis celui de
  // l'auteur), à droite le jugement, les actions et GitHub, qui restent sous la
  // main pendant qu'on défile. Sous 1100 px, la décision passe au-dessus.
  function buildSignalDetailHTML() {
    var s = findSignal(state.signalId);
    if (!s) return buildHeaderHTML() + '<div class="view-body"></div>';
    var project = state.projectsById[s.projectId];
    var analyst = project && project.watch && state.agentsById[project.watch.agentId];
    var labels = (s.labels || []).map(function (l) { return '<span class="signal-label">' + escapeHtml(l) + '</span>'; }).join('');
    var byline = [];
    if (s.author) byline.push(t('vigie.by', { login: s.author }));
    if (s.createdAt) byline.push(t(s.kind === 'upkeep' ? 'vigie.detail.foundBy' : 'vigie.detail.opened', { when: timeAgo(s.createdAt) }));

    var reviewHTML = '';
    if (s.status === 'ready') {
      var heading = t(s.kind === 'upkeep' ? 'vigie.detail.plan' : (s.kind === 'pr' ? 'vigie.detail.review' : 'vigie.detail.reply'));
      reviewHTML = '<section class="sd-section sd-review">' +
        '<div class="sd-section-head"><h2>' + escapeHtml(heading) + '</h2>' +
          // Copier se tient aussi près du texte, sauf quand c'est déjà l'action
          // principale du panneau : deux boutons pour un geste, c'est un de trop.
          (s.review && signalPrimaryAction(s) !== 'copy' ? '<button class="btn-quiet" data-action="signal-copy" data-signal-id="' + s.id + '">' + escapeHtml(t('vigie.detail.copy')) + '</button>' : '') +
        '</div>' +
        (s.review
          ? '<div class="msg-text sd-prose">' + renderMarkdown(s.review) + '</div>' +
            (s.kind === 'upkeep' ? '' : '<div class="sd-note">' + escapeHtml(t(s.kind === 'pr' ? 'vigie.reviewNote' : 'vigie.replyNote')) + '</div>')
          : '<div class="sd-empty">' + escapeHtml(t('vigie.detail.nothingToAdd')) + '</div>') +
      '</section>';
    } else if (s.status !== 'error') {
      reviewHTML = '<section class="sd-section sd-review sd-live" data-signal-live="' + s.id + '">' + signalLiveInnerHTML(s) + '</section>';
    }
    // La discussion : les derniers messages sur la forge, ceux que l'analyste a lus.
    var discussionHTML = (s.comments || []).length ? '<section class="sd-section sd-discussion">' +
      '<div class="sd-section-head"><h2>' + escapeHtml(t('vigie.detail.discussion')) + '</h2>' +
        (s.newComments ? '<span class="signal-comments">' + escapeHtml(tCount('vigie.newComments', s.newComments)) + '</span>' : '') + '</div>' +
      s.comments.map(function (c) {
        return '<div class="sd-comment"><div class="sd-comment-head"><span class="sd-comment-author">@' + escapeHtml(c.author || '?') + '</span>' +
          '<span class="sd-comment-when">' + escapeHtml(timeAgo(c.createdAt)) + '</span></div>' +
          '<div class="msg-text sd-prose">' + renderMarkdown(c.body || '') + '</div></div>';
      }).join('') + '</section>' : '';
    var bodyHTML = s.kind === 'upkeep' ? '' : '<section class="sd-section sd-source">' +
      '<div class="sd-section-head"><h2>' + escapeHtml(t(s.kind === 'pr' ? 'vigie.detail.prBody' : 'vigie.detail.issueBody', { login: s.author || '?' })) + '</h2></div>' +
      (s.body ? '<div class="msg-text sd-prose">' + renderMarkdown(s.body) + '</div>' : '<div class="sd-empty">' + escapeHtml(t('vigie.detail.noBody')) + '</div>') +
    '</section>';

    // Le panneau de décision.
    var verdict = '';
    if (s.status === 'ready') {
      verdict = '<div class="sd-decision">' + signalDecisionHTML(s) + '</div>' +
        '<p class="sd-verdict">' + escapeHtml(s.verdict) + '</p>';
    } else if (s.status === 'error') {
      verdict = '<p class="sd-verdict signal-verdict-error">' + escapeHtml(signalErrorText(s)) + '</p>';
    } else {
      verdict = '<p class="sd-verdict signal-verdict-pending">' + escapeHtml(t('vigie.detail.waiting')) + '</p>';
    }
    var who = analyst
      ? '<div class="sd-analyst"><span class="agent-avatar-sm" style="background:' + softColor(analyst.color) + '">' + analyst.emoji + '</span>' +
        escapeHtml(t('vigie.detail.readBy', { agent: analyst.name })) +
        (s.analyzedAt ? '<span class="sd-analyst-when">' + escapeHtml(timeAgo(s.analyzedAt)) + '</span>' : '') + '</div>'
      : '';
    var github = s.url
      ? '<a class="sd-github" href="' + escapeHtml(s.url) + '" target="_blank" rel="noopener noreferrer" data-action="external">' +
          '<span class="sd-github-icon">' + signalKindIconHTML(s.kind) + '</span>' +
          '<span class="sd-github-text"><span class="sd-github-ref mono">' + escapeHtml(s.repoName) + '#' + s.number + '</span>' +
          '<span class="sd-github-sub">' + escapeHtml(t('vigie.detail.onGithub')) + '</span></span>' +
          externalIconHTML() +
        '</a>'
      : '';
    var aside = '<aside class="sd-aside"><div class="sd-panel">' +
        who + verdict +
        buildSignalActionsHTML(s, false, { noGithub: true }) +
      '</div>' + github + '</aside>';

    return buildHeaderHTML() + '<div class="view-body"><div class="sd-page">' +
      '<header class="sd-head">' +
        '<div class="sd-eyebrow">' + signalKindIconHTML(s.kind) + '<span>' + escapeHtml(t(s.kind === 'upkeep' ? 'vigie.detail.kindUpkeep' : (s.kind === 'pr' ? 'vigie.detail.kindPr' : 'vigie.detail.kindIssue'))) + '</span>' +
          (project ? '<span class="sd-dot">·</span><span>' + escapeHtml(project.name) + '</span>' : '') + '</div>' +
        '<h1 class="sd-title">' + escapeHtml(s.title) + '</h1>' +
        '<div class="sd-byline">' + escapeHtml(byline.join(' · ')) + labels + '</div>' +
      '</header>' +
      aside +
      '<div class="sd-main">' + reviewHTML + bodyHTML + discussionHTML + '</div>' +
    '</div></div>';
  }

  // Le journal dit ce que la Vigie a lu et pourquoi elle n'a rien remonté :
  // une Vigie muette doit pouvoir s'expliquer. projectId restreint au projet.
  function buildVigieLogHTML(projectId) {
    var entries = ((state.watch && state.watch.log) || []).filter(function (e) {
      return !projectId || e.projectId === projectId;
    });
    var open = !!state.vigieLogOpen;
    var html = '<div class="vigie-log">' +
      '<button class="btn-quiet vigie-log-toggle" data-action="vigie-log-toggle" aria-expanded="' + (open ? 'true' : 'false') + '">' +
      escapeHtml(t(open ? 'vigie.log.hide' : 'vigie.log.show')) + '</button>';
    if (open) {
      html += entries.length
        ? '<ul class="vigie-log-list">' + entries.map(function (e) {
            return '<li class="' + (e.error ? 'vigie-log-error' : '') + '"><span class="vigie-log-at mono">' + escapeHtml(formatTime(e.at)) + '</span>' +
              '<span>' + escapeHtml(watchLogText(e)) + '</span></li>';
          }).join('') + '</ul>'
        : '<div class="modal-note">' + escapeHtml(t('vigie.log.empty')) + '</div>';
    }
    return html + '</div>';
  }

  function watchLogText(e) {
    var project = state.projectsById[e.projectId];
    var where = e.repoName + (project && project.name !== e.repoName ? ' (' + project.name + ')' : '');
    var kind = t('vigie.kind.' + e.item);
    if (e.kind === 'error') return t('vigie.log.error', { repo: where, error: e.error });
    if (e.kind === 'upkeep') {
      return e.error ? t('vigie.log.upkeepFailed', { repo: where, error: e.error === 'stopped' ? t('vigie.stopped') : e.error }) : t('vigie.log.upkeep', { repo: where, n: e.added || 0 });
    }
    if (e.kind === 'analysis') {
      if (e.error) return t('vigie.log.analysisFailed', { kind: kind, n: e.number, repo: where, error: e.error });
      return t('vigie.log.analysis', { kind: kind, n: e.number, repo: where, s: Math.max(1, Math.round((e.durationMs || 0) / 1000)) });
    }
    var parts = [t('vigie.log.list', { repo: where, kind: kind, open: e.open, added: e.added })];
    if (e.updated) parts.push(t('vigie.log.updated', { n: e.updated }));
    if (e.known) parts.push(t('vigie.log.known', { n: e.known }));
    if (e.drafts) parts.push(t('vigie.log.drafts', { n: e.drafts }));
    if (e.older) {
      var d = new Date(e.from);
      parts.push(t('vigie.log.older', { n: e.older, date: isNaN(d.getTime()) ? '' : d.toLocaleDateString(state.lang === 'en' ? 'en-US' : 'fr-FR') }));
    }
    return parts.join(', ');
  }

  // Dans la vue d'un projet, ce que la Vigie y a trouvé, au-dessus du kanban.
  // Rien du tout tant que la Vigie n'est pas activée ou n'a rien à dire : le
  // kanban reste le sujet de la page.
  function findSignal(id) {
    return state.signals.filter(function (s) { return s.id === id; })[0] || null;
  }

  function upsertSignal(sig) {
    var i = state.signals.findIndex(function (x) { return x.id === sig.id; });
    if (i >= 0) state.signals[i] = sig; else state.signals.push(sig);
  }

  function refreshVigie() {
    renderSidebar();
    if (state.screen === 'vigie' || state.screen === 'kanban' || state.screen === 'signal' || state.screen === 'inbox') renderMain();
  }

  function vigieFail(err) {
    state.vigieError = (err && err.message) || t('common.networkError');
    refreshVigie();
  }

  function setProjectWatch(projectId, enabled, agentId) {
    state.vigieError = '';
    var body = { enabled: enabled };
    if (agentId) body.agentId = agentId;
    api('/api/projects/' + encodeURIComponent(projectId) + '/watch', { method: 'PUT', body: body })
      .then(function (p) { upsertProject(p); refreshVigie(); })
      .catch(vigieFail);
  }

  function vigieCheck() {
    state.vigieError = '';
    state.watch = Object.assign({}, state.watch || {}, { checking: true });
    refreshVigie();
    api('/api/watch/check', { method: 'POST' }).catch(vigieFail);
  }

  function patchSignal(id, body) {
    return api('/api/signals/' + encodeURIComponent(id), { method: 'PATCH', body: body })
      .then(function (sig) { upsertSignal(sig); refreshVigie(); return sig; });
  }

  // Classer est optimiste : c'est le geste qu'on répète en vidant la liste, il
  // ne doit pas attendre le serveur.
  function dismissSignal(id) {
    var s = findSignal(id);
    if (!s) return;
    var before = s.status;
    s.status = 'dismissed';
    if (state.signalId === id) { goBack(); }
    refreshVigie();
    patchSignal(id, { status: 'dismissed' }).catch(function (err) { s.status = before; vigieFail(err); });
  }

  function copySignalReview(id, el) {
    var s = findSignal(id);
    if (!s || !s.review) return;
    writeClipboard(s.review, function () {
      var before = el.textContent;
      el.textContent = t('vigie.copied');
      setTimeout(function () { el.textContent = before; }, 1300);
    });
  }

  // Un signal devient un chantier : la carte porte le constat en contexte (que
  // chaque tâche reçoit), le signal est marqué traité, puis la planification
  // s'ouvre pour découper le travail. Rien ne démarre sans que l'humain relise
  // et crée les tâches proposées.
  function signalToCard(id, el) {
    var s = findSignal(id);
    if (!s) return;
    var ctx = [s.kind === 'upkeep' ? t('vigie.upkeep.promptHeader', { repo: s.repoName, title: s.title }) : t('issues.promptHeader', { n: s.number, title: s.title })];
    if (s.url) ctx.push(s.url);
    if (s.verdict) ctx.push('', s.verdict);
    if (s.review) ctx.push('', s.review);
    if (el) el.disabled = true;
    api('/api/cards', { method: 'POST', body: { projectId: s.projectId, title: s.title, contextPrompt: ctx.join('\n') } })
      .then(function (card) {
        upsertCard(card);
        return patchSignal(id, { status: 'handled', cardId: card.id }).then(function () { return card; });
      })
      .then(function (card) {
        navigateTo('#/p/' + encodeURIComponent(card.projectId) + '/c/' + encodeURIComponent(card.id));
        setTimeout(function () { openPlanModal(card.id); }, 0);
      })
      .catch(function (err) { if (el) el.disabled = false; vigieFail(err); });
  }

  // Une issue devient un chantier : la carte est créée, le signal marqué traité,
  // et la création de tâche s'ouvre pré-remplie avec l'issue et l'analyse.
  // L'humain relit la consigne avant que le moindre agent démarre, comme pour
  // un import d'issue ordinaire.
  function startWorkstreamFromSignal(id) {
    var s = findSignal(id);
    if (!s) return;
    if (s.kind === 'upkeep') {
      var up = [t('vigie.upkeep.promptHeader', { repo: s.repoName, title: s.title }), '', s.verdict];
      if (s.review) up.push('', s.review);
      openNewTaskInProject(s.projectId, { title: s.title, prompt: up.join('\n'), repoName: s.repoName, waitsForTaskId: '' }, s.id);
      return;
    }
    var parts = [t('issues.promptHeader', { n: s.number, title: s.title })];
    if (s.url) parts.push(s.url);
    if (s.body) parts.push('', s.body);
    if (s.review) parts.push('', t('vigie.promptAnalysis'), s.review);
    openNewTaskInProject(s.projectId, { title: s.title, prompt: parts.join('\n'), repoName: s.repoName, waitsForTaskId: '' }, s.id);
  }

  // Notification système : seulement pour un signal qui vient d'être analysé,
  // qui n'est pas du bruit, et quand on ne regarde pas déjà la Vigie.
  function notifySignal(sig) {
    if (!canNotify() || Notification.permission !== 'granted') return;
    if (sig.priority === 'low') return;
    if (!document.hidden && state.screen === 'vigie') return;
    var project = state.projectsById[sig.projectId];
    var title = t('vigie.kind.' + sig.kind) + ' ' + signalRefText(sig) + (project ? ' · ' + project.name : '');
    try {
      var n = new Notification(title, { body: sig.verdict, tag: 'sillage-signal-' + sig.id });
      n.onclick = function () { window.focus(); openSignal(sig.id); n.close(); };
    } catch (e) {}
  }

  function onSignalEvent(sig) {
    if (!sig || !sig.id) return;
    var prev = findSignal(sig.id);
    upsertSignal(sig);
    if (sig.status === 'ready' && (!prev || prev.status !== 'ready')) notifySignal(sig);
    refreshVigie();
  }

  function onWatchEvent(w) {
    var prev = state.watch || {};
    state.watch = w || {};
    function key(c) { return c ? [c.kind, c.projectId, c.repoName, c.signalId].join('/') : ''; }
    var cur = key(state.watch.current);
    var was = key(prev.current);
    // Une étape de plus ne change que les zones de suivi ; le reste de l'écran
    // Vigie (journal, relève) ne se reconstruit que quand il a bougé.
    var moved = cur !== was || !!prev.checking !== !!state.watch.checking ||
      prev.checkedAt !== state.watch.checkedAt || (prev.log || []).length !== (state.watch.log || []).length;
    if (moved && state.screen === 'vigie') renderMain();
    patchSignalLive();
  }

  function stopSignal(id) {
    api('/api/signals/' + encodeURIComponent(id) + '/stop', { method: 'POST' }).catch(vigieFail);
  }

  // ---------------------------------------------------------------------
  // Rendu : Kanban
  // ---------------------------------------------------------------------

  function buildKanbanCardHTML(c, colKey) {
    var agents = activeAgentsForCard(c.id);
    var avatarsHTML = agents.map(function (a) {
      return '<span class="card-avatar" style="background:' + softColor(a.color) + '">' + a.emoji + '</span>';
    }).join('');
    // Même pastille animée que sur une tâche en cours (task-glyph-running) :
    // un agent qui tourne dans le chantier se repère sans lire chaque tâche.
    var runningGlyph = agents.length
      ? '<span class="task-glyph task-glyph-running" title="' + escapeHtml(t('status.running')) + '"><span class="task-spinner"></span></span>'
      : '';
    // Une barre de progression n'est pas une alerte : de l'encre grise tant que
    // le chantier avance, le vert seulement quand il est livré. Le bleu accent
    // reste réservé aux liens et aux anneaux de focus.
    var barColor = colKey === 'done' ? 'var(--green-live)' : 'var(--muted-3)';
    // Le rail n'apparaît qu'une fois qu'il a quelque chose à montrer. À zéro il
    // ne dit rien, et sur une carte aussi légère il devient le trait le plus
    // fort de la carte : un trait qui ne veut rien dire attire l'œil pour rien.
    var progressHTML = (c.tasksTotal && c.progress)
      ? '<div class="progress-track"><div class="progress-fill" style="width:' + c.progress + '%; background:' + barColor + '"></div></div>'
      : '';
    var liveHTML = c.liveActivity ? '<div class="card-live"><span class="live-dot"></span><span class="live-text mono">' +
      escapeHtml(c.liveActivity) + '</span></div>' : '';
    var attention = c.reviewCount ? '<span class="card-attention">' + escapeHtml(t('kanban.card.reviewCount', { n: c.reviewCount })) + '</span>' : '';
    var awaitingShip = c.awaitingShip
      ? '<span class="card-awaiting-ship">' + shipIconHTML() + escapeHtml(t('kanban.card.awaitingShip')) + '</span>'
      : '';
    // Les colonnes suivent les tâches : seul un chantier sans tâche se range à
    // la main (le classer, ou le remettre à lancer).
    var menuItemsHTML = '';
    if (!c.tasksTotal && !state.tasks.some(function (tk) { return tk.cardId === c.id; })) {
      var target = c.column === 'done' ? 'soon' : 'done';
      menuItemsHTML = '<button class="card-menu-item" data-action="move-card" data-card-id="' + c.id + '" data-column="' + target + '">' +
        escapeHtml(t('cardMenu.moveTo', { column: t(target === 'done' ? 'lane.shipped' : 'lane.next') })) + '</button>';
    }
    return '<article class="kanban-card" data-action="open-card" data-card-id="' + c.id + '">' +
      '<div class="card-top">' +
        runningGlyph +
        '<h3 class="card-title ' + (colKey === 'done' ? 'card-title-done' : '') + '">' + escapeHtml(c.title) + '</h3>' +
        '<div class="card-avatars">' + avatarsHTML + '</div>' +
        (menuItemsHTML
          ? '<button class="card-menu-btn" data-action="toggle-card-menu" data-card-id="' + c.id + '">⋯</button>' +
            '<div class="card-menu hidden" data-card-menu="' + c.id + '">' + menuItemsHTML + '</div>'
          : '') +
      '</div>' +
      liveHTML +
      // Une carte porte une décision, pas un tableau de bord. Restent l'avancement
      // (combien de tâches sont passées) et ce qui attend un humain ; les
      // compteurs de livrables et de messages ne décidaient de rien et se lisent
      // dans le chantier. Une carte sans tâche ne dit même pas « 0/0 ».
      '<div class="card-meta">' +
        (c.tasksTotal
          ? '<span>' + (c.tasksDone || 0) + '/' + c.tasksTotal + ' ' + escapeHtml(t('kanban.card.tasksLabel')) + '</span>'
          : '<span>' + escapeHtml(t('kanban.card.noTask')) + '</span>') +
        awaitingShip +
        attention +
      '</div>' +
      progressHTML +
      '</article>';
  }

  function buildPinnedLinksHTML(project) {
    var links = (project && project.links) || [];
    if (!links.length) return '';
    var items = links.map(function (l) {
      var host = '';
      try { host = new URL(l.url).host; } catch (e) {}
      var title = l.title || l.url;
      var iconHTML = host
        ? '<img src="https://' + escapeHtml(host) + '/favicon.ico" alt="" class="pinned-link-icon" onerror="window.__pinnedLinkIconError(this)">'
        : '<span class="pinned-link-fallback">🔗</span>';
      return '<a class="pinned-link" href="' + escapeHtml(l.url) + '" target="_blank" rel="noopener" title="' + escapeHtml(title) + '">' + iconHTML + '</a>';
    }).join('');
    return '<div class="pinned-links">' + items + '</div>';
  }

  // Tri stable et déterministe, comme compareTasksForList : updatedAt
  // décroissant (les chantiers actifs remontent), ref décroissante en
  // départage.
  // Le flux se lit « à qui la main » plutôt que « bientôt, en cours, terminé » :
  // avec des agents, l'avancement n'est plus le goulot, l'attention humaine
  // l'est. Les colonnes sont dérivées de l'état des tâches, sauf « Livré » qui
  // reste la colonne done du serveur (tout terminal et livré).
  var LANE_ORDER = ['next', 'agents', 'you', 'shipped'];
  function cardLane(c) {
    if (c.column === 'done') return 'shipped';
    if (c.reviewCount || c.awaitingShip) return 'you';
    var tasks = state.tasks.filter(function (tk) { return tk.cardId === c.id; });
    if (tasks.some(function (tk) { return tk.status === 'running'; })) return 'agents';
    if (!tasks.length || tasks.some(function (tk) { return tk.status === 'waiting'; })) return 'next';
    return 'you'; // du travail rendu ou refusé, rien qui tourne : une décision vous attend
  }
  // Un signal analysé attend une décision humaine ; en file ou en cours
  // d'analyse, c'est un agent qui a la main.
  function signalLane(sg) {
    return (sg.status === 'queued' || sg.status === 'analyzing') ? 'agents' : 'you';
  }

  function compareCardsForColumn(a, b) {
    var diff = new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime();
    if (diff !== 0) return diff;
    return (b.ref || 0) - (a.ref || 0);
  }

  function buildKanbanHTML() {
    var project = state.projectsById[state.projectId];
    var cards = state.cards.filter(function (c) { return c.projectId === state.projectId; });

    var signals = projectWatched(project) ? vigieSignals().filter(function (sg) { return sg.projectId === state.projectId; }) : [];
    var colsHTML = LANE_ORDER.map(function (key) {
      var list = cards.filter(function (c) { return cardLane(c) === key; }).sort(compareCardsForColumn);
      var laneSignals = signals.filter(function (sg) { return signalLane(sg) === key; });
      var addBtn = key === 'next'
        ? '<button class="add-card-btn" data-action="open-new-card">' + escapeHtml(t('kanban.addCard')) + '</button>'
        : '';
      return '<section class="kanban-col kanban-lane-' + key + '">' +
        '<div class="col-head"><span class="col-dot"></span>' +
        '<span class="col-label">' + escapeHtml(t('lane.' + key)) + '</span><span class="col-count">' + (list.length + laneSignals.length) + '</span></div>' +
        list.map(function (c) { return buildKanbanCardHTML(c, c.column === 'done' ? 'done' : key); }).join('') +
        laneSignals.map(function (sg) { return buildSignalCardHTML(sg, { actions: sg.status === 'ready' }); }).join('') +
        addBtn +
        '</section>';
    }).join('');

    var head = '<div class="kanban-head"><div class="kanban-title-col">' +
      '<div class="kanban-title-row"><h1>' + escapeHtml(project ? project.name : '') + '</h1>';
    if (project) {
      head += '<button class="icon-btn" data-action="open-edit-project" title="' + escapeHtml(t('project.editTooltip')) + '" aria-label="' + escapeHtml(t('project.editTooltip')) + '">✎</button>';
    }
    head += '</div>';
    if (project) {
      fetchClaudeSessions(project.id);
      head += '<div class="kanban-pills">' + buildAutoconfPillHTML(project) + buildSessionsPillHTML(project) + '</div>';
    }
    if (project) {
      head += '<div class="kanban-actions">' +
        buildCreateButtonHTML('open-new-project-task', t('header.newTask'), t('kanban.newTaskTooltip'), 'btn-primary') +
        '<button class="btn-quiet" data-action="open-new-card">' + escapeHtml(t('kanban.newWorkstream')) + '</button>' +
        '</div>';
    }
    if (project && project.description) {
      head += '<div class="kanban-description">' + escapeHtml(project.description) + '</div>';
    }
    head += '</div>';
    head += '</div>';
    head += buildPinnedLinksHTML(project);

    // Un projet n'est vide que sans chantier ni signal : une PR à relire, c'est
    // déjà quelque chose en cours.
    var emptyNote = cards.length === 0 && signals.length === 0
      ? '<button class="empty-cta" data-action="open-new-project-task">' +
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
    var glyphHTML = buildTaskGlyphHTML(t);
    return '<div class="task-row ' + (selected ? 'selected' : '') + '" data-action="open-task" data-task-id="' + t.id + '">' +
      glyphHTML +
      '<div class="task-main">' +
        '<div class="task-title-line"><span class="' + titleClass + '">' + escapeHtml(t.title) + '</span>' +
        (isNew ? '<span class="badge-new">' + escapeHtml(t2('badge.new')) + '</span>' : '') +
        buildBehindBadgeHTML(t) + '</div>' +
        '<div class="task-meta-line">' + projectTag +
        '<span class="mono">#' + t.ref + '</span>' +
        '<span class="agent-chip"><span class="agent-avatar-sm" style="background:' + softColor(agent.color) + '">' + agent.emoji + '</span>' + escapeHtml(agent.name) + '</span>' +
        '<span>' + escapeHtml(timeAgo(t.updatedAt)) + '</span>' + liveLine +
        '</div>' +
      '</div>' +
      buildTaskRowStateHTML(t) +
      buildTaskCountsHTML(t) +
      '</div>';
  }

  // Compteurs d'une ligne de tâche. Un zéro n'apprend rien : « 0 nouveaux
  // commits » sur chaque ligne, c'est du texte à sauter des yeux à chaque fois.
  // Ne restent que les deux nombres qui disent l'ampleur du travail rendu, et
  // seulement quand ils existent. Le nombre de messages se lit dans l'onglet
  // Conversation, qui le porte déjà.
  function buildTaskCountsHTML(task) {
    var parts = [];
    if (task.filesCount) {
      parts.push('<span title="' + escapeHtml(tCount('task.filesCount', task.filesCount)) + '">' +
        escapeHtml(tCount('task.filesCount', task.filesCount)) + '</span>');
    }
    if (task.commitsCount) {
      parts.push('<span>' + escapeHtml(tCount('ship.commits', task.commitsCount)) + '</span>');
    }
    if (!parts.length) return '';
    return '<div class="task-counts">' + parts.join('') + '</div>';
  }

  // Pastille d'état d'une tâche : le repère qu'on cherche des yeux en balayant
  // une liste. Cercle teinté (une couleur par état) plutôt qu'un caractère nu,
  // avec l'état en infobulle. Deux états animés : l'agent travaille (fuseau
  // gris), un rebase automatique est en cours (fuseau ambre).
  function buildTaskGlyphHTML(task) {
    var glyph = STATUS_GLYPH[task.status] || STATUS_GLYPH.running;
    if (task.rebasing) {
      return '<span class="task-glyph task-glyph-rebasing" title="' + escapeHtml(t2('status.rebasing')) + '">' +
        '<span class="task-spinner"></span></span>';
    }
    if (task.status === 'running') {
      return '<span class="task-glyph task-glyph-running" title="' + escapeHtml(t2('status.running')) + '">' +
        '<span class="task-spinner"></span></span>';
    }
    return '<span class="task-glyph task-glyph-' + task.status + '" title="' + escapeHtml(t2('status.' + task.status)) + '"' +
      ' style="color:' + glyph.color + '">' + glyph.icon + '</span>';
  }

  // Retard d'une tâche sur la branche de son chantier, tel que l'annonce le
  // dernier aperçu de livraison (state.deliveryByCard, rechargé à l'ouverture
  // du chantier et toutes les 60 s). Absent de la table = à jour.
  function taskBehindCount(task) {
    var prev = state.deliveryByCard[task.cardId];
    if (!prev || !prev.behind) return 0;
    return prev.behind[task.id] || 0;
  }

  // Badge « en retard » : un compteur discret, avec l'explication en infobulle.
  function buildBehindBadgeHTML(task) {
    var n = taskBehindCount(task);
    if (!n) return '';
    var tip = tCount('behind.taskTooltip', n, { base: task.base || '' });
    return '<span class="behind-badge" title="' + escapeHtml(tip) + '">↓' + n + '</span>';
  }

  // État d'une tâche dans la liste : lisible sans clic (acceptée, refusée), et
  // pour une tâche à relire, les deux boutons qui apparaissent au survol.
  // Un clic sur ces boutons agit tout de suite : l'acceptation est locale
  // (fusion dans la branche du chantier) et réversible, donc sans confirmation.
  function buildTaskRowStateHTML(t) {
    // Pendant un rebase, ni Accepter ni Refuser : le worktree de la tâche est en
    // train d'être rejoué, et le serveur refuserait de toute façon.
    if (t.rebasing) {
      return '<div class="task-row-state task-row-state-rebasing">' + escapeHtml(t2('status.rebasing')) + '</div>';
    }
    if (t.status === 'waiting') {
      var dep = state.tasksById[t.waitsForTaskId];
      var waitLabel = dep ? t2('taskRow.waitingFor', { title: dep.title }) : t2('status.waiting');
      return '<div class="task-row-state task-row-state-waiting" title="' + escapeHtml(waitLabel) + '">' + escapeHtml(t2('status.waiting')) + '</div>';
    }
    if (t.status === 'review') {
      // Acceptation en vol : les deux boutons se figent le temps de la fusion,
      // sans changer de libellé (« Acceptation… » élargirait la ligne entière
      // pour un dixième de seconde). Le retour, c'est l'onde de playWake.
      var busy = !!state.acceptingByTask[t.id];
      var busyAttr = busy ? ' disabled' : '';
      return '<div class="task-row-actions' + (busy ? ' task-row-actions-busy' : '') + '">' +
        '<button class="row-btn row-btn-accept"' + busyAttr + ' data-action="accept-task" data-task-id="' + t.id + '" title="' + escapeHtml(t2('action.acceptTooltip')) + '">' + escapeHtml(t2('action.accept')) + '</button>' +
        '<button class="row-btn"' + busyAttr + ' data-action="refuse-task" data-task-id="' + t.id + '" title="' + escapeHtml(t2('action.refuseTooltip')) + '">' + escapeHtml(t2('action.refuse')) + '</button>' +
        '</div>';
    }
    if (t.status === 'accepted' || t.status === 'cancelled') {
      return '<div class="task-row-state task-row-state-' + t.status + '">' + escapeHtml(t2('status.' + t.status)) + '</div>';
    }
    return '';
  }
  // Alias local pour éviter un conflit de nom avec le paramètre `t` (Task) ci-dessus.
  function t2(key, vars) { return t(key, vars); }

  function taskMatchesFilter(t, filter) {
    if (filter === 'all') return true;
    if (filter === 'finished') return t.status === 'accepted' || t.status === 'cancelled';
    return t.status === filter;
  }

  // Tri stable et déterministe : updatedAt décroissant, ref décroissante en
  // départage (deux tâches ne changent jamais de position relative tant que
  // ni l'un ni l'autre de ces deux champs ne change réellement ; ouvrir une
  // tâche ne doit provoquer aucun re-tri visible).
  function compareTasksForList(a, b) {
    var diff = new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime();
    if (diff !== 0) return diff;
    return (b.ref || 0) - (a.ref || 0);
  }

  function buildTaskListPaneInnerHTML(tasksAll, opts) {
    var counts = tasksAll.reduce(function (acc, t) { acc[t.status] = (acc[t.status] || 0) + 1; return acc; }, {});
    var finishedCount = (counts.accepted || 0) + (counts.cancelled || 0);
    var filters = [
      { key: 'all', label: t('filter.all', { n: tasksAll.length }) },
      { key: 'waiting', label: t('filter.waiting', { n: counts.waiting || 0 }) },
      { key: 'running', label: t('filter.running', { n: counts.running || 0 }) },
      { key: 'review', label: t('filter.review', { n: counts.review || 0 }) },
      { key: 'finished', label: t('filter.finished', { n: finishedCount }) }
    ];
    // Un filtre vide ne filtre rien : il disparaît, sauf s'il est actif (on ne
    // retire pas le sol sous les pieds de qui vient de cliquer dessus).
    var filterCounts = { all: tasksAll.length, waiting: counts.waiting || 0, running: counts.running || 0, review: counts.review || 0, finished: finishedCount };
    filters = filters.filter(function (f) { return f.key === 'all' || filterCounts[f.key] || state.taskFilter === f.key; });
    var filterHTML = filters.length < 2 ? '' : filters.map(function (f) {
      return '<button class="pill ' + (state.taskFilter === f.key ? 'pill-active' : '') + '" data-action="set-filter" data-filter="' + f.key + '">' + escapeHtml(f.label) + '</button>';
    }).join('');
    // Le plan se tient à côté de « Nouvelle tâche », pas dans la barre du haut
    // déjà occupée par Recette et Livrer : les deux boutons créent du travail,
    // c'est le même geste à deux échelles (une tâche, ou le découpage entier).
    var newTaskHTML = state.cardId
      ? '<div class="filter-pills-actions">' +
        '<button class="btn-neutral plan-btn" data-action="open-plan" title="' + escapeHtml(t('plan.buttonTooltip')) + '">' + escapeHtml(t('plan.button')) + '</button>' +
        buildCreateButtonHTML('open-new-task', t('header.newTask'), t('header.newTaskTooltip'), 'btn-outline') +
        '</div>'
      : '';
    var visible = tasksAll.filter(function (t) { return taskMatchesFilter(t, state.taskFilter); })
      .sort(compareTasksForList);
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
    return '<div class="filter-pills-row"><div class="filter-pills">' + filterHTML + '</div>' + newTaskHTML + '</div>' +
      '<div class="task-list">' + rowsHTML + '</div>';
  }

  function buildTaskListShell(tasksAll, opts) {
    var paneInnerHTML = buildTaskListPaneInnerHTML(tasksAll, opts);
    var task = state.taskId ? state.tasksById[state.taskId] : null;
    var panelHTML = task ? buildDetailPanelHTML(task) : '';
    var card = state.cardId ? state.cardsById[state.cardId] : null;
    var bodyClass = 'view-body work-body' + (opts.bodyClass ? ' ' + opts.bodyClass : '') + (task ? ' has-panel' : '') + (task && state.panelExpanded ? ' panel-expanded' : '');
    return buildHeaderHTML() + (opts.bannerHTML || '') + (card ? buildShipBarHTML(card) : '') +
      '<div class="' + bodyClass + '" style="padding:0;">' +
      '<div class="task-list-pane">' + (opts.paneTopHTML || '') + paneInnerHTML + '</div>' +
      panelHTML +
      '</div>';
  }

  // ---------------------------------------------------------------------
  // Livraison d'un chantier (barre d'en-tête + récapitulatif)
  // ---------------------------------------------------------------------

  // Rafraîchissement périodique de l'aperçu tant qu'une vue chantier est
  // ouverte : c'est cet appel qui constate côté serveur les branches fusionnées
  // à la main (elles passent alors « acceptées », voir autoAcceptMergedTasks).
  var deliveryPollTimer = null;
  var deliveryPollInterval = 60000;

  function syncDeliveryPolling() {
    if (deliveryPollTimer) { clearInterval(deliveryPollTimer); deliveryPollTimer = null; }
    if (!state.cardId) return;
    deliveryPollTimer = setInterval(function () {
      if (!state.cardId) return;
      loadDelivery(state.cardId);
    }, deliveryPollInterval);
  }

  // loadDelivery récupère l'aperçu de livraison d'un chantier (lecture seule
  // côté serveur : mode, dépôts, commits, avertissements, blocage).
  // Le réglage de livraison d'un projet vient de changer : les aperçus déjà
  // chargés annoncent l'ancien mode (« pousse la branche et ouvre une pull
  // request » alors que le projet fusionne désormais en local). On les jette, et
  // on recharge tout de suite celui du chantier ouvert.
  function invalidateDeliveryForProject(projectId) {
    state.cards.forEach(function (c) {
      if (c.projectId === projectId) delete state.deliveryByCard[c.id];
    });
    var current = state.cardId && state.cardsById[state.cardId];
    if (current && current.projectId === projectId) loadDelivery(state.cardId);
  }

  function loadDelivery(cardId) {
    var key = 'delivery:' + cardId;
    if (state.loading[key]) return;
    state.loading[key] = true;
    api('/api/cards/' + cardId + '/delivery').then(function (data) {
      state.deliveryByCard[cardId] = data;
    }).catch(function () {
      delete state.deliveryByCard[cardId];
    }).then(function () {
      delete state.loading[key];
      refreshShipBar(cardId);
      // Les badges de retard vivent aussi dans les lignes de tâches et dans
      // l'en-tête du détail : deux rafraîchissements ciblés, jamais un rendu
      // complet, pour ne pas toucher au fil de conversation ni au composeur.
      if (state.cardId === cardId) refreshTaskListAndFilters();
      if (state.taskId) patchDetailHead(state.taskId);
      // Toutes les surfaces du sillage viennent d'être reconstruites : c'est le
      // moment de l'onde, jamais avant (voir armWake).
      playPendingWake();
    });
  }

  function shipBlockerLabel(blocker) {
    switch (blocker) {
      case 'no-tasks': return t('ship.blocked.noTasks');
      case 'nothing-accepted': return t('ship.blocked.nothingAccepted');
      case 'nothing-to-ship': return t('ship.blocked.nothingToShip');
      default: return '';
    }
  }

  // Dépôts qui ont réellement quelque chose à livrer (commits non poussés, ou
  // pas encore fusionnés dans la branche de destination).
  function deliverableRepos(prev) {
    return (prev && prev.repos ? prev.repos : []).filter(function (r) { return r.pending > 0; });
  }

  // Livraison partielle : des tâches sont encore en cours ou à relire. Ce n'est
  // pas un blocage (la branche du chantier ne contient que le travail accepté,
  // donc rien de non relu ne peut partir), seulement ce qu'il faut savoir avant
  // de cliquer : ce qui reste se livrera plus tard.
  function partialTaskCount(prev) {
    if (!prev || !deliverableRepos(prev).length) return 0;
    return (prev.counts && prev.counts.pending) || 0;
  }

  // Les deux modes qui fusionnent dans la branche de destination, par
  // opposition aux deux qui livrent la branche du chantier elle-même.
  function deliveryModeMerges(mode) {
    return mode === 'merge' || mode === 'merge-push';
  }

  // deliveryActionLabel annonce l'action exacte, avant tout clic.
  function deliveryActionLabel(prev) {
    var repos = deliverableRepos(prev);
    if (!repos.length) return t('ship.subtext.nothing');
    if (repos.length > 1) return t('ship.subtext.repos', { n: repos.length });
    var r = repos[0];
    var base = prev.target || r.base;
    if (prev.mode === 'merge') return t('ship.subtext.merge', { branch: r.branch, base: base });
    if (prev.mode === 'merge-push') return t('ship.subtext.mergePush', { branch: r.branch, base: base });
    if (prev.mode === 'push') return t('ship.subtext.push', { branch: r.branch });
    if (prev.provider === 'gitlab') return t('ship.subtext.mr', { branch: r.branch, base: r.base });
    if (prev.provider === 'github') return t('ship.subtext.pr', { branch: r.branch, base: r.base });
    return t('ship.subtext.unknownForge', { branch: r.branch });
  }

  function deliveryWarningText(warning) {
    var w = String(warning || '');
    if (w.indexOf('gh not found') === 0) return t('delivery.warning.ghMissing');
    if (w.indexOf('glab not found') === 0) return t('delivery.warning.glabMissing');
    if (w.indexOf('no \'origin\' remote') === 0) return t('delivery.warning.noRemote');
    if (w.indexOf('unknown forge') === 0) return t('delivery.warning.unknownForge');
    return w;
  }

  // Lien d'installation à côté de l'avertissement, dans les réglages du projet
  // (voir buildDeliveryPanelHTML) : le CLI manquant se règle une fois, pas par
  // un rappel silencieux qu'on ne peut pas suivre.
  var CLI_DOC_LINKS = { gh: 'https://cli.github.com/', glab: 'https://gitlab.com/gitlab-org/cli' };
  function deliveryWarningLinkHTML(warning) {
    var s = String(warning || '');
    var cli = s.indexOf('gh not found') === 0 ? 'gh' : (s.indexOf('glab not found') === 0 ? 'glab' : '');
    if (!cli) return '';
    return ' <a href="' + CLI_DOC_LINKS[cli] + '" target="_blank" rel="noopener noreferrer">' +
      escapeHtml(t('delivery.warning.' + cli + 'InstallLink')) + '</a>';
  }

  // buildShipLinksHTML : ce qui est déjà livré (une ligne par dépôt).
  function buildShipLinksHTML(prev) {
    // La pull request reste affichée même après l'arrivée de travail nouveau
    // (elle existe toujours, et la prochaine livraison la mettra à jour) ; la
    // mention « poussée » ou « fusionnée » ne s'affiche que si tout est livré,
    // et seulement si le chantier n'est pas déjà « arrivé » : l'ancre de
    // l'en-tête le dit déjà (voir buildShipButtonHTML), et un simple texte ne
    // vaut pas le doublon, contrairement à un lien cliquable.
    var anchored = shipAlreadyOnTarget(prev);
    var shipped = (prev.repos || []).filter(function (r) { return r.prUrl || (r.shippedAt && !anchored); });
    if (!shipped.length) return '';
    var items = shipped.map(function (r) {
      if (r.prUrl) {
        return '<a class="ship-link" href="' + escapeHtml(r.prUrl) + '" target="_blank" rel="noopener">' +
          escapeHtml(t('ship.prLink')) + (prev.repos.length > 1 ? ' (' + escapeHtml(r.repoName) + ')' : '') + '</a>';
      }
      // Pas d'URL : fusion (locale ou poussée), ou branche poussée sans PR.
      var note;
      if (prev.mode === 'merge') note = t('ship.mergedInto', { base: prev.target || r.base });
      else if (prev.mode === 'merge-push') note = t('ship.mergedPushed', { base: prev.target || r.base });
      else note = t('ship.pushedBranch', { branch: r.branch });
      return '<span class="ship-link-note">' + escapeHtml(note) + '</span>';
    }).join('');
    return '<div class="ship-links">' + items + '</div>';
  }

  // Bateau de la livraison : un SVG inline plutôt qu'un emoji (illisible à cette
  // taille, et dessiné différemment par chaque système) et plutôt qu'une police
  // d'icônes (le frontend n'a aucune dépendance et reste embarqué). Grand-voile,
  // foc et coque, en currentColor : l'icône suit la couleur du bouton.
  function repoIconHTML() {
    return '<svg class="icon-link-svg" viewBox="0 0 16 16" width="15" height="15" aria-hidden="true" focusable="false">' +
      '<path fill="currentColor" d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0 0 16 8c0-4.42-3.58-8-8-8Z"/>' +
      '</svg>';
  }

  function sponsorIconHTML() {
    return '<svg class="icon-link-svg" viewBox="0 0 16 16" width="15" height="15" aria-hidden="true" focusable="false">' +
      '<path fill="currentColor" d="M4.25 2.5c-1.336 0-2.75 1.164-2.75 3 0 2.15 1.58 4.144 3.365 5.682A20.565 20.565 0 0 0 8 13.393a20.561 20.561 0 0 0 3.135-2.211C12.92 9.644 14.5 7.65 14.5 5.5c0-1.836-1.414-3-2.75-3-1.373 0-2.609.986-3.029 2.456a.75.75 0 0 1-1.442 0C6.859 3.486 5.623 2.5 4.25 2.5Z"/>' +
      '</svg>';
  }

  function shipIconHTML() {
    return '<svg class="btn-icon" viewBox="0 0 16 16" width="14" height="14" aria-hidden="true" focusable="false">' +
      '<path fill="currentColor" d="M7.6 1.5a.5.5 0 0 1 .95-.2l3.2 6.4a.5.5 0 0 1-.45.72H8.1a.5.5 0 0 1-.5-.5V1.5Z"/>' +
      '<path fill="currentColor" d="M6.3 8.44H4.1a.5.5 0 0 1-.42-.77l2.2-3.4a.5.5 0 0 1 .92.27v3.4a.5.5 0 0 1-.5.5Z"/>' +
      '<path fill="currentColor" d="M2 9.9h12a.6.6 0 0 1 .56.82l-1.1 2.7a2.2 2.2 0 0 1-2.04 1.37H4.58A2.2 2.2 0 0 1 2.54 13.4l-1.1-2.7A.6.6 0 0 1 2 9.9Z"/>' +
      '</svg>';
  }

  // Ancre : le chantier est arrivé à destination, il n'y a plus rien à livrer.
  // Remplace le bouton, plutôt qu'un bouton désactivé qui laisserait croire
  // qu'il reste quelque chose à faire.
  function anchorIconHTML() {
    return '<svg class="btn-icon" viewBox="0 0 16 16" width="14" height="14" aria-hidden="true" focusable="false">' +
      '<path fill="currentColor" d="M8 1a2.1 2.1 0 0 1 .7 4.08V6.7h1.6a.65.65 0 0 1 0 1.3H8.7v5.06a4.3 4.3 0 0 0 3.4-3.36h-.85a.5.5 0 0 1-.37-.84l1.6-1.75a.5.5 0 0 1 .74 0l1.6 1.75a.5.5 0 0 1-.37.84h-.93A5.7 5.7 0 0 1 8 15a5.7 5.7 0 0 1-5.52-5.2H1.55a.5.5 0 0 1-.37-.84l1.6-1.75a.5.5 0 0 1 .74 0l1.6 1.75a.5.5 0 0 1-.37.84H3.9a4.3 4.3 0 0 0 3.4 3.36V8H5.7a.65.65 0 0 1 0-1.3h1.6V5.08A2.1 2.1 0 0 1 8 1Zm0 1.3a.8.8 0 1 0 0 1.6.8.8 0 0 0 0-1.6Z"/>' +
      '</svg>';
  }

  // Icônes de la conversation : mêmes règles que les autres (SVG en ligne, en
  // currentColor, jamais d'emoji illisible sous 16px).
  function copyIconHTML() {
    return '<svg viewBox="0 0 16 16" width="13" height="13" aria-hidden="true" focusable="false">' +
      '<path fill="currentColor" d="M6.25 1.5h4.9c.9 0 1.6.7 1.6 1.6v6.05c0 .9-.7 1.6-1.6 1.6h-4.9c-.9 0-1.6-.7-1.6-1.6V3.1c0-.9.7-1.6 1.6-1.6Zm0 1.3a.3.3 0 0 0-.3.3v6.05c0 .17.13.3.3.3h4.9a.3.3 0 0 0 .3-.3V3.1a.3.3 0 0 0-.3-.3h-4.9Z"/>' +
      '<path fill="currentColor" d="M3.5 5.05a.65.65 0 0 1 .65.65v6.6c0 .17.13.3.3.3h5.2a.65.65 0 0 1 0 1.3h-5.2c-.9 0-1.6-.7-1.6-1.6V5.7a.65.65 0 0 1 .65-.65Z"/>' +
      '</svg>';
  }

  // Coche nue de la confirmation de copie, à ne pas confondre avec la pastille
  // cerclée des checks de projet (checkIconHTML plus bas).
  function copiedIconHTML() {
    return '<svg viewBox="0 0 16 16" width="13" height="13" aria-hidden="true" focusable="false">' +
      '<path fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" d="M3.2 8.6l3 3L12.8 4.9"/>' +
      '</svg>';
  }

  // Trombone du bouton joindre : discret, au trait, comme les autres icônes de
  // bouton (jamais d'emoji, illisible sous 16px).
  function attachIconHTML() {
    return '<svg viewBox="0 0 16 16" width="14" height="14" aria-hidden="true" focusable="false">' +
      '<path fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" d="M11.9 7.3 7.2 12a2.55 2.55 0 0 1-3.6-3.6l4.9-4.9a1.7 1.7 0 0 1 2.4 2.4l-4.9 4.9a.85.85 0 0 1-1.2-1.2l4.5-4.5"/>' +
      '</svg>';
  }

  // Issue ouverte : le cercle pointé des forges, pas leur logo. Le logo GitHub
  // serait une marque posée dans l'interface, et il ne dit pas « issue ».
  function issueIconHTML() {
    return '<svg viewBox="0 0 16 16" width="14" height="14" aria-hidden="true" focusable="false">' +
      '<circle cx="8" cy="8" r="6.2" fill="none" stroke="currentColor" stroke-width="1.5"/>' +
      '<circle cx="8" cy="8" r="1.9" fill="currentColor"/>' +
      '</svg>';
  }

  // Croix de retrait d'une vignette jointe.
  function closeIconHTML() {
    return '<svg viewBox="0 0 16 16" width="11" height="11" aria-hidden="true" focusable="false">' +
      '<path fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" d="M4.2 4.2l7.6 7.6M11.8 4.2l-7.6 7.6"/>' +
      '</svg>';
  }

  function arrowDownIconHTML() {
    return '<svg viewBox="0 0 16 16" width="12" height="12" aria-hidden="true" focusable="false">' +
      '<path fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" d="M8 3.2v9.2M4.3 8.9 8 12.6l3.7-3.7"/>' +
      '</svg>';
  }

  // Flèche vers l'extérieur : ouverture dans une application tierce, pas le
  // logo d'un éditeur en particulier.
  function externalIconHTML() {
    return '<svg viewBox="0 0 16 16" width="12" height="12" aria-hidden="true" focusable="false">' +
      '<path fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" d="M6.5 3H3.6c-.6 0-1.1.5-1.1 1.1v7.3c0 .6.5 1.1 1.1 1.1h7.3c.6 0 1.1-.5 1.1-1.1V9.5M9 2.5h4v4M7.2 8.3l5.6-5.6"/>' +
      '</svg>';
  }

  // Total de commits de la branche du chantier (base..branche), tous dépôts
  // confondus : un chantier multi-dépôts en a une somme, pas une liste.
  function deliveryTotalCommits(prev) {
    var repos = (prev && prev.repos) || [];
    return repos.reduce(function (sum, r) { return sum + (r.commits || 0); }, 0);
  }

  // Chip « N commits », dans l'esprit d'un badge GitHub : à côté du bouton de
  // livraison ou de l'ancre « déjà arrivé », peu importe l'état du chantier.
  function buildCommitsChipHTML(prev) {
    if (!prev || !prev.repos || !prev.repos.length) return '';
    return '<span class="ship-commit-chip">' + escapeHtml(tCount('ship.commits', deliveryTotalCommits(prev))) + '</span>';
  }

  // Chantier entièrement arrivé dans sa branche de destination : plus rien à
  // livrer, ni par Sillage ni à la main (voir mergedIntoTarget côté serveur).
  function shipAlreadyOnTarget(prev) {
    var repos = (prev && prev.repos) || [];
    if (!repos.length) return false;
    return repos.every(function (r) { return r.mergedIntoTarget; });
  }

  // Fusion impossible : les deux modes de fusion sont fast-forward uniquement, et
  // la destination a avancé de son côté. Livrer échouerait à coup sûr ; le badge
  // de retard du chantier dit déjà quoi faire (rebaser).
  function shipNeedsRebase(prev) {
    if (!prev || !deliveryModeMerges(prev.mode)) return false;
    return deliverableRepos(prev).some(function (r) { return !r.fastForwardable; });
  }

  // buildShipButtonHTML : le bouton lui-même, affiché dans l'en-tête. L'état
  // vient de la carte (shipReady / shipBlocker, toujours à jour via SSE) ;
  // l'aperçu (branche, base, commits) vient de GET /api/cards/{id}/delivery,
  // chargé à la demande. Le tout dans un emplacement stable (.ship-slot), parce
  // que ce n'est pas toujours un bouton qui s'y affiche.
  function buildShipButtonHTML(card) {
    var prev = state.deliveryByCard[card.id];
    if (!prev) loadDelivery(card.id);
    var blocked = !card.shipReady;
    var sub;
    if (blocked) sub = shipBlockerLabel(card.shipBlocker);
    else if (prev) sub = deliveryActionLabel(prev);
    else sub = t('common.loading');

    if (prev && shipAlreadyOnTarget(prev)) {
      var base = prev.target || (prev.repos[0] && prev.repos[0].base) || '';
      var label = t('ship.alreadyOnTarget', { base: base });
      var tip = t('ship.alreadyOnTargetSub', { base: base });
      return '<span class="ship-slot"><span class="ship-anchored" title="' + escapeHtml(tip) + '">' +
        anchorIconHTML() + escapeHtml(label) + '</span>' + buildCommitsChipHTML(prev) + '</span>';
    }
    if (prev && shipNeedsRebase(prev)) {
      sub = t('ship.blocked.behindTarget', { base: catchUpTarget(prev) });
    }
    var ready = !blocked && !!prev && deliverableRepos(prev).length > 0 && !shipNeedsRebase(prev);
    // Rattraper la destination est le geste qui débloque la livraison : le
    // bouton se tient donc à côté de celui qu'il débloque.
    var catchUp = '';
    if (prev && cardNeedsCatchUp(prev)) {
      var target = catchUpTarget(prev);
      catchUp = '<button class="btn-outline catch-up-btn" data-action="catch-up-card" data-card-id="' + card.id + '"' +
        ' title="' + escapeHtml(t('catchUp.tooltip', { base: target })) + '">' +
        escapeHtml(t('catchUp.button', { base: target })) + '</button>';
    }
    return '<span class="ship-slot">' + catchUp +
      '<button class="btn-green ship-btn" ' + (ready ? '' : 'disabled title="' + escapeHtml(sub) + '"') +
      ' data-action="open-ship-modal" data-card-id="' + card.id + '">' +
      shipIconHTML() + escapeHtml(t('ship.button')) + '</button>' + buildCommitsChipHTML(prev) + '</span>';
  }

  // Un chantier a besoin d'être rattrapé dès que sa destination a avancé sans
  // lui, quel que soit le mode : en fusion la livraison est bloquée, en pull
  // request elle partirait sur une base périmée.
  function cardNeedsCatchUp(prev) {
    var repos = (prev && prev.repos) || [];
    if (!repos.length || shipAlreadyOnTarget(prev)) return false;
    return repos.some(function (r) { return r.behind > 0 || !r.fastForwardable; });
  }

  function catchUpTarget(prev) {
    var repos = (prev && prev.repos) || [];
    return prev.target || (repos[0] && repos[0].base) || '';
  }

  // Le CLI de la forge (gh/glab) manquant est un problème d'environnement, pas
  // du chantier ouvert : il est identique sur toutes les cartes du projet, et se
  // règle une fois pour toutes (voir buildDeliveryPanelHTML, réglages du
  // projet). L'annoncer ici en plus ferait doublon sans rien apporter.
  function isCliMissingWarning(w) {
    var s = String(w || '');
    return s.indexOf('gh not found') === 0 || s.indexOf('glab not found') === 0;
  }

  // buildShipBarHTML : l'aperçu de livraison (compteurs, avertissements, liens
  // déjà livrés) ; le bouton lui-même et le mode de livraison sont dans
  // l'en-tête et la modale Ship (voir buildShipButtonHTML, buildShipModalHTML).
  // N'affiche rien (bandeau invisible, mais toujours présent dans le DOM pour
  // que refreshShipBar retrouve son nœud) tant qu'il n'y a rien à signaler.
  function buildShipBarHTML(card) {
    var prev = state.deliveryByCard[card.id];
    var blocked = !card.shipReady;
    var anchored = !!(prev && shipAlreadyOnTarget(prev));
    var sub = '';
    // Le bouton désactivé porte déjà la raison en infobulle (voir
    // buildShipButtonHTML) : la répéter ici la rend visible sans avoir à
    // survoler, ce qui reste utile pour "bloqué" et "en retard". "Déjà arrivé"
    // et "prêt à livrer" ont chacun leur propre annonce ailleurs (l'ancre, la
    // modale Ship) : rien à répéter ici.
    if (!anchored) {
      if (blocked) sub = shipBlockerLabel(card.shipBlocker);
      else if (prev && shipNeedsRebase(prev)) {
        sub = t('ship.blocked.behindTarget', { base: prev.target || (deliverableRepos(prev)[0] || {}).base || '' });
      }
    }
    var warningsList = (prev && prev.warnings ? prev.warnings : []).filter(function (w) { return !isCliMissingWarning(w); });
    var warningsHTML = warningsList.map(function (w) {
      return '<div class="ship-bar-warning">⚠ ' + escapeHtml(deliveryWarningText(w)) + '</div>';
    }).join('');
    var catchUpError = state.catchUpErrorByCard[card.id];
    if (catchUpError) warningsHTML += '<div class="ship-bar-warning">⚠ ' + escapeHtml(catchUpError) + '</div>';
    var partial = partialTaskCount(prev);
    var partialHTML = partial > 0
      ? '<span class="ship-bar-partial" title="' + escapeHtml(tCount('ship.partialNote', partial)) + '">' +
        escapeHtml(tCount('ship.partial', partial)) + '</span>'
      : '';
    // Pas de compteurs de tâches ici : les filtres de la liste juste en dessous
    // les portent déjà (« Toutes 7 · À relire 0 · Traitées 7 »). Pas de puce de
    // commits non plus : elle est déjà dans l'en-tête, à côté du bouton (voir
    // buildShipButtonHTML) ; la répéter ici ferait doublon.
    var subHTML = sub ? '<span class="ship-bar-sub">' + escapeHtml(sub) + '</span>' : '';
    var behindHTML = buildCardBehindHTML(prev);
    var linksHTML = prev ? buildShipLinksHTML(prev) : '';
    var hasContent = !!(subHTML || partialHTML || warningsHTML || behindHTML || linksHTML);
    var hasWarning = !!(warningsHTML);
    var cls = 'ship-bar' + (hasContent ? (hasWarning ? ' ship-bar-warn' : '') : ' ship-bar-empty');
    return '<div class="' + cls + '">' +
        '<div class="ship-bar-state">' +
          subHTML +
          partialHTML +
          warningsHTML +
          behindHTML +
          linksHTML +
        '</div>' +
      '</div>';
  }

  // Retard du chantier lui-même sur sa base (la release) : une ligne par dépôt
  // concerné, seulement quand il y en a. Ce retard ne bloque pas la livraison,
  // il annonce le rebase à faire pour livrer proprement.
  function buildCardBehindHTML(prev) {
    if (!prev || !prev.repos) return '';
    return prev.repos.filter(function (r) { return r.behind > 0; }).map(function (r) {
      var label = tCount('behind.cardLabel', r.behind, { base: r.base || '' });
      var tip = tCount('behind.cardTooltip', r.behind, { base: r.base || '' });
      return '<span class="ship-bar-behind" title="' + escapeHtml(tip) + '">' + escapeHtml(label) + '</span>';
    }).join('');
  }

  // Rafraîchit la barre de livraison et le bouton Ship de l'en-tête (pas la
  // liste, pas le détail).
  function refreshShipBar(cardId) {
    if (state.cardId !== cardId) return false;
    var card = state.cardsById[cardId];
    if (!card) return false;
    var bar = document.querySelector('.ship-bar');
    if (bar) {
      var barWrapper = document.createElement('div');
      barWrapper.innerHTML = buildShipBarHTML(card);
      var freshBar = barWrapper.firstElementChild;
      if (freshBar) bar.replaceWith(freshBar);
    }
    // Emplacement, pas bouton : selon l'état du chantier on y met le bouton de
    // livraison ou l'ancre « déjà arrivé ».
    var slot = document.querySelector('.ship-slot');
    if (slot) {
      var slotWrapper = document.createElement('div');
      slotWrapper.innerHTML = buildShipButtonHTML(card);
      var freshSlot = slotWrapper.firstElementChild;
      if (freshSlot) slot.replaceWith(freshSlot);
    }
    return !!(bar || slot);
  }

  function buildShipRepoRowHTML(prev, r) {
    var meta;
    if (r.pending > 0) {
      // Les commits annoncés sont ceux qui restent à livrer, pas le total du
      // chantier : un chantier déjà livré puis repris ne compte que le neuf.
      meta = tCount('ship.commits', r.pending) + ' · ' + tCount('ship.files', r.files);
    } else {
      meta = r.shippedAt ? t('ship.repoShipped') : t('ship.repoNothing');
    }
    var target = deliveryModeMerges(prev.mode) ? (prev.target || r.base) : r.base;
    return '<div class="ship-repo">' +
      '<div class="ship-repo-head"><span class="repo-chip">' + escapeHtml(r.repoName) + '</span>' +
      '<span class="mono ship-repo-branch">' + escapeHtml(r.branch) + ' → ' + escapeHtml(target) + '</span></div>' +
      '<div class="ship-repo-meta">' + escapeHtml(meta) + '</div>' +
      '</div>';
  }

  function buildShipResultsHTML(cardId) {
    var result = state.shipResultByCard[cardId];
    if (!result) return '';
    // La cible réelle d'une fusion vient du réglage du projet ; la réponse de
    // livraison ne porte que la base de la branche de chantier.
    var target = (state.deliveryByCard[cardId] || {}).target || '';
    var rows = (result.repos || []).map(function (r) {
      var label;
      if (r.error) label = '✕ ' + r.error;
      else if (r.skipped) label = '· ' + t('ship.repoNothing');
      else if (r.merged) label = '✓ ' + t(r.pushed ? 'ship.mergedPushed' : 'ship.mergedInto', { base: target || r.base });
      else label = '✓ ' + (r.prUrl || r.branch);
      return '<div class="ship-result-row ' + (r.error ? 'ship-result-error' : '') + '">' +
        '<span class="repo-chip">' + escapeHtml(r.repoName) + '</span> ' + escapeHtml(label) + '</div>';
    }).join('');
    return '<div class="ship-results">' + rows + '</div>';
  }

  function buildShipModalHTML(card, prev) {
    var reposHTML = (prev.repos || []).map(function (r) { return buildShipRepoRowHTML(prev, r); }).join('');
    var warnings = (prev.warnings || []).map(function (w) {
      return '<div class="modal-note modal-note-warning">⚠ ' + escapeHtml(deliveryWarningText(w)) + '</div>';
    }).join('');
    // La règle du mode choisi est répétée dans le récapitulatif : ce panneau
    // est la confirmation, il doit dire ce qui va sortir de la machine.
    var modeNote = deliveryModeNote(prev.mode);
    var mergeNote = modeNote ? '<div class="modal-note">' + escapeHtml(modeNote) + '</div>' : '';
    // Livraison partielle : le récapitulatif étant la confirmation, c'est ici
    // qu'il faut dire que tout le chantier ne partira pas, et que ce n'est pas
    // une impasse (le reste se livrera après acceptation).
    var partial = partialTaskCount(prev);
    var partialNote = partial > 0
      ? '<div class="modal-note">' + escapeHtml(tCount('ship.partialNote', partial)) + '</div>'
      : '';
    return '<div class="modal modal-sm">' +
      '<div class="modal-head"><span class="modal-title">' + escapeHtml(t('ship.modalTitle')) + '</span>' +
      '<button class="icon-btn" data-action="close-modal" aria-label="' + escapeHtml(t('common.close')) + '">✕</button></div>' +
      '<div class="ship-modal-sub">' + escapeHtml(deliveryActionLabel(prev)) + '</div>' +
      reposHTML + partialNote + mergeNote + warnings + buildShipSessionsWarningHTML(card) +
      buildShipResultsHTML(card.id) +
      '<div id="ship-modal-error" class="modal-error hidden"></div>' +
      '<div class="modal-foot"><button class="btn-outline" data-action="close-modal">' + escapeHtml(t('common.cancel')) + '</button>' +
      '<button class="btn-green" data-action="submit-ship" data-card-id="' + card.id + '">' +
      shipIconHTML() + escapeHtml(t('ship.modalConfirm')) + '</button></div>' +
      '</div>';
  }

  // Le bouton d'action de cette modale EST la confirmation : deux clics au
  // total pour livrer, le second en connaissance de cause (branche, base,
  // dépôts concernés, avertissements).
  // Une session Claude Code ouverte à la main dans un dépôt du projet peut
  // écrire là où la livraison fusionne : on le dit, sans bloquer.
  function buildShipSessionsWarningHTML(card) {
    var active = claudeSessionList(card.projectId).filter(function (x) { return x.active; });
    if (!active.length) return '';
    return '<div class="modal-note modal-note-warning">' + escapeHtml(t('sessions.shipWarning', { repo: active[0].repoName, title: active[0].title })) + '</div>';
  }

  function openShipModal(cardId) {
    var card = state.cardsById[cardId];
    var prev = state.deliveryByCard[cardId];
    if (!card || !prev || !card.shipReady || !deliverableRepos(prev).length) return;
    delete state.shipResultByCard[cardId];
    fetchClaudeSessions(card.projectId, true);
    openModal(buildShipModalHTML(card, prev));
  }

  // markShipSailing : le temps de la livraison, le bouton de la modale ne se
  // reclique pas (pousser deux fois n'a rien de gratuit) et dit qu'il travaille,
  // son bateau avançant au rythme d'une houle. Un push peut prendre plusieurs
  // secondes : sans ça, le clic paraît sans effet.
  function markShipSailing() {
    var btn = document.querySelector('[data-action="submit-ship"]');
    if (!btn) return;
    btn.disabled = true;
    btn.classList.add('btn-sailing');
    btn.innerHTML = shipIconHTML() + escapeHtml(t('ship.shipping'));
  }

  // La modale survit à un échec (elle porte le message d'erreur) : le bouton
  // doit y redevenir cliquable, sinon la seule sortie serait de tout refermer.
  function restoreShipButton() {
    var btn = document.querySelector('[data-action="submit-ship"]');
    if (!btn) return;
    btn.disabled = false;
    btn.classList.remove('btn-sailing');
    btn.innerHTML = shipIconHTML() + escapeHtml(t('ship.modalConfirm'));
  }

  function submitShip(cardId) {
    var errEl = document.getElementById('ship-modal-error');
    markShipSailing();
    api('/api/cards/' + cardId + '/ship', { method: 'POST', body: { confirm: true } }).then(function (res) {
      if (res && res.card) upsertCard(res.card);
      state.shipResultByCard[cardId] = res;
      var failed = (res.repos || []).filter(function (r) { return r.error; });
      if (!failed.length) armWake('ship', cardId);
      loadDelivery(cardId);
      if (failed.length) {
        var card = state.cardsById[cardId];
        var prev = state.deliveryByCard[cardId];
        if (card && prev) openModal(buildShipModalHTML(card, prev));
        return;
      }
      closeModal();
      renderMain();
    }).catch(function (e) {
      restoreShipButton();
      if (!errEl) return;
      errEl.textContent = (e instanceof ApiError && e.message) || t('ship.errorFailed');
      errEl.classList.remove('hidden');
    });
  }

  // Rattrapage de la destination : le geste mécanique d'abord (fusion de la
  // destination dans la branche du chantier, côté serveur). Il passe tout seul
  // dans le cas courant ; quand il conflicte, rien n'est modifié et c'est là
  // qu'un agent est utile, pas avant.
  function catchUpCard(cardId) {
    delete state.catchUpErrorByCard[cardId];
    api('/api/cards/' + cardId + '/catch-up', { method: 'POST', body: {} }).then(function (res) {
      if (res && res.card) upsertCard(res.card);
      var conflicted = (res.repos || []).filter(function (r) { return r.conflictFilePaths; });
      var failed = (res.repos || []).filter(function (r) { return r.error && !r.conflictFilePaths; });
      if (failed.length) state.catchUpErrorByCard[cardId] = failed[0].error;
      // Succès : rien à annoncer, l'écran le dit déjà (le bouton disparaît, le
      // retard tombe, la livraison s'active). loadDelivery redessine la barre.
      loadDelivery(cardId);
      // Chaque dépôt a sa propre cible et ses propres fichiers en conflit : un
      // chantier multi-dépôts peut buter sur plusieurs à la fois, chacun réglé
      // par sa propre tâche (une tâche = un worktree = un dépôt).
      if (conflicted.length) openModal(buildCatchUpConflictHTML(cardId, conflicted));
    }).catch(function (e) {
      state.catchUpErrorByCard[cardId] = (e instanceof ApiError && e.message) || t('catchUp.errorFailed');
      refreshShipBar(cardId);
    });
  }

  // Conflit de rattrapage : dire quels fichiers, rappeler que rien n'a bougé, et
  // proposer le seul chemin qui reste (une tâche d'agent par dépôt en conflit).
  function buildCatchUpConflictHTML(cardId, conflicted) {
    var multi = conflicted.length > 1;
    var title = multi ? t('catchUp.conflictTitleMulti') : t('catchUp.conflictTitle', { base: conflicted[0].target });
    var itemsHTML = conflicted.map(function (r) {
      var files = r.conflictFilePaths.split(' ').join(', ');
      var repoLabel = multi ? '<div class="catchup-conflict-repo mono">' + escapeHtml(r.repoName) + '</div>' : '';
      return '<div class="catchup-conflict-item">' + repoLabel +
        '<div class="modal-note">' + escapeHtml(t('catchUp.conflictBody', { base: r.target, files: files })) + '</div>' +
        '<button class="btn-green" data-action="catch-up-ask-agent" data-card-id="' + cardId + '"' +
        ' data-repo-name="' + escapeHtml(r.repoName) + '" data-target="' + escapeHtml(r.target) + '" data-files="' + escapeHtml(files) + '">' +
        escapeHtml(t('catchUp.askAgent')) + (multi ? ' (' + escapeHtml(r.repoName) + ')' : '') + '</button>' +
        '</div>';
    }).join('');
    return '<div class="modal modal-sm">' +
      '<div class="modal-head"><span class="modal-title">' + escapeHtml(title) + '</span>' +
      '<button class="icon-btn" data-action="close-modal" aria-label="' + escapeHtml(t('common.close')) + '">✕</button></div>' +
      itemsHTML +
      '<div class="modal-foot"><button class="btn-outline" data-action="close-modal">' + escapeHtml(t('common.cancel')) + '</button></div>' +
      '</div>';
  }

  // Confier le conflit d'un dépôt à un agent : la modale de création de tâche,
  // pré-remplie et pré-choisie sur ce dépôt (une tâche ne travaille que dans un
  // seul worktree). L'humain garde la main sur l'agent choisi et sur le texte
  // avant d'envoyer.
  function askAgentToCatchUp(cardId, repoName, target, files) {
    openNewTaskModal(cardId);
    setTimeout(function () {
      var titleEl = document.getElementById('new-task-title');
      var promptEl = document.getElementById('new-task-prompt');
      var repoEl = document.getElementById('new-task-repo');
      if (titleEl) titleEl.value = t('catchUp.taskTitle', { base: target });
      if (promptEl) promptEl.value = t('catchUp.taskPrompt', { base: target, files: files });
      if (repoEl && repoName) repoEl.value = repoName;
      if (promptEl) promptEl.focus();
    }, 0);
  }

  // ---------------------------------------------------------------------
  // Recette manuelle (panneau, runs, journal)
  // ---------------------------------------------------------------------

  // Une recette est la commande d'un dépôt (Repo.previewCmd) lancée par Sillage
  // dans le worktree d'un chantier ou d'une tâche. Le panneau montre une ligne
  // par dépôt, le journal en direct, et toujours le chemin du worktree : sans
  // commande configurée, ce chemin est le seul repli, et il suffit.

  function upsertPreview(run) {
    if (!run || !run.id) return;
    var i = state.previews.findIndex(function (x) { return x.id === run.id; });
    if (i >= 0) state.previews[i] = run; else state.previews.push(run);
    // Un seul run par worktree côté serveur : le frontend applique la même
    // règle, sinon un ancien run resterait affiché à côté de son remplaçant.
    state.previews = state.previews.filter(function (x) {
      return x.id === run.id || x.dir !== run.dir;
    });
  }

  function previewRunning() {
    return state.previews.filter(function (r) { return r.status === 'running'; });
  }

  // Le run d'un worktree donné : c'est le worktree qui identifie une recette,
  // pas le chantier (une tâche du même chantier a le sien).
  function previewRunForDir(dir) {
    if (!dir) return null;
    var found = null;
    state.previews.forEach(function (r) { if (r.dir === dir) found = r; });
    return found;
  }

  function previewStatusLabel(run) {
    if (!run) return t('preview.statusIdle');
    if (run.status === 'running') return t('preview.statusRunning');
    if (run.status === 'stopped') return t('preview.statusStopped');
    if (run.status === 'failed') return t('preview.statusFailed');
    return t('preview.statusExited', { code: run.exitCode });
  }

  // Les cibles recettables du panneau : une par dépôt du chantier, ou la seule
  // du worktree d'une tâche. Chacune porte sa commande (celle de son dépôt) et
  // le run en cours s'il y en a un.
  function previewTargets(scope) {
    if (!scope) return [];
    if (scope.kind === 'task') {
      var task = state.tasksById[scope.id];
      if (!task) return [];
      var project = state.projectsById[task.projectId] || {};
      var repo = (project.repos || []).filter(function (r) { return r.name === task.repoName; })[0] || {};
      return [{
        repoName: task.repoName, dir: task.worktreeDir, cmd: repo.previewCmd || '',
        run: previewRunForDir(task.worktreeDir), taskId: task.id
      }];
    }
    if (scope.kind === 'card') {
      var card = state.cardsById[scope.id];
      if (!card) return [];
      var proj = state.projectsById[card.projectId] || {};
      return (card.branches || []).map(function (b) {
        var r = (proj.repos || []).filter(function (x) { return x.name === b.repoName; })[0] || {};
        return {
          repoName: b.repoName, dir: b.worktreeDir, cmd: r.previewCmd || '',
          run: previewRunForDir(b.worktreeDir), cardId: card.id
        };
      });
    }
    // Portée « tout » (compteur de la sidebar) : les runs en cours, sans dépôt
    // sans run, puisqu'on vient y chercher ce qui tourne.
    return previewRunning().map(function (run) {
      return { repoName: run.repoName, dir: run.dir, cmd: run.cmd, run: run, cardId: run.cardId, taskId: run.taskId };
    });
  }

  // Bouton de l'en-tête de chantier. Une pastille apparaît quand une recette de
  // ce chantier tourne : c'est l'information qu'on cherche des yeux en revenant
  // sur l'écran.
  // Le libellé nomme la portée : les deux boutons de recette (chantier et tâche)
  // sont visibles en même temps dès qu'une tâche est ouverte, et deux boutons
  // « Recette » côte à côte n'apprennent pas lequel lance quoi.
  function buildPreviewButtonHTML(card) {
    var running = previewRunning().some(function (r) { return r.cardId === card.id; });
    return '<button class="btn-outline preview-btn' + (running ? ' preview-btn-on' : '') + '"' +
      ' data-action="open-card-preview" data-card-id="' + card.id + '"' +
      ' title="' + escapeHtml(t('preview.tooltip')) + '">' +
      (running ? '<span class="preview-dot preview-dot-on"></span>' : '') +
      escapeHtml(t('preview.buttonCard')) + '</button>';
  }

  // Bouton du panneau de détail : la recette d'une tâche tourne dans le worktree
  // de la tâche, donc sur son incrément, avant l'acceptation.
  function buildTaskPreviewButtonHTML(task) {
    if (!task.worktreeDir) return '';
    var run = previewRunForDir(task.worktreeDir);
    var running = !!run && run.status === 'running';
    // Secondaire, à largeur fixe, et du même poids que Refuser : ce sont les deux
    // options qui accompagnent Accepter. Un aplat gris se lisait plus fort que
    // Refuser, alors que refuser engage bien davantage.
    return '<button class="btn-outline preview-task-btn' + (running ? ' preview-btn-on' : '') + '"' +
      ' data-action="open-task-preview" data-task-id="' + task.id + '"' +
      ' title="' + escapeHtml(t('preview.taskTooltip')) + '">' +
      (running ? '<span class="preview-dot preview-dot-on"></span>' : '') +
      escapeHtml(t('preview.button')) + '</button>';
  }

  function buildPreviewRowHTML(target) {
    var run = target.run;
    var running = !!run && run.status === 'running';
    var dotClass = running ? 'preview-dot-on' : (run && run.status === 'failed' ? 'preview-dot-fail' : '');
    var startAction = target.taskId
      ? 'data-action="start-task-preview" data-task-id="' + target.taskId + '"'
      : 'data-action="start-card-preview" data-card-id="' + target.cardId + '" data-repo-name="' + escapeHtml(target.repoName) + '"';

    var right;
    if (!target.cmd) {
      right = '<span class="preview-none">' + escapeHtml(t('preview.noCmd')) + '</span>';
    } else if (running) {
      right = '<button class="btn-outline btn-sm" data-action="stop-preview" data-run-id="' + run.id + '">' +
        escapeHtml(t('preview.stop')) + '</button>';
    } else {
      right = '<button class="btn-outline btn-sm" ' + startAction + '>' +
        escapeHtml(run ? t('preview.restart') : t('preview.start')) + '</button>';
    }

    var middle = '<span class="preview-status">' + escapeHtml(previewStatusLabel(run)) + '</span>';
    if (running && run.url) {
      middle = '<a class="preview-url mono" href="' + escapeHtml(run.url) + '" target="_blank" rel="noopener">' +
        escapeHtml(run.url) + '</a>';
    }
    var logBtn = run
      ? '<button class="detail-link" data-action="show-preview-log" data-run-id="' + run.id + '">' + escapeHtml(t('preview.showLog')) + '</button>'
      : '';

    return '<div class="preview-row">' +
      '<span class="preview-dot ' + dotClass + '"></span>' +
      '<span class="repo-chip">' + escapeHtml(target.repoName || '') + '</span>' +
      middle + logBtn +
      '<span class="preview-row-actions">' + right + '</span>' +
      '</div>' +
      buildPreviewWorktreeHTML(target.dir);
  }

  // Le chemin du worktree, toujours affiché : c'est le repli universel de la
  // recette, celui qui marche sans aucune configuration.
  function buildPreviewWorktreeHTML(dir) {
    if (!dir) return '';
    return '<div class="preview-worktree">' +
      '<span class="preview-worktree-label">' + escapeHtml(t('preview.worktree')) + '</span>' +
      '<code class="mono preview-worktree-path">' + escapeHtml(dir) + '</code>' +
      '<button class="detail-link" data-action="copy-path" data-path="' + escapeHtml(dir) + '">' +
      escapeHtml(t('preview.copyPath')) + '</button>' +
      '</div>';
  }

  function previewScopeTitle(scope) {
    if (scope.kind === 'task') {
      var task = state.tasksById[scope.id];
      return t('preview.taskTitle', { ref: task ? task.ref : '' });
    }
    if (scope.kind === 'card') {
      var card = state.cardsById[scope.id];
      return t('preview.cardTitle', { title: card ? card.title : '' });
    }
    return t('preview.allTitle');
  }

  function buildPreviewLogHTML(scope) {
    var lines = (scope.runId && state.previewLogByRun[scope.runId]) || [];
    var body = lines.length
      ? lines.map(function (l) { return escapeHtml(l); }).join('\n')
      : '<span class="preview-log-empty">' + escapeHtml(t('preview.logEmpty')) + '</span>';
    return '<pre class="preview-log" id="preview-log">' + body + '</pre>';
  }

  function buildPreviewModalHTML(scope) {
    var targets = previewTargets(scope);
    var hint = '';
    if (scope.kind === 'task') {
      var task = state.tasksById[scope.id];
      hint = '<div class="modal-note">' + escapeHtml(t('preview.taskScope', { ref: task ? task.ref : '' })) + '</div>';
    } else if (scope.kind === 'card' && targets.length === 0) {
      hint = '<div class="modal-note">' + escapeHtml(t('preview.noBranch')) + '</div>';
    }
    var missingCmd = targets.some(function (x) { return !x.cmd; });
    var cmdHint = missingCmd
      ? '<div class="modal-note">' + escapeHtml(t('preview.noCmdHint')) + ' ' +
        '<button class="detail-link" data-action="open-preview-settings">' + escapeHtml(t('preview.openSettings')) + '</button></div>'
      : '';

    return '<div class="modal modal-lg">' +
      '<div class="modal-head"><span class="modal-title">' + escapeHtml(previewScopeTitle(scope)) + '</span>' +
      '<button class="icon-btn" data-action="close-modal" aria-label="' + escapeHtml(t('common.close')) + '">✕</button></div>' +
      hint +
      '<div class="preview-rows">' + targets.map(buildPreviewRowHTML).join('') + '</div>' +
      cmdHint +
      '<div id="preview-modal-error" class="modal-error hidden"></div>' +
      buildPreviewLogHTML(scope) +
      '<div class="modal-foot"><button class="btn-outline" data-action="close-modal">' + escapeHtml(t('common.close')) + '</button></div>' +
      '</div>';
  }

  function openPreviewModal(kind, id) {
    var scope = { kind: kind, id: id, runId: null };
    var targets = previewTargets(scope);
    // Journal ouvert d'office sur le run en cours : c'est ce qu'on vient voir.
    var withRun = targets.filter(function (x) { return x.run; })[0];
    if (withRun) scope.runId = withRun.run.id;
    openModal(buildPreviewModalHTML(scope));
    state.previewScope = scope; // après openModal, qui remet la portée à zéro
    if (scope.runId) loadPreviewLog(scope.runId);
  }

  // Lien « Ouvrir les réglages du projet » du panneau de recette, quand un dépôt
  // n'a pas de commande : on retrouve le projet depuis la portée du panneau
  // (un chantier ou une tâche, jamais la portée « all ») et on ouvre directement
  // l'onglet Dépôts, où la commande se déclare.
  function openPreviewSettings() {
    var scope = state.previewScope;
    if (!scope) return;
    var projectId = null;
    if (scope.kind === 'card') {
      var card = state.cardsById[scope.id];
      projectId = card && card.projectId;
    } else if (scope.kind === 'task') {
      var task = state.tasksById[scope.id];
      projectId = task && task.projectId;
    }
    if (!projectId) return;
    state.projectId = projectId;
    openEditProjectModal();
    setProjectTab('repos');
  }

  // Rafraîchit le panneau sans le rouvrir : les lignes et leurs boutons, pas le
  // journal (qui reçoit ses lignes une par une, voir appendPreviewLogLine).
  function refreshPreviewModal() {
    var scope = state.previewScope;
    if (!scope || !state.modal) return;
    var rows = document.querySelector('.preview-rows');
    if (!rows) return;
    rows.innerHTML = previewTargets(scope).map(buildPreviewRowHTML).join('');
  }

  function previewError(message) {
    var el = document.getElementById('preview-modal-error');
    if (!el) return;
    el.textContent = message;
    el.classList.remove('hidden');
  }

  function startCardPreview(cardId, repoName) {
    api('/api/cards/' + cardId + '/preview', { method: 'POST', body: { repoName: repoName || '' } })
      .then(onPreviewStarted)
      .catch(function (e) { previewError((e instanceof ApiError && e.message) || t('preview.errorStartFailed')); });
  }

  function startTaskPreview(taskId) {
    api('/api/tasks/' + taskId + '/preview', { method: 'POST', body: {} })
      .then(onPreviewStarted)
      .catch(function (e) { previewError((e instanceof ApiError && e.message) || t('preview.errorStartFailed')); });
  }

  function onPreviewStarted(run) {
    if (!run) return;
    upsertPreview(run);
    // Un nouveau run repart d'un journal vide : celui du run précédent ne
    // raconte plus ce qui se passe.
    state.previewLogByRun[run.id] = [];
    if (state.previewScope) state.previewScope.runId = run.id;
    refreshPreviewModal();
    renderPreviewLog();
    renderSidebar();
  }

  function stopPreview(runId) {
    api('/api/previews/' + runId + '/stop', { method: 'POST', body: {} })
      .catch(function (e) { previewError((e instanceof ApiError && e.message) || t('preview.errorStopFailed')); });
  }

  function showPreviewLog(runId) {
    if (!state.previewScope) return;
    state.previewScope.runId = runId;
    refreshPreviewModal();
    renderPreviewLog();
    loadPreviewLog(runId);
  }

  function loadPreviewLog(runId) {
    api('/api/previews/' + runId + '/log').then(function (data) {
      if (!data) return;
      state.previewLogByRun[runId] = data.lines || [];
      if (state.previewScope && state.previewScope.runId === runId) renderPreviewLog();
    }).catch(function () {});
  }

  function renderPreviewLog() {
    var scope = state.previewScope;
    var el = document.getElementById('preview-log');
    if (!scope || !el) return;
    var wrapper = document.createElement('div');
    wrapper.innerHTML = buildPreviewLogHTML(scope);
    var fresh = wrapper.firstElementChild;
    if (fresh) {
      el.replaceWith(fresh);
      fresh.scrollTop = fresh.scrollHeight;
    }
  }

  // Une ligne de journal arrive : append pur, jamais de reconstruction. Une
  // installation de dépendances en produit des milliers.
  function appendPreviewLogLine(runId, line) {
    var buffer = state.previewLogByRun[runId] || (state.previewLogByRun[runId] = []);
    buffer.push(line);
    if (buffer.length > 2000) buffer.splice(0, buffer.length - 2000);
    if (!state.previewScope || state.previewScope.runId !== runId) return;
    var el = document.getElementById('preview-log');
    if (!el) return;
    var atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 40;
    var empty = el.querySelector('.preview-log-empty');
    if (empty) el.textContent = '';
    el.appendChild(document.createTextNode((el.textContent ? '\n' : '') + line));
    if (atBottom) el.scrollTop = el.scrollHeight;
  }

  function writeClipboard(text, done) {
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(text).then(done).catch(function () {});
      return;
    }
    // Repli pour les contextes non sécurisés (http sur une IP de réseau local),
    // où l'API clipboard n'existe pas.
    var ta = document.createElement('textarea');
    ta.value = text;
    ta.setAttribute('readonly', 'readonly');
    ta.style.position = 'fixed';
    ta.style.opacity = '0';
    document.body.appendChild(ta);
    ta.select();
    try { document.execCommand('copy'); done(); } catch (e) {}
    document.body.removeChild(ta);
  }

  function copyPathToClipboard(path, el) {
    writeClipboard(path, function () {
      if (!el) return;
      var before = el.textContent;
      el.textContent = t('preview.copied');
      setTimeout(function () { el.textContent = before; }, 1200);
    });
  }

  // Bouton de copie en icône : la coche remplace l'icône le temps d'un battement,
  // seule confirmation possible sur un bouton de 20 px. Le libellé accessible
  // suit, sinon un lecteur d'écran annoncerait encore « copier ».
  function copyToClipboardIcon(text, el) {
    if (!el || el.classList.contains('copied')) return;
    writeClipboard(text, function () {
      var beforeTitle = el.getAttribute('title') || '';
      el.classList.add('copied');
      el.innerHTML = copiedIconHTML();
      el.setAttribute('title', t('chat.copied'));
      el.setAttribute('aria-label', t('chat.copied'));
      setTimeout(function () {
        el.classList.remove('copied');
        el.innerHTML = copyIconHTML();
        el.setAttribute('title', beforeTitle);
        el.setAttribute('aria-label', beforeTitle);
      }, 1300);
    });
  }

  // Le code copié se relit dans le DOM plutôt que dans un attribut : le bloc est
  // déjà là, le dupliquer doublerait le poids du fil.
  function copyCodeBlock(el) {
    var wrap = el.closest('.code-wrap');
    var code = wrap ? wrap.querySelector('code') : null;
    if (code) copyToClipboardIcon(code.textContent, el);
  }

  // Contexte de liste courant (carte ou boîte de réception), utilisé à la fois
  // par le rendu complet et par le rafraîchissement ciblé sur événement SSE.
  function currentTaskListContext() {
    if (state.cardId) {
      return {
        tasksAll: state.tasks.filter(function (t) { return t.cardId === state.cardId; }),
        opts: {
          showProject: false,
          emptyMsg: t('work.emptyCard'),
          emptyCta: { action: 'open-new-task', label: t('work.emptyCardAction') }
        }
      };
    }
    if (state.screen === 'inbox') {
      return {
        tasksAll: state.tasks.filter(function (t) { return t.unread || t.status === 'review'; }),
        opts: { showProject: true, emptyMsg: t('inbox.empty'), bodyClass: 'inbox-body', paneTopHTML: buildInboxHelloHTML() + buildInboxSignalsHTML() + buildAwaitingShipBannerHTML() }
      };
    }
    return null;
  }

  // Chantiers dont tout le travail est terminal mais qui n'ont pas été livrés
  // (voir Card.awaitingShip) : sans ce bandeau, rien dans la boîte de réception
  // ne les signale, puisque leurs tâches sont acceptées/refusées et n'y
  // apparaissent donc jamais.
  // Une seule porte pour « qu'est-ce qui m'attend » : les signaux prêts de la
  // Vigie remontent ici, une ligne chacun, avec le jugement en tête. Le détail
  // (commentaires, actions) reste sur l'écran Vigie.
  // L'accueil salue, puis dit en une phrase ce qui vous attend. Pas de
  // tableau de bord : une phrase se lit d'un coup d'œil, des chiffres se
  // déchiffrent.
  function buildInboxHelloHTML() {
    var h = new Date().getHours();
    var name = (state.settings && state.settings.displayName) || '';
    var hello = t(h >= 18 || h < 5 ? 'hello.evening' : 'hello.day') + (name ? ', ' + name : '');
    var review = state.tasks.filter(function (tk) { return tk.status === 'review'; }).length;
    var signals = vigieSignals().filter(function (sg) { return sg.status === 'ready' && sg.priority !== 'low'; }).length;
    var ship = state.cards.filter(function (c) { return c.awaitingShip; }).length;
    var running = state.tasks.filter(function (tk) { return tk.status === 'running'; }).length +
      state.signals.filter(function (sg) { return sg.status === 'analyzing'; }).length;
    var parts = [];
    if (review) parts.push(tCount('hello.review', review));
    if (signals) parts.push(tCount('hello.signals', signals));
    if (ship) parts.push(tCount('hello.ship', ship));
    var line;
    if (parts.length) {
      var list = parts.length > 1 ? parts.slice(0, -1).join(', ') + ' ' + t('hello.and') + ' ' + parts[parts.length - 1] : parts[0];
      line = t('hello.waiting', { list: list });
    } else {
      line = t('hello.nothing');
    }
    if (running) line += ' ' + tCount('hello.running', running);
    return '<div class="inbox-hello"><h1>' + escapeHtml(hello) + '</h1><p>' + escapeHtml(line) + '</p></div>';
  }

  function buildInboxSignalsHTML() {
    var signals = vigieSignals().filter(function (s) { return s.status === 'ready' && s.priority !== 'low'; });
    if (!signals.length) return '';
    var rows = signals.map(function (s) {
      return '<button class="inbox-signal-row" data-action="open-signal" data-signal-id="' + s.id + '">' +
        (s.decision ? '<span class="signal-decision signal-decision-' + escapeHtml(s.decision) + '">' + escapeHtml(t('vigie.decision.' + s.decision)) + '</span>' : '') +
        '<span class="inbox-signal-title">' + escapeHtml(s.title) + '</span>' +
        '<span class="inbox-signal-ref mono">' + escapeHtml(signalRefText(s)) + '</span>' +
      '</button>';
    }).join('');
    return '<div class="awaiting-ship-banner inbox-signals">' +
      '<div class="awaiting-ship-head">' + escapeHtml(t('inbox.signals.title')) + '</div>' + rows + '</div>';
  }

  function buildAwaitingShipBannerHTML() {
    var cards = state.cards.filter(function (c) { return c.awaitingShip; });
    if (!cards.length) return '';
    var rows = cards.map(function (c) {
      var p = state.projectsById[c.projectId];
      return '<div class="awaiting-ship-row" data-action="open-card" data-card-id="' + c.id + '">' +
        shipIconHTML() +
        '<span class="awaiting-ship-title">' + escapeHtml(c.title) + '</span>' +
        (p ? '<span class="awaiting-ship-project">' + escapeHtml(p.name) + '</span>' : '') +
        '</div>';
    }).join('');
    return '<div class="awaiting-ship-banner">' +
      '<div class="awaiting-ship-head">' + escapeHtml(t('inbox.awaitingShip.title')) + '</div>' +
      rows +
      '</div>';
  }

  function buildWorkHTML() {
    var ctx = currentTaskListContext();
    return buildTaskListShell(ctx.tasksAll, ctx.opts);
  }

  function buildInboxHTML() {
    var ctx = currentTaskListContext();
    return buildTaskListShell(ctx.tasksAll, ctx.opts);
  }

  // Rafraîchit uniquement les pilules de filtre + la liste de tâches (pas le
  // panneau de détail, jamais le conteneur des messages) : utilisé par les
  // gestionnaires SSE task pour éviter un rendu complet.
  function refreshTaskListAndFilters() {
    var ctx = currentTaskListContext();
    if (!ctx) return false;
    var pane = document.querySelector('.task-list-pane');
    if (!pane) return false;
    pane.innerHTML = buildTaskListPaneInnerHTML(ctx.tasksAll, ctx.opts);
    return true;
  }

  // ---------------------------------------------------------------------
  // Rendu : panneau de détail de tâche
  // ---------------------------------------------------------------------

  // L'action principale d'une tâche à relire est l'acceptation (fusion locale
  // dans la branche du chantier) : sans confirmation, puisque rien ne sort de
  // la machine. La seule action sortante du produit est le Ship du chantier.
  function primaryActionInfo(task) {
    switch (task.status) {
      case 'waiting':
        return { label: t('action.startNow'), cls: 'btn-neutral', action: 'start-task', kind: 'plain' };
      case 'running':
        return { label: t('action.interrupt'), cls: 'btn-neutral', action: 'interrupt', kind: 'plain' };
      case 'review':
        return { label: t('action.accept'), cls: 'btn-green', action: 'accept-task', kind: 'plain' };
      case 'accepted':
      case 'cancelled':
        return { label: t('action.reopen'), cls: 'btn-neutral', action: 'reopen', kind: 'plain' };
      default:
        return { label: '', cls: 'btn-neutral', action: '', kind: 'plain' };
    }
  }

  // Bandeau d'état du panneau de détail. Un rebase automatique en cours prend le
  // dessus sur le statut : c'est l'information du moment, et elle explique
  // pourquoi la tâche n'accepte rien pendant ce temps.
  function buildStatusBadgeHTML(task) {
    if (task.rebasing) {
      return '<div class="status-badge status-badge-rebasing">' +
        '<span class="status-badge-icon"><span class="task-spinner"></span></span>' +
        '<span class="status-badge-label">' + escapeHtml(t('status.rebasing')) + '</span>' +
        '<span class="status-badge-note">' + escapeHtml(t('status.rebasingNote')) + '</span></div>';
    }
    var glyph = STATUS_GLYPH[task.status] || STATUS_GLYPH.running;
    var label = t('status.' + task.status);
    return '<div class="status-badge"><span class="status-badge-icon" style="color:' + glyph.color + '">' + glyph.icon + '</span>' +
      '<span class="status-badge-label">' + escapeHtml(label) + '</span></div>';
  }

  // Pastille du check de projet (ex. "go test ./...") : un rond avec coche ou
  // croix plutôt qu'un glyphe unicode, la commande elle-même n'apparaît qu'au
  // survol (attribut title) pour ne pas encombrer la barre d'actions.
  function checkIconHTML(ok) {
    var mark = ok
      ? '<path fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round" d="M4.8 8.3l2 2 4.4-4.4"/>'
      : '<path fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" d="M5.5 5.5l5 5M10.5 5.5l-5 5"/>';
    return '<svg class="check-svg" viewBox="0 0 16 16" width="14" height="14" aria-hidden="true" focusable="false">' +
      '<circle cx="8" cy="8" r="7" fill="none" stroke="currentColor" stroke-width="1.3"/>' +
      mark +
      '</svg>';
  }

  function renderChecks(checks) {
    if (!checks || checks.length === 0) return '';
    return checks.map(function (c) {
      return '<span class="check ' + (c.ok ? 'check-ok' : 'check-fail') + '" title="' + escapeHtml(c.label) + '">' + checkIconHTML(c.ok) + '</span>';
    }).join('');
  }

  // ---------------------------------------------------------------------
  // Réassignation d'une tâche à un autre agent
  // ---------------------------------------------------------------------

  function agentWarningText(warning) {
    if (!warning) return '';
    if (warning.indexOf('codex sandbox is blocked') !== -1) return t('agent.warning.codexSandbox');
    if (warning.indexOf('agy cannot work headlessly') !== -1) return t('agent.warning.agyPolicy');
    if (warning.indexOf('KIRO_API_KEY is not set') !== -1) return t('agent.warning.kiroApiKey');
    var m = /^(\S+) CLI not found in PATH$/.exec(warning);
    if (m) return t('agent.warning.cliNotFound', { cli: m[1] });
    return warning;
  }

  // Complète agentWarningText() : commande de contournement copiable et lien
  // vers la doc OpenAI. Uniquement pour la bannière d'avertissement (le title
  // d'un tooltip ne peut contenir ni HTML ni bouton).
  var CODEX_SANDBOX_FIX_CMD = 'sudo sysctl kernel.apparmor_restrict_unprivileged_userns=0';
  var AGENT_CLI_INSTALL = {
    claude: {
      command: 'curl -fsSL https://claude.ai/install.sh | bash',
      url: 'https://docs.anthropic.com/en/docs/claude-code/getting-started'
    },
    codex: {
      command: 'npm install -g @openai/codex',
      url: 'https://github.com/openai/codex'
    },
    copilot: {
      command: 'curl -fsSL https://gh.io/copilot-install | bash',
      url: 'https://docs.github.com/en/copilot/how-tos/copilot-cli/set-up-copilot-cli/install-copilot-cli'
    },
    agy: {
      command: 'curl -fsSL https://antigravity.google/cli/install.sh | bash',
      url: 'https://antigravity.google/docs/cli/install'
    },
    kiro: {
      command: 'curl -fsSL https://cli.kiro.dev/install | bash',
      url: 'https://kiro.dev/docs/cli/installation/'
    }
  };
  function agentWarningExtrasHTML(warning, agentId) {
    if (!warning) return '';
    var missing = /^(\S+) CLI not found in PATH$/.exec(warning);
    if (missing && AGENT_CLI_INSTALL[missing[1]]) {
      var install = AGENT_CLI_INSTALL[missing[1]];
      return '<div class="agent-warning-fallback">' + escapeHtml(t('agent.warning.installHint')) + '</div>' +
        '<div class="agent-warning-cmd">' +
          '<code class="mono">' + escapeHtml(install.command) + '</code>' +
          '<button data-action="copy-path" data-path="' + escapeHtml(install.command) + '">' + escapeHtml(t('agent.warning.copyCmd')) + '</button>' +
        '</div>' +
        '<a class="agent-warning-link" href="' + escapeHtml(install.url) + '" target="_blank" rel="noopener noreferrer">' +
          escapeHtml(t('agent.warning.installLink')) + '</a>';
    }
    if (warning.indexOf('agy cannot work headlessly') !== -1) {
      var fixKey = 'agent-fix:' + agentId;
      var fixLabel = isPendingConfirm(fixKey) ? t('agent.warning.agyPolicyFixConfirm') : t('agent.warning.agyPolicyFix');
      return '<div class="agent-warning-fallback">' + escapeHtml(t('agent.warning.agyPolicyHint')) + '</div>' +
        '<div class="agent-warning-action">' +
          '<button class="btn-outline" data-action="confirm-click" data-confirm-key="' + fixKey + '" data-confirm-action="agent-fix" data-confirm-id="' + escapeHtml(agentId) + '" data-default-label="' + escapeHtml(t('agent.warning.agyPolicyFix')) + '" data-confirm-label="' + escapeHtml(t('agent.warning.agyPolicyFixConfirm')) + '">' + escapeHtml(fixLabel) + '</button>' +
        '</div>' +
        '<a class="agent-warning-link" href="https://antigravity.google/docs/cli/reference" target="_blank" rel="noopener noreferrer">' +
          escapeHtml(t('agent.warning.agyPolicyLink')) + '</a>';
    }
    if (warning.indexOf('KIRO_API_KEY is not set') !== -1) {
      return '<div class="agent-warning-fallback">' + escapeHtml(t('agent.warning.kiroApiKeyHint')) + '</div>' +
        '<a class="agent-warning-link" href="https://kiro.dev/docs/cli/authentication/" target="_blank" rel="noopener noreferrer">' +
          escapeHtml(t('agent.warning.kiroApiKeyLink')) + '</a>';
    }
    if (warning.indexOf('codex sandbox is blocked') !== -1) {
      return '<div class="agent-warning-cmd">' +
        '<code class="mono">' + escapeHtml(CODEX_SANDBOX_FIX_CMD) + '</code>' +
        '<button data-action="copy-path" data-path="' + escapeHtml(CODEX_SANDBOX_FIX_CMD) + '">' + escapeHtml(t('agent.warning.copyCmd')) + '</button>' +
        '</div>' +
        '<div class="agent-warning-fallback">' + escapeHtml(t('agent.warning.codexSandboxFallback')) + '</div>' +
        '<a class="agent-warning-link" href="https://developers.openai.com/codex/concepts/sandboxing" target="_blank" rel="noopener noreferrer">' +
        escapeHtml(t('agent.warning.codexSandboxLink')) + '</a>';
    }
    return '';
  }

  function buildReassignMenuHTML(task) {
    var others = state.agents.filter(function (a) { return a.id !== task.agentId; });
    var items = others.map(function (a) {
      var warn = a.warning ? '<span class="agent-warning" title="' + escapeHtml(agentWarningText(a.warning)) + '">⚠</span>' : '';
      return '<button class="reassign-item" data-action="reassign-task" data-task-id="' + task.id + '" data-agent-id="' + a.id + '">' +
        '<span class="reassign-item-avatar" style="background:' + softColor(a.color) + '">' + a.emoji + '</span>' +
        '<span class="reassign-item-info"><span class="reassign-item-name">' + escapeHtml(a.name) + warn + '</span>' +
        '<span class="reassign-item-model mono">' + escapeHtml(a.model || '') + '</span></span>' +
        '</button>';
    }).join('');
    return '<div class="reassign-menu hidden" data-reassign-menu="' + task.id + '">' + items + '</div>';
  }

  function buildAgentChipHTML(task, agent, soft) {
    var others = state.agents.filter(function (a) { return a.id !== task.agentId; });
    var canReassign = task.status !== 'running' && others.length > 0;
    var innerHTML = '<span class="agent-avatar-sm" style="background:' + soft + '">' + agent.emoji + '</span>' + escapeHtml(agent.name);
    if (!canReassign) {
      return '<span class="agent-chip">' + innerHTML + '</span>';
    }
    return '<span class="agent-chip-wrap">' +
      '<button class="agent-chip agent-chip-btn" data-action="toggle-reassign-menu" data-task-id="' + task.id + '" title="' + escapeHtml(t('reassign.tooltip')) + '" aria-label="' + escapeHtml(t('reassign.tooltip')) + '">' + innerHTML + '</button>' +
      buildReassignMenuHTML(task) +
      '</span>';
  }

  function closeAllReassignMenus() {
    document.querySelectorAll('.reassign-menu').forEach(function (m) { m.classList.add('hidden'); });
  }
  function toggleReassignMenu(taskId) {
    var el = document.querySelector('.reassign-menu[data-reassign-menu="' + taskId + '"]');
    if (!el) return;
    var willOpen = el.classList.contains('hidden');
    closeAllReassignMenus();
    if (willOpen) el.classList.remove('hidden');
  }
  function doReassignTask(taskId, agentId) {
    closeAllReassignMenus();
    api('/api/tasks/' + taskId, { method: 'PATCH', body: { agentId: agentId } }).then(function (task) {
      upsertTask(task);
      renderMain();
    }).catch(function (e) {
      if (e instanceof ApiError) showDetailError(taskId, e.message || t('errors.reassignFailed'));
    });
  }

  function buildDetailHeadHTML(task) {
    var agent = state.agentsById[task.agentId] || { emoji: '?', name: '?', model: '?', color: '#ccc' };
    var soft = softColor(agent.color);
    var taskProject = state.projectsById[task.projectId];
    var multiRepo = !!(taskProject && taskProject.repos && taskProject.repos.length > 1);
    var action = primaryActionInfo(task);
    var tabs = ['chat', 'diff', 'files', 'history'];
    var tabLabels = { chat: t('tabs.conversation'), diff: t('tabs.diff'), files: t('tabs.deliverables'), history: t('tabs.history') };
    var tabCounts = { chat: task.messagesCount || 0, diff: task.filesCount || 0, files: (task.docsCount || 0) + (task.filesCount || 0), history: (task.commandLog || []).length };
    var tabDataAttr = { chat: 'conversation', diff: 'diff', files: 'deliverables', history: 'history' };

    var tabsHTML = tabs.map(function (tk) {
      var active = state.panelTab === tk;
      var isNew = tk === 'chat' && task.unread && !active;
      return '<button class="tab ' + (active ? 'tab-active' : '') + '" role="tab" data-tab="' + tabDataAttr[tk] + '" data-action="set-tab" data-panel-tab="' + tk + '">' +
        escapeHtml(tabLabels[tk]) + '<span class="tab-count">' + tabCounts[tk] + '</span>' +
        (isNew ? '<span class="tab-dot"></span>' : '') + '</button>';
    }).join('');

    // Acceptation en vol : contrairement à la ligne de liste, le bouton du
    // détail est seul sur sa ligne et peut porter le libellé de l'attente.
    var accepting = task.status === 'review' && !!state.acceptingByTask[task.id];
    var primaryBtnHTML = '<button id="task-primary-action" class="btn-action ' + action.cls + '"' +
      (accepting ? ' disabled' : '') + ' data-action="' + action.action + '" data-task-id="' + task.id + '">' +
      escapeHtml(accepting ? t('action.accepting') : action.label) + '</button>';

    // Refuser une tâche en revue, ou annuler une tâche en cours : même action
    // côté API (/cancel), seul le libellé change selon le contexte.
    // Le retard se règle par l'agent, pas par le serveur : le bouton envoie un
    // message dans le fil (mis en file si l'agent tourne encore), lui seul sait
    // résoudre un conflit de rebase.
    var behind = taskBehindCount(task);
    var behindRow = '';
    if (behind > 0) {
      behindRow = '<div class="behind-row">' +
        '<span class="behind-badge" title="' + escapeHtml(tCount('behind.taskTooltip', behind, { base: task.base || '' })) + '">↓' + behind + '</span>' +
        '<span class="behind-text">' + escapeHtml(tCount('behind.taskTooltip', behind, { base: task.base || '' })) + '</span>' +
        '<button class="detail-link" data-action="ask-rebase" data-task-id="' + task.id + '" title="' + escapeHtml(t('behind.askRebaseTooltip', { base: task.base || '' })) + '">' + escapeHtml(t('behind.askRebase')) + '</button>' +
        '</div>';
    }

    // Tâche qui n'a pas encore démarré : dit clairement quoi attend, et si la
    // tâche attendue n'existe plus (supprimée), le dit aussi. « Démarrer
    // maintenant » (bouton primaire) reste toujours disponible pour débloquer.
    var waitingRow = '';
    if (task.status === 'waiting') {
      var waitDep = state.tasksById[task.waitsForTaskId];
      var waitText = waitDep ? t('waitingFor.note', { title: waitDep.title }) : t('waitingFor.unknownTask');
      waitingRow = '<div class="behind-row"><span class="behind-text">' + escapeHtml(waitText) + '</span></div>';
    }

    // Refuser (ou annuler) est l'autre issue de la revue : un vrai bouton à côté
    // de l'action principale, plus un lien perdu sous elle. Même action côté API
    // (/cancel), seul le libellé change selon le contexte.
    var secondaryBtnHTML = '';
    if (task.status === 'waiting' || task.status === 'running' || task.status === 'review') {
      var cancelDefault = task.status === 'review' ? t('action.refuse') : t('action.cancelTask');
      var cancelKey = 'task-cancel:' + task.id;
      var cancelLabel = isPendingConfirm(cancelKey) ? t('action.cancelTaskConfirm') : cancelDefault;
      secondaryBtnHTML = '<button class="btn-outline btn-action-secondary"' + (accepting ? ' disabled' : '') + ' data-action="confirm-click" data-confirm-key="' + cancelKey + '" data-confirm-action="task-cancel" data-confirm-id="' + task.id + '" data-default-label="' + escapeHtml(cancelDefault) + '" data-confirm-label="' + escapeHtml(t('action.cancelTaskConfirm')) + '">' + escapeHtml(cancelLabel) + '</button>';
    }

    // La suppression part dans le menu « ⋯ » de l'en-tête : une action qu'on
    // fait une fois n'a pas à occuper une ligne rouge au-dessus des onglets
    // qu'on ouvre vingt fois par jour.
    var deleteTaskKey = 'task-delete:' + task.id;
    var deleteTaskLabel = isPendingConfirm(deleteTaskKey) ? t('task.deleteConfirm') : t('task.delete');
    var menuHTML = '<div class="detail-menu-wrap">' +
      '<button class="icon-btn" data-action="toggle-task-menu" title="' + escapeHtml(t('task.moreActions')) + '" aria-label="' + escapeHtml(t('task.moreActions')) + '">⋯</button>' +
      '<div class="card-menu detail-menu hidden" data-task-menu="' + task.id + '">' +
        '<button class="card-menu-item card-menu-item-danger" data-action="confirm-click" data-confirm-key="' + deleteTaskKey + '" data-confirm-action="task-delete" data-confirm-id="' + task.id + '" data-default-label="' + escapeHtml(t('task.delete')) + '" data-confirm-label="' + escapeHtml(t('task.deleteConfirm')) + '">' + escapeHtml(deleteTaskLabel) + '</button>' +
      '</div>' +
    '</div>';

    return '<div class="detail-head">' +
        '<div class="detail-head-row">' +
          buildTaskGlyphHTML(task) +
          '<div class="detail-head-main">' +
            '<div class="detail-title">' + escapeHtml(task.title) + '</div>' +
            '<div class="detail-meta">' +
              buildAgentChipHTML(task, agent, soft) +
              '<span class="mono">' + escapeHtml(task.branch || '') + '</span>' +
              (multiRepo && task.repoName ? '<span class="repo-chip">' + escapeHtml(task.repoName) + '</span>' : '') +
              (task.commitsCount ? '<span class="repo-chip">' + escapeHtml(tCount('ship.commits', task.commitsCount)) + '</span>' : '') +
            '</div>' +
          '</div>' +
          menuHTML +
          '<button class="icon-btn" data-action="close-panel" aria-label="' + escapeHtml(t('common.close')) + '">✕</button>' +
        '</div>' +
        buildStatusBadgeHTML(task) +
        behindRow +
        waitingRow +
        '<div class="action-row">' + primaryBtnHTML +
          secondaryBtnHTML +
          buildTaskPreviewButtonHTML(task) +
          '<span class="checks">' + renderChecks(task.checks) + '</span>' +
        '</div>' +
        '<div class="tabs">' + tabsHTML + '</div>' +
      '</div>';
  }

  // La poignée d'agrandissement se tient sur la frontière qu'elle déplace : le
  // bord gauche du panneau, à mi-hauteur. Une flèche dit dans quel sens le
  // panneau va partir ; F fait la même chose au clavier.
  function panelHandleLabel() {
    return (state.panelExpanded ? t('panel.collapse') : t('panel.expand')) + ' (F)';
  }
  function buildPanelHandleHTML() {
    return '<button class="panel-handle" data-action="toggle-panel-expand" aria-label="' + escapeHtml(panelHandleLabel()) + '" title="' + escapeHtml(panelHandleLabel()) + '">' +
      '<svg viewBox="0 0 16 16" width="14" height="14" aria-hidden="true" focusable="false"><path fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" d="M10 3.5 5.5 8l4.5 4.5"/></svg>' +
      '</button>';
  }

  function buildDetailPanelHTML(task) {
    var agent = state.agentsById[task.agentId] || { emoji: '?', name: '?', model: '?', color: '#ccc' };
    var err = state.detailErrorByTask[task.id];
    var bodyHTML = '';
    if (state.panelTab === 'chat') bodyHTML = buildConversationHTML(task, agent);
    else if (state.panelTab === 'diff') bodyHTML = buildDiffHTML(task);
    else if (state.panelTab === 'files') bodyHTML = buildDeliverablesHTML(task);
    else if (state.panelTab === 'history') bodyHTML = buildHistoryHTML(task);

    return '<aside class="detail-panel">' +
      buildPanelHandleHTML() +
      (err ? '<div class="detail-error">' + escapeHtml(err) + '</div>' : '') +
      buildDetailHeadHTML(task) +
      '<div class="tab-body">' + bodyHTML + '</div>' +
      '</aside>';
  }

  // Rafraîchit uniquement l'en-tête du détail (glyphe, badge de statut,
  // bouton d'action, compteurs d'onglets...) sans toucher à .tab-body ni au
  // conteneur des messages, qui restent des frères du même parent.
  function patchDetailHead(taskId) {
    var task = state.tasksById[taskId];
    var headEl = document.querySelector('.detail-panel .detail-head');
    if (!task || !headEl) return false;
    var wrapper = document.createElement('div');
    wrapper.innerHTML = buildDetailHeadHTML(task);
    var newHead = wrapper.firstElementChild;
    if (newHead) headEl.replaceWith(newHead);
    return true;
  }

  // Rafraîchit la ligne d'activité live d'une seule ligne de la liste de
  // tâches, sans reconstruire la liste ni toucher au panneau de détail.
  function patchTaskRowLiveLine(taskId, line, task) {
    var row = document.querySelector('.task-row[data-task-id="' + taskId + '"]');
    if (!row) return false;
    var metaLine = row.querySelector('.task-meta-line');
    if (!metaLine) return false;
    var existing = metaLine.querySelector('.live-line');
    var showLive = !!(task && task.status === 'running' && line);
    if (showLive) {
      var html = '<span class="live-line mono"><span class="live-dot"></span>' + escapeHtml(line) + '</span>';
      if (existing) existing.outerHTML = html;
      else metaLine.insertAdjacentHTML('beforeend', html);
    } else if (existing) {
      existing.remove();
    }
    return true;
  }

  // ---------------------------------------------------------------------
  // Onglet Conversation
  // ---------------------------------------------------------------------

  function parseReassignMarker(text) {
    var m = /^\[reassigned:([^\]]+)\]$/.exec(String(text || '').trim());
    return m ? m[1] : null;
  }

  // parseMarker extrait la charge d'un message marqueur "[<kind>:<valeur>]"
  // (posé par le backend, volontairement neutre : la phrase est localisée ici).
  function parseMarker(kind, text) {
    var m = new RegExp('^\\[' + kind + ':([^\\]]+)\\]$').exec(String(text || '').trim());
    return m ? m[1] : null;
  }

  function buildMessageHTML(m, agent) {
    var reassignedAgentId = parseReassignMarker(m.text);
    if (reassignedAgentId) {
      var targetAgent = state.agentsById[reassignedAgentId];
      var targetName = targetAgent ? targetAgent.name : reassignedAgentId;
      return '<div class="msg-system">' + escapeHtml(t('chat.reassignedTo', { name: targetName })) + '</div>';
    }
    var acceptedBranch = parseMarker('accepted', m.text);
    if (acceptedBranch) {
      return '<div class="msg-system">' + escapeHtml(t('chat.accepted', { branch: acceptedBranch })) + '</div>';
    }
    // Acceptation constatée : la branche de la tâche était déjà contenue dans
    // celle du chantier (voir autoAcceptMergedTasks côté serveur).
    var autoAcceptedBranch = parseMarker('auto-accepted', m.text);
    if (autoAcceptedBranch) {
      return '<div class="msg-system">' + escapeHtml(t('chat.autoAccepted', { branch: autoAcceptedBranch })) + '</div>';
    }
    var conflictFiles = parseMarker('merge-conflict', m.text);
    if (conflictFiles) {
      return '<div class="msg-system msg-system-warning">' + escapeHtml(t('chat.mergeConflict', { files: conflictFiles.split(' ').join(', ') })) + '</div>';
    }
    // Rebase automatique après l'acceptation d'une autre tâche du chantier
    // (voir rebaseSiblingTasks côté serveur) : réussi, ou abandonné sur conflit.
    var rebasedBranch = parseMarker('rebased', m.text);
    if (rebasedBranch) {
      return '<div class="msg-system">' + escapeHtml(t('chat.rebased', { branch: rebasedBranch })) + '</div>';
    }
    var rebaseConflictFiles = parseMarker('rebase-conflict', m.text);
    if (rebaseConflictFiles) {
      return '<div class="msg-system msg-system-warning">' + escapeHtml(t('chat.rebaseConflict', { files: rebaseConflictFiles.split(' ').join(', ') })) + '</div>';
    }
    // Outil refusé à l'agent : sans cette ligne, le refus ne se voit nulle part
    // et la tâche prend des tours que personne ne sait expliquer.
    var deniedTool = parseMarker('tool-denied', m.text);
    if (deniedTool) {
      return '<div class="msg-system msg-system-warning">' + escapeHtml(t('chat.toolDenied', { tool: deniedTool })) + '</div>';
    }
    // Agent perdu avec le processus qui le portait : Sillage s'est arrêté
    // pendant qu'il travaillait (voir resetTransientTaskFlags côté serveur).
    if (parseMarker('interrupted', m.text)) {
      return '<div class="msg-system msg-system-warning">' + escapeHtml(t('chat.interrupted')) + '</div>';
    }
    var isUser = m.author === 'user';
    var emoji = isUser ? '🙂' : (agent.emoji || '');
    var bg = isUser ? '#eeece6' : softColor(agent.color);
    var name = m.authorName || (isUser ? t('chat.you') : (agent.name || ''));
    // Copier la réponse d'un agent est fréquent (on la porte ailleurs) ; copier
    // sa propre question ne l'est pas. Le bouton ne s'affiche donc que sur les
    // messages d'agent, et seulement au survol.
    var actions = isUser ? '' :
      '<span class="msg-actions">' +
        '<button class="msg-copy" data-action="copy-message" data-copy="' + escapeHtml(m.text) + '"' +
        ' title="' + escapeHtml(t('chat.copyMessage')) + '" aria-label="' + escapeHtml(t('chat.copyMessage')) + '">' +
        copyIconHTML() + '</button>' +
      '</span>';
    return '<div class="message">' +
      '<span class="msg-avatar" style="background:' + bg + '">' + emoji + '</span>' +
      '<div class="msg-body">' +
        '<div class="msg-head"><span class="msg-name">' + escapeHtml(name) + '</span><span class="msg-time">' + formatTime(m.createdAt) + '</span>' + actions + '</div>' +
        (m.text ? '<div class="msg-text">' + renderMarkdown(m.text) + '</div>' : '') +
        buildMessageAttachmentsHTML(m) +
      '</div></div>';
  }

  // Les images jointes s'affichent en vignettes sous le texte du message. La
  // source est l'endpoint du serveur (le fichier est sur le disque, jamais dans
  // l'état) ; le clic ouvre l'image en grand.
  function buildMessageAttachmentsHTML(m) {
    var atts = m.attachments || [];
    if (!atts.length) return '';
    var items = atts.map(function (a) {
      var url = '/api/tasks/' + encodeURIComponent(m.taskId) + '/attachments/' + encodeURIComponent(a.id);
      return '<button class="msg-image" data-action="open-attachment" data-src="' + escapeHtml(url) + '" data-name="' + escapeHtml(a.name || '') + '"' +
        ' title="' + escapeHtml(t('chat.imageOpen')) + '">' +
        '<img src="' + escapeHtml(url) + '" alt="' + escapeHtml(a.name || '') + '" loading="lazy">' +
      '</button>';
    }).join('');
    return '<div class="msg-images">' + items + '</div>';
  }

  // Image en grand : une modale sans habillage, l'image seule sur le voile.
  function openAttachmentModal(src, name) {
    openModal('<div class="modal modal-image">' +
      '<img src="' + escapeHtml(src) + '" alt="' + escapeHtml(name || '') + '">' +
    '</div>');
  }

  // L'attente d'un agent prend le gabarit d'un message, au pied du fil, à la
  // place exacte qu'occupera sa réponse : quand elle arrive, rien ne saute. La
  // ligne du dessous dit ce qu'il fait à l'instant (l'outil en cours, la même
  // information que le fil d'activité de la liste de tâches) ou, à défaut, qu'il
  // travaille. Sans elle, un agent qui réfléchit trois minutes est indiscernable
  // d'un agent bloqué.
  function buildThinkingHTML(task, agent) {
    if (!task || task.status !== 'running') return '';
    return '<div class="msg-thinking" id="conversation-thinking">' +
      '<span class="msg-avatar" style="background:' + softColor(agent.color) + '">' + (agent.emoji || '') + '</span>' +
      '<div class="msg-body">' +
        '<div class="msg-head"><span class="msg-name">' + escapeHtml(agent.name || '') + '</span>' +
          '<span class="thinking-dots" aria-hidden="true"><i></i><i></i><i></i></span></div>' +
        '<div class="thinking-line" aria-live="polite">' + thinkingLineInnerHTML(task, agent) + '</div>' +
      '</div></div>';
  }

  function thinkingLineInnerHTML(task, agent) {
    if (task.liveActivity) {
      return '<span class="live-line mono"><span class="live-dot"></span>' + escapeHtml(task.liveActivity) + '</span>';
    }
    return '<span class="thinking-label">' + escapeHtml(t('chat.working')) + '</span>';
  }

  // Met l'indicateur à jour sans toucher aux messages. Appelé sur chaque
  // événement de tâche et d'activité, donc plusieurs fois par seconde quand un
  // agent travaille : seule la ligne du bas est réécrite, sinon l'animation des
  // points repartirait de zéro à chaque outil.
  function patchConversationThinking(taskId) {
    if (state.taskId !== taskId || state.panelTab !== 'chat') return;
    var container = document.getElementById('conversation-list');
    if (!container) return;
    var task = state.tasksById[taskId];
    var agent = (task && state.agentsById[task.agentId]) || { emoji: '', name: '', color: '#ccc' };
    var existing = document.getElementById('conversation-thinking');
    if (!task || task.status !== 'running') {
      if (existing) existing.remove();
      if (!container.children.length) {
        container.innerHTML = '<div class="empty-note">' + escapeHtml(t('conversation.empty')) + '</div>';
      }
      return;
    }
    var atBottom = isScrolledToBottom(container) && !isSelectionActiveIn(container);
    if (existing) {
      var line = existing.querySelector('.thinking-line');
      if (line) line.innerHTML = thinkingLineInnerHTML(task, agent);
    } else {
      var emptyNote = container.querySelector('.empty-note');
      if (emptyNote) emptyNote.remove();
      container.insertAdjacentHTML('beforeend', buildThinkingHTML(task, agent));
    }
    if (atBottom) container.scrollTop = container.scrollHeight;
  }

  function buildConversationHTML(task, agent) {
    var msgs = state.messagesByTask[task.id];
    if (!msgs) {
      loadMessages(task.id);
      return '<div class="conv-loading">' + escapeHtml(t('common.loading')) + '</div>';
    }
    var thinking = buildThinkingHTML(task, agent);
    var items = msgs.map(function (m) { return buildMessageHTML(m, agent); }).join('');
    if (!items && !thinking) items = '<div class="empty-note">' + escapeHtml(t('conversation.empty')) + '</div>';
    return '<div class="conversation" id="conversation-list">' + items + thinking + '</div>' +
      '<div class="composer-wrap">' + buildConvJumpHTML() + '<div class="composer">' +
        buildAttachDraftsHTML(task.id) +
        '<textarea id="composer-input" rows="1" placeholder="' + escapeHtml(t('chat.placeholder', { name: agent.name })) + '"></textarea>' +
        '<div class="composer-row">' +
          '<button class="composer-attach" data-action="attach-image" data-target="composer-file"' +
            ' title="' + escapeHtml(t('chat.attach')) + '" aria-label="' + escapeHtml(t('chat.attach')) + '">' + attachIconHTML() + '</button>' +
          '<input type="file" id="composer-file" class="hidden" accept="image/png,image/jpeg,image/gif,image/webp" multiple data-scope="task:' + task.id + '">' +
          '<span class="composer-model"><span class="agent-avatar-sm" style="background:' + softColor(agent.color) + '">' + agent.emoji + '</span>' + escapeHtml(agent.model || '') + '</span>' +
          '<button class="btn-send" data-action="send-message" data-task-id="' + task.id + '" disabled>' + escapeHtml(t('chat.send')) + '</button>' +
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

  // ---------------------------------------------------------------------
  // Confort de conversation : rendu incrémental (jamais de reconstruction
  // du fil pour un simple événement SSE ; voir onMessageEvent/onTaskEvent/
  // onActivityEvent plus bas).
  // ---------------------------------------------------------------------

  function isScrolledToBottom(el) {
    return (el.scrollHeight - el.scrollTop - el.clientHeight) < 32;
  }

  function isSelectionActiveIn(container) {
    var sel = window.getSelection();
    if (!sel || sel.isCollapsed || sel.rangeCount === 0) return false;
    try {
      var range = sel.getRangeAt(0);
      return container.contains(range.commonAncestorContainer);
    } catch (e) {
      return false;
    }
  }

  // Ajoute un seul message au fil sans jamais reconstruire le conteneur :
  // le seul cas où le fil est touché pour un événement SSE. L'auto-scroll
  // (mutation non essentielle) est sauté si l'utilisateur est en train de
  // sélectionner du texte dans le fil, ou s'il n'était pas déjà en bas.
  function appendMessageToConversationDOM(m) {
    var container = document.getElementById('conversation-list');
    if (!container) return;
    var emptyNote = container.querySelector('.empty-note');
    if (emptyNote) emptyNote.remove();
    var selecting = isSelectionActiveIn(container);
    var wasAtBottom = !selecting && isScrolledToBottom(container);
    var task = state.tasksById[m.taskId];
    var agent = (task && state.agentsById[task.agentId]) || { emoji: '', name: '', model: '', color: '#ccc' };
    var wrapper = document.createElement('div');
    wrapper.innerHTML = buildMessageHTML(m, agent);
    var node = wrapper.firstElementChild;
    if (node) {
      // Le message monte de quelques pixels en arrivant : il n'était pas là. Les
      // messages déjà présents, eux, ne portent jamais cette classe (voir .msg-in).
      node.classList.add('msg-in');
      // L'indicateur d'attente reste le dernier élément du fil : l'agent qui
      // vient de parler continue de travailler jusqu'à la fin de son tour.
      var thinking = document.getElementById('conversation-thinking');
      if (thinking) container.insertBefore(node, thinking);
      else container.appendChild(node);
    }
    if (wasAtBottom) {
      container.scrollTop = container.scrollHeight;
    } else if (m.author !== 'user') {
      // L'utilisateur lit plus haut : on ne le déplace pas, mais il doit savoir
      // qu'il y a du nouveau en bas.
      state.convUnseen = true;
    }
    refreshConvJump();
  }

  // ---------------------------------------------------------------------
  // Retour en bas du fil
  // ---------------------------------------------------------------------

  function buildConvJumpHTML() {
    return '<button class="conv-jump hidden" id="conv-jump" data-action="jump-to-bottom"></button>';
  }

  // La pastille n'existe que quand on n'est pas en bas d'un fil qui défile.
  // Recalculée à chaque événement de défilement, elle ne réécrit son contenu que
  // si son état a changé, sinon défiler repeindrait pour rien.
  function refreshConvJump() {
    var btn = document.getElementById('conv-jump');
    var container = document.getElementById('conversation-list');
    if (!btn || !container) return;
    var visible = !isScrolledToBottom(container);
    if (!visible) state.convUnseen = false;
    var signature = visible ? (state.convUnseen ? 'new' : 'jump') : 'off';
    if (state.convJumpState === signature) return;
    state.convJumpState = signature;
    btn.classList.toggle('hidden', !visible);
    if (!visible) return;
    var label = state.convUnseen ? t('chat.newMessage') : t('chat.jumpToBottom');
    btn.innerHTML = (state.convUnseen ? '<span class="conv-jump-dot"></span>' : arrowDownIconHTML()) + escapeHtml(label);
    btn.setAttribute('aria-label', label);
  }

  // Le seul défilement animé de l'interface : celui-là, l'utilisateur l'a
  // demandé, et voir le fil passer lui dit d'où il revient.
  function jumpConversationToBottom() {
    var el = document.getElementById('conversation-list');
    if (!el) return;
    state.convUnseen = false;
    if (prefersReducedMotion() || !el.scrollTo) {
      el.scrollTop = el.scrollHeight;
    } else {
      el.scrollTo({ top: el.scrollHeight, behavior: 'smooth' });
    }
    refreshConvJump();
  }

  // ---------------------------------------------------------------------
  // Composeur
  // ---------------------------------------------------------------------

  // Le champ suit ce qu'on écrit, jusqu'à la limite posée par le CSS
  // (max-height) : au-delà, c'est lui qui défile, pas le fil qui disparaît.
  function autoGrowComposer(el) {
    if (!el) return;
    el.style.height = 'auto';
    el.style.height = el.scrollHeight + 'px';
  }

  function refreshComposerState() {
    var el = document.getElementById('composer-input');
    if (!el) return;
    var btn = document.querySelector('.composer [data-action="send-message"]');
    if (!btn) return;
    // Une image seule vaut instruction : le bouton s'allume aussi quand le
    // champ est vide mais qu'une pièce jointe attend.
    btn.disabled = !el.value.trim() && !attachDrafts(btn.getAttribute('data-task-id')).length;
  }

  // ---------------------------------------------------------------------
  // Images jointes (Ctrl+V ou bouton joindre)
  //
  // Deux endroits s'en servent avec le même code : le composeur d'une
  // conversation (brouillon par tâche dans state.attachDrafts) et la modale de
  // création de tâche (brouillon dans modalAttachments). Les images vivent en
  // data-URI le temps qu'on écrive (la vignette et l'envoi lisent la même
  // chose, aucun aller-retour serveur avant d'appuyer sur Envoyer), puis
  // partent en base64 dans le corps JSON : le serveur les pose sur le disque
  // et n'en donne que le chemin à l'agent.
  // ---------------------------------------------------------------------

  var ATTACH_MAX_BYTES = 8 * 1024 * 1024;
  var ATTACH_MAX_COUNT = 6;
  var ATTACH_TYPES = ['image/png', 'image/jpeg', 'image/gif', 'image/webp'];
  var attachKeyN = 0;

  function attachDrafts(taskId) {
    return (taskId && state.attachDrafts[taskId]) || [];
  }

  // Bande de vignettes d'un brouillon. `scope` est repris tel quel sur le
  // bouton de retrait, pour que le clic sache à quel brouillon il s'adresse.
  function buildAttachStripHTML(list, scope) {
    if (!list.length) return '';
    var items = list.map(function (a) {
      return '<div class="attach-chip">' +
        '<img src="' + escapeHtml(a.dataUrl) + '" alt="' + escapeHtml(a.name) + '">' +
        '<button class="attach-remove" data-action="remove-attachment" data-scope="' + escapeHtml(scope) + '" data-key="' + a.key + '"' +
          ' title="' + escapeHtml(t('chat.attachRemove')) + '" aria-label="' + escapeHtml(t('chat.attachRemove')) + '">' + closeIconHTML() + '</button>' +
      '</div>';
    }).join('');
    return items;
  }

  function buildAttachDraftsHTML(taskId) {
    var drafts = attachDrafts(taskId);
    if (!drafts.length) return '';
    return '<div class="attach-strip">' + buildAttachStripHTML(drafts, 'task:' + taskId) + '</div>';
  }

  // readAttachmentFiles filtre des fichiers (presse-papiers ou sélecteur) et
  // lit ceux qui passent. `add` reçoit chaque image lue, `fail` chaque refus :
  // ce qui est refusé le dit et n'entre pas.
  function readAttachmentFiles(files, alreadyThere, add, fail) {
    if (!files || !files.length) return;
    var queue = [];
    for (var i = 0; i < files.length; i++) {
      var f = files[i];
      if (!f) continue;
      var name = f.name || 'image';
      if (ATTACH_TYPES.indexOf(f.type) === -1) {
        fail(t('chat.attachUnsupported', { name: name }));
        continue;
      }
      if (f.size > ATTACH_MAX_BYTES) {
        fail(t('chat.attachTooLarge', { name: name }));
        continue;
      }
      if (alreadyThere + queue.length >= ATTACH_MAX_COUNT) {
        fail(t('chat.attachTooMany'));
        break;
      }
      queue.push(f);
    }
    queue.forEach(function (file) {
      var reader = new FileReader();
      reader.onload = function () {
        attachKeyN++;
        add({
          key: 'a' + attachKeyN,
          name: file.name || 'image',
          mime: file.type,
          size: file.size,
          dataUrl: String(reader.result || '')
        });
      };
      reader.onerror = function () { fail(t('chat.attachFailed', { name: file.name || 'image' })); };
      reader.readAsDataURL(file);
    });
  }

  // Brouillon du composeur d'une conversation.
  function addAttachmentFiles(taskId, files) {
    if (!taskId) return;
    readAttachmentFiles(files, attachDrafts(taskId).length, function (att) {
      var list = state.attachDrafts[taskId] || (state.attachDrafts[taskId] = []);
      list.push(att);
      renderMain();
    }, function (msg) {
      showDetailError(taskId, msg);
    });
  }

  function removeAttachment(scope, key) {
    if (scope === 'new-task') { removeModalAttachment(key); return; }
    var taskId = scope.indexOf('task:') === 0 ? scope.slice(5) : '';
    var list = attachDrafts(taskId).filter(function (a) { return a.key !== key; });
    if (list.length) state.attachDrafts[taskId] = list;
    else delete state.attachDrafts[taskId];
    renderMain();
  }

  // Le corps JSON ne transporte que la charge utile base64, sans le préfixe
  // « data:<mime>;base64, » que le lecteur de fichiers ajoute.
  function attachPayloadOf(list) {
    return list.map(function (a) {
      var comma = a.dataUrl.indexOf(',');
      return { name: a.name, mime: a.mime, data: comma === -1 ? '' : a.dataUrl.slice(comma + 1) };
    });
  }

  function attachPayload(taskId) {
    return attachPayloadOf(attachDrafts(taskId));
  }

  function syncComposer() {
    var el = document.getElementById('composer-input');
    if (!el) return;
    autoGrowComposer(el);
    refreshComposerState();
  }

  // askRebase demande le rebase à l'agent au lieu de le faire côté serveur :
  // un rebase serveur échouerait sur le premier conflit, alors que l'agent sait
  // le résoudre. Le message part par le chemin normal, donc il se met en file
  // tout seul si l'agent tourne encore (voir la mécanique pending du runner).
  function askRebase(taskId) {
    var task = state.tasksById[taskId];
    if (!task) return;
    var behind = taskBehindCount(task);
    if (!behind) return;
    var text = tCount('behind.rebasePrompt', behind, { base: task.base || '' });
    api('/api/tasks/' + taskId + '/messages', { method: 'POST', body: { text: text } })
      .catch(function () {});
  }

  function sendMessage(taskId) {
    var el = document.getElementById('composer-input');
    if (!el) return;
    var text = el.value.trim();
    var attachments = attachPayload(taskId);
    if (!text && !attachments.length) return;
    el.value = '';
    // La bande de vignettes part avec le message : elle est rendue par le
    // composeur, donc il faut le reconstruire.
    if (attachments.length) {
      delete state.attachDrafts[taskId];
      renderMain();
      el = document.getElementById('composer-input') || el;
    }
    el.disabled = true;
    syncComposer();
    api('/api/tasks/' + taskId + '/messages', { method: 'POST', body: { text: text, attachments: attachments } })
      .catch(function () {})
      .then(function () {
        var el2 = document.getElementById('composer-input');
        if (el2) el2.disabled = false;
        syncComposer();
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

  // Le diff se lit d'un bout à l'autre, comme une revue : tous les fichiers à
  // la suite, chacun sous un en-tête qui colle en haut pendant qu'on le lit,
  // avec les numéros de ligne des deux côtés et, sur une ligne modifiée, le
  // fragment qui change surligné. L'arbre des fichiers sert à sauter de l'un à
  // l'autre : colonne à gauche quand le panneau est large, menu sous le résumé
  // sinon (requête de conteneur dans style.css, aucun rendu à refaire). Replier
  // un fichier se fait en place, pour que l'animation parte.

  var DIFF_GENERATED_RE = /(^|\/)(go\.sum|package-lock\.json|yarn\.lock|pnpm-lock\.yaml|composer\.lock|Cargo\.lock|Gemfile\.lock|poetry\.lock|uv\.lock|bun\.lockb)$/;
  var DIFF_HUNK_RE = /^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@ ?(.*)$/;
  var DIFF_BIG_LINES = 600;

  function diffFileLines(f) {
    return (f.hunks || []).reduce(function (n, h) { return n + (h.lines || []).length; }, 0);
  }

  // Ajout ou suppression pure : tous les hunks partent (ou arrivent) de rien.
  function diffFileStatus(f) {
    var hunks = f.hunks || [];
    if (!hunks.length) return '';
    var added = true, deleted = true;
    hunks.forEach(function (h) {
      var m = DIFF_HUNK_RE.exec(h.header || '');
      if (!m) { added = deleted = false; return; }
      if (!(m[1] === '0' && m[2] === '0')) added = false;
      if (!(m[3] === '0' && m[4] === '0')) deleted = false;
    });
    return added ? 'new' : (deleted ? 'deleted' : '');
  }

  function diffFileCollapsed(taskId, f) {
    var map = state.diffCollapsed[taskId] || {};
    if (Object.prototype.hasOwnProperty.call(map, f.path)) return map[f.path];
    return DIFF_GENERATED_RE.test(f.path) || diffFileLines(f) > DIFF_BIG_LINES;
  }

  function diffPathHTML(path) {
    var i = path.lastIndexOf('/');
    var dir = i >= 0 ? path.slice(0, i + 1) : '';
    var name = i >= 0 ? path.slice(i + 1) : path;
    return '<span class="diff-path-dir">' + escapeHtml(dir) + '</span><span class="diff-path-name">' + escapeHtml(name) + '</span>';
  }

  // Cinq cases, comme GitHub : la part d'ajout et de suppression d'un coup d'œil.
  function diffRatioHTML(add, del) {
    var total = add + del;
    var greens = total ? Math.round(5 * add / total) : 0;
    if (add && !greens) greens = 1;
    if (del && greens === 5) greens = 4;
    var reds = total ? 5 - greens : 0;
    var html = '';
    for (var i = 0; i < 5; i++) {
      html += '<i class="' + (i < greens ? 'r-add' : (i < greens + reds ? 'r-del' : '')) + '"></i>';
    }
    return '<span class="diff-ratio" aria-hidden="true">' + html + '</span>';
  }

  function diffStatsHTML(add, del) {
    return '<span class="diff-stats"><span class="diff-add">+' + add + '</span><span class="diff-del">−' + del + '</span></span>';
  }

  // Le fragment qui change entre une ligne supprimée et celle qui la remplace :
  // préfixe et suffixe communs retirés. Rien de surligné quand presque tout
  // change, le surlignage ne dirait alors rien de plus que la couleur de ligne.
  function diffWordSplit(a, b) {
    var max = Math.min(a.length, b.length), p = 0, q = 0;
    while (p < max && a.charAt(p) === b.charAt(p)) p++;
    while (q < max - p && a.charAt(a.length - 1 - q) === b.charAt(b.length - 1 - q)) q++;
    var changed = Math.max(a.length, b.length) - p - q;
    if (p + q === 0 || changed > 0.7 * Math.max(a.length, b.length)) return null;
    return { p: p, q: q };
  }

  function diffTextHTML(text, cut) {
    if (!cut) return escapeHtml(text);
    var end = text.length - cut.q;
    if (end <= cut.p) return escapeHtml(text);
    return escapeHtml(text.slice(0, cut.p)) + '<span class="dw">' + escapeHtml(text.slice(cut.p, end)) + '</span>' + escapeHtml(text.slice(end));
  }

  function buildDiffHunksHTML(f) {
    var html = '';
    var prevEnd = null;
    (f.hunks || []).forEach(function (h) {
      var m = DIFF_HUNK_RE.exec(h.header || '');
      var oldN = m ? +m[1] : 0, newN = m ? +m[3] : 0;
      if (m && prevEnd !== null && oldN > prevEnd) {
        html += '<div class="diff-gap">' + escapeHtml(tCount('diff.skipped', oldN - prevEnd)) + '</div>';
      }
      html += '<div class="diff-hunk-head"><span class="diff-hunk-range">' + escapeHtml(m ? '@@ -' + m[1] + (m[2] !== undefined ? ',' + m[2] : '') + ' +' + m[3] + (m[4] !== undefined ? ',' + m[4] : '') + ' @@' : h.header) + '</span>' +
        (m && m[5] ? '<span class="diff-hunk-ctx">' + escapeHtml(m[5]) + '</span>' : '') + '</div>';
      var lines = h.lines || [];
      // Appariement des blocs « suppressions puis ajouts » ligne à ligne.
      var cuts = new Array(lines.length);
      for (var i = 0; i < lines.length;) {
        if (lines[i].type !== 'del') { i++; continue; }
        var d0 = i;
        while (i < lines.length && lines[i].type === 'del') i++;
        var a0 = i;
        while (i < lines.length && lines[i].type === 'add') i++;
        var pairs = Math.min(a0 - d0, i - a0);
        for (var k = 0; k < pairs; k++) {
          var cut = diffWordSplit(lines[d0 + k].text, lines[a0 + k].text);
          if (!cut) continue;
          cuts[d0 + k] = cut;
          cuts[a0 + k] = cut;
        }
      }
      lines.forEach(function (l, idx) {
        var o = '', n = '', mark = ' ';
        if (l.type === 'del') { o = oldN++; mark = '−'; }
        else if (l.type === 'add') { n = newN++; mark = '+'; }
        else { o = oldN++; n = newN++; }
        html += '<div class="dl dl-' + l.type + '"><span class="ln">' + o + '</span><span class="ln">' + n + '</span><span class="dm">' + mark + '</span><span class="dt">' + diffTextHTML(l.text, cuts[idx]) + '</span></div>';
      });
      prevEnd = oldN;
    });
    return html;
  }

  function diffToolLabel(diff) {
    return diff.tool ? t('diff.openIn', { tool: diff.tool }) : t('diff.openDifftool');
  }

  function buildDiffFileHTML(task, diff, f) {
    var collapsed = diffFileCollapsed(task.id, f);
    var status = diffFileStatus(f);
    var generated = DIFF_GENERATED_RE.test(f.path);
    // Largeur des colonnes de numéros : le plus grand numéro du fichier.
    var maxLine = 1;
    (f.hunks || []).forEach(function (h) {
      var m = DIFF_HUNK_RE.exec(h.header || '');
      if (m) maxLine = Math.max(maxLine, +m[1] + +(m[2] || 1), +m[3] + +(m[4] || 1));
    });
    var badge = status ? '<span class="diff-badge diff-badge-' + status + '">' + escapeHtml(t('diff.' + status)) + '</span>' :
      (generated ? '<span class="diff-badge">' + escapeHtml(t('diff.generated')) + '</span>' : '');
    return '<section class="diff-file' + (collapsed ? ' collapsed' : '') + '" data-diff-path="' + escapeHtml(f.path) + '" style="--ln-w:' + (String(maxLine).length + 1) + 'ch">' +
      '<header class="diff-file-head">' +
        '<button class="diff-file-toggle" data-action="diff-toggle-file" data-task-id="' + task.id + '" data-path="' + escapeHtml(f.path) + '" aria-expanded="' + (collapsed ? 'false' : 'true') + '" title="' + escapeHtml(collapsed ? t('diff.expand') : t('diff.collapse')) + '">' +
          '<span class="diff-chevron" aria-hidden="true"></span>' +
          '<span class="diff-file-path mono">' + diffPathHTML(f.path) + '</span>' +
        '</button>' +
        badge +
        diffStatsHTML(f.additions, f.deletions) +
        diffRatioHTML(f.additions, f.deletions) +
        '<button class="icon-btn-sm diff-file-open" data-action="open-difftool" data-task-id="' + task.id + '" data-path="' + escapeHtml(f.path) + '" title="' + escapeHtml(t('diff.openFile', { tool: diff.tool || 'difftool' })) + '" aria-label="' + escapeHtml(t('diff.openFile', { tool: diff.tool || 'difftool' })) + '">' + externalIconHTML() + '</button>' +
      '</header>' +
      '<div class="diff-file-body"><div class="diff-file-inner">' + (collapsed ? '' : buildDiffHunksHTML(f)) + '</div></div>' +
    '</section>';
  }

  function buildDiffTreeHTML(task, diff) {
    return diff.files.map(function (f) {
      var status = diffFileStatus(f);
      return '<button class="diff-tree-row" data-action="diff-jump" data-task-id="' + task.id + '" data-path="' + escapeHtml(f.path) + '" title="' + escapeHtml(f.path) + '">' +
        '<span class="diff-tree-dot' + (status ? ' dot-' + status : '') + '" aria-hidden="true"></span>' +
        '<span class="diff-tree-name mono">' + diffPathHTML(f.path) + '</span>' +
        diffStatsHTML(f.additions, f.deletions) +
      '</button>';
    }).join('');
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
    var add = 0, del = 0;
    diff.files.forEach(function (f) { add += f.additions; del += f.deletions; });
    var allCollapsed = diff.files.every(function (f) { return diffFileCollapsed(task.id, f); });
    var toolBtn = '<button class="btn-outline diff-open-all" data-action="open-difftool" data-task-id="' + task.id + '" data-path="" title="' + escapeHtml(diff.tool ? t('diff.openAllHint') : t('diff.noToolHint')) + '">' +
      externalIconHTML() + '<span>' + escapeHtml(diffToolLabel(diff)) + '</span></button>';
    return '<div class="diff-view' + (state.diffTreeOpen ? ' tree-open' : '') + '" data-diff-task="' + task.id + '">' +
      '<div class="diff-toolbar">' +
        '<button class="diff-summary" data-action="diff-tree-toggle" aria-expanded="' + (state.diffTreeOpen ? 'true' : 'false') + '">' +
          '<span class="diff-summary-count">' + escapeHtml(tCount('diff.files', diff.files.length)) + '</span>' +
          diffStatsHTML(add, del) + diffRatioHTML(add, del) +
          '<span class="diff-chevron diff-summary-chevron" aria-hidden="true"></span>' +
        '</button>' +
        '<span class="diff-toolbar-spacer"></span>' +
        '<button class="btn-quiet" data-action="diff-toggle-all" data-task-id="' + task.id + '">' + escapeHtml(allCollapsed ? t('diff.expandAll') : t('diff.collapseAll')) + '</button>' +
        toolBtn +
      '</div>' +
      '<div class="diff-layout">' +
        '<nav class="diff-tree" aria-label="' + escapeHtml(tCount('diff.files', diff.files.length)) + '">' + buildDiffTreeHTML(task, diff) + '</nav>' +
        '<div class="diff-scroll" id="diff-scroll">' +
          diff.files.map(function (f) { return buildDiffFileHTML(task, diff, f); }).join('') +
          '<div class="diff-end mono">' + escapeHtml(diff.branch || '') + ' → ' + escapeHtml(diff.base || '') + '</div>' +
        '</div>' +
      '</div>' +
    '</div>';
  }

  // Le pied n'existe plus (la branche est en bas du fil) : rien à patcher.
  function patchDiffFooter() { return true; }

  function findDiffFile(taskId, path) {
    var diff = state.diffByTask[taskId];
    return diff && diff.files ? diff.files.filter(function (f) { return f.path === path; })[0] : null;
  }

  function diffSectionEl(path) {
    var all = document.querySelectorAll('.diff-file');
    for (var i = 0; i < all.length; i++) if (all[i].getAttribute('data-diff-path') === path) return all[i];
    return null;
  }

  // Replie ou déplie en place : le contenu d'un fichier replié n'est pas rendu
  // (un go.sum de 3000 lignes n'a rien à faire dans le DOM), il l'est au
  // premier dépliage, puis la hauteur s'anime.
  function setDiffFileCollapsed(taskId, path, collapsed) {
    state.diffCollapsed[taskId] = state.diffCollapsed[taskId] || {};
    state.diffCollapsed[taskId][path] = collapsed;
    var sec = diffSectionEl(path);
    var f = findDiffFile(taskId, path);
    if (!sec || !f) return;
    var inner = sec.querySelector('.diff-file-inner');
    if (!collapsed && inner && !inner.firstChild) inner.innerHTML = buildDiffHunksHTML(f);
    sec.classList.toggle('collapsed', collapsed);
    var btn = sec.querySelector('.diff-file-toggle');
    if (btn) {
      btn.setAttribute('aria-expanded', collapsed ? 'false' : 'true');
      btn.title = collapsed ? t('diff.expand') : t('diff.collapse');
    }
    refreshDiffToggleAll(taskId);
  }

  function refreshDiffToggleAll(taskId) {
    var diff = state.diffByTask[taskId];
    var btn = document.querySelector('[data-action="diff-toggle-all"]');
    if (!diff || !btn) return;
    var allCollapsed = diff.files.every(function (f) { return diffFileCollapsed(taskId, f); });
    btn.textContent = allCollapsed ? t('diff.expandAll') : t('diff.collapseAll');
  }

  function toggleDiffFile(taskId, path) {
    var f = findDiffFile(taskId, path);
    if (f) setDiffFileCollapsed(taskId, path, !diffFileCollapsed(taskId, f));
  }

  function toggleDiffAll(taskId) {
    var diff = state.diffByTask[taskId];
    if (!diff) return;
    var collapse = !diff.files.every(function (f) { return diffFileCollapsed(taskId, f); });
    diff.files.forEach(function (f) { setDiffFileCollapsed(taskId, f.path, collapse); });
  }

  function toggleDiffTree() {
    state.diffTreeOpen = !state.diffTreeOpen;
    var view = document.querySelector('.diff-view');
    if (view) view.classList.toggle('tree-open', state.diffTreeOpen);
    var btn = document.querySelector('.diff-summary');
    if (btn) btn.setAttribute('aria-expanded', state.diffTreeOpen ? 'true' : 'false');
  }

  // Saut vers un fichier : déplié s'il était replié, son en-tête amené en haut.
  // Le défilement s'anime, c'est un geste de l'utilisateur.
  function jumpToDiffFile(taskId, path) {
    var f = findDiffFile(taskId, path);
    if (!f) return;
    if (diffFileCollapsed(taskId, f)) setDiffFileCollapsed(taskId, path, false);
    if (state.diffTreeOpen) toggleDiffTree();
    var sec = diffSectionEl(path);
    var scroller = document.getElementById('diff-scroll');
    if (!sec || !scroller) return;
    scroller.scrollTo({ top: sec.offsetTop - scroller.offsetTop, behavior: 'smooth' });
    markDiffTreeActive(path);
    diffJump = { path: path, at: Date.now() };
  }

  function markDiffTreeActive(path) {
    var rows = document.querySelectorAll('.diff-tree-row');
    for (var i = 0; i < rows.length; i++) rows[i].classList.toggle('active', rows[i].getAttribute('data-path') === path);
  }

  // Le fichier lu est celui dont l'en-tête a passé le haut de la zone.
  function currentDiffPath() {
    var scroller = document.getElementById('diff-scroll');
    if (!scroller) return null;
    var secs = scroller.querySelectorAll('.diff-file');
    var top = scroller.scrollTop + 8, cur = secs.length ? secs[0] : null;
    for (var i = 0; i < secs.length; i++) {
      if (secs[i].offsetTop - scroller.offsetTop <= top) cur = secs[i]; else break;
    }
    return cur ? cur.getAttribute('data-diff-path') : null;
  }

  // Dernier saut demandé : pendant que le défilement animé court, c'est lui le
  // fichier courant, sinon deux J rapides viseraient deux fois le même.
  var diffJump = null;
  var diffSpyFrame = 0;
  function onDiffScroll(e) {
    if (!e.target || e.target.id !== 'diff-scroll' || diffSpyFrame) return;
    diffSpyFrame = requestAnimationFrame(function () {
      diffSpyFrame = 0;
      var p = currentDiffPath();
      if (p) markDiffTreeActive(p);
    });
  }

  // j / k : fichier suivant ou précédent, depuis celui qu'on lit.
  function stepDiffFile(dir) {
    var task = state.taskId ? state.tasksById[state.taskId] : null;
    var diff = task ? state.diffByTask[task.id] : null;
    if (!diff || !diff.files || !diff.files.length) return;
    var cur = diffJump && Date.now() - diffJump.at < 700 ? diffJump.path : currentDiffPath();
    var idx = diff.files.findIndex(function (f) { return f.path === cur; });
    var scroller = document.getElementById('diff-scroll');
    // Au tout début, « suivant » mène au premier fichier lui-même s'il n'est pas en haut.
    var next = Math.max(0, Math.min(diff.files.length - 1, idx + dir));
    if (dir > 0 && scroller && scroller.scrollTop < 4 && idx <= 0 && cur === diff.files[0].path && !(diffJump && Date.now() - diffJump.at < 700)) next = Math.min(1, diff.files.length - 1);
    jumpToDiffFile(task.id, diff.files[next].path);
  }

  // doOpenDifftool : demande au serveur de lancer l'outil de diff configuré
  // par git (diff.tool) sur le fichier actif, dans le worktree de la tâche.
  // Local, pas de rendu à rafraîchir ; seul un échec (aucun outil configuré)
  // remonte à l'écran.
  function doOpenDifftool(taskId, path) {
    api('/api/tasks/' + taskId + '/difftool', { method: 'POST', body: { path: path } })
      .catch(function (e) { if (e instanceof ApiError) showDetailError(taskId, e.message || t('errors.genericFailed')); });
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
  // Onglet Historique
  // ---------------------------------------------------------------------

  // Commandes jouées par l'agent (tool_use), déjà portées par task.commandLog :
  // pas de chargement asynchrone, contrairement à Diff/Livrables. Les plus
  // récentes en premier, pour repérer tout de suite ce que l'agent vient de faire.
  function buildHistoryHTML(task) {
    var log = task.commandLog || [];
    if (!log.length) {
      return '<div class="empty-note">' + escapeHtml(t('history.empty')) + '</div>';
    }
    var items = log.slice().reverse().map(function (e) {
      return '<div class="history-item"><span class="history-time mono">' + escapeHtml(formatTime(e.at)) + '</span>' +
        '<span class="history-text mono">' + escapeHtml(e.text) + '</span></div>';
    }).join('');
    return '<div class="history-list">' + items + '</div>';
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
  // doStartWaitingTask : démarre une tâche "waiting" avant que sa dépendance ne
  // soit acceptée (dépendance refusée/supprimée, ou changement d'avis). Un
  // clic, pas de confirmation : aucune donnée n'est perdue, la tâche allait
  // démarrer tôt ou tard.
  function doStartWaitingTask(taskId) {
    api('/api/tasks/' + taskId + '/start', { method: 'POST' }).then(function (task) {
      upsertTask(task); renderMain();
    }).catch(function (e) { if (e instanceof ApiError) showDetailError(taskId, e.message || t('errors.genericFailed')); });
  }
  function doReopen(taskId) {
    var previous = state.tasksById[taskId];
    var cardId = previous ? previous.cardId : null;
    api('/api/tasks/' + taskId + '/reopen', { method: 'POST' }).then(function (task) {
      upsertTask(task);
      if (cardId) loadDelivery(cardId);
      renderMain();
    }).catch(function (e) { if (e instanceof ApiError) showDetailError(taskId, e.message || t('errors.genericFailed')); });
  }
  // Le sillage se joue après le rafraîchissement qui a reconstruit les surfaces,
  // pas au retour du POST : l'aperçu de livraison arrive juste derrière et
  // remplace la ligne de la tâche, l'en-tête du détail et la barre du chantier,
  // ce qui emporterait une onde lancée trop tôt. D'où ce drapeau, armé par
  // l'acceptation ou la livraison et consommé par loadDelivery. Un aperçu déjà
  // en vol au moment de l'armement fait aussi bien l'affaire : sa propre fin
  // reconstruit les mêmes surfaces.
  var pendingWake = null; // { kind: 'accept' | 'ship', id }
  function armWake(kind, id) {
    pendingWake = { kind: kind, id: id };
    // Périmé au bout de trois secondes : passé ce délai, plus rien à l'écran ne
    // rattacherait l'onde au geste qui l'a déclenchée.
    setTimeout(function () {
      if (pendingWake && pendingWake.kind === kind && pendingWake.id === id) pendingWake = null;
    }, 3000);
  }
  function playPendingWake() {
    var w = pendingWake;
    if (!w) return;
    pendingWake = null;
    if (w.kind === 'accept') playAcceptWake(w.id);
    else playShipWake(w.id);
  }

  // playAcceptWake : l'onde passe là où quelque chose vient de changer, et
  // nulle part ailleurs. La ligne de la tâche (les deux boutons y ont laissé
  // place à « Acceptée ») et, si le détail est ouvert sur cette tâche, son
  // bandeau d'état.
  function playAcceptWake(taskId) {
    playWake(document.querySelector('.task-row[data-task-id="' + taskId + '"]'));
    if (state.taskId === taskId) playWake(document.querySelector('.detail-panel .status-badge'));
  }

  // playShipWake : le chantier vient de partir. L'onde traverse l'emplacement
  // de livraison de l'en-tête (le bouton y a cédé la place à l'ancre) et la
  // barre de livraison quand elle a quelque chose à montrer (les liens de pull
  // request) ; sur une barre masquée, elle n'aurait aucune surface.
  function playShipWake(cardId) {
    if (state.cardId !== cardId) return;
    playWake(document.querySelector('.ship-slot'));
    var bar = document.querySelector('.ship-bar');
    if (bar && !bar.classList.contains('ship-bar-empty')) playWake(bar);
  }

  // acceptErrorText localise le conflit de fusion (le seul échec attendu),
  // avec le message de l'API en repli pour tout le reste.
  function acceptErrorText(e) {
    if (e.status === 409) return t('errors.acceptConflict');
    return e.message || t('errors.acceptFailed');
  }

  // doAcceptTask : un clic, pas de confirmation (fusion locale dans la branche
  // du chantier, réversible par « Rouvrir »). Un conflit laisse la tâche en
  // revue et affiche l'invite à faire reprendre la base par l'agent.
  function doAcceptTask(taskId) {
    var task = state.tasksById[taskId];
    var cardId = task ? task.cardId : null;
    if (state.acceptingByTask[taskId]) return;
    state.acceptingByTask[taskId] = true;
    renderMain();
    api('/api/tasks/' + taskId + '/accept', { method: 'POST' }).then(function (res) {
      delete state.acceptingByTask[taskId];
      if (res && res.task) upsertTask(res.task);
      armWake('accept', taskId);
      if (cardId) loadDelivery(cardId);
      renderMain();
      // Sans chantier à recharger, personne ne viendra consommer le drapeau.
      if (!cardId) playPendingWake();
    }).catch(function (e) {
      delete state.acceptingByTask[taskId];
      if (!(e instanceof ApiError)) { renderMain(); return; }
      // Le message marqueur de conflit arrive par SSE : rien à recharger ici.
      // L'erreur s'affiche dans le panneau de détail : si l'acceptation venait
      // de la liste (panneau fermé), ouvrir la tâche, sinon le message serait
      // invisible et le clic semblerait sans effet.
      showDetailError(taskId, acceptErrorText(e));
      if (state.taskId !== taskId) openTask(taskId);
    });
  }
  function doCancelTask(taskId) {
    var task = state.tasksById[taskId];
    var cardId = task ? task.cardId : null;
    api('/api/tasks/' + taskId + '/cancel', { method: 'POST' }).then(function (task) {
      upsertTask(task);
      if (cardId) loadDelivery(cardId);
      renderMain();
    }).catch(function (e) { if (e instanceof ApiError) showDetailError(taskId, e.message || t('errors.cancelFailed')); });
  }
  function doDeleteTask(taskId) {
    api('/api/tasks/' + taskId, { method: 'DELETE', body: { confirm: true } }).then(function () {
      removeTaskLocally(taskId);
      closePanel();
    }).catch(function (e) { if (e instanceof ApiError) showDetailError(taskId, e.message || t('errors.deleteTaskFailed')); });
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
  // Un seul panneau de détail est ouvert à la fois : pas besoin d'identifiant.
  function toggleTaskMenu() {
    var el = document.querySelector('.detail-menu');
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
    if (state.signalId) return buildSignalDetailHTML();
    if (state.projectId) return buildKanbanHTML();
    if (state.screen === 'inbox') return buildInboxHTML();
    if (state.screen === 'vigie') return buildVigieHTML();
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
    var prevDiff = document.getElementById('diff-scroll');
    var prevDiffScroll = prevDiff ? prevDiff.scrollTop : null;

    mainEl.innerHTML = buildMainHTML();

    var newDiff = document.getElementById('diff-scroll');
    if (newDiff) {
      if (prevDiffScroll !== null) newDiff.scrollTop = prevDiffScroll;
      var cur = currentDiffPath();
      if (cur) markDiffTreeActive(cur);
    }

    var newComposer = document.getElementById('composer-input');
    if (newComposer) {
      if (draft) newComposer.value = draft;
      // Le champ reprend la main même sans brouillon : coller une image le
      // reconstruit alors qu'il est vide, et on écrit juste après.
      if (hadFocus) {
        newComposer.focus();
        if (selStart !== null) { try { newComposer.setSelectionRange(selStart, selStart); } catch (e) {} }
      }
    }
    var newConv = document.getElementById('conversation-list');
    if (newConv) {
      if (state.pendingConversationScroll) {
        scrollConversationToBottom();
        state.pendingConversationScroll = false;
      } else if (prevScroll !== null) {
        newConv.scrollTop = prevScroll;
      }
      // La pastille et le champ sont neufs : leur état peint est à refaire, quelle
      // que soit la valeur retenue avant le rendu.
      state.convJumpState = null;
      refreshConvJump();
    }
    if (newComposer) syncComposer();
  }

  // ---------------------------------------------------------------------
  // Recherche (Ctrl/Cmd+K)
  // ---------------------------------------------------------------------

  function overlayBgCloseSearch(e) { if (e.target === e.currentTarget) closeSearch(); }

  // Score de pertinence d'un texte pour une requête déjà connue « contenue » :
  // correspondance exacte ou en tête > en tête de mot > au milieu d'un mot.
  // Permet de faire remonter les meilleurs résultats plutôt que de garder
  // l'ordre de création des entités.
  function searchTextScore(text, query) {
    var idx = text.toLowerCase().indexOf(query);
    if (idx === -1) return -1;
    if (idx === 0) return text.length === query.length ? 3 : 2;
    if (/[^a-z0-9]/i.test(text.charAt(idx - 1))) return 1;
    return 0;
  }
  function searchRefScore(ref, query, isNumericQuery) {
    if (!isNumericQuery) return -1;
    var s = String(ref);
    if (s === query) return 5;
    if (s.indexOf(query) === 0) return 4;
    return -1;
  }
  function rankSearchResults(list, query, isNumericQuery, textOf, refOf, limit) {
    var scored = [];
    list.forEach(function (item) {
      var score = searchTextScore(textOf(item), query);
      if (refOf) score = Math.max(score, searchRefScore(refOf(item), query, isNumericQuery));
      if (score > -1) scored.push({ item: item, score: score });
    });
    scored.sort(function (a, b) { return b.score - a.score; });
    return scored.slice(0, limit).map(function (x) { return x.item; });
  }

  function buildSearchResultsHTML(q) {
    var query = q.trim().toLowerCase();
    if (!query) return '<div class="empty-note">' + escapeHtml(t('search.typeToSearch')) + '</div>';
    var isNumericQuery = /^\d+$/.test(query);
    var projects = rankSearchResults(state.projects, query, isNumericQuery, function (p) { return p.name; }, null, 20);
    var cards = rankSearchResults(state.cards, query, isNumericQuery, function (c) { return c.title; }, function (c) { return c.ref; }, 20);
    var tasks = rankSearchResults(state.tasks, query, isNumericQuery, function (t) { return t.title; }, function (t) { return t.ref; }, 20);
    if (!projects.length && !cards.length && !tasks.length) return '<div class="empty-note">' + escapeHtml(t('search.noResults')) + '</div>';
    var html = '';
    if (projects.length) {
      html += '<div class="search-group-label">' + escapeHtml(t('common.projects')) + '</div>' + projects.map(function (p) {
        return '<button class="search-result" data-action="search-goto-project" data-project-id="' + p.id + '"><span class="hash">#</span>' + escapeHtml(p.name) + '</button>';
      }).join('');
    }
    if (cards.length) {
      html += '<div class="search-group-label">' + escapeHtml(t('common.workstreamsWord')) + '</div>' + cards.map(function (c) {
        var p = state.projectsById[c.projectId];
        return '<button class="search-result" data-action="search-goto-card" data-card-id="' + c.id + '"><span class="mono">#' + c.ref + '</span> ' +
          escapeHtml(c.title) + '<span class="muted-sm">' + (p ? escapeHtml(p.name) : '') + '</span></button>';
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
      '<div class="search-results" id="search-results">' + buildSearchResultsHTML(q) + '</div>' +
      '<div class="search-foot"><span class="kbd">↑ ↓ ' + escapeHtml(t('shortcuts.searchNav')) + '</span>' +
      '<span class="kbd">⏎ ' + escapeHtml(t('shortcuts.searchOpen')) + '</span></div></div>';
  }

  // Le résultat actif (surligné, cible de la touche Entrée) est repéré par sa
  // position dans la liste : la liste est reconstruite à chaque frappe.
  function searchResultEls() {
    return Array.prototype.slice.call(document.querySelectorAll('#search-results .search-result'));
  }
  function highlightSearchResult() {
    var els = searchResultEls();
    if (!els.length) { searchIndex = 0; return; }
    if (searchIndex >= els.length) searchIndex = els.length - 1;
    if (searchIndex < 0) searchIndex = 0;
    els.forEach(function (el, i) {
      var on = i === searchIndex;
      el.classList.toggle('active', on);
      if (on && el.scrollIntoView) el.scrollIntoView({ block: 'nearest' });
    });
  }
  function moveSearchResult(delta) {
    var els = searchResultEls();
    if (!els.length) return;
    searchIndex = (searchIndex + delta + els.length) % els.length;
    highlightSearchResult();
  }
  function activateSearchResult() {
    var els = searchResultEls();
    if (els[searchIndex]) els[searchIndex].click();
  }

  function openSearch() {
    state.searchOpen = true;
    searchIndex = 0;
    var root = document.getElementById('search-overlay');
    root.classList.remove('hidden');
    root.innerHTML = buildSearchHTML('');
    root.addEventListener('click', overlayBgCloseSearch);
    var input = document.getElementById('search-input');
    input.focus();
    input.addEventListener('input', function () {
      var results = document.getElementById('search-results');
      if (!results) return;
      results.innerHTML = buildSearchResultsHTML(input.value);
      searchIndex = 0;
      highlightSearchResult();
    });
  }
  function closeSearch() {
    state.searchOpen = false;
    searchIndex = 0;
    var root = document.getElementById('search-overlay');
    root.classList.add('hidden');
    root.innerHTML = '';
    root.removeEventListener('click', overlayBgCloseSearch);
  }

  // ---------------------------------------------------------------------
  // Modales génériques
  // ---------------------------------------------------------------------

  function overlayBgCloseModal(e) { if (e.target === e.currentTarget) closeModal(); }

  // Après openModal, le formulaire est déjà dans le DOM : focaliser tout de
  // suite, sans setTimeout, pour ne perdre aucune frappe.
  function focusById(id) {
    var el = document.getElementById(id);
    if (el) el.focus();
  }

  function openModal(html) {
    state.modal = true;
    // Le panneau de recette est une modale comme les autres : ouvrir autre chose
    // ferme sa portée, sinon les rafraîchissements SSE viseraient un DOM parti.
    state.previewScope = null;
    var root = document.getElementById('modal-root');
    root.innerHTML = '<div class="modal-overlay" id="modal-overlay">' + html + '</div>';
    document.getElementById('modal-overlay').addEventListener('click', overlayBgCloseModal);
  }
  function closeModal() {
    state.modal = null;
    state.previewScope = null;
    resetPlanWait();
    resetIssuePick();
    var root = document.getElementById('modal-root');
    root.innerHTML = '';
  }

  // ---------------------------------------------------------------------
  // Aide des raccourcis clavier (touche « ? »)
  // ---------------------------------------------------------------------

  function shortcutSections() {
    var mod = modKeyLabel();
    return [
      { label: t('shortcuts.sectionGlobal'), rows: [
        [[mod + '+K', '/'], t('shortcuts.search')],
        [['N'], t('shortcuts.create')],
        [['?'], t('shortcuts.help')],
        [['Esc'], t('shortcuts.escape')]
      ] },
      { label: t('shortcuts.sectionForm'), rows: [
        [[mod + '+⏎'], t('shortcuts.submit')],
        [[mod + '+⇧+⏎'], t('shortcuts.submitAnother')],
        [['←', '→', '↑', '↓'], t('shortcuts.pickAgent')],
        [['Tab'], t('shortcuts.tab')]
      ] },
      { label: t('shortcuts.sectionSearch'), rows: [
        [['↑', '↓'], t('shortcuts.searchNav')],
        [['⏎'], t('shortcuts.searchOpen')]
      ] },
      { label: t('shortcuts.sectionTask'), rows: [
        [['⏎'], t('shortcuts.sendMessage')],
        [['F'], t('shortcuts.expandPanel')],
        [['J', 'K'], t('shortcuts.diffNav')]
      ] }
    ];
  }

  function buildShortcutsModalHTML() {
    var body = shortcutSections().map(function (section) {
      var rows = section.rows.map(function (row) {
        var keys = row[0].map(function (k) { return '<kbd class="kbd-key">' + escapeHtml(k) + '</kbd>'; }).join('');
        return '<div class="shortcut-row"><span class="shortcut-keys">' + keys + '</span>' +
          '<span class="shortcut-desc">' + escapeHtml(row[1]) + '</span></div>';
      }).join('');
      return '<div class="modal-label">' + escapeHtml(section.label) + '</div>' + rows;
    }).join('');
    return '<div class="modal modal-sm">' +
      '<div class="modal-head"><span class="modal-title">' + escapeHtml(t('shortcuts.title')) + '</span>' +
      '<button class="icon-btn" data-action="close-modal" aria-label="' + escapeHtml(t('common.close')) + '">✕</button></div>' +
      body +
      '<div class="modal-foot"><button class="btn-outline" data-action="close-modal">' + escapeHtml(t('common.close')) + '</button></div>' +
      '</div>';
  }

  function openShortcutsModal() {
    openModal(buildShortcutsModalHTML());
  }

  // Nouvelle tâche

  function buildAgentChoiceWarningHTML(agent) {
    if (!agent || !agent.warning) return '';
    return '<div class="agent-choice-warning">⚠ ' + escapeHtml(agentWarningText(agent.warning)) +
      agentWarningExtrasHTML(agent.warning, agent.id) + '</div>';
  }

  // Groupe de radios : un seul arrêt de tabulation, les flèches changent
  // d'agent (tabindex roulant, cf. moveAgentChoice). Partagé par la modale de
  // création de tâche et celle du plan de tâches, qui n'en montre qu'une partie.
  function buildAgentChoicesHTML(agents) {
    return agents.map(function (a) {
      var isSel = a.id === modalAgentId;
      var warn = a.warning ? '<span class="agent-warning" title="' + escapeHtml(agentWarningText(a.warning)) + '">⚠</span>' : '';
      return '<button class="agent-choice ' + (isSel ? 'selected' : '') + '" data-action="pick-agent" data-agent-id="' + a.id + '"' +
        ' role="radio" aria-checked="' + (isSel ? 'true' : 'false') + '" tabindex="' + (isSel ? '0' : '-1') + '">' +
        '<span class="agent-choice-avatar" style="background:' + softColor(a.color) + '">' + a.emoji + '</span>' +
        '<span class="agent-choice-info"><span class="agent-choice-name">' + escapeHtml(a.name) + warn + '</span>' +
        '<span class="agent-choice-model mono">' + escapeHtml(a.model || '') + '</span></span></button>';
    }).join('');
  }

  function buildNewTaskModalHTML(card) {
    var agentChoices = buildAgentChoicesHTML(state.agents);
    var selected = state.agentsById[modalAgentId];
    var project = state.projectsById[card.projectId];
    var repos = (project && project.repos) || [];
    var draft = newTaskDraft || { title: '', prompt: '', repoName: '', waitsForTaskId: '' };
    var repoSelectHTML = '';
    if (repos.length > 1) {
      var repoOptions = repos.map(function (r) {
        return '<option value="' + escapeHtml(r.name) + '"' + (r.name === draft.repoName ? ' selected' : '') + '>' + escapeHtml(r.name) + '</option>';
      }).join('');
      repoSelectHTML = '<div class="modal-label">' + escapeHtml(t('newTask.repoLabel')) + '</div>' +
        '<select id="new-task-repo" class="modal-input">' + repoOptions + '</select>';
    }
    // Importer une issue GitHub, c'est remplir ce formulaire-là : un titre et
    // une consigne, relus et corrigés avant qu'un agent démarre. D'où une
    // entrée ici plutôt qu'un troisième bouton au-dessus de la liste des
    // tâches. Visible seulement si un dépôt du projet est sur github.com.
    var githubRepos = (project && project.githubRepos) || [];
    var importHTML = githubRepos.length && !newTaskSignalId
      ? '<div class="modal-import-row">' +
        '<button class="attach-btn" data-action="open-issue-picker" data-card-id="' + card.id + '" title="' + escapeHtml(t('issues.importTooltip')) + '">' +
        issueIconHTML() + '<span>' + escapeHtml(t('issues.import')) + '</span></button>' +
        '</div>'
      : '';
    // Démarrer après : uniquement les tâches du chantier pas encore terminales
    // (une tâche accepted/cancelled ne fera plus jamais rien). Absent du
    // formulaire s'il n'y a rien à attendre, pour ne pas encombrer le cas
    // courant.
    var waitCandidates = state.tasks.filter(function (tk) {
      return tk.cardId === card.id && (tk.status === 'running' || tk.status === 'review' || tk.status === 'waiting');
    });
    var waitsForSelectHTML = '';
    if (waitCandidates.length > 0) {
      var waitOptions = '<option value="">' + escapeHtml(t('newTask.waitsForNone')) + '</option>' +
        waitCandidates.map(function (tk) {
          return '<option value="' + escapeHtml(tk.id) + '"' + (tk.id === draft.waitsForTaskId ? ' selected' : '') + '>#' + tk.ref + ' ' + escapeHtml(tk.title) + '</option>';
        }).join('');
      waitsForSelectHTML = '<div class="modal-label">' + escapeHtml(t('newTask.waitsForLabel')) + '</div>' +
        '<select id="new-task-waits-for" class="modal-input">' + waitOptions + '</select>';
    }
    return '<div class="modal">' +
      '<div class="modal-head"><span class="modal-title">' + escapeHtml(t('newTask.title')) + '</span><span class="modal-sub">' + escapeHtml(card.title) + '</span>' +
      '<button class="icon-btn" data-action="close-modal" aria-label="' + escapeHtml(t('common.close')) + '">✕</button></div>' +
      importHTML +
      '<input id="new-task-title" class="modal-input" placeholder="' + escapeHtml(t('newTask.titlePlaceholder')) + '" value="' + escapeHtml(draft.title) + '">' +
      // Plus de lignes quand la consigne est déjà là : une issue importée se
      // relit, et trois lignes avec un ascenseur ne se relisent pas.
      '<textarea id="new-task-prompt" class="modal-textarea" placeholder="' + escapeHtml(t('newTask.promptPlaceholder')) + '" rows="' + (draft.prompt ? 8 : 3) + '">' + escapeHtml(draft.prompt) + '</textarea>' +
      // Joindre une image dès la création : le trombone et la bande de
      // vignettes du composeur, au même endroit de la lecture (sous le champ
      // qu'elles accompagnent).
      '<div class="modal-attach-row">' +
        // Le libellé fait partie du bouton : une icône seule de 14px est une
        // cible trop petite, et le texte à côté donnait envie de cliquer là où
        // il ne se passait rien.
        '<button class="attach-btn" data-action="attach-image" data-target="new-task-file">' +
          attachIconHTML() + '<span>' + escapeHtml(t('newTask.attachHint')) + '</span></button>' +
        '<input type="file" id="new-task-file" class="hidden" accept="image/png,image/jpeg,image/gif,image/webp" multiple data-scope="new-task">' +
      '</div>' +
      '<div class="attach-strip" id="new-task-attach-strip">' + buildAttachStripHTML(modalAttachments, 'new-task') + '</div>' +
      repoSelectHTML +
      '<div class="modal-label modal-label-row"><span>' + escapeHtml(t('newTask.agentLabel')) + '</span>' +
      '<span class="modal-label-hint">' + escapeHtml(t('newTask.agentHint')) + '</span></div>' +
      '<div class="agent-choices" id="agent-choices" role="radiogroup" aria-label="' + escapeHtml(t('newTask.agentLabel')) + '">' + agentChoices + '</div>' +
      '<div id="agent-choice-warning">' + buildAgentChoiceWarningHTML(selected) + '</div>' +
      (project && project.contextPrompt ? '<div class="project-context-note">' + escapeHtml(t('newTask.projectContextNote')) + '</div>' : '') +
      (card && card.contextPrompt ? '<div class="project-context-note">' + escapeHtml(t('newTask.workstreamContextNote')) + '</div>' : '') +
      waitsForSelectHTML +
      '<div id="new-task-error" class="modal-error hidden"></div>' +
      '<div id="new-task-note" class="modal-note modal-note-success hidden" role="status" aria-live="polite"></div>' +
      // Pas d'indice ici : les deux libellés de boutons disent déjà ce qui suit
      // la création, et trois boutons remplissent la largeur de la modale.
      '<div class="modal-foot">' +
      '<button class="btn-outline" data-action="close-modal">' + escapeHtml(t('common.cancel')) + '</button>' +
      '<button class="btn-neutral" data-action="submit-new-task-another" data-card-id="' + card.id + '"' +
      ' title="' + escapeHtml(t('newTask.submitAnotherTooltip', { mod: modKeyLabel() })) + '">' + escapeHtml(t('newTask.submitAnother')) +
      '<span class="kbd kbd-in-btn" aria-hidden="true">' + escapeHtml(modKeyLabel()) + '⇧⏎</span></button>' +
      '<button class="btn-green" data-action="submit-new-task" data-card-id="' + card.id + '"' +
      ' title="' + escapeHtml(t('newTask.submitTooltip', { mod: modKeyLabel() })) + '">' + escapeHtml(t('newTask.submit')) +
      '<span class="kbd kbd-in-btn" aria-hidden="true">' + escapeHtml(modKeyLabel()) + '⏎</span></button></div>' +
      '</div>';
  }

  // En mode projet, pas de chantier : une pseudo-carte porte le projet et le
  // nom affiché sous le titre de la modale. Son id vide dit au reste du code
  // qu'aucune tâche du chantier n'existe encore.
  function newTaskCard(cardId) {
    if (cardId) return state.cardsById[cardId] || null;
    var project = newTaskProjectId && state.projectsById[newTaskProjectId];
    if (!project) return null;
    return { id: '', projectId: project.id, title: project.name, contextPrompt: '' };
  }

  // Une tâche, tout de suite, sans passer par la création d'un chantier : le
  // cas courant. Le chantier existe quand même (c'est la branche, donc la PR),
  // mais l'humain n'a pas à le nommer deux fois.
  function openNewTaskInProject(projectId, draft, signalId) {
    if (!state.projectsById[projectId]) return;
    newTaskProjectId = projectId;
    newTaskSignalId = signalId || null;
    var defaultAgent = state.agentsById.bolt || state.agents.find(function (a) { return !a.warning; }) || state.agents[0];
    modalAgentId = defaultAgent ? defaultAgent.id : null;
    modalAttachments = [];
    newTaskDraft = draft || null;
    renderNewTaskModal('', draft && draft.prompt ? 'new-task-prompt' : 'new-task-title');
  }

  function openNewTaskModal(cardId) {
    newTaskProjectId = null;
    newTaskSignalId = null;
    var card = state.cardsById[cardId];
    if (!card) return;
    var defaultAgent = state.agentsById.bolt || state.agents.find(function (a) { return !a.warning; }) || state.agents[0];
    modalAgentId = defaultAgent ? defaultAgent.id : null;
    modalAttachments = [];
    newTaskDraft = null;
    renderNewTaskModal(cardId, 'new-task-title');
  }

  // Rouvre la modale de création avec le brouillon en cours : c'est le retour
  // du choix d'une issue, qui a occupé le même emplacement entre-temps.
  // focusId dit où reprendre la saisie : le titre pour un formulaire vierge, la
  // consigne quand une issue vient de la remplir (c'est elle qu'on relit).
  function renderNewTaskModal(cardId, focusId) {
    var card = newTaskCard(cardId);
    if (!card) return;
    openModal(buildNewTaskModalHTML(card));
    // Focus synchrone : la modale est déjà dans le DOM, et un focus différé
    // laisse filer les premières frappes de qui tape juste après N.
    var el = document.getElementById(focusId || 'new-task-title');
    if (!el) return;
    el.focus();
    if (el.value) el.setSelectionRange(el.value.length, el.value.length);
  }

  // Le formulaire vit dans le DOM ; le relire avant de céder la place est ce
  // qui permet de revenir sans avoir rien perdu.
  function captureNewTaskDraft() {
    var titleEl = document.getElementById('new-task-title');
    var promptEl = document.getElementById('new-task-prompt');
    var repoEl = document.getElementById('new-task-repo');
    var waitsForEl = document.getElementById('new-task-waits-for');
    newTaskDraft = {
      title: titleEl ? titleEl.value : '',
      prompt: promptEl ? promptEl.value : '',
      repoName: repoEl ? repoEl.value : '',
      waitsForTaskId: waitsForEl ? waitsForEl.value : ''
    };
  }

  // La modale n'est pas rendue par renderMain : seule sa bande de vignettes est
  // réécrite, pour ne pas perdre le titre et le prompt déjà saisis.
  function refreshModalAttachStrip() {
    var el = document.getElementById('new-task-attach-strip');
    if (el) el.innerHTML = buildAttachStripHTML(modalAttachments, 'new-task');
  }

  function addModalAttachmentFiles(files) {
    readAttachmentFiles(files, modalAttachments.length, function (att) {
      modalAttachments.push(att);
      refreshModalAttachStrip();
    }, function (msg) {
      var errEl = document.getElementById('new-task-error');
      if (errEl) { errEl.textContent = msg; errEl.classList.remove('hidden'); }
    });
  }

  function removeModalAttachment(key) {
    modalAttachments = modalAttachments.filter(function (a) { return a.key !== key; });
    refreshModalAttachStrip();
  }

  function pickAgentInModal(agentId, focusIt) {
    modalAgentId = agentId;
    document.querySelectorAll('#agent-choices .agent-choice').forEach(function (el) {
      var isSel = el.getAttribute('data-agent-id') === agentId;
      el.classList.toggle('selected', isSel);
      el.setAttribute('aria-checked', isSel ? 'true' : 'false');
      el.setAttribute('tabindex', isSel ? '0' : '-1');
      if (isSel && focusIt) el.focus();
    });
    var warning = document.getElementById('agent-choice-warning');
    var a = state.agentsById[agentId];
    if (warning) warning.innerHTML = buildAgentChoiceWarningHTML(a);
  }

  // Flèches dans le groupe d'agents : décale la sélection de `delta` en
  // bouclant, et emmène le focus avec elle (tabindex roulant).
  function moveAgentChoice(delta) {
    var choices = Array.prototype.slice.call(document.querySelectorAll('#agent-choices .agent-choice'));
    if (choices.length < 2) return;
    var ids = choices.map(function (el) { return el.getAttribute('data-agent-id'); });
    var i = ids.indexOf(modalAgentId);
    if (i < 0) i = 0;
    var next = (i + delta + ids.length) % ids.length;
    pickAgentInModal(ids[next], true);
  }

  // mode 'chat' : la conversation s'ouvre. mode 'another' : le formulaire reste
  // ouvert, agent et dépôt conservés, titre et prompt vidés (saisie en série).
  function submitNewTask(cardId, mode) {
    var titleEl = document.getElementById('new-task-title');
    var promptEl = document.getElementById('new-task-prompt');
    var errEl = document.getElementById('new-task-error');
    var noteEl = document.getElementById('new-task-note');
    if (!titleEl || !promptEl) return;
    var title = titleEl.value.trim();
    var prompt = promptEl.value.trim();
    if (!title) { errEl.textContent = t('newTask.errorTitleRequired'); errEl.classList.remove('hidden'); titleEl.focus(); return; }
    if (!modalAgentId) { errEl.textContent = t('newTask.errorAgentRequired'); errEl.classList.remove('hidden'); return; }
    errEl.classList.add('hidden');
    var body = { title: title, agentId: modalAgentId };
    if (cardId) body.cardId = cardId;
    else body.projectId = newTaskProjectId;
    if (!cardId && newTaskSignalId) body.signalId = newTaskSignalId;
    if (prompt) body.prompt = prompt;
    if (modalAttachments.length) body.attachments = attachPayloadOf(modalAttachments);
    var repoEl = document.getElementById('new-task-repo');
    if (repoEl && repoEl.value) body.repoName = repoEl.value;
    var waitsForEl = document.getElementById('new-task-waits-for');
    if (waitsForEl && waitsForEl.value) body.waitsForTaskId = waitsForEl.value;
    api('/api/tasks', { method: 'POST', body: body }).then(function (task) {
      upsertTask(task);
      // Un seul signal par tâche : « Créer et enchaîner » repart sans lui.
      if (!cardId) newTaskSignalId = null;
      if (mode === 'another') {
        titleEl.value = '';
        promptEl.value = '';
        // Les images sont parties avec la tâche créée : la suivante repart nue.
        modalAttachments = [];
        refreshModalAttachStrip();
        if (noteEl) {
          noteEl.textContent = t('newTask.created', { ref: task.ref, title: task.title });
          noteEl.classList.remove('hidden');
        }
        titleEl.focus();
        renderMain();
        return;
      }
      closeModal();
      // Chantier implicite : la carte vient d'être créée côté serveur, et son
      // événement SSE peut arriver après cette réponse. La route vers la tâche
      // a besoin de la carte, d'où un rechargement de l'état avant d'y aller.
      if (!cardId && !state.cardsById[task.cardId]) {
        api('/api/state').then(function (data) { hydrateState(data); openTask(task.id); }).catch(function () { openTask(task.id); });
        return;
      }
      openTask(task.id);
    }).catch(function (e) {
      errEl.textContent = (e instanceof ApiError && e.message) || t('newTask.errorCreateFailed');
      errEl.classList.remove('hidden');
    });
  }

  // Dépôts de projet (lignes éditables de l'onglet Dépôts). La création de
  // projet ne passe pas par ici : elle ne demande qu'un chemin (voir
  // buildNewProjectModalHTML), le nom du dépôt étant déduit côté serveur.

  function buildRepoRowsHTML() {
    return modalRepos.map(function (r, i) {
      var canRemove = modalRepos.length > 1;
      return '<div class="repo-block">' +
        '<div class="repo-row">' +
          '<input class="modal-input repo-row-name" placeholder="' + escapeHtml(t('project.repoNamePlaceholder')) + '" value="' + escapeHtml(r.name) + '">' +
          '<input class="modal-input mono repo-row-path" placeholder="' + escapeHtml(t('newProject.pathPlaceholder')) + '" value="' + escapeHtml(r.path) + '">' +
          (canRemove ? '<button class="icon-btn repo-row-remove" data-action="remove-repo-row" data-index="' + i + '" aria-label="' + escapeHtml(t('project.removeRepo')) + '">✕</button>' : '') +
        '</div>' +
        '<div class="repo-preview-row">' +
          '<input class="modal-input mono repo-row-preview-cmd" placeholder="' + escapeHtml(t('project.previewCmdPlaceholder')) + '" value="' + escapeHtml(r.previewCmd || '') + '">' +
          '<input class="modal-input mono repo-row-preview-url" placeholder="' + escapeHtml(t('project.previewUrlPlaceholder')) + '" value="' + escapeHtml(r.previewUrl || '') + '">' +
        '</div>' +
        '</div>';
    }).join('');
  }
  // Une ligne est lue bloc par bloc (`.repo-block`), jamais champ par champ à
  // l'échelle du document : l'en-tête des colonnes porte les mêmes classes
  // `repo-row-name`/`repo-row-path` sur des <span>, et une lecture globale le
  // comptait comme une ligne, décalant d'un cran la commande et l'URL de recette
  // (la dernière ligne perdait les siennes à chaque enregistrement).
  function captureRepoRowsFromDOM() {
    var blocks = document.querySelectorAll('#repo-rows .repo-block');
    if (blocks.length === 0) return;
    modalRepos = Array.prototype.map.call(blocks, function (block) {
      function val(selector) {
        var el = block.querySelector(selector);
        return el ? el.value : '';
      }
      return {
        name: val('input.repo-row-name'), path: val('input.repo-row-path'),
        previewCmd: val('input.repo-row-preview-cmd'),
        previewUrl: val('input.repo-row-preview-url')
      };
    });
  }
  function refreshRepoRowsUI() {
    var container = document.getElementById('repo-rows');
    if (container) container.innerHTML = buildRepoRowsHTML();
  }
  function addRepoRow() {
    captureRepoRowsFromDOM();
    modalRepos.push({ name: '', path: '', previewCmd: '', previewUrl: '' });
    refreshRepoRowsUI();
  }
  function removeRepoRow(index) {
    captureRepoRowsFromDOM();
    if (modalRepos.length <= 1) return;
    modalRepos.splice(index, 1);
    refreshRepoRowsUI();
  }
  // Variables de recette : une puce par variable (nom + à quoi elle sert),
  // plus un exemple concret. Un bloc à part de `.modal-section-hint` pour
  // pouvoir mettre `$SILLAGE_PORT` en avant visuellement (code), là où la
  // phrase unique d'origine le noyait dans du texte.
  var PREVIEW_VARS = [
    ['SILLAGE_ID', 'project.previewVarIdDesc'],
    ['SILLAGE_N', 'project.previewVarNDesc'],
    ['SILLAGE_PORT', 'project.previewVarPortDesc'],
    ['SILLAGE_DIR', 'project.previewVarDirDesc'],
    ['SILLAGE_BRANCH', 'project.previewVarBranchDesc']
  ];
  function buildPreviewHintHTML() {
    var vars = PREVIEW_VARS.map(function (v) {
      return '<li class="preview-hint-var">' +
        '<code>$' + v[0] + '</code>' +
        '<span>' + escapeHtml(t(v[1])) + '</span>' +
        '</li>';
    }).join('');
    return '<div class="modal-section-hint preview-hint">' +
      '<div>' + escapeHtml(t('project.previewHintIntro')) + '</div>' +
      '<ul class="preview-hint-vars">' + vars + '</ul>' +
      '<div class="preview-hint-example">' +
        escapeHtml(t('project.previewExampleLabel')) +
        ' <code>python3 -m http.server $SILLAGE_PORT</code>' +
      '</div>' +
    '</div>';
  }
  // L'aide ne s'affiche que quand la liste est vide : une fois les chemins
  // saisis, ils se lisent seuls et la phrase n'est plus que du bruit. Les deux
  // colonnes portent alors un en-tête, sans quoi une ligne remplie ne dit plus
  // lequel des deux champs est le nom (les placeholders sont masqués).
  // La branche de base vit ici, pas dans Livraison : le serveur s'en sert aussi
  // comme point de départ des branches de chantier (voir CreateCardWorktree),
  // c'est donc un réglage des dépôts, pas un sous-réglage de la livraison.
  function buildBaseBranchFieldHTML() {
    return fieldRowHTML('project.baseBranch', 'project-delivery-target',
        '<input id="project-delivery-target" class="modal-input mono" placeholder="' + escapeHtml(t('delivery.targetPlaceholder')) + '" value="' + escapeHtml(projectDraft.target) + '">') +
      '<div class="modal-note">' + escapeHtml(t('project.baseBranchHint')) + '</div>';
  }
  function buildRepoSectionHTML() {
    var hasPath = modalRepos.some(function (r) { return (r.path || '').trim(); });
    var head = hasPath
      ? '<div class="repo-row repo-row-head">' +
          '<span class="repo-row-name">' + escapeHtml(t('project.repoName')) + '</span>' +
          '<span class="repo-row-path">' + escapeHtml(t('project.repoPath')) + '</span>' +
        '</div>'
      : '<div class="modal-section-hint">' + escapeHtml(t('project.reposHint')) + '</div>';
    // Le rappel des variables de recette est ici, à côté du champ où l'on écrit
    // la commande : c'est le seul endroit où l'on en a besoin.
    return head +
      '<div id="repo-rows">' + buildRepoRowsHTML() + '</div>' +
      '<button class="add-repo-link" data-action="add-repo-row">' + escapeHtml(t('project.addRepo')) + '</button>' +
      buildPreviewHintHTML() +
      buildBaseBranchFieldHTML();
  }
  function collectReposForSubmit() {
    captureRepoRowsFromDOM();
    return modalRepos.map(function (r) {
      return {
        name: (r.name || '').trim(), path: (r.path || '').trim(),
        previewCmd: (r.previewCmd || '').trim(), previewUrl: (r.previewUrl || '').trim()
      };
    }).filter(function (r) { return r.path; });
  }
  function reposToBody(repos) {
    return repos.map(function (r) {
      var o = { path: r.path };
      if (r.name) o.name = r.name;
      if (r.previewCmd) o.previewCmd = r.previewCmd;
      if (r.previewUrl) o.previewUrl = r.previewUrl;
      return o;
    });
  }
  // Réglage de livraison (« ce que Ship veut dire dans ce projet »), présenté en
  // quatre cartes à cocher : chaque mode porte sa propre phrase de conséquence,
  // toutes lisibles d'un coup. Un menu déroulant obligeait à changer la
  // sélection pour découvrir ce que chaque option ferait. La création de projet
  // ne pose pas la question : le serveur déduit le mode des remotes des dépôts.
  var DELIVERY_MODES = [
    { value: 'pr', labelKey: 'delivery.modePr', noteKey: 'delivery.prNote' },
    { value: 'push', labelKey: 'delivery.modePush', noteKey: 'delivery.pushNote' },
    { value: 'merge', labelKey: 'delivery.modeMerge', noteKey: 'delivery.mergeNote' },
    { value: 'merge-push', labelKey: 'delivery.modeMergePush', noteKey: 'delivery.mergePushNote' }
  ];

  // La phrase de conséquence d'un mode, partagée par les cartes et par le
  // récapitulatif de livraison (qui est la confirmation de la seule action
  // sortante du produit).
  function deliveryModeNote(mode) {
    for (var i = 0; i < DELIVERY_MODES.length; i++) {
      if (DELIVERY_MODES[i].value === mode) return t(DELIVERY_MODES[i].noteKey);
    }
    return '';
  }

  function buildDeliveryPanelHTML(project) {
    var mode = projectDraft ? projectDraft.mode : ((project.delivery && project.delivery.mode) || 'pr');
    var cards = DELIVERY_MODES.map(function (m) {
      var checked = m.value === mode;
      return '<label class="choice-card' + (checked ? ' choice-card-active' : '') + '">' +
        '<input type="radio" name="project-delivery-mode" value="' + m.value + '"' + (checked ? ' checked' : '') + '>' +
        '<span class="choice-card-text">' +
          '<span class="choice-card-title">' + escapeHtml(t(m.labelKey)) + '</span>' +
          '<span class="choice-card-desc">' + escapeHtml(t(m.noteKey)) + '</span>' +
        '</span>' +
        '</label>';
    }).join('');
    var warning = project.deliveryWarning
      ? '<div class="modal-note modal-note-warning">⚠ ' + escapeHtml(deliveryWarningText(project.deliveryWarning)) +
        deliveryWarningLinkHTML(project.deliveryWarning) + '</div>'
      : '';
    return '<div class="modal-section-hint">' + escapeHtml(t('delivery.label')) + '</div>' +
      '<div class="choice-cards" role="radiogroup">' + cards + '</div>' +
      warning;
  }

  function refreshDeliveryCards() {
    var cards = document.querySelectorAll('.choice-cards .choice-card');
    Array.prototype.forEach.call(cards, function (card) {
      var input = card.querySelector('input[type="radio"]');
      card.classList.toggle('choice-card-active', !!(input && input.checked));
    });
  }

  // Liens épinglés (modale d'édition de projet uniquement)

  function buildLinkRowsHTML() {
    if (!modalLinks.length) return '<div class="empty-note-sm">' + escapeHtml(t('project.linksEmpty')) + '</div>';
    return modalLinks.map(function (l, i) {
      var titleNote = l.title ? '<span class="link-row-title">' + escapeHtml(l.title) + '</span>' : '';
      return '<div class="link-row">' +
        '<input class="modal-input mono link-row-url" placeholder="https://…" value="' + escapeHtml(l.url) + '">' +
        titleNote +
        '<button class="icon-btn link-row-remove" data-action="remove-link-row" data-index="' + i + '" aria-label="' + escapeHtml(t('project.removeRepo')) + '">✕</button>' +
        '</div>';
    }).join('');
  }
  function captureLinksFromDOM() {
    var urlInputs = document.querySelectorAll('.link-row-url');
    if (urlInputs.length === 0) return;
    modalLinks = Array.prototype.map.call(urlInputs, function (input, i) {
      return { url: input.value, title: modalLinks[i] ? modalLinks[i].title : '' };
    });
  }
  function refreshLinkRowsUI() {
    var container = document.getElementById('link-rows');
    if (container) container.innerHTML = buildLinkRowsHTML();
  }
  function addLinkRow() {
    var input = document.getElementById('new-link-url');
    var url = input ? input.value.trim() : '';
    var errEl = document.getElementById('project-modal-error');
    if (!url) return;
    if (!/^https?:\/\//i.test(url)) {
      if (errEl) { errEl.textContent = t('project.linksInvalidUrl'); errEl.classList.remove('hidden'); }
      return;
    }
    captureLinksFromDOM();
    if (modalLinks.length >= 12) {
      if (errEl) { errEl.textContent = t('project.linksMax'); errEl.classList.remove('hidden'); }
      return;
    }
    modalLinks.push({ url: url, title: '' });
    input.value = '';
    if (errEl) errEl.classList.add('hidden');
    refreshLinkRowsUI();
  }
  function removeLinkRow(index) {
    captureLinksFromDOM();
    modalLinks.splice(index, 1);
    refreshLinkRowsUI();
  }
  function buildLinksSectionHTML() {
    return '<div class="modal-section-hint">' + escapeHtml(t('project.linksHint')) + '</div>' +
      '<div id="link-rows">' + buildLinkRowsHTML() + '</div>' +
      '<div class="link-add-row">' +
        '<input id="new-link-url" class="modal-input mono" placeholder="https://…">' +
        '<button class="add-repo-link" data-action="add-link-row">' + escapeHtml(t('project.addLink')) + '</button>' +
      '</div>';
  }
  function collectLinksForSubmit() {
    captureLinksFromDOM();
    return modalLinks.map(function (l) {
      return { url: (l.url || '').trim(), title: (l.title || '').trim() };
    }).filter(function (l) { return l.url; });
  }
  function linksToBody(links) {
    return links.map(function (l) {
      var o = { url: l.url };
      if (l.title) o.title = l.title;
      return o;
    });
  }

  // Nouveau projet : une seule question, le chemin d'un dépôt git. Le nom du
  // projet vient du basename (voir AddProject) et le mode de livraison des
  // remotes (voir detectDelivery) : rien d'autre n'est demandé à froid, tout se
  // règle ensuite dans la modale de réglages.

  function buildNewProjectModalHTML() {
    return '<div class="modal modal-sm">' +
      '<div class="modal-head"><span class="modal-title">' + escapeHtml(t('newProject.title')) + '</span><button class="icon-btn" data-action="close-modal" aria-label="' + escapeHtml(t('common.close')) + '">✕</button></div>' +
      '<label class="modal-label" for="new-project-path">' + escapeHtml(t('newProject.pathLabel')) + '</label>' +
      '<input id="new-project-path" class="modal-input mono" placeholder="' + escapeHtml(t('newProject.pathPlaceholder')) + '">' +
      '<div class="modal-note">' + escapeHtml(t('newProject.hint')) + '</div>' +
      '<div id="new-project-error" class="modal-error hidden"></div>' +
      '<div class="modal-foot"><button class="btn-outline" data-action="close-modal">' + escapeHtml(t('common.cancel')) + '</button>' +
      '<button class="btn-green" data-action="submit-new-project">' + escapeHtml(t('common.create')) + '</button></div>' +
      '</div>';
  }
  function openNewProjectModal() {
    openModal(buildNewProjectModalHTML());
    focusById('new-project-path');
  }
  function submitNewProject() {
    var errEl = document.getElementById('new-project-error');
    var path = document.getElementById('new-project-path').value.trim();
    if (!path) { errEl.textContent = t('newProject.errorPathRequired'); errEl.classList.remove('hidden'); return; }
    api('/api/projects', { method: 'POST', body: { repos: [{ path: path }] } }).then(function (project) {
      upsertProject(project);
      closeModal();
      goProject(project.id);
      // On force la modale sur Livraison : c'est le seul réglage qu'un nouveau
      // projet ne devine pas correctement tout seul (mode déduit des remotes,
      // à confirmer), et le seul qu'on ne veut pas laisser passer.
      openEditProjectModal('delivery', project.id);
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
      '<div class="modal-label">' + escapeHtml(t('project.contextPrompt')) + '</div>' +
      '<textarea id="new-card-context" class="modal-textarea" rows="3" placeholder="' + escapeHtml(t('project.contextPromptPlaceholder')) + '"></textarea>' +
      '<div id="new-card-error" class="modal-error hidden"></div>' +
      '<div class="modal-foot"><button class="btn-outline" data-action="close-modal">' + escapeHtml(t('common.cancel')) + '</button>' +
      // Planifier tout de suite est le prolongement naturel de la création :
      // le titre du chantier vient d'être écrit, c'est là qu'on sait le moins
      // par quoi commencer.
      '<button class="btn-neutral" data-action="submit-new-card-plan" title="' + escapeHtml(t('newCard.submitAndPlanTooltip')) + '">' + escapeHtml(t('newCard.submitAndPlan')) + '</button>' +
      '<button class="btn-green" data-action="submit-new-card">' + escapeHtml(t('common.create')) + '</button></div>' +
      '</div>';
  }
  function openNewCardModal() {
    if (!state.projectId) return;
    openModal(buildNewCardModalHTML());
    focusById('new-card-title');
  }
  // mode 'plan' : le chantier créé, la modale de plan prend la place de
  // celle-ci, sur la page du chantier.
  function submitNewCard(mode) {
    var titleEl = document.getElementById('new-card-title');
    var contextEl = document.getElementById('new-card-context');
    var errEl = document.getElementById('new-card-error');
    var title = titleEl.value.trim();
    var contextPrompt = contextEl.value.trim();
    if (!title) { errEl.textContent = t('newCard.errorTitleRequired'); errEl.classList.remove('hidden'); return; }
    api('/api/cards', { method: 'POST', body: { projectId: state.projectId, title: title, contextPrompt: contextPrompt } }).then(function (card) {
      upsertCard(card);
      closeModal();
      goCard(card.id);
      if (mode === 'plan') openPlanModal(card.id);
    }).catch(function (e) {
      errEl.textContent = (e instanceof ApiError && e.message) || t('newCard.errorCreateFailed');
      errEl.classList.remove('hidden');
    });
  }

  // Plan de tâches : demander à une IA de découper le chantier en une
  // succession d'étapes. La proposition n'est jamais appliquée toute seule,
  // c'est un brouillon relu et corrigé dans la modale avant que la moindre
  // tâche existe (voir docs/SPEC-API.md, « Plan de tâches »).

  // Les CLI dont Sillage garantit la lecture seule pendant une planification
  // (canPlan dans internal/server/plan.go) : la route refuse les autres, la
  // liste ne les propose donc pas.
  var PLAN_CAPABLE_CLIS = ['claude', 'codex', 'fake'];

  function planCapableAgents() {
    return state.agents.filter(function (a) { return PLAN_CAPABLE_CLIS.indexOf(a.cli) >= 0; });
  }

  function buildPlanFormHTML(card) {
    var agents = planCapableAgents();
    if (!agents.length) {
      return '<div class="modal-error">' + escapeHtml(t('plan.errorNoAgent')) + '</div>' +
        '<div class="modal-foot"><button class="btn-outline" data-action="close-modal">' + escapeHtml(t('common.close')) + '</button></div>';
    }
    var project = state.projectsById[card.projectId];
    var repos = (project && project.repos) || [];
    var repoSelectHTML = '';
    if (repos.length > 1) {
      var repoOptions = repos.map(function (r) {
        return '<option value="' + escapeHtml(r.name) + '"' + (r.name === planRepoName ? ' selected' : '') + '>' + escapeHtml(r.name) + '</option>';
      }).join('');
      repoSelectHTML = '<div class="modal-label">' + escapeHtml(t('plan.repoLabel')) + '</div>' +
        '<select id="plan-repo" class="modal-input">' + repoOptions + '</select>';
    }
    return '<div class="modal-hint">' + escapeHtml(t('plan.intro')) + '</div>' +
      '<div class="modal-label">' + escapeHtml(t('plan.extraLabel')) + '</div>' +
      '<textarea id="plan-extra" class="modal-textarea" rows="3" placeholder="' + escapeHtml(t('plan.extraPlaceholder')) + '"' + (planBusy ? ' disabled' : '') + '>' + escapeHtml(planExtra) + '</textarea>' +
      repoSelectHTML +
      '<div class="modal-label modal-label-row"><span>' + escapeHtml(t('plan.agentLabel')) + '</span>' +
      '<span class="modal-label-hint">' + escapeHtml(t('plan.agentHint')) + '</span></div>' +
      '<div class="agent-choices" id="agent-choices" role="radiogroup" aria-label="' + escapeHtml(t('plan.agentLabel')) + '">' +
      buildAgentChoicesHTML(agents) + '</div>' +
      '<div id="agent-choice-warning">' + buildAgentChoiceWarningHTML(state.agentsById[modalAgentId]) + '</div>' +
      '<div id="plan-error" class="modal-error hidden"></div>' +
      buildPlanThinkingHTML() +
      '<div class="modal-foot"><button class="btn-outline" data-action="close-modal">' + escapeHtml(t('common.cancel')) + '</button>' +
      '<button class="btn-green" data-action="submit-plan"' + (planBusy ? ' disabled' : '') + '>' +
      (planBusy ? planPendingLabelHTML(t('plan.proposing')) : escapeHtml(t('plan.propose'))) + '</button></div>';
  }

  // L'attente d'un plan est la même attente que celle d'un agent dans une
  // conversation, et se montre donc de la même façon (voir buildThinkingHTML) :
  // l'avatar de l'agent qui respire, son nom, les trois points, et une ligne
  // qui ondule sous lui. Elle prend la place que la proposition occupera, pour
  // que rien ne saute quand elle arrive.
  // Le compteur est ce qui distingue vraiment « ça travaille » de « c'est
  // figé » : lire un dépôt prend parfois deux minutes, et une animation seule
  // ne dit pas si elle tourne depuis cinq secondes ou depuis cinq minutes. Il
  // survit à prefers-reduced-motion, qui éteint tout le reste.
  function buildPlanThinkingHTML() {
    if (!planBusy) return '';
    var agent = state.agentsById[modalAgentId] || { emoji: '', name: '', color: '#ccc' };
    return '<div class="msg-thinking plan-thinking" id="plan-thinking">' +
      '<span class="msg-avatar" style="background:' + softColor(agent.color) + '">' + (agent.emoji || '') + '</span>' +
      '<div class="msg-body">' +
        '<div class="msg-head"><span class="msg-name">' + escapeHtml(agent.name || '') + '</span>' +
          '<span class="thinking-dots" aria-hidden="true"><i></i><i></i><i></i></span></div>' +
        '<div class="thinking-line" aria-live="polite">' +
          '<span class="thinking-label">' + escapeHtml(t(planWaitKey)) + '</span>' +
          '<span class="thinking-elapsed mono">' + escapeHtml(planElapsedText()) + '</span>' +
        '</div>' +
        // L'ordre de grandeur annoncé une fois vaut mieux qu'une barre de
        // progression inventée : on ne sait pas où en est l'agent.
        (planWaitKey === 'plan.proposing' ? '<div class="plan-thinking-note">' + escapeHtml(t('plan.waitNote')) + '</div>' : '') +
      '</div></div>';
  }

  // Trois points dans le bouton lui-même : un bouton grisé au libellé fixe est
  // le premier endroit où l'on croit l'interface bloquée, parce que c'est là
  // qu'on vient de cliquer.
  function planPendingLabelHTML(label) {
    return escapeHtml(label) + '<span class="thinking-dots btn-dots" aria-hidden="true"><i></i><i></i><i></i></span>';
  }

  function planElapsedText() {
    if (!planStartedAt) return '';
    var s = Math.max(0, Math.round((Date.now() - planStartedAt) / 1000));
    var m = Math.floor(s / 60);
    var rest = s % 60;
    return m + ':' + (rest < 10 ? '0' : '') + rest;
  }

  // Seul le texte du compteur est réécrit, jamais la ligne entière : réinjecter
  // le libellé relancerait son ondulation à zéro chaque seconde.
  function startPlanTicker(waitKey) {
    planWaitKey = waitKey;
    planStartedAt = Date.now();
    stopPlanTicker();
    planTicker = setInterval(function () {
      var el = document.querySelector('#plan-thinking .thinking-elapsed');
      if (!el) { stopPlanTicker(); return; }
      el.textContent = planElapsedText();
    }, 1000);
  }

  function stopPlanTicker() {
    if (planTicker) { clearInterval(planTicker); planTicker = null; }
  }

  // Appelé par closeModal : une modale fermée ne laisse ni minuteur qui tourne,
  // ni réponse en vol capable de la rouvrir.
  function resetPlanWait() {
    stopPlanTicker();
    if (planBusy) { planBusy = false; planRunToken += 1; }
  }

  function buildPlanProposalHTML() {
    var stepsHTML = planDraft.steps.map(function (s, i) {
      return '<div class="plan-step">' +
        '<div class="plan-step-head">' +
          '<span class="plan-step-num">' + (i + 1) + '</span>' +
          '<input class="modal-input plan-step-title" data-index="' + i + '" value="' + escapeHtml(s.title) + '" placeholder="' + escapeHtml(t('plan.stepTitlePlaceholder')) + '">' +
          '<button class="icon-btn" data-action="plan-remove-step" data-index="' + i + '" title="' + escapeHtml(t('plan.removeStep')) + '" aria-label="' + escapeHtml(t('plan.removeStep')) + '">✕</button>' +
        '</div>' +
        '<textarea class="modal-textarea plan-step-prompt" data-index="' + i + '" rows="3" placeholder="' + escapeHtml(t('plan.stepPromptPlaceholder')) + '">' + escapeHtml(s.prompt) + '</textarea>' +
        '</div>';
    }).join('');
    return '<div class="modal-label">' + escapeHtml(tCount('plan.stepsTitle', planDraft.steps.length)) + '</div>' +
      '<div class="plan-steps">' + stepsHTML + '</div>' +
      '<label class="plan-chain-row">' +
        '<input type="checkbox" id="plan-chain"' + (planDraft.chain ? ' checked' : '') + '>' +
        '<span>' + escapeHtml(t('plan.chainLabel')) + '</span>' +
      '</label>' +
      '<div class="modal-note">' + escapeHtml(t('plan.chainHint')) + '</div>' +
      '<div id="plan-error" class="modal-error hidden"></div>' +
      buildPlanThinkingHTML() +
      '<div class="modal-foot"><button class="btn-outline" data-action="close-modal">' + escapeHtml(t('common.cancel')) + '</button>' +
      '<button class="btn-neutral" data-action="plan-restart"' + (planBusy ? ' disabled' : '') + '>' + escapeHtml(t('plan.again')) + '</button>' +
      '<button class="btn-green" data-action="plan-create-tasks"' + (planBusy ? ' disabled' : '') + '>' +
      (planBusy ? planPendingLabelHTML(t('plan.creating')) : escapeHtml(tCount('plan.create', planDraft.steps.length))) + '</button></div>';
  }

  function buildPlanModalHTML(card) {
    return '<div class="modal">' +
      '<div class="modal-head"><span class="modal-title">' + escapeHtml(t('plan.title')) + '</span>' +
      '<span class="modal-sub">' + escapeHtml(card.title) + '</span>' +
      '<button class="icon-btn" data-action="close-modal" aria-label="' + escapeHtml(t('common.close')) + '">✕</button></div>' +
      (planDraft ? buildPlanProposalHTML() : buildPlanFormHTML(card)) +
      '</div>';
  }

  function renderPlanModal() {
    var card = state.cardsById[planCardId];
    if (!card) { closeModal(); return; }
    openModal(buildPlanModalHTML(card));
  }

  function openPlanModal(cardId) {
    var card = state.cardsById[cardId];
    if (!card) return;
    planCardId = cardId;
    planDraft = null;
    planBusy = false;
    planExtra = '';
    planRepoName = '';
    stopPlanTicker();
    // L'agent déjà choisi ailleurs est repris s'il sait planifier, sinon le
    // premier agent en bon état de la liste.
    var agents = planCapableAgents();
    var kept = agents.filter(function (a) { return a.id === modalAgentId; })[0];
    var fallback = agents.filter(function (a) { return !a.warning; })[0] || agents[0];
    modalAgentId = (kept || fallback || {}).id || null;
    renderPlanModal();
  }

  function planErrorText(e, fallbackKey) {
    return (e instanceof ApiError && e.message) || t(fallbackKey);
  }

  function showPlanError(text) {
    var el = document.getElementById('plan-error');
    if (!el) return;
    el.textContent = text;
    el.classList.remove('hidden');
  }

  function submitPlan() {
    if (planBusy) return;
    var card = state.cardsById[planCardId];
    if (!card) return;
    if (!modalAgentId) { showPlanError(t('plan.errorAgentRequired')); return; }
    var extraEl = document.getElementById('plan-extra');
    var repoEl = document.getElementById('plan-repo');
    planExtra = extraEl ? extraEl.value.trim() : '';
    planRepoName = repoEl ? repoEl.value : '';
    var body = { agentId: modalAgentId };
    if (planExtra) body.prompt = planExtra;
    if (planRepoName) body.repoName = planRepoName;
    planBusy = true;
    var token = ++planRunToken;
    startPlanTicker('plan.proposing');
    renderPlanModal();
    api('/api/cards/' + planCardId + '/plan', { method: 'POST', body: body }).then(function (plan) {
      if (token !== planRunToken) return; // modale fermée entre-temps
      planBusy = false;
      stopPlanTicker();
      planDraft = { steps: plan.steps || [], repoName: plan.repoName || '', chain: true };
      renderPlanModal();
    }).catch(function (e) {
      if (token !== planRunToken) return;
      planBusy = false;
      stopPlanTicker();
      renderPlanModal();
      showPlanError(planErrorText(e, 'plan.errorFailed'));
    });
  }

  // Ce qui a été corrigé à la main dans la modale est relu avant tout
  // re-rendu : retirer une étape réécrit la liste entière.
  function syncPlanDraftFromDOM() {
    if (!planDraft) return;
    document.querySelectorAll('.plan-step-title').forEach(function (el) {
      var s = planDraft.steps[parseInt(el.getAttribute('data-index'), 10)];
      if (s) s.title = el.value.trim();
    });
    document.querySelectorAll('.plan-step-prompt').forEach(function (el) {
      var s = planDraft.steps[parseInt(el.getAttribute('data-index'), 10)];
      if (s) s.prompt = el.value.trim();
    });
    var chainEl = document.getElementById('plan-chain');
    if (chainEl) planDraft.chain = chainEl.checked;
  }

  function removePlanStep(index) {
    if (!planDraft || planBusy) return;
    syncPlanDraftFromDOM();
    planDraft.steps.splice(index, 1);
    if (!planDraft.steps.length) { planDraft = null; }
    renderPlanModal();
  }

  function restartPlan() {
    if (planBusy) return;
    planDraft = null;
    renderPlanModal();
  }

  // Les tâches sont créées une par une, dans l'ordre : chaînées, chacune
  // attend l'acceptation de la précédente, donc seule la première démarre un
  // agent. Une erreur en cours de route garde les étapes restantes dans la
  // modale plutôt que de tout perdre.
  function createPlanTasks() {
    if (!planDraft || planBusy) return;
    syncPlanDraftFromDOM();
    var steps = planDraft.steps.filter(function (s) { return s.title; });
    if (!steps.length) { showPlanError(t('plan.errorEmpty')); return; }
    var chain = planDraft.chain;
    var repoName = planDraft.repoName;
    planBusy = true;
    var token = ++planRunToken;
    // Chaque tâche crée un worktree git : trois étapes se comptent en secondes,
    // pas en millisecondes. Même attente, donc même indicateur.
    startPlanTicker('plan.creating');
    renderPlanModal();

    var i = 0;
    var previousId = null;
    function createNext() {
      if (i >= steps.length) return Promise.resolve();
      var body = { cardId: planCardId, title: steps[i].title, agentId: modalAgentId };
      if (steps[i].prompt) body.prompt = steps[i].prompt;
      if (repoName) body.repoName = repoName;
      if (chain && previousId) body.waitsForTaskId = previousId;
      return api('/api/tasks', { method: 'POST', body: body }).then(function (task) {
        upsertTask(task);
        previousId = task.id;
        i += 1;
        return createNext();
      });
    }
    createNext().then(function () {
      if (token !== planRunToken) return;
      planBusy = false;
      stopPlanTicker();
      planDraft = null;
      closeModal();
      renderSidebar();
      renderMain();
    }).catch(function (e) {
      if (token !== planRunToken) return;
      planBusy = false;
      stopPlanTicker();
      planDraft.steps = steps.slice(i);
      renderPlanModal();
      showPlanError(t('plan.errorPartial', { n: i, msg: planErrorText(e, 'plan.errorFailed') }));
      renderSidebar();
    });
  }

  // Import d'une issue GitHub : choisir une issue ouverte d'un dépôt du projet
  // pour remplir le formulaire de création de tâche. Même principe que le plan
  // de tâches, et même raison : la forge propose un titre et un texte, l'humain
  // les relit et les corrige avant qu'un agent démarre. Rien n'est écrit sur
  // GitHub, ni ici ni plus tard (voir internal/server/issues.go).

  // Plafond de la liste, à garder aligné sur issueListLimit
  // (internal/server/issues.go) : c'est lui qui décide de l'affichage de la
  // note « les N plus récentes ».
  var ISSUE_LIST_LIMIT = 50;

  // Les dépôts du projet qui sont sur github.com, calculés côté serveur depuis
  // le remote origin (ProjectOut.GithubRepos).
  function issueRepoChoices(card) {
    var project = state.projectsById[card.projectId];
    return (project && project.githubRepos) || [];
  }

  function openIssuePicker(cardId) {
    var card = newTaskCard(cardId);
    if (!card) return;
    var repos = issueRepoChoices(card);
    if (!repos.length) return;
    issueCardId = cardId;
    // Le dépôt déjà choisi dans le formulaire s'il est sur GitHub, sinon le
    // premier : chercher une issue ailleurs que là où la tâche travaillera
    // n'aurait pas de sens.
    var wanted = (newTaskDraft && newTaskDraft.repoName) || '';
    issueRepoName = repos.indexOf(wanted) >= 0 ? wanted : repos[0];
    issueQuery = '';
    issueList = null;
    issueError = '';
    issueBusy = false;
    renderIssuePicker();
    loadIssues();
  }

  // Appelé par closeModal : ni recherche en attente, ni réponse en vol capable
  // de réécrire une liste qui n'est plus là.
  function resetIssuePick() {
    if (issueSearchTimer) { clearTimeout(issueSearchTimer); issueSearchTimer = null; }
    if (issueBusy) { issueBusy = false; issueRunToken += 1; }
  }

  function renderIssuePicker() {
    var card = newTaskCard(issueCardId);
    if (!card) { closeModal(); return; }
    openModal(buildIssuePickerHTML(card));
    setTimeout(function () {
      var el = document.getElementById('issue-search');
      if (!el) return;
      el.focus();
      if (el.value) el.setSelectionRange(el.value.length, el.value.length);
    }, 0);
  }

  function buildIssuePickerHTML(card) {
    var repos = issueRepoChoices(card);
    var repoSelectHTML = '';
    if (repos.length > 1) {
      var repoOptions = repos.map(function (name) {
        return '<option value="' + escapeHtml(name) + '"' + (name === issueRepoName ? ' selected' : '') + '>' + escapeHtml(name) + '</option>';
      }).join('');
      repoSelectHTML = '<div class="modal-label">' + escapeHtml(t('issues.repoLabel')) + '</div>' +
        '<select id="issue-repo" class="modal-input">' + repoOptions + '</select>';
    }
    return '<div class="modal">' +
      '<div class="modal-head"><span class="modal-title">' + escapeHtml(t('issues.title')) + '</span>' +
      '<span class="modal-sub">' + escapeHtml(card.title) + '</span>' +
      '<button class="icon-btn" data-action="close-modal" aria-label="' + escapeHtml(t('common.close')) + '">✕</button></div>' +
      // La recherche part sur la forge, pas sur la liste déjà chargée : un
      // dépôt actif a plus d'issues ouvertes qu'on n'en affiche.
      '<input id="issue-search" class="modal-input" placeholder="' + escapeHtml(t('issues.searchPlaceholder')) + '" value="' + escapeHtml(issueQuery) + '">' +
      repoSelectHTML +
      '<div class="issue-list" id="issue-list">' + buildIssueListHTML() + '</div>' +
      '<div class="modal-foot"><button class="btn-outline" data-action="cancel-issue-picker">' + escapeHtml(t('common.cancel')) + '</button></div>' +
      '</div>';
  }

  // Seule cette portion est réécrite à chaque recherche : réinjecter la modale
  // entière volerait le focus et le curseur du champ en train d'être tapé.
  function buildIssueListHTML() {
    if (issueError) return '<div class="modal-error">' + escapeHtml(issueError) + '</div>';
    if (issueList === null) return '<div class="issue-empty">' + escapeHtml(t('issues.loading')) + '</div>';
    if (!issueList.length) {
      return '<div class="issue-empty">' + escapeHtml(t(issueQuery ? 'issues.emptySearch' : 'issues.empty')) + '</div>';
    }
    var rowsHTML = issueList.map(buildIssueRowHTML).join('');
    // La liste est plafonnée côté serveur : le dire, plutôt que de laisser
    // croire que ce sont là toutes les issues ouvertes du dépôt.
    var capHTML = issueList.length >= ISSUE_LIST_LIMIT
      ? '<div class="modal-note">' + escapeHtml(t('issues.cap', { n: issueList.length })) + '</div>'
      : '';
    return rowsHTML + capHTML;
  }

  function buildIssueRowHTML(issue) {
    var labelsHTML = (issue.labels || []).map(function (l) {
      return '<span class="issue-label">' + escapeHtml(l) + '</span>';
    }).join('');
    var meta = [];
    if (issue.author) meta.push(t('issues.by', { login: issue.author }));
    if (issue.updatedAt) meta.push(timeAgo(issue.updatedAt));
    return '<button class="issue-row" data-action="pick-issue" data-number="' + issue.number + '">' +
      '<span class="issue-row-num mono">#' + issue.number + '</span>' +
      '<span class="issue-row-body">' +
        '<span class="issue-row-title">' + escapeHtml(issue.title) + '</span>' +
        (labelsHTML || meta.length
          ? '<span class="issue-row-meta">' + labelsHTML +
            (meta.length ? '<span class="issue-row-when">' + escapeHtml(meta.join(' · ')) + '</span>' : '') +
            '</span>'
          : '') +
      '</span></button>';
  }

  function refreshIssueList() {
    var el = document.getElementById('issue-list');
    if (!el) return;
    el.innerHTML = buildIssueListHTML();
    // Liste précédente conservée mais estompée pendant une nouvelle recherche :
    // la vider à chaque frappe ferait clignoter la modale.
    el.classList.toggle('issue-list-busy', issueBusy && issueList !== null);
  }

  function loadIssues() {
    if (!issueCardId) return;
    issueError = '';
    issueBusy = true;
    var token = ++issueRunToken;
    refreshIssueList();
    var url = '/api/projects/' + newTaskCard(issueCardId).projectId + '/issues' +
      '?repoName=' + encodeURIComponent(issueRepoName) + '&q=' + encodeURIComponent(issueQuery);
    api(url).then(function (res) {
      if (token !== issueRunToken) return; // recherche dépassée, ou modale fermée
      issueBusy = false;
      issueRepoName = res.repoName || issueRepoName;
      issueList = res.issues || [];
      refreshIssueList();
    }).catch(function (e) {
      if (token !== issueRunToken) return;
      issueBusy = false;
      issueList = [];
      issueError = (e instanceof ApiError && e.message) || t('issues.errorFailed');
      refreshIssueList();
    });
  }

  // Une frappe ne part pas tout de suite sur le réseau : gh interroge l'API
  // GitHub, une requête par caractère serait une requête pour rien.
  function scheduleIssueSearch(value) {
    issueQuery = value;
    if (issueSearchTimer) clearTimeout(issueSearchTimer);
    issueSearchTimer = setTimeout(function () { issueSearchTimer = null; loadIssues(); }, 350);
  }

  function searchIssuesNow(value) {
    if (issueSearchTimer) { clearTimeout(issueSearchTimer); issueSearchTimer = null; }
    issueQuery = value;
    loadIssues();
  }

  function changeIssueRepo(name) {
    if (name === issueRepoName) return;
    issueRepoName = name;
    issueList = null;
    loadIssues();
  }

  // L'issue choisie remplit le formulaire, elle ne crée rien : le titre, la
  // consigne (référence, lien, corps de l'issue) et le dépôt visé. Tout reste
  // modifiable, et aucun agent n'a encore bougé.
  function pickIssue(number) {
    var issue = (issueList || []).filter(function (i) { return i.number === number; })[0];
    if (!issue) return;
    var cardId = issueCardId;
    var parts = [t('issues.promptHeader', { n: issue.number, title: issue.title })];
    if (issue.url) parts.push(issue.url);
    if (issue.body) parts.push('', issue.body);
    newTaskDraft = {
      title: issue.title,
      prompt: parts.join('\n'),
      repoName: issueRepoName,
      waitsForTaskId: (newTaskDraft && newTaskDraft.waitsForTaskId) || ''
    };
    closeModal();
    renderNewTaskModal(cardId, 'new-task-prompt');
  }

  // Renoncer au choix revient au formulaire, pas au kanban : la saisie qui
  // l'attendait est toujours là (voir captureNewTaskDraft).
  function cancelIssuePicker() {
    var cardId = issueCardId;
    closeModal();
    renderNewTaskModal(cardId, 'new-task-title');
  }

  // Édition d'un chantier (titre + contexte)

  function buildCardEditModalHTML(card) {
    var deleteKey = 'card-delete:' + card.id;
    var deletePending = isPendingConfirm(deleteKey);
    var deleteLabel = deletePending ? t('workstream.deleteConfirm') : t('workstream.delete');
    var taskCount = card.tasksTotal || 0;
    return '<div class="modal modal-sm">' +
      '<div class="modal-head"><span class="modal-title">' + escapeHtml(t('workstream.editTitle')) + '</span><button class="icon-btn" data-action="close-modal" aria-label="' + escapeHtml(t('common.close')) + '">✕</button></div>' +
      '<div class="modal-label">' + escapeHtml(t('newCard.titleLabel')) + '</div><input id="card-edit-title" class="modal-input" value="' + escapeHtml(card.title) + '">' +
      '<div class="modal-label">' + escapeHtml(t('project.contextPrompt')) + '</div>' +
      '<textarea id="card-edit-context" class="modal-textarea" rows="3" placeholder="' + escapeHtml(t('project.contextPromptPlaceholder')) + '">' + escapeHtml(card.contextPrompt || '') + '</textarea>' +
      '<div id="card-edit-error" class="modal-error hidden"></div>' +
      '<div class="modal-delete-row">' +
        '<button class="delete-link" data-action="confirm-click" data-confirm-key="' + deleteKey + '" data-confirm-action="card-delete" data-confirm-id="' + card.id + '" data-default-label="' + escapeHtml(t('workstream.delete')) + '" data-confirm-label="' + escapeHtml(t('workstream.deleteConfirm')) + '">' + escapeHtml(deleteLabel) + '</button>' +
        '<div class="modal-delete-subtext">' + escapeHtml(tCount('workstream.deleteSubtext', taskCount)) + '</div>' +
      '</div>' +
      '<div class="modal-foot"><button class="btn-outline" data-action="close-modal">' + escapeHtml(t('common.cancel')) + '</button>' +
      '<button class="btn-green" data-action="submit-card-edit" data-card-id="' + card.id + '">' + escapeHtml(t('common.save')) + '</button></div>' +
      '</div>';
  }
  function openEditCardModal() {
    var card = state.cardsById[state.cardId];
    if (!card) return;
    openModal(buildCardEditModalHTML(card));
    focusById('card-edit-title');
  }
  function submitCardEdit(cardId) {
    var title = document.getElementById('card-edit-title').value.trim();
    var contextPrompt = document.getElementById('card-edit-context').value.trim();
    var errEl = document.getElementById('card-edit-error');
    if (!title) { errEl.textContent = t('newCard.errorTitleRequired'); errEl.classList.remove('hidden'); return; }
    api('/api/cards/' + cardId, { method: 'PATCH', body: { title: title, contextPrompt: contextPrompt } }).then(function (card) {
      upsertCard(card);
      closeModal();
      renderSidebar();
      renderMain();
    }).catch(function (e) {
      errEl.textContent = (e instanceof ApiError && e.message) || t('workstream.errorSaveFailed');
      errEl.classList.remove('hidden');
    });
  }
  function doDeleteCard(cardId) {
    var card = state.cardsById[cardId];
    var projectId = card ? card.projectId : state.projectId;
    api('/api/cards/' + cardId, { method: 'DELETE', body: { confirm: true } }).then(function () {
      removeCardLocally(cardId);
      closeModal();
      goProject(projectId);
    }).catch(function (e) {
      var errEl = document.getElementById('card-edit-error');
      if (errEl) {
        errEl.textContent = (e instanceof ApiError && e.message) || t('errors.deleteCardFailed');
        errEl.classList.remove('hidden');
      }
    });
  }

  // Nouvel agent / édition d'agent

  // Suggestions sobres et variées pour l'emoji d'agent ; la saisie libre
  // reste possible (datalist, pas de select fermé).
  var AGENT_EMOJI_SUGGESTIONS = ['🐝', '🦉', '🦊', '🪶', '🐙', '🦫', '🐢', '🦅', '🐺', '🦭', '🐬', '🦋', '🐞', '🪲', '🐌', '🦔', '🐿️', '🦜', '🦢', '🐋', '⚡', '🔧', '🎨', '📐'];
  function buildAgentEmojiDatalistHTML() {
    var options = AGENT_EMOJI_SUGGESTIONS.map(function (e) { return '<option value="' + e + '"></option>'; }).join('');
    return '<datalist id="agent-emoji-options">' + options + '</datalist>';
  }

  // buildAgentQuotaHTML : quota du fournisseur cli de l'agent (voir
  // AgentOut.quota côté serveur). Seul codex publie cette info ; claude et
  // fake affichent un message "non disponible" plutôt qu'une donnée inventée.
  function buildAgentQuotaHTML(agent) {
    var windowLabel = function (w) {
      if (w.label === '5h') return t('agent.quotaWindow5h');
      if (w.label === 'week') return t('agent.quotaWindowWeek');
      return w.label;
    };
    var body;
    if (agent.quota && agent.quota.windows && agent.quota.windows.length) {
      var rows = agent.quota.windows.map(function (w) {
        return '<div class="agent-quota-row">' +
          '<span class="agent-quota-window">' + escapeHtml(windowLabel(w)) + '</span>' +
          '<span class="agent-quota-percent">' + escapeHtml(t('agent.quotaUsedPercent', { percent: Math.round(w.usedPercent) })) + '</span>' +
          '<span class="agent-quota-resets">' + escapeHtml(t('agent.quotaResetsIn', { time: formatResetCountdown(w.resetsAt) })) + '</span>' +
          '</div>';
      }).join('');
      body = '<div class="agent-quota-box">' + rows + '</div>' +
        '<div class="modal-note">' + escapeHtml(t('agent.quotaUpdatedAt', { time: timeAgo(agent.quota.updatedAt) })) + '</div>';
    } else {
      body = '<div class="modal-note">' + escapeHtml(t('agent.quotaUnavailable', { cli: agent.cli })) + '</div>';
    }
    return '<div class="modal-label">' + escapeHtml(t('agent.quotaTitle')) + '</div>' + body;
  }

  function buildAgentModalHTML(agent) {
    var isEdit = !!agent;
    var title = isEdit ? t('agent.editTitle') : t('agent.newTitle');
    var cliValues = ['claude', 'codex', 'copilot', 'agy', 'kiro', 'fake'];
    var cliOptions = cliValues.map(function (cli) {
      return '<option value="' + cli + '"' + (agent && agent.cli === cli ? ' selected' : '') + '>' + escapeHtml(t('agent.cli.' + cli)) + '</option>';
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
    var warningExtras = agent ? agentWarningExtrasHTML(agent.warning, agent.id) : '';
    var warningBanner = (agent && agent.warning) ? '<div class="agent-warning-banner">⚠ ' + escapeHtml(agentWarningText(agent.warning)) + warningExtras + '</div>' : '';
    var quotaSection = isEdit ? buildAgentQuotaHTML(agent) : '';
    return '<div class="modal">' +
      '<div class="modal-head"><span class="modal-title">' + escapeHtml(title) + '</span><button class="icon-btn" data-action="close-modal" aria-label="' + escapeHtml(t('common.close')) + '">✕</button></div>' +
      warningBanner +
      quotaSection +
      '<div class="modal-label">' + escapeHtml(t('agent.name')) + '</div><input id="agent-name" class="modal-input" value="' + (agent ? escapeHtml(agent.name) : '') + '">' +
      '<div class="agent-form-row">' +
        '<div><div class="modal-label">' + escapeHtml(t('agent.emoji')) + '</div><input id="agent-emoji" class="modal-input agent-emoji-input" maxlength="4" list="agent-emoji-options" value="' + (agent ? escapeHtml(agent.emoji || '') : '') + '">' + buildAgentEmojiDatalistHTML() + '</div>' +
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
    focusById('agent-name');
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
  // Correction de la configuration machine d'un agent (aujourd'hui : la
  // politique d'exécution d'agy). L'avertissement disparaît sur place : la
  // bannière devient sa propre confirmation, sans re-rendu de la modale ouverte
  // (elle peut être celle de l'agent comme le choix d'agent d'une tâche).
  function doFixAgentWarning(agentId) {
    api('/api/agents/' + agentId + '/fix-warning', { method: 'POST' }).then(function (agent) {
      var i = state.agents.findIndex(function (a) { return a.id === agent.id; });
      if (i >= 0) state.agents[i] = agent;
      reindex();
      renderSidebar();
      var banners = document.querySelectorAll('.agent-warning-banner, .agent-choice-warning');
      for (var b = 0; b < banners.length; b++) {
        banners[b].textContent = '✓ ' + t('agent.warning.agyPolicyFixed');
      }
    }).catch(function (e) {
      // Un échec ici (settings.json illisible, droits) doit se voir là où on a
      // cliqué : la modale d'agent a son emplacement d'erreur, le choix d'agent
      // d'une nouvelle tâche non.
      var msg = (e instanceof ApiError && e.message) || t('agent.errorSaveFailed');
      showAgentModalError(msg);
      var actions = document.querySelectorAll('.agent-warning-action');
      for (var a = 0; a < actions.length; a++) {
        var line = actions[a].querySelector('.agent-warning-error');
        if (!line) {
          line = document.createElement('div');
          line.className = 'agent-warning-fallback agent-warning-error';
          actions[a].appendChild(line);
        }
        line.textContent = msg;
      }
      patchConfirmButtons('agent-fix:' + agentId);
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


  // ---------------------------------------------------------------------
  // Sessions Claude Code hors Sillage (voir claude_sessions.go)
  // ---------------------------------------------------------------------
  // Lues à la demande, pas poussées par SSE : c'est un dossier du disque que
  // Sillage ne surveille pas. Rafraîchies toutes les 30 s tant que le projet
  // est à l'écran, pour que « en cours » reste vrai.

  var CLAUDE_SESSIONS_TTL = 30000;

  function fetchClaudeSessions(projectId, force) {
    var cur = state.claudeSessions[projectId];
    if (!force && cur && (cur.loading || Date.now() - cur.at < CLAUDE_SESSIONS_TTL)) return;
    state.claudeSessions[projectId] = Object.assign({ list: [] }, cur || {}, { loading: true });
    api('/api/projects/' + encodeURIComponent(projectId) + '/claude-sessions').then(function (res) {
      state.claudeSessions[projectId] = { at: Date.now(), list: (res && res.sessions) || [] };
      refreshClaudeSessionsUI(projectId);
    }).catch(function () {
      state.claudeSessions[projectId] = { at: Date.now(), list: (cur && cur.list) || [] };
    });
  }

  function claudeSessionList(projectId) {
    var cur = state.claudeSessions[projectId];
    return (cur && cur.list) || [];
  }

  function buildSessionsPillHTML(project) {
    var list = claudeSessionList(project.id);
    var attr = ' data-sessions-pill="' + project.id + '"';
    if (!list.length) return '<span' + attr + ' hidden></span>';
    var active = list.filter(function (x) { return x.active; }).length;
    return '<button class="autoconf-pill sessions-pill' + (active ? ' sessions-pill-active' : '') + '"' + attr +
      ' data-action="open-claude-sessions" data-project-id="' + project.id + '">' +
      (active ? '<span class="sessions-dot"></span>' + escapeHtml(t('sessions.pill.active')) : escapeHtml(tCount('sessions.pill.recent', list.length))) +
      '</button>';
  }

  function buildClaudeSessionsInnerHTML(projectId) {
    var list = claudeSessionList(projectId);
    if (!list.length) return '<div class="modal-note">' + escapeHtml(t('sessions.empty')) + '</div>';
    return '<div class="sessions-list">' + list.map(function (x) {
      var meta = [x.repoName];
      if (x.branch) meta.push(x.branch);
      meta.push(timeAgo(x.updatedAt));
      meta.push(tCount('sessions.prompts', x.promptsCount));
      var prompts = (x.prompts || []);
      return '<article class="session-row">' +
        '<div class="session-head"><span class="session-title">' + escapeHtml(x.title) + '</span>' +
          (x.active ? '<span class="session-active"><span class="sessions-dot"></span>' + escapeHtml(t('sessions.active')) + '</span>' : '') + '</div>' +
        '<div class="session-meta">' + escapeHtml(meta.join(' · ')) + '</div>' +
        (x.lastPrompt ? '<div class="session-last">' + escapeHtml(x.lastPrompt) + '</div>' : '') +
        (prompts.length > 1 ? '<details class="session-prompts"><summary>' + escapeHtml(t('sessions.allPrompts')) + '</summary><ol>' +
          prompts.map(function (pr) { return '<li>' + escapeHtml(pr) + '</li>'; }).join('') + '</ol></details>' : '') +
        '<div class="session-actions">' +
          '<button class="btn-outline" data-action="claude-session-copy" data-project-id="' + projectId + '" data-session-id="' + escapeHtml(x.id) + '">' + escapeHtml(t('sessions.copy')) + '</button>' +
          '<button class="btn-quiet" data-action="claude-session-workstream" data-project-id="' + projectId + '" data-session-id="' + escapeHtml(x.id) + '">' + escapeHtml(t('sessions.workstream')) + '</button>' +
        '</div>' +
      '</article>';
    }).join('') + '</div>';
  }

  function openClaudeSessionsModal(projectId) {
    openModal('<div class="modal">' +
      '<div class="modal-head"><span class="modal-title">' + escapeHtml(t('sessions.title')) + '</span><button class="icon-btn" data-action="close-modal" aria-label="' + escapeHtml(t('common.close')) + '">✕</button></div>' +
      '<div class="modal-section-hint">' + escapeHtml(t('sessions.hint')) + '</div>' +
      '<div id="claude-sessions-body" data-project-id="' + projectId + '">' + buildClaudeSessionsInnerHTML(projectId) + '</div>' +
      '</div>');
    fetchClaudeSessions(projectId, true);
  }

  function refreshClaudeSessionsUI(projectId) {
    var project = state.projectsById[projectId];
    if (!project) return;
    var pill = document.querySelector('[data-sessions-pill="' + projectId + '"]');
    if (pill) pill.outerHTML = buildSessionsPillHTML(project);
    var body = document.getElementById('claude-sessions-body');
    // La modale n'est réécrite que si rien n'y est déplié : une liste qui se
    // replie sous les yeux pendant qu'on lit, c'est pire qu'une date en retard.
    if (body && body.getAttribute('data-project-id') === projectId && !body.querySelector('details[open]')) {
      body.innerHTML = buildClaudeSessionsInnerHTML(projectId);
    }
  }

  function findClaudeSession(projectId, id) {
    return claudeSessionList(projectId).filter(function (x) { return x.id === id; })[0] || null;
  }

  function copyClaudeSessionCmd(projectId, id, el) {
    var x = findClaudeSession(projectId, id);
    if (!x) return;
    writeClipboard(x.resumeCmd, function () {
      var before = el.textContent;
      el.textContent = t('sessions.copied');
      setTimeout(function () { el.textContent = before; }, 1300);
    });
  }

  // Pas une reprise de la session (elle est rangée sous son dossier de
  // lancement, et ses changements non commités sont dans le dépôt principal) :
  // une tâche neuve, dans son worktree, qui part de ce qu'on y demandait.
  function claudeSessionToWorkstream(projectId, id) {
    var x = findClaudeSession(projectId, id);
    if (!x) return;
    var parts = [t('sessions.promptHeader', { title: x.title })];
    if (x.branch) parts.push(t('sessions.promptBranch', { cwd: x.cwd, branch: x.branch }));
    if ((x.prompts || []).length) {
      parts.push('', t('sessions.promptAsks'));
      x.prompts.forEach(function (pr) { parts.push('- ' + pr.replace(/\n+/g, ' ')); });
    }
    closeModal();
    openNewTaskInProject(projectId, { title: x.title, prompt: parts.join('\n'), repoName: x.repoName, waitsForTaskId: '' });
  }

  setInterval(function () {
    // Le kanban et l'écran d'un chantier : partout où l'on peut livrer.
    if (state.projectId && !document.hidden && ['kanban', 'work'].indexOf(state.screen) >= 0) fetchClaudeSessions(state.projectId);
  }, CLAUDE_SESSIONS_TTL);

  // ---------------------------------------------------------------------
  // Configuration automatique d'un projet (voir autoconfig.go)
  // ---------------------------------------------------------------------
  // Un agent en lecture seule propose les commandes ; rien n'est enregistré
  // sans l'humain. « Reprendre ces valeurs » remplit le brouillon de la
  // modale, et c'est Enregistrer qui écrit : une commande qui s'exécutera après
  // chaque tâche se relit avant.

  var autoconfApplied = {};

  function projectIsConfigured(p) {
    return !!(p.checkCmd || (p.repos || []).some(function (r) { return r.previewCmd; }));
  }

  function autoconfIconHTML() {
    return '<svg viewBox="0 0 16 16" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true" focusable="false">' +
      '<path d="M2.5 13.5l7-7M8.5 5.5l2 2"/><path d="M12 1.8v2.4M10.8 3h2.4M13.6 6.6v1.8M12.7 7.5h1.8M6 1.6v1.6M5.2 2.4h1.6"/></svg>';
  }

  function defaultAutoconfAgent(project) {
    var agents = analystAgents();
    var watchAgent = project.watch && project.watch.agentId;
    if (watchAgent && agents.some(function (a) { return a.id === watchAgent; })) return watchAgent;
    var real = agents.filter(function (a) { return a.cli !== 'fake' && !a.warning; })[0];
    return (real || agents[0] || {}).id || '';
  }

  function buildAutoconfSectionHTML(project) {
    return '<section id="autoconf-section" class="autoconf" data-project-id="' + project.id + '">' + buildAutoconfInnerHTML(project) + '</section>';
  }

  function autoconfCode(v) {
    return v ? '<code class="mono">' + escapeHtml(v) + '</code>' : '<span class="autoconf-none">' + escapeHtml(t('autoconf.nothing')) + '</span>';
  }

  function autoconfRowHTML(label, value, current) {
    return '<div class="autoconf-row"><div class="autoconf-label">' + escapeHtml(label) + '</div><div class="autoconf-value">' + autoconfCode(value) +
      (current && current !== value ? '<div class="autoconf-current">' + escapeHtml(t('autoconf.current', { v: current })) + '</div>' : '') +
      '</div></div>';
  }

  function buildAutoconfInnerHTML(project) {
    var run = state.autoconfigs[project.id];
    var pid = project.id;
    var agent = run && state.agentsById[run.agentId];
    var agentName = agent ? agent.name : '';
    var head = '<div class="autoconf-head"><span class="autoconf-icon">' + autoconfIconHTML() + '</span>' +
      '<div class="autoconf-head-text"><div class="autoconf-title">' + escapeHtml(t('autoconf.title')) + '</div>';
    if (!run) {
      var agents = analystAgents();
      head += '<div class="autoconf-hint">' + escapeHtml(t('autoconf.hint')) + '</div></div></div>';
      if (!agents.length) return head + '<div class="modal-note">' + escapeHtml(t('autoconf.noAgent')) + '</div>';
      var chosen = defaultAutoconfAgent(project);
      return head + '<div class="autoconf-actions">' +
        (agents.length > 1 ? '<select id="autoconf-agent" class="modal-input autoconf-agent" aria-label="' + escapeHtml(t('autoconf.with')) + '">' +
          agents.map(function (a) { return '<option value="' + a.id + '"' + (a.id === chosen ? ' selected' : '') + '>' + a.emoji + ' ' + escapeHtml(a.name) + '</option>'; }).join('') +
        '</select>' : '') +
        '<button class="btn-primary" data-action="autoconf-start" data-project-id="' + pid + '">' + escapeHtml(t('autoconf.start')) + '</button></div>';
    }
    if (run.status === 'running') {
      var steps = run.steps || [];
      return head + '<div class="autoconf-live"><span class="task-spinner"></span>' +
          '<span>' + escapeHtml(t('autoconf.running', { agent: agentName, repo: run.repoName || '' })) + ' · ' + elapsedHTML(run.startedAt) +
          (run.timeoutSec ? ' · ' + escapeHtml(t('vigie.live.limit', { n: Math.round(run.timeoutSec / 60) })) : '') + '</span>' +
          '<button class="btn-quiet" data-action="autoconf-stop" data-project-id="' + pid + '">' + escapeHtml(t('autoconf.stop')) + '</button></div></div></div>' +
        (steps.length ? '<ol class="live-steps">' + steps.map(function (line, i) {
          return '<li class="mono' + (i === steps.length - 1 ? ' live-step-now' : '') + '">' + escapeHtml(line) + '</li>';
        }).join('') + '</ol>' : '<div class="sd-live-now">' + escapeHtml(t('vigie.live.start')) + '</div>');
    }
    var again = '<button class="btn-quiet" data-action="autoconf-start" data-project-id="' + pid + '">' + escapeHtml(t('autoconf.retry')) + '</button>' +
      '<button class="btn-quiet" data-action="autoconf-dismiss" data-project-id="' + pid + '">' + escapeHtml(t('autoconf.dismiss')) + '</button>';
    if (run.status === 'error') {
      return head + '<div class="autoconf-hint signal-verdict-error">' + escapeHtml(run.error === 'stopped' ? t('autoconf.stopped') : t('autoconf.error', { error: run.error })) + '</div></div></div>' +
        '<div class="autoconf-actions">' + again + '</div>';
    }
    var prop = run.proposal || { repos: [] };
    var draft = projectDraft || {};
    var rows = autoconfRowHTML(t('autoconf.check'), prop.checkCmd, draft.checkCmd);
    (prop.repos || []).forEach(function (r) {
      var multi = (prop.repos || []).length > 1;
      var current = (modalRepos || []).filter(function (m) { return m.name === r.name; })[0] || {};
      if (r.error) {
        rows += '<div class="autoconf-row"><div class="autoconf-label">' + escapeHtml(r.name) + '</div><div class="autoconf-value signal-verdict-error">' + escapeHtml(t('autoconf.error', { error: r.error })) + '</div></div>';
        return;
      }
      rows += autoconfRowHTML(t('autoconf.preview') + (multi ? ' · ' + r.name : ''), r.previewCmd, current.previewCmd);
      if (r.previewUrl) rows += autoconfRowHTML(t('autoconf.url') + (multi ? ' · ' + r.name : ''), r.previewUrl, current.previewUrl);
      var found = [['install', r.install], ['build', r.build], ['test', r.test]].filter(function (x) { return x[1]; });
      if (found.length) {
        rows += '<div class="autoconf-row autoconf-row-minor"><div class="autoconf-label">' + escapeHtml(t('autoconf.found') + (multi ? ' · ' + r.name : '')) + '</div><div class="autoconf-value">' +
          found.map(function (x) { return '<div><span class="autoconf-step">' + escapeHtml(t('autoconf.' + x[0])) + '</span> <code class="mono">' + escapeHtml(x[1]) + '</code></div>'; }).join('') +
        '</div></div>';
      }
      if (r.notes) {
        rows += '<div class="autoconf-row autoconf-row-minor"><div class="autoconf-label">' + escapeHtml(t('autoconf.notes') + (multi ? ' · ' + r.name : '')) + '</div>' +
          '<div class="autoconf-value"><div class="autoconf-notes">' + escapeHtml(r.notes) + '</div></div></div>';
      }
    });
    if ((prop.permissionProfiles || []).length) {
      rows += '<div class="autoconf-row"><div class="autoconf-label">' + escapeHtml(t('autoconf.profiles')) + '</div><div class="autoconf-value">' +
        prop.permissionProfiles.map(function (id) {
          var pr = PERMISSION_PROFILES.filter(function (x) { return x.id === id; })[0];
          return '<span class="signal-label">' + escapeHtml(pr ? t(pr.labelKey) : id) + '</span>';
        }).join(' ') + '</div></div>';
    }
    var applied = autoconfApplied[pid];
    return head + '<div class="autoconf-hint">' + escapeHtml(t('autoconf.ready', { agent: agentName })) + '</div></div></div>' +
      '<div class="autoconf-proposal">' + rows + '</div>' +
      '<div class="autoconf-warning">' + escapeHtml(applied ? t('autoconf.applied') : t('autoconf.warning')) + '</div>' +
      '<div class="autoconf-actions">' +
        (applied ? '' : '<button class="btn-primary" data-action="autoconf-apply" data-project-id="' + pid + '">' + escapeHtml(t('autoconf.apply')) + '</button>') +
        again + '</div>';
  }

  function refreshAutoconfUI(projectId) {
    var el = document.getElementById('autoconf-section');
    var project = state.projectsById[projectId];
    if (el && project && el.getAttribute('data-project-id') === projectId) el.innerHTML = buildAutoconfInnerHTML(project);
    var pill = document.querySelector('[data-autoconf-pill="' + projectId + '"]');
    if (pill && project) pill.outerHTML = buildAutoconfPillHTML(project);
    else if (state.screen === 'kanban' && state.projectId === projectId) renderMain();
  }

  function onAutoconfigEvent(r) {
    if (!r || !r.projectId) return;
    if (!r.status) delete state.autoconfigs[r.projectId];
    else state.autoconfigs[r.projectId] = r;
    refreshAutoconfUI(r.projectId);
  }

  function autoconfFail(projectId, err) {
    state.autoconfigs[projectId] = { projectId: projectId, status: 'error', error: (err && err.message) || t('common.networkError') };
    refreshAutoconfUI(projectId);
  }

  function startAutoconf(projectId) {
    var project = state.projectsById[projectId];
    if (!project) return;
    var sel = document.getElementById('autoconf-agent');
    var agentId = (sel && sel.value) || defaultAutoconfAgent(project);
    delete autoconfApplied[projectId];
    api('/api/projects/' + encodeURIComponent(projectId) + '/autoconfigure', { method: 'POST', body: { agentId: agentId } })
      .then(function (run) { state.autoconfigs[projectId] = run; refreshAutoconfUI(projectId); })
      .catch(function (e) { autoconfFail(projectId, e); });
  }

  function stopAutoconf(projectId) {
    api('/api/projects/' + encodeURIComponent(projectId) + '/autoconfigure/stop', { method: 'POST' }).catch(function () {});
  }

  function forgetAutoconf(projectId) {
    delete state.autoconfigs[projectId];
    delete autoconfApplied[projectId];
    refreshAutoconfUI(projectId);
    api('/api/projects/' + encodeURIComponent(projectId) + '/autoconfigure', { method: 'DELETE' }).catch(function () {});
  }

  // Les valeurs proposées vont dans le brouillon, champ par champ ; un champ
  // pour lequel l'agent n'a rien trouvé garde ce qui y était.
  function applyAutoconf(projectId) {
    var run = state.autoconfigs[projectId];
    if (!run || !run.proposal || !projectDraft) return;
    captureProjectDraftFromDOM();
    var prop = run.proposal;
    if (prop.checkCmd) projectDraft.checkCmd = prop.checkCmd;
    (prop.repos || []).forEach(function (r) {
      var m = (modalRepos || []).filter(function (x) { return x.name === r.name; })[0];
      if (!m || r.error) return;
      if (r.previewCmd) { m.previewCmd = r.previewCmd; m.previewUrl = r.previewUrl || ''; }
    });
    (prop.permissionProfiles || []).forEach(function (id) {
      if (projectDraft.permissionProfiles.indexOf(id) < 0) projectDraft.permissionProfiles.push(id);
    });
    autoconfApplied[projectId] = true;
    markProjectDraftDirty();
    refreshAutoconfUI(projectId);
  }

  function openAutoconf(projectId, start) {
    state.projectId = projectId;
    openEditProjectModal('general', projectId);
    if (start && !state.autoconfigs[projectId]) startAutoconf(projectId);
  }

  // Sous le titre du projet : l'invitation quand rien n'est configuré, ou
  // l'état d'une configuration en cours.
  function buildAutoconfPillHTML(project) {
    var run = state.autoconfigs[project.id];
    var attr = ' data-autoconf-pill="' + project.id + '"';
    if (run && run.status === 'running') {
      var agent = state.agentsById[run.agentId];
      return '<button class="autoconf-pill"' + attr + ' data-action="autoconf-open" data-project-id="' + project.id + '"><span class="task-spinner"></span>' +
        escapeHtml(t('autoconf.runningShort', { agent: agent ? agent.name : '' })) + ' · ' + elapsedHTML(run.startedAt) + '</button>';
    }
    if (run && run.status === 'ready' && !autoconfApplied[project.id]) {
      return '<button class="autoconf-pill autoconf-pill-ready"' + attr + ' data-action="autoconf-open" data-project-id="' + project.id + '">' +
        autoconfIconHTML() + escapeHtml(t('autoconf.readyShort')) + '</button>';
    }
    if (!projectIsConfigured(project) && analystAgents().length) {
      return '<div class="autoconf-nudge"' + attr + '>' + escapeHtml(t('autoconf.nudge')) +
        ' <button class="detail-link" data-action="autoconf-open" data-start="1" data-project-id="' + project.id + '">' + escapeHtml(t('autoconf.nudgeAction')) + '</button></div>';
    }
    return '<span' + attr + ' hidden></span>';
  }

  // Réglages de projet : une modale à panneaux, pas un formulaire de onze
  // champs empilés. La colonne de gauche EST le découpage (aucun filet
  // horizontal), chaque panneau ne montre que deux ou trois choses, et ce qui ne
  // se règle qu'une fois (dépôts, livraison, suppression) ne pèse plus sur ce
  // qu'on change souvent (nom, description, instructions).

  var PROJECT_TABS = [
    { key: 'general', labelKey: 'project.tabGeneral' },
    { key: 'repos', labelKey: 'project.tabRepos' },
    { key: 'instructions', labelKey: 'project.tabInstructions' },
    { key: 'permissions', labelKey: 'project.tabPermissions' },
    { key: 'delivery', labelKey: 'project.tabDelivery' },
    { key: 'links', labelKey: 'project.tabLinks' },
    { key: 'danger', labelKey: 'project.tabDanger', danger: true }
  ];

  // Un champ court : libellé à gauche, saisie à droite, tous alignés sur la même
  // verticale. Deux fois moins haut qu'un libellé empilé, et l'alignement se lit
  // comme de l'ordre.
  function fieldRowHTML(labelKey, inputId, inputHTML) {
    return '<div class="field-row"><label class="modal-label" for="' + inputId + '">' + escapeHtml(t(labelKey)) + '</label>' +
      '<div class="field-row-input">' + inputHTML + '</div></div>';
  }

  function buildProjectGeneralPanelHTML() {
    var project = state.projectsById[state.projectId];
    return (project ? buildAutoconfSectionHTML(project) : '') + fieldRowHTML('project.name', 'project-edit-name',
        '<input id="project-edit-name" class="modal-input" value="' + escapeHtml(projectDraft.name) + '">') +
      fieldRowHTML('project.description', 'project-description',
        '<input id="project-description" class="modal-input" placeholder="' + escapeHtml(t('project.descriptionPlaceholder')) + '" value="' + escapeHtml(projectDraft.description) + '">');
  }

  function buildProjectInstructionsPanelHTML() {
    return '<div class="modal-section-hint">' + escapeHtml(t('project.instructionsHint')) + '</div>' +
      '<textarea id="project-context-prompt" class="modal-textarea modal-textarea-tall" rows="12" placeholder="' + escapeHtml(t('project.contextPromptPlaceholder')) + '">' + escapeHtml(projectDraft.contextPrompt) + '</textarea>' +
      '<label class="modal-label" for="project-edit-checkcmd">' + escapeHtml(t('project.checkCmd')) + '</label>' +
      '<input id="project-edit-checkcmd" class="modal-input mono" placeholder="go test ./..." value="' + escapeHtml(projectDraft.checkCmd) + '">';
  }

  var PERMISSION_PROFILES = [
    { id: 'ruby', labelKey: 'project.profile.ruby', descKey: 'project.profile.rubyDesc' },
    { id: 'node', labelKey: 'project.profile.node', descKey: 'project.profile.nodeDesc' },
    { id: 'php', labelKey: 'project.profile.php', descKey: 'project.profile.phpDesc' },
    { id: 'python', labelKey: 'project.profile.python', descKey: 'project.profile.pythonDesc' },
    { id: 'go', labelKey: 'project.profile.go', descKey: 'project.profile.goDesc' },
    { id: 'rust', labelKey: 'project.profile.rust', descKey: 'project.profile.rustDesc' }
  ];

  function buildProjectPermissionsPanelHTML() {
    var selected = projectDraft.permissionProfiles || [];
    var cards = PERMISSION_PROFILES.map(function (profile) {
      var checked = selected.indexOf(profile.id) >= 0;
      return '<label class="choice-card' + (checked ? ' choice-card-active' : '') + '">' +
        '<input type="checkbox" name="project-permission-profile" value="' + profile.id + '"' + (checked ? ' checked' : '') + '>' +
        '<span class="choice-card-text">' +
          '<span class="choice-card-title">' + escapeHtml(t(profile.labelKey)) + '</span>' +
          '<span class="choice-card-desc">' + escapeHtml(t(profile.descKey)) + '</span>' +
        '</span>' +
      '</label>';
    }).join('');
    return '<div class="modal-section-hint">' + escapeHtml(t('project.permissionProfilesHint')) + '</div>' +
      '<div class="choice-cards permission-profile-cards">' + cards + '</div>' +
      '<label class="modal-label" for="project-allowed-tools">' + escapeHtml(t('project.allowedTools')) + '</label>' +
      '<textarea id="project-allowed-tools" class="modal-textarea mono" rows="4" placeholder="' + escapeHtml(t('project.allowedToolsPlaceholder')) + '">' + escapeHtml(projectDraft.allowedTools) + '</textarea>' +
      '<div class="modal-note">' + escapeHtml(t('project.allowedToolsHint')) + '</div>';
  }

  function refreshPermissionProfileCards() {
    var cards = document.querySelectorAll('.permission-profile-cards .choice-card');
    Array.prototype.forEach.call(cards, function (card) {
      var input = card.querySelector('input[type="checkbox"]');
      card.classList.toggle('choice-card-active', !!(input && input.checked));
    });
  }

  function buildProjectDangerPanelHTML(project) {
    var deleteKey = 'project-delete:' + project.id;
    var deleteLabel = isPendingConfirm(deleteKey) ? t('project.deleteConfirm') : t('project.delete');
    var cardCount = state.cards.filter(function (c) { return c.projectId === project.id; }).length;
    var taskCount = state.tasks.filter(function (tk) { return tk.projectId === project.id; }).length;
    return '<div class="danger-box">' +
      '<div class="modal-section-hint">' + escapeHtml(t('project.deleteSubtext', { cards: cardCount, tasks: taskCount })) + '</div>' +
      '<div class="modal-note">' + escapeHtml(t('project.deleteWarning')) + '</div>' +
      '<div class="danger-panel-action">' +
        '<button class="btn-danger" data-action="confirm-click" data-confirm-key="' + deleteKey + '" data-confirm-action="project-delete" data-confirm-id="' + project.id + '" data-default-label="' + escapeHtml(t('project.delete')) + '" data-confirm-label="' + escapeHtml(t('project.deleteConfirm')) + '">' + escapeHtml(deleteLabel) + '</button>' +
      '</div>' +
    '</div>';
  }

  function buildProjectPanelHTML(project) {
    if (projectModalTab === 'repos') return buildRepoSectionHTML();
    if (projectModalTab === 'instructions') return buildProjectInstructionsPanelHTML();
    if (projectModalTab === 'permissions') return buildProjectPermissionsPanelHTML();
    if (projectModalTab === 'delivery') return buildDeliveryPanelHTML(project);
    if (projectModalTab === 'links') return buildLinksSectionHTML();
    if (projectModalTab === 'danger') return buildProjectDangerPanelHTML(project);
    return buildProjectGeneralPanelHTML();
  }

  function buildProjectTabsHTML() {
    return PROJECT_TABS.map(function (tb) {
      var active = projectModalTab === tb.key;
      return '<button class="ptab' + (active ? ' ptab-active' : '') + (tb.danger ? ' ptab-danger' : '') + '"' +
        ' role="tab" aria-selected="' + (active ? 'true' : 'false') + '"' +
        ' data-action="set-project-tab" data-project-tab="' + tb.key + '">' + escapeHtml(t(tb.labelKey)) + '</button>';
    }).join('');
  }

  function buildProjectModalBodyHTML(project) {
    return '<nav class="ptabs-nav" role="tablist">' + buildProjectTabsHTML() + '</nav>' +
      '<div class="ptabs-panel" role="tabpanel">' + buildProjectPanelHTML(project) + '</div>';
  }

  function buildProjectModalHTML(project) {
    return '<div class="modal modal-tabbed">' +
      '<div class="modal-head"><span class="modal-title">' + escapeHtml(t('project.editTitle')) + '</span><button class="icon-btn" data-action="close-modal" aria-label="' + escapeHtml(t('common.close')) + '">✕</button></div>' +
      '<div class="ptabs" id="project-modal-body">' + buildProjectModalBodyHTML(project) + '</div>' +
      '<div id="project-modal-error" class="modal-error hidden"></div>' +
      '<div class="modal-foot">' +
        '<span id="project-unsaved" class="modal-hint' + (projectDraftDirty ? '' : ' hidden') + '">' + escapeHtml(t('project.unsaved')) + '</span>' +
        '<button class="btn-outline" data-action="close-modal">' + escapeHtml(t('common.cancel')) + '</button>' +
        '<button class="btn-green" data-action="submit-project-edit" data-project-id="' + project.id + '">' + escapeHtml(t('common.save')) + '</button>' +
      '</div>' +
      '</div>';
  }

  // splitAllowedTools : une entrée par ligne côté saisie, une liste côté API.
  // Le serveur retire de toute façon les vides et les espaces superflus
  // (NormalizeAllowedTools), on ne duplique pas la validation ici.
  function splitAllowedTools(value) {
    return String(value || '').split('\n').map(function (line) {
      return line.trim();
    }).filter(function (line) { return line !== ''; });
  }

  // Le brouillon retient les champs simples pendant qu'on navigue d'un panneau à
  // l'autre : seul le panneau visible existe dans le DOM.
  function captureProjectDraftFromDOM() {
    if (!projectDraft) return;
    var fields = {
      'project-edit-name': 'name',
      'project-description': 'description',
      'project-delivery-target': 'target',
      'project-context-prompt': 'contextPrompt',
      'project-edit-checkcmd': 'checkCmd',
      'project-allowed-tools': 'allowedTools'
    };
    Object.keys(fields).forEach(function (id) {
      var el = document.getElementById(id);
      if (el) projectDraft[fields[id]] = el.value;
    });
    var mode = document.querySelector('input[name="project-delivery-mode"]:checked');
    if (mode) projectDraft.mode = mode.value;
    var profileInputs = document.querySelectorAll('input[name="project-permission-profile"]');
    if (profileInputs.length) {
      projectDraft.permissionProfiles = Array.prototype.filter.call(profileInputs, function (input) {
        return input.checked;
      }).map(function (input) { return input.value; });
    }
    captureRepoRowsFromDOM();
    captureLinksFromDOM();
  }

  function setProjectTab(tabKey) {
    var project = state.projectsById[state.projectId];
    if (!project) return;
    captureProjectDraftFromDOM();
    projectModalTab = tabKey;
    var body = document.getElementById('project-modal-body');
    if (body) body.innerHTML = buildProjectModalBodyHTML(project);
  }

  function openEditProjectModal(initialTab, projectId) {
    var project = state.projectsById[projectId || state.projectId];
    if (!project) return;
    projectModalTab = initialTab || 'general';
    projectDraftDirty = false;
    projectDraft = {
      name: project.name || '',
      description: project.description || '',
      target: (project.delivery && project.delivery.target) || '',
      contextPrompt: project.contextPrompt || '',
      checkCmd: project.checkCmd || '',
      // Une entrée par ligne dans le champ, une liste côté API.
      allowedTools: (project.allowedTools || []).join('\n'),
      permissionProfiles: (project.permissionProfiles || []).slice(),
      mode: (project.delivery && project.delivery.mode) || 'pr'
    };
    var repos = (project.repos && project.repos.length) ? project.repos : [{ name: '', path: '' }];
    modalRepos = repos.map(function (r) {
      return {
        name: r.name || '', path: r.path || '',
        previewCmd: r.previewCmd || '', previewUrl: r.previewUrl || ''
      };
    });
    modalLinks = (project.links || []).map(function (l) { return { url: l.url || '', title: l.title || '' }; });
    openModal(buildProjectModalHTML(project));
    focusById('project-edit-name');
  }

  // Une saisie quelque part dans la modale allume la mention « modifications non
  // enregistrées » : changer d'onglet ne doit pas donner l'impression d'avoir
  // perdu ce qu'on venait d'écrire.
  function markProjectDraftDirty() {
    if (!projectDraft || projectDraftDirty) return;
    projectDraftDirty = true;
    var el = document.getElementById('project-unsaved');
    if (el) el.classList.remove('hidden');
  }

  function submitProjectEdit(projectId) {
    var errEl = document.getElementById('project-modal-error');
    captureProjectDraftFromDOM();
    var name = projectDraft.name.trim();
    if (!name) {
      errEl.textContent = t('project.errorNameRequired');
      errEl.classList.remove('hidden');
      setProjectTab('general');
      return;
    }
    var repos = collectReposForSubmit();
    if (repos.length === 0) {
      errEl.textContent = t('project.errorReposRequired');
      errEl.classList.remove('hidden');
      setProjectTab('repos');
      return;
    }
    var links = collectLinksForSubmit();
    var body = {
      name: name, checkCmd: projectDraft.checkCmd.trim(), repos: reposToBody(repos),
      description: projectDraft.description.trim(), contextPrompt: projectDraft.contextPrompt.trim(),
      allowedTools: splitAllowedTools(projectDraft.allowedTools),
      permissionProfiles: projectDraft.permissionProfiles.slice(),
      links: linksToBody(links),
      delivery: { mode: projectDraft.mode, target: projectDraft.target.trim() }
    };
    api('/api/projects/' + projectId, { method: 'PATCH', body: body }).then(function (project) {
      upsertProject(project);
      invalidateDeliveryForProject(project.id);
      // Une proposition reprise puis enregistrée n'a plus rien à dire.
      if (autoconfApplied[project.id]) { delete autoconfApplied[project.id]; forgetAutoconf(project.id); }
      closeModal();
      renderSidebar(); renderMain();
    }).catch(function (e) {
      errEl.textContent = (e instanceof ApiError && e.message) || t('project.errorSaveFailed');
      errEl.classList.remove('hidden');
    });
  }
  function doDeleteProject(projectId) {
    api('/api/projects/' + projectId, { method: 'DELETE', body: { confirm: true } }).then(function () {
      closeModal();
      goAllProjects();
      fetchStateSilently();
    }).catch(function (e) {
      var errEl = document.getElementById('project-modal-error');
      if (errEl) {
        errEl.textContent = (e instanceof ApiError && e.message) || t('errors.deleteProjectFailed');
        errEl.classList.remove('hidden');
      }
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

  // Onglet Statistiques de Settings : consommation de tokens par projet,
  // sans prix (charge mentale). Le contenu vit dans #usage-section pour
  // permettre un patch ciblé sur l'événement SSE `tokens`, sans reconstruire
  // toute la modale.
  function buildStatsSectionInnerHTML() {
    var global = (state.tokens && state.tokens.global) || {};
    var rows = state.projects
      .map(function (p) { return { name: p.name, total: tokenTotal(p.tokens) }; })
      .sort(function (a, b) { return b.total - a.total; })
      .map(function (r) {
        return '<div class="usage-project-row"><span class="usage-project-name">' + escapeHtml(r.name) + '</span>' +
          '<span class="usage-project-value">' + formatTokens(r.total) + ' ' + escapeHtml(t('tokens.unit')) + '</span></div>';
      }).join('');
    return '<div class="usage-global">Σ ' + formatTokens(tokenTotal(global)) + ' ' + escapeHtml(t('tokens.unit')) + '</div>' +
      (rows ? '<div class="usage-project-list">' + rows + '</div>' : '<div class="empty-note-sm">' + escapeHtml(t('usage.empty')) + '</div>');
  }
  function buildStatsSectionHTML() {
    return '<div id="usage-section">' + buildStatsSectionInnerHTML() + '</div>';
  }

  function buildSettingsTabsHTML() {
    var tabs = [
      { key: 'general', label: t('settings.tabGeneral') },
      { key: 'sync', label: t('settings.tabSync') },
      { key: 'agents', label: t('sidebar.agentsHeading') },
      { key: 'update', label: t('update.title') },
      { key: 'stats', label: t('settings.tabStats') }
    ];
    return '<div class="tabs">' + tabs.map(function (tb) {
      var active = settingsModalTab === tb.key;
      return '<button class="tab ' + (active ? 'tab-active' : '') + '" role="tab" data-action="set-settings-tab" data-settings-tab="' + tb.key + '">' +
        escapeHtml(tb.label) + '</button>';
    }).join('') + '</div>';
  }
  function setSettingsTab(tabKey) {
    settingsModalTab = (tabKey === 'stats' || tabKey === 'update' || tabKey === 'agents' || tabKey === 'sync') ? tabKey : 'general';
    refreshWorkspaceModalBody();
  }

  // Onglet Agents : la liste complète, sortie de la barre latérale où elle ne
  // servait qu'à ouvrir cette même modale d'édition. Une ligne par agent, son
  // modèle en sous-titre, son avertissement d'installation s'il en a un.
  function buildSettingsAgentsHTML() {
    var rows = state.agents.map(function (a) {
      var warn = a.warning
        ? '<span class="agent-warning-sm" title="' + escapeHtml(agentWarningText(a.warning)) + '">⚠</span>'
        : '';
      return '<button class="agent-manage-row" data-action="edit-agent" data-agent-id="' + a.id + '">' +
        '<span class="agent-avatar" style="background:' + softColor(a.color) + '">' + a.emoji + '</span>' +
        '<span class="agent-manage-main">' +
          '<span class="agent-manage-name">' + escapeHtml(a.name) + warn + '</span>' +
          '<span class="agent-manage-model">' + escapeHtml(a.model || a.cli || '') + '</span>' +
        '</span>' +
        (a.active ? '<span class="agent-dot"></span>' : '') +
      '</button>';
    }).join('');
    return '<div class="agent-manage-list">' +
        (rows || '<div class="empty-note-sm">' + escapeHtml(t('sidebar.noAgents')) + '</div>') +
      '</div>' +
      '<div class="secondary-row"><button class="btn-outline btn-block" data-action="open-new-agent">+ ' +
        escapeHtml(t('sidebar.newAgentTooltip')) + '</button></div>';
  }

  // Accès depuis le téléphone ou un autre portable : un interrupteur, les
  // adresses à taper, et l'empreinte du certificat pour qui veut vérifier
  // l'avertissement du navigateur au lieu de le subir.
  var lanError = '';
  function buildLanSectionInnerHTML() {
    var lan = state.lan || {};
    var html = '<div class="modal-label">' + escapeHtml(t('lan.heading')) + '</div>' +
      '<label class="workspace-autosync-row">' +
        '<input type="checkbox" data-action="toggle-lan"' + (lan.enabled ? ' checked' : '') + (lan.noPassword && !lan.enabled ? ' disabled' : '') + '>' +
        '<span>' + escapeHtml(t('lan.toggle')) + '</span>' +
      '</label>';
    if (lan.noPassword && !lan.enabled) {
      html += '<div class="update-note">' + escapeHtml(t('lan.noPassword')) + '</div>';
    } else if (lan.enabled) {
      var urls = lan.urls || [];
      html += urls.length
        ? urls.map(function (u) {
            return '<div class="agent-warning-cmd"><code class="mono">' + escapeHtml(u) + '</code>' +
              '<button data-action="copy-path" data-path="' + escapeHtml(u) + '">' + escapeHtml(t('lan.copy')) + '</button></div>';
          }).join('')
        : '<div class="update-note">' + escapeHtml(t('lan.noAddress')) + '</div>';
      html += '<div class="update-note">' + escapeHtml(t('lan.certNote')) + '</div>' +
        (lan.fingerprint ? '<div class="lan-fingerprint mono">' + escapeHtml(lan.fingerprint) + '</div>' : '');
    } else {
      html += '<div class="update-note">' + escapeHtml(t('lan.offNote')) + '</div>';
    }
    var err = lanError || lan.error;
    if (err) html += '<div class="modal-error">' + escapeHtml(err) + '</div>';
    return html + '<div class="preferences-divider"></div>';
  }
  function refreshLanSection() {
    var el = document.getElementById('lan-section');
    if (el) el.innerHTML = buildLanSectionInnerHTML();
  }
  function saveLanSetting(enabled) {
    lanError = '';
    api('/api/settings', { method: 'PATCH', body: { lan: enabled } }).then(function (settings) {
      if (settings) state.settings = settings;
    }).catch(function (e) {
      lanError = (e instanceof ApiError && e.message) || t('preferences.errorSaveFailed');
    }).then(refreshLanSection);
  }

  function buildSettingsGeneralHTML() {
    return buildPreferencesSectionHTML() +
      '<div id="lan-section">' + buildLanSectionInnerHTML() + '</div>' +
      '<button class="btn-outline btn-block" data-action="logout">' + escapeHtml(t('nav.logout')) + '</button>';
  }

  // Onglet Sauvegarde : synchronisation git de l'espace de travail, séparée
  // des préférences pour qu'on ne la croise qu'en la cherchant.
  function buildSettingsSyncHTML() {
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

    var autoSyncRow = '';
    if (gitEnabled && hasRemote) {
      var lastSyncErrorHTML = ws.lastSyncError ? '<div class="workspace-sync-error">' + escapeHtml(ws.lastSyncError) + '</div>' : '';
      autoSyncRow = '<label class="workspace-autosync-row">' +
        '<input type="checkbox" id="workspace-autosync-checkbox" data-action="toggle-autosync"' + (ws.autoSync ? ' checked' : '') + '>' +
        '<span>' + escapeHtml(t('workspace.autoSync')) + '</span>' +
        '</label>' +
        lastSyncErrorHTML;
    }

    return '<div class="workspace-state">' + escapeHtml(stateLabel) + '</div>' +
      '<div class="modal-label">' + escapeHtml(t('workspace.remoteLabel')) + '</div>' +
      '<div class="workspace-remote-row">' +
        '<input id="workspace-remote-input" class="modal-input mono" placeholder="git@github.com:vous/sillage-workspace.git" value="' + escapeHtml(ws.remote || '') + '">' +
        // Sans git initialisé, pas de PATCH possible : le bouton « Activer » ci-dessous
        // embarque le remote saisi et fait l'initialisation et le branchement d'un coup.
        (gitEnabled ? '<button class="link-save-btn" data-action="save-workspace-remote">' + escapeHtml(t('common.save')) + '</button>' : '') +
      '</div>' +
      '<div class="workspace-warning">' + escapeHtml(t('workspace.privateWarning')) + '</div>' +
      '<div id="workspace-modal-error" class="modal-error hidden"></div>' +
      '<div id="workspace-sync-message" class="workspace-sync-message hidden"></div>' +
      '<div class="secondary-row">' + primaryHTML + '</div>' +
      autoSyncRow +
      '<div class="workspace-meta">' +
        '<span>' + escapeHtml(t('workspace.lastCommit', { time: lastCommit })) + '</span>' +
        '<span>' + escapeHtml(t('workspace.lastSync', { time: lastSync })) + '</span>' +
      '</div>' +
      dirtyNote;
  }

  // Le dépôt et le sponsoring : sur tous les onglets de la modale, pas
  // seulement Général, pour rester accessibles quel que soit l'onglet ouvert.
  function buildAboutLinksHTML() {
    return '<div class="about-links">' +
      '<a href="https://github.com/Halleck45/sillage" target="_blank" rel="noopener noreferrer">' + repoIconHTML() + escapeHtml(t('sidebar.repoTooltip')) + '</a>' +
      '<a href="https://github.com/sponsors/Halleck45/" target="_blank" rel="noopener noreferrer">' + sponsorIconHTML() + escapeHtml(t('sidebar.sponsorTooltip')) + '</a>' +
    '</div>';
  }
  function buildWorkspaceModalBodyHTML() {
    var inner = settingsModalTab === 'stats' ? buildStatsSectionHTML()
      : settingsModalTab === 'update' ? buildUpdateSectionHTML()
      : settingsModalTab === 'agents' ? buildSettingsAgentsHTML()
      : settingsModalTab === 'sync' ? buildSettingsSyncHTML()
      : buildSettingsGeneralHTML();
    return buildSettingsTabsHTML() + '<div class="settings-tab-body">' + inner + '</div>' + buildAboutLinksHTML();
  }
  function buildWorkspaceModalHTML() {
    return '<div class="modal modal-sm">' +
      '<div class="modal-head"><span class="modal-title">' + escapeHtml(t('workspace.title')) + '</span><button class="icon-btn" data-action="close-modal" aria-label="' + escapeHtml(t('common.close')) + '">✕</button></div>' +
      '<div id="workspace-modal-body">' + buildWorkspaceModalBodyHTML() + '</div>' +
      '</div>';
  }
  function openWorkspaceModal() {
    settingsModalTab = 'general';
    openModal(buildWorkspaceModalHTML());
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
  function saveAutoSync(enabled) {
    var errEl = document.getElementById('workspace-modal-error');
    api('/api/workspace', { method: 'PATCH', body: { autoSync: enabled } }).then(function (ws) {
      state.workspace = ws || state.workspace;
      refreshWorkspaceModalBody();
    }).catch(function (e) {
      if (errEl) {
        errEl.textContent = (e instanceof ApiError && e.message) || t('workspace.errorAutoSyncFailed');
        errEl.classList.remove('hidden');
      }
      refreshWorkspaceModalBody();
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
  // Mise à jour de Sillage
  // ---------------------------------------------------------------------

  // La modale est le seul endroit qui parle de versions : la sidebar ne porte
  // qu'une ligne, et seulement quand il y a réellement quelque chose à faire.
  // Ces quatre variables ne décrivent que l'opération en cours côté navigateur
  // (l'état, lui, vient de state.update et du SSE `update`).
  var updateChecking = false;
  var updateApplying = false;
  var updateMessage = '';
  var updateError = '';

  function updateState() { return state.update || { current: 'dev', method: 'dev' }; }

  // Ligne de pied de sidebar : un rappel de version, qui devient un appel à
  // l'action quand une version existe. Rien d'autre : une mise à jour n'est
  // jamais urgente au point de couvrir l'écran.
  function buildUpdateLineHTML() {
    var u = updateState();
    if (u.available) {
      return '<button class="update-line update-line-available" data-action="open-update-modal">' +
        '<span class="update-dot"></span>' + escapeHtml(t('update.sidebarAvailable')) + '</button>';
    }
    if (u.method === 'dev') return '';
    return '<button class="update-line" data-action="open-update-modal">' +
      escapeHtml(t('update.currentVersion', { version: u.current })) + '</button>';
  }

  function updateMethodLabel(u) {
    if (u.method === 'binary') {
      var path = u.path || '';
      var dir = path.indexOf('/') >= 0 ? path.slice(0, path.lastIndexOf('/')) : path;
      return dir ? t('update.method.binary', { dir: dir }) : '';
    }
    if (u.method === 'brew' || u.method === 'go' || u.method === 'unknown') return t('update.method.' + u.method);
    return '';
  }

  function buildUpdateSectionInnerHTML() {
    var u = updateState();
    var isDev = u.method === 'dev';

    var headline;
    if (isDev) headline = t('update.devBuild');
    else if (u.available) headline = t('update.availableHeadline', { latest: u.latest, current: u.current });
    else headline = t('update.upToDate');

    var metaBits = [t('update.currentVersion', { version: u.current })];
    var methodLabel = updateMethodLabel(u);
    if (methodLabel) metaBits.push(methodLabel);
    metaBits.push(u.checkedAt ? t('update.lastChecked', { time: timeAgo(u.checkedAt) }) : t('update.neverChecked'));

    var notesHTML = (u.available && u.releaseUrl)
      ? '<a class="update-notes-link" href="' + escapeHtml(u.releaseUrl) + '" target="_blank" rel="noopener noreferrer">' +
        escapeHtml(t('update.releaseNotes')) + '</a>'
      : '';

    // Un blocage temporaire (agent au travail, recette en cours) laisse le
    // bouton visible mais éteint : la raison est plus utile qu'un bouton absent.
    var blockerHTML = (u.available && u.blocker)
      ? '<div class="update-blocker">' + escapeHtml(t('update.blocker.' + u.blocker)) + '</div>'
      : '';

    // `applying` vient aussi du serveur : une mise à jour lancée depuis un autre
    // onglet (ou avant la fermeture de cette modale) reste visible ici.
    var applying = updateApplying || !!u.applying;
    var actionHTML = '';
    if (u.available && u.selfUpdatable) {
      var disabled = applying || !!u.blocker;
      actionHTML = '<button class="btn-green btn-block" data-action="update-apply"' + (disabled ? ' disabled' : '') + '>' +
        escapeHtml(applying ? t('update.applying') : t('update.apply')) + '</button>';
    }

    // La commande à la main est toujours offerte quand une version existe :
    // c'est le seul recours des installations que Sillage ne peut pas toucher,
    // et un repli rassurant pour les autres.
    var manualHTML = '';
    if (u.available && u.command) {
      manualHTML = '<div class="update-manual">' +
        '<div class="update-manual-intro">' + escapeHtml(t('update.manualIntro')) + '</div>' +
        '<div class="agent-warning-cmd">' +
          '<code class="mono">' + escapeHtml(u.command) + '</code>' +
          '<button data-action="copy-path" data-path="' + escapeHtml(u.command) + '">' + escapeHtml(t('agent.warning.copyCmd')) + '</button>' +
        '</div>' +
      '</div>';
    }

    var checkHTML = isDev ? '' :
      '<button class="update-check-btn" data-action="update-check"' + (updateChecking ? ' disabled' : '') + '>' +
      escapeHtml(updateChecking ? t('update.checking') : t('update.checkNow')) + '</button>';

    var autoHTML = '<label class="workspace-autosync-row">' +
      '<input type="checkbox" data-action="toggle-update-check"' + (u.checkEnabled ? ' checked' : '') + '>' +
      '<span>' + escapeHtml(t('update.autoCheckLab')) + '</span>' +
      '</label>' +
      '<div class="update-note">' + escapeHtml(t('update.autoCheckNote')) + '</div>';

    // Lancement à l'ouverture de session. Absent de la réponse = aucune
    // réponse sûre (installation hors Homebrew, brew muet) : on n'affiche rien
    // plutôt que d'affirmer à tort que ce n'est pas configuré.
    var serviceHTML = '';
    if (u.service) {
      var registered = !!u.service.registered;
      var flagsNote = (!registered && u.service.customFlags)
        ? '<div class="update-note">' + escapeHtml(t('update.serviceFlagsNote')) + '</div>'
        : '';
      var serviceCmd = (!registered && u.service.command)
        ? '<div class="agent-warning-cmd">' +
            '<code class="mono">' + escapeHtml(u.service.command) + '</code>' +
            '<button data-action="copy-path" data-path="' + escapeHtml(u.service.command) + '">' + escapeHtml(t('agent.warning.copyCmd')) + '</button>' +
          '</div>'
        : '';
      serviceHTML = '<div class="preferences-divider"></div>' +
        '<div class="modal-label">' + escapeHtml(t('update.serviceHeading')) + '</div>' +
        '<div class="update-service-state' + (registered ? ' update-service-on' : '') + '">' +
          escapeHtml(t(registered ? 'update.serviceOn' : 'update.serviceOff')) + '</div>' +
        serviceCmd +
        flagsNote;
    }

    // Une vérification qui a échoué se dit : sinon « à jour » serait une
    // promesse que le serveur n'a pas pu tenir.
    var checkErrorHTML = (!updateError && u.error)
      ? '<div class="update-blocker">' + escapeHtml(t('update.errorCheckFailed')) + '</div>'
      : '';

    return '<div class="update-headline">' + escapeHtml(headline) + '</div>' +
      notesHTML +
      '<div class="update-meta">' + metaBits.map(escapeHtml).join(' · ') + '</div>' +
      checkErrorHTML +
      blockerHTML +
      (updateMessage ? '<div class="workspace-sync-message">' + escapeHtml(updateMessage) + '</div>' : '') +
      (updateError ? '<div class="modal-error">' + escapeHtml(updateError) + '</div>' : '') +
      (actionHTML ? '<div class="secondary-row">' + actionHTML + '</div>' : '') +
      manualHTML +
      '<div class="preferences-divider"></div>' +
      checkHTML +
      autoHTML +
      serviceHTML;
  }

  // Le contenu vit dans #update-section pour permettre un patch ciblé sur
  // l'événement SSE `update`, sans reconstruire toute la modale (même principe
  // que #usage-section pour les statistiques).
  function buildUpdateSectionHTML() {
    return '<div id="update-section">' + buildUpdateSectionInnerHTML() + '</div>';
  }

  // Ouvre les réglages directement sur l'onglet Mises à jour : c'est la cible
  // de la ligne de pied de sidebar.
  function openUpdateModal() {
    updateMessage = '';
    updateError = '';
    settingsModalTab = 'update';
    openModal(buildWorkspaceModalHTML());
  }

  function refreshUpdateSection() {
    var body = document.getElementById('update-section');
    if (body) body.innerHTML = buildUpdateSectionInnerHTML();
  }

  function doUpdateCheck() {
    updateChecking = true;
    updateError = '';
    refreshUpdateSection();
    api('/api/update/check', { method: 'POST', body: {} }).then(function (u) {
      if (u) state.update = u;
    }).catch(function (e) {
      updateError = (e instanceof ApiError && e.message) || t('update.errorCheckFailed');
    }).then(function () {
      updateChecking = false;
      refreshUpdateSection();
      renderSidebar();
    });
  }

  function saveUpdateCheckSetting(enabled) {
    api('/api/settings', { method: 'PATCH', body: { updateCheck: enabled } }).then(function (settings) {
      if (settings) state.settings = settings;
      if (state.update) state.update.checkEnabled = enabled;
      refreshUpdateSection();
    }).catch(function (e) {
      updateError = (e instanceof ApiError && e.message) || t('preferences.errorSaveFailed');
      refreshUpdateSection();
    });
  }

  function doUpdateApply() {
    updateApplying = true;
    updateError = '';
    updateMessage = '';
    refreshUpdateSection();
    api('/api/update/apply', { method: 'POST', body: { confirm: true } }).then(function (res) {
      var version = (res && res.version) || '';
      if (res && res.restarting) {
        updateMessage = t('update.applied', { version: version });
        refreshUpdateSection();
        waitForRestart();
        return;
      }
      updateApplying = false;
      updateMessage = t('update.appliedNoRestart', { version: version });
      refreshUpdateSection();
    }).catch(function (e) {
      updateApplying = false;
      updateError = (e instanceof ApiError && e.message) || t('update.errorApplyFailed');
      refreshUpdateSection();
    });
  }

  // Le serveur se remplace par sa nouvelle version : on attend qu'il réponde à
  // nouveau, puis on recharge la page. Le rechargement est indispensable, le
  // frontend étant embarqué dans le binaire : sans lui, l'ancienne interface
  // continuerait de tourner contre la nouvelle API. Un mot de passe configuré
  // ramènera l'écran de connexion, les sessions vivant en mémoire.
  function waitForRestart() {
    var deadline = Date.now() + 60000;
    updateMessage = t('update.reconnecting');
    refreshUpdateSection();
    (function poll() {
      setTimeout(function () {
        fetch('/api/update', { credentials: 'same-origin' }).then(function () {
          window.location.reload();
        }).catch(function () {
          if (Date.now() < deadline) poll();
          else {
            updateApplying = false;
            updateError = t('common.networkError');
            refreshUpdateSection();
          }
        });
      }, 1500);
    })();
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

  // Ces trois gestionnaires sont le coeur du rendu incrémental (confort de
  // conversation) : ils ne rappellent JAMAIS renderMain()/render() pour une
  // tâche déjà ouverte, afin de ne jamais reconstruire le conteneur des
  // messages. Seuls les éléments concernés sont mis à jour (liste de tâches,
  // en-tête du détail, pied du diff) ; le fil lui-même ne reçoit qu'un append
  // de nœud (onMessageEvent) et n'est donc jamais recréé.
  function onTaskEvent(task) {
    // Un agent qui travaille publie beaucoup d'événements de tâche : l'aperçu
    // de livraison (git log + git diff par dépôt) n'est rechargé que si le
    // statut a réellement changé.
    var previous = state.tasksById[task.id];
    var statusChanged = !previous || previous.status !== task.status;
    // Un rebase automatique qui vient de finir a effacé le retard de la tâche :
    // l'aperçu (qui porte les compteurs de retard) n'est plus à jour.
    var rebaseFinished = !!(previous && previous.rebasing && !task.rebasing);
    upsertTask(task);
    // Une tâche déjà ouverte peut redevenir "non lue" côté serveur (fin
    // d'exécution d'agent, voir finalize() dans runner.go) : la remarquer
    // lue immédiatement, sinon le badge reste bloqué à 1 sans raison visible.
    if (state.taskId === task.id && task.unread) {
      task.unread = false;
      api('/api/tasks/' + task.id + '/read', { method: 'POST' }).catch(function () {});
    }
    refreshTaskListAndFilters();
    if (state.taskId === task.id) {
      patchDetailHead(task.id);
      if (state.panelTab === 'chat') patchConversationThinking(task.id);
      if (state.panelTab === 'diff') patchDiffFooter(task.id);
    }
    if ((statusChanged || rebaseFinished) && state.cardId && task.cardId === state.cardId) {
      loadDelivery(state.cardId);
    }
    renderSidebar();
  }

  function onMessageEvent(m) {
    if (state.messagesByTask[m.taskId]) {
      var exists = state.messagesByTask[m.taskId].some(function (x) { return x.id === m.id; });
      if (!exists) state.messagesByTask[m.taskId].push(m);
    }
    if (state.taskId === m.taskId && state.panelTab === 'chat') {
      appendMessageToConversationDOM(m);
    }
  }

  function onActivityEvent(payload) {
    var task = state.tasksById[payload.taskId];
    if (task) task.liveActivity = payload.line;
    patchTaskRowLiveLine(payload.taskId, payload.line, task);
    patchConversationThinking(payload.taskId);
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
    var usageSection = document.getElementById('usage-section');
    if (usageSection) usageSection.innerHTML = buildStatsSectionInnerHTML();
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

  // La vérification périodique a trouvé quelque chose (ou l'application d'une
  // mise à jour a changé d'état) : la ligne de sidebar et la modale ouverte
  // suivent, sans rendu complet.
  function onUpdateEvent(u) {
    state.update = u;
    renderSidebar();
    refreshUpdateSection();
  }

  // Recette : l'état d'un run change (démarré, terminé, arrêté). Le panneau et
  // les boutons se mettent à jour sans rendu complet ; le journal, lui, reçoit
  // ses lignes une par une (onPreviewLogEvent).
  function onPreviewEvent(run) {
    upsertPreview(run);
    refreshPreviewModal();
    renderSidebar();
    if (state.screen === 'work') refreshPreviewButton();
    if (state.taskId) patchDetailHead(state.taskId);
  }

  function onPreviewLogEvent(payload) {
    if (!payload || !payload.runId) return;
    appendPreviewLogLine(payload.runId, payload.line);
  }

  // Le bouton Recette de l'en-tête porte la pastille « une recette tourne » :
  // remplacé seul, pour ne pas reconstruire l'écran à chaque événement.
  function refreshPreviewButton() {
    var card = state.cardsById[state.cardId];
    var btn = document.querySelector('.topbar-actions .preview-btn');
    if (!card || !btn) return;
    var wrapper = document.createElement('div');
    wrapper.innerHTML = buildPreviewButtonHTML(card);
    var fresh = wrapper.firstElementChild;
    if (fresh) btn.replaceWith(fresh);
  }

  function onSettingsEvent(settings) {
    state.settings = settings || state.settings;
    refreshWorkspaceModalBody();
  }

  // Suppressions (tâche/chantier/projet) : purge locale + sortie propre si
  // l'objet actuellement affiché est celui qui vient de disparaître.
  function onTaskDeletedEvent(payload) {
    if (!payload || !payload.taskId) return;
    var wasOpen = state.taskId === payload.taskId;
    removeTaskLocally(payload.taskId);
    if (wasOpen) {
      closePanel();
    } else {
      renderSidebar();
      refreshTaskListAndFilters();
    }
  }

  function onCardDeletedEvent(payload) {
    if (!payload || !payload.cardId) return;
    var wasOpen = state.cardId === payload.cardId;
    removeCardLocally(payload.cardId);
    if (wasOpen) {
      goProject(payload.projectId);
    } else {
      renderSidebar();
      if (state.projectId === payload.projectId) renderMain();
    }
  }

  function onProjectDeletedEvent(payload) {
    if (!payload || !payload.projectId) return;
    var wasCurrent = state.projectId === payload.projectId;
    if (wasCurrent) goAllProjects();
    fetchStateSilently();
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
    es.addEventListener('update', function (e) { try { onUpdateEvent(JSON.parse(e.data)); } catch (er) {} });
    es.addEventListener('settings', function (e) { try { onSettingsEvent(JSON.parse(e.data)); } catch (er) {} });
    es.addEventListener('preview', function (e) { try { onPreviewEvent(JSON.parse(e.data)); } catch (er) {} });
    es.addEventListener('signal', function (e) { try { onSignalEvent(JSON.parse(e.data)); } catch (er) {} });
    es.addEventListener('watch', function (e) { try { onWatchEvent(JSON.parse(e.data)); } catch (er) {} });
    es.addEventListener('autoconfig', function (e) { try { onAutoconfigEvent(JSON.parse(e.data)); } catch (er) {} });
    es.addEventListener('lan', function (e) { try { state.lan = JSON.parse(e.data) || {}; refreshLanSection(); } catch (er) {} });
    es.addEventListener('previewLog', function (e) { try { onPreviewLogEvent(JSON.parse(e.data)); } catch (er) {} });
    es.addEventListener('taskDeleted', function (e) { try { onTaskDeletedEvent(JSON.parse(e.data)); } catch (er) {} });
    es.addEventListener('cardDeleted', function (e) { try { onCardDeletedEvent(JSON.parse(e.data)); } catch (er) {} });
    es.addEventListener('projectDeleted', function (e) { try { onProjectDeletedEvent(JSON.parse(e.data)); } catch (er) {} });
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
    if (!el) { closeAllCardMenus(); closeAllReassignMenus(); closeAllProjectMenus(); return; }
    var action = el.getAttribute('data-action');
    // Le menu « ⋯ » du détail d'une tâche partage la classe .card-menu (donc sa
    // fermeture au clic ailleurs) : il faut exempter les deux bascules.
    if (action !== 'toggle-card-menu' && action !== 'toggle-task-menu') closeAllCardMenus();
    if (action !== 'toggle-reassign-menu') closeAllReassignMenus();
    if (action !== 'toggle-project-menu') closeAllProjectMenus();
    switch (action) {
      case 'nav-inbox': goInbox(); break;
      case 'nav-vigie': goVigie(); break;
      case 'vigie-check': vigieCheck(); break;
      case 'vigie-upkeep': api('/api/projects/' + encodeURIComponent(el.getAttribute('data-project-id')) + '/upkeep', { method: 'POST' }).catch(vigieFail); break;
      case 'vigie-upkeep-stop': api('/api/projects/' + encodeURIComponent(el.getAttribute('data-project-id')) + '/upkeep/stop', { method: 'POST' }).catch(vigieFail); break;
      case 'vigie-watch': setProjectWatch(el.getAttribute('data-project-id'), el.getAttribute('data-enabled') === '1'); break;
      case 'vigie-notify':
        if (canNotify()) Notification.requestPermission().then(function () { refreshVigie(); });
        break;
      case 'vigie-log-toggle': state.vigieLogOpen = !state.vigieLogOpen; renderMain(); break;
      case 'signal-dismiss': dismissSignal(el.getAttribute('data-signal-id')); break;
      case 'signal-stop': stopSignal(el.getAttribute('data-signal-id')); break;
      case 'open-claude-sessions': openClaudeSessionsModal(el.getAttribute('data-project-id')); break;
      case 'claude-session-copy': copyClaudeSessionCmd(el.getAttribute('data-project-id'), el.getAttribute('data-session-id'), el); break;
      case 'claude-session-workstream': claudeSessionToWorkstream(el.getAttribute('data-project-id'), el.getAttribute('data-session-id')); break;
      case 'autoconf-start': startAutoconf(el.getAttribute('data-project-id')); break;
      case 'autoconf-stop': stopAutoconf(el.getAttribute('data-project-id')); break;
      case 'autoconf-dismiss': forgetAutoconf(el.getAttribute('data-project-id')); break;
      case 'autoconf-apply': applyAutoconf(el.getAttribute('data-project-id')); break;
      case 'autoconf-open': openAutoconf(el.getAttribute('data-project-id'), el.getAttribute('data-start') === '1'); break;
      case 'signal-retry': patchSignal(el.getAttribute('data-signal-id'), { status: 'queued' }).catch(vigieFail); break;
      case 'signal-copy': copySignalReview(el.getAttribute('data-signal-id'), el); break;
      case 'signal-card': signalToCard(el.getAttribute('data-signal-id'), el); break;
      case 'signal-workstream': startWorkstreamFromSignal(el.getAttribute('data-signal-id')); break;
      case 'nav-projects': goAllProjects(); break;
      case 'nav-project': goProject(el.getAttribute('data-project-id')); break;
      case 'open-card': goCard(el.getAttribute('data-card-id')); break;
      case 'go-back': goBack(); break;
      case 'open-task': openTask(el.getAttribute('data-task-id')); break;
      case 'close-panel': closePanel(); break;
      case 'toggle-panel-expand': togglePanelExpand(); break;
      case 'set-filter': setFilter(el.getAttribute('data-filter')); break;
      case 'set-tab': setTab(el.getAttribute('data-panel-tab')); break;
      case 'toggle-card-menu': toggleCardMenu(el.getAttribute('data-card-id')); break;
      case 'toggle-task-menu': toggleTaskMenu(); break;
      case 'toggle-reassign-menu': toggleReassignMenu(el.getAttribute('data-task-id')); break;
      case 'reassign-task': doReassignTask(el.getAttribute('data-task-id'), el.getAttribute('data-agent-id')); break;
      case 'toggle-project-menu': toggleProjectMenu(el.getAttribute('data-project-id'), el, el.getAttribute('data-menu-scope')); break;
      case 'mark-project-read': markProjectAllRead(el.getAttribute('data-project-id')); break;
      case 'toggle-project-sidebar-visibility': toggleProjectSidebarVisibility(el.getAttribute('data-project-id')); break;
      case 'move-card': moveCard(el.getAttribute('data-card-id'), el.getAttribute('data-column')); break;
      case 'open-new-card': openNewCardModal(); break;
      case 'open-new-task': if (state.cardId) openNewTaskModal(state.cardId); break;
      case 'open-new-project-task': if (state.projectId) openNewTaskInProject(state.projectId); break;
      case 'open-plan': if (state.cardId) openPlanModal(state.cardId); break;
      case 'submit-plan': submitPlan(); break;
      case 'plan-remove-step': removePlanStep(parseInt(el.getAttribute('data-index'), 10)); break;
      case 'plan-restart': restartPlan(); break;
      case 'plan-create-tasks': createPlanTasks(); break;
      // Le formulaire de création cède la place à la liste d'issues : ce qui y
      // était saisi est relu avant, pour revenir dessus intact.
      case 'open-issue-picker': {
        // La modale porte son chantier (vide en mode projet) : c'est lui, pas
        // l'écran derrière, qui dit où la tâche sera créée.
        var pickerCardId = el.getAttribute('data-card-id') || '';
        captureNewTaskDraft();
        openIssuePicker(pickerCardId);
        break;
      }
      case 'pick-issue': pickIssue(parseInt(el.getAttribute('data-number'), 10)); break;
      case 'cancel-issue-picker': cancelIssuePicker(); break;
      case 'open-new-project': openNewProjectModal(); break;
      case 'open-edit-project': openEditProjectModal(); break;
      case 'set-project-tab': setProjectTab(el.getAttribute('data-project-tab')); break;
      case 'submit-project-edit': submitProjectEdit(el.getAttribute('data-project-id')); break;
      case 'open-edit-card': openEditCardModal(); break;
      case 'submit-card-edit': submitCardEdit(el.getAttribute('data-card-id')); break;
      case 'add-repo-row': addRepoRow(); break;
      case 'remove-repo-row': removeRepoRow(parseInt(el.getAttribute('data-index'), 10)); break;
      case 'add-link-row': addLinkRow(); break;
      case 'remove-link-row': removeLinkRow(parseInt(el.getAttribute('data-index'), 10)); break;
      case 'open-new-agent': openNewAgentModal(); break;
      case 'edit-agent': openEditAgentModal(el.getAttribute('data-agent-id')); break;
      case 'submit-agent': submitAgent(el.getAttribute('data-agent-id') || null); break;
      case 'close-modal': closeModal(); break;
      case 'pick-agent': pickAgentInModal(el.getAttribute('data-agent-id')); break;
      case 'submit-new-task': submitNewTask(el.getAttribute('data-card-id'), 'chat'); break;
      case 'submit-new-task-another': submitNewTask(el.getAttribute('data-card-id'), 'another'); break;
      case 'open-shortcuts': openShortcutsModal(); break;
      case 'submit-new-project': submitNewProject(); break;
      case 'submit-new-card': submitNewCard(); break;
      case 'submit-new-card-plan': submitNewCard('plan'); break;
      case 'interrupt': doInterrupt(el.getAttribute('data-task-id')); break;
      case 'start-task': doStartWaitingTask(el.getAttribute('data-task-id')); break;
      case 'reopen': doReopen(el.getAttribute('data-task-id')); break;
      case 'accept-task': doAcceptTask(el.getAttribute('data-task-id')); break;
      case 'refuse-task': doCancelTask(el.getAttribute('data-task-id')); break;
      case 'ask-rebase': askRebase(el.getAttribute('data-task-id')); break;
      case 'open-ship-modal': openShipModal(el.getAttribute('data-card-id')); break;
      case 'submit-ship': submitShip(el.getAttribute('data-card-id')); break;
      case 'catch-up-card': catchUpCard(el.getAttribute('data-card-id')); break;
      case 'open-card-preview': openPreviewModal('card', el.getAttribute('data-card-id')); break;
      case 'open-task-preview': openPreviewModal('task', el.getAttribute('data-task-id')); break;
      case 'open-all-previews': openPreviewModal('all', null); break;
      case 'start-card-preview':
        startCardPreview(el.getAttribute('data-card-id'), el.getAttribute('data-repo-name'));
        break;
      case 'start-task-preview': startTaskPreview(el.getAttribute('data-task-id')); break;
      case 'stop-preview': stopPreview(el.getAttribute('data-run-id')); break;
      case 'show-preview-log': showPreviewLog(el.getAttribute('data-run-id')); break;
      case 'open-preview-settings': openPreviewSettings(); break;
      case 'copy-path': copyPathToClipboard(el.getAttribute('data-path'), el); break;
      case 'catch-up-ask-agent':
        askAgentToCatchUp(el.getAttribute('data-card-id'), el.getAttribute('data-repo-name'), el.getAttribute('data-target'), el.getAttribute('data-files'));
        break;
      case 'confirm-click': handleConfirmClickDispatch(el); break;
      case 'diff-toggle-file': toggleDiffFile(el.getAttribute('data-task-id'), el.getAttribute('data-path')); break;
      case 'diff-toggle-all': toggleDiffAll(el.getAttribute('data-task-id')); break;
      case 'diff-tree-toggle': toggleDiffTree(); break;
      case 'diff-jump': jumpToDiffFile(el.getAttribute('data-task-id'), el.getAttribute('data-path')); break;
      case 'open-difftool': doOpenDifftool(el.getAttribute('data-task-id'), el.getAttribute('data-path')); break;
      case 'send-message': sendMessage(el.getAttribute('data-task-id')); break;
      case 'attach-image': openAttachPicker(el.getAttribute('data-target')); break;
      case 'remove-attachment': removeAttachment(el.getAttribute('data-scope') || '', el.getAttribute('data-key')); break;
      case 'open-attachment': openAttachmentModal(el.getAttribute('data-src'), el.getAttribute('data-name')); break;
      case 'jump-to-bottom': jumpConversationToBottom(); break;
      case 'copy-message': copyToClipboardIcon(el.getAttribute('data-copy') || '', el); break;
      case 'copy-code': copyCodeBlock(el); break;
      case 'open-search': openSearch(); break;
      case 'search-goto-project': closeSearch(); goProject(el.getAttribute('data-project-id')); break;
      case 'search-goto-card': closeSearch(); goCard(el.getAttribute('data-card-id')); break;
      case 'search-goto-task': closeSearch(); openTaskFromSearch(el.getAttribute('data-task-id')); break;
      case 'logout': doLogout(); break;
      case 'toggle-sidebar': toggleSidebarDrawer(); break;
      case 'close-sidebar-drawer': closeSidebarDrawer(); break;
      case 'set-lang': setLang(el.getAttribute('data-lang')); break;
      case 'open-workspace-modal': openWorkspaceModal(); break;
      case 'open-settings-agents': closeSidebarDrawer(); settingsModalTab = 'agents'; openModal(buildWorkspaceModalHTML()); break;
      case 'open-signal': openSignal(el.getAttribute('data-signal-id')); break;
      case 'external': break; // lien sortant dans une carte : le navigateur l'ouvre, la carte reste fermée
      case 'set-settings-tab': setSettingsTab(el.getAttribute('data-settings-tab')); break;
      case 'save-workspace-remote': saveWorkspaceRemote(); break;
      case 'toggle-autosync': saveAutoSync(el.checked); break;
      case 'save-display-name': saveDisplayName(); break;
      case 'activate-workspace-git': activateWorkspaceGit(); break;
      case 'toggle-onboarding-card': toggleOnboardingCard(el.getAttribute('data-key')); break;
      case 'submit-onboarding': submitOnboarding(el.getAttribute('data-mode')); break;
      case 'open-update-modal': openUpdateModal(); break;
      case 'update-check': doUpdateCheck(); break;
      case 'update-apply': doUpdateApply(); break;
      case 'toggle-update-check': saveUpdateCheckSetting(el.checked); break;
      case 'toggle-lan': saveLanSetting(el.checked); break;
    }
  }

  // Une frappe dans un champ de saisie n'est jamais un raccourci d'écran.
  function isTypingTarget(el) {
    if (!el) return false;
    var tag = el.tagName;
    return tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT' || el.isContentEditable === true;
  }

  // Action de création principale de l'écran courant (raccourci « N »).
  function createShortcutAction() {
    if (state.screen === 'work' && state.cardId) return function () { openNewTaskModal(state.cardId); };
    if (state.screen === 'kanban' && state.projectId) return function () { openNewTaskInProject(state.projectId); };
    if (state.screen === 'projects') return openNewProjectModal;
    return null;
  }

  function focusableInModal(modal) {
    var sel = 'button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [href], [tabindex]:not([tabindex="-1"])';
    return Array.prototype.slice.call(modal.querySelectorAll(sel)).filter(function (el) {
      return el.getAttribute('tabindex') !== '-1' && el.offsetParent !== null;
    });
  }

  // Piège la tabulation dans la modale ouverte : le focus ne part jamais
  // derrière l'overlay, où les boutons sont inaccessibles à la souris.
  function trapModalTab(e, modal) {
    var items = focusableInModal(modal);
    if (!items.length) return;
    var first = items[0], last = items[items.length - 1];
    var current = document.activeElement;
    var outside = !modal.contains(current);
    if (e.shiftKey && (outside || current === first)) {
      e.preventDefault();
      last.focus();
    } else if (!e.shiftKey && (outside || current === last)) {
      e.preventDefault();
      first.focus();
    }
  }

  var AGENT_ARROW_KEYS = { ArrowRight: 1, ArrowDown: 1, ArrowLeft: -1, ArrowUp: -1 };

  // Bouton de validation d'une modale : le vert du pied, quelle que soit l'action.
  function primarySubmitButton(modal) {
    return modal.querySelector('.modal-foot .btn-green');
  }

  // Raccourcis actifs dans une modale : validation, validation en série,
  // choix de l'agent aux flèches, tabulation piégée.
  function onModalKeydown(e) {
    var modal = document.querySelector('#modal-root .modal');
    if (!modal) return;
    if (e.key === 'Tab') { trapModalTab(e, modal); return; }
    var mod = e.ctrlKey || e.metaKey;
    if (e.key === 'Enter' && mod) {
      var wanted = e.shiftKey ? 'submit-new-task-another' : null;
      var btn = wanted ? modal.querySelector('[data-action="' + wanted + '"]') : primarySubmitButton(modal);
      if (btn) { e.preventDefault(); btn.click(); }
      return;
    }
    // Entrée dans un champ d'une ligne (jamais dans un textarea) : valider.
    if (e.key === 'Enter' && !e.shiftKey && e.target && e.target.tagName === 'INPUT' && modal.contains(e.target)) {
      // Sauf dans le champ d'ajout d'un lien, dont la validation est l'ajout :
      // enregistrer le projet y perdrait l'URL qu'on vient de taper.
      if (e.target.id === 'new-link-url') { e.preventDefault(); addLinkRow(); return; }
      // Sauf dans la recherche d'issues : Entrée y lance la recherche sans
      // attendre la temporisation, c'est le geste qu'on vient de faire.
      if (e.target.id === 'issue-search') { e.preventDefault(); searchIssuesNow(e.target.value); return; }
      // Sauf dans le titre d'une étape de plan : la validation y créerait toutes
      // les tâches (worktrees compris) alors qu'on est en train de relire la
      // proposition, étape par étape.
      if (e.target.classList.contains('plan-step-title')) { e.preventDefault(); return; }
      var primary = primarySubmitButton(modal);
      if (primary) { e.preventDefault(); primary.click(); }
      return;
    }
    if (modal.querySelector('#agent-choices') && AGENT_ARROW_KEYS[e.key] && !isTypingTarget(e.target)) {
      e.preventDefault();
      moveAgentChoice(AGENT_ARROW_KEYS[e.key]);
    }
  }

  // Aucun raccourci avant l'authentification (l'écran de login n'en a pas).
  function shellVisible() {
    var shell = document.getElementById('shell');
    return !!shell && !shell.classList.contains('hidden');
  }

  // Même principe de délégation que onGlobalClick, pour les champs dont la
  // valeur pilote un affichage : cocher un mode de livraison déplace la carte
  // active. On repeint les classes plutôt que le panneau, pour ne pas voler le
  // focus au bouton radio qu'on vient d'atteindre aux flèches.
  function onGlobalChange(e) {
    if (e.target && (e.target.id === 'composer-file' || e.target.id === 'new-task-file')) {
      addAttachmentsForScope(e.target.getAttribute('data-scope') || '', e.target.files);
      // Le champ est vidé pour que rechoisir le même fichier redéclenche
      // l'événement (sinon la valeur ne change pas et rien ne se passe).
      e.target.value = '';
      return;
    }
    if (e.target && e.target.classList && e.target.classList.contains('vigie-agent-select')) {
      setProjectWatch(e.target.getAttribute('data-project-id'), true, e.target.value);
      return;
    }
    if (e.target && e.target.id === 'issue-repo') {
      changeIssueRepo(e.target.value);
      return;
    }
    if (e.target && e.target.name === 'project-delivery-mode') {
      markProjectDraftDirty();
      captureProjectDraftFromDOM();
      refreshDeliveryCards();
      return;
    }
    if (e.target && e.target.name === 'project-permission-profile') {
      markProjectDraftDirty();
      captureProjectDraftFromDOM();
      refreshPermissionProfileCards();
    }
  }

  // Toute saisie dans la modale de réglages d'un projet allume la mention
  // « modifications non enregistrées » du pied.
  function onGlobalInput(e) {
    if (e.target && e.target.id === 'composer-input') {
      autoGrowComposer(e.target);
      refreshComposerState();
      return;
    }
    if (e.target && e.target.id === 'issue-search') {
      scheduleIssueSearch(e.target.value);
      return;
    }
    var body = document.getElementById('project-modal-body');
    if (body && e.target && body.contains(e.target)) markProjectDraftDirty();
  }

  function openAttachPicker(id) {
    var input = document.getElementById(id || 'composer-file');
    if (input) input.click();
  }

  function addAttachmentsForScope(scope, files) {
    if (scope === 'new-task') { addModalAttachmentFiles(files); return; }
    if (scope.indexOf('task:') === 0) addAttachmentFiles(scope.slice(5), files);
  }

  // Ctrl+V d'une image : la capture d'écran qu'on vient de prendre est le cas
  // d'usage principal, elle n'a pas de fichier à choisir. Marche dans le
  // composeur d'une conversation comme dans les champs de la modale de
  // création. Le collage de texte n'est jamais intercepté (aucun fichier
  // image dedans).
  function onGlobalPaste(e) {
    var target = e.target;
    if (!target) return;
    var scope = '';
    if (target.id === 'composer-input') {
      var btn = document.querySelector('.composer [data-action="send-message"]');
      scope = btn ? 'task:' + btn.getAttribute('data-task-id') : '';
    } else if (target.id === 'new-task-prompt' || target.id === 'new-task-title') {
      scope = 'new-task';
    }
    if (!scope) return;
    var data = e.clipboardData;
    if (!data) return;
    var files = [];
    var items = data.items || [];
    for (var i = 0; i < items.length; i++) {
      if (items[i].kind !== 'file') continue;
      var file = items[i].getAsFile();
      if (file && file.type && file.type.indexOf('image/') === 0) files.push(file);
    }
    if (!files.length) return;
    e.preventDefault();
    addAttachmentsForScope(scope, files);
  }

  // Le défilement du fil se capte à la racine (phase de capture : « scroll » ne
  // remonte pas), pour survivre à la recréation du conteneur à chaque rendu.
  function onGlobalScroll(e) {
    if (e.target && e.target.id === 'conversation-list') refreshConvJump();
    // Le menu « ⋯ » d'un projet est en position fixe, ancrée sur son bouton au
    // moment du clic : un défilement du conteneur qui le contient le laisserait
    // flotter loin de ce bouton plutôt que de le suivre.
    if (e.target && e.target.classList && (e.target.classList.contains('project-list') || e.target.classList.contains('view-body'))) closeAllProjectMenus();
  }

  function onGlobalKeydown(e) {
    if (!shellVisible()) return;
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
      if (state.taskId && !isTypingTarget(e.target)) { closePanel(); return; }
    }
    if (state.searchOpen) {
      if (e.key === 'ArrowDown') { e.preventDefault(); moveSearchResult(1); return; }
      if (e.key === 'ArrowUp') { e.preventDefault(); moveSearchResult(-1); return; }
      if (e.key === 'Enter') { e.preventDefault(); activateSearchResult(); return; }
      return;
    }
    if (state.modal) { onModalKeydown(e); return; }
    if (e.target && e.target.id === 'composer-input' && e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      var wrap = e.target.closest('.composer');
      var btn = wrap ? wrap.querySelector('[data-action="send-message"]') : null;
      if (btn) sendMessage(btn.getAttribute('data-task-id'));
      return;
    }
    // Raccourcis d'écran : seulement hors saisie et sans modificateur.
    if (isTypingTarget(e.target) || mod || e.altKey) return;
    if (e.key === 'n' || e.key === 'N') {
      var create = createShortcutAction();
      if (create) { e.preventDefault(); create(); }
      return;
    }
    if (e.key === '/') { e.preventDefault(); openSearch(); return; }
    if (state.taskId && (e.key === 'f' || e.key === 'F')) { e.preventDefault(); togglePanelExpand(); return; }
    if (state.taskId && state.panelTab === 'diff' && (e.key === 'j' || e.key === 'k')) {
      e.preventDefault(); stepDiffFile(e.key === 'j' ? 1 : -1); return;
    }
    if (e.key === '?') { e.preventDefault(); openShortcutsModal(); }
  }

  document.addEventListener('DOMContentLoaded', function () {
    applyStaticTranslations();
    var form = document.getElementById('login-form');
    if (form) form.addEventListener('submit', onLoginSubmit);
    document.addEventListener('click', onGlobalClick);
    document.addEventListener('change', onGlobalChange);
    document.addEventListener('input', onGlobalInput);
    document.addEventListener('paste', onGlobalPaste);
    document.addEventListener('scroll', onGlobalScroll, true);
    document.addEventListener('keydown', onGlobalKeydown);
    document.addEventListener('scroll', onDiffScroll, true);
    window.addEventListener('hashchange', applyRoute);
    boot();
  });
})();
