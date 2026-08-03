# Sillage : contrat d'API (v0.3.4)

Serveur Go sur `:8787`. Frontend statique servi sur `/`. Tout le JSON est en camelCase.
Auth par cookie de session (`sillage_session`, HttpOnly). Toute route `/api/*` (sauf `/api/login`) renvoie `401` sans session valide : le frontend affiche alors l'écran de connexion.

Mono-utilisateur : un seul mot de passe partagé (pas de comptes, pas de rôles). Tous les messages d'erreur API (`{"error": "..."}`) sont en anglais, courts (ex : `"confirmation required"`, `"task not found"`, `"invalid project path: not a git repository"`).

## Modèles

```jsonc
Tokens  { "input": 0, "output": 0, "costUsd": 0.0 }

Repo    { "name": "api", "path": "/abs/path" }   // name court et unique dans le projet

Project { "id": "p1", "name": "sillage", "description": "...", "repos": [Repo, ...], "unread": 2,
          "tokens": Tokens, "checkCmd": "go test ./...", "contextPrompt": "..." }
          // description : une phrase, affichée sous le nom. checkCmd/contextPrompt/description peuvent être vides.

Card    { "id": "c1", "projectId": "p1", "column": "soon|doing|done", "title": "...",
          "tasksTotal": 4, "tasksDone": 1, "docsCount": 2, "messagesCount": 12,
          "reviewCount": 1, "progress": 25, "liveActivity": "..." | null, "contextPrompt": "..." }
          // Card = chantier (vocabulaire produit) ; nom technique inchangé. contextPrompt :
          // texte libre transmis aux agents (voir plus bas), peut être vide.

Agent   { "id": "bolt", "name": "Bolt", "emoji": "🐝", "color": "#f2b705",
          "model": "claude-sonnet-5", "cli": "claude", "contextPrompt": "...",
          "active": true, "warning": "..." }
          // active = une tâche running lui est assignée. warning : calculé à chaque
          // liste d'agents (jamais persisté dans state.json), vide si tout va bien ;
          // voir "Santé des agents" ci-dessous.

Task    { "id": "t1", "cardId": "c1", "projectId": "p1", "ref": 482, "title": "...",
          "agentId": "bolt", "repoName": "api", "branch": "sillage/482-slug",
          "status": "running|review|shipped|done|cancelled",
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
                  "dirty": false, "lastCommitAt": "..." | null, "lastSyncAt": "..." | null }

Settings { "displayName": "Ada", "lang": "fr" }   // lang : ""|"fr"|"en"
```

