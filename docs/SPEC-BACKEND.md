# Sillage : spec interne backend (Go)

Binaire unique `sillage`. Go 1.24. Seule dépendance externe autorisée : `golang.org/x/crypto/bcrypt`.
Frontend embarqué via `//go:embed web`. Respecter SPEC-API.md à la lettre.

## Arborescence

```
main.go                      // flags, config, seed, routes, embed web/
internal/server/
  models.go                  // structs JSON de SPEC-API.md + State
  store.go                   // Store: état en mémoire + persistance JSON atomique
  auth.go                    // bcrypt, sessions, middleware, rate-limit login
  sse.go                     // Hub SSE
  git.go                     // worktrees, diff parse, commits, push
  runner.go                  // adaptateurs claude / codex / fake
  handlers.go                // REST
  store_test.go, git_test.go // tests unitaires
```

## Flags et config

- `-addr` (défaut `127.0.0.1:8787`), `-data` (défaut `~/.local/share/sillage`).
- `config.json` dans dataDir : `{ "passwordHash": "..." }`. Le mot de passe est optionnel : sans `SILLAGE_PASSWORD` défini au démarrage ni hash déjà présent dans `config.json`, `LoadPasswordHash` renvoie un hash vide, aucun fichier n'est écrit, et le serveur tourne sans authentification (voir `withMiddleware`). Env `SILLAGE_PASSWORD` (si présent au démarrage) fournit le hash bcrypt en mémoire, jamais persisté.
- `state.json` dans dataDir : tout l'état (projects, cards, tasks, messages, agents, refCounter, usages tokens). Écriture atomique : fichier temp + `os.Rename`, à chaque mutation (fichier petit).

## Store

