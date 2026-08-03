# Sillage : contrat d'API (v0.3.7)

Serveur Go sur `:8787`. Frontend statique servi sur `/`. Tout le JSON est en camelCase.

Chaque fichier statique est servi avec un `ETag` (empreinte de son contenu embarqué, calculée au démarrage) et `Cache-Control: no-cache` : le navigateur revalide à chaque chargement et reçoit `304 Not Modified` tant que le contenu n'a pas changé. Sans ces en-têtes, `embed.FS` ne fournissant aucune date de modification, la réponse n'aurait aucun validateur et le navigateur garderait indéfiniment son ancienne copie de `app.js` après un rebuild.

Auth par cookie de session (`sillage_session`, HttpOnly). Toute route `/api/*` (sauf `/api/login`) renvoie `401` sans session valide : le frontend affiche alors l'écran de connexion.

Mono-utilisateur : un seul mot de passe partagé (pas de comptes, pas de rôles). Tous les messages d'erreur API (`{"error": "..."}`) sont en anglais, courts (ex : `"confirmation required"`, `"task not found"`, `"invalid project path: not a git repository"`).

## Modèles

```jsonc
Tokens  { "input": 0, "output": 0, "costUsd": 0.0 }

Repo    { "name": "api", "path": "/abs/path" }   // name court et unique dans le projet

Link    { "url": "https://github.com/org/repo", "title": "org/repo" }
          // title fourni par l'utilisateur, ou récupéré best-effort (voir plus bas), ou nom d'hôte.

Delivery { "mode": "pr|merge", "target": "main", "stackedPrs": false }
          // Ce que « livrer » veut dire dans ce projet (voir "Livraison d'un chantier").
          // target vide = branche par défaut du dépôt. stackedPrs : réservé, ignoré.

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
          "shipBlocker": "|no-tasks|tasks-pending|nothing-accepted|nothing-to-ship",
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
          "unread": true, "updatedAt": "2026-08-02T10:00:00Z", "tokens": Tokens }
          // repoName : nom du Repo du projet utilisé pour le worktree

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
| GET | `/api/state` | | `{projects, cards, tasks, agents, workspace, settings, tokens:{global:Tokens}}` |
| GET | `/api/workspace` | | WorkspaceStatus |
| POST | `/api/workspace/setup` | `{mode:"local"\|"init"\|"clone", remote?}` | WorkspaceStatus (voir ci-dessous) |
| PATCH | `/api/workspace` | `{remote?, autoSync?}` | WorkspaceStatus (au moins un des deux champs requis ; `remote` définit/remplace origin, 400 si git non initialisé ; `autoSync:true` exige git initialisé ET un remote déjà défini, celui fourni dans le même appel compte, 400 sinon ; voir "Synchronisation automatique" ci-dessous) |
| POST | `/api/workspace/sync` | `{confirm:true}` | `{output, lastSyncAt}` (voir ci-dessous). **Validation humaine obligatoire** : refus 400 sans `confirm` |
| PATCH | `/api/settings` | `{displayName?, lang?}` | Settings (`lang` doit être `""`, `"fr"` ou `"en"`, 400 sinon) |
| POST | `/api/agents` | `{name, emoji?, color?, cli, model?, contextPrompt?}` | Agent (name et cli requis ; cli ∈ {claude, codex, fake} ; id = slug du name, 400 si déjà pris) |
| PATCH | `/api/agents/{id}` | mêmes champs, tous optionnels | Agent |
| DELETE | `/api/agents/{id}` | | 204 (400 si une tâche référence encore l'agent) |
| POST | `/api/projects` | `{name, path}` ou `{name, repos:[{name?,path}, ...], description?, contextPrompt?, links?:[{url,title?}, ...], delivery?}` | Project (400 si un path est invalide/pas un dépôt git, noms de repo dupliqués, mode de livraison inconnu, ou lien invalide/en trop grand nombre). `delivery` absent = déduit des remotes des dépôts (voir « Livraison d'un chantier ») |
| PATCH | `/api/projects/{id}` | `{name?, description?, checkCmd?, contextPrompt?, repos?, links?, delivery?}` | Project (repos/links, si fournis, remplacent la liste entière ; retirer un repo ne casse pas les tâches existantes) |
| DELETE | `/api/projects/{id}` | `{confirm:true}` | 204 (voir « Suppressions » ci-dessous). **Validation humaine obligatoire** : refus 400 sans `confirm` |
| POST | `/api/cards` | `{projectId, title, column?, contextPrompt?}` | Card. `column`, si fourni, doit valoir `"soon"` (400 `"cards are created in the soon column"` sinon) : les cartes se créent toujours dans "Bientôt" |
| PATCH | `/api/cards/{id}` | `{column?, title?, contextPrompt?}` | Card. `title`, si fourni, doit être non vide. Le déplacement manuel de colonne (toutes colonnes acceptées) reste indépendant de l'auto-déplacement (voir plus bas) |
| DELETE | `/api/cards/{id}` | `{confirm:true}` | 204 (voir « Suppressions » ci-dessous). **Validation humaine obligatoire** : refus 400 sans `confirm` |
| GET | `/api/cards/{id}/delivery` | | DeliveryPreview : ce que la livraison ferait, avant tout clic (voir « Livraison d'un chantier »). Lecture seule |
| POST | `/api/cards/{id}/ship` | `{confirm:true}` | ShipResponse : **la seule action sortante du produit** (push + pull request, ou fusion locale). 400 sans `confirm`, 409 si le chantier n'est pas livrable |
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
| GET | `/api/events` | | SSE |

### Cycle de vie des tâches et auto-déplacement de carte

Statuts : `running → review → accepted`, plus `cancelled` (via `/cancel`, le « refuser » de l'UI). Une tâche ne se livre plus seule : elle est acceptée (fusionnée) dans la branche de son chantier, et c'est le chantier qui se livre (voir « Livraison d'un chantier »). `/reopen` accepte accepted/cancelled et ramène en `review` ; le merge déjà fait n'est pas annulé, la prochaine acceptation fusionnera les nouveaux commits.

Compatibilité, au chargement de state.json : `ready` (antérieur à la v0.3.4) migre vers `review` ; `shipped` et `done` migrent vers `accepted`.

Après une acceptation réussie : un Message marqueur est ajouté au fil (`author="agent"`, `authorName=""`, texte figé `"[accepted:<workstreamBranch>]"`). Une acceptation simplement constatée (la branche de la tâche était déjà contenue dans celle du chantier) pose `"[auto-accepted:<workstreamBranch>]"`. En cas de conflit, la fusion est annulée (`git merge --abort`), la tâche **reste** en `review`, un marqueur `"[merge-conflict:<fichiers séparés par des espaces>]"` est ajouté au fil, et la réponse est 409 `{"error":"merge conflict with the workstream branch","conflictFilePaths":"..."}`. Le frontend détecte ces marqueurs et affiche une ligne système localisée.

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

**Réglage du projet** (`Project.delivery`) : `mode:"pr"` pousse la branche du chantier puis ouvre la pull request (GitHub, `gh pr create`) ou la merge request (GitLab, `glab mr create`) ; en l'absence du CLI ou en cas d'échec, repli sur une URL de création pré-remplie en lecture seule (`https://<host>/<path>/compare/<base>...<branch>?expand=1`, ou `https://<host>/<path>/-/merge_requests/new?merge_request[source_branch]=...&merge_request[target_branch]=...`). La branche est poussée dans les deux cas : le repli ne dégrade que l'ouverture de la requête. `mode:"merge"` fusionne la branche du chantier dans `target` **en local, en fast-forward uniquement, et ne pousse jamais rien** (worktree transitoire dédié ; repli dans le dépôt lui-même seulement si `target` y est la branche courante et l'arbre propre ; 409 sinon).

