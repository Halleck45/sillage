# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commandes

```bash
go build -o sillage . && SILLAGE_PASSWORD=dev ./sillage   # build + lancement local (http://127.0.0.1:8787)
go test ./...                                             # tests
go test ./internal/server -run TestReassignTask -v         # un seul test
go vet ./...                                              # vet
gofmt -l .                                                # doit ne rien afficher
node --check web/app.js                                   # syntaxe du frontend
```

La CI (`.github/workflows/ci.yml`) exige les cinq : `gofmt -l .` vide, `go vet ./...`, `go test ./...`, `go build`, `node --check web/app.js`.

Flags du binaire : `-addr` (défaut `127.0.0.1:8787`), `-data` (défaut `~/.local/share/sillage`), `-version`, `-app` (ouvre la fenêtre d'application, démarre le serveur s'il ne répond pas), `-install-app` (lanceur du menu d'applications, Linux). Env : `SILLAGE_PASSWORD` (remplace le hash du mot de passe au démarrage, pratique en dev), `SILLAGE_CODEX_SANDBOX` (mode sandbox de codex, défaut `workspace-write`), `SILLAGE_BROWSER` (navigateur de la fenêtre d'application).

Le frontend est embarqué via `//go:embed web` : toute modification de `web/` demande un rebuild + redémarrage pour être visible.

## Ce que fait le produit

Sillage pilote des agents IA de code (Claude Code, Codex CLI) depuis un kanban web. Un projet regroupe un ou plusieurs dépôts git ; un chantier (carte) est un objectif **et une branche de feature** (`sillage/ws-<ref>-<slug>`, un worktree dédié par dépôt touché) ; une tâche est un agent qui travaille dans son propre **worktree git dédié** sur une branche `sillage/<ref>-<slug>` partant de celle du chantier.

L'humain relit une tâche et l'accepte (fusion locale dans la branche du chantier) ou la refuse ; puis il livre le chantier, seule action sortante du produit. Voir `docs/SPEC-LIVRAISON.md`.

Une acceptation rebase en tâche de fond les autres tâches en revue du chantier (`rebaseSiblingTasks` dans `handlers.go`), pour qu'elles repartent du travail accepté au lieu de conflicter chacune à son tour. Ce que « livrer » veut dire est un réglage du projet (`Project.Delivery`, quatre modes : `pr`, `push`, `merge`, `merge-push`).

Pour tester le flux complet sans coût, utiliser l'agent seedé Écho 🧪 (`cli:"fake"`, simulé, sans process externe).

## Invariants non négociables

Ce sont les promesses de sécurité du produit (voir `CONTRIBUTING.md`) :

