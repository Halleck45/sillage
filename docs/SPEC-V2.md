# Sillage v0.2 : spec des évolutions

Note : la section 1 (multi-utilisateur) a été implémentée puis retirée le 2026-08-02, hors périmètre du produit.

Deltas par rapport à docs/SPEC-API.md et docs/SPEC-BACKEND.md. Mot d'ordre : SIMPLE. Pas de framework, pas de nouvelle dépendance Go, pas de refonte.

## 1. Multi-utilisateur (simple, espace partagé)

Un seul espace de travail partagé par tous les utilisateurs. Pas de permissions par projet, pas de non-lus par utilisateur (l'état unread reste global).

Modèle : `User { id, name, role: "admin"|"member", passwordHash }` stocké dans state.json (hash jamais exposé par l'API : struct de sortie sans le champ).

- Migration au démarrage : si state.json ne contient aucun utilisateur et que config.json a un `passwordHash`, créer `admin` (role admin) avec ce hash. Si aucun hash nulle part : générer le mot de passe initial comme avant, pour `admin`.
- `SILLAGE_PASSWORD` (env) : remplace le mot de passe de `admin` au démarrage (créé si absent).
- POST `/api/login` : `{username, password}` (le champ `password` seul n'est plus accepté). Sessions : token -> {userID, expiry}.
- GET `/api/me` -> `{id, name, role}`.
- Admin uniquement (403 sinon) :
  - GET `/api/users` -> liste (sans hash)
  - POST `/api/users` `{name, password, role?}` (role défaut member ; name unique, non vide)
  - PATCH `/api/users/{id}` `{password?, role?}`
  - DELETE `/api/users/{id}` : refusé (400) pour soi-même et pour le dernier admin
- `Message` gagne `authorName` (nom de l'utilisateur pour author=user ; nom de l'agent pour author=agent). Renseigné à la création, affiché dans le fil.
- GET `/api/state` inclut `me: {id, name, role}`.

## 2. CRUD agents

- POST `/api/agents` `{name, emoji, color, cli, model, contextPrompt}` : name et cli obligatoires ; cli dans {claude, codex, fake} ; id = slug du name (unique, 400 sinon).
- PATCH `/api/agents/{id}` : mêmes champs, tous optionnels.
- DELETE `/api/agents/{id}` : 400 si une tâche référence l'agent.
- SSE `agents` déjà existant : publier après chaque mutation.
- Accessible à tous les utilisateurs connectés (pas réservé admin).

## 3. Projets : édition

- PATCH `/api/projects/{id}` `{name?, checkCmd?}`. `checkCmd` exposé dans le JSON Project. SSE : republier le state léger nécessaire (event `project` avec le Project complet ; le front met à jour).
- DELETE : hors périmètre v0.2.

## 4. Ouvrir la PR

- POST `/api/tasks/{id}/pr` `{confirm:true}` (400 sans confirm) ; tâche `shipped` uniquement (400 sinon).
- Implémentation : dans le worktree, `gh pr create --head <branch> --title <titre tâche> --body <corps>` (timeout 60 s, env hérité). Corps : une ligne sobre "Created with Sillage" + résumé (titre). Si `gh` réussit : extraire l'URL de stdout -> `{url}`.
- Fallback sans gh ou en échec : si `git remote get-url origin` pointe vers github.com (ssh ou https), renvoyer `{url}` = `https://github.com/<owner>/<repo>/compare/<base>...<branch>?expand=1` (aucune commande d'écriture). Sinon 502 `{error}`.
- C'est une action externe : le bouton frontend applique le même double clic de confirmation que le ship.

## 5. i18n frontend + erreurs backend en anglais

- Backend : TOUS les messages d'erreur API passent en anglais, courts (ex : "confirmation required", "task not found", "invalid project path: not a git repository"). Les logs/commentaires restent comme ils sont.
- Frontend : dictionnaire `I18N = {fr: {...}, en: {...}}` dans app.js, fonction `t(key, vars?)`. TOUTES les chaînes visibles passent par t() (y compris dates relatives, pluriels simples via clés distinctes, placeholders, tooltips, modales, états vides).
- Langue initiale : localStorage `sillage.lang`, sinon `navigator.language` (fr* -> fr, sinon en).
- Bascule : petit sélecteur discret "FR · EN" dans le pied de la sidebar (près des tokens), re-render immédiat, persisté.
- Les noms de colonnes kanban, statuts, filtres, workflow, boutons d'action, login, paramètres : tout est traduit. Les contenus utilisateur (titres, messages, prompts) ne sont jamais traduits.
- Seeds des nouveaux installs (agents par défaut) : contextPrompt en anglais.

## 6. UI nouvelles surfaces (rester zen, une action primaire par écran)

- Sidebar, section Agents : un "+" discret dans l'en-tête de section -> modale "Nouvel agent". Clic sur un agent -> modale d'édition (mêmes champs : nom, emoji, couleur, cli en select, modèle, prompt de contexte) avec bouton principal "Enregistrer" et lien discret "Supprimer" (confirmation inline, erreur si refus serveur).
- Pied de sidebar : icône ⚙ (admin seulement) -> modale "Utilisateurs" : liste (nom, rôle), ajout (nom + mot de passe + rôle), réinitialisation de mot de passe, suppression (avec les gardes serveur). Sobre, une colonne.
- En-tête kanban : petite icône ✎ à côté du nom du projet -> modale "Projet" (nom, commande de check avec placeholder `go test ./...`).
- Détail de tâche, statut shipped : sous l'action principale "Rouvrir", un bouton secondaire "Ouvrir la PR" (double clic de confirmation, ouvre l'URL renvoyée dans un nouvel onglet). Également remplacé dans la barre du diff (l'ancien bouton désactivé).
- Login : champs identifiant + mot de passe (placeholder "admin" pour l'identifiant au premier lancement).

## 7. Tests

- Backend : tests unitaires pour la migration users, les gardes (dernier admin, agent référencé), le parse d'URL github (ssh + https) du fallback PR.
- e2e existants à adapter (login username+password).
- `go vet`, `gofmt`, `go test`, `node --check` verts.

## Hors périmètre v0.2 (ne pas faire)

Permissions par projet, non-lus par utilisateur, i18n backend, suppression de projets, OAuth, HTTPS intégré.
