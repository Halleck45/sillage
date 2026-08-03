# Changelog

## v0.5.0 (2026-08-03)

- Workstreams are feature branches. A workstream now owns a `sillage/ws-<ref>-<slug>` branch per repository, with its own worktree; task branches start from it and are merged into it when accepted, so a task created after an acceptance builds on the work already accepted.
- Shipping moved from the task to the workstream: one button at the top of the workstream view, with a one-line announcement of what it will do, and a recap panel whose action button is the confirmation. Two clicks instead of four, and the second one informed.
- Reviewing a task is now Accept / Refuse, revealed on hover in the task list. Accepting merges locally (no confirmation, reversible with Reopen); a merge conflict leaves the task in review and says which files clash.
- Accepting a task rebases the workstream's other tasks in review on top of it, in the background, so the next one no longer starts behind. Their pending work is committed first and never lost, agents at work are left alone, and a conflicting rebase is aborted (nothing changed) and reported in the thread. The task row shows an amber spinner while it happens.
- Task state icons are now tinted pills instead of bare glyphs: the state of a list is readable at a glance.
- The Ship button tells the truth about what it can do: an anchor and "Already on <branch>" once the workstream has landed in its target branch (merged by Sillage or by hand), and a disabled button when the target has moved on and a fast-forward merge can no longer succeed.
- In that last case, a "Catch up with <branch>" button sits next to Ship: it merges the target branch into the workstream branch (a merge, not a rebase, so the accepted task branches keep their base) and shipping becomes possible again. A conflicting catch-up is aborted, leaves the workstream untouched, and offers to hand the job to an agent in a prefilled task.
- The workstream header no longer repeats the task counts: the list filters right below already carry them.
- Buttons are black across the app, the Ship button carries a boat, and the sidebar wears the product logo.
- New per-project delivery setting, four modes: open a pull request (GitHub via `gh`) or a merge request (GitLab via `glab`); push the workstream branch without opening anything; merge into a target branch locally without ever pushing; or merge into the target branch and push it. Detected from the repository remotes at project creation, editable there and in the project settings. Both merge modes are fast-forward only; the pushing one catches the target branch up from `origin` first and refuses, before writing anything, if it has really diverged. Never `--force`.
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
