# Changelog

## v0.1.0 (2026-08-02)

First usable release.

- Kanban per project (Soon / Doing / Done), cards with live activity, counters and progress.
- Tasks in dedicated git worktrees on `sillage/<ref>-slug` branches, PR-style list, detail panel with conversation / diff / deliverables tabs.
- Agent adapters: Claude Code (stream-json, session resume, token and cost tracking), Codex CLI (best effort, workspace-write sandbox), and Écho, a free built-in simulated agent.
- Workflow running / review / ready / shipped with a single primary action per state; shipping performs the only `git push` in the codebase and requires explicit confirmation.
- Token accounting per task, per project and global, live over SSE.
- Password auth (bcrypt), sessions, login rate limiting, CSRF protection, localhost by default.
- Responsive UI (drawer sidebar on mobile), French.