1. **`git push` n'existe qu'à deux endroits, tous deux dans `internal/server/git.go`** : `pushBranch()` (dépôts de projet, deux appelants seulement : `Ship()` pour la branche d'un chantier, `mergeThenPush()` pour la branche de destination du mode `merge-push`) et `SyncPush()` (synchronisation de l'espace de données, jamais un dépôt de projet). Aucune entrée capable de pousser dans les allowlists d'outils des agents, aucun flag de contournement de permissions (`--dangerously-skip-permissions` interdit). Les deux modes de fusion n'acceptent que le fast-forward ; `merge` ne pousse jamais rien, `merge-push` ne pousse que la branche de destination, jamais en force.
2. **Les actions sortantes exigent `{"confirm": true}`** sur une requête authentifiée : livraison d'un chantier (la seule action sortante côté projets), sync de l'espace de travail.
3. Le serveur reste sûr sur un portable : localhost par défaut, mot de passe bcrypt, rate-limit du login, `Content-Type: application/json` obligatoire sur les mutations (protection CSRF avec SameSite=Lax).
   Une seule route sans session : `GET /preview/<jeton>/…` (`handlePreviewStatic`), qui sert une recette statique (`previewUrl` en `file:///`). Le jeton de 128 bits du run est son secret, et le contenu part avec une CSP `sandbox` : ne jamais retirer l'une de ces deux protections.
   L'accès depuis le réseau local est un réglage explicite (`Settings.lan`, `lan.go`) : un second listener, toujours en HTTPS (certificat auto-signé dans `<dataDir>/tls/`), refusé tant qu'aucun mot de passe n'est défini. Le listener de `-addr` ne change pas.
   Sillage fait de lui-même trois sortes d'appels réseau, tous en lecture : `git fetch origin <destination>` à l'ouverture d'un projet (`target.go`, au plus une fois toutes les cinq minutes par dépôt, pour annoncer un main local en retard), la vérification de mise à jour (`update.go`, un GET vers l'API GitHub, désactivable dans les réglages, un binaire téléchargé n'est jamais posé sans vérification de son sha256) et la Vigie (`watch.go`, `gh pr list` / `issue list` / `pr diff` / `pr view` et `issue view` (l'état seul, pour constater une PR mergée ou une issue fermée) / `api user` sur les seuls projets qui l'ont activée, et, hors Vigie, `gh pr list --head <branche>` et `gh pr view <n>` pour retrouver la PR d'une branche de chantier ou en adopter une (`pullrequest.go`, aperçu et livraison en mode `pr`/`push`, création d'une tâche avec `pullRequest`), plus le `git fetch origin <tête>` de cette adoption, plus, pour la veille d'entretien, un GET par technologie vers endoflife.date dans `upkeep.go`, qui n'envoie rien du projet). **Rien n'est jamais écrit sur une forge en dehors de la livraison** : les revues rédigées par la Vigie restent sur la machine.
4. Une seule dépendance Go externe : `golang.org/x/crypto`. En ajouter une demande une très bonne raison.

## Les specs sont le contrat

- `docs/SPEC-API.md` : contrat HTTP/JSON/SSE complet (modèles, endpoints, cycle de vie des tâches, événements SSE, règles UI). **À lire avant toute modification de l'API et à mettre à jour dans le même commit** si le contrat change.
- `docs/SPEC-BACKEND.md` : internes (store, auth, SSE, git, adaptateurs d'agents, seed).
- `docs/SPEC-RECETTE.md` : spec de fournée de la recette manuelle (une commande par dépôt, lancée dans le worktree d'un chantier ou d'une tâche, journal en direct). Le lot 1 est implémenté ; le contrat courant est SPEC-API.md §« Recette manuelle ». Le §5 garde la trace des pistes écartées.
- `docs/SPEC-LIVRAISON.md` : spec de fournée du modèle de livraison (chantier = branche de feature, acceptation, Ship de chantier, réglage de livraison). Contexte du « pourquoi » et lots restants ; le contrat courant reste SPEC-API.md.

## Architecture

Binaire Go unique, frontend vanilla embarqué, persistance JSON, temps réel SSE.

```
main.go                     flags, dataDir, init mot de passe, Store, embed web/, écoute
app.go                      mode app : fenêtre Chromium --app, lanceur .desktop
instance.go                 une instance par dossier de données (flock sur sillage.lock + adresse publiée)
internal/server/
  models.go                 tous les structs JSON de SPEC-API.md
  store.go                  état en mémoire + persistance atomique + compteurs dérivés
  handlers.go               routage (net/http ServeMux avec patterns méthode+chemin), middlewares
  runner.go                 adaptateurs claude / codex / agy / kiro / fake, un process max par tâche
  acp.go                    client ACP (Agent Client Protocol) : JSON-RPC 2.0 sur stdio, sessions,
                            lecture d'un tour (acpTurn), refus de toute demande d'autorisation
  acp_runner.go             runACP : l'adaptateur commun des agents ACP (un acpAgentSpec par agent)
  copilot.go                adaptateur copilot (par ACP) : règles d'outils, COPILOT_HOME isolé par tâche
  preview.go                recette manuelle : un process par worktree, journal en mémoire
  preview_handlers.go       routes de recette (lancer, arrêter, journal)
  attachments.go            images jointes aux messages : écriture sur disque, service HTTP,
                            chemins ajoutés au texte envoyé au CLI
  pullrequest.go            PR d'une branche de chantier : recherche (gh pr list --head), état,
                            adoption de la branche d'une PR existante (adoptPRBranch)
  git.go                    worktrees (chantier + tâche), parser de diff unifié, commits, fusions
                            (MergeBranch, MergeLocal, MergeAndPush), pushBranch + SyncPush
                            (les deux seuls push), Ship, OpenPR (gh/glab)
  update.go                 version du binaire, détection de mise à jour (GitHub), application
                            selon le mode d'installation (brew / binaire), redémarrage sur place
  lan.go                    accès réseau local optionnel : second listener HTTPS, certificat auto-signé
  readonly.go               agent en lecture seule lu au fil de l'eau (claude stream-json, codex --json) :
                            relecture de la Vigie et configuration automatique remontent chaque outil
  autoconfig.go             configuration automatique d'un projet : proposition seulement, jamais enregistrée
  upkeep.go                 veille d'entretien de la Vigie (versions publiées via endoflife.date, constats "upkeep")
  claude_sessions.go        sessions Claude Code lancées à la main dans les dépôts, lues en lecture seule
  watch.go                  la Vigie : relève périodique des PR/issues GitHub (gh, lecture seule),
                            file d'analyse ; signal_analysis.go : prompt et lecture du verdict ;
                            signals_store.go : persistance ; watch_handlers.go : routes
  workspace.go              dataDir en dépôt git optionnel (setup, clone, commit auto throttlé)
  auth.go                   bcrypt, sessions en mémoire, rate-limit login
  sse.go                    Hub pub/sub
web/                        index.html + style.css + app.js (SPA vanilla, zéro dépendance),
                            manifest.json + sw.js (application installable)
```

### Fenêtre d'application (`app.go`, `web/sw.js`)

L'« app » n'est pas un runtime embarqué : c'est une fenêtre `--app` d'un navigateur Chromium déjà installé, ouverte sur le serveur. Pas de webview natif (Wails, Tauri, webview_go) : cgo casserait la compilation croisée `CGO_ENABLED=0` de la release, et WebKitGTK rend mal le `backdrop-filter` du design. Fermer la fenêtre n'arrête rien, les agents vivent dans le serveur.

- **Une instance par dossier de données, pas par port** (`instance.go`) : `flock` sur `<dataDir>/sillage.lock`, pris avant `NewStore` (qui réécrit `state.json` dès le chargement) et tenu toute la vie du process, puis l'adresse réelle publiée dans ce fichier une fois le port lié. Verrou déjà pris : `-app` lit l'adresse, ouvre la fenêtre dessus et sort sans charger l'état, quel que soit `-addr` ; un lancement ordinaire refuse de démarrer et dit où est l'instance. Le noyau libère le verrou à la mort du process (pas de verrou périmé après un crash) ; `O_CLOEXEC` l'empêche de passer aux agents et le relâche à l'`exec` de `restartInPlace`.
- Verrou libre : démarrage normal, fenêtre ouverte une fois le port lié. Si le bind échoue et qu'un Sillage répond sur ce port (binaire antérieur au verrou), `-app` ouvre la fenêtre dessus. Le drapeau est retiré d'`os.Args` : `restartInPlace` rouvrirait sinon une fenêtre à chaque mise à jour, et le statut du service brew verrait des arguments personnalisés.
- Le navigateur est lancé dans son propre groupe de process (un Ctrl+C sur le serveur ne le ferme pas), avec le profil de l'utilisateur : la fenêtre rejoint le process existant et reprend la session.
- `sw.js` ne cache rien. Il ne sert qu'à la page « Sillage ne tourne pas » quand l'app installée s'ouvre sans serveur. Les règles de routage statique (`addRoutes`) sortent du service worker tout ce qui n'est pas une navigation, plus `/api/` et `/preview/` : ne jamais y ajouter de cache, ni intercepter le flux SSE.

### Store (`store.go`)

`Store` est à la fois l'état en mémoire et le format sur disque : ses champs exportés sont sérialisés tel quel dans `<dataDir>/state.json`. Points structurants :

- Un `sync.Mutex` protège tout. Les helpers `recomputeCard/Project/Agent/All` doivent être appelés **verrou tenu** et recalculent les champs dérivés (progression, compteurs, `unread`, tokens agrégés, `active`).
- `recomputeCard` porte aussi deux règles produit : l'état du bouton de livraison (`ShipReady`/`ShipBlocker`, voir `shipReadiness`) et la colonne de la carte, qui ne passe à `done` que si toutes ses tâches sont terminales **et** que le chantier a été livré (`CardBranch.ShippedAt`). « Terminé » veut dire livré ; un chantier livré qui reçoit du travail nouveau en ressort.
- `save()` écrit un fichier temporaire puis `os.Rename` (atomique), et arme le commit git de l'espace de travail. Ce commit est **throttlé** (`workspaceCommitInterval`, 15 min) et non debouncé : un minuteur en attente n'est jamais repoussé, sinon un agent actif (plusieurs sauvegardes par seconde) empêcherait tout commit. Chaque commit stockant un blob complet de `state.json`, commiter à chaque sauvegarde gonfle le dépôt en objets libres pour aucun gain.
- `stateFormatVersion` (const dans `store.go`) est le format de `state.json`, écrit dans le fichier à chaque sauvegarde. **À incrémenter dès qu'un champ persisté est ajouté, renommé ou change de sens** : c'est la seule protection contre un binaire plus ancien, qui sinon charge le fichier puis le réécrit en supprimant en silence les champs qu'il ne connaît pas (`NewStore` sauvegarde immédiatement après le chargement). Un fichier de format supérieur fait échouer `loadStoreFile` (`ErrStateTooNew`) avant toute écriture, et `main` refuse de démarrer. `Store.WrittenBy` garde la version qui a écrit en dernier : `DowngradeWarning()` prévient au démarrage quand deux versions publiées du même format se marchent dessus (ce que le format ne peut pas voir). Attention : la protection n'est effective qu'entre binaires qui la portent tous les deux.
- Les migrations de format se font au chargement dans `loadStoreFile` (`migrateLegacyRepos`, `migrateLegacyWorkspace`, `migrateTaskStatuses`, `migrateLegacyDelivery`, `migrateAgentSeeds`) en relisant le JSON brut : ajouter une migration là, pas ailleurs. `resetTransientTaskFlags`, au même endroit, éteint les états qui ne décrivent qu'un processus disparu avec l'arrêt du serveur : `Task.Rebasing`, et les tâches `running` qui repassent en `review` non lues, avec un marqueur `[interrupted:server-restart]` au fil et leurs compteurs relus depuis git (`TaskWorkCounts`).
- `AgentOut.Warning` (santé de l'agent : binaire absent du PATH, sandbox codex bloqué par AppArmor) est calculé à chaque `ListAgents` et **jamais persisté** : `Agent` n'a pas ce champ.
- `ReloadFromDisk()` remplace le contenu sans changer le pointeur `Store`, pour que sessions et abonnements SSE survivent au rapatriement d'un espace de travail.

### Runner (`runner.go`)

La vérification projet (`check.go`) tourne en tâche de fond après chaque exécution et chaque rebase, jamais en même temps que l'agent : tout ce qui touche au worktree appelle d'abord `CancelCheck`. Son résultat et le verdict de relecture (`Task.Review`) sont attachés à une empreinte du travail (`WorkFingerprint`, stable à travers un rebase sans conflit) : une empreinte différente de `Task.WorkSha` veut dire « périmé ».

Un `procHandle` par tâche au maximum (`map[taskID]*procHandle`). Un message envoyé pendant qu'un agent tourne est mis en file (`pending`) et rejoué à la fin de l'exécution en cours. Les process sont lancés avec `Setpgid` pour qu'`Interrupt` puisse SIGINT le groupe (puis SIGKILL après 5 s).

Six adaptateurs, sélectionnés par `agent.cli` :

- **claude** : `claude -p --output-format stream-json --verbose --permission-mode acceptEdits --allowedTools <claudeAllowedTools> [--append-system-prompt ...] [--resume <sessionId>]`. Parse le JSONL : `system/init` → `sessionId` stocké (reprise de conversation), blocs `text` → Messages, blocs `tool_use` → ligne d'activité live, `result` → tokens et coût. `claudeAllowedTools` est une constante figée : ne pas y ajouter d'outil capable de pousser. Une session introuvable à la reprise (« No conversation found » : créée sous un autre compte ou un autre `CLAUDE_CONFIG_DIR`, ou effacée) est oubliée, et l'exécution repart une fois sans `--resume`, l'historique du fil rejoué comme pour codex ; toute autre erreur remonte avec la dernière ligne de stderr.
- **codex** : `codex exec --json --sandbox <SILLAGE_CODEX_SANDBOX|workspace-write> -C <worktree>`. Pas de reprise de session : l'historique est rejoué via `buildTranscript`. Les événements de tokens portent des **totaux cumulés** : ne garder que le dernier et l'ajouter une seule fois en fin de process (`parseCodexTokenStream`).
- **copilot** : `copilot --acp ...`, piloté par le client ACP (`acp.go`, `runACP` dans `acp_runner.go`). Outils read/write/shell autorisés par drapeaux, avec refus prioritaires pour `git push`, `gh`, `glab` et les réglages Copilot du dépôt ; ces règles s'appliquent en mode ACP, et toute demande d'autorisation qui remonte quand même est refusée (`acpRejectAll`) : Sillage n'accorde jamais rien en cours de route. Jamais `--allow-all` ni le mode autopilot. Session reprise par `session/load` (`Task.SessionID`), outils en direct, tokens lus sur la réponse du prompt. Chaque tâche a son `COPILOT_HOME` sous `<dataDir>/agent-homes/` (`prepareCopilotHome`) : Copilot traite un hook utilisateur en échec comme un refus, donc sans cet isolement un hook écrit pour un autre outil rend la tâche muette. Seuls les noms de compte y sont recopiés, le jeton reste dans le trousseau. Le texte reçu avant l'événement `assistant.turn_start` est jeté (avertissements de démarrage écrits dans le canal des messages). Interrompre envoie d'abord `session/cancel` (`procHandle.softStop`), pour que la session soit enregistrée.
- **agy** : `agy --sandbox --print-timeout=60m --add-dir <worktree> --add-dir <dossier git commun> --print <prompt>`. Sandbox toujours forcé, jamais `--dangerously-skip-permissions`. Sortie texte finale, pas de reprise ni de tokens. Trois contraintes du CLI : `--print` prend le prompt **en valeur** (donc toujours en dernier) ; en sandbox agy ignore le répertoire de travail du process, d'où `--add-dir` sur le worktree ; et le sandbox ne montre que les répertoires ajoutés, donc sans le dossier git commun du dépôt (`GitCommonDir`) toute commande git échoue, le modèle la relance hors sandbox et cette demande auto-refusée tue la session. Ce second répertoire est monté en lecture seule et le sandbox n'a pas de réseau : lectures git seulement, ni écriture de hooks ni push. Les autorisations n'existent pas en drapeau : elles viennent de `~/.gemini/antigravity-cli/settings.json` (`toolPermission` **et** deux règles `read_file`/`write_file` sur la racine des worktrees), que Sillage lit pour avertir (`antigravityWorksHeadlessly` dans `store.go`) et n'écrit que sur clic explicite (`fixAntigravityToolPermission`, `POST /api/agents/{id}/fix-warning`). En mode print, toute demande de confirmation est auto-refusée et la session s'arrête sans rien écrire : une seule suffit à rendre la tâche muette.
- **kiro** : `kiro-cli chat --no-interactive --agent sillage --wrap never <prompt>`, avec `KIRO_API_KEY` obligatoire. Chaque run crée un `KIRO_HOME` temporaire hors du dépôt et y pose un profil aux outils read/write/shell/grep/glob, sans MCP hérité. `write` et `shell` passent par leurs réglages granulaires ; les refus prioritaires couvrent `git push`, `gh`, `glab` et la reconfiguration de Kiro. Ne jamais employer `--trust-all-tools`. Sortie texte finale, pas de reprise ni de tokens ; le profil temporaire est supprimé en fin de run.
- **fake** : simule ~3 s de travail, écrit `SILLAGE-TEST.md` dans le worktree, produit un usage fictif. Aucun process externe.

Le prompt d'un départ frais (lancement initial, ou premier message après réassignation, `sessionId` vide) est préfixé par `Task: <title>\n\n` (`contextualizeCliInput`). Le contexte projet s'ajoute au system prompt pour claude (`buildSystemPrompt`) et en préfixe du prompt pour codex, copilot, agy et kiro ; copilot, agy et kiro reçoivent aussi le contexte agent dans ce préfixe. Pour copilot, le préfixe n'accompagne que le premier prompt d'une session (`acpAgentSpec.FreshPrompt`).

Toute mutation d'état publie les événements SSE correspondants via les helpers `publishTask/publishMessage/publishCards/publishTokens/publishAgents/publishActivity/...` : oublier une publication laisse le frontend désynchronisé jusqu'au prochain reload.

### Frontend (`web/app.js`)

SPA vanilla d'un seul fichier, dans une IIFE, découpée en sections commentées (i18n, constantes, état, utilitaires, couche API, hydratation, routage, rendus, modales, SSE, amorçage). Conventions :

- **Rendu par chaînes HTML** (`build*HTML()`) réinjectées dans `#sidebar` / `#main` / `#modal-root`. Toujours passer les données utilisateur par `escapeHtml`.
- **Délégation d'événements globale** : un seul `onGlobalClick` dispatche sur `data-action` (+ `data-*` pour les paramètres). Ajouter un bouton = ajouter un `data-action` et un `case`.
- **Routage par hash** (`#/inbox`, `#/projects`, `#/p/{id}`, cartes et tâches) : `navigateTo()` change le hash, `applyRoute()` applique l'état au `hashchange`.
- **i18n** : chaque chaîne visible passe par `t('cle')` / `tCount()`, et **doit exister dans les deux dictionnaires `fr` et `en`** de `I18N` (le fallback est `fr`). Langue détectée du navigateur, surchargée par les Settings et `localStorage`.
- **État** : hydraté par `GET /api/state`, maintenu par SSE (`EventSource`), re-fetch complet à la reconnexion.
- Confirmations en deux temps génériques (`data-action="confirm-click"`) pour la sync de l'espace de travail, le refus d'une tâche et les suppressions : le bouton devient « Confirmer ? » avant d'agir. La livraison d'un chantier n'utilise pas ce mécanisme : son récapitulatif **est** la confirmation (voir `openShipModal`).
- Aucune image binaire dans `web/` : la marque (`.brand-mark`, barre latérale et page de connexion) est le logo `docs/logo-sillage.png` redessiné en SVG pixel par pixel, en data-URI dans `style.css`. Même principe pour les icônes de bouton (`shipIconHTML`, `anchorIconHTML`) : SVG inline en `currentColor`, jamais d'emoji (illisible sous 16px) ni de police d'icônes.
- Barre de livraison du chantier (`buildShipBarHTML`) : l'état du bouton vient de la carte (`shipReady`/`shipBlocker`, à jour via SSE), l'annonce et les compteurs de commits viennent de `GET /api/cards/{id}/delivery`, rechargé à l'ouverture, à chaque changement de statut et toutes les 60 s (`syncDeliveryPolling`).
- **Sortie de commande = terminal** (`terminalHTML`, `ansiToHTML`) : journal de vérification et journal de recette ont le même bandeau (état, commande, copier) sur fond d'encre, les couleurs ANSI de la commande, sinon les lignes d'échec en rouge et de succès en vert ; curseur clignotant tant que ça tourne. Tout ce qui est « en cours » (agent, relecture de la Vigie, vérification, étape d'outil) porte la classe `.shimmer`, et les outils d'un agent se lisent en chaîne (`.live-steps`).
- **Dictée** (`toggleDictation`) : bouton « Parler » du composeur, Web Speech API du navigateur, absent sans l'API ou hors contexte sûr. Rien ne passe par Sillage, mais Chrome envoie l'audio à son service de reconnaissance : c'est le navigateur qui appelle le réseau, pas le serveur.
- **Micro-interactions de la conversation**, toutes dans le prolongement du rendu incrémental (le fil n'est jamais reconstruit sur un événement SSE) :
  - un message arrivé en direct monte de quelques pixels (`.msg-in`, posée par `appendMessageToConversationDOM` seulement) ; le fil reconstruit à l'ouverture d'un onglet n'anime rien ;
  - l'attente d'un agent est une ligne au gabarit d'un message au pied du fil (`buildThinkingHTML`), qui porte l'outil en cours ou « Travaille… ». `patchConversationThinking` ne réécrit que sa ligne du bas, sinon l'animation des points repartirait de zéro à chaque événement d'activité ; elle reste le dernier élément du conteneur, les messages s'insèrent avant elle ;
  - un message reçu pendant qu'on lit plus haut ne déplace personne : il allume la pastille « Nouveau message » (`refreshConvJump`, alimentée par un écouteur de défilement en phase de capture pour survivre aux rendus). Son clic est le seul défilement animé de l'interface ;
  - copier un message d'agent ou un bloc de code se révèle au survol (opacité seule) et se confirme par une coche (`copyToClipboardIcon`) ;
  - le composeur grandit avec le texte (`autoGrowComposer`) et son bouton d'envoi s'efface quand il n'y a rien à envoyer.
- **Le sillage des deux gestes qui valident du travail** (accepter une tâche, livrer un chantier), la seule animation du produit déclenchée par un succès et non par un clic :
  - une onde verte traverse une fois les surfaces qui viennent de changer (`playWake`, classe `.wake` et `@keyframes om-wake`) : la ligne de la tâche et le bandeau d'état du détail pour une acceptation (`playAcceptWake`), l'emplacement de livraison et la barre du chantier pour une livraison (`playShipWake`). C'est un fond animé sur un `::after` en `inset 0`, jamais un enfant déplacé : rien ne déborde, rien ne se relayoute, et une animation qui ne part pas finit invisible ;
  - l'onde se joue **après** le rafraîchissement qui a reconstruit les surfaces, pas au retour du POST : l'aperçu de livraison arrive juste derrière et remplace la ligne, l'en-tête du détail et la barre du chantier. D'où le drapeau `pendingWake`, armé par `armWake` et consommé à la fin de `loadDelivery`, périmé au bout de trois secondes ;
  - les deux gestes ne se recliquent pas pendant leur aller-retour : `state.acceptingByTask` fige les boutons de la ligne et le bouton primaire du détail (« Acceptation… »), et le bouton de la modale de livraison passe en `.btn-sailing` (« Livraison… », son bateau au rythme d'une houle) puisqu'un push peut prendre plusieurs secondes.

### Le design (`web/style.css`)

Verre translucide, et le calme comme critère de décision : devant deux options, prendre celle qui demande le moins d'attention. Trois couches, décrites en tête de `style.css`, à respecter pour toute surface ajoutée :

1. **L'ambiance** : le fond de la page (`body`), cinq taches colorées et **resserrées** sur `--bg`, en `background-attachment: fixed`. Deux d'entre elles tombent dans la colonne de gauche, pour que la barre latérale ait une teinte qui varie du haut vers le bas. Deux pièges, faits une fois chacun :
   - `--bg` doit être franchement plus soutenu que les surfaces. Du verre posé sur un fond aussi clair que lui donne un écran délavé où plus rien ne se distingue.
   - **Les taches doivent avoir des bords.** Des lavis très étalés, une fois floutés, redonnent exactement le même dégradé : le `backdrop-filter` n'a alors rien à réfracter et la surface reste un aplat, quel que soit le rayon du flou. C'est la raison pour laquelle la barre latérale n'avait « pas d'effet glass ». Les taches sont donc courtes et colorées ; c'est le verre par-dessus qui les rend discrètes, pas leur pâleur.
2. **Le verre**, par densité croissante : `--glass-rail` (la barre latérale, la plus fine : elle porte la navigation, pas le travail), `--glass` (le panneau `.main`), `--glass-card` (cartes et champs posés dessus), `--glass-float` (modales, 0.88 minimum : en dessous le texte de derrière fantôme au travers), `--glass-menu` (menus, 0.97 : un menu se pose directement sur du texte, sans le voile assombri d'une modale pour l'aider). Trois choses font qu'une surface se lit comme du verre, et le flou n'est que la troisième : `--sheen` (le reflet en diagonale, posé en `background-image` par-dessus la couleur de verre), `--glass-edge` (l'arête claire en ombre interne sur les bords haut et gauche), puis `backdrop-filter: var(--blur)`. Plus un trait `--border` autour et `border-top-color: var(--glass-line)`. Les deux panneaux (barre latérale et `.main`) sont séparés par une rainure d'ambiance de 10 px : c'est elle, plus qu'un trait, qui donne au verre son épaisseur, et c'est en comparant la tache brute dans la rainure à sa version filtrée sous le panneau qu'on voit le verre.
3. **L'encre** : le texte, les traits (`--border*`, `--hover`, tous des `rgba` d'encre diluée, jamais une couleur, pour se poser sur n'importe quelle densité de verre) et les voiles de sens (`--green-bg`, `--amber-bg`, `--red-bg`, translucides eux aussi).

Les règles qui en découlent :

- **Le mouvement est réservé à ce que l'utilisateur déclenche.** Rien ne bouge au survol (pas de `transform` sur un bouton, une carte ou une ligne : seules la teinte et l'ombre changent) parce que survoler n'est pas une intention ; un appui, si, donc tout ce qui se clique s'enfonce d'un `scale()` (voir « Micro-interactions »), et une surface qui arrive monte de quelques pixels (`om-rise`, modales et menus) parce qu'elle n'était pas là. Les animations de fond (`om-pulse`, `om-spin`, `om-fade`) sont lentes, et `prefers-reduced-motion` désactive tout, y compris l'enfoncement.
- **Le bloc « Transitions des éléments cliquables » reste le dernier du fichier.** `transition` est un raccourci qui repose `transition-property` en entier : toute règle plus bas qui en déclare un ferait disparaître `transform` de la liste et rendrait l'enfoncement sec.
- **Pas de couleur pour décorer.** Le noir (`--solid`) pour les boutons pleins, le bleu `--accent` réservé aux liens et aux anneaux de focus, le vert au « livré », l'ambre au « en retard ». Un `@supports not (backdrop-filter)` remonte l'opacité des surfaces là où le verre n'existe pas.
- **Une échelle de boutons, toujours en pilule** (`border-radius: 999px`) : 32 px dans les barres et les modales, 44 px pour l'action qui conclut un écran (accepter une tâche). Le bouton plein est un dégradé noir avec liseré (`--solid-fill`, `--solid-lift`), pas un aplat. `.btn-neutral` est le même bouton secondaire que `.btn-outline` : un aplat gris se lisait comme désactivé.
- **Police du système** (`--font-ui` : SF, puis Inter, Segoe, `system-ui`) et `--font-mono` pour tout le code. Les gris de texte tiennent 4,5:1 jusqu'à `--muted-3` ; `--muted-5` n'écrit jamais de texte à lire.
- **Icônes d'interface en SVG au trait** (`uiIconHTML`, grille 16, trait 1.5) : jamais de glyphe unicode (⌂ ◎ ✎ ⚙), dont la graisse et la ligne de base changent d'une police à l'autre.

### Ce que les écrans ne montrent pas

Épurer est une décision produit, pas un réglage de CSS. Les choix en vigueur, à ne pas défaire par inadvertance en rajoutant « juste une information » :

- **La barre latérale ne liste que les agents qui travaillent, ou qui sont mal installés alors qu'ils ont déjà eu une tâche**, sous le titre « Au travail », et la section disparaît quand il n'y en a aucun. Les CLI jamais employés et mal installés tiennent sur une ligne (« 3 agents à installer ») qui ouvre Réglages → Agents : un avertissement permanent pour un outil qu'on n'utilise pas, c'est du bruit. La liste complète vit dans Réglages → Agents (`buildSettingsAgentsHTML`), qui est aussi le seul endroit pour en créer un. Sept lignes d'agents au repos, c'était la moitié de la barre pour une information qu'on ne relit jamais.
- **Une carte de kanban ne porte que l'avancement et ce qui attend un humain.** Pas de compteur de livrables ni de messages : ils ne décident de rien. Le rail de progression n'apparaît qu'au-dessus de 0 %.
- **L'accueil est la boîte de réception**, pas la liste des projets, et les signaux prêts de la Vigie y remontent en tête (`buildInboxSignalsHTML`) : une seule porte pour « qu'est-ce qui m'attend ». Les filtres vides n'y apparaissent pas.
- **Sur un grand écran, le contenu se tient au centre** (`.page-col`, padding calculé de la boîte de réception, `.kanban-body > *`) : une colonne collée à gauche sur 2560 px, c'est 2000 px de vide à traverser. La fiche d'un signal a deux colonnes : la lecture à gauche, la décision à droite qui reste visible au défilement (`.sd-aside`), avec la carte GitHub sous le panneau.
- **L'accueil salue et résume en une phrase** (`buildInboxHelloHTML`) plutôt que d'aligner des compteurs : une phrase se lit, des chiffres se déchiffrent.
- **Le kanban se lit « à qui la main »** (`LANE_ORDER`, `cardLane`, `signalLane`) : À lancer, Les agents travaillent, À vous, Livré. Avec des agents, le goulot n'est plus l'avancement mais l'attention humaine. Les PR et issues de la Vigie y sont des cartes comme les chantiers, et s'ouvrent sur leur fiche (`buildSignalDetailHTML`, route `#/p/{id}/s/{signalId}`) ; pas de bloc à part au-dessus des colonnes.
- **Le jugement d'un signal choisit son bouton principal** (`signalPrimaryAction`) : « Merger sur GitHub » pour une PR mergeable, « Copier les commentaires » quand il y a des retours, « Confier à un agent » pour une issue à corriger ou un constat d'entretien, avec « En faire un chantier » à côté (le chantier porte le constat en contexte, puis la planification s'ouvre). Un seul bouton plein par signal.
- **Un agent qui travaille montre qu'il avance** : un compteur qui tourne (`elapsedHTML`, mis à jour chaque seconde), ses derniers outils (`.live-steps`), le délai d'abandon et un bouton Arrêter. « Muse relit » sans rien d'autre ne distingue pas un agent lent d'un agent planté. Un signal en file dit combien passent avant lui. Ces zones se mettent à jour en place (`patchSignalLive`), sans reconstruire l'écran.
- **Les sessions Claude Code lancées à la main se voient, sans devenir des tâches** (`buildSessionsPillHTML`, sous le titre du projet, puis une modale) : elles n'ont ni worktree ni diff à relire. Reprendre se fait dans le terminal (commande copiée), ou en en faisant un chantier neuf.
- **La Vigie suit la discussion et l'entretien** : un commentaire d'un tiers relance la relecture (pastille « N nouveaux commentaires »), et la veille hebdomadaire rend des cartes « Suggestion » qui se transforment en tâche ou en chantier. Une PR mergée ou une issue fermée sur la forge sort d'elle-même de la Vigie (`closeGoneSignals`) et se range dans « Livré » : personne n'a à cocher « terminé ».
- **La configuration automatique propose, l'humain enregistre** (`buildAutoconfSectionHTML`, en tête de l'onglet Général des réglages ; invitation sous le titre d'un projet qui n'a ni vérification ni recette). « Reprendre ces valeurs » remplit le brouillon, Enregistrer écrit.
- **Le diff se lit comme une revue** (`buildDiffHTML`) : tous les fichiers à la suite sous des en-têtes collants, numéros de ligne des deux côtés, fragment modifié surligné, écarts entre hunks annoncés ; l'arbre des fichiers est une colonne quand le panneau est large, un menu sinon (requête de conteneur, pas de rendu). Fichiers générés (lockfiles) et très longs repliés d'office. « Ouvrir dans <outil> » ouvre toutes les modifications (`--dir-diff`), chaque en-tête a le sien pour un fichier. Raccourcis J/K (fichiers), F (agrandir). Agrandir et replier se font **en place** (classe posée sur l'élément existant) : un `renderMain` couperait l'animation.
- **La colonne « Livré » ne grandit pas sans fin** : « Archiver » dans son en-tête range d'un coup les chantiers livrés et les PR et issues fermées sur la forge, et un chantier fini (livré ou abandonné) s'archive aussi un par un, depuis le « ⋯ » de sa carte ou de son titre. Les archives se replient sous la colonne (`Card.archivedAt`, rien n'est détruit), et du travail nouveau en fait ressortir le chantier. Supprimer un chantier est dans le même « ⋯ » du titre, plus dans la modale d'édition.
- **Une ligne de tâche ne montre aucun zéro** (`buildTaskCountsHTML`). « 0 nouveaux commits » sur chaque ligne, c'est du texte à sauter des yeux à chaque fois.
- **Une action rare ne prend pas la place d'une action fréquente.** Supprimer une tâche, ou la faire « Repartir de zéro » (code remis à la pointe du chantier, session oubliée, consigne d'origine remise dans le composeur), est dans le menu « ⋯ » de l'en-tête du détail, pas en rouge au-dessus des onglets ; Refuser est un bouton à côté d'Accepter, pas un lien sous lui.
- **Deux boutons voisins ne portent jamais le même libellé** : la recette du chantier et celle de la tâche sont visibles en même temps, d'où `preview.buttonCard` (« Recette du chantier ») face à `preview.button` (« Recette »).

## Conventions de langue

- Commentaires Go et specs `docs/` : **français**.
- Messages d'erreur de l'API (`{"error": "..."}`) : **anglais**, courts (`"confirmation required"`, `"task not found"`).
- Chaînes visibles de l'UI : jamais en dur, toujours via `t()` avec les clés `fr` + `en`.
- README, CONTRIBUTING, CHANGELOG : anglais (`README.fr.md` est la version française du README).
- Pas de tiret cadratin ni demi-cadratin comme ponctuation.

## Tests

Tout est dans `internal/server/*_test.go`, sans dépendance de test externe. Deux patrons à réutiliser :

- `NewStore(t.TempDir())` pour les tests de store (roundtrip, compteurs dérivés, transitions de statut, validations).
- Un vrai dépôt git créé dans `t.TempDir()` via le helper `runTestGit` pour les tests git et workspace (worktree + diff, clone, sync, conflit de rebase, « SyncPush ne touche jamais un dépôt de projet »).
- `newDeliveryFixture` (git_test.go) pour tout ce qui touche la livraison : dépôt git réel avec remote bare, serveur, projet et chantier prêts, plus les helpers `addTask` / `accept` / `ship` / `delivery`. Les tests couvrent le flux accepter → livrer, le conflit de fusion, les quatre modes de livraison (`merge` ne pousse jamais, `merge-push` pousse la branche de destination et refuse une divergence réelle, `push` ne crée aucune pull request), l'acceptation automatique d'une branche fusionnée à la main et la relivraison après travail nouveau.

Le parser de diff se teste sur une fixture inline (`TestParseDiffFixture`).

## Données à l'exécution

`~/.local/share/sillage/` : `sillage.lock` (verrou d'instance et adresse du serveur en cours), `state.json` (tout l'état), `config.json` (`passwordHash`), `worktrees/<taskId>/` (un worktree git par tâche), `worktrees/ws-<cardId>-<repo>/` (un par branche de chantier), `worktrees/.merge-<cardId>-<repo>/` (transitoire, mode de livraison « fusion locale »), `attachments/<taskId>/` (les images jointes aux messages, supprimées avec la tâche), `package-cache/<taskId>/<profil>/` (caches de toolchain hors diff, supprimés avec la tâche), `agent-homes/<taskId>/copilot/` (dossier de configuration isolé de Copilot, où vit sa session, supprimé avec la tâche), et optionnellement un dépôt git de l'espace de travail qui ne versionne que `state.json`, `config.json` et `.gitignore`.
