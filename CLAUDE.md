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

Flags du binaire : `-addr` (défaut `127.0.0.1:8787`), `-data` (défaut `~/.local/share/sillage`). Env : `SILLAGE_PASSWORD` (remplace le hash du mot de passe au démarrage, pratique en dev), `SILLAGE_CODEX_SANDBOX` (mode sandbox de codex, défaut `workspace-write`).

Le frontend est embarqué via `//go:embed web` : toute modification de `web/` demande un rebuild + redémarrage pour être visible.

## Ce que fait le produit

Sillage pilote des agents IA de code (Claude Code, Codex CLI) depuis un kanban web. Un projet regroupe un ou plusieurs dépôts git ; une carte est un objectif ; une tâche est un agent qui travaille dans un **worktree git dédié** sur une branche `sillage/<ref>-<slug>`. L'humain relit, accepte, livre.

Pour tester le flux complet sans coût, utiliser l'agent seedé Écho 🧪 (`cli:"fake"`, simulé, sans process externe).

## Invariants non négociables

Ce sont les promesses de sécurité du produit (voir `CONTRIBUTING.md`) :

1. **`git push` n'existe qu'à deux endroits** : `Ship()` dans `internal/server/git.go` (branche d'une tâche) et `SyncPush()` dans `internal/server/workspace.go` (synchronisation de l'espace de données, jamais un dépôt de projet). Aucune entrée capable de pousser dans les allowlists d'outils des agents, aucun flag de contournement de permissions (`--dangerously-skip-permissions` interdit).
2. **Les actions sortantes exigent `{"confirm": true}`** sur une requête authentifiée : ship, ouverture de PR, sync de l'espace de travail.
3. Le serveur reste sûr sur un portable : localhost par défaut, mot de passe bcrypt, rate-limit du login, `Content-Type: application/json` obligatoire sur les mutations (protection CSRF avec SameSite=Lax).
4. Une seule dépendance Go externe : `golang.org/x/crypto`. En ajouter une demande une très bonne raison.

## Les specs sont le contrat

