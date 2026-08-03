# Sillage

[![CI](https://github.com/Halleck45/sillage/actions/workflows/ci.yml/badge.svg)](https://github.com/Halleck45/sillage/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/Halleck45/sillage)](https://github.com/Halleck45/sillage/releases/latest)
[![Go version](https://img.shields.io/github/go-mod/go-version/Halleck45/sillage)](go.mod)
[![License: MIT](https://img.shields.io/badge/license-MIT-2f7d54)](LICENSE)

**A calm web dashboard to pilot AI coding agents (Claude Code, Codex CLI) across your projects.**

Sillage runs your agents in isolated git worktrees, streams their activity live, tracks every token they spend, and never lets anything leave your machine without your explicit approval. **One self-contained binary, zero dependencies, zero configuration**: download, run, open a browser tab.

![Task view: PR-style task list, conversation with the agent, one primary action](docs/screenshots/task-detail.png)

## Why

Working with several AI agents quickly turns into tab hell: five terminals, three repos, "wait, which task was I reviewing?". Sillage is built around one idea: **reduce mental load**. Each project is a kanban; each card is a small goal; each task is one agent working on one branch. You review, you approve, you ship. The wake (sillage, in French) of your work stays visible: branches, diffs, deliverables.

## Features

- **Kanban per project** (Soon / Doing / Done) of **workstreams**: small goals grouping tasks, moving across columns automatically as work starts and finishes.
- **Tasks like pull requests**: status icons, live activity line, counters, filter pills, deep links (`#/p/…/c/…/t/…`). Lists inform, the detail panel acts, one primary action per state.
- **Conversation with the agent** (markdown, code blocks, incremental rendering that never breaks your text selection), **colored diff**, **deliverables** in tabs. Messages sent while the agent is busy are queued and delivered when it finishes.
- **Human validation is structural, not cosmetic**: agents work in a dedicated git worktree on a `sillage/<ref>-...` branch and are never allowed to push. `git push` exists in exactly one place in the code, triggered only by the ship button after an explicit confirmation. Shipping posts a system line with a direct link to the pushed branch.
- **Layered context for agents**: each agent gets its own prompt, plus the project context, plus the workstream context. Write things once, at the right level.
- **One or several git repositories per project**; each task works on exactly one of them.
- **Token accounting everywhere**: input/output/cost per task, per project and global, updated live.
- **Workspace backed by git** (optional): your projects, conversations and settings live in a local git repository with automatic local commits; sync it manually to a private remote to move between machines.
- **Agents managed from the UI** (name, emoji, CLI, model, context prompt), with upfront health warnings (missing CLI, blocked sandbox) and one-click task reassignment to another agent.
- **Open a PR** on GitHub for a shipped task (`gh pr create`, with the same explicit confirmation as shipping).
- **Pinned links per project**: a discreet favicon row for your sites, repos and dashboards.
- **English and French UI**, auto-detected, switchable in preferences.
- **Mobile friendly** (drawer sidebar), great over Tailscale from your phone.
- **Built-in free agent** (Écho 🧪) to try the whole workflow without spending a token.

![Kanban view](docs/screenshots/kanban.png)

## Quickstart

Sillage is a single static binary with the web UI embedded: no database, no runtime, no config files to write. The only things it talks to are `git` and your agent CLIs ([Claude Code](https://docs.anthropic.com/en/docs/claude-code) and/or [Codex CLI](https://github.com/openai/codex), logged in; the built-in free agent works without either).

```bash
# Linux (amd64)
curl -fsSL https://github.com/Halleck45/sillage/releases/latest/download/sillage_linux_amd64 -o sillage
chmod +x sillage && ./sillage
```

macOS: replace with `sillage_darwin_arm64` (Apple Silicon) or `sillage_darwin_amd64` (Intel). Also available: `linux_arm64`, [checksums](https://github.com/Halleck45/sillage/releases/latest), or `go install github.com/Halleck45/sillage@latest` if you prefer building it yourself.

Then open http://127.0.0.1:8787. On first start, a password is generated and printed once in the terminal. To choose your own: `SILLAGE_PASSWORD=yourpassword sillage`.

Add a project (any local git repository), create a card, create a task, pick an agent: it starts working immediately in its own worktree. Try the free agent Écho first if you want to see the flow without any API cost.

## Security model

- Agents run headless with a fixed allowlist of tools: file edits and read-only commands inside the task worktree. Never `git push`, never permission bypass flags.
- The server binds to `127.0.0.1` by default. Password login (bcrypt), HttpOnly session cookies, login rate limiting, JSON content-type enforcement on mutations.
- **Never expose the HTTP port directly to the internet.** For remote access use [Tailscale](https://tailscale.com) (`sillage -addr <tailscale-ip>:8787`) or a TLS reverse proxy; the session cookie switches to Secure automatically behind `X-Forwarded-Proto: https`.
- Data lives in `~/.local/share/sillage` (JSON state, atomic writes) plus one git worktree per task.

Note for codex agents: Sillage runs them with `--sandbox workspace-write`. On machines where AppArmor blocks bubblewrap (`bwrap: Operation not permitted`), either allow unprivileged user namespaces (`sudo sysctl kernel.apparmor_restrict_unprivileged_userns=0`) or start Sillage with `SILLAGE_CODEX_SANDBOX=danger-full-access`, knowing that the remaining containment is Sillage's own (dedicated worktree, no push, human validation).

## Agents

Four agents are seeded on first run and stored in the state file (editable there for now, UI coming):

| Agent | CLI | Role |
|---|---|---|
| Bolt 🐝 | claude (sonnet) | pragmatic backend developer |
| Muse 🦊 | claude (opus) | product, specs, documentation |
| Otto 🦉 | codex | infrastructure, CI, tooling |
| Écho 🧪 | built-in fake | free local agent for demos and tests |

Each agent has a context prompt appended to its system prompt, and a model. Conversations resume the underlying CLI session, so follow-up messages keep full context.

## Architecture

Single Go binary, embedded vanilla JS frontend (no framework, no build step), JSON file persistence, SSE for realtime. The HTTP/JSON/SSE contract is documented in [docs/SPEC-API.md](docs/SPEC-API.md) and the internals in [docs/SPEC-BACKEND.md](docs/SPEC-BACKEND.md).

```
main.go               server bootstrap, flags, embedded web/
internal/server/      store (JSON), auth, SSE hub, git worktrees, agent runners, handlers
web/                  index.html, style.css, app.js
```

## Roadmap

- Project deletion and archiving.
- More agent CLI adapters.
- More languages for the UI.

Multi-user support was tried and deliberately removed: Sillage stays a personal tool, and that simplicity is a feature.

Contributions welcome: see [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[MIT](LICENSE)