`Store` avec `sync.Mutex`, méthodes CRUD, recalcul des compteurs dérivés (Card.progress, counts, Card.shipReady/shipBlocker, Project.unread, Agent.active) à chaque lecture/émission. IDs : `p1`, `c1`, `t1`... (compteur) ; ref : compteur global démarrant à 100, partagé par les tâches et les chantiers (le nom de branche d'un chantier en a besoin).

Branches de chantier : `GetCardBranch`, `SetCardBranch`, `MarkCardBranchShipped(cardID, repoName, prURL, at)` et `MarkCardBranchPending(cardID, repoName)` (remise à `shippedAt=nil` quand du travail nouveau arrive ; `prUrl` conservée). Tous recalculent la carte, verrou tenu.
Tokens : cumulés dans Task.tokens ; agrégats projet/global calculés en sommant les tâches.

## Auth

- Sessions en mémoire : token 32 octets crypto/rand hex, map[token]expiry (30 jours). Cookie `sillage_session` HttpOnly SameSite=Lax, Secure si TLS ou `X-Forwarded-Proto: https`.
- Rate-limit login : max 5 échecs/minute par IP (map + fenêtre glissante), réponse 429.
- Toute mutation exige `Content-Type: application/json` (403 sinon) : protection CSRF avec SameSite=Lax.
- Middleware auth sur tout `/api/*` sauf `/api/login`. Les fichiers statiques sont publics mais ne contiennent rien de sensible.

## SSE (sse.go)

Hub : `Subscribe() (ch chan Event, unsub func())`, `Publish(Event{Name, Data any})`. Handler `/api/events` : headers SSE, flush après chaque événement, ping commentaire `: ping` toutes les 25 s, buffer 64, client lent = drop de l'événement (non bloquant).

## Git (git.go)

Deux niveaux de worktree, un par branche (voir docs/SPEC-LIVRAISON.md) :

- **chantier** : `<dataDir>/worktrees/ws-<cardId>-<slug(repoName)>`, branche `sillage/ws-<card.ref>-<slug(titre)>`, créée depuis `project.delivery.target` (ou la branche courante du dépôt si vide). C'est là que les tâches acceptées sont fusionnées, et c'est cette branche que la livraison pousse.
- **tâche** : `<dataDir>/worktrees/<taskId>`, branche `sillage/<ref>-<slug>`, créée depuis la branche du chantier (stockée comme `base` dans la tâche). Jamais poussée.

- `CreateWorktree(repoPath, dataDir, taskID, branch, base) (dir, base, error)` et `CreateCardWorktree(repoPath, dataDir, cardID, repoName, branch, base)` : `git worktree add -B <branch> <dir> <base>` (base vide = branche courante du dépôt).
- `Diff(dir, base)` : d'abord `git add -A -N` (fait apparaître les fichiers non suivis), puis `git diff <base>` ; parser le unified diff en structure de SPEC-API.md (fichiers, additions/deletions, hunks, lignes ctx/add/del). Parser robuste : ignorer les lignes binaires, gérer rename.
- `Commits(dir, base)` : `git log <base>..HEAD --pretty=format:%h|%s|%cr`. `CountCommits(dir, from, to)` : `git rev-list --count from..to` (sert à `pending`, ce qu'il reste à livrer).
- `CommitAll(dir, message)` : `git add -A` puis commit si l'arbre est sale (sans erreur s'il est propre). `IsWorktreeClean(dir)`, `HeadSha(dir)`, `IsBranchMergedInto(dir, branch, target)` : constats en lecture seule.
- `MergeBranch(dir, branch, message)` : `git merge --no-ff -m <message> <branch>` dans le worktree du chantier (un merge commit par tâche acceptée, l'historique reste lisible). En cas de conflit : liste les fichiers en conflit (`git diff --name-only --diff-filter=U`), `git merge --abort`, et enveloppe `ErrMergeConflict`. Aucune résolution automatique.
- `RebaseOnto(dir, onto)` : `git rebase <onto>` dans le worktree d'une tâche, pour la faire repartir du travail accepté. Conflit → `git rebase --abort` (l'arbre revient intact) et `ErrRebaseConflict` avec les fichiers en conflit. L'appelant garantit worktree propre et aucun agent en cours : un rebase réécrit l'historique de la branche de la tâche.
- `pushBranch(dir, branch, setUpstream)` : `git push [-u] origin <branch>`, jamais de `--force`. **C'est, avec `SyncPush` (espace de travail uniquement), le SEUL endroit du code qui exécute `git push`** ; deux appelants seulement, `Ship` et `mergeThenPush`.
- `Ship(dir, branch, title)` : commit de sécurité si l'arbre est sale, puis `pushBranch` de la branche du chantier. Erreur claire si pas de remote. Sert les modes `pr` et `push` : leur seule différence est l'ouverture de la pull request, faite par l'appelant.
- `DetectForge(dir) ForgeInfo` : remote `origin` parsé (ssh scp-like ou URL) → fournisseur `github` (github.com) / `gitlab` (hôte contenant `gitlab`) / `""`. `RemoteURL` vide distingue « pas de remote » de « forge inconnue ». Jamais persisté.
- `OpenPR(dir, branch, base, title)` : `gh pr create` ou `glab mr create` si le binaire est présent, sinon URL de création pré-remplie (`GithubCompareURL` / `GitlabNewMRURL`). Aucune écriture git : la branche a déjà été poussée par `Ship`.
- `MergeLocal(repoPath, dataDir, cardID, repoName, target, source)` : fusion locale **fast-forward uniquement, jamais de push**. D'abord un worktree transitoire dédié (`worktrees/.merge-<cardId>-<repoName>`, retiré à la fin) ; si `target` y est déjà empruntée (cas courant : c'est la branche courante du dépôt de travail), repli dans le dépôt lui-même, autorisé seulement si `target` **est** la branche courante et l'arbre propre (jamais de changement de branche, jamais de stash). Sinon `ErrTargetBusy`. Divergence → `ErrTargetDiverged`, avec la commande à jouer à la main.
- `MergeAndPush(...)` : même chemin (les deux passent par `mergeIntoTarget`), plus le push de `target` (mode `merge-push`). `fastForwardFromRemote` rattrape d'abord `origin/<target>` : `git fetch origin <target>` (référence distante absente = rien à rattraper, la branche sera créée au push), puis rien à faire si le local contient déjà le distant (`git merge-base --is-ancestor FETCH_HEAD HEAD`), sinon `git merge --ff-only FETCH_HEAD`, sinon `ErrTargetDiverged`. Le rattrapage vient **avant** la fusion : un refus avant écriture vaut mieux qu'une fusion locale suivie d'un push rejeté.
- Toutes les commandes git : `exec.Command` avec `Dir` fixé, env hérité + `GIT_TERMINAL_PROMPT=0`, timeout 60 s (120 s pour push) via context.

