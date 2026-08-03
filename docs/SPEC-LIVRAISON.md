# Spec : livraison par chantier (branche de feature, accepter, Ship)

Spec de fournée : le « pourquoi » du modèle de livraison. Les lots 1 et 2 sont implémentés
(voir §8) ; `docs/SPEC-API.md` et `docs/SPEC-BACKEND.md` sont le contrat courant et font foi
sur les détails. Les écarts entre l'intention initiale et ce qui a été construit sont notés
dans le texte, pour que la décision reste traçable.

## 1. Le problème

Aujourd'hui la livraison est une action de **tâche** : `POST /api/tasks/{id}/ship` (push de la
branche de la tâche), puis `POST /api/tasks/{id}/pr` (ouverture de la PR), chacune avec sa
confirmation, la seconde dans le pied du panneau diff et grisée jusqu'à ce que la première ait
réussi. Quatre clics, dont deux à aller chercher.

La cause n'est pas le nombre de clics, c'est le modèle : la hiérarchie produit
(projet > chantier > tâches) n'existe pas dans git. Chaque tâche crée sa branche
`sillage/<ref>-<slug>` depuis la branche courante du dépôt. Trois tâches d'un même chantier sont
trois branches sœurs de `main`, sans lien : trois pushs, trois PRs, aucune vue d'ensemble, et les
conflits entre tâches d'un même chantier se découvrent au merge final. On livre des tâches alors
que l'unité de valeur relue est le chantier.

## 2. Le modèle cible

**Le chantier est une branche de feature. La tâche est un incrément local qu'on accepte ou refuse.**

- Une branche de chantier par couple **(chantier, dépôt)** : `sillage/ws-<cardRef>-<slug>`,
  créée à la première tâche du chantier sur ce dépôt, depuis la branche cible du projet
  (voir §4). Un chantier qui touche deux dépôts a deux branches, et livrera deux PRs.
- Chaque branche de chantier a son **worktree dédié** (`worktrees/ws-<cardId>-<repoName>/`).
  Aucune opération n'a jamais lieu dans le dépôt de travail de l'utilisateur.
- Les branches de tâche partent désormais de la branche du chantier, pas de `main`. Effet de
  bord voulu : une tâche créée après une acceptation démarre **sur le travail déjà accepté**,
  les agents s'enchaînent au lieu de se marcher dessus.
- Les branches de tâche ne sont plus jamais poussées. Elles sont locales, fusionnées dans la
  branche du chantier à l'acceptation.
- **Une seule action sortante dans tout le produit** : le Ship du chantier.

Modèle mental : *je relis une tâche, je l'accepte ; je relis le chantier, je le livre.*

## 3. Cycle de vie d'une tâche

Statuts : `running` → `review` → `accepted`, plus `cancelled`.

