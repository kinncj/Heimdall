---
id: high-fidelity-terminal-visual-experience-0004
story: docs/stories/high-fidelity-terminal-visual-experience-20260626121705-0004/Story.md
target: tui
status: approved
created_at: 2026-06-26
---

# Wireframe — high-fidelity-terminal-visual-experience-0004

Target: **tui**. Animated, depth/3D-style visuals that improve readability, with automatic
graceful degradation. Both render modes are shown side by side, now including the GPU/power
gradient gauges, plus the auto-detection ladder. No specific colours are chosen here.

Legend (every state = symbol + text, never colour alone):
  ● online/ok   ◐ degraded/enrolling   ○ offline/no-internet   ⏱ stale   ⚠ error
  ⚿ needs helper (insufficient permission)   — unavailable (vendor/path)   ✓ installed   – absent
  ↑ relaying   ↯ reconnecting   ⚡ rate-limited   ✋ refused   ▸ focus (also inverse)
Min width ≈ 80 cols: optional columns (TREND/GPU/PWR/NET) collapse first; core metrics never drop.

## High-fidelity vs degraded — incl. GPU/power gauges

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ Heimdall · render modes                   ● 6/7 online                 ⏱ 2026-06-26 14:03:12 ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                                ║
║   ┌ HIGH-FIDELITY  truecolor + anim ──────────┐  ┌ DEGRADED  NO_COLOR / dumb / slow ─────────┐ ║
║   │ dgx-spark        ● ONLINE  ⠹ live         │  │ dgx-spark        [ON] ONLINE  busy        │ ║
║   │ ╓─ depth panel (shadow) ──────╖           │  │ +- plain panel ---------------+           │ ║
║   │ CPU  █████████▓▒░░ 88% ↑                  │  │ CPU  [######## ] 88% ^                    │ ║
║   │ MEM  ██████▓▒░░░░░ 63% →                  │  │ MEM  [######   ] 63% =                    │ ║
║   │ STO  █████▓▒░░░░░░ 54% →                  │  │ STO  [#####    ] 54% =                    │ ║
║   │ TEMP ████████▓▒░░░ 74°C ↑                 │  │ TEMP [#######  ] 74C ^                    │ ║
║   │ GPU  ██████████▓▒░ 91% ↑                  │  │ GPU  [######## ] 91% ^                    │ ║
║   │ VRAM ███████▓▒░░░░ 28/40GB                │  │ VRAM [######   ] 28/40GB                  │ ║
║   │ PWR  ██████▓▒░░░░░ 142W                   │  │ PWR  [#####    ] 142W                     │ ║
║   │ prof balanced (read-only)                 │  │ prof balanced (read-only)                 │ ║
║   │ ╙─────────────────────────────╜           │  │ +-----------------------------+           │ ║
║   │ gradient █▓▒░ · braille spinner           │  │ ascii # = bars · no animation             │ ║
║   └───────────────────────────────────────────┘  └───────────────────────────────────────────┘ ║
║                                                                                                ║
║   Same host, two modes. Every critical field — host, CONN, CPU/MEM/STO/TEMP and the new        ║
║   GPU/VRAM/PWR gauges + read-only profile — is present and readable in BOTH.                   ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║ STATUS render=high-fidelity · v toggles mode · auto-detected                                   ║
║ q quit · ↑/↓ nav · ⏎ detail · r refresh · v fidelity · ? help                                  ║
╚════════════════════════════════════════════════════════════════════════════════════════════════╝
```

**Annotations**

- Left high-fidelity: depth panel + █▓▒░ gradient gauges (CPU/MEM/STO/TEMP/GPU/VRAM/PWR) + braille spinner ⠹.
- Right degraded: plain ASCII # / = bars, static 'busy', NO_COLOR-safe, no animation.
- Parity: host, CONN and every metric incl. GPU/VRAM/PWR and the read-only profile appear in BOTH.
- All states use symbol + text (● ONLINE / [ON] ONLINE); styling never hides monitoring data.
- v cycles fidelity; all other bindings identical between modes (keyboard-reachable).

## Auto-detection & degrade ladder

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ Heimdall · render modes                   ● 6/7 online                 ⏱ 2026-06-26 14:03:12 ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                                ║
║   ┌ AUTO-DETECTION & DEGRADE LADDER ─────────────────────────────────────────────────────┐     ║
║   │ probe                         → decision                                             │     ║
║   │ NO_COLOR set / TERM=dumb      → degraded (ascii, no colour)                          │     ║
║   │ COLORTERM=truecolor + 256col  → high-fidelity (gradient gauges + depth)              │     ║
║   │ frame budget exceeded (slow)  → drop animation first, keep glyphs                    │     ║
║   │ narrow width (<80 cols)       → drop GPU/PWR/TREND gauges, keep numbers              │     ║
║   │                                                                                      │     ║
║   │ manual override: v cycles  auto → high → degraded → auto                             │     ║
║   │ Critical data (host/CONN/CPU/MEM/STO/TEMP/GPU/PWR) renders at every rung.            │     ║
║   └──────────────────────────────────────────────────────────────────────────────────────┘     ║
║                                                                                                ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║ STATUS auto-detect: COLORTERM=truecolor → high-fidelity                                        ║
║ v cycle fidelity · q quit · ? help                                                             ║
╚════════════════════════════════════════════════════════════════════════════════════════════════╝
```

**Annotations**

- Mode is chosen from terminal probes (NO_COLOR, TERM, COLORTERM, colour count) and a frame-time budget.
- Ordered degradation: drop animation → drop gradient/depth → drop GPU/PWR/TREND gauges → keep numbers.
- Critical data renders at every rung; <80 cols collapses gauges first (min-width-resize).
- Degraded is the accessible baseline and is reachable manually with v as well as automatically.

