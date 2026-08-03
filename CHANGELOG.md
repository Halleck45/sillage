# Changelog

## Unreleased

- Keyboard-first creation: `N` creates the object of the current screen (task, workstream or project), `?` opens a shortcut cheatsheet, `/` opens search, `Esc` closes the task panel.
- "Create and add another" in the new-task form (`Ctrl/⌘+Shift+Enter`): the form stays open with the agent and repository kept, so a batch of tasks can be queued in a row. "Create and chat" is `Ctrl/⌘+Enter`.
- Creation buttons now carry a `+` pictogram and their shortcut badge.
- Better keyboard navigation: the agent picker is an arrow-navigable radio group, `Tab` stays inside the open modal, and search results are reachable with `↑` `↓` and `Enter`.

## v0.4.0 (2026-08-03)

First public release. Highlights since 0.2: workstreams with layered agent context, one or several git repositories per project, git-backed workspace sync, simplified review-then-ship workflow, task lifecycle (finish, cancel, reassign, delete), message queueing while agents run, deep links, pinned project links, agent health warnings, incremental conversation rendering.

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