## Runner (runner.go)

`Runner` gère au plus un process par tâche (map[taskID]*exec.Cmd + mutex). `Start(task, agent, userText)` :

1. statut tâche → `running`, publier SSE `task`.
2. Ajouter le Message user (sauf lancement initial où le prompt = titre + description).
3. Lancer le process selon `agent.cli` (voir ci-dessous), cwd = worktree de la tâche.
4. Goroutine : lire stdout ligne par ligne (JSONL), publier `activity`, créer les Messages agent, cumuler les tokens (SSE `tokens`).
5. Fin de process : lancer le check éventuel du projet (`Project.checkCmd`, timeout 120 s) → Task.checks ; statut → `review` (sauf si interrompu : déjà review) ; `unread=true` ; `liveActivity=null` ; SSE `task`, `cards`, `agents`, `activity(null)`.

`Interrupt(taskID)` : SIGINT au groupe de process (Setpgid à la création), 5 s, puis SIGKILL. Statut → review.

### Adaptateur claude (`cli:"claude"`)

```
claude -p --output-format stream-json --verbose \
  --model <agent.model> \
  --permission-mode acceptEdits \
  --allowedTools "Read,Edit,Write,Glob,Grep,WebFetch,Bash(go build:*),Bash(go test:*),Bash(go vet:*),Bash(ls:*),Bash(cat:*),Bash(mkdir:*)" \
  --append-system-prompt <agent.contextPrompt> \
  [--resume <sessionID si connu>] \
  <texte utilisateur>
```

JSONL émis, à parser (champs utiles seulement) :

- `{"type":"system","subtype":"init","session_id":"..."}` → stocker session_id sur la tâche.
- `{"type":"assistant","message":{"content":[{"type":"text","text":"..."} , {"type":"tool_use","name":"Edit","input":{"file_path":"..."}}]}}` :
  - bloc `text` → Message agent (SSE `message`).
  - bloc `tool_use` → ligne d'activité `"<name> · <résumé input>"` (file_path, command ou pattern tronqué à 80 chars) → Task.liveActivity + SSE `activity`.
