# Sillage : contrat d'API (v0.2)

Serveur Go sur `:8787`. Frontend statique servi sur `/`. Tout le JSON est en camelCase.
Auth par cookie de session (`sillage_session`, HttpOnly). Toute route `/api/*` (sauf `/api/login`) renvoie `401` sans session valide : le frontend affiche alors l'écran de connexion.

Mono-utilisateur : un seul mot de passe partagé (pas de comptes, pas de rôles). Tous les messages d'erreur API (`{"error": "..."}`) sont en anglais, courts (ex : `"confirmation required"`, `"task not found"`, `"invalid project path: not a git repository"`).

## Modèles

```jsonc
Tokens  { "input": 0, "output": 0, "costUsd": 0.0 }

Project { "id": "p1", "name": "sillage", "path": "/abs/path", "unread": 2,
          "tokens": Tokens, "checkCmd": "go test ./..." }   // checkCmd peut être vide

Card    { "id": "c1", "projectId": "p1", "column": "soon|doing|done", "title": "...",
          "tasksTotal": 4, "tasksDone": 1, "docsCount": 2, "messagesCount": 12,
          "reviewCount": 1, "progress": 25, "liveActivity": "..." | null }

Agent   { "id": "bolt", "name": "Bolt", "emoji": "🐝", "color": "#f2b705",
          "model": "claude-sonnet-5", "cli": "claude", "contextPrompt": "...",
          "active": true }        // active = une tâche running lui est assignée

Task    { "id": "t1", "cardId": "c1", "projectId": "p1", "ref": 482, "title": "...",
          "agentId": "bolt", "branch": "sillage/482-slug",
          "status": "running|review|ready|shipped",
          "messagesCount": 5, "filesCount": 3, "docsCount": 1,
          "checks": [ { "label": "go test", "ok": true } ],   // [] si aucun
          "liveActivity": "Edit · internal/server/store.go" | null,
          "unread": true, "updatedAt": "2026-08-02T10:00:00Z", "tokens": Tokens }

Message { "id": "m1", "taskId": "t1", "author": "user|agent", "authorName": "Bolt",
          "text": "markdown...", "createdAt": "..." }
          // authorName : nom de l'agent pour author="agent" ; vide pour author="user"
          // (mono-utilisateur : le frontend affiche alors "Vous"/"You")
```

`Card.progress` = tasksDone/tasksTotal en % (0 si vide). Les compteurs de Card sont calculés côté serveur. `tasksDone` = tâches `shipped`. `reviewCount` = tâches `review`.

## Endpoints

| Méthode | Route | Corps | Réponse |
|---|---|---|---|
| POST | `/api/login` | `{password}` | 204 ou 401 `{error}` |
| POST | `/api/logout` | | 204 |
| GET | `/api/state` | | `{projects, cards, tasks, agents, tokens:{global:Tokens}}` |
| POST | `/api/agents` | `{name, emoji?, color?, cli, model?, contextPrompt?}` | Agent (name et cli requis ; cli ∈ {claude, codex, fake} ; id = slug du name, 400 si déjà pris) |
| PATCH | `/api/agents/{id}` | mêmes champs, tous optionnels | Agent |
| DELETE | `/api/agents/{id}` | | 204 (400 si une tâche référence encore l'agent) |
| POST | `/api/projects` | `{name, path}` | Project (400 si path invalide/pas un dépôt git) |
| PATCH | `/api/projects/{id}` | `{name?, checkCmd?}` | Project |
| POST | `/api/cards` | `{projectId, title, column?}` | Card |
| PATCH | `/api/cards/{id}` | `{column}` | Card |
| POST | `/api/tasks` | `{cardId, title, agentId, prompt?}` | Task (créée `running`, agent lancé avec title+prompt) |
| GET | `/api/tasks/{id}` | | `{task, messages}` |
| POST | `/api/tasks/{id}/messages` | `{text}` | 202 ; relance l'agent (statut → running) |
| POST | `/api/tasks/{id}/interrupt` | | Task (→ review) |
| POST | `/api/tasks/{id}/accept` | | Task (review → ready) |
| POST | `/api/tasks/{id}/ship` | `{confirm:true}` | `{task, output}` : fait le `git push` réel. **Validation humaine obligatoire** : refus 400 sans `confirm` |
| POST | `/api/tasks/{id}/pr` | `{confirm:true}` | `{url}` : ouvre la pull request (voir ci-dessous). **Validation humaine obligatoire** : refus 400 sans `confirm` ; tâche `shipped` uniquement (400 sinon) |
| POST | `/api/tasks/{id}/reopen` | | Task (shipped → review) |
| POST | `/api/tasks/{id}/read` | | 204 (unread=false) |
| GET | `/api/tasks/{id}/diff` | | voir ci-dessous |
| GET | `/api/tasks/{id}/deliverables` | | voir ci-dessous |
| GET | `/api/events` | | SSE |

### Ouvrir la PR

`POST /api/tasks/{id}/pr` ne pousse jamais rien : la branche a déjà été poussée par `/ship`. Implémentation : `gh pr create` dans le worktree (si `gh` est disponible) ; en cas d'échec ou d'absence de `gh`, repli sur une URL de comparaison GitHub en lecture seule (`https://github.com/<owner>/<repo>/compare/<base>...<branch>?expand=1`), déduite de `git remote get-url origin` (ssh ou https). Si ni l'un ni l'autre n'aboutit (remote non-GitHub ou absent) : 502 `{error}`.

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

Le frontend maintient son état en mémoire à partir de `/api/state` + SSE ; reconnexion SSE automatique (EventSource natif) et re-fetch de `/api/state` à la reconnexion.

## Règles produit à respecter côté UI

- Push / livraison uniquement via le bouton du détail de tâche avec confirmation explicite (double clic de confirmation ou bouton qui devient « Confirmer le push ? »). Jamais automatique. Même règle pour l'ouverture de la PR.
- Tokens visibles : total global en bas de sidebar (`Σ 12,4k tokens · 0,84 $`), par projet dans l'en-tête kanban, par tâche dans le détail (sous l'en-tête).
- Une tâche s'ouvre → POST `/read`.