À la création d'un projet sans champ `delivery`, le mode est déduit : `"pr"` si tous les dépôts pointent vers une forge connue (github.com, ou un hôte contenant `gitlab`), sinon `"merge"` avec `target` = branche courante du premier dépôt. Le fournisseur n'est jamais persisté : il est redéduit du remote `origin` à chaque opération.

**Conditions de livraison** (`Card.shipReady`, sinon `Card.shipBlocker`) : au moins une tâche, aucune tâche `running` ni `review`, au moins une tâche `accepted`, au moins une branche de chantier. `POST /api/cards/{id}/ship` refuse en 409 avec le même vocabulaire.

**Acceptation automatique des branches fusionnées à la main.** Une branche de tâche peut être fusionnée dans celle du chantier hors de Sillage (`git merge` dans un terminal). `GET /api/cards/{id}/delivery` le constate au passage : une tâche `review` dont la branche est entièrement contenue dans celle du chantier passe `accepted`, avec un message marqueur `[auto-accepted:<branche du chantier>]` (le frontend affiche une ligne système localisée) et les événements SSE `task`, `message`, `cards`, `agents`. Quatre garde-fous : aucun agent en cours pour la tâche, `filesCount > 0` (une tâche vide est contenue par construction), son worktree propre (du travail non commité n'est par définition pas fusionné), et sa branche effectivement contenue. Aucune écriture git : l'appel ne fait que constater. Le frontend rappelle cet endpoint à l'ouverture de la vue chantier et toutes les 60 s tant qu'elle reste ouverte.

**Continuer un chantier déjà livré.** `CardBranch.shippedAt` signifie « livré, et rien de nouveau depuis ». Il est remis à `null` dès que du travail nouveau apparaît sur ce dépôt : création d'une tâche sur le chantier, ou acceptation qui ajoute réellement des commits à la branche (comparaison de SHA avant/après la fusion). Le chantier quitte alors la colonne `done`, et le bouton de livraison redevient actif une fois tout accepté ou refusé. `prUrl` est conservée : la pull request existe toujours, une nouvelle livraison la met à jour par un simple push (aucune tentative d'en ouvrir une seconde). Une acceptation automatique (voir ci-dessus) n'ajoute aucun commit et ne remet donc jamais le chantier en « non livré ».

