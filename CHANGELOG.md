# Changelog

## Unreleased

- Codex agents can now commit from Sillage's linked task worktrees: the runner discovers external Git metadata and grants only the required directories to `workspace-write`, keeping shared hooks and repository configuration read-only.
- GitHub Copilot (`copilot`) and Google Antigravity (`agy`) are now full agent types, with safe non-interactive runners and two new seeded profiles. Missing Claude, Codex, Copilot, or Antigravity CLIs are shown as not connected; clicking the agent displays the official installation command, also directly from the new-task picker.
- Antigravity asks for approval before every command, which nobody can give while an agent works alone: its tasks used to end silently, with an empty conversation and an empty diff. The agent now carries a warning with the one settings line that lets commands run inside the sandbox Sillage already forces, and a "Set it for me" button that writes it: your other Antigravity settings are kept, and an unreadable file is never overwritten. And a text agent that exits successfully without saying anything is no longer silent: what it wrote on stderr lands in the conversation. Antigravity agents are also told that their `.git` is a pointer file they must not open, because that one read is refused and would end their run.
- A task can now be created to wait for another one: pick "Start after" in the new task form and it stays queued until the task you're waiting on gets accepted, then starts on its own with the prompt you gave it. Useful when a task genuinely needs code that another task hasn't landed yet. A waiting task can still be started manually at any time, or cancelled like any other.
- A workstream can now be shipped while some of its tasks are still running or waiting for review: what is accepted goes out now, the rest ships later. Nothing unreviewed can leak, because a workstream branch only ever contains accepted work; the ship bar and the recap announce the partial delivery ("2 tasks are not accepted yet"). The `tasks-pending` ship blocker is gone. A partially shipped workstream stays in "Doing": the "Done" column still means shipped *and* finished.
- Creating a project asks one question: the path to a git repository. The project name comes from the repository, the delivery mode from its remotes; description, instructions and links stay empty until you need them. `name` is now optional on `POST /api/projects`.
- Project settings became a modal with side navigation instead of eleven fields stacked in one scrolling column: General, Repositories, Instructions, Delivery, Links, Delete. Two or three fields per panel, no horizontal rules, and the project context prompt finally gets room to breathe.
- The delivery mode is picked from four cards, each carrying the sentence describing what it will do, instead of a dropdown with a caption that only changed once you had already changed your selection.
- The target branch moved to General under its real name, "Base branch": Sillage also uses it as the branch workstreams start from, so it is the project's reference branch, not a delivery sub-setting.
- Deleting a project moved out of an always-visible red block into its own panel at the bottom of the navigation. It keeps its two-step confirmation.
- Pressing `Enter` in the "add a link" field now adds the link instead of saving the project and dropping the URL you had just typed.
- Manual preview: a "Preview" button on a workstream and on a task runs the project in the matching worktree, so trying the software out no longer means opening a terminal and remembering the command. Sillage knows nothing about stacks; it runs one command per repository, written in the project settings.
- That command receives four variables, so a project can isolate itself without Sillage knowing what a database or a container is: `$SILLAGE_ID` (`ws-107`, `t-482`) for names, `$SILLAGE_N` (`107`) for arithmetic such as `PORT=$((4000 + SILLAGE_N))`, plus `$SILLAGE_DIR` and `$SILLAGE_BRANCH`. They derive from the workstream and task reference numbers: small, stable over time (so a preview database survives between sessions) and unique within a project, with no new state to allocate or persist.
- A server that stays alive and a script that finishes go through the same path: the panel streams stdout and stderr as they arrive, shows a clickable URL when the repository declares one, and reports "finished (exit 0)" otherwise. Web and non-web projects are covered by the same code, with no command type to choose.
- Starting a preview replaces the one already running in that worktree, and everything is killed when Sillage stops (`SIGINT`/`SIGTERM`), process groups included, so no forgotten server keeps a port. A counter at the bottom of the sidebar always says how many previews are running.
- With no command configured, the panel still shows the worktree path with a copy button: previewing works on every project without any setup.
- A conflict with the workstream branch, whether at accept time or during the automatic sibling rebase, no longer just leaves a note in the thread for a human to notice: the agent is immediately asked to rebase onto the workstream branch and resolve it, so the task goes back to "running" on its own instead of sitting stuck in review.
- New preview variable `$SILLAGE_PORT` (`4000 + $SILLAGE_N`, precomputed): a small workstream or task reference (e.g. 118) written directly as a port, or with the `$((4000 + SILLAGE_N))` offset forgotten, lands on a privileged port that refuses to bind (`permission denied`) with no clue why. `$SILLAGE_PORT` removes the arithmetic from the command entirely.
- The preview panel now links straight to the project's settings when a repository has no preview command configured, instead of only describing where to go look for it.
- The preview variables reminder next to the command field is now a row of chips (one per variable, with what it's for) plus a worked example, `python3 -m http.server $SILLAGE_PORT`, instead of one dense sentence listing all five.
- Sillage now knows its own version (`sillage -version`) and tells you when a newer one is out: a new Settings > Updates tab holds the version you run, where it was installed from, when it last looked, and the buttons to check or update; the only reminder outside it is one line at the bottom of the sidebar, never a banner. It checks GitHub for the latest release number once a day, which is a read (nothing from your machine is sent, no identifier, no telemetry) and can be switched off right there.
- Updating is one click when Sillage can do it safely: `brew update && brew upgrade sillage` for a Homebrew install, or a download whose sha256 is verified against the release's `checksums.txt` before anything is written, for a plain binary. Then Sillage restarts itself in place and the page reloads. The button stays off, with the reason, while an agent is working or a preview is running; a `go install` setup, or a binary in a directory Sillage cannot write to, gets the exact command to copy instead.
- `state.json` now carries the format version that wrote it, and Sillage refuses to start on a file written by a newer version instead of loading what it understands and silently dropping the rest on the next save. That silent rewrite was real: an older binary opening a current workspace erased workstream branches, delivery settings, preview commands and per-task counters, with no error anywhere. The refusal names the way out (`brew upgrade sillage`, or another `-data`), and nothing is written before the check, including when pulling a workspace from a remote. Two released versions sharing one format get a warning line at startup instead, since that case cannot destroy the file structure but can still drop fields.
- Settings > Updates now also says whether Sillage starts when you log in, and hands you the command when it doesn't. Homebrew installs only, because that is the only setup with a registry to ask (`brew services info`); anywhere else, the absence of a systemd unit or launch agent proves nothing, so Sillage says nothing rather than telling someone their working setup does not exist. It also warns when the running instance carries flags the service would not.
- Homebrew users can keep Sillage running in the background, started at login: `brew services start sillage`. The formula sets an explicit `PATH` so agent CLIs (`claude`, `codex`, `copilot`, `agy`, `gh`) stay visible to a service, which does not inherit your shell.

## v0.5.0 (2026-08-03)

- The password is now optional: no login is required unless `SILLAGE_PASSWORD` is set. No more random password generated and printed on first start.
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
