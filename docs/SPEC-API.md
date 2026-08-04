# Sillage : contrat d'API (v0.3.7)

Serveur Go sur `:8787`. Frontend statique servi sur `/`. Tout le JSON est en camelCase.

Chaque fichier statique est servi avec un `ETag` (empreinte de son contenu embarqué, calculée au démarrage) et `Cache-Control: no-cache` : le navigateur revalide à chaque chargement et reçoit `304 Not Modified` tant que le contenu n'a pas changé. Sans ces en-têtes, `embed.FS` ne fournissant aucune date de modification, la réponse n'aurait aucun validateur et le navigateur garderait indéfiniment son ancienne copie de `app.js` après un rebuild.

Auth par cookie de session (`sillage_session`, HttpOnly). Toute route `/api/*` (sauf `/api/login`) renvoie `401` sans session valide : le frontend affiche alors l'écran de connexion. Le mot de passe est optionnel : sans `SILLAGE_PASSWORD` défini au démarrage (ni hash existant dans `config.json`), aucun mot de passe n'est configuré et le serveur ne demande aucune authentification (toute route `/api/*` répond directement, jamais de `401`).

Mono-utilisateur : un seul mot de passe partagé (pas de comptes, pas de rôles). Tous les messages d'erreur API (`{"error": "..."}`) sont en anglais, courts (ex : `"confirmation required"`, `"task not found"`, `"invalid project path: not a git repository"`).

## Modèles

```jsonc
Tokens  { "input": 0, "output": 0, "costUsd": 0.0 }

Repo    { "name": "api", "path": "/abs/path", "previewCmd": "make serve PORT=$((4000 + SILLAGE_N))",
          "previewUrl": "http://127.0.0.1:$((4000 + SILLAGE_N))" }
          // name court et unique dans le projet.
          // previewCmd : commande de recette manuelle, lancée dans le worktree d'un chantier
          // ou d'une tâche (voir "Recette manuelle"). Vide = pas de recette pour ce dépôt.
          // previewUrl : optionnelle, http(s) uniquement, mêmes variables que la commande.

Link    { "url": "https://github.com/org/repo", "title": "org/repo" }
          // title fourni par l'utilisateur, ou récupéré best-effort (voir plus bas), ou nom d'hôte.

Delivery { "mode": "pr|push|merge|merge-push", "target": "main", "stackedPrs": false }
          // Ce que « livrer » veut dire dans ce projet (voir "Livraison d'un chantier") :
          //   pr         pousse la branche du chantier, puis ouvre la pull/merge request
          //   push       pousse la branche du chantier, sans rien ouvrir
          //   merge      fusionne dans target, en local, sans jamais pousser
          //   merge-push fusionne dans target puis pousse target
          // target vide = branche par défaut du dépôt. stackedPrs : réservé, ignoré.
          // target n'est pas qu'une destination de livraison : c'est aussi la branche
          // d'où partent les branches de chantier (voir "Livraison d'un chantier").
          // C'est donc la branche de référence du projet, présentée comme telle dans
          // l'UI (« Branche de base », onglet Général) et non sous la livraison.

Project { "id": "p1", "name": "sillage", "description": "...", "repos": [Repo, ...],
          "links": [Link, ...], "unread": 2,
          "tokens": Tokens, "checkCmd": "go test ./...", "contextPrompt": "...",
          "delivery": Delivery, "deliveryWarning": "..." }
          // description : une phrase, affichée sous le nom. checkCmd/contextPrompt/description peuvent être vides.
          // links : au plus 12, http(s) uniquement (voir "Liens épinglés" ci-dessous).
          // deliveryWarning : calculé à chaque lecture (jamais persisté), vide si tout va
          // bien ; voir "Santé de la livraison" ci-dessous.

CardBranch { "repoName": "api", "branch": "sillage/ws-101-refonte-auth", "base": "main",
             "worktreeDir": "...", "prUrl": "", "shippedAt": null }
          // Branche de feature du chantier sur un dépôt (une par dépôt touché).
          // shippedAt : livré ET rien de nouveau depuis ; remis à null dès qu'une tâche
          // est ajoutée au chantier ou qu'une acceptation ajoute des commits.
          // prUrl : conservée après cette remise à zéro (la PR existe toujours).

Card    { "id": "c1", "projectId": "p1", "ref": 101, "column": "soon|doing|done", "title": "...",
          "tasksTotal": 4, "tasksDone": 1, "docsCount": 2, "messagesCount": 12,
          "reviewCount": 1, "progress": 25, "liveActivity": "..." | null,
          "branches": [CardBranch, ...], "shipReady": false,
          "shipBlocker": "|no-tasks|nothing-accepted|nothing-to-ship",
          "contextPrompt": "..." }
          // Card = chantier (vocabulaire produit) ; nom technique inchangé. contextPrompt :
          // texte libre transmis aux agents (voir plus bas), peut être vide.
          // ref : référence courte du projet (compteur partagé avec les tâches), utilisée
          // dans le nom de la branche du chantier.
          // shipReady/shipBlocker : dérivés, état du bouton de livraison (voir plus bas).

Agent   { "id": "bolt", "name": "Bolt", "emoji": "🐝", "color": "#f2b705",
          "model": "claude-sonnet-5", "cli": "claude", "contextPrompt": "...",
          "active": true, "warning": "..." }
          // active = une tâche running lui est assignée. warning : calculé à chaque
          // liste d'agents (jamais persisté dans state.json), vide si tout va bien ;
          // voir "Santé des agents" ci-dessous.

Task    { "id": "t1", "cardId": "c1", "projectId": "p1", "ref": 482, "title": "...",
          "agentId": "bolt", "repoName": "api", "branch": "sillage/482-slug",
          "status": "running|review|accepted|cancelled",
          "messagesCount": 5, "filesCount": 3, "docsCount": 1,
          "checks": [ { "label": "go test", "ok": true } ],   // [] si aucun
          "liveActivity": "Edit · internal/server/store.go" | null,
          "unread": true, "updatedAt": "2026-08-02T10:00:00Z", "tokens": Tokens,
          "rebasing": false }
          // repoName : nom du Repo du projet utilisé pour le worktree
          // rebasing : un rebase automatique de cette tâche est en cours (voir
          //   "Rebase automatique après une acceptation"). État volatile, remis à false au
          //   chargement de state.json. N'affecte jamais updatedAt.

Message { "id": "m1", "taskId": "t1", "author": "user|agent", "authorName": "Bolt",
          "text": "markdown...", "createdAt": "..." }
          // authorName : nom de l'agent pour author="agent" ; displayName des Settings pour
          // author="user" (vide si non renseigné, le frontend affiche alors "Vous"/"You")

WorkspaceStatus { "setupDone": true, "gitEnabled": true, "remote": "git@host:org/repo.git",
                  "dirty": false, "lastCommitAt": "..." | null, "lastSyncAt": "..." | null,
                  "autoSync": false, "lastSyncError": "" }
                  // autoSync : persisté (Workspace.autoSync). lastSyncError : en mémoire
                  // uniquement (jamais persisté, remis à "" à chaque redémarrage) ; voir
                  // "Synchronisation automatique" ci-dessous.

Settings { "displayName": "Ada", "lang": "fr" }   // lang : ""|"fr"|"en"

PreviewRun { "id": "pv1", "projectId": "p1", "cardId": "c1", "taskId": "",
             "repoName": "api", "cmd": "make serve PORT=4101",
             "url": "http://127.0.0.1:4101", "dir": ".../worktrees/ws-c1-api",
             "status": "running|exited|stopped|failed", "exitCode": 0, "error": "",
             "startedAt": "...", "endedAt": null }
          // Une exécution de recette manuelle (voir "Recette manuelle" ci-dessous).
          // taskId vide = run de chantier. cmd et url sont ce qui a réellement été lancé et
          // affiché (variables substituées). status : exited = sortie naturelle (exitCode),
          // stopped = arrêt humain, failed = n'a pas démarré (error).
          // JAMAIS persisté : ni dans state.json, ni dans l'espace de travail git. Un
          // redémarrage du serveur ne laisse ni process ni run.
```

