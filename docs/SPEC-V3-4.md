# Sillage : simplification du workflow de livraison + confort de conversation

Retours du propriétaire. Erreurs API en anglais, i18n fr+en.

## 1. Suppression de l'état « ready » (accepted)

Le cycle devient : running → review → shipped (plus done/cancelled inchangés).

- Backend : POST /accept supprimé (la route disparaît). POST /ship accepté depuis `review` (confirm toujours obligatoire, inchangé). finish : autorisé depuis review/shipped (inchangé sauf disparition de ready). Migration au chargement : toute tâche `ready` devient `review`.
- Après un ship réussi : ajouter un Message marqueur `"[shipped:<branch>]"` (author=agent, authorName vide) ; si le remote origin du repo est github, le backend fournit aussi l'URL de branche dans la réponse ship `{task, output, branchUrl}` (`https://github.com/<o>/<r>/tree/<branch>`, vide sinon).
- Frontend : le bouton « Accepter » disparaît ; depuis review l'action primaire est « Livrer » (vert, double clic de confirmation). La barre de workflow 3 segments disparaît au profit d'un badge de statut sobre dans l'en-tête du détail (icône + libellé : ◐ En cours, ◍ À relire, ✓ Livré, ✓ Terminé, ⊘ Annulé). Pilules de filtre : Toutes / À relire / Terminées. Le marqueur `[shipped:...]` s'affiche comme ligne système « Livré : branche <branch> poussée » avec lien externe vers branchUrl si connue (stockée côté client depuis la réponse du ship ; sinon ligne sans lien).
- Nettoyer les clés i18n mortes (accepted, ready...).

## 2. Conversation : rendu incrémental (copier-coller cassé)

Problème : la conversation est reconstruite à chaque événement SSE (task, activity...), la sélection de texte saute.

- Le fil de messages ne doit JAMAIS être reconstruit intégralement quand seul du contenu s'y ajoute : événement `message` -> append d'un nœud DOM ; événement `activity`/`tokens`/`task` -> mise à jour ciblée des seuls éléments concernés (ligne live, compteurs, bouton d'action, badge de statut) sans toucher au conteneur des messages.
- Si un re-render complet est inévitable (changement d'onglet, navigation), préserver le scroll. Pendant qu'une sélection de texte est active dans le fil (window.getSelection non vide à l'intérieur du conteneur), différer toute mutation non essentielle du fil.
- Vérifier le même confort sur l'onglet Diff.
