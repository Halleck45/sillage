# Contributing to Sillage

Thanks for considering a contribution. Sillage values calm and simplicity, in the product and in the code.

## Getting started

```bash
git clone https://github.com/Halleck45/sillage
cd sillage
go build -o sillage . && SILLAGE_PASSWORD=dev ./sillage
```

The frontend is plain HTML/CSS/JS in `web/`, embedded in the binary: rebuild and restart to see changes. The free agent Écho 🧪 lets you exercise the whole task workflow without any API cost.

## Before opening a PR

- `go vet ./...`, `go test ./...` and `gofmt -l .` must be clean (CI enforces this).
- `node --check web/app.js` for frontend changes.
- Keep the dependency count where it is: the only external Go dependency is `golang.org/x/crypto`. Adding a dependency needs a very good reason.
- Read the two design documents first: [docs/SPEC-API.md](docs/SPEC-API.md) (HTTP/JSON/SSE contract) and [docs/SPEC-BACKEND.md](docs/SPEC-BACKEND.md) (internals). If your change alters the contract, update the spec in the same PR.

## Non-negotiable invariants

These are the product's security promises. PRs that weaken them will not be merged:

1. Agents never get push rights: no `git push` outside `pushBranch()` and `SyncPush()` in `internal/server/git.go`, no permission-bypass flags for agent CLIs. `pushBranch()` has exactly two callers, both in the ship path: `Ship()` (workstream branch) and `mergeThenPush()` (target branch, `merge-push` mode only). `SyncPush()` only ever operates on the data directory, never on a project repository.

   Agent tool permissions are stated, not implied by omission. Part of the allowlist is a per-project setting the human types, so the ban lives in an explicit deny list passed to the agent CLI, covering everything push-capable (`git push`, `gh`, `glab`) plus the files that steer the agent itself (`.claude/settings*.json`, hooks). Deny beats allow, so no project setting can open those doors. That deny list is a backstop against a slip, not a sandbox: the real containment is the dedicated worktree, the throwaway branch, and human review before anything leaves. Grow it when a new outbound binary appears; never shrink it.
2. Shipping requires an explicit human confirmation (`{"confirm": true}`) on an authenticated request. Shipping a workstream is the only outbound action in the product.
3. Merge delivery is fast-forward only, never switches branches in your working repository, and fast-forwards nothing but the target branch. The `merge` mode never pushes at all; `merge-push` pushes the target branch and nothing else, never with `--force`, and refuses (before writing anything) when the target has really diverged from `origin`. Which of the two applies is a per-project setting the human made, not a default.
4. The server must stay safe to run on a laptop: localhost by default, hashed password, rate-limited login.
5. Manual preview commands come from the project settings only, typed by the human (same trust level as `checkCmd`). No command is ever read from a repository file, because repository files are written by agents: that would turn a branch into an execution vector triggered by a human who believes they are just launching their app. Preview runs happen in a Sillage worktree, never in your own working repository, and none of them survives the server: SIGINT/SIGTERM stops them all.

## Style

- Go: idiomatic, small surface, no premature abstraction.
- UI: white, dense but airy, few words. Lists inform, the detail panel acts. One primary action per state.
- User-facing strings are currently in French; i18n is on the roadmap and help is welcome there.

## Reporting bugs and proposing features

Use the issue templates. For security issues, see [SECURITY.md](SECURITY.md): do not open a public issue.
