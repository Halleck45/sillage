# Sillage : stabilité de la liste de tâches + liens épinglés par projet

## 1. La liste de tâches ne doit pas se réordonner au clic

Bug : ouvrir une tâche déclenche POST /read, qui met à jour updatedAt, ce qui re-trie la liste sous le curseur.

- Backend : marquer une tâche lue (handleRead / la mutation Unread=false) ne doit PAS toucher UpdatedAt. Introduire un moyen explicite (ex : variante d'UpdateTask sans bump, utilisée uniquement par read). Test : read ne change pas UpdatedAt.
- Frontend : vérifier que le tri de la liste est stable et déterministe (updatedAt desc puis ref desc en départage) et que l'ouverture d'une tâche ne provoque aucun re-tri visible.

## 2. Liens épinglés par projet

Besoin : épingler des URLs par projet (site, dépôts, dashboards) sous forme d'une rangée discrète de favicons en haut de la vue projet. Un lien = une URL ; au survol : le titre de la page ; au clic : ouverture dans un nouvel onglet.

- Modèle : `Project.links: [{url, title}]` (persisté, exposé).
- API : POST/PATCH /api/projects acceptent `links` (liste remplaçante). Validation : http(s) uniquement (400 sinon), max 12 liens. Pour chaque lien SANS titre fourni, le serveur tente de récupérer le `<title>` de la page à l'enregistrement : GET avec timeout 5 s, taille lue plafonnée à 64 Ko, http/https seulement, jamais de redirection vers autre chose que http(s) ; en échec, titre = nom d'hôte. Cette récupération est best-effort et ne bloque jamais l'enregistrement (échec silencieux -> hostname).
- Frontend :
  - En-tête kanban : rangée discrète de favicons (16-18 px, coins légèrement arrondis, opacité 0.85, hover 1) entre le titre/description et les colonnes, ou à droite du titre si plus harmonieux. favicon = `https://<host>/favicon.ico` chargée directement (PAS de service tiers) avec fallback sur un glyphe 🔗 gris si l'image ne charge pas (onerror). title HTML = titre stocké. target=_blank rel="noopener".
  - Gestion dans la modale projet (✎) : section « Liens épinglés » : lignes url (+ titre affiché en petit une fois connu), ajout par champ + bouton discret, suppression par ✕ de ligne. Envoyées via `links` au PATCH.
  - i18n fr+en. Aucun lien : rien ne s'affiche dans l'en-tête (pas d'état vide).