- **Accepter** (`POST /api/tasks/{id}/accept`, depuis `review`) : dans le worktree de la tâche,
  `git add -A` puis commit `Sillage: <titre>` si l'arbre est sale ; puis, dans le worktree du
  chantier, `git merge --no-ff <branche de tâche>` avec pour sujet `Sillage: <titre>`. Aucun
  réseau, aucune confirmation (l'action est locale et réversible par `/reopen`).
  Le `--no-ff` est délibéré : un merge commit par tâche acceptée rend lisible, dans l'historique
  du chantier, ce qui a été accepté et par quelle tâche.
- **Refuser** : c'est `POST /api/tasks/{id}/cancel`, déjà existant (statut `cancelled`). Pas de
  nouvel endpoint : refuser et abandonner sont la même chose côté git (la branche reste, rien
  n'est fusionné, rien n'est supprimé). Seul le libellé UI change selon le contexte.
- **Rouvrir** (`/reopen`, depuis `accepted`/`cancelled`) : retour en `review`. Le merge déjà
  fait n'est **pas** annulé ; la prochaine acceptation fusionne les nouveaux commits. Si le
  chantier était déjà livré, le Ship suivant repousse la branche et met la PR à jour.
- **Acceptation constatée** (*ajouté à l'implémentation*) : à l'ouverture de la vue chantier, une
  tâche en revue dont la branche est déjà entièrement contenue dans celle du chantier passe
  `accepted` d'elle-même, avec un marqueur `[auto-accepted:<branche>]`. Trois garde-fous : aucun
  agent en cours, worktree de la tâche propre, branche effectivement contenue. Aucune écriture
  git : on ne fait que constater ce qui est déjà fusionné, pour ne pas demander à l'humain
  d'accepter une deuxième fois le même travail.
- **Conflit à l'acceptation** : `git merge --abort`, la tâche **reste** en `review`, réponse
  `409 {"error":"merge conflict with the workstream branch"}` avec la liste des fichiers en
  conflit. Un message marqueur `[merge-conflict:<fichiers>]` est ajouté au fil, le frontend
  affiche une ligne système localisée invitant à demander à l'agent de reprendre la base. La
  résolution assistée (rebase de la tâche sur le chantier en un clic) est du lot 3.

Disparaissent : `POST /api/tasks/{id}/ship`, `POST /api/tasks/{id}/pr`,
`POST /api/tasks/{id}/finish`, et les statuts `shipped` et `done`.

Migration au chargement (`loadStoreFile`, comme `migrateLegacyRepos`) : `shipped` et `done`
deviennent `accepted`. Les chantiers existants ne sont **pas** rétro-convertis en branches de
feature : leurs tâches gardent leur base et leur branche ; le nouveau modèle s'applique aux
chantiers créés ensuite. Une tâche existante sans branche de chantier ne peut pas être acceptée
(l'UI l'indique et propose le seul chemin sain : rouvrir dans un nouveau chantier).

## 4. Réglage projet : ce que « livrer » veut dire

Nouveau champ persisté sur `Project` :

```jsonc
Delivery { "mode": "pr" | "push" | "merge" | "merge-push", "target": "", "stackedPrs": false }
          // target : branche de destination (base de la PR en mode "pr", branche fusionnée
          //          dans les modes "merge" et "merge-push"). "" = branche par défaut
          //          de chaque dépôt.
          // stackedPrs : lot 3, ignoré avant.
```

*Étendu après coup (lot 2 bis)* : les deux modes initiaux (`pr`, `merge`) forçaient deux choix à la
fois, « avec ou sans pull request » et « local ou remote ». Or ces deux questions sont
indépendantes et se répondent différemment d'un projet à l'autre. Les quatre modes sont donc le
produit des deux axes : livrer la branche du chantier (`pr` avec pull request, `push` sans) ou
faire avancer la branche de destination (`merge` en local, `merge-push` en poussant).

**Le réglage n'est jamais une question posée à froid.** À la création du projet il est
pré-rempli par détection du remote `origin` de chaque dépôt, et modifiable :

| Remote `origin` détecté | Défaut proposé |
|---|---|
| `github.com` | `mode:"pr"` |
| hôte contenant `gitlab` | `mode:"pr"` |
| autre hôte, ou aucun remote | `mode:"merge"`, `target` = branche courante du dépôt |

Le fournisseur n'est **pas** persisté : il est redéduit du remote à chaque opération (un remote
peut changer). Seuls `mode`, `target` et `stackedPrs` sont dans `state.json`.

### Mode `pr`

Push de la branche de chantier, puis ouverture de la PR/MR :

- GitHub : `gh pr create --head <branche> --base <target>`. Repli si `gh` est absent ou échoue :
  URL de comparaison en lecture seule (comportement actuel, conservé).
- GitLab : `glab mr create --source-branch <branche> --target-branch <target>`. Repli :
  `<origin>/-/merge_requests/new?merge_request[source_branch]=<branche>&merge_request[target_branch]=<target>`.

La branche est poussée dans les deux cas : le repli ne dégrade que l'ouverture de la PR, jamais
la livraison du code.

### Mode `push`

Le mode `pr` sans la PR : la branche de chantier est poussée, et c'est tout. Pour les projets où
la revue se passe ailleurs (ou pas du tout), et pour les forges que Sillage ne connaît pas. Ce
comportement existait déjà comme *repli* du mode `pr` sur forge inconnue ; en faire un mode
explicite évite de devoir mentir au réglage pour obtenir ce qu'on veut.

### Mode `merge`

Fusion locale de la branche de chantier dans `target`, dans un worktree transitoire dédié
(`worktrees/.merge-<cardId>-<repoName>/`, créé puis retiré dans l'opération). Trois règles
fermes, à écrire dans l'UI :

1. **Jamais de push.** Le mode `merge` fusionne et s'arrête. Pousser `main` reste une décision
   humaine, prise dans son terminal. Sans cette règle, l'invariant « `git push` n'existe qu'à
   deux endroits » resterait vrai dans le code mais vide dans les faits.
2. **Jamais de changement de branche dans le dépôt de travail de l'utilisateur.**
   *Précisé à l'implémentation* : la règle initiale (« jamais dans le dépôt de travail ») rendait
   le mode inutilisable dans le cas courant, puisque git refuse d'emprunter dans un worktree
   transitoire une branche déjà empruntée, et `main` est presque toujours la branche courante du
   dépôt de travail. Le repli retenu : fusionner **dans** ce dépôt, mais seulement si `target`
   **est** sa branche courante et que son arbre est propre, donc sans jamais changer de branche,
   sans stash et sans possibilité de perdre du travail. Sinon, refus avec la commande à jouer à
   la main (`ErrTargetBusy`).
3. **Fast-forward uniquement.** Si `target` a divergé, aucune tentative de résolution :
   `409 {"error":"target branch has diverged..."}` avec la commande à copier.

### Mode `merge-push`

Même fusion que `merge`, mais `target` est poussée ensuite : le projet dont livrer veut dire
« faire avancer la branche principale, remote comprise », sans passer par une pull request.

La règle 1 du mode `merge` (jamais de push) devient donc, pour ce mode, un **choix explicite du
projet** : c'est un réglage que l'humain est allé poser, pas un défaut. Les deux autres règles
tiennent telles quelles, et une quatrième s'ajoute :

4. **Rattraper le remote d'abord.** `git fetch origin <target>`, puis fast-forward de `target`
   locale sur `FETCH_HEAD` si le distant est en avance. Sans ce rattrapage, le push serait rejeté
   dès que `target` a bougé côté remote, et l'utilisateur se retrouverait avec une fusion locale
   à moitié livrée. Une divergence réelle (les deux côtés ont avancé) est un refus
   `ErrTargetDiverged` **avant** toute écriture. Jamais de `--force`, jamais de merge commit de
   réconciliation : c'est un rattrapage, pas une résolution.

Le push passe par la même primitive que celui de la branche de chantier (`pushBranch`), pour que
l'invariant « `git push` n'existe qu'à deux endroits » reste vérifiable par un grep.

### Santé de la livraison

Champ calculé à chaque lecture, **jamais persisté** (même pattern que `AgentOut.Warning`) :
`ProjectOut.deliveryWarning`, en anglais, vide si tout va bien.

- `gh not found in PATH; Sillage will fall back to a prefilled pull request URL`
- `glab not found in PATH; Sillage will fall back to a prefilled merge request URL`
- `no 'origin' remote on repository <name>; nothing can be pushed`
- lot 3 : `gh stacked pull requests plugin not installed; one pull request per workstream`

Ces avertissements s'affichent dans les réglages du projet, à côté du choix de mode, et dans le
récapitulatif de livraison. Ils n'empêchent jamais rien.

## 5. Le Ship du chantier

### Aperçu (lecture seule)

`GET /api/cards/{id}/delivery` :

```jsonc
{ "mode": "pr", "target": "main", "provider": "github",
  "ready": false, "blocker": "tasks-pending",
  "warnings": ["gh not found in PATH; ..."],
  "counts": { "accepted": 3, "refused": 1, "pending": 1 },
  "repos": [ { "repoName": "api", "branch": "sillage/ws-7-refonte-auth", "base": "main",
               "commits": 4, "files": 11, "pending": 4, "prUrl": "", "shippedAt": null } ] }
```

C'est cet appel qui alimente à la fois le sous-texte du bouton et le panneau de récapitulatif :
l'utilisateur sait **avant** de cliquer ce qui va se passer, et sur quels dépôts.

*Ajouté à l'implémentation* : `repos[].pending`, le nombre de commits qui restent **réellement**
à livrer (non poussés en mode `pr`, non encore fusionnés en mode `merge`), distinct de `commits`
qui décrit le contenu de la livraison. Sans lui, le bouton restait actif après une livraison
réussie et un second clic ne faisait rien, ce qui est exactement le genre de faux positif que
cette fournée cherche à supprimer. `pending` à zéro partout signifie « rien à livrer » : bouton
inactif, et dépôt marqué `skipped` si la livraison est tout de même appelée.

### Action

`POST /api/cards/{id}/ship {confirm:true}` (400 `"confirmation required"` sans le drapeau :
invariant inchangé). Pour chaque dépôt du chantier ayant au moins un commit : push de la branche
de chantier, puis PR/MR ou fusion locale selon `project.delivery.mode`.

Réponse : `{card, repos:[{repoName, branch, base, pushed, prUrl, output, error}]}`. Un dépôt en
échec n'annule pas les autres : chaque ligne porte son propre `error`, et l'UI affiche le détail
par dépôt. Une nouvelle livraison est toujours possible (elle ne fait qu'ajouter des commits à
une branche déjà poussée, donc un push ordinaire, jamais de force).

### Conditions (règle produit, appliquée aux deux bouts)

Le Ship est **possible** si et seulement si :

1. le chantier a au moins une tâche ;
2. aucune tâche n'est `running` ni `review` (tout est accepté ou refusé) ;
3. au moins une tâche est `accepted` ;
4. au moins un dépôt du chantier a des commits à livrer.

Champs dérivés sur `Card`, recalculés dans `recomputeCard` (verrou tenu) :
`shipReady: bool` et `shipBlocker: "" | "no-tasks" | "tasks-pending" | "nothing-accepted" | "nothing-to-ship"`.
Côté API, `POST /api/cards/{id}/ship` refuse en 409 avec le même vocabulaire ; côté UI, le
bouton est grisé avec le blocage en infobulle (« 1 tâche encore à relire »).

### Colonne du chantier

Changement de la règle actuelle : `card.column` passe à `"done"` **après une livraison
réussie**, plus quand toutes les tâches sont terminales. Toutes les tâches acceptées ou refusées
sans livraison, c'est `"doing"` avec le Ship actif. La colonne « Terminé » retrouve son sens :
livré. Le déplacement manuel (`PATCH /api/cards/{id}`) reste indépendant.

### Un chantier n'est pas figé par sa livraison

Un chantier livré doit pouvoir continuer à vivre : on relit, on livre, puis on ajoute une tâche
et on relivre. `CardBranch.shippedAt` ne veut donc pas dire « a été livré une fois » mais
**« livré, et rien de nouveau depuis »**. Il est remis à zéro dès qu'un travail nouveau apparaît
sur ce dépôt :

- création d'une tâche sur le chantier (une tâche de plus veut dire que tout n'est plus livré) ;
- acceptation qui ajoute réellement des commits à la branche du chantier (comparaison du SHA de
  HEAD avant et après la fusion, la sortie de git étant localisée selon la machine).

Conséquences : la carte quitte la colonne « Terminé », le bouton redevient actif une fois tout
accepté ou refusé, et la livraison suivante ne pousse que le neuf. `prUrl` est **conservée** : la
pull request existe toujours et le push la met à jour ; Sillage ne tente jamais d'en ouvrir une
seconde pour la même branche.

### Acceptation automatique : les merges faits à la main

Le chantier est une branche git ordinaire : rien n'empêche de fusionner soi-même une branche de
tâche dedans, depuis un terminal. Si Sillage l'ignorait, la tâche resterait « à relire » pour
toujours et bloquerait le bouton de livraison, ce qui punirait exactement le travail à la main
que le produit doit accepter.

`GET /api/cards/{id}/delivery` **constate** donc, avant de calculer l'aperçu : une tâche `review`
dont la branche est entièrement contenue dans celle du chantier passe `accepted`, avec le
marqueur `[auto-accepted:<branche>]` au fil. Le frontend appelle cet endpoint à l'ouverture de la
vue chantier, puis toutes les 60 secondes tant qu'elle reste ouverte.

Quatre garde-fous, parce qu'une décision automatique doit être conservatrice :

1. **aucun agent en cours** pour cette tâche (il travaille peut-être encore) ;
2. **la tâche a produit quelque chose** (`filesCount > 0`) : une tâche vide est contenue dans la
   branche du chantier par construction, la déclarer fusionnée ne constaterait rien ;
3. **worktree propre** : du travail non commité n'est par définition pas fusionné, et le déclarer
   accepté le perdrait de vue ;
4. **branche effectivement contenue** dans celle du chantier (`git merge-base --is-ancestor`).

Aucune écriture git : cette détection ne fait que lire, elle n'ajoute aucun commit et ne remet
donc jamais le chantier en « non livré ».

## 6. DX : ce que l'utilisateur voit

**Liste des tâches du chantier.** Chaque ligne porte son état, lisible sans clic :
`running` (activité live), `review` (point d'attention), `accepted` (✓ vert),
`cancelled` (grisé, libellé « Refusée »). Au **survol** d'une ligne en `review`, deux boutons
apparaissent à droite : **Accepter** et **Refuser**. Un clic, pas de confirmation, effet
immédiat (la ligne bascule en ✓ et le compteur d'en-tête se met à jour). Accepter depuis la
liste ne demande pas d'ouvrir la tâche : la relecture du diff reste disponible dans le détail
pour qui la veut, elle n'est plus un péage.

**En-tête du chantier.** Une ligne d'état et un bouton :

```
Refonte auth                                   [ Livrer ]
3 acceptées · 1 refusée · 1 à relire           Pousse sillage/ws-7-refonte-auth et ouvre une PR vers main (api)
```

- Le sous-texte annonce l'action exacte, tirée de `GET /api/cards/{id}/delivery`.
- Le bouton est grisé tant que `shipReady` est faux, avec le blocage en infobulle.
- Un clic ouvre un récapitulatif : mode, dépôts concernés, branche → base, nombre de commits et
  de fichiers, avertissements de santé. **Le bouton d'action de ce panneau est la confirmation.**
  Deux clics au total, le second en connaissance de cause. Plus de bouton « Confirmer ? »
  qui remplace un libellé sans rien montrer.
- Après livraison, la ligne devient « PR #123 ouverte » (cliquable), une par dépôt.

Le pied du panneau diff perd ses deux boutons (Push, PR) : la livraison n'est plus une action de
tâche. Il ne reste que la lecture du diff.

**Pourquoi une étape d'acceptation revient alors qu'on l'avait retirée (v0.3.4).** L'ancien
`ready`/`/accept` était un péage sans contenu : un état de plus avant un ship qui, de toute
façon, ne concernait que cette tâche. Ici, accepter **fait** quelque chose (la fusion dans la
branche du chantier) et sert de porte au seul bouton sortant du produit. C'est la même
mécanique, ce n'est pas le même prix.

## 7. Ce qui n'est pas dans ce lot

- **PRs empilées GitHub** (`stackedPrs`) : reporté au lot 3, volontairement. Une fois le
  chantier devenu une branche de feature, le besoin couvert par le stacking (des PRs enchaînées
  dépendantes) est absorbé. Le stacking ne redevient utile que pour un mode de revue plus fin
  (une PR par tâche à l'intérieur du chantier), qui est une décision produit distincte. Le faire
  maintenant ferait passer la livraison de deux chemins de code (gh, URL de repli) à cinq
  (stack, gh, URL GitHub, glab, URL GitLab) sur un flux dont la DX n'est pas encore bonne.
- **Rebase assisté** d'une tâche en conflit sur la branche du chantier.
- **Diff agrégé du chantier** (`GET /api/cards/{id}/diff`) : les compteurs de l'aperçu suffisent
  d'abord ; la vue complète viendra si la relecture au niveau chantier se révèle nécessaire.

## 8. Découpage

1. **Chantier = branche** (fait). Branche et worktree par (chantier, dépôt), tâches basées sur la
   branche du chantier, `/accept`, refus via `/cancel`, `shipReady`/`shipBlocker`, Ship de
   chantier, retrait des routes de tâche, migration des statuts. C'est là qu'est l'essentiel du gain.
2. **Réglage de livraison** (fait). `Project.delivery`, détection à la création, mode `merge`,
   GitLab, `deliveryWarning`.
3. **PRs empilées** (option), rebase assisté, diff de chantier : à faire, si le besoin se confirme
   à l'usage.

## 9. Impacts sur les specs courantes (faits pour les lots 1 et 2)

- `SPEC-API.md` : modèles `Project` (+`delivery`, +`deliveryWarning`), `Card` (+`branches`,
  +`shipReady`, +`shipBlocker`), `Task` (statuts) ; table des endpoints (ajout de `/accept`,
  `/api/cards/{id}/ship`, `/api/cards/{id}/delivery` ; retrait de `/ship`, `/pr`, `/finish` de
  tâche) ; section « Cycle de vie des tâches et auto-déplacement de carte » (statuts, règle de
  colonne) ; section « Ouvrir la PR » remplacée par « Livraison d'un chantier » ; règles UI
  (le point « push uniquement via le bouton du détail de tâche » devient « uniquement via le
  bouton de l'en-tête de chantier »).
- `SPEC-BACKEND.md` : worktrees de chantier, merge d'acceptation, adaptateurs `gh`/`glab`,
  worktree transitoire du mode `merge`.
- `CONTRIBUTING.md` : l'invariant `git push` reste vrai (`Ship()` opère désormais sur une
  branche de chantier, `SyncPush()` inchangé) ; préciser que le mode `merge` ne pousse jamais.

## 10. Questions tranchées à l'implémentation

- `Card.ref` : le compteur de référence est bien partagé entre tâches et chantiers (une seule
  suite de références courtes dans un projet, `sillage/ws-<ref>-<slug>`).
- Chantier multi-dépôts et livraison partielle : le Ship est rejouable tel quel. Chaque dépôt
  porte son propre résultat et sa propre erreur ; ceux qui n'ont plus rien à livrer sont
  `skipped`. Pas de bouton « réessayer ce dépôt » à maintenir.
