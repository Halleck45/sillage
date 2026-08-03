# Sillage : chantiers (ex-cartes) et contexte de chantier

Décision produit : l'objet entre le projet et la tâche s'appelle désormais un **chantier** (anglais : **workstream**). Le modèle et l'API gardent `Card`/`cards` (pas de renommage technique, pas de migration) : seul le vocabulaire visible change.

## 1. Contexte de chantier (backend)

- `Card` gagne `contextPrompt` (texte libre, persisté, exposé).
- POST `/api/cards` : accepte `contextPrompt` optionnel. PATCH `/api/cards/{id}` : accepte désormais `{column?, title?, contextPrompt?}` (title non vide si fourni).
- Runner : le contexte transmis aux agents devient la concaténation, chaque bloc seulement s'il est non vide :
  - claude (--append-system-prompt) : contexte agent, puis `Project context:\n...`, puis `Workstream context:\n...` (celui de la carte de la tâche), séparés par des lignes vides.
  - codex : préfixe du prompt : `Project context:\n...\n\nWorkstream context:\n...\n\n---\n\n`.
  - fake : rien.
- Tests : PATCH title/contextPrompt, et construction du prompt combiné (fonction extraite testable).

## 2. Vocabulaire et UI (frontend)

- i18n : toutes les chaînes visibles qui disaient « carte »/« card » passent à « chantier »/« workstream » : « Nouveau chantier »/« New workstream », « Titre du chantier »/« Workstream title », états vides, tooltips, compteurs éventuels. Les colonnes du kanban ne changent pas.
- Modale nouveau chantier : ajouter la textarea « Contexte pour les agents » (mêmes codes visuels que celle du projet).
- Édition d'un chantier : petite icône ✎ à côté du titre du chantier dans l'en-tête de la vue travail (liste de tâches) : modale avec titre + contexte. PATCH /api/cards/{id}.
- Modale nouvelle tâche : sous « + contexte du projet », afficher aussi « + contexte du chantier » si défini.
- i18n fr+en pour tout, aucun texte en dur.
