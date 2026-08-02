# Sillage v0.3 : multi-dépôts par projet + synchronisation git de l'espace de travail

Deltas par rapport à docs/SPEC-API.md. Mot d'ordre inchangé : SIMPLE. Pas de nouvelle dépendance. Toutes les erreurs API en anglais.

## 1. Plusieurs dépôts git par projet

Modèle : `Project.repos: [{name, path}]` (name court affichable, unique dans le projet ; path absolu d'un dépôt git valide). Le champ legacy `path` disparaît du modèle mais reste accepté en entrée.

- Migration au chargement du state : projet avec `path` non vide et `repos` vide -> `repos = [{name: basename(path), path}]`.
- POST `/api/projects` : accepte `{name, path}` (devient un repo unique) OU `{name, repos:[{name,path},...]}`. Validation : au moins un repo, chaque path est un dépôt git (`git rev-parse --git-dir`), names uniques et non vides (défaut : basename du path).
- PATCH `/api/projects/{id}` : `{name?, checkCmd?, repos?}` ; repos = liste complète remplaçante, même validation. Retirer un repo ne casse pas les tâches existantes (leur worktree vit sa vie).
- POST `/api/tasks` : nouveau champ `repoName`, optionnel si le projet n'a qu'un repo (défaut), 400 `"repoName required (project has several repositories)"` s'il y en a plusieurs et qu'il manque, 400 si inconnu. Le worktree est créé depuis le repo choisi.
- `Task` gagne `repoName` (persisté, exposé).
- `Card.docsCount/messagesCount` etc. inchangés.

Frontend :
- Modale projet (création et édition) : liste éditable de dépôts (lignes nom + chemin, bouton discret "+ dépôt", retrait par ligne). En création, un seul champ chemin visible par défaut avec "+ dépôt" pour en ajouter (le nom se déduit du basename, éditable).
- Modale nouvelle tâche : si le projet a plusieurs dépôts, un select "Dépôt" (défaut : premier). Sinon rien d'affiché.
- Détail de tâche : si le projet a plusieurs dépôts, afficher le nom du repo en petit chip gris à côté de la branche.

## 2. Synchronisation git de l'espace de travail

Le répertoire de données (dataDir) devient un dépôt git optionnel. Seuls `state.json`, `config.json` et `.gitignore` sont versionnés ; `.gitignore` contient `worktrees/` et `*.tmp`.

Modèle persisté (dans state.json) : `Workspace { setupDone: bool, syncRemote: string, lastSyncAt: time|null }`.

### Comportement

- **Commit local automatique** : si dataDir est un dépôt git (`.git` présent), après chaque sauvegarde du state, un commit debounced (2 s) : `git add state.json config.json .gitignore && git commit -m "sillage: update"` (silencieux si rien à committer). Jamais de push automatique.
- **Onboarding** (après login) : si `!workspace.setupDone`, le frontend affiche une modale de bienvenue avec 3 choix :
  1. "Travailler en local" -> POST setup `{mode:"local"}` : marque setupDone, pas de git.
  2. "Initialiser un dépôt de sauvegarde" (remote optionnel) -> `{mode:"init", remote?}` : `git init` dataDir (branche main), .gitignore, premier commit ; si remote fourni : `git remote add origin <remote>` (pas de push automatique).
  3. "Rapatrier un espace existant" (URL requise) -> `{mode:"clone", remote}` : clone dans un répertoire temporaire ; vérifier que `state.json` existe à la racine du clone, sinon 400 `"remote does not look like a Sillage workspace"` ; puis remplacer state.json/config.json/.git dans dataDir par ceux du clone et recharger le store en mémoire (les sessions actives restent valides). Attention : le mot de passe devient celui de l'espace rapatrié (l'UI l'affiche comme avertissement avant de confirmer).
- Un install existant (state déjà rempli, setupDone absent) : migration -> setupDone=true, mode local (ne JAMAIS afficher l'onboarding sur un espace déjà utilisé ; l'activation de git se fait alors via la modale réglages).

### API

- GET `/api/workspace` -> `{setupDone, gitEnabled, remote, dirty, lastCommitAt, lastSyncAt}` (dirty = commit auto en attente ou modifications non commitées).
- POST `/api/workspace/setup` `{mode:"local"|"init"|"clone", remote?}` : une seule fois (400 si setupDone), sauf `{mode:"init"}` autorisé plus tard pour activer git sur un espace local (depuis les réglages).
- PATCH `/api/workspace` `{remote}` : définit/remplace origin (`git remote add/set-url`). 400 si git non initialisé.
- POST `/api/workspace/sync` `{confirm:true}` (action externe : 400 sans confirm) : commit d'abord si nécessaire, puis `git pull --rebase origin main`, puis `git push -u origin main`. En cas de conflit de rebase : `git rebase --abort` puis 409 `{"error":"sync conflict: the remote workspace diverged, resolve manually in <dataDir>"}`. Timeouts 120 s. Succès -> `{output, lastSyncAt}` et workspace.lastSyncAt mis à jour.
- SSE : event `workspace` avec l'objet workspace après chaque changement.
- GET `/api/state` inclut `workspace`.

Toutes les commandes git de ce module : GIT_TERMINAL_PROMPT=0, jamais interactives. Le push de sync est le seul autre endroit autorisé à exécuter `git push`, dans une fonction dédiée `SyncPush` de git.go, clairement commentée ; il ne pousse QUE le dépôt de l'espace de travail (dataDir), jamais un dépôt de projet.

### UI réglages et sync

- Pied de sidebar : icône ⚙ (pour tous) -> modale "Espace de travail" : état (local seul / git sans remote / git + remote), champ remote (placeholder `git@github.com:vous/sillage-workspace.git`), avertissement sobre "Dépôt privé recommandé : l'espace contient vos conversations", bouton principal "Synchroniser" (double clic de confirmation, comme ship/PR), dernier commit et dernière sync affichés. Si git non activé : bouton "Activer la sauvegarde git" (-> setup init).
- Après une sync réussie : toast/inline discret "Synchronisé il y a un instant".
- i18n : toutes les nouvelles chaînes en fr ET en.

## 3. Tests

- Backend : migration path->repos ; validation repos (path invalide, names dupliqués) ; repoName requis si plusieurs repos ; workspace setup local/init (clone testé avec un remote git local file://) ; sync avec remote bare local (commit auto + push + pull) ; conflit de sync -> 409 et rebase abort ; SyncPush ne touche jamais un repo de projet.
- `gofmt`, `go vet`, `go test`, `node --check` verts.

## Hors périmètre v0.3

Sync automatique vers le remote, résolution de conflits dans l'UI, multi-branches de l'espace, chiffrement.
