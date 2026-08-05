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
  preview.go                 // superviseur des recettes manuelles (process + journal)
  preview_handlers.go        // routes de recette
  update.go                  // version, détection de mise à jour, application, redémarrage
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

## Format de state.json

- `stateFormatVersion` (const, store.go) est écrit dans le fichier (`FormatVersion`) à chaque `save()`. **À incrémenter dès qu'un champ persisté est ajouté, renommé ou change de sens.**
- Au chargement, `FormatVersion > stateFormatVersion` fait échouer `loadStoreFile` avec `ErrStateTooNew`, **avant la moindre écriture** : `NewStore` sauvegarde immédiatement après avoir chargé, donc laisser passer un fichier trop récent le mutilerait à la seconde près (les champs exportés du Store sont sérialisés tels quels : ce que le binaire ne connaît pas disparaît). `main` refuse alors de démarrer, avec le message et la sortie de secours (`brew upgrade sillage`, ou un autre `-data`).
- `CloneWorkspace` applique le même refus au `state.json` du clone **avant** `ReplaceWorkspaceFiles` : sinon le rapatriement laisserait le répertoire de données avec un fichier illisible sous un serveur qui tourne encore.
- `Store.WrittenBy` garde la version de Sillage qui a écrit le fichier en dernier ("dev" pour une compilation locale). `DowngradeWarning()` prévient au démarrage quand le fichier vient d'une version publiée plus récente que le binaire **à format égal** : le chargement réussit, mais les champs apparus entre les deux tomberont à la première sauvegarde. Silencieux dès qu'une version "dev" est en jeu, faute de pouvoir comparer.
- Limite à connaître : le garde-fou n'existe qu'entre binaires qui le portent tous les deux. Une version antérieure à son introduction ignore `FormatVersion` et le supprime même du fichier en le réécrivant. Le filet de secours reste le dépôt git de l'espace de travail (commits `sillage: update`, throttlés à 15 min).

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

Le quota de compte (`rate_limits`) n'est PAS porté par ce flux stdout, seulement par le fichier de session que codex écrit de son côté même en mode `exec` (`~/.codex/sessions/AAAA/MM/JJ/rollout-...-<thread_id>.jsonl`) : `readCodexRateLimits` retrouve ce fichier via le `thread_id` capturé sur l'événement `thread.started` du flux, lit son dernier `rate_limits` et met à jour `Store.CodexQuota` (best-effort, jamais bloquant). Voir SPEC-API.md « Quota des agents ».

### Adaptateur fake (`cli:"fake"`)

Sans exec : goroutine qui simule ~3 s de travail : 3 lignes d'activité espacées, écrit/complète un fichier `SILLAGE-TEST.md` dans le worktree (contenu horodaté), un Message agent de synthèse, usage fictif `{input:1200, output:340, costUsd:0.004}`. Sert aux tests et à la démo sans coût.

## Superviseur de recette (preview.go)

Voir `docs/SPEC-RECETTE.md` pour le « pourquoi », SPEC-API.md §« Recette manuelle » pour le contrat.

- `PreviewSupervisor` : `map[worktreeDir]*previewProc` protégée par un mutex, plus un compteur de runs (`pv1`, `pv2`, ...). Vit dans `Server`, **rien dans `Store`** : aucun run, aucun journal n'atteint `state.json`.
- `Start(previewTarget)` : arrête d'abord le run du même worktree et **attend sa mort** (le port doit être libre), puis `sh -c <cmd>` avec `Dir` = le worktree, `Setpgid: true`, et l'environnement du serveur enrichi des quatre variables (`previewEnv`). Un échec de `cmd.Start()` n'est pas une erreur HTTP : le run est enregistré en `failed` avec son message, parce que c'est une information de recette (binaire absent, droits), pas une panne du produit.
- Un seul tuyau (`os.Pipe`) pour stdout **et** stderr : les erreurs doivent apparaître dans l'ordre où elles arrivent. Une goroutine (`pump`) scanne les lignes, remplit le tampon circulaire (2000 lignes) et publie `previewLog`. Le hub SSE abandonne les événements des clients lents, donc une commande très bavarde ne bloque jamais le process recetté.
- `wait` distingue l'arrêt humain (`stopped`, drapeau atomique posé par `Stop`) de la sortie naturelle (`exited` + `exitCode`), puis publie `preview`.
- `expandPreviewURL` développe `Repo.previewUrl` en lançant `printf '%s' "<gabarit>"` dans le même environnement : le gabarit accepte donc la même syntaxe que la commande, `$(( ))` compris, sans réimplémenter l'arithmétique du shell en Go. `shellDoubleQuote` échappe `"`, `\` et le backtick, mais **pas** `$` (c'est tout l'intérêt).
- `killGroup(cmd, done)` (dans runner.go) est partagé avec l'interruption d'un agent : SIGINT au groupe, SIGKILL après 5 s. Viser le groupe et non le seul `sh` est indispensable : un `npm run dev` laisse sinon son node en vie sur le port.
- `Server.Shutdown()` appelle `StopAll()` (en parallèle, avec attente) puis arrête l'auto-sync. `main.go` l'appelle sur SIGINT/SIGTERM avant `httpSrv.Shutdown`. Les agents, eux, ne sont pas interrompus : leur travail est dans un worktree et survit au redémarrage.

## Mises à jour (update.go)

