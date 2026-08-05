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

Flags du binaire : `-addr` (défaut `127.0.0.1:8787`), `-data` (défaut `~/.local/share/sillage`), `-version`. Env : `SILLAGE_PASSWORD` (remplace le hash du mot de passe au démarrage, pratique en dev), `SILLAGE_CODEX_SANDBOX` (mode sandbox de codex, défaut `workspace-write`).

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
   Le seul appel réseau que Sillage fait de lui-même est la vérification de mise à jour (`update.go`) : un GET vers l'API GitHub, en lecture, désactivable dans les réglages. Aucune donnée de la machine n'y est envoyée, et un binaire téléchargé n'est jamais posé sans vérification de son sha256.
4. Une seule dépendance Go externe : `golang.org/x/crypto`. En ajouter une demande une très bonne raison.

## Les specs sont le contrat

- `docs/SPEC-API.md` : contrat HTTP/JSON/SSE complet (modèles, endpoints, cycle de vie des tâches, événements SSE, règles UI). **À lire avant toute modification de l'API et à mettre à jour dans le même commit** si le contrat change.
- `docs/SPEC-BACKEND.md` : internes (store, auth, SSE, git, adaptateurs d'agents, seed).
- `docs/SPEC-RECETTE.md` : spec de fournée de la recette manuelle (une commande par dépôt, lancée dans le worktree d'un chantier ou d'une tâche, journal en direct). Le lot 1 est implémenté ; le contrat courant est SPEC-API.md §« Recette manuelle ». Le §5 garde la trace des pistes écartées.
- `docs/SPEC-LIVRAISON.md` : spec de fournée du modèle de livraison (chantier = branche de feature, acceptation, Ship de chantier, réglage de livraison). Contexte du « pourquoi » et lots restants ; le contrat courant reste SPEC-API.md.

## Architecture

Binaire Go unique, frontend vanilla embarqué, persistance JSON, temps réel SSE.

```
main.go                     flags, dataDir, init mot de passe, Store, embed web/, ListenAndServe
internal/server/
  models.go                 tous les structs JSON de SPEC-API.md
  store.go                  état en mémoire + persistance atomique + compteurs dérivés
  handlers.go               routage (net/http ServeMux avec patterns méthode+chemin), middlewares
  runner.go                 adaptateurs claude / codex / copilot / agy / fake, un process max par tâche
  preview.go                recette manuelle : un process par worktree, journal en mémoire
  preview_handlers.go       routes de recette (lancer, arrêter, journal)
  git.go                    worktrees (chantier + tâche), parser de diff unifié, commits, fusions
                            (MergeBranch, MergeLocal, MergeAndPush), pushBranch + SyncPush
                            (les deux seuls push), Ship, OpenPR (gh/glab)
  update.go                 version du binaire, détection de mise à jour (GitHub), application
                            selon le mode d'installation (brew / binaire), redémarrage sur place
  workspace.go              dataDir en dépôt git optionnel (setup, clone, commit auto throttlé)
  auth.go                   bcrypt, sessions en mémoire, rate-limit login
  sse.go                    Hub pub/sub
web/                        index.html + style.css + app.js (SPA vanilla, zéro dépendance)
```

### Store (`store.go`)

`Store` est à la fois l'état en mémoire et le format sur disque : ses champs exportés sont sérialisés tel quel dans `<dataDir>/state.json`. Points structurants :

- Un `sync.Mutex` protège tout. Les helpers `recomputeCard/Project/Agent/All` doivent être appelés **verrou tenu** et recalculent les champs dérivés (progression, compteurs, `unread`, tokens agrégés, `active`).
- `recomputeCard` porte aussi deux règles produit : l'état du bouton de livraison (`ShipReady`/`ShipBlocker`, voir `shipReadiness`) et la colonne de la carte, qui ne passe à `done` que si toutes ses tâches sont terminales **et** que le chantier a été livré (`CardBranch.ShippedAt`). « Terminé » veut dire livré ; un chantier livré qui reçoit du travail nouveau en ressort.
- `save()` écrit un fichier temporaire puis `os.Rename` (atomique), et arme le commit git de l'espace de travail. Ce commit est **throttlé** (`workspaceCommitInterval`, 15 min) et non debouncé : un minuteur en attente n'est jamais repoussé, sinon un agent actif (plusieurs sauvegardes par seconde) empêcherait tout commit. Chaque commit stockant un blob complet de `state.json`, commiter à chaque sauvegarde gonfle le dépôt en objets libres pour aucun gain.
- `stateFormatVersion` (const dans `store.go`) est le format de `state.json`, écrit dans le fichier à chaque sauvegarde. **À incrémenter dès qu'un champ persisté est ajouté, renommé ou change de sens** : c'est la seule protection contre un binaire plus ancien, qui sinon charge le fichier puis le réécrit en supprimant en silence les champs qu'il ne connaît pas (`NewStore` sauvegarde immédiatement après le chargement). Un fichier de format supérieur fait échouer `loadStoreFile` (`ErrStateTooNew`) avant toute écriture, et `main` refuse de démarrer. `Store.WrittenBy` garde la version qui a écrit en dernier : `DowngradeWarning()` prévient au démarrage quand deux versions publiées du même format se marchent dessus (ce que le format ne peut pas voir). Attention : la protection n'est effective qu'entre binaires qui la portent tous les deux.
- Les migrations de format se font au chargement dans `loadStoreFile` (`migrateLegacyRepos`, `migrateLegacyWorkspace`, `migrateTaskStatuses`, `migrateLegacyDelivery`, `migrateAgentSeeds`) en relisant le JSON brut : ajouter une migration là, pas ailleurs. `resetTransientTaskFlags`, au même endroit, éteint les états qui ne décrivent qu'une opération en cours (`Task.Rebasing`).
- `AgentOut.Warning` (santé de l'agent : binaire absent du PATH, sandbox codex bloqué par AppArmor) est calculé à chaque `ListAgents` et **jamais persisté** : `Agent` n'a pas ce champ.
- `ReloadFromDisk()` remplace le contenu sans changer le pointeur `Store`, pour que sessions et abonnements SSE survivent au rapatriement d'un espace de travail.

### Runner (`runner.go`)

Un `procHandle` par tâche au maximum (`map[taskID]*procHandle`). Un message envoyé pendant qu'un agent tourne est mis en file (`pending`) et rejoué à la fin de l'exécution en cours. Les process sont lancés avec `Setpgid` pour qu'`Interrupt` puisse SIGINT le groupe (puis SIGKILL après 5 s).

Cinq adaptateurs, sélectionnés par `agent.cli` :

- **claude** : `claude -p --output-format stream-json --verbose --permission-mode acceptEdits --allowedTools <claudeAllowedTools> [--append-system-prompt ...] [--resume <sessionId>]`. Parse le JSONL : `system/init` → `sessionId` stocké (reprise de conversation), blocs `text` → Messages, blocs `tool_use` → ligne d'activité live, `result` → tokens et coût. `claudeAllowedTools` est une constante figée : ne pas y ajouter d'outil capable de pousser.
- **codex** : `codex exec --json --sandbox <SILLAGE_CODEX_SANDBOX|workspace-write> -C <worktree>`. Pas de reprise de session : l'historique est rejoué via `buildTranscript`. Les événements de tokens portent des **totaux cumulés** : ne garder que le dernier et l'ajouter une seule fois en fin de process (`parseCodexTokenStream`).
- **copilot** : `copilot --autopilot ... -p <prompt>`, outils read/write/shell autorisés avec refus prioritaires pour `git push`, `gh`, `glab` et les réglages Copilot du dépôt. MCP GitHub et contrôle distant désactivés. Sortie texte finale, pas de reprise ni de tokens.
- **agy** : `agy --sandbox --print-timeout=60m --add-dir <worktree> --print <prompt>`. Sandbox toujours forcé, jamais `--dangerously-skip-permissions`. Sortie texte finale, pas de reprise ni de tokens. Deux contraintes du CLI : `--print` prend le prompt **en valeur** (donc toujours en dernier), et en sandbox agy ignore le répertoire de travail du process, d'où `--add-dir` sur le worktree. Les autorisations n'existent pas en drapeau : elles viennent de `~/.gemini/antigravity-cli/settings.json` (`toolPermission` **et** deux règles `read_file`/`write_file` sur la racine des worktrees), que Sillage lit pour avertir (`antigravityWorksHeadlessly` dans `store.go`) et n'écrit que sur clic explicite (`fixAntigravityToolPermission`, `POST /api/agents/{id}/fix-warning`). En mode print, toute demande de confirmation est auto-refusée et la session s'arrête sans rien écrire : une seule suffit à rendre la tâche muette.
- **fake** : simule ~3 s de travail, écrit `SILLAGE-TEST.md` dans le worktree, produit un usage fictif. Aucun process externe.

Le prompt d'un départ frais (lancement initial, ou premier message après réassignation, `sessionId` vide) est préfixé par `Task: <title>\n\n` (`contextualizeCliInput`). Le contexte projet s'ajoute au system prompt pour claude (`buildSystemPrompt`) et en préfixe du prompt pour codex, copilot et agy ; ces deux derniers reçoivent aussi le contexte agent dans ce préfixe.

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

`~/.local/share/sillage/` : `state.json` (tout l'état), `config.json` (`passwordHash`), `worktrees/<taskId>/` (un worktree git par tâche), `worktrees/ws-<cardId>-<repo>/` (un par branche de chantier), `worktrees/.merge-<cardId>-<repo>/` (transitoire, mode de livraison « fusion locale »), et optionnellement un dépôt git de l'espace de travail qui ne versionne que `state.json`, `config.json` et `.gitignore`.
