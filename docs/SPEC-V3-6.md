# Sillage : suppression de tâches, chantiers et projets

Actions destructives : jamais dans les listes ; confirmation double clic côté UI ET `{"confirm":true}` côté API (400 "confirmation required" sinon). Erreurs en anglais, i18n fr+en.

## Backend

- DELETE `/api/tasks/{id}` body `{confirm:true}` : si running, interrompre le process d'abord (même mécanique que cancel) ; supprime la tâche et ses messages ; retire le worktree (`git worktree remove --force` puis `worktree prune` dans le repo d'origine, best-effort : un échec de nettoyage n'empêche pas la suppression) ; ne supprime JAMAIS la branche (elle peut être poussée). SSE : `cards` (compteurs), `agents`, `tokens`, et un event `taskDeleted` `{taskId, cardId, projectId}`.
- DELETE `/api/cards/{id}` body `{confirm:true}` : cascade sur ses tâches (même traitement chacune). SSE `cards` + `taskDeleted` par tâche + event `cardDeleted` `{cardId, projectId}`.
- DELETE `/api/projects/{id}` body `{confirm:true}` : cascade chantiers + tâches. SSE `projectDeleted` `{projectId}` (le front recharge l'état à cet event, plus simple et sûr).
- Les tokens déjà consommés par les tâches supprimées disparaissent des agrégats (comportement assumé : les compteurs reflètent l'existant). Publier `tokens` après suppression.
- Tests : suppression simple, cascade chantier, cascade projet, tâche running interrompue puis supprimée, worktree effectivement retiré (repo git réel en TempDir), refus sans confirm.

## Frontend

- Détail de tâche : lien discret rouge « Supprimer la tâche » sous les actions secondaires (double clic de confirmation). Après suppression : retour à la liste (hash de la carte), la tâche disparaît.
- Modale ✎ du chantier : lien discret « Supprimer le chantier » avec sous-texte « Supprime aussi ses N tâches » (double clic). Après : retour au kanban.
- Modale ✎ du projet : lien discret « Supprimer le projet » avec sous-texte « Supprime aussi N chantiers et N tâches » (double clic). Après : retour à Tous les projets.
- Écoute des events SSE `taskDeleted`/`cardDeleted`/`projectDeleted` : purge de l'état local (ou re-fetch pour projectDeleted) et sortie propre si l'objet affiché a disparu.
- Réutiliser le mécanisme de confirmation générique existant. i18n fr+en.
