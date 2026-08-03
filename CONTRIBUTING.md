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

1. Agents never get push rights: no `git push` outside `Ship()` and `SyncPush()` in `internal/server/git.go`, no push-capable entries in agent tool allowlists, no permission-bypass flags for agent CLIs. `Ship()` pushes a workstream branch; `SyncPush()` only ever operates on the data directory, never on a project repository.
2. Shipping requires an explicit human confirmation (`{"confirm": true}`) on an authenticated request. Shipping a workstream is the only outbound action in the product.
3. Local-merge delivery never pushes and never fast-forwards anything but the target branch, and never switches branches in your working repository: pushing a shared branch stays a human decision, taken in a terminal.
4. The server must stay safe to run on a laptop: localhost by default, hashed password, rate-limited login.

## Style

- Go: idiomatic, small surface, no premature abstraction.
- UI: white, dense but airy, few words. Lists inform, the detail panel acts. One primary action per state.
- User-facing strings are currently in French; i18n is on the roadmap and help is welcome there.

## Reporting bugs and proposing features

Use the issue templates. For security issues, see [SECURITY.md](SECURITY.md): do not open a public issue.
