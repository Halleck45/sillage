# Sillage

App web (binaire Go unique) pour piloter des agents IA (Claude Code, Codex) sur plusieurs projets, sans jongler entre 50 fenêtres : kanban par projet, tâches avec conversation, diff, livrables, suivi des tokens, et validation humaine obligatoire avant tout push.

## Lancer

```bash
go build -o sillage . && ./sillage
```

- Écoute par défaut sur `http://127.0.0.1:8787`.
- Au premier lancement, un mot de passe est généré et affiché une seule fois dans le terminal (hash stocké dans `~/.local/share/sillage/config.json`). Pour le changer : relancer avec `SILLAGE_PASSWORD=nouveau ./sillage` (le hash est remplacé).
- Données : `~/.local/share/sillage/state.json` (écriture atomique). Worktrees git des tâches : `~/.local/share/sillage/worktrees/`.

## Accès distant (téléphone)

Le serveur ne parle que HTTP : ne jamais l'exposer nu sur Internet. Deux options sûres :

1. **Tailscale** (recommandé) : `./sillage -addr 100.x.y.z:8787` (ou `-addr 0.0.0.0:8787` si le pare-feu ne laisse passer que tailscale0), puis ouvrir l'URL Tailscale depuis le téléphone. Chiffré de bout en bout, zéro config TLS.
2. Reverse proxy TLS (Caddy, nginx) devant `127.0.0.1:8787` ; le cookie de session passe en `Secure` automatiquement derrière `X-Forwarded-Proto: https`.

## Principes

- **Validation humaine obligatoire** : les agents travaillent dans un worktree git isolé, sur une branche `sillage/<ref>-...`. Aucun agent n'a le droit de pousser. Le `git push` n'existe qu'à un seul endroit du code et n'est déclenché que par le bouton « Pousser et livrer » avec confirmation explicite.
- **Tokens** : entrée/sortie/coût cumulés par tâche, agrégés par projet et en global (sidebar).
- **Agents** : Bolt 🐝 (claude, sonnet), Muse 🦊 (claude, opus), Otto 🦉 (codex), Écho 🧪 (agent factice local, gratuit, pour essayer l'interface).

## Note sur Otto (codex)

Par défaut, les agents codex tournent en `--sandbox workspace-write`. Sur certaines machines (dont celle-ci), AppArmor bloque bwrap (`bwrap: Operation not permitted`) et codex ne peut alors rien écrire. Deux issues :

1. Autoriser les namespaces non privilégiés (le vrai correctif, à faire en connaissance de cause) : `sudo sysctl kernel.apparmor_restrict_unprivileged_userns=0`.
2. Lancer Sillage avec `SILLAGE_CODEX_SANDBOX=danger-full-access` : codex perd son isolation système ; le confinement restant est celui d'Sillage (worktree dédié, aucun push possible par l'agent, validation humaine avant livraison).

## Développement

- `go test ./...`, `go vet ./...`.
- Specs : `SPEC-API.md` (contrat HTTP/SSE), `SPEC-BACKEND.md` (interne), maquette dans `mockups/`.
