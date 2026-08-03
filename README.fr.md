# Sillage

**Un tableau de bord web, calme et épuré, pour piloter des agents IA (Claude Code, Codex CLI) sur tous vos projets.**

Sillage fait travailler vos agents dans des worktrees git isolés, diffuse leur activité en direct, compte chaque token dépensé, et ne laisse rien sortir de votre machine sans votre validation explicite. Un binaire, un onglet de navigateur, fini le jonglage entre terminaux.

*[English README](README.md)*

![Vue tâche : liste style PR, conversation avec l'agent, une action primaire](docs/screenshots/task-detail.png)

## Pourquoi

Travailler avec plusieurs agents IA tourne vite à l'enfer des onglets : cinq terminaux, trois dépôts, « attends, je relisais quelle tâche déjà ? ». Sillage est construit autour d'une idée : **réduire la charge mentale**. Chaque projet est un kanban ; chaque carte un petit objectif ; chaque tâche un agent sur une branche. Vous relisez, vous validez, vous livrez. Le sillage de votre travail reste visible : branches, diffs, livrables.

## Fonctionnalités

- **Kanban par projet** (Bientôt / En cours / Terminé) de **chantiers** : des mini-objectifs qui regroupent des tâches et changent de colonne automatiquement quand le travail démarre et se termine.
- **Tâches façon pull requests** : icônes d'état, ligne d'activité en direct, compteurs, filtres en pilules, liens profonds (`#/p/…/c/…/t/…`). Les listes informent, le détail agit, une action primaire par état.
- **Conversation avec l'agent** (markdown, blocs de code, rendu incrémental qui ne casse jamais la sélection de texte), **diff coloré**, **livrables** en onglets. Les messages envoyés pendant que l'agent travaille sont mis en file et transmis dès qu'il termine.
- **Validation humaine structurelle** : les agents travaillent dans un worktree git dédié sur une branche `sillage/<ref>-...` et n'ont jamais le droit de pousser. `git push` n'existe qu'à un seul endroit du code, déclenché uniquement par le bouton de livraison après confirmation explicite. La livraison affiche une ligne système avec le lien direct vers la branche poussée.
- **Contexte en couches pour les agents** : chaque agent reçoit son propre prompt, plus le contexte du projet, plus celui du chantier. On écrit les choses une fois, au bon niveau.
- **Un ou plusieurs dépôts git par projet** ; chaque tâche travaille sur exactement un dépôt.
- **Comptage des tokens partout** : entrée/sortie/coût par tâche, par projet et global, en temps réel.
- **Espace de travail versionné git** (optionnel) : projets, conversations et réglages vivent dans un dépôt git local avec commits automatiques ; synchronisation manuelle vers un remote privé pour passer d'une machine à l'autre.
- **Agents gérés depuis l'interface** (nom, emoji, CLI, modèle, prompt), avec avertissements de santé en amont (CLI absent, sandbox bloqué) et réassignation d'une tâche à un autre agent en un clic.
- **Ouvrir une PR** GitHub pour une tâche livrée (`gh pr create`, avec la même confirmation explicite que la livraison).
- **Liens épinglés par projet** : une rangée discrète de favicons pour tes sites, dépôts et dashboards.
- **UI en français et en anglais**, détection automatique, bascule dans les préférences.
- **Mobile** (sidebar en tiroir), parfait via Tailscale depuis le téléphone.
- **Agent gratuit intégré** (Écho 🧪) pour essayer tout le workflow sans dépenser un token.

![Vue kanban](docs/screenshots/kanban.png)

## Démarrage

Prérequis : Go 1.24+, git, et au moins un CLI d'agent ([Claude Code](https://docs.anthropic.com/en/docs/claude-code) et/ou [Codex CLI](https://github.com/openai/codex)) connecté.

```bash
go install github.com/Halleck45/sillage@latest
sillage
```

Ou depuis les sources : `go build -o sillage . && ./sillage`

Puis ouvrir http://127.0.0.1:8787. Au premier lancement, un mot de passe est généré et affiché une seule fois dans le terminal. Pour choisir le vôtre : `SILLAGE_PASSWORD=votremotdepasse sillage`.

Ajoutez un projet (n'importe quel dépôt git local), créez une carte, une tâche, choisissez un agent : il démarre immédiatement dans son worktree. Commencez par Écho si vous voulez voir le flux sans aucun coût d'API.

## Modèle de sécurité

- Agents en mode headless avec une liste blanche d'outils figée : éditions de fichiers et commandes en lecture dans le worktree de la tâche. Jamais `git push`, jamais de contournement des permissions.
- Serveur sur `127.0.0.1` par défaut. Mot de passe (bcrypt), cookies de session HttpOnly, rate-limit de connexion, Content-Type JSON exigé sur les mutations.
- **Ne jamais exposer le port HTTP nu sur Internet.** Pour l'accès distant : [Tailscale](https://tailscale.com) (`sillage -addr <ip-tailscale>:8787`) ou un reverse proxy TLS ; le cookie passe en Secure derrière `X-Forwarded-Proto: https`.
- Données dans `~/.local/share/sillage` (état JSON, écritures atomiques) plus un worktree git par tâche.

Note codex : Sillage le lance avec `--sandbox workspace-write`. Sur les machines où AppArmor bloque bubblewrap (`bwrap: Operation not permitted`), soit autoriser les namespaces non privilégiés (`sudo sysctl kernel.apparmor_restrict_unprivileged_userns=0`), soit lancer Sillage avec `SILLAGE_CODEX_SANDBOX=danger-full-access`, en sachant que le confinement restant est celui de Sillage (worktree dédié, pas de push, validation humaine).

## Agents

Quatre agents sont créés au premier lancement (modifiables dans le fichier d'état pour l'instant, UI à venir) : Bolt 🐝 (claude/sonnet, dev backend), Muse 🦊 (claude/opus, produit et docs), Otto 🦉 (codex, infra), Écho 🧪 (factice local, gratuit). Chaque agent a un prompt de contexte et un modèle ; les conversations reprennent la session du CLI sous-jacent.

## Architecture

Binaire Go unique, frontend vanilla embarqué (pas de framework, pas de build), persistance JSON, SSE. Contrat HTTP/JSON/SSE : [docs/SPEC-API.md](docs/SPEC-API.md) ; internes : [docs/SPEC-BACKEND.md](docs/SPEC-BACKEND.md).

## Feuille de route

Suppression et archivage de projets, autres adaptateurs de CLI d'agents, autres langues d'interface.

Le multi-utilisateur a été essayé puis volontairement retiré : Sillage reste un outil personnel, et cette simplicité est une fonctionnalité.

Contributions bienvenues : voir [CONTRIBUTING.md](CONTRIBUTING.md).

## Licence

[MIT](LICENSE)