`Card.progress` = tasksDone/tasksTotal en % (0 si vide). Les compteurs de Card sont calculés côté serveur. `tasksDone` = tâches `accepted`. `reviewCount` = tâches `review`. Les tâches `cancelled` (refusées) sont exclues de `tasksTotal`/`tasksDone`/`progress`, mais comptent comme terminales pour l'auto-déplacement de carte (voir plus bas).

`Project.repos` : un projet regroupe un ou plusieurs dépôts git. `Repo.name` est unique dans le projet (défaut : basename du chemin si omis en entrée). L'ancien champ `path` (v0.1/v0.2, un seul dépôt) n'est plus exposé dans le modèle mais reste accepté en entrée de `POST /api/projects` (voir ci-dessous) ; les projets existants migrent automatiquement au chargement (`repos = [{name: basename(path), path}]`).

`Project.contextPrompt` et `Card.contextPrompt` (le contexte du chantier de la tâche), s'ils sont non vides, sont transmis à l'agent lors du lancement d'une tâche (voir runner.go). Chaque bloc n'est ajouté que s'il est non vide :
- claude : ajoutés à `--append-system-prompt`, séparés par des lignes vides, dans cet ordre : contexte de l'agent, puis `Project context:\n<project.contextPrompt>`, puis `Workstream context:\n<card.contextPrompt>`.
- codex : préfixent le prompt utilisateur, dans le même ordre (sans le contexte agent, propre à claude) : `Project context:\n<project.contextPrompt>\n\nWorkstream context:\n<card.contextPrompt>\n\n---\n\n<prompt>`.
- fake : ignorés.

## Endpoints

