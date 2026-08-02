# Sillage : fournée UX (post-v0.3)

Retours du propriétaire, à implémenter après la v0.3. Erreurs API en anglais, i18n fr+en pour toute nouvelle chaîne, simplicité d'abord.

## 1. Cartes : création uniquement dans « Bientôt »

- Frontend : le bouton « + Ajouter » ne s'affiche que dans la colonne soon ; la modale nouvelle carte n'a plus de choix de colonne.
- Backend : POST /api/cards ignore/refuse column != "soon" (400 `"cards are created in the soon column"` si fourni et différent).
- Le déplacement manuel de cartes entre colonnes (menu ⋯) reste inchangé.

## 2. Routage par URL (hash)

- Hash routing dans app.js : `#/inbox`, `#/projects`, `#/p/{projectId}`, `#/p/{projectId}/c/{cardId}`, `#/p/{projectId}/c/{cardId}/t/{taskId}` (+ `?tab=diff|deliverables` optionnel pour l'onglet).
- Navigation interne = pushState sur le hash ; popstate/hashchange restaure la vue ; au chargement initial, l'URL est appliquée après le fetch de /api/state (id inconnu -> retour à `#/projects` sans erreur).
- Aucune modif serveur.

## 3. En-tête projet allégé

- Supprimer du header kanban les compteurs « N en cours · N à relire · N tâches ». Garder uniquement les tokens du projet (`Σ ... · ...$`).

## 4. Cycle de vie des tâches : Terminer et Annuler

- Nouveaux statuts Task : `done` et `cancelled` (en plus de running/review/ready/shipped).
- Backend :
  - POST /api/tasks/{id}/finish : autorisé depuis review/ready/shipped -> done. 400 sinon (`"task must be reviewed before finishing"` pour running).
  - POST /api/tasks/{id}/cancel : autorisé depuis running/review/ready -> cancelled (si running : interrompre l'agent d'abord, même mécanique que interrupt). 400 depuis shipped/done.
  - reopen : accepté depuis shipped, done ET cancelled -> review.
  - Compteurs : tasksDone = shipped+done ; les cancelled sont EXCLUES de tasksTotal et de la progression ; reviewCount inchangé.
  - **Auto-déplacement de carte** : après chaque changement de statut de tâche, si la carte a au moins une tâche et que toutes ses tâches sont shipped/done/cancelled -> card.column = "done" automatiquement (SSE cards). Si une tâche redevient active (reopen/nouvelle tâche) et que la carte est en done -> retour en "doing".
- Frontend :
  - Icônes liste : done = ✓ vert (comme shipped), cancelled = ⊘ gris avec titre barré discret.
  - Détail : sous l'action primaire, deux liens discrets contextuels : « Marquer comme terminé » (review/ready/shipped) et « Annuler la tâche » (running/review/ready, confirmation double clic). Pour done/cancelled, l'action primaire devient « Rouvrir la tâche ».
  - Pilules de filtre : Toutes / À relire / Prêt à livrer / Terminées (shipped+done+cancelled).
  - Workflow 3 segments : done affiche le segment « Livré » renommé... NON : garder les segments actuels ; pour done/cancelled afficher à la place un bandeau simple (« Tâche terminée » / « Tâche annulée »).

## 5. Contexte et description de projet

- Project gagne `description` (une phrase, affichée sous le nom dans l'en-tête kanban, gris discret) et `contextPrompt` (texte libre multi-lignes).
- POST/PATCH /api/projects acceptent ces champs ; modale projet : champ description + textarea « Contexte pour les agents » (placeholder explicite).
- **Câblage agents dès maintenant** (runner.go) : si project.contextPrompt non vide :
  - claude : l'ajouter au --append-system-prompt : contexte agent, puis ligne vide, puis `Project context:\n<contextPrompt>`.
  - codex : préfixer le prompt utilisateur de `Project context:\n<contextPrompt>\n\n---\n\n`.
  - fake : rien.
- La modale nouvelle tâche affiche, sous l'aperçu du contexte de l'agent, une ligne discrète « + contexte du projet » si défini.

## 6. Réglages globaux : prénom et langue

- Nouveau modèle persisté dans state.json : `Settings { displayName: string, lang: "fr"|"en"|"" }` ; exposé dans GET /api/state (`settings`) ; PATCH `/api/settings` `{displayName?, lang?}` (lang validée, 400 sinon) ; SSE event `settings`.
- `displayName` : utilisé comme `authorName` des messages utilisateur à leur création (champ déjà existant, vide jusqu'ici). Frontend : affiche authorName s'il est non vide, sinon « Vous »/« You » comme aujourd'hui.
- `lang` : si non vide, prime sur localStorage/navigator au boot ; la bascule de langue écrit désormais via PATCH /api/settings ET localStorage (ce dernier sert avant login).
- UI : la modale ⚙ « Espace de travail » gagne une petite section « Préférences » en haut : champ Prénom + sélecteur de langue (les mêmes FR · EN). Le sélecteur du pied de sidebar est retiré au profit de la modale.

## Versions

Ne plus tagger à chaque fournée : commits sur main uniquement, release quand le propriétaire le demandera.
