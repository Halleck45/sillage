# Spec : recette manuelle (lancer un chantier ou une tâche pour s'en servir)

Spec de fournée (issue #128). **Le lot 1 (§6) est implémenté** ; `docs/SPEC-API.md`
§« Recette manuelle » est le contrat courant et fait foi sur les détails. Ce document porte le
« pourquoi » et la trace des pistes écartées (§5). Vocabulaire : « recette » côté UI et docs,
`preview` côté code (`Repo.PreviewCmd`, `PreviewRun`), parce que le produit ne prononce pas de
verdict : il amène l'humain devant le logiciel en marche.

## 1. Le problème

Sillage sait faire relire du **code** (diff, acceptation, livraison). Il ne sait pas faire éprouver
un **comportement**. Or ce qui décide vraiment d'une livraison, ce n'est pas « le diff est-il
correct ? » mais « est-ce que ça marche quand je m'en sers ? ». Aujourd'hui la seule réponse est de
sortir du produit : ouvrir un terminal, retrouver le chemin du worktree, se souvenir de la commande
de lancement de ce projet, éviter le port déjà pris par un autre chantier.

Deux pièges, écartés en §5 : croire que recetter c'est « lancer un serveur web » (ça exclut les CLI,
les jobs, les bibliothèques, le mobile), et croire qu'un agent peut recetter à notre place (il peut
préparer le terrain, jamais décider si on est d'accord).

## 2. Le modèle

**Un dépôt de projet porte une commande de recette. Sillage la lance dans le worktree du chantier
ou de la tâche, streame sa sortie, et affiche un lien et un bouton Arrêter.**

Sillage ne sait rien des stacks : il exécute une commande écrite par l'humain. Deux variables
injectées suffisent à rendre l'isolation possible sans que le produit apprenne ce qu'est une base
de données ou un conteneur.

### Déclaration : deux champs par dépôt

Sur `Repo`, à côté du chemin, dans la modale de réglages du projet :

- `previewCmd` : la commande, lancée via `sh -c` dans le worktree. Vide = pas de recette pour ce
  dépôt.
- `previewUrl` : optionnel, affiché en lien cliquable quand le process tourne
  (ex. `http://127.0.0.1:$((4000 + SILLAGE_N))`). Les variables y sont substituées comme dans la
  commande.

La commande est portée par le **dépôt** et non par le projet parce que les dépôts sont déjà une
liste dans les réglages : un projet à trois dépôts obtient trois commandes sans aucune UI de liste
nouvelle, et ça correspond exactement à `CardBranch` (une branche de chantier par dépôt).

### Les variables injectées

| Variable | Valeur | Pour |
| --- | --- | --- |
| `SILLAGE_ID` | `ws-107` (chantier) ou `t-482` (tâche) | noms : base de données, conteneur, répertoire |
| `SILLAGE_N` | `107` ou `482` | arithmétique : `PORT=$((4000 + SILLAGE_N))` |
| `SILLAGE_DIR` | le worktree (aussi le répertoire courant) | chemins absolus dans la commande |
| `SILLAGE_BRANCH` | la branche recettée | affichage, bannière de debug |

`SILLAGE_ID` et `SILLAGE_N` dérivent de `Card.Ref` et `Task.Ref`, qui existent déjà, sont petits,
sont stables dans le temps, et **partagent un seul compteur par projet** (tranché dans
`SPEC-LIVRAISON.md` §10). Conséquences, toutes gratuites :

- aucun état nouveau à persister : pas d'allocateur, pas de table de slots, pas de compteur ;
- aucune collision entre la recette d'un chantier et celle d'une de ses tâches ;
- **stabilité** : même valeur à chaque lancement, donc la base de recette et son contenu survivent
  entre deux sessions. À charge du projet d'écrire une commande idempotente (créer-si-absent) ;
  c'est son métier, pas celui de Sillage.

Deux projets différents peuvent partager la valeur `107`. Ce n'est pas un problème réel : la
commande est écrite par dépôt et nomme son propre applicatif (`app_$SILLAGE_ID`, pas `db_$SILLAGE_ID`).

### Exemples

```sh
# app web
npm ci && npm run dev -- --port $((4000 + SILLAGE_N))
# URL : http://127.0.0.1:$((4000 + SILLAGE_N))

# API avec base isolée, idempotente
make db-ensure DB=api_$SILLAGE_ID && make serve DB=api_$SILLAGE_ID PORT=$((4000 + SILLAGE_N))

# compose
COMPOSE_PROJECT_NAME=app-$SILLAGE_ID PORT=$((4000 + SILLAGE_N)) docker compose up

# CLI : on veut juste voir la sortie
go build -o /tmp/$SILLAGE_ID . && /tmp/$SILLAGE_ID --demo

# pipeline sur fixture
python -m etl --input fixtures/small.csv --out /tmp/$SILLAGE_ID
```