Voir SPEC-API.md §« Mises à jour de Sillage » pour le contrat.

- La version vit dans `main.version` (ldflags de release) et est poussée dans le paquet par `server.SetVersion` au démarrage. `buildVersion` non publiable (`"dev"`) éteint tout : `updateChecksEnabled()` est faux, aucune goroutine ne démarre, aucun appel réseau n'est fait. C'est aussi ce qui garantit qu'aucun test ne sort sur le réseau (`NewServer` est appelé sans jamais passer par `SetVersion`).
- `updateTracker` (dans `Server`, jamais dans `Store`) garde la dernière version vue, la date de vérification et la dernière erreur. Rien de tout ça n'atteint `state.json` : ce serait une observation de la machine courante dans un fichier qui peut être rapatrié sur une autre.
- `detectInstall()` est mémoïsé (`sync.Once`) : un binaire ne se déplace pas pendant qu'il tourne. La détection Homebrew passe par `filepath.EvalSymlinks` (le PATH ne montre qu'un lien vers le Cellar). L'écriture possible du dossier est testée réellement (fichier temporaire créé puis supprimé) : les permissions seules ne disent rien d'un montage en lecture seule.
- `startUpdateChecker`/`stopUpdateChecker` suivent le patron de `startAutoSync` (même canal `stop`, même idempotence). Une première vérification part 3 s après le démarrage, puis un ticker de 24 h.
- `applyUpdate()` refuse si un agent tourne (`Runner.RunningCount()`) ou si une recette tourne (`PreviewSupervisor.RunningCount()`) : on ne remplace pas un binaire sous les pieds d'un process qu'on a lancé. Le drapeau `applying` empêche deux applications simultanées et est publié en SSE.
- `downloadRelease` vérifie le sha256 pendant l'écriture (`io.MultiWriter`) contre le `checksums.txt` de la release, n'accepte qu'une empreinte sha256 bien formée comme référence (`parseChecksums`), et ne pose le fichier qu'après vérification, par `os.Rename` depuis un temporaire du même dossier. Aucun résidu en cas d'échec.
- `restartInPlace` fait un `syscall.Exec` : le handler répond **avant** (avec un flush et 500 ms de grâce), sinon le navigateur ne verrait qu'une socket fermée. Après Homebrew, la cible est le `sillage` du PATH, pas le chemin de départ : le Cellar de l'ancienne version peut avoir disparu.
- `refreshServiceStatus()` sonde le lancement à l'ouverture de session via `brew services info --json`, uniquement pour une installation Homebrew (les autres modes n'ont pas de registre à interroger). Appelé dans une goroutine au démarrage, indépendamment de la version et du réglage de vérification, et à chaque `POST /api/update/check`. `serviceRegistered` retombe sur `status` quand le champ `registered` est absent (brew ancien) et renvoie `ok=false` quand rien ne permet de conclure : le champ `service` reste alors absent de la réponse, plutôt que d'affirmer que rien n'est configuré.
- Indirections de test : `fetchLatestReleaseFn`, `downloadReleaseFn`, `brewUpgradeFn`, `detectInstallFn`, `releaseDownloadBase`, `lookPathFn`, `brewServiceInfoFn`. Même intention que `syncPushFn`/`workspaceGitEnabledFn` dans workspace.go.

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
- POST /cards/{id}/preview, /tasks/{id}/preview : aucune confirmation (lancer un process local est réversible et n'a rien de sortant). Le nom de dépôt est optionnel quand le chantier n'a qu'une branche (`cardBranchByRepo`). L'identité vient de la référence courte : `ws-<card.ref>` ou `t-<task.ref>`.
- reopen : accepted/cancelled → review.
- Après chaque mutation : publier les SSE nécessaires (`task`, `cards`, `tokens`, `agents`).

## Tests

- store_test.go : roundtrip save/load, compteurs dérivés, signature du format à la sauvegarde, refus d'un state.json plus récent (fichier laissé intact), avertissement de retour en arrière entre versions publiées.
- git_test.go : parser de diff sur une fixture inline (2 fichiers, add/del/ctx, fichier nouveau) ; test worktree+diff sur un repo git temporaire créé dans t.TempDir().
- preview_test.go : commande qui finit (journal, code de retour), stderr capturé, les quatre variables et le répertoire d'exécution (le worktree, jamais le dépôt du projet), identité propre d'une tâche, URL développée avec arithmétique, arrêt qui tue le groupe de process, relancement qui remplace le run, `StopAll` qui ne laisse rien vivant, refus sans commande ou sans branche de chantier, URL non http(s) refusée, tampon de journal plafonné.
- update_test.go : sonde de service (aucun process brew hors installation Homebrew, PID qui identifie l'instance de service, silence quand brew ne conclut pas, repli sur `status`, plus une fixture de la sortie réelle de `brew services info --json`), comparaison de versions (`0.10.0 > 0.9.9`, rc ignorée), lecture de checksums (lignes invalides écartées), compilation locale qui ne sort jamais sur le réseau, chaque `blocker` et son `selfUpdatable`, refus tant qu'un agent travaille, refus sans `confirm`, et le test de sécurité : un sha256 qui ne correspond pas ne remplace rien et ne laisse aucun temporaire (serveur HTTP local via `httptest`).
- `go vet ./...` et `go test ./...` doivent passer. `go build` doit produire le binaire.