`Card.progress` = tasksDone/tasksTotal en % (0 si vide). Les compteurs de Card sont calculés côté serveur. `tasksDone` = tâches `shipped`+`done`. `reviewCount` = tâches `review`. Les tâches `cancelled` sont exclues de `tasksTotal`/`tasksDone`/`progress`, mais comptent comme terminales pour l'auto-déplacement de carte (voir plus bas).

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
| PATCH | `/api/workspace` | `{remote}` | WorkspaceStatus (définit/remplace origin ; 400 si git non initialisé) |
| POST | `/api/workspace/sync` | `{confirm:true}` | `{output, lastSyncAt}` (voir ci-dessous). **Validation humaine obligatoire** : refus 400 sans `confirm` |
| PATCH | `/api/settings` | `{displayName?, lang?}` | Settings (`lang` doit être `""`, `"fr"` ou `"en"`, 400 sinon) |
| POST | `/api/agents` | `{name, emoji?, color?, cli, model?, contextPrompt?}` | Agent (name et cli requis ; cli ∈ {claude, codex, fake} ; id = slug du name, 400 si déjà pris) |
| PATCH | `/api/agents/{id}` | mêmes champs, tous optionnels | Agent |
| DELETE | `/api/agents/{id}` | | 204 (400 si une tâche référence encore l'agent) |
| POST | `/api/projects` | `{name, path}` ou `{name, repos:[{name?,path}, ...], description?, contextPrompt?}` | Project (400 si un path est invalide/pas un dépôt git, ou noms de repo dupliqués) |
| PATCH | `/api/projects/{id}` | `{name?, description?, checkCmd?, contextPrompt?, repos?}` | Project (repos, si fourni, remplace la liste entière ; retirer un repo ne casse pas les tâches existantes) |
| POST | `/api/cards` | `{projectId, title, column?, contextPrompt?}` | Card. `column`, si fourni, doit valoir `"soon"` (400 `"cards are created in the soon column"` sinon) : les cartes se créent toujours dans "Bientôt" |
| PATCH | `/api/cards/{id}` | `{column?, title?, contextPrompt?}` | Card. `title`, si fourni, doit être non vide. Le déplacement manuel de colonne (toutes colonnes acceptées) reste indépendant de l'auto-déplacement (voir plus bas) |
| POST | `/api/tasks` | `{cardId, title, agentId, prompt?, repoName?}` | Task (créée `running`, agent lancé avec title+prompt). `repoName` optionnel si le projet n'a qu'un repo ; sinon 400 `"repoName required (project has several repositories)"` ou 400 si inconnu |
| PATCH | `/api/tasks/{id}` | `{agentId}` | Task : réassigne l'agent (voir « Réassignation » ci-dessous). 400 si `status=running` (`"interrupt the agent before reassigning"`) ou si l'agent est inconnu |
| GET | `/api/tasks/{id}` | | `{task, messages}` |
| POST | `/api/tasks/{id}/messages` | `{text}` | 202 ; relance l'agent (statut → running) |
| POST | `/api/tasks/{id}/interrupt` | | Task (running → review) |
| POST | `/api/tasks/{id}/ship` | `{confirm:true}` | `{task, output, branchUrl}` : fait le `git push` réel, accepté depuis `review` (voir ci-dessous). **Validation humaine obligatoire** : refus 400 sans `confirm` |
| POST | `/api/tasks/{id}/pr` | `{confirm:true}` | `{url}` : ouvre la pull request (voir ci-dessous). **Validation humaine obligatoire** : refus 400 sans `confirm` ; tâche `shipped` uniquement (400 sinon) |
| POST | `/api/tasks/{id}/finish` | | Task (review/shipped → done). 400 depuis `running` (`"task must be reviewed before finishing"`) ou depuis done/cancelled |
| POST | `/api/tasks/{id}/cancel` | | Task (running/review → cancelled). Si `running`, l'agent est d'abord interrompu (même mécanique qu'`/interrupt`). 400 depuis shipped/done/cancelled |
| POST | `/api/tasks/{id}/reopen` | | Task (shipped/done/cancelled → review). 400 sinon |
| POST | `/api/tasks/{id}/read` | | 204 (unread=false) |
| GET | `/api/tasks/{id}/diff` | | voir ci-dessous |
| GET | `/api/tasks/{id}/deliverables` | | voir ci-dessous |
| GET | `/api/events` | | SSE |

### Cycle de vie des tâches et auto-déplacement de carte

Statuts : `running → review → shipped`, plus `done` (via `/finish`) et `cancelled` (via `/cancel`). L'état intermédiaire `ready` (et l'étape d'acceptation manuelle `/accept`) a disparu : `ship` est accepté directement depuis `review`. `/reopen` accepte shipped/done/cancelled et ramène en `review`. Compatibilité : au chargement de state.json, toute tâche encore `ready` (installation antérieure à la v0.3.4) migre vers `review`.

Après un ship réussi : un Message marqueur est ajouté au fil (`author="agent"`, `authorName=""`, texte figé `"[shipped:<branch>]"` ; le frontend détecte ce marqueur et affiche une ligne système localisée). La réponse inclut aussi `branchUrl` : l'URL de la branche sur GitHub (`https://github.com/<owner>/<repo>/tree/<branch>`) si le remote `origin` du dépôt est un dépôt github.com, chaîne vide sinon (jamais d'erreur pour cette information optionnelle).

Après chaque changement de statut de tâche, la carte est automatiquement replacée : si elle a au moins une tâche et que toutes ses tâches sont terminales (`shipped`/`done`/`cancelled`), `card.column` passe à `"done"` ; si une tâche redevient active (reopen, nouvelle tâche) alors que la carte est en `"done"`, elle repasse en `"doing"`. Le déplacement manuel (PATCH `/api/cards/{id}`) reste indépendant de cette règle. Chaque changement republie l'événement SSE `cards`.

### Réassignation d'une tâche à un autre agent

`PATCH /api/tasks/{id} {agentId}` : refusé si la tâche est `running` (l'interrompre d'abord via `/interrupt`) ou si l'agent est inconnu. Effets : `task.agentId` change, `task.sessionId` est vidé (le nouvel agent ne peut pas reprendre la session CLI de l'ancien), et un Message est ajouté au fil avec `author="agent"`, `authorName=""` et un texte figé `"[reassigned:<agentId>]"` : le frontend détecte ce marqueur et affiche une ligne système localisée (`author`/`authorName` volontairement neutres pour rester i18n-propre côté backend). SSE `task` + `message` + `agents`.

Départ frais : quand la session CLI est vide (lancement initial, ou premier message après une réassignation), le texte envoyé au CLI est préfixé par un rappel minimal : `Task: <title>\n\n<texte>` (pas de préfixe si le texte est vide).

