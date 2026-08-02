# Changelog

## v0.2.0 (2026-08-02)

- English and French UI with a sidebar language switch (auto-detected from the browser).
- Agent management from the UI: create, edit and delete agents (deletion refused while tasks reference them).
- "Open a PR" for shipped tasks: `gh pr create` on the already-pushed branch, or a read-only GitHub compare URL as fallback; explicit confirmation required, never pushes.
- Per-project check command editable from the UI.
- Messages carry the author's display name.
- API error messages are now in English.
- Releases now ship raw binaries instead of tar.gz archives.

## v0.1.0 (2026-08-02)

First usable release.

- Kanban per project (Soon / Doing / Done), cards with live activity, counters and progress.
- Tasks in dedicated git worktrees on `sillage/<ref>-slug` branches, PR-style list, detail panel with conversation / diff / deliverables tabs.
- Agent adapters: Claude Code (stream-json, session resume, token and cost tracking), Codex CLI (best effort, workspace-write sandbox), and Écho, a free built-in simulated agent.
- Workflow running / review / ready / shipped with a single primary action per state; shipping performs the only `git push` in the codebase and requires explicit confirmation.
- Token accounting per task, per project and global, live over SSE.
- Password auth (bcrypt), sessions, login rate limiting, CSRF protection, localhost by default.
- Responsive UI (drawer sidebar on mobile), French.
