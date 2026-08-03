# Changelog

## Unreleased

- Workstreams are feature branches. A workstream now owns a `sillage/ws-<ref>-<slug>` branch per repository, with its own worktree; task branches start from it and are merged into it when accepted, so a task created after an acceptance builds on the work already accepted.
- Shipping moved from the task to the workstream: one button at the top of the workstream view, with a one-line announcement of what it will do, and a recap panel whose action button is the confirmation. Two clicks instead of four, and the second one informed.
- Reviewing a task is now Accept / Refuse, revealed on hover in the task list. Accepting merges locally (no confirmation, reversible with Reopen); a merge conflict leaves the task in review and says which files clash.
- New per-project delivery setting: open a pull request (GitHub via `gh`) or a merge request (GitLab via `glab`), or merge locally into a branch of your choice. Detected from the repository remotes at project creation. Local merge is fast-forward only and never pushes.
- Health warnings when `gh` or `glab` is missing, when a repository has no `origin` remote, or when the forge is unknown. Everything keeps working: the fallback is a prefilled pull request URL.
- Task statuses are now `running / review / accepted / cancelled` (`shipped` and `done` migrate to `accepted`). The Done column now means shipped, not merely reviewed.
- Branches you merge by hand into the workstream branch are noticed: the task is marked accepted (only when no agent is running and its worktree is clean), on opening the workstream view and every 60 seconds after that.
- A shipped workstream keeps living: adding a task, or accepting one that brings new commits, marks it as not fully shipped again, so the button comes back. The next delivery pushes only the new commits and updates the existing pull request instead of opening a second one.
- Removed: per-task ship, per-task "Open the PR", and "Mark as completed".
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