`GET /api/cards/{id}/delivery` (aucune écriture git ; peut accepter des tâches déjà fusionnées, voir ci-dessus) :

```jsonc
{ "mode": "pr", "target": "main", "provider": "github|gitlab|",
  "ready": false, "blocker": "tasks-pending",
  "warnings": ["gh not found in PATH; ..."],
  "counts": { "accepted": 3, "refused": 1, "pending": 1 },
  "repos": [ { "repoName": "api", "branch": "sillage/ws-101-refonte-auth", "base": "main",
               "commits": 4, "files": 11, "pending": 4, "prUrl": "", "shippedAt": null } ] }
```

`commits`/`files` décrivent le contenu de la livraison (`base..branche`) ; `pending` est ce qu'il reste réellement à livrer (commits non poussés en mode `pr`, non encore fusionnés en mode `merge`). `pending` à zéro partout signifie « rien à livrer » : le bouton est inactif et la livraison marque le dépôt `skipped`.

`POST /api/cards/{id}/ship` `{confirm:true}` (400 `"confirmation required"` sinon) traite chaque dépôt du chantier et répond :

```jsonc
{ "card": Card, "mode": "pr",
  "repos": [ { "repoName": "api", "branch": "...", "base": "main", "pushed": true,
               "merged": false, "skipped": false, "prUrl": "https://...",
               "output": "...", "error": "" } ] }
```

Un dépôt en échec n'annule pas les autres : chaque ligne porte sa propre erreur, et la livraison est rejouable telle quelle (les dépôts déjà livrés sont `skipped`).

### Santé de la livraison

`Project.deliveryWarning` (dans `ProjectOut`, jamais `Project` lui-même : le champ n'est pas persisté) est recalculé à chaque lecture d'état ou mutation de projet. Vide en mode `merge` (aucun binaire externe, aucun remote requis) ; sinon, par ordre de priorité : `"no 'origin' remote on repository <name>; nothing can be pushed"`, `"unknown forge on repository <name>; the branch will be pushed without opening a pull request"`, `"gh not found in PATH; Sillage will fall back to a prefilled pull request URL"`, `"glab not found in PATH; ..."`. Aucun de ces avertissements n'empêche quoi que ce soit.

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
- `taskDeleted` : `{taskId, cardId, projectId}` (après `DELETE /api/tasks/{id}`, une fois par tâche supprimée y compris en cascade).
- `cardDeleted` : `{cardId, projectId}` (après `DELETE /api/cards/{id}`).
- `projectDeleted` : `{projectId}` (après `DELETE /api/projects/{id}` ; le frontend recharge l'état complet plutôt que de rejouer la cascade).

Le frontend maintient son état en mémoire à partir de `/api/state` + SSE ; reconnexion SSE automatique (EventSource natif) et re-fetch de `/api/state` à la reconnexion.

## Règles produit à respecter côté UI

- Push / livraison uniquement via le bouton « Livrer » de l'en-tête du chantier, en deux clics : le bouton ouvre un récapitulatif (mode, dépôts, branche → base, commits, fichiers, avertissements) dont le bouton d'action **est** la confirmation (`{"confirm":true}`). Jamais automatique. Le bouton est grisé, avec la raison en infobulle, tant que `card.shipReady` est faux ou qu'il n'y a rien à livrer. Même règle de confirmation explicite pour la synchronisation de l'espace de travail (`/api/workspace/sync`).
- Accepter ou refuser une tâche s'obtient en un clic depuis la liste de tâches (boutons révélés au survol d'une tâche à relire) : ces actions sont locales et réversibles (`/reopen`), donc sans confirmation. L'état de chaque tâche (acceptée, refusée) est lisible dans la liste sans ouvrir la tâche.
- Tokens : jamais affichés dans le flux de travail (kanban, détail de tâche, liste des projets), pour ne pas ajouter de charge mentale. Seul endroit visible : Réglages > onglet Statistiques, consommation par projet, sans prix.
- Une tâche s'ouvre → POST `/read`.

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
  | Formulaire | `Entrée` dans un champ d'une ligne | Valider |
  | Formulaire | `←` `→` `↑` `↓` | Changer d'agent dans le sélecteur |
  | Formulaire | `Tab` | Circuler dans les champs, sans jamais sortir de la modale |
  | Recherche | `↑` `↓` puis `Entrée` | Parcourir les résultats et ouvrir le résultat actif |
  | Tâche | `Entrée` | Envoyer le message du composeur |

- Aucun raccourci d'écran ne se déclenche pendant une saisie (champ, zone de texte, liste déroulante) ni avant l'authentification.
- Le sélecteur d'agent est un groupe de radios (`role="radiogroup"`, `aria-checked`) à tabulation roulante : un seul arrêt de `Tab` pour tout le groupe, les flèches déplacent la sélection et le focus ensemble.
