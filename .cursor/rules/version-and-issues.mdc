---
description: Version & issue tracking — every change gets a SemVer class, an issue with the matching type:* label, a major/minor milestone, and a card on the configured project board (project.config.yaml)
alwaysApply: true
---

## Version & Issue Tracking (mandatory)

Every change we work on is tracked. When you start a piece of work (and before you
finish it):

1. **Classify the SemVer bump** — **major** (breaking), **minor** (new
   backward-compatible feature), or **patch** (bug fix / no API change).
2. **Milestone — major & minor only.** Ensure a milestone `vX.Y.0` exists for the
   target minor/major; create it if missing. **Patches never get their own
   milestone** — attach them to their minor's milestone (`vX.Y.0`).
3. **Issue — every change, patches included.** Create a GitHub issue, labelled
   the matching **`type:*`** label (`type:bug` fix · `type:feature` feature · `type:docs`/`type:refactor`/`type:chore`, per `gh-labels-milestones`), assigned
   to the target major/minor milestone.
4. **Project board.** Add the issue to the project board — read its number from
   `github.project_number` in `project.config.yaml` (never hard-code it) — and set
   **Status** (`In progress` while working, `Done` when shipped).
5. **Link & close.** The PR references the issue (`Closes #N`); merging closes it.
   On release the milestone version equals the release tag.
6. **No Claude/Anthropic attribution** anywhere — commits, PR bodies, issues,
   releases.

Mechanics live in the maple skills — use them, don't hand-roll:
`gh-labels-milestones` (milestone/label upsert), `gh-issues` (create/label/
milestone), `gh-projects` (board add + Status). Config: `project.config.yaml → github`.