- `docs/SPEC-API.md` : contrat HTTP/JSON/SSE complet (modèles, endpoints, cycle de vie des tâches, événements SSE, règles UI). **À lire avant toute modification de l'API et à mettre à jour dans le même commit** si le contrat change.
- `docs/SPEC-BACKEND.md` : internes (store, auth, SSE, git, adaptateurs d'agents, seed).
- `docs/SPEC-V2.md`, `SPEC-V3.md`, `SPEC-V3-1.md`, `SPEC-V3-2.md` : specs de fournées successives, historiques (contexte du « pourquoi »), pas des références courantes.

## Architecture

Binaire Go unique, frontend vanilla embarqué, persistance JSON, temps réel SSE.

```
main.go                     flags, dataDir, init mot de passe, Store, embed web/, ListenAndServe
internal/server/
  models.go                 tous les structs JSON de SPEC-API.md
  store.go                  état en mémoire + persistance atomique + compteurs dérivés
  handlers.go               routage (net/http ServeMux avec patterns méthode+chemin), middlewares
  runner.go                 adaptateurs claude / codex / fake, un process max par tâche
  git.go                    worktrees, parser de diff unifié, commits, Ship (push), OpenPR
  workspace.go              dataDir en dépôt git optionnel (setup, commit auto, SyncPush)
  auth.go                   bcrypt, sessions en mémoire, rate-limit login
  sse.go                    Hub pub/sub
web/                        index.html + style.css + app.js (SPA vanilla, zéro dépendance)
```

### Store (`store.go`)

`Store` est à la fois l'état en mémoire et le format sur disque : ses champs exportés sont sérialisés tel quel dans `<dataDir>/state.json`. Points structurants :

- Un `sync.Mutex` protège tout. Les helpers `recomputeCard/Project/Agent/All` doivent être appelés **verrou tenu** et recalculent les champs dérivés (progression, compteurs, `unread`, tokens agrégés, `active`).
- `recomputeCard` porte aussi une règle produit : la colonne de la carte suit ses tâches (toutes terminales → `done` ; du travail actif → `doing`).
- `save()` écrit un fichier temporaire puis `os.Rename` (atomique), et arme un commit git debounced (2 s) de l'espace de travail.
- Les migrations de format se font au chargement dans `loadStoreFile` (`migrateLegacyRepos`, `migrateLegacyWorkspace`) en relisant le JSON brut : ajouter une migration là, pas ailleurs.
- `AgentOut.Warning` (santé de l'agent : binaire absent du PATH, sandbox codex bloqué par AppArmor) est calculé à chaque `ListAgents` et **jamais persisté** : `Agent` n'a pas ce champ.
- `ReloadFromDisk()` remplace le contenu sans changer le pointeur `Store`, pour que sessions et abonnements SSE survivent au rapatriement d'un espace de travail.

### Runner (`runner.go`)

Un `procHandle` par tâche au maximum (`map[taskID]*procHandle`). Un message envoyé pendant qu'un agent tourne est mis en file (`pending`) et rejoué à la fin de l'exécution en cours. Les process sont lancés avec `Setpgid` pour qu'`Interrupt` puisse SIGINT le groupe (puis SIGKILL après 5 s).

Trois adaptateurs, sélectionnés par `agent.cli` :

- **claude** : `claude -p --output-format stream-json --verbose --permission-mode acceptEdits --allowedTools <claudeAllowedTools> [--append-system-prompt ...] [--resume <sessionId>]`. Parse le JSONL : `system/init` → `sessionId` stocké (reprise de conversation), blocs `text` → Messages, blocs `tool_use` → ligne d'activité live, `result` → tokens et coût. `claudeAllowedTools` est une constante figée : ne pas y ajouter d'outil capable de pousser.
- **codex** : `codex exec --json --sandbox <SILLAGE_CODEX_SANDBOX|workspace-write> -C <worktree>`. Pas de reprise de session : l'historique est rejoué via `buildTranscript`. Les événements de tokens portent des **totaux cumulés** : ne garder que le dernier et l'ajouter une seule fois en fin de process (`parseCodexTokenStream`).
- **fake** : simule ~3 s de travail, écrit `SILLAGE-TEST.md` dans le worktree, produit un usage fictif. Aucun process externe.

Le prompt d'un départ frais (lancement initial, ou premier message après réassignation, `sessionId` vide) est préfixé par `Task: <title>\n\n` (`contextualizeCliInput`). Le contexte projet s'ajoute au system prompt pour claude (`buildSystemPrompt`) et en préfixe du prompt pour codex.

Toute mutation d'état publie les événements SSE correspondants via les helpers `publishTask/publishMessage/publishCards/publishTokens/publishAgents/publishActivity/...` : oublier une publication laisse le frontend désynchronisé jusqu'au prochain reload.

### Frontend (`web/app.js`)

SPA vanilla d'un seul fichier, dans une IIFE, découpée en sections commentées (i18n, constantes, état, utilitaires, couche API, hydratation, routage, rendus, modales, SSE, amorçage). Conventions :

- **Rendu par chaînes HTML** (`build*HTML()`) réinjectées dans `#sidebar` / `#main` / `#modal-root`. Toujours passer les données utilisateur par `escapeHtml`.
- **Délégation d'événements globale** : un seul `onGlobalClick` dispatche sur `data-action` (+ `data-*` pour les paramètres). Ajouter un bouton = ajouter un `data-action` et un `case`.
- **Routage par hash** (`#/inbox`, `#/projects`, `#/p/{id}`, cartes et tâches) : `navigateTo()` change le hash, `applyRoute()` applique l'état au `hashchange`.
- **i18n** : chaque chaîne visible passe par `t('cle')` / `tCount()`, et **doit exister dans les deux dictionnaires `fr` et `en`** de `I18N` (le fallback est `fr`). Langue détectée du navigateur, surchargée par les Settings et `localStorage`.
- **État** : hydraté par `GET /api/state`, maintenu par SSE (`EventSource`), re-fetch complet à la reconnexion.
- Confirmations en deux temps génériques (`data-action="confirm-click"`) pour ship, PR, sync, suppressions : le bouton devient « Confirmer ? » avant d'agir.

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

Le parser de diff se teste sur une fixture inline (`TestParseDiffFixture`).

## Données à l'exécution

`~/.local/share/sillage/` : `state.json` (tout l'état), `config.json` (`passwordHash`), `worktrees/<taskId>/` (un worktree git par tâche), et optionnellement un dépôt git de l'espace de travail qui ne versionne que `state.json`, `config.json` et `.gitignore`.
