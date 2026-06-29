---
adr: "0019"
title: "v2 — log search and top sorting in the detail modals"
status: accepted
date: "2026-06-29"
supersedes: null
superseded_by: null
deciders:
  - "Kinn Coelho Juliao"
---

# 0019 — v2: log search and top sorting in the detail modals

> Ships on `feature/sockets` toward **v2.0.0**. Dashboard-only — it operates on the
> data already pushed in v1.6.0 (ADR 0017) and needs no socket transport, so it can
> land independently of [ADR 0018](0018-v2-persistent-socket-mediation.md).

## 1. Context

v1.6.0 added the `l` (logs) and `t` (top) detail-view modals. Two refinements were
requested:

- **Log search.** A way to find lines in the log modal, using the app's existing
  `/` keybind but scoped to that modal, returning the matching lines with their
  timestamps.
- **Top sorting.** Sort the process table — CPU-descending by default — with an
  `s`-key picker, and **persist** the chosen sort to the dashboard's config so it
  becomes the default on next launch.

## 2. Goals / Non-Goals

**Goals:**
- `/` inside the log view filters the buffered + live lines to substring matches,
  keeping timestamps; scoped to the modal, independent of the grid filter.
- `t` sorts CPU-desc by default; `s` opens a sort picker; the choice re-sorts live
  and **persists to the dashboard config JSON** as the new default.

**Non-Goals:**
- Regex / fuzzy log search (substring, case-insensitive — consistent with the grid
  filter).
- Sorting the grid or other views; this is the top modal only.

## 3. Proposal

### 3.1 Log search (`/` in the log view)
- `/` opens an inline search input scoped to `modalLogView` (own `logQuery` /
  `logSearching` state, not the grid's `filter`). `enter` keeps the query, `esc`
  clears it and exits the input; with no active input, `esc` still steps back to
  the source list.
- The log body shows only lines whose text contains the query (case-insensitive),
  each with its `HH:MM:SS` timestamp. The title shows the active query.

### 3.2 Top sorting (`s` picker, persisted)
- Sort options are a registry (`{key, label, less}`), mirroring the grouping /
  column / matcher registries: `cpu` (desc, default), `mem` (desc), `pid` (asc),
  `command` (asc). Adding a sort is registering one.
- The top modal sorts a copy of the process table by the active option before
  rendering. `s` opens a small picker modal (`modalTopSort`); `↑/↓`/`enter`
  selects, `esc` cancels.
- On select, the key persists to the dashboard config file under `top-sort`. A new
  `top-sort` catalog option (default `cpu`) is loaded on startup, so the persisted
  choice becomes the default. Persistence is a small JSON merge into the located
  config path (so a single runtime change doesn't require rewriting the full
  resolved catalog), injected into the model as a closure (Dependency Inversion).

## 4. Alternatives Considered

| Option | Why not |
|---|---|
| Reuse the grid `/` filter state for log search | couples two unrelated searches; leaks log queries into the grid |
| A separate prefs file for the sort | the dashboard already has a config; one source of truth is simpler |
| Persist via `options.Sink.Write` (whole catalog) | needs a mutated `Resolved`; a scoped JSON merge is lower-risk for one runtime change |

## 5. Trade-offs and Risks
- The sort picker adds a fourth modal state; the `esc` unwind contract must stay
  consistent (picker → top, search input → log view, view → list → detail).
- Persisting on every change writes the config file; cheap, but it means the file
  is touched from the TUI, not only at first-run/flags.

## 6. Impact
**SRE/Security:** No new surface; dashboard-local. Config write is owner-only
(0600), same as today. **Team:** One more registry + modal; localized.

## 7. Decision
Add a modal-scoped `/` log search and an `s` top-sort picker (CPU-desc default)
that persists to the dashboard config, both built from the existing registry +
filter-input patterns. Dashboard-only; ships on `feature/sockets`.

Status: **accepted**
