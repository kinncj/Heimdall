# Dashboard screenshots

Documentation captures of the `heimdall-dashboard` TUI, generated from the
synthetic `--demo` fleet so they need no hub and stay reproducible.

## Regenerate

```bash
make screenshots          # or: bash scripts/gen-screenshots.sh
```

The generator drives `heimdall-dashboard --demo --snapshot` at a matrix of
sizes, views, and themes. The snapshot path honours `COLUMNS`/`LINES`, so the
same binary renders a wide grid (all metric columns) and a narrow / portrait grid
(columns dropped right-to-left) — the responsive layout, captured headlessly.

It runs with an isolated `HEIMDALL_CONFIG_DIR`, so it never rewrites your real
`~/.config/heimdall/dashboard.json`.

## Outputs

Each step degrades gracefully — absent tools are skipped, so the script always
produces the lower-fidelity artifacts.

| Dir      | Tool        | What                                                       |
|----------|-------------|------------------------------------------------------------|
| `ansi/`  | none        | Raw ANSI frames — always produced; diff-able, re-renderable |
| `html/`  | `aha`       | Styled, embeddable HTML renders of each frame              |
| `gif/`   | `vhs`       | Animated GIFs of the interactive modal flows              |

These are build artifacts — **git-ignored**, not committed. The committed sources
are this README, the generator (`scripts/gen-screenshots.sh`), and the VHS tapes
under `tapes/`.

## Static frames

| Name            | Size    | View   | Shows                                  |
|-----------------|---------|--------|----------------------------------------|
| `grid-wide`     | 110×30  | grid   | Full column set                        |
| `grid-narrow`   | 64×24   | grid   | Responsive column drop (portrait)      |
| `grid-light`    | 110×30  | grid   | Light theme                            |
| `detail-wide`   | 110×34  | detail | Host drilldown                         |
| `detail-small`  | 88×20   | detail | Scrollable detail on a small screen    |

## Animated flows (`tapes/`)

The interactive modals can't be captured from a single snapshot, so they are
driven by [VHS](https://github.com/charmbracelet/vhs) tapes — the committed,
reproducible source. Install `vhs` and re-run `make screenshots` to render them.

| Tape                 | Flow                                                  |
|----------------------|-------------------------------------------------------|
| `logs-modal.tape`    | `l` logs modal, then `/` search                       |
| `top-modal.tape`     | `t` process view, then `s` sort picker (persisted)    |
| `command-modal.tape` | `c` on-demand command modal (allow-listed, read-only) |
