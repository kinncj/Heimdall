# Using the Dashboard — Driving Heimdall After Setup

You have a hub running and at least one host reporting (if not, start with the
[Quickstart](01-quickstart.md) or [Monitor a Fleet](02-monitor-a-fleet.md)). This
guide is the other half: once metrics are flowing, how do you actually *drive* the
terminal UI — navigate the fleet, drill into a host, read logs, run a command,
search and filter, and open the full-screen top view.

Everything here is keyboard-first; the mouse is additive, never required. Press
**`?`** at any time for the in-app key reference.

```
heimdall-dashboard            # talks to a local hub
heimdall-dashboard --hub host:8443
heimdall-dashboard --demo     # fake fleet, no setup — a safe place to learn the keys
```

## 1. The fleet grid

The opening screen is your whole fleet, one row per host, with live metric columns
(CPU, MEM, DISK, TEMP, GPU, PWR — as many as the width allows) and a state badge.

| Key | Action |
|---|---|
| `↑/↓` · `j/k` | move the selection |
| `⏎` | open the selected host's **detail** view |
| `t` | open the full-screen **top** view for the selected host |
| `g` | cycle grouping: none → hub → OS → tag keys → none |
| `/` | filter the fleet by host name, tag, hub, OS, or state |
| `r` | refresh now |
| `?` | toggle the key reference |
| `esc` | clear an active filter, else quit |
| `q` · `ctrl+c` | quit |

**Filtering** (`/`): type to narrow, `⏎` keeps it, `esc` clears it. A filter that
matches nothing shows an empty state rather than a stale list.

## 2. Host detail

Press `⏎` on a host to open its detail view — the full metric read-out plus the
entry points to logs, processes, and commands.

| Key | Action |
|---|---|
| `↑/↓` | previous / next host (stay in detail while you walk the fleet) |
| `⇧↑/↓` · `pgup/pgdn` · wheel | scroll the detail body |
| `home/end` | jump to top / bottom |
| `l` | **logs** (if the host streams any) |
| `p` | **process** table (if offered) |
| `c` | **command** picker (if the host allows commands) |
| `t` | full-screen **top** view |
| `esc` | back to the fleet |

`l`, `p`, and `c` only appear when the host actually offers them — a daemon opts
into logs, a process table, and allow-listed commands separately.

## 3. Logs (`l`)

Pick a source from the list (`↑/↓`, `⏎`), then stream it live.

| Key | Action |
|---|---|
| `↑/↓` · wheel | scroll; the view starts pinned to the newest line |
| `/` | **search** — type to filter lines (case-insensitive) |
| `esc` | clear an active search, then step back to the source list |

## 4. Process table (`p`)

A live top-style process list for the host.

| Key | Action |
|---|---|
| `↑/↓` · wheel | scroll |
| `s` | change the sort column — `↑/↓` to pick (CPU%, MEM%, PID, COMMAND), `⏎` to apply |
| `/` | **filter by command** (case-insensitive); `esc` clears it |
| `esc` | clear an active filter, then back |

The chosen sort is remembered as your default for next time. Sort and filter stack —
filter to `chrome`, sort by `MEM%`, and you see the heaviest Chrome processes.

## 5. Commands (`c`)

Run an allow-listed diagnostic on the host and read the result inline. Filtering
works at **both** levels here.

**The picker:**

| Key | Action |
|---|---|
| `↑/↓` | pick a command |
| `/` | **filter by name** — type to narrow the list |
| `⏎` | run the selected command |
| `esc` | clear an active filter, then close |

**The result:**

| Key | Action |
|---|---|
| `↑/↓` · wheel | scroll the output |
| `/` | **filter the output lines** (case-insensitive) — handy for a long dump |
| `esc` | clear an active filter, then step back to the command list |

## 6. Full-screen top view (`t`)

A mactop/btop-style read-out of one host — per-core CPU, memory, power, GPU/NPU,
network, disk, and processes. It has its own focus and scrolling; the short version:

- `↑/↓` and the wheel scroll the **whole view** first (always works, any resolution).
- `tab` steps the focus ring into each panel and wraps back to the whole view.
- The mouse wheel/click targets the panel under the pointer.

The full keys and panel reference are in [Top View (Hliðskjálf)](12-top-view.md).

## Two things that work the same everywhere

**Search / filter (`/`).** The same key opens a filter in the logs, the command
picker, the command result, and the fleet grid. Type to narrow, `esc` to clear.
This is one shared idiom (codename **Himinbjörg**) — learn it once.

**Your selection stays in view.** In any list — command picker, log sources, sort —
the highlighted row is kept on screen as you move, so a long list never scrolls the
cursor out from under you.

## Mouse

- **Wheel** scrolls whatever is under the pointer — the grid, a modal body, or a
  panel in the top view.
- **Click** in the top view focuses the panel under the pointer.

Everything the mouse does has a keyboard equivalent, so Heimdall is fully usable over
a plain SSH session or a screen reader that scrapes the terminal.

## Full key reference

The in-app help (`?`) is the canonical, always-current list:

| Context | Keys |
|---|---|
| **Fleet** | `↑/↓·j/k` move · `⏎` detail · `t` top · `g` group · `/` filter · `r` refresh |
| **Host detail** | `↑/↓` prev/next host · `⇧↑/↓·wheel` scroll · `l/p/c` logs/procs/command · `t` top |
| **Modals** | `↑/↓·wheel` pick/scroll · `⏎` open/run · `/` search/filter · `s` sort · `esc` back one level |
| **Top view** | `↑/↓·wheel` scroll (whole view first) · `tab` focus a panel · `pgup/pgdn` page · `esc` back |
| **General** | `?` help · `q·ctrl+c` quit · `esc` back one level |

## Learn with no setup

`heimdall-dashboard --demo` renders a fake fleet with live-looking metrics, logs,
and processes — every key above works, so it is the safe place to build muscle
memory. See [Demo Mode](08-demo-mode.md).
