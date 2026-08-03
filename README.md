# Sillage

[![CI](https://github.com/Halleck45/sillage/actions/workflows/ci.yml/badge.svg)](https://github.com/Halleck45/sillage/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/Halleck45/sillage)](https://github.com/Halleck45/sillage/releases/latest)
[![Go version](https://img.shields.io/github/go-mod/go-version/Halleck45/sillage)](go.mod)
[![License: MIT](https://img.shields.io/badge/license-MIT-2f7d54)](LICENSE)

**A calm web dashboard to pilot AI coding agents (Claude Code, Codex CLI) across your projects.**

Sillage runs your agents in isolated git worktrees, streams their activity live, and never lets anything leave your machine without your explicit approval. **One self-contained binary, zero dependencies, zero configuration**: download, run, open a browser tab.

![Task view: PR-style task list, conversation with the agent, one primary action](docs/screenshots/task-detail.png)

## Why

Working with several AI agents quickly turns into tab hell: five terminals, three repos, "wait, which task was I reviewing?". Sillage is built around one idea: **reduce mental load**. Each project is a kanban; each workstream is a small goal and a feature branch; each task is one agent working on its own branch. You review a task and accept it, then you ship the workstream: one outbound action, once. The wake (sillage, in French) of your work stays visible: branches, diffs, deliverables.

## Features

- **Delegate, then watch**: a kanban of workstreams per project, tasks like pull requests, live agent activity streamed to your browser.
- **Agents can never push**: each task runs in an isolated git worktree; accepting a task merges it locally into the workstream branch, and shipping that branch is the only outbound action, behind one confirmed click. For project repositories, `git push` exists in exactly one function in the code, reached only from that click.
- **Ship your way**: you decide, per project, what Ship does: open a pull request (GitHub, GitLab), push the branch without opening anything, merge into the branch of your choice locally, or merge into it and push it. Fast-forward only, never `--force`. Branches you merged by hand are noticed, and a shipped workstream can be picked up again.
- **Real conversations**: chat with the agent mid-task, review the colored diff. Messages sent while it works are queued, context never gets lost.
- **Layered context**: agent prompt + project context + workstream context. Write things once, at the right level.
- **Token usage, out of the way**: no counters cluttering the kanban or the task view; per-project token totals live in Settings > Statistics.
- **Your workspace is a git repo** (optional): conversations, projects and settings versioned locally, synced to a private remote to follow you across machines.

- **Keyboard first**: `N` creates what the screen holds, `Ctrl/⌘+K` searches, `?` lists every shortcut. In the new-task form, "Create and add another" (`Ctrl/⌘+Shift+Enter`) keeps the agent selected so you can queue a batch of tasks without touching the mouse.

Also: multiple repositories per project, task reassignment, agent health warnings, pinned project links, English and French UI, works from your phone, and a free built-in agent to try everything without spending a token.

![Kanban view](docs/screenshots/kanban.png)

## Quickstart

Sillage is a single static binary with the web UI embedded: no database, no runtime, no config files to write. The only things it talks to are `git` and your agent CLIs ([Claude Code](https://docs.anthropic.com/en/docs/claude-code) and/or [Codex CLI](https://github.com/openai/codex), logged in; the built-in free agent works without either).

```bash
curl -fsSL https://raw.githubusercontent.com/Halleck45/sillage/main/install.sh | sh
sillage
```

Or with Homebrew:

```bash
brew tap halleck45/sillage https://github.com/Halleck45/sillage
brew trust halleck45/sillage   # Homebrew requires explicit trust for third-party taps
brew install sillage
```

Prefer doing it by hand? Grab a binary from the [releases page](https://github.com/Halleck45/sillage/releases/latest), `chmod +x`, run it. Or `go install github.com/Halleck45/sillage@latest`.

Then open http://127.0.0.1:8787. No login is required by default. To require a password, set `SILLAGE_PASSWORD=yourpassword sillage`.

Add a project (any local git repository), create a card, create a task, pick an agent: it starts working immediately in its own worktree. Try the free agent Écho first if you want to see the flow without any API cost.

## Security model

- Agents run headless with a fixed allowlist of tools: file edits and read-only commands inside the task worktree. Never `git push`, never permission bypass flags.
- The server binds to `127.0.0.1` by default. Optional password login (bcrypt, `SILLAGE_PASSWORD`), HttpOnly session cookies, login rate limiting, JSON content-type enforcement on mutations.
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