**Il n'y a pas de genre de recette.** Un serveur qui reste en vie et un script qui finit, c'est le
même code : Sillage lance, streame, et affiche soit « en cours » avec Arrêter, soit
« terminé (code 0) ». Web et non-web passent par le même chemin, sans champ de type à choisir.

## 3. Le run

- Un run appartient à un **worktree** : celui du chantier (cas normal, on éprouve l'état intégré)
  ou celui d'une tâche (on éprouve un incrément avant de l'accepter). Même mécanique, deux points
  d'entrée.
- Un seul run par worktree à la fois : relancer arrête le précédent.
- Statuts : `running` → `exited` (avec `exitCode`) | `stopped` (arrêt humain) | `failed` (n'a pas
  démarré).
- Supervision identique au runner d'agents : `Setpgid`, `Interrupt` envoie SIGINT au groupe puis
  SIGKILL après 5 s. `main.go` gagne un arrêt propre (SIGINT/SIGTERM) qui tue tous les runs :
  rien ne doit survivre à la fermeture de Sillage. En cas de SIGKILL du serveur, un process peut
  rester orphelin ; on l'accepte (le port occupé le signalera), on ne construit pas de registre
  sur disque pour ça.
- **Pas de plafond ni de TTL d'inactivité** dans ce lot. À la place, la visibilité : un compteur
  « Recettes en cours (2) » en bas de sidebar, qui ouvre la liste avec liens et boutons Arrêter.
  Un utilisateur qui voit ce qui tourne l'arrête lui-même ; un TTL qui coupe un serveur pendant
  qu'on s'en sert est plus agaçant qu'utile.
- **Rien n'est persisté** : les runs et leur journal (tampon circulaire, 2000 dernières lignes)
  vivent en mémoire. `state.json` est versionné dans l'espace de travail : y verser des logs le
  ferait gonfler pour rien.

## 4. Ce que l'utilisateur voit

En-tête du chantier, à gauche de Livrer, un bouton **Recette** qui ouvre un tiroir :

```
Recette : Refonte auth
  web    ● en cours      http://127.0.0.1:4107          [ Arrêter ]
  api    ○ arrêtée                                      [ Lancer ]
  worker  pas de commande de recette (Réglages du projet)

  [journal en direct du run sélectionné]
```

- Une ligne par dépôt du chantier (par `CardBranch`). Le journal s'affiche en direct, y compris
  pendant l'installation : c'est là que se voient les erreurs de dépendances, qui sont la moitié
  du sujet.
- Le lien n'est cliquable que si `previewUrl` est renseignée. Pas de sonde de disponibilité : si
  ça répond 404 la première seconde, on rafraîchit.
- Une ligne sans commande affiche le **chemin du worktree avec un bouton de copie** et un lien
  vers les réglages. Ce repli est le minimum garanti : la recette reste possible sur 100 % des
  projets sans aucune configuration.
- Détail d'une tâche : le même tiroir, sur le worktree de la tâche, un seul dépôt
  (`Task.RepoName`), avec un bandeau « vous recettez la tâche #482, pas le chantier ».
- Un chantier livré garde son tiroir : recetter après livraison reste légitime.
- Chaînes visibles via `t()`, clés `fr` et `en`, comme le reste.

## 5. Décisions écartées, et pourquoi

- **Un bouton « Lancer le serveur » avec détection de stack** (Sillage devine npm, Makefile,
  compose). Séduisant en démo, faux en pratique : chaque devinette rate, et le code de devinette
  grossit sans fin. Une ligne de config obtient le même résultat pour tous les types de projets.
  La détection peut revenir comme *suggestion préremplie* à la création d'un projet, jamais comme
  mécanique d'exécution.
- **Un champ dérivé rempli par l'IA** (« l'humain écrit l'intention, un agent en déduit la
  commande finale isolée »). Écarté : personne ne possède ce champ. Le jour où ça casse, on
  debugue une commande shell qu'on n'a pas écrite, dans une modale de réglages ; et elle est
  périmée par construction, calculée un jour contre l'état du dépôt de ce jour-là. S'y ajoute un
  appel LLM dans une validation de formulaire (latence, CLI absent, hors-ligne, échec à gérer).
  Cette intelligence existe déjà **mieux placée** : une tâche, où l'agent peut *essayer* sa
  commande au lieu de la deviner, et dont l'humain relit le résultat avant de le coller dans le
  champ. Un bouton « demander à un agent » qui pré-remplirait cette tâche est du sucre, pas de
  l'architecture (lot 2).
