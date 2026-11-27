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
  runner.go                  // adaptateurs claude / codex / agy / kiro / fake
  acp.go, acp_runner.go      // client ACP (JSON-RPC sur stdio) et adaptateur commun runACP
  copilot.go                 // adaptateur copilot, par ACP
  preview.go                 // superviseur des recettes manuelles (process + journal)
  preview_handlers.go        // routes de recette
  attachments.go             // images jointes : disque, service HTTP, chemins passés au CLI
  update.go                  // version, détection de mise à jour, application, redémarrage
  handlers.go                // REST
  store_test.go, git_test.go // tests unitaires
```

## Flags et config

- `-addr` (défaut `127.0.0.1:8787`), `-data` (défaut `~/.local/share/sillage`).
- Une seule instance par dossier de données (`instance.go`) : `flock(LOCK_EX|LOCK_NB)` sur `<dataDir>/sillage.lock` avant le chargement de l'état, fichier gardé ouvert jusqu'à l'arrêt, `{pid, addr, version}` écrit dedans une fois le port lié. Verrou pris par un autre process : lancement ordinaire refusé (code 1, message avec le pid et l'URL de l'instance).
- `-app` : fenêtre d'application (`app.go`). Verrou pris : lit l'adresse publiée (attente jusqu'à 15 s, le temps que l'instance charge son état), vérifie que `GET /` répond avec le titre « Sillage », ouvre la fenêtre et sort. Verrou libre : démarre le serveur et ouvre la fenêtre dès que le port est lié. Navigateur : `SILLAGE_BROWSER`, puis Chrome, Chromium, Edge, Brave, Vivaldi en `--app=<url>`, à défaut `xdg-open` / `open`. `-install-app` (Linux) écrit `$XDG_DATA_HOME/applications/sillage.desktop` et l'icône `icons/hicolor/scalable/apps/sillage.svg` (référencée par chemin absolu : un `icon-theme.cache` préexistant la masquerait), en reportant `-addr` et `-data` s'ils diffèrent des défauts.
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
- `resetTransientTaskFlags` clôt le chargement (après les migrations) en éteignant ce qui ne décrivait qu'un processus du serveur précédent : `Task.Rebasing`, et les tâches `running`, qui repassent en `review` non lues, sans `liveActivity`, avec un marqueur `[interrupted:server-restart]` au fil (`appendMessage`, sans verrou ni sauvegarde : personne ne voit encore le Store) et leurs compteurs relus par `TaskWorkCounts` (git.go, best-effort : worktree disparu = zéros). Un `recomputeAll` suit, les compteurs de carte dépendant du statut. Voir SPEC-API.md § « Cycle de vie des tâches » pour le pourquoi côté produit.
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
- `RepairWorktrees(repoPath)` : `git worktree repair` lancé sans argument depuis `repoPath`, best-effort. Un worktree lié garde en dur, dans son fichier `.git` pointeur, le chemin absolu du dépôt au moment de sa création : si ce dossier est déplacé sur le disque, tous ses worktrees liés (tâches et chantiers) deviennent illisibles (« not a git repository ») jusqu'à cette réparation. Appelée par `repairMovedRepoWorktrees` (handlers.go) depuis `handleUpdateProject`, pour chaque repo dont le `path` a changé (comparaison par `Name` entre l'état du projet avant et après la mise à jour) : le cas visé est un dossier de code déplacé, avec des tâches ou un chantier déjà en cours dessus.
- `Diff(dir, base)` : d'abord `git add -A -N` (fait apparaître les fichiers non suivis), puis `git diff <base>` ; parser le unified diff en structure de SPEC-API.md (fichiers, additions/deletions, hunks, lignes ctx/add/del). Parser robuste : ignorer les lignes binaires, gérer rename.
- `Commits(dir, base)` : `git log <base>..HEAD --pretty=format:%h|%s|%cr`. `CountCommits(dir, from, to)` : `git rev-list --count from..to` (sert à `pending`, ce qu'il reste à livrer).
- `CommitAll(dir, message)` : `git add -A` puis commit si l'arbre est sale (sans erreur s'il est propre). `IsWorktreeClean(dir)`, `HeadSha(dir)`, `IsBranchMergedInto(dir, branch, target)` : constats en lecture seule.
- `FastForwardBranch(dir, branch)` : `git merge --ff-only <branch>` dans le worktree du chantier, à l'acceptation, après `RebaseOnto` si la tâche ne part plus de la pointe du chantier : aucun merge commit. `MergeBranch(dir, branch)` : `git merge --no-edit <branch>` (fast-forward si possible), pour le rattrapage de la destination. En cas de conflit : liste les fichiers en conflit (`git diff --name-only --diff-filter=U`), `git merge --abort`, et enveloppe `ErrMergeConflict`. Aucune résolution automatique.
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
5. Fin de process : compteurs et `workSha` relus depuis git ; statut → `review` (sauf si interrompu : déjà review) ; `unread=true` ; `liveActivity=null` ; SSE `task`, `cards`, `agents`, `activity(null)`. Puis la vérification du projet (`Project.checkCmd`) part en tâche de fond (`check.go`, `Runner.startCheck`, délai 10 min, groupe de process tué à l'annulation) ; son résultat (`Task.checks`, avec la fin de la sortie) arrive par SSE `task`. Tout départ d'agent, rebase, acceptation, refus ou suppression appelle d'abord `Runner.CancelCheck`.

`Interrupt(taskID)` : SIGINT au groupe de process (Setpgid à la création), 5 s, puis SIGKILL. Statut → review.

Images jointes : `Message`/`Start` reçoivent les `Attachment` déjà écrites sur le disque par le handler (attachments.go). Le Message stocké garde le texte de l'utilisateur ; le texte transmis au CLI reçoit en plus les chemins absolus (`withAttachmentPaths`), y compris quand le message passe par la file `pending`. Aucun adaptateur ne transporte l'image elle-même : tous lisent le fichier.

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

`codex exec --json --sandbox <mode> -C <worktree> [--add-dir <git-dir>...] [--model <model>] <texte>` ; parser plusieurs formes de JSONL :
`{"msg":{"type":"agent_message","message":"..."}}`, `{"type":"item.completed","item":{"type":"agent_message","text":"..."}}`, `token_count` → usage si présent. En cas d'échec de parsing, capturer stdout brut en un Message final. Pas de resume en v1 (chaque message relance `codex exec` avec le dernier texte).

En mode `workspace-write`, les worktrees liés gardent leur index/HEAD et leurs objets/références Git hors du répertoire de tâche. Sillage résout ces chemins avec `git rev-parse` et les passe séparément via `--add-dir` : git-dir propre au worktree, puis `objects`, `refs`, `logs` et stockages optionnels existants du common-dir. Le common-dir entier n'est pas exposé pour un worktree lié, afin que `hooks` et la configuration du dépôt d'origine restent en lecture seule. La résolution est best-effort : un répertoire qui n'est plus un dépôt Git ne bloque pas le démarrage de Codex. Les caches des profils de permissions explicitement activés sont eux aussi passés séparément via `--add-dir`, sous `<dataDir>/package-cache/<taskId>/<profil>` ; aucun dossier système n'est élargi.

### Profils de permissions technologiques

`Project.permissionProfiles` accepte `ruby`, `node`, `php`, `python`, `go` et `rust`. Ce sont des choix humains dans les réglages, jamais une détection depuis le dépôt. Chaque profil ajoute ses commandes à l'allowlist Claude et injecte aux process d'agent ainsi qu'au check final des variables qui redirigent les caches de paquets/build vers `<dataDir>/package-cache/<taskId>/<profil>`. Le profil Ruby fixe aussi `BUNDLE_BIN`, `BUNDLE_PATH` et `GEM_HOME` : `bundle install` ne tente donc plus d'écrire dans `/usr/local/bin`. Le profil PHP fixe `COMPOSER_HOME` et `COMPOSER_CACHE_DIR`, puis place les outils Composer globaux dans le `PATH`. Le cache reste propre à une tâche, est réutilisé sur ses tours suivants, n'apparaît jamais dans le diff Git et tombe avec la suppression de la tâche. `workspaceGitignore` exclut `package-cache/` si le dataDir est synchronisé par Git.

Le quota de compte (`rate_limits`) n'est PAS porté par ce flux stdout, seulement par le fichier de session que codex écrit de son côté même en mode `exec` (`~/.codex/sessions/AAAA/MM/JJ/rollout-...-<thread_id>.jsonl`) : `readCodexRateLimits` retrouve ce fichier via le `thread_id` capturé sur l'événement `thread.started` du flux, lit son dernier `rate_limits` et met à jour `Store.CodexQuota` (best-effort, jamais bloquant). Voir SPEC-API.md « Quota des agents ».

### Client ACP (`acp.go`, `acp_runner.go`)

Sillage sait piloter un agent par l'[Agent Client Protocol](https://agentclientprotocol.com) : du JSON-RPC 2.0 sur stdio, un objet par ligne, où Sillage tient le rôle de l'éditeur. `acp.go` porte le protocole (`acpClient` : `initialize`, `session/new`, `session/load`, `session/prompt`, `session/cancel`, lecture des `session/update`, réponse aux `session/request_permission`), `acp_runner.go` l'adaptateur commun `runACP`. Brancher un nouvel agent ACP revient à écrire un `acpAgentSpec` : binaire, arguments, environnement, et les quelques écarts au protocole commun.

Un process par tour, comme les autres adaptateurs. Le déroulé de `runACP` :

1. `initialize` annonce un client sans accès fichiers ni terminal (`fs.readTextFile`, `fs.writeTextFile` et `terminal` à `false`). L'agent garde ses propres outils, bornés par ses propres règles ; toute méthode non servie reçoit « method not found ».
2. `Task.SessionID` connu et agent capable de reprise (`agentCapabilities.loadSession`) : `session/load`. L'agent rejoue alors tout son historique en `session/update` avant de répondre ; ces mises à jour sont ignorées, `acpTurn` ne lit rien hors d'un prompt. Session introuvable : elle est oubliée, une session neuve part avec l'historique du fil rejoué dans le prompt (`prepareCliInput`), comme pour claude. L'identifiant d'une session neuve n'est gardé que si l'agent sait reprendre, sinon le tour suivant se croirait dispensé de rejouer le fil.
3. Le premier prompt d'une session neuve reçoit le contexte (`acpAgentSpec.FreshPrompt`) ; une session reprise le connaît déjà et ne reçoit que le message.
4. Pendant le tour, `acpTurn` transforme le flux : les `agent_message_chunk` s'accumulent et deviennent un Message à chaque nouvel appel d'outil puis à la fin du tour ; chaque `tool_call` donne une ligne d'activité (`acpToolLine`, « genre · détail », chemins rendus relatifs au worktree) ajoutée à `Task.CommandLog` ; un outil en échec avec le code `denied`, ou une demande d'autorisation refusée, pose un marqueur `[tool-denied:...]`, une fois par outil et par tour.
5. La réponse de `session/prompt` porte l'usage : `inputTokens - cachedReadTokens` s'ajoute à `Tokens.Input` (même convention que claude : entrée neuve et écriture de cache, pas la relecture), `outputTokens` à `Tokens.Output`. Pas de coût.
6. L'entrée du process est fermée ; s'il ne sort pas en 3 s son groupe est tué.

**Autorisations.** `acpRejectAll` est la politique par défaut, et la seule employée : toute `session/request_permission` reçoit l'option de refus, ou une annulation s'il n'y en a pas. Ce qu'un agent a le droit de faire est fixé d'avance par ses propres règles (ses drapeaux) ; une demande qui remonte jusqu'à Sillage sort de ce cadre, et personne n'est là pour en juger. Aucune autorisation n'est donc jamais accordée en cours de route, ce qui garde l'invariant 1 hors de portée de ce chemin.

**Interruption.** `procHandle.softStop` envoie `session/cancel` avant tout signal : l'agent répond `stopReason: "cancelled"`, enregistre sa session et sort, ce qui garde la reprise possible au message suivant. Passé `softStopGrace` (5 s), `killGroup` prend le relais. Le texte en attente d'un tour annulé n'est pas publié (phrase coupée, ou accusé de réception de l'annulation).

Un tour qui finit sans le moindre message est un échec (même règle que `runTextCLI`), de même qu'un `stopReason` autre que `end_turn`.

### Adaptateur GitHub Copilot (`cli:"copilot"`, `copilot.go`)

`copilot --acp --no-ask-user --allow-tool=read,write,shell --deny-tool=<refus figés> --disable-builtin-mcps --no-auto-update --sync-session=off [--model=<model>]`, piloté par le client ACP ci-dessus. Reprise de session, outils en direct et compteurs de tokens ; pas de coût ni de quota.

Les autorisations sont entièrement dans ces drapeaux, que le mode ACP applique : lecture, écriture et shell passent sans demande, les refus figés (`git push`, `gh`, `glab`, réglages Copilot du dépôt) l'emportent sur toute autorisation, et le reste (chemin hors du worktree, URL) remonte en demande que `acpRejectAll` refuse. Ni `--allow-all` ni mode autopilot, qui revient à tout autoriser. Le MCP GitHub intégré est désactivé. `--no-remote` et `--no-remote-export` ont été retirés du CLI (il refuse de démarrer avec) : l'export des sessions est coupé par `remoteExport: false` dans les réglages du dossier isolé, et une session annonce `remoteSteerable: false` par défaut.

**Dossier de configuration isolé.** Chaque tâche a son `COPILOT_HOME` : `<dataDir>/agent-homes/<taskId>/copilot/` (`prepareCopilotHome`), supprimé avec la tâche. Copilot charge depuis son dossier les hooks, serveurs MCP et plugins de l'utilisateur, et un hook en échec vaut refus : un seul hook `PermissionRequest` écrit pour un autre outil suffit à refuser toutes les écritures d'une tâche, sans message utile. Le dossier isolé n'hérite de rien. Il persiste d'un tour à l'autre, parce que c'est là que Copilot range la session à reprendre. À chaque tour, `settings.json` est réécrit (`remoteExport` et `autoUpdate` à `false`) et `config.json` reprend deux clés de celui de l'utilisateur, `lastLoggedInUser` et `loggedInUsers` : des noms de compte, pas des secrets, le jeton restant dans le trousseau du système où le CLI le retrouve par compte. Un utilisateur qui a choisi le stockage du jeton en clair (`storeTokenPlaintext`) n'est donc pas authentifié par ce chemin et doit passer `COPILOT_GITHUB_TOKEN` ou `GH_TOKEN`.

`copilotEnv` retire en plus un `GITHUB_TOKEN` classique (`ghp_`), avec lequel le CLI refuse de démarrer.

**Début de tour.** Copilot écrit ses avertissements de démarrage (« Run /login again to use Connectors ») dans le canal des messages, sans rien qui les distingue d'une réponse. Le client s'abonne donc à l'événement `assistant.turn_start` (`clientCapabilities._meta["github.com/copilot"].events`) et jette le texte reçu avant le premier de chaque tour (`acpTurn.discardPrelude`). Le même `_meta` porte `interactive: false`.

### Adaptateur Antigravity (`cli:"agy"`)

`agy --sandbox --print-timeout=60m --add-dir <worktree> [--add-dir <dossier git commun>] [--model <model>] --print <texte>`. Le texte contient les contextes agent, projet et chantier avant le prompt ; le stdout texte devient un Message agent final. Pas de reprise de session ni de compteurs de tokens. Le mode sandbox est toujours forcé ; `--dangerously-skip-permissions` n'est jamais utilisé.

Trois particularités du CLI dictent la ligne de commande :

- `--print` est un alias de `--prompt` et prend le prompt **en valeur**, ce n'est pas un booléen. Il doit rester le dernier drapeau : `agy --print --sandbox <texte>` fait exécuter le prompt « --sandbox », l'agent répond une explication du mode sandbox et ne touche à rien.
- en mode sandbox, agy ignore le répertoire de travail du process et retombe sur son espace interne (`~/.gemini/antigravity-cli/scratch`). `--add-dir <worktree>` est donc obligatoire, sinon la tâche écrit ses fichiers hors du dépôt et le diff reste vide.
- le sandbox ne montre **que** les répertoires ajoutés. Le worktree d'une tâche est un worktree lié : son vrai dossier git est celui du dépôt d'origine, hors du sandbox, donc toute commande git y échoue par « ce n'est pas un dépôt git ». Le modèle relance alors la commande hors sandbox (`BypassSandbox: true`), demande auto-refusée en mode print : la session s'arrête sans rien produire (observé sur un simple `git status`). D'où le second `--add-dir`, le dossier git commun résolu par `GitCommonDir` (`git rev-parse --path-format=absolute --git-common-dir`), best-effort : sans lui l'agent travaille quand même sur les fichiers, il perd seulement git.

Ce second répertoire n'ouvre aucune porte (invariant 1) : un `--add-dir` hors espace de travail principal est monté **en lecture seule** (vérifié : `touch` y répond « système de fichiers accessible en lecture seulement » et `git commit` échoue sur `index.lock`), et le sandbox n'a pas de réseau (la résolution DNS y échoue). L'agent gagne donc exactement les lectures git de l'allowlist claude (`status`, `log`, `diff`, `show`), sans écriture de hooks ni de config, et sans push. Le prompt le lui dit (`antigravityWorktreeNote`), avec la consigne de ne jamais relancer une commande hors sandbox mais de rapporter l'échec : c'est ce qui évite les tâches muettes.

Les autorisations ne s'expriment **pas** en ligne de commande : le seul drapeau existant est `--dangerously-skip-permissions`, que Sillage n'utilise jamais. Tout vient de `~/.gemini/antigravity-cli/settings.json`. En mode print, **toute** demande de confirmation est auto-refusée et la session s'arrête aussitôt : stdout vide, sortie 0, explication sur stderr. Une seule demande suffit donc à rendre une tâche muette, et deux réglages sont nécessaires (`antigravityWorksHeadlessly` dans `store.go`) :

- `toolPermission` (valeurs `always-proceed`, `request-review`, `strict`, `proceed-in-sandbox`) : le défaut `request-review` fait demander un accord à chaque commande. Sillage attend `proceed-in-sandbox`, cohérent avec le `--sandbox` toujours forcé.
- deux règles dans `permissions.allow` : `read_file(<dataDir>/worktrees)` et `write_file(<dataDir>/worktrees)`. Les fichiers ordinaires du worktree passent sans règle, mais certains chemins demandent un accord : `.git`, qui dans un worktree lié est un fichier d'une ligne qu'un modèle en exploration essaie d'ouvrir (vérifié), et des modifications de fichiers existants (observé une fois sur `web/index.html`, non reproduit ; le message de la CLI demande exactement une règle `write_file(<cible>)`).

Faits vérifiés sur la syntaxe des règles, qui expliquent la forme retenue : une cible **dossier** couvre tout ce qu'elle contient, les motifs (`read_file(<dossier>/*)`, `read_file(<racine>/**)`) ne correspondent à rien, et ajouter ces règles ne restreint rien par ailleurs (une lecture hors worktree était déjà refusée sans elles, elle l'est toujours avec). Deux règles sur la racine des worktrees suffisent donc, sans rien accorder au-delà du terrain de travail des agents.

Sillage n'écrit ce fichier que sur clic explicite de l'humain (`POST /api/agents/{id}/fix-warning` → `fixAntigravityToolPermission`) : relecture des clés et des règles existantes, ajout de ce qui manque seulement, écriture atomique (fichier temporaire + rename) en conservant les droits, refus pur et simple si le JSON existant est illisible.

Reste un mode d'échec assumé : un agent qui veut lire ou écrire **hors** de son worktree fait tomber sa session (par exemple `/etc/hostname`, vérifié). C'est le comportement souhaitable, et il est maintenant visible dans la conversation. L'adaptateur prévient aussi l'agent dans son prompt (`antigravityWorktreeNote`) que `.git` est un pointeur qu'il ne doit pas ouvrir.

`runTextCLI` (partagé avec kiro) traite une sortie 0 sans rien sur stdout comme un échec et remonte stderr : sans ça, la tâche finissait « à relire » avec une conversation vide, ce qui ne laissait aucune trace du refus.

### Adaptateur Kiro (`cli:"kiro"`)

`kiro-cli chat --no-interactive --agent sillage --wrap never <texte>`. Le texte contient les contextes agent, projet et chantier avant le prompt ; le stdout texte devient un Message agent final. Pas de reprise de session ni de compteurs de tokens : le transcript récent est rejoué dans le prompt comme pour les autres adaptateurs one-shot.

Le mode headless exige `KIRO_API_KEY` dans l'environnement du processus Sillage. La santé de l'agent signale son absence avant le premier run.

Chaque exécution reçoit un `KIRO_HOME` créé par `os.MkdirTemp`, distinct du dépôt et du profil Kiro de l'utilisateur. Il contient un agent `sillage.json` avec `read`, `write`, `shell`, `grep` et `glob`, sans MCP hérité (`includeMcpJson:false`). `write` n'est pas placé dans `allowedTools` : son `allowedPaths:["**"]` limite l'approbation automatique aux chemins relatifs au worktree. Même règle pour `shell`, qui passe par `toolsSettings` afin que les refus soient évalués avant l'autorisation générale des commandes locales. Les refus figés couvrent `git push` (y compris `git -C ... push`), `gh`, `glab` et les sous-commandes Kiro qui modifieraient agents, réglages ou intégrations. Ne jamais remplacer ce profil par `--trust-all-tools`, qui court-circuiterait cette granularité. Le `KIRO_HOME` temporaire est supprimé après la mort du process, succès, erreur ou interruption.

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

Agents : Bolt 🐝 `#f2b705` claude/sonnet ; Muse 🦊 `#d0662f` claude/opus ; Otto 🦉 `#4f7d2f` codex/(modèle vide = défaut) ; Fably 🪶 `#6b4fbb` claude/fable ; Octo 🐙 `#24292f` copilot/(modèle par défaut) ; Astro 🚀 `#4285f4` agy/(modèle par défaut) ; Kiro 🟣 `#7c3aed` kiro/(modèle par défaut) ; Écho 🧪 `#777` fake (agent de test local, gratuit). Pas de projets seedés. Chaque nouvelle vague d'agents externes est migrée une seule fois via `AgentSeedVersion`, afin qu'une suppression volontaire reste définitive ; le passage de la version 1 à 2 n'ajoute que Kiro et ne recrée donc pas un profil Copilot ou Antigravity supprimé auparavant.

## Handlers

Suivre SPEC-API.md. Points d'attention :
- POST /projects : vérifier que path existe et `git rev-parse --git-dir` réussit, sinon 400 avec message explicite. Sans champ `delivery`, déduire le mode des remotes (`detectDelivery`).
- POST /tasks : garantir la branche du chantier sur ce dépôt (`ensureCardBranch`, créée à la première tâche), puis le worktree de la tâche depuis cette branche ; si échec, 400 et pas de tâche. Lancer l'agent avec `title` + `prompt`. `ensureCardBranch` remet aussi le chantier en « non livré » (`MarkCardBranchPending`) s'il l'était : une tâche de plus veut dire que tout n'est plus livré.
- POST /tasks/{id}/accept : commit du worktree de la tâche, puis `MergeBranch` dans le worktree du chantier. Conflit → 409, marqueur `[merge-conflict:...]`, tâche laissée en `review`. Aucune confirmation (rien ne sort de la machine). Si le SHA de HEAD du chantier a changé (des commits ont réellement rejoint la branche), `MarkCardBranchPending` : le chantier redevient livrable. Refuse en 409 une tâche dont `Rebasing` est vrai. Termine en lançant `rebaseSiblingTasks` dans une goroutine (`rebaseWG` pour que les tests l'attendent, `rebaseMu` pour sérialiser) : les autres tâches en revue du chantier sont rejouées sur la branche du chantier, avec `SetTaskRebasing` (jamais de `UpdatedAt`) autour de chaque rebase et un marqueur `[rebased:...]` ou `[rebase-conflict:...]` au fil.
- GET /cards/{id}/delivery : aucune écriture git ; commence par constater les acceptations déjà faites (`autoAcceptMergedTasks` : tâche en revue, aucun agent en cours, `filesCount > 0`, worktree propre, branche déjà contenue dans celle du chantier) avant de calculer l'aperçu. Publie `task`/`message`/`cards`/`agents` pour chaque tâche ainsi acceptée.
- POST /cards/{id}/catch-up : pour chaque branche de chantier, `MergeBranch(worktree du chantier, destination)` (sujet par défaut de git) — donc une fusion, jamais un rebase (les branches de tâche descendent de celle du chantier). Ne fait rien si la destination y est déjà (`upToDate`), refuse si le worktree du chantier est sale, annule et rapporte les fichiers en cas de conflit. Un rattrapage réussi appelle `MarkCardBranchPending`.
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