### Santé des agents

`Agent.warning` (dans `AgentOut`, jamais `Agent` lui-même : le champ n'existe pas dans le modèle persisté) est recalculé à chaque liste d'agents (`GET /api/state`, événement SSE `agents`) :
- `cli=codex` : si `/proc/sys/kernel/apparmor_restrict_unprivileged_userns` vaut `1` et que `SILLAGE_CODEX_SANDBOX` n'est pas définie → `"codex sandbox is blocked on this machine (AppArmor); see README (SILLAGE_CODEX_SANDBOX)"`.
- `cli=codex` ou `cli=claude` : si le binaire correspondant est introuvable dans le PATH → `"<cli> CLI not found in PATH"`.
- Sinon (ou `cli=fake`) : chaîne vide.

### Ouvrir la PR

`POST /api/tasks/{id}/pr` ne pousse jamais rien : la branche a déjà été poussée par `/ship`. Implémentation : `gh pr create` dans le worktree (si `gh` est disponible) ; en cas d'échec ou d'absence de `gh`, repli sur une URL de comparaison GitHub en lecture seule (`https://github.com/<owner>/<repo>/compare/<base>...<branch>?expand=1`), déduite de `git remote get-url origin` (ssh ou https). Si ni l'un ni l'autre n'aboutit (remote non-GitHub ou absent) : 502 `{error}`.

### Espace de travail (synchronisation git de dataDir)

Le répertoire de données (dataDir) peut devenir un dépôt git optionnel, pour sauvegarder/synchroniser `state.json`, `config.json` et `.gitignore` (seuls fichiers versionnés ; `.gitignore` exclut `worktrees/` et `*.tmp`). Un commit local est tenté silencieusement si git est activé, throttlé à au plus un par quart d'heure (`workspaceCommitInterval`) : chaque commit stocke un blob complet de `state.json`, et un agent qui travaille déclenche plusieurs sauvegardes par seconde. La sauvegarde de `state.json` elle-même reste atomique et immédiate à chaque mutation ; le commit n'est qu'un point de restauration. Le dépôt de l'espace de travail est configuré en `gc.auto 256` pour que git compacte souvent. Jamais de push automatique.

- `mode:"local"` : marque `setupDone`, aucun git. Refusé (400) si déjà fait.
- `mode:"init"` : `git init` (branche main), écrit `.gitignore`, premier commit ; `remote` optionnel (`git remote add`, jamais de push). Rejouable plus tard (depuis les réglages) pour activer git sur un espace resté local, même si `setupDone` est déjà vrai.
- `mode:"clone"` (`remote` requis) : clone dans un répertoire temporaire, vérifie que `state.json` existe à sa racine (sinon 400 `"remote does not look like a Sillage workspace"`), puis remplace `state.json`/`config.json`/`.git` de dataDir par ceux du clone et recharge le store en mémoire, sans redémarrage (les sessions actives restent valides). Le mot de passe devient celui de l'espace rapatrié. Refusé (400) si déjà fait.
- `POST /api/workspace/sync` : commit d'abord si nécessaire, puis `git pull --rebase origin main`, puis `git push -u origin main` (fonction dédiée `SyncPush`, qui n'opère jamais sur un dépôt de projet). En cas de conflit de rebase : `git rebase --abort` puis 409 `{"error":"sync conflict: the remote workspace diverged, resolve manually in <dataDir>"}`.
- Compatibilité : un state.json existant sans champ `workspace` (installation antérieure à la v0.3) migre vers `setupDone=true` en mode local au chargement : l'onboarding ne s'affiche jamais sur un espace déjà utilisé.

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
- `workspace` : WorkspaceStatus (après setup, changement de remote ou sync).
- `settings` : Settings (après PATCH `/api/settings`).

Le frontend maintient son état en mémoire à partir de `/api/state` + SSE ; reconnexion SSE automatique (EventSource natif) et re-fetch de `/api/state` à la reconnexion.

## Règles produit à respecter côté UI

- Push / livraison uniquement via le bouton du détail de tâche avec confirmation explicite (double clic de confirmation ou bouton qui devient « Confirmer le push ? »). Jamais automatique. Même règle pour l'ouverture de la PR et pour la synchronisation de l'espace de travail (`/api/workspace/sync`).
- Tokens visibles : total global en bas de sidebar (`Σ 12,4k tokens · 0,84 $`), par projet dans l'en-tête kanban, par tâche dans le détail (sous l'en-tête).
- Une tâche s'ouvre → POST `/read`.