- **Des genres de recette** (`service` / `run` / `manual`) : complication inutile, voir §2.
- **Allocateur de port et sonde de disponibilité** : `SILLAGE_N` suffit à dériver un port stable,
  et l'attente de disponibilité se remplace par un rafraîchissement. Le jour où deux chantiers
  du même projet se battent pour un port, on saura pourquoi ajouter un allocateur.
- **`.sillage/recipes.json` dans le dépôt** : bien plus pratique (la recette suit le code), mais ce
  fichier est écrit par des agents : une branche pourrait faire exécuter une commande arbitraire
  sur le portable, déclenchée par un clic de l'humain qui croit lancer son app. Il faudrait une
  approbation par empreinte, donc une UI et un état de plus. La commande dans les réglages
  supprime la question entière : elle est tapée par l'humain, au même niveau de confiance que
  `checkCmd`, qui existe déjà.
- **Sign-off et blocage de livraison** (« recette OK » attaché au sha, blocage `signoff-missing`).
  Formaliser un verdict quand une seule personne recette et livre, c'est le péage vide de l'ancien
  `ready` analysé dans `SPEC-LIVRAISON.md` §6. Ne devient utile que si quelqu'un d'autre recette,
  ou si on se fait avoir en livrant sans essayer.
- **Un reverse-proxy Sillage** pour une URL stable : cookies, WebSockets, chemins absolus et
  redirections en font un chantier à part, et ça dépasse « localhost par défaut » (invariant 3).

Ligne de scope qui tient : **Sillage amène l'humain à la recette, il ne gère pas son environnement
de développement.** Il lance une commande dans un worktree ; il n'orchestre ni bases, ni
conteneurs, ni cycles de vie longs.

## 6. Découpage

1. **Le socle** (fait). `Repo.previewCmd` / `Repo.previewUrl`, substitution des quatre
   variables, lancement et arrêt depuis le chantier et depuis une tâche, journal en direct,
   compteur en sidebar, repli « chemin du worktree ». Réutilise la supervision de process de
   `runner.go` (`killGroup`, partagé avec l'interruption d'un agent), plus un arrêt propre du
   serveur dans `main.go` : sans lui, un Ctrl+C laissait les serveurs de recette en vie.
2. **Si l'usage le demande.** Plusieurs commandes nommées par dépôt (avec un menu sur le bouton),
   TTL d'inactivité, bouton « demander à un agent » pour rédiger la commande, suggestion
   préremplie à la création du projet, fichiers produits rendus par la couche livrables existante.
3. **Garé volontairement.** Sign-off et gate de livraison, `.sillage/recipes.json` avec
   approbation, allocateur de port, proxy d'URL stable.

## 7. Impacts sur les specs courantes (faits pour le lot 1)

- `SPEC-API.md` : `Repo` (+`previewCmd`, +`previewUrl`) ; modèle `PreviewRun` ; `State`
  (+`previews`, pour l'hydratation) ; endpoints `POST /api/cards/{id}/preview {repoName?}`,
  `POST /api/tasks/{id}/preview`, `POST /api/previews/{id}/stop`, `GET /api/previews/{id}/log` ;
  événements SSE `preview` et `previewLog` ; section « Recette manuelle » (exécution, variables,
  journal, arrêt) ; règle UI du bouton. Les mutations n'exigent **pas** `{"confirm": true}` :
  lancer ou arrêter un process local est réversible et n'a rien de sortant.
- `SPEC-BACKEND.md` : section « Superviseur de recette » (un run par worktree, tuyau unique
  stdout+stderr, tampon de journal, `expandPreviewURL`, `killGroup` partagé, `Server.Shutdown`).
- `CONTRIBUTING.md` : cinquième invariant. La recette n'introduit aucun chemin de push ; la
  commande vient des réglages, jamais d'un fichier de dépôt ; les runs n'ont lieu que dans un
  worktree de Sillage et ne survivent pas au serveur.

## 8. Écarts entre l'intention et l'implémentation

Notés pour que la décision reste traçable :

- **Panneau, pas tiroir.** Le §4 décrivait un tiroir ; l'implémentation réutilise la modale
  existante (`openModal`), qui a déjà son piège de tabulation, sa fermeture au fond et à `Esc`.
  Un tiroir aurait été un nouveau composant pour le même contenu.
- **`previewUrl` développée par le shell**, et non substituée en Go. Sans ça, il aurait fallu
  deux syntaxes de variables : celle de la commande (shell, avec `$(( ))`) et celle de l'URL
  (substitution maison). Une seule règle à documenter valait le coût d'un `sh` de plus au
  lancement.
- **Pas de champ de recette à la création d'un projet** : les lignes de dépôt en mode création
  n'affichent que le chemin. On ne sait pas encore comment le projet se lance ; le champ arrive
  dans les réglages, quand la question a un sens.