| Méthode | Route | Corps | Réponse |
|---|---|---|---|
| POST | `/api/login` | `{password}` | 204 ou 401 `{error}` |
| POST | `/api/logout` | | 204 |
| GET | `/api/state` | | `{projects, cards, tasks, agents, workspace, settings, previews, tokens:{global:Tokens}}` |
| GET | `/api/workspace` | | WorkspaceStatus |
| POST | `/api/workspace/setup` | `{mode:"local"\|"init"\|"clone", remote?}` | WorkspaceStatus (voir ci-dessous) |
| PATCH | `/api/workspace` | `{remote?, autoSync?}` | WorkspaceStatus (au moins un des deux champs requis ; `remote` définit/remplace origin, 400 si git non initialisé ; `autoSync:true` exige git initialisé ET un remote déjà défini, celui fourni dans le même appel compte, 400 sinon ; voir "Synchronisation automatique" ci-dessous) |
| POST | `/api/workspace/sync` | `{confirm:true}` | `{output, lastSyncAt}` (voir ci-dessous). **Validation humaine obligatoire** : refus 400 sans `confirm` |
| PATCH | `/api/settings` | `{displayName?, lang?}` | Settings (`lang` doit être `""`, `"fr"` ou `"en"`, 400 sinon) |
| POST | `/api/agents` | `{name, emoji?, color?, cli, model?, contextPrompt?}` | Agent (name et cli requis ; cli ∈ {claude, codex, fake} ; id = slug du name, 400 si déjà pris) |
| PATCH | `/api/agents/{id}` | mêmes champs, tous optionnels | Agent |
| DELETE | `/api/agents/{id}` | | 204 (400 si une tâche référence encore l'agent) |
| POST | `/api/projects` | `{name?, path}` ou `{name?, repos:[{name?,path}, ...], description?, contextPrompt?, links?:[{url,title?}, ...], delivery?}` | Project (400 si aucun dépôt, un path invalide/pas un dépôt git, noms de repo dupliqués, mode de livraison inconnu, ou lien invalide/en trop grand nombre). Seul le chemin d'un dépôt est requis : `name` absent ou vide = basename du premier dépôt, `delivery` absent = déduit des remotes des dépôts (voir « Livraison d'un chantier ») |
| PATCH | `/api/projects/{id}` | `{name?, description?, checkCmd?, contextPrompt?, repos?, links?, delivery?}` | Project (repos/links, si fournis, remplacent la liste entière ; retirer un repo ne casse pas les tâches existantes ; `previewCmd`/`previewUrl` se posent sur chaque Repo, et une `previewUrl` non http(s) est refusée en 400) |
| DELETE | `/api/projects/{id}` | `{confirm:true}` | 204 (voir « Suppressions » ci-dessous). **Validation humaine obligatoire** : refus 400 sans `confirm` |
| POST | `/api/cards` | `{projectId, title, column?, contextPrompt?}` | Card. `column`, si fourni, doit valoir `"soon"` (400 `"cards are created in the soon column"` sinon) : les cartes se créent toujours dans "Bientôt" |
| PATCH | `/api/cards/{id}` | `{column?, title?, contextPrompt?}` | Card. `title`, si fourni, doit être non vide. Le déplacement manuel de colonne (toutes colonnes acceptées) reste indépendant de l'auto-déplacement (voir plus bas) |
| DELETE | `/api/cards/{id}` | `{confirm:true}` | 204 (voir « Suppressions » ci-dessous). **Validation humaine obligatoire** : refus 400 sans `confirm` |
| GET | `/api/cards/{id}/delivery` | | DeliveryPreview : ce que la livraison ferait, avant tout clic (voir « Livraison d'un chantier »). Lecture seule |
| POST | `/api/cards/{id}/ship` | `{confirm:true}` | ShipResponse : **la seule action sortante du produit** (push + pull request, ou fusion locale). 400 sans `confirm`, 409 si le chantier n'est pas livrable |
| POST | `/api/cards/{id}/catch-up` | | CatchUpResponse : fusionne la branche de destination dans celle du chantier pour débloquer la livraison (voir « Livraison d'un chantier »). Local, aucune confirmation ; un conflit est annulé et rapporté par dépôt |
| POST | `/api/tasks` | `{cardId, title, agentId, prompt?, repoName?}` | Task (créée `running`, agent lancé avec title+prompt). `repoName` optionnel si le projet n'a qu'un repo ; sinon 400 `"repoName required (project has several repositories)"` ou 400 si inconnu |
| PATCH | `/api/tasks/{id}` | `{agentId}` | Task : réassigne l'agent (voir « Réassignation » ci-dessous). 400 si `status=running` (`"interrupt the agent before reassigning"`) ou si l'agent est inconnu |
| DELETE | `/api/tasks/{id}` | `{confirm:true}` | 204 (voir « Suppressions » ci-dessous). **Validation humaine obligatoire** : refus 400 sans `confirm` |
| GET | `/api/tasks/{id}` | | `{task, messages}` |
| POST | `/api/tasks/{id}/messages` | `{text}` | 202 ; relance l'agent (statut → running) |
| POST | `/api/tasks/{id}/interrupt` | | Task (running → review) |
| POST | `/api/tasks/{id}/accept` | | `{task, workstreamBranch, output}` : commite le worktree de la tâche puis fusionne sa branche dans celle du chantier. Accepté depuis `review` uniquement (400 sinon). **Aucune confirmation** : action locale, réversible par `/reopen`. 409 en cas de conflit (voir ci-dessous) |
| POST | `/api/tasks/{id}/cancel` | | Task (running/review → cancelled, c'est le « refuser » de l'UI). Si `running`, l'agent est d'abord interrompu (même mécanique qu'`/interrupt`). 400 depuis accepted/cancelled |
| POST | `/api/tasks/{id}/reopen` | | Task (accepted/cancelled → review). 400 sinon |
| POST | `/api/tasks/{id}/read` | | 204 (unread=false, ne modifie jamais `updatedAt` : voir plus bas) |
| GET | `/api/tasks/{id}/diff` | | voir ci-dessous |
| GET | `/api/tasks/{id}/deliverables` | | voir ci-dessous |
| POST | `/api/cards/{id}/preview` | `{repoName?}` | PreviewRun : lance la recette dans le worktree de la branche de chantier. `repoName` optionnel si le chantier n'a qu'une branche ; 400 `"repoName is required"` sinon, 400 si aucune branche de chantier ou aucune commande configurée. **Aucune confirmation** : local et réversible |
| POST | `/api/tasks/{id}/preview` | | PreviewRun : même chose dans le worktree de la tâche |
| POST | `/api/previews/{id}/stop` | | 204 (SIGINT au groupe de process, SIGKILL après 5 s). 404 si le run est inconnu ; sans effet si déjà terminé |
| GET | `/api/previews/{id}/log` | | `{runId, lines: ["..."]}` : le tampon du journal (2000 dernières lignes), la suite arrive en SSE |
| GET | `/api/events` | | SSE |

### Cycle de vie des tâches et auto-déplacement de carte

Statuts : `running → review → accepted`, plus `cancelled` (via `/cancel`, le « refuser » de l'UI). Une tâche ne se livre plus seule : elle est acceptée (fusionnée) dans la branche de son chantier, et c'est le chantier qui se livre (voir « Livraison d'un chantier »). `/reopen` accepte accepted/cancelled et ramène en `review` ; le merge déjà fait n'est pas annulé, la prochaine acceptation fusionnera les nouveaux commits.

Compatibilité, au chargement de state.json : `ready` (antérieur à la v0.3.4) migre vers `review` ; `shipped` et `done` migrent vers `accepted`.

Après une acceptation réussie : un Message marqueur est ajouté au fil (`author="agent"`, `authorName=""`, texte figé `"[accepted:<workstreamBranch>]"`). Une acceptation simplement constatée (la branche de la tâche était déjà contenue dans celle du chantier) pose `"[auto-accepted:<workstreamBranch>]"`. En cas de conflit, la fusion est annulée (`git merge --abort`), la tâche **reste** en `review`, un marqueur `"[merge-conflict:<fichiers séparés par des espaces>]"` est ajouté au fil, et la réponse est 409 `{"error":"merge conflict with the workstream branch","conflictFilePaths":"..."}`. Le frontend détecte ces marqueurs et affiche une ligne système localisée.

#### Rebase automatique après une acceptation

Une acceptation fait avancer la branche du chantier, ce qui met en retard les autres tâches du même chantier : chacune conflicterait à son tour à son acceptation, pour une reprise presque toujours mécanique. Une acceptation réussie déclenche donc, **en tâche de fond** (la réponse HTTP n'attend pas), le rebase des autres tâches du chantier sur la branche du chantier.

Quatre garde-fous, dans cet ordre : même dépôt et statut `review` (une tâche `running` appartient à son agent) ; aucun agent en cours pour cette tâche ; réellement en retard (branche du chantier pas déjà contenue) ; worktree propre après commit du travail en attente (`Sillage: <titre>`, le même commit que celui de l'acceptation ; les agents ne commitent pas toujours, et un rebase perdrait leur travail). Les rebases sont sérialisés entre eux.

Pendant l'opération, `Task.rebasing` vaut `true` (événement SSE `task` à l'entrée et à la sortie ; `updatedAt` n'est jamais touché). `POST /api/tasks/{id}/accept` répond 409 `{"error":"task is being rebased, retry in a moment"}` tant qu'il vaut `true`.

Résultat au fil de la tâche : `"[rebased:<workstreamBranch>]"` en cas de succès. En cas de conflit, `git rebase --abort` remet le worktree exactement dans son état d'avant (rien n'est modifié, la branche est inchangée) et le marqueur est `"[rebase-conflict:<fichiers séparés par des espaces>]"` : la reprise redevient l'affaire de l'agent (bouton « Demander le rebase »). Un échec sans conflit ne pose aucun marqueur.

Après chaque changement de statut de tâche, la carte est automatiquement replacée : si elle a au moins une tâche, qu'elles sont toutes terminales (`accepted`/`cancelled`) **et** que le chantier a été livré, `card.column` passe à `"done"` ; sinon elle passe (ou reste) en `"doing"`. La colonne « Terminé » veut donc dire livré, pas seulement relu. Une carte déjà livrée en ressort dès qu'un travail nouveau y apparaît (voir « Continuer un chantier déjà livré »). Les cartes antérieures aux branches de chantier (`branches` vide) sont considérées livrées, pour garder leur colonne historique. Le déplacement manuel (PATCH `/api/cards/{id}`) reste indépendant de cette règle. Chaque changement republie l'événement SSE `cards`.

### Lecture d'une tâche : `updatedAt` inchangé

`POST /api/tasks/{id}/read` (ouverture d'une tâche) ne met jamais à jour `updatedAt` (variante dédiée `Store.MarkTaskRead`, distincte de la mutation générique `UpdateTask` utilisée par toutes les autres actions) : une liste de tâches triée par `updatedAt` ne doit pas se réordonner sous le curseur quand on ouvre simplement une tâche pour la lire.

### Liens épinglés de projet

`Project.links` : au plus 12 liens, http(s) uniquement (400 sinon, `file://` et tout autre schéma refusés). Pour chaque lien envoyé sans `title`, le serveur tente de récupérer le `<title>` de la page à l'enregistrement (POST ou PATCH) : GET avec timeout global de 5 s, lecture plafonnée à 64 Ko, redirections suivies uniquement si elles restent en http(s) (au plus 5). Cette récupération est best-effort et ne bloque ni n'échoue jamais l'enregistrement : en cas d'échec (timeout, erreur HTTP, page sans `<title>`, hôte injoignable...), `title` devient le nom d'hôte de l'URL.

### Réassignation d'une tâche à un autre agent

`PATCH /api/tasks/{id} {agentId}` : refusé si la tâche est `running` (l'interrompre d'abord via `/interrupt`) ou si l'agent est inconnu. Effets : `task.agentId` change, `task.sessionId` est vidé (le nouvel agent ne peut pas reprendre la session CLI de l'ancien), et un Message est ajouté au fil avec `author="agent"`, `authorName=""` et un texte figé `"[reassigned:<agentId>]"` : le frontend détecte ce marqueur et affiche une ligne système localisée (`author`/`authorName` volontairement neutres pour rester i18n-propre côté backend). SSE `task` + `message` + `agents`.

Départ frais : quand la session CLI est vide (lancement initial, ou premier message après une réassignation), le texte envoyé au CLI est préfixé par un rappel minimal : `Task: <title>\n\n<texte>` (pas de préfixe si le texte est vide).

### Suppressions (tâches, chantiers, projets)

Actions destructives, jamais déclenchées depuis les listes : confirmation double clic côté UI ET `{"confirm":true}` côté API (400 `"confirmation required"` sinon, pour les trois routes).

- `DELETE /api/tasks/{id}` : si la tâche est `running`, l'agent est d'abord interrompu (même mécanique que `/interrupt`/`/cancel` : SIGINT puis SIGKILL après 5 s). La tâche et ses messages sont ensuite supprimés du store. Le worktree est retiré du dépôt d'origine (`git worktree remove --force` puis `worktree prune`), **best-effort** : un échec de ce nettoyage n'empêche jamais la suppression de la tâche. La branche git n'est **jamais** supprimée (elle peut avoir été poussée). SSE : `cards` (compteurs recalculés), `agents`, `tokens`, et un événement dédié `taskDeleted`.
- `DELETE /api/cards/{id}` : cascade sur toutes les tâches du chantier (même traitement individuel que ci-dessus). SSE : `taskDeleted` par tâche supprimée, puis `cards`, `agents`, `tokens`, puis `cardDeleted`.
- `DELETE /api/projects/{id}` : cascade sur tous les chantiers du projet, puis leurs tâches (même traitement individuel). Ne publie que `projectDeleted` : le frontend recharge l'état complet (`GET /api/state`) à réception de cet événement, plus simple et plus sûr qu'une longue série d'événements de cascade.
- Les tokens déjà consommés par les tâches supprimées disparaissent des agrégats (`tokens.global`, `tokens.projects`, `tokens.tasks`) : les compteurs reflètent uniquement l'existant. L'événement `tokens` est republié après toute suppression de tâche (isolée ou en cascade).

### Santé des agents

`Agent.warning` (dans `AgentOut`, jamais `Agent` lui-même : le champ n'existe pas dans le modèle persisté) est recalculé à chaque liste d'agents (`GET /api/state`, événement SSE `agents`) :
- `cli=codex` : si `/proc/sys/kernel/apparmor_restrict_unprivileged_userns` vaut `1` et que `SILLAGE_CODEX_SANDBOX` n'est pas définie → `"codex sandbox is blocked on this machine (AppArmor); see README (SILLAGE_CODEX_SANDBOX)"`.
- `cli=codex` ou `cli=claude` : si le binaire correspondant est introuvable dans le PATH → `"<cli> CLI not found in PATH"`.
- Sinon (ou `cli=fake`) : chaîne vide.

### Livraison d'un chantier

Le chantier est une branche de feature : une branche `sillage/ws-<card.ref>-<slug>` par couple (chantier, dépôt), avec son worktree dédié, créée à la première tâche du chantier sur ce dépôt depuis `project.delivery.target` (ou la branche courante du dépôt si `target` est vide). Les branches de tâche partent de cette branche et y sont fusionnées à l'acceptation ; elles ne sont jamais poussées.

**Réglage du projet** (`Project.delivery.mode`), quatre comportements. L'API l'accepte à la création comme à la modification ; l'UI ne le demande qu'après coup (onglet Livraison des réglages du projet), la création s'appuyant sur la déduction décrite plus bas :

- `"pr"` : pousse la branche du chantier puis ouvre la pull request (GitHub, `gh pr create`) ou la merge request (GitLab, `glab mr create`) ; en l'absence du CLI ou en cas d'échec, repli sur une URL de création pré-remplie en lecture seule (`https://<host>/<path>/compare/<base>...<branch>?expand=1`, ou `https://<host>/<path>/-/merge_requests/new?merge_request[source_branch]=...&merge_request[target_branch]=...`). La branche est poussée dans les deux cas : le repli ne dégrade que l'ouverture de la requête.
- `"push"` : pousse la branche du chantier, et s'arrête là. Aucune pull request n'est ouverte, `prUrl` reste vide.
- `"merge"` : fusionne la branche du chantier dans `target` **en local, en fast-forward uniquement, et ne pousse jamais rien** (worktree transitoire dédié ; repli dans le dépôt lui-même seulement si `target` y est la branche courante et l'arbre propre ; 409 sinon).
- `"merge-push"` : comme `"merge"`, mais `target` est **poussée** ensuite. Avant la fusion, `target` est rattrapée depuis `origin` (`git fetch origin <target>` puis fast-forward) : une vraie divergence est un 409 `"target branch has diverged: ..."` **avant** toute écriture, jamais une fusion locale suivie d'un push rejeté. Jamais de `--force`. La branche du chantier, elle, n'est pas poussée.

Un mode inconnu est refusé en 400 (`"invalid delivery mode: must be pr, push, merge or merge-push"`).

À la création d'un projet sans champ `delivery`, le mode est déduit : `"pr"` si tous les dépôts pointent vers une forge connue (github.com, ou un hôte contenant `gitlab`), sinon `"merge"` avec `target` = branche courante du premier dépôt. Le fournisseur n'est jamais persisté : il est redéduit du remote `origin` à chaque opération.

**Conditions de livraison** (`Card.shipReady`, sinon `Card.shipBlocker`) : au moins une tâche, au moins une tâche `accepted`, au moins une branche de chantier. `POST /api/cards/{id}/ship` refuse en 409 avec le même vocabulaire.

**Livraison partielle.** Des tâches encore `running` ou `review` ne bloquent pas la livraison : la branche du chantier ne contient que le travail accepté, donc livrer un chantier inachevé n'envoie jamais rien qui n'ait été relu. On livre ce qui est prêt, on livre le reste ensuite (voir « Continuer un chantier déjà livré »). `counts.pending` de l'aperçu compte ces tâches ; l'UI l'annonce dans la barre de livraison et dans le récapitulatif de la modale, comme une information et non comme un avertissement. La colonne `done`, elle, exige toujours « tout terminal ET livré » : un chantier livré partiellement reste en `doing`.

Deux états supplémentaires ne se calculent qu'avec git, donc dans l'aperçu (`GET /api/cards/{id}/delivery`) et non dans `shipReady`, qui est recalculé sous verrou sans jamais lancer de commande :

- **Chantier arrivé à destination** (`repos[].mergedIntoTarget` partout) : il n'y a plus rien à livrer, que Sillage l'ait fait ou qu'un humain ait fusionné à la main. L'UI remplace le bouton par une ancre et « Déjà sur `<destination>` ».
- **Fusion fast-forward impossible** (modes `merge`/`merge-push`, un dépôt à livrer avec `fastForwardable: false`) : la destination a avancé de son côté, la livraison échouerait en 409 `ErrTargetDiverged`. Le bouton est désactivé et annonce le rattrapage à faire, au lieu de promettre une livraison certaine d'échouer.

Pourquoi « en retard » suffit à bloquer : une fusion fast-forward avance la destination jusqu'à la branche du chantier. Si la branche est en retard, cette avance serait un recul (les commits de la destination disparaîtraient), et git refuse. Être derrière ne veut pas dire que le travail du chantier est déjà arrivé : c'est `mergedIntoTarget` qui répond à cette question, pas `behind`.

**Rattraper la destination** : `POST /api/cards/{id}/catch-up` (aucun corps requis) fusionne la branche de destination **dans** la branche du chantier, un dépôt après l'autre : `git merge <destination>` dans le worktree du chantier, avec pour sujet `Sillage: catch up with <destination>`. La branche du chantier contient alors la destination, donc la livraison redevient fast-forward.

Une fusion, pas un rebase : les branches des tâches acceptées descendent de la branche du chantier, réécrire son historique les laisserait orphelines. Aucun réseau (la destination est une branche locale), donc aucune confirmation, comme `/accept`. Réponse `{card, repos:[{repoName, target, merged, upToDate, conflictFilePaths, output, error}]}` en 200 même en cas de conflit : chaque dépôt porte son résultat.

Un conflit annule la fusion (`git merge --abort`) : le worktree du chantier revient intact, `merged` reste faux et `conflictFilePaths` liste les fichiers. Le serveur ne résout jamais un conflit ; l'UI propose alors de confier le rattrapage à un agent (modale de création de tâche pré-remplie, l'humain choisit l'agent et valide). Un rattrapage réussi ajoute des commits à la branche : `MarkCardBranchPending` est appelé, comme à l'acceptation.

**Acceptation automatique des branches fusionnées à la main.** Une branche de tâche peut être fusionnée dans celle du chantier hors de Sillage (`git merge` dans un terminal). `GET /api/cards/{id}/delivery` le constate au passage : une tâche `review` dont la branche est entièrement contenue dans celle du chantier passe `accepted`, avec un message marqueur `[auto-accepted:<branche du chantier>]` (le frontend affiche une ligne système localisée) et les événements SSE `task`, `message`, `cards`, `agents`. Quatre garde-fous : aucun agent en cours pour la tâche, `filesCount > 0` (une tâche vide est contenue par construction), son worktree propre (du travail non commité n'est par définition pas fusionné), et sa branche effectivement contenue. Aucune écriture git : l'appel ne fait que constater. Le frontend rappelle cet endpoint à l'ouverture de la vue chantier et toutes les 60 s tant qu'elle reste ouverte.

**Continuer un chantier déjà livré.** `CardBranch.shippedAt` signifie « livré, et rien de nouveau depuis ». Il est remis à `null` dès que du travail nouveau apparaît sur ce dépôt : création d'une tâche sur le chantier, ou acceptation qui ajoute réellement des commits à la branche (comparaison de SHA avant/après la fusion). Le chantier quitte alors la colonne `done`, et le bouton de livraison redevient actif dès qu'une acceptation a remis des commits à livrer. `prUrl` est conservée : la pull request existe toujours, une nouvelle livraison la met à jour par un simple push (aucune tentative d'en ouvrir une seconde). Une acceptation automatique (voir ci-dessus) n'ajoute aucun commit et ne remet donc jamais le chantier en « non livré ».

`GET /api/cards/{id}/delivery` (aucune écriture git ; peut accepter des tâches déjà fusionnées, voir ci-dessus) :

```jsonc
{ "mode": "pr", "target": "main", "provider": "github|gitlab|",
  "ready": true, "blocker": "",
  "warnings": ["gh not found in PATH; ..."],
  "counts": { "accepted": 3, "refused": 1, "pending": 1 },
  "repos": [ { "repoName": "api", "branch": "sillage/ws-101-refonte-auth", "base": "main",
               "commits": 4, "files": 11, "pending": 4, "behind": 2,
               "mergedIntoTarget": false, "fastForwardable": true,
               "prUrl": "", "shippedAt": null } ],
  "behind": { "t12": 2 } }
```

`mergedIntoTarget` (la branche du chantier est entièrement contenue dans la branche de destination) et `fastForwardable` (la destination est un ancêtre de la branche, donc la fusion fast-forward passerait) situent le chantier par rapport à sa destination. Les deux se lisent par `git merge-base --is-ancestor`, sans rien écrire.

`commits`/`files` décrivent le contenu de la livraison (`base..branche`) ; `pending` est ce qu'il reste réellement à livrer (commits non poussés dans les modes de branche `pr`/`push`, non encore fusionnés dans `target` dans les modes de fusion `merge`/`merge-push`). `pending` à zéro partout signifie « rien à livrer » : le bouton est inactif et la livraison marque le dépôt `skipped`.

**Retard sur la base** (`behind`, deux niveaux, jamais persistés, calculés par `git rev-list --count`) :

- `repos[].behind` = commits que la base a et que la branche du chantier n'a pas (`branche..base`) : le chantier est en retard sur la release. Ne bloque pas la livraison, l'annonce seulement.
- `behind` (table à la racine, par identifiant de tâche) = commits que la branche du chantier a et que la branche de la tâche n'a pas. Renseignée uniquement pour les tâches `running` et `review` ; une tâche à jour n'y figure pas. C'est le retard qui produit un conflit à l'acceptation, typiquement après l'acceptation d'une autre tâche du même chantier (deux commits par acceptation : celui de la tâche, puis le commit de fusion `--no-ff`).

Une révision manquante (branche jamais poussée, worktree retiré) rend `0` : mieux vaut n'annoncer aucun retard qu'un retard faux.

Règles UI : badge `↓{n}` ambre dans la ligne de tâche et dans l'en-tête du panneau de détail, avec l'explication en infobulle ; ligne de retard du chantier dans la barre de livraison. Le bouton **Demander le rebase** de l'en-tête n'appelle aucun endpoint git : il poste un message dans le fil de la tâche (`POST /api/tasks/{id}/messages`) demandant à l'agent de rebaser, ce qui le met en file si l'agent tourne encore.

Un retard qui subsiste veut donc dire que le rebase automatique de l'acceptation (voir ci-dessus) ne s'est pas appliqué ou a conflicté : agent au travail, tâche pas en revue, ou conflit réel. Le serveur ne rejoue jamais une branche autrement que par ce rebase, et n'y résout jamais un conflit : ça reste l'affaire de l'agent.

`POST /api/cards/{id}/ship` `{confirm:true}` (400 `"confirmation required"` sinon) traite chaque dépôt du chantier et répond :

```jsonc
{ "card": Card, "mode": "pr",
  "repos": [ { "repoName": "api", "branch": "...", "base": "main", "pushed": true,
               "merged": false, "skipped": false, "prUrl": "https://...",
               "output": "...", "error": "" } ] }
```

Un dépôt en échec n'annule pas les autres : chaque ligne porte sa propre erreur, et la livraison est rejouable telle quelle (les dépôts déjà livrés sont `skipped`). `pushed`/`merged` décrivent ce qui a réellement eu lieu : `pushed` seul dans les modes `pr` et `push`, `merged` seul en `merge`, les deux en `merge-push`.

### Santé de la livraison

`Project.deliveryWarning` (dans `ProjectOut`, jamais `Project` lui-même : le champ n'est pas persisté) est recalculé à chaque lecture d'état ou mutation de projet. Vide en mode `merge`, le seul qui ne touche jamais au réseau (aucun binaire externe, aucun remote requis). Sinon, par ordre de priorité : `"no 'origin' remote on repository <name>; nothing can be pushed"` (tout mode qui pousse), puis, en mode `pr` uniquement, `"unknown forge on repository <name>; the branch will be pushed without opening a pull request"`, `"gh not found in PATH; Sillage will fall back to a prefilled pull request URL"`, `"glab not found in PATH; ..."`. Les modes `push` et `merge-push` n'ont besoin d'aucune forge reconnue : ils poussent, point. Aucun de ces avertissements n'empêche quoi que ce soit.

### Espace de travail (synchronisation git de dataDir)

Le répertoire de données (dataDir) peut devenir un dépôt git optionnel, pour sauvegarder/synchroniser `state.json`, `config.json` et `.gitignore` (seuls fichiers versionnés ; `.gitignore` exclut `worktrees/` et `*.tmp`). Un commit local est tenté silencieusement si git est activé, throttlé à au plus un par quart d'heure (`workspaceCommitInterval`) : chaque commit stocke un blob complet de `state.json`, et un agent qui travaille déclenche plusieurs sauvegardes par seconde. La sauvegarde de `state.json` elle-même reste atomique et immédiate à chaque mutation ; le commit n'est qu'un point de restauration. Le dépôt de l'espace de travail est configuré en `gc.auto 256` pour que git compacte souvent. Jamais de push automatique en dehors de la synchronisation automatique optionnelle (`autoSync`, voir ci-dessous), qui reste désactivée par défaut.

- `mode:"local"` : marque `setupDone`, aucun git. Refusé (400) si déjà fait.
- `mode:"init"` : `git init` (branche main), écrit `.gitignore`, premier commit ; `remote` optionnel (`git remote add`, jamais de push). Rejouable plus tard (depuis les réglages) pour activer git sur un espace resté local, même si `setupDone` est déjà vrai.
- `mode:"clone"` (`remote` requis) : clone dans un répertoire temporaire, vérifie que `state.json` existe à sa racine (sinon 400 `"remote does not look like a Sillage workspace"`), puis remplace `state.json`/`config.json`/`.git` de dataDir par ceux du clone et recharge le store en mémoire, sans redémarrage (les sessions actives restent valides). Le mot de passe devient celui de l'espace rapatrié. Refusé (400) si déjà fait.
- `POST /api/workspace/sync` : commit d'abord si nécessaire, puis `git pull --rebase origin main`, puis `git push -u origin main` (fonction dédiée `SyncPush`, qui n'opère jamais sur un dépôt de projet). En cas de conflit de rebase : `git rebase --abort` puis 409 `{"error":"sync conflict: the remote workspace diverged, resolve manually in <dataDir>"}`. Une synchronisation manuelle réussie efface toujours `lastSyncError`, y compris si l'auto-sync était en pause sur conflit (voir ci-dessous).
- Compatibilité : un state.json existant sans champ `workspace` (installation antérieure à la v0.3) migre vers `setupDone=true` en mode local au chargement : l'onboarding ne s'affiche jamais sur un espace déjà utilisé.

### Synchronisation automatique de l'espace de travail (`autoSync`)

`Workspace.autoSync` (persisté, `false` par défaut) active une synchronisation périodique automatique de l'espace de travail (dataDir uniquement, jamais un dépôt de projet), avec la même logique que la synchronisation manuelle (`SyncPush`).

- Activation : `PATCH /api/workspace {autoSync:true}` exige que git soit initialisé ET qu'un remote soit déjà défini (celui fourni dans le même appel compte) ; 400 sinon (`"git must be initialized with a remote before enabling automatic sync"`).
- Une goroutine dédiée (ticker de 15 min) est démarrée au démarrage du serveur si `autoSync` est déjà actif dans state.json, et à chaque activation via PATCH ; elle est arrêtée proprement à la désactivation (`autoSync:false`).
- Chaque tick : si `autoSync` n'est plus actif, ou si une erreur de conflit non résolue est en attente (voir plus bas), le tick ne fait rien. Sinon, il appelle la même logique que `POST /api/workspace/sync` :
  - **Succès** : `lastSyncAt` est mis à jour, `lastSyncError` est vidé, l'événement SSE `workspace` est republié.
  - **Échec** (hors conflit) : `lastSyncError` est renseigné avec le message d'erreur, l'événement SSE `workspace` est republié, et `autoSync` **reste actif** (nouvelle tentative au tick suivant, 15 min plus tard).
  - **Conflit** (`ErrSyncConflict`) : `lastSyncError` est renseigné, l'événement SSE `workspace` est republié, et la synchronisation automatique **se met en pause** : `autoSync` reste `true`, mais les ticks suivants ne tentent plus rien tant que `lastSyncError` signale ce conflit. Seule une synchronisation manuelle réussie (`POST /api/workspace/sync`) efface l'erreur et relance l'auto-sync au tick suivant.
- `lastSyncError` n'est jamais persisté : un redémarrage du serveur repart avec `lastSyncError` vide (mais respecte `autoSync` s'il était actif).

### Recette manuelle

Sillage ne sait rien des stacks : il exécute la commande de recette d'un dépôt (`Repo.previewCmd`, écrite par l'humain dans les réglages du projet) dans le worktree de ce qu'on veut éprouver. Voir `docs/SPEC-RECETTE.md` pour le « pourquoi » et les pistes écartées.

- **Exécution** : `sh -c <previewCmd>`, répertoire courant = le worktree du chantier (`CardBranch.worktreeDir`) ou de la tâche (`Task.worktreeDir`), **jamais le dépôt de travail de l'utilisateur**. Le process est lancé dans son propre groupe (`Setpgid`) pour que l'arrêt tue aussi ses enfants (un `npm run dev` qui lance node, un `make` qui lance un serveur).
- **Un seul run par worktree** : relancer arrête le précédent et attend sa mort avant de démarrer le nouveau (deux serveurs sur le même port ne servent à personne).
- **Pas de genre de commande** : un serveur qui reste en vie et un script qui finit passent par le même chemin. Le statut dit lequel c'était (`running`, puis `exited` avec son `exitCode`).
- **Variables injectées** dans l'environnement, en plus de celui du serveur :

  | Variable | Valeur | Pour |
  | --- | --- | --- |
  | `SILLAGE_ID` | `ws-<cardRef>` ou `t-<taskRef>` | noms : base de données, conteneur, répertoire |
  | `SILLAGE_N` | `<cardRef>` ou `<taskRef>` | arithmétique : `PORT=$((4000 + SILLAGE_N))` |
  | `SILLAGE_DIR` | le worktree (aussi le répertoire courant) | chemins absolus |
  | `SILLAGE_BRANCH` | la branche recettée | affichage, bannière de debug |

  `SILLAGE_ID`/`SILLAGE_N` dérivent de `Card.ref` et `Task.ref` : petits, **stables dans le temps** (la base de recette et son contenu survivent entre deux sessions) et uniques dans un projet, puisque le compteur de références est partagé entre chantiers et tâches. Aucun état nouveau n'est persisté pour ça : ni allocateur, ni table de slots. À charge du projet d'écrire une commande idempotente (créer-si-absent).
- **`previewUrl`** est développée par le shell dans le même environnement que la commande : elle accepte donc exactement la même syntaxe, arithmétique comprise. Elle est validée à l'enregistrement (`http(s)` uniquement, parce qu'elle devient un lien cliquable) et rendue dans `PreviewRun.url` une fois substituée. Pas de sonde de disponibilité : le lien est cliquable dès le lancement.
- **Journal** : tampon circulaire en mémoire de 2000 lignes, stdout et stderr mêlés dans l'ordre d'arrivée. Lu une fois à l'ouverture (`GET /api/previews/{id}/log`), puis complété ligne à ligne par l'événement SSE `previewLog`. Jamais écrit sur disque.
- **Rien ne survit à l'arrêt du serveur** : SIGINT/SIGTERM déclenche l'arrêt de tous les runs (SIGINT au groupe, SIGKILL après 5 s) avant la fermeture du serveur HTTP. Il n'y a **ni plafond ni TTL d'inactivité** : à la place, l'interface affiche en permanence le nombre de recettes en cours, et c'est l'humain qui arrête (un TTL qui coupe un serveur pendant qu'on s'en sert est plus agaçant qu'utile).
- **Sans commande configurée**, il n'y a rien à lancer : l'interface affiche le chemin du worktree avec un bouton de copie. Ce repli est le minimum garanti, disponible sur 100 % des projets sans configuration.

### Diff

```jsonc
{ "branch": "sillage/482-slug", "base": "main",
  "files": [ { "path": "a/b.go", "additions": 10, "deletions": 2,
               "hunks": [ { "header": "@@ -1,4 +1,6 @@",
                            "lines": [ {"type":"ctx|add|del", "text":"..."} ] } ] } ] }
```

### Livrables

```jsonc
{ "code":   [ { "kind": "commit|branch", "title": "...", "meta": "abc1234 · il y a 2 h" } ],
  "docs":   [ { "kind": "doc",   "title": "README.md", "meta": "+12 −0", "path": "README.md" } ],
  "images": [ { "kind": "image", "title": "...", "meta": "...", "path": "..." } ] }
```

### SSE `/api/events`

Événements nommés, `data` = JSON :

- `task` : Task complet (création ou changement d'état/compteurs).
- `message` : Message (nouveau message agent ou user).
- `activity` : `{taskId, line}` : ligne d'activité live (peut être `null` = terminé).
- `tokens` : `{global: Tokens, projects: {projectId: Tokens}, tasks: {taskId: Tokens}}`.
- `cards` : liste des Cards recalculées du projet touché.
- `agents` : liste des Agents (pour l'indicateur d'activité, et après chaque mutation CRUD).
- `project` : Project complet (après PATCH `/api/projects/{id}`).
- `workspace` : WorkspaceStatus (après setup, changement de remote/autoSync, sync manuelle, ou tick d'auto-sync).
- `settings` : Settings (après PATCH `/api/settings`).
- `preview` : PreviewRun (lancement, sortie, arrêt d'une recette manuelle).
- `previewLog` : `{runId, line}` : une ligne de sortie d'un run de recette.
- `taskDeleted` : `{taskId, cardId, projectId}` (après `DELETE /api/tasks/{id}`, une fois par tâche supprimée y compris en cascade).
- `cardDeleted` : `{cardId, projectId}` (après `DELETE /api/cards/{id}`).
- `projectDeleted` : `{projectId}` (après `DELETE /api/projects/{id}` ; le frontend recharge l'état complet plutôt que de rejouer la cascade).

Le frontend maintient son état en mémoire à partir de `/api/state` + SSE ; reconnexion SSE automatique (EventSource natif) et re-fetch de `/api/state` à la reconnexion.

## Règles produit à respecter côté UI

- Push / livraison uniquement via le bouton « Livrer » de l'en-tête du chantier, en deux clics : le bouton ouvre un récapitulatif (mode, dépôts, branche → base, commits, fichiers, avertissements) dont le bouton d'action **est** la confirmation (`{"confirm":true}`). Jamais automatique. Le bouton est grisé, avec la raison en infobulle, tant que `card.shipReady` est faux ou qu'il n'y a rien à livrer. Il reste actif quand des tâches sont encore en cours ou à relire : le récapitulatif annonce alors la livraison partielle (« n tâches ne sont pas encore acceptées ») plutôt que d'interdire d'envoyer ce qui est prêt. Même règle de confirmation explicite pour la synchronisation de l'espace de travail (`/api/workspace/sync`).
- Accepter ou refuser une tâche s'obtient en un clic depuis la liste de tâches (boutons révélés au survol d'une tâche à relire) : ces actions sont locales et réversibles (`/reopen`), donc sans confirmation. L'état de chaque tâche (acceptée, refusée) est lisible dans la liste sans ouvrir la tâche.
- Recette manuelle : le bouton « Recette » est présent dans l'en-tête du chantier (avant « Livrer » : on éprouve, puis on livre) et dans le panneau de détail d'une tâche (action secondaire, la principale reste « Accepter »). Il est **toujours affiché**, même sans commande configurée : le panneau propose alors le chemin du worktree à copier. Lancer et arrêter sont des actions locales, sans confirmation. Une pastille verte sur le bouton, et un compteur en bas de sidebar, disent ce qui tourne.
- Tokens : jamais affichés dans le flux de travail (kanban, détail de tâche, liste des projets), pour ne pas ajouter de charge mentale. Seul endroit visible : Réglages > onglet Statistiques, consommation par projet, sans prix.
- Une tâche s'ouvre → POST `/read`.

### Créer un projet : une seule question

La modale « Nouveau projet » ne demande **que le chemin d'un dépôt git**. Rien d'autre n'est posé à froid : le nom du projet est le basename du dépôt, le mode de livraison est déduit des remotes, la description, les instructions et les liens restent vides. Tout se règle ensuite, quand le besoin apparaît.

La règle vaut aussi côté serveur (`name` optionnel sur `POST /api/projects`) : elle n'est pas une astuce de formulaire.

### Réglages d'un projet : une modale à panneaux

L'édition d'un projet est une modale à navigation latérale, pas une colonne de champs empilés. Six panneaux, deux ou trois champs chacun :

| Panneau | Contenu |
|---|---|
| Général | nom, description, branche de base (`delivery.target`) |
| Dépôts | les lignes nom + chemin, « + dépôt » |
| Instructions | `contextPrompt` en grand (≥ 12 lignes), puis `checkCmd` |
| Livraison | les quatre modes en cartes à cocher, plus l'avertissement de santé s'il y en a un |
| Liens | les liens épinglés |
| Supprimer | conséquences chiffrées, puis le bouton de suppression en deux temps |

Règles qui tiennent ce découpage :

- **La colonne de navigation est le seul découpage** : aucun filet horizontal, aucun bloc encadré. Le panneau garde une hauteur minimale pour que la modale ne saute pas d'un onglet à l'autre.
- **La branche de base est dans Général, pas dans Livraison** : le serveur s'en sert aussi comme point de départ des branches de chantier, ce n'est donc pas un sous-réglage de la livraison.
- **Le mode de livraison se choisit en cartes**, chacune portant sa phrase de conséquence (les quatre sont lisibles d'un coup). Une liste déroulante obligerait à changer la sélection pour découvrir ce que chaque option ferait, alors que c'est le réglage qui décide si le produit pousse ou non.
- **La suppression est un panneau**, en bas de la colonne, détaché des autres : ce qu'on ne touche qu'une fois ne pèse pas en permanence sur ce qu'on change souvent. Elle garde la confirmation en deux temps.
- **Les libellés des champs courts sont en ligne** (libellé à gauche, saisie à droite, alignés d'un champ à l'autre) ; seuls les grands champs gardent leur libellé au-dessus.
- **L'aide est conditionnelle** : la phrase sur les chemins de dépôts ne s'affiche que si aucun chemin n'est encore saisi.
- **Un seul enregistrement pour toute la modale** : les champs saisis sont conservés d'un panneau à l'autre, le pied affiche « Modifications non enregistrées » dès la première frappe, et une erreur de validation ramène sur le panneau fautif.

### Création de tâche et navigation clavier

- Le bouton de création principal d'un écran porte trois éléments : un pictogramme `+`, son libellé, et le badge du raccourci (`N`). Le badge disparaît sous 700 px de large (pas de clavier physique attendu).
- La modale « Nouvelle tâche » offre deux validations : **Créer et discuter** (la conversation s'ouvre, `Ctrl/⌘+Entrée`) et **Créer et enchaîner** (la modale reste ouverte, `Ctrl/⌘+Maj+Entrée`). Après un enchaînement, l'agent et le dépôt choisis sont conservés, le titre et le prompt sont vidés, le focus revient au titre, et une ligne de confirmation annonce la tâche créée (`Tâche #{ref} créée : {titre}`).
- Raccourcis clavier, tous documentés dans la modale d'aide ouverte par `?` (et par le lien « Raccourcis : ? » en bas de sidebar) :

  | Portée | Touche | Effet |
  | --- | --- | --- |
  | Partout | `Ctrl/⌘+K`, `/` | Ouvrir la recherche |
  | Partout | `N` | Créer l'objet de l'écran courant (tâche dans un chantier, chantier dans un projet, projet dans « Tous les projets ») |
  | Partout | `?` | Ouvrir l'aide des raccourcis |
  | Partout | `Esc` | Fermer recherche / modale / tiroir, ou refermer le panneau d'une tâche |
  | Formulaire | `Ctrl/⌘+Entrée` | Valider |
  | Formulaire | `Ctrl/⌘+Maj+Entrée` | Valider et enchaîner (modale Nouvelle tâche) |
  | Formulaire | `Entrée` dans un champ d'une ligne | Valider (sauf dans le champ d'ajout d'un lien, où la validation est l'ajout du lien) |
  | Formulaire | `←` `→` `↑` `↓` | Changer d'agent dans le sélecteur |
  | Formulaire | `Tab` | Circuler dans les champs, sans jamais sortir de la modale |
  | Recherche | `↑` `↓` puis `Entrée` | Parcourir les résultats et ouvrir le résultat actif |
  | Tâche | `Entrée` | Envoyer le message du composeur |

- Aucun raccourci d'écran ne se déclenche pendant une saisie (champ, zone de texte, liste déroulante) ni avant l'authentification.
- Le sélecteur d'agent est un groupe de radios (`role="radiogroup"`, `aria-checked`) à tabulation roulante : un seul arrêt de `Tab` pour tout le groupe, les flèches déplacent la sélection et le focus ensemble.