- `{"type":"result","subtype":"success","result":"...","session_id":"...","total_cost_usd":0.0123,"usage":{"input_tokens":N,"output_tokens":N,"cache_read_input_tokens":N,"cache_creation_input_tokens":N}}` → cumuler tokens (input = input+cache_creation, le cache_read compté à part n'est PAS ajouté à input ; costUsd += total_cost_usd). Si `result` non vide et différent du dernier message texte, l'ajouter comme Message.
- Lignes non-JSON ou types inconnus : ignorer silencieusement.
- `is_error:true` ou exit code != 0 → Message agent `⚠️ <détail>` et statut review.

Jamais de `--dangerously-skip-permissions`. Jamais `git push` dans allowedTools.

### Adaptateur codex (`cli:"codex"`), best-effort

`codex exec --json -C <worktree> [--model <model>] <texte>` ; parser plusieurs formes de JSONL :
`{"msg":{"type":"agent_message","message":"..."}}`, `{"type":"item.completed","item":{"type":"agent_message","text":"..."}}`, `token_count` → usage si présent. En cas d'échec de parsing, capturer stdout brut en un Message final. Pas de resume en v1 (chaque message relance `codex exec` avec le dernier texte).

### Adaptateur fake (`cli:"fake"`)

Sans exec : goroutine qui simule ~3 s de travail : 3 lignes d'activité espacées, écrit/complète un fichier `SILLAGE-TEST.md` dans le worktree (contenu horodaté), un Message agent de synthèse, usage fictif `{input:1200, output:340, costUsd:0.004}`. Sert aux tests et à la démo sans coût.

## Seed (premier lancement)

Agents : Bolt 🐝 `#f2b705` claude/sonnet (contexte : dev backend pragmatique) ; Muse 🦊 `#d0662f` claude/opus (produit, specs, docs) ; Otto 🦉 `#4f7d2f` codex/(modèle vide = défaut) (infra) ; Écho 🧪 `#777` fake (agent de test local, gratuit). Pas de projets seedés.

## Handlers

Suivre SPEC-API.md. Points d'attention :
- POST /projects : vérifier que path existe et `git rev-parse --git-dir` réussit, sinon 400 avec message explicite. Sans champ `delivery`, déduire le mode des remotes (`detectDelivery`).
- POST /tasks : garantir la branche du chantier sur ce dépôt (`ensureCardBranch`, créée à la première tâche), puis le worktree de la tâche depuis cette branche ; si échec, 400 et pas de tâche. Lancer l'agent avec `title` + `prompt`. `ensureCardBranch` remet aussi le chantier en « non livré » (`MarkCardBranchPending`) s'il l'était : une tâche de plus veut dire que tout n'est plus livré.
- POST /tasks/{id}/accept : commit du worktree de la tâche, puis `MergeBranch` dans le worktree du chantier. Conflit → 409, marqueur `[merge-conflict:...]`, tâche laissée en `review`. Aucune confirmation (rien ne sort de la machine). Si le SHA de HEAD du chantier a changé (des commits ont réellement rejoint la branche), `MarkCardBranchPending` : le chantier redevient livrable. Refuse en 409 une tâche dont `Rebasing` est vrai. Termine en lançant `rebaseSiblingTasks` dans une goroutine (`rebaseWG` pour que les tests l'attendent, `rebaseMu` pour sérialiser) : les autres tâches en revue du chantier sont rejouées sur la branche du chantier, avec `SetTaskRebasing` (jamais de `UpdatedAt`) autour de chaque rebase et un marqueur `[rebased:...]` ou `[rebase-conflict:...]` au fil.
- GET /cards/{id}/delivery : aucune écriture git ; commence par constater les acceptations déjà faites (`autoAcceptMergedTasks` : tâche en revue, aucun agent en cours, `filesCount > 0`, worktree propre, branche déjà contenue dans celle du chantier) avant de calculer l'aperçu. Publie `task`/`message`/`cards`/`agents` pour chaque tâche ainsi acceptée.
- POST /cards/{id}/catch-up : pour chaque branche de chantier, `MergeBranch(worktree du chantier, destination, "Sillage: catch up with <destination>")` — donc une fusion, jamais un rebase (les branches de tâche descendent de celle du chantier). Ne fait rien si la destination y est déjà (`upToDate`), refuse si le worktree du chantier est sale, annule et rapporte les fichiers en cas de conflit. Un rattrapage réussi appelle `MarkCardBranchPending`.
- POST /cards/{id}/ship : exiger `{"confirm":true}` sinon 400 `"confirmation required"` ; 409 si `card.shipReady` est faux (message dérivé de `shipBlocker`) ; traiter chaque dépôt indépendamment, un échec n'annule pas les autres (erreur portée par la ligne de résultat). Un dépôt dont `pending` vaut 0 est `skipped`. Une `prUrl` déjà connue est réutilisée : le push met à jour la pull request existante, aucune tentative d'en ouvrir une seconde.
- reopen : accepted/cancelled → review.
- Après chaque mutation : publier les SSE nécessaires (`task`, `cards`, `tokens`, `agents`).

## Tests

- store_test.go : roundtrip save/load, compteurs dérivés.
- git_test.go : parser de diff sur une fixture inline (2 fichiers, add/del/ctx, fichier nouveau) ; test worktree+diff sur un repo git temporaire créé dans t.TempDir().
- `go vet ./...` et `go test ./...` doivent passer. `go build` doit produire le binaire.
