@story:fullscreen-top-view-0025 @priority:high @ui
Feature: Full-screen single-host top view

  # --- Keybind swap (p = processes, t = top) ---

  @story @priority:high
  Scenario: Pressing p opens the renamed process table
    Given the dashboard is showing the fleet grid
    When the operator presses "p"
    Then the process table opens labelled "processes"

  @story @priority:high
  Scenario: Pressing t opens the full-screen top view
    Given the dashboard has a focused host
    When the operator presses "t"
    Then the full-screen single-host top view opens for that host

  @story @priority:high
  Scenario: Pressing esc or q exits the top view
    Given the operator is in the full-screen top view
    When the operator presses "esc" or "q"
    Then the dashboard returns to the fleet grid

  # --- Panel content (reads existing collected data) ---

  @story @priority:high
  Scenario: The top view shows every system panel for the focused host
    Given the operator opens the top view for a focused host
    Then it shows per-core CPU bars, a CPU utilisation sparkline, and CPU frequency
    And it shows memory used and swap with a memory-bandwidth sparkline
    And it shows power for package, cpu, gpu, and npu with a power sparkline
    And it shows GPU and NPU utilisation, VRAM, and temperature
    And it shows network and disk sparklines
    And it shows the process list

  @story @priority:high
  Scenario: Time-series graphs are drawn from the existing history buffers
    Given the focused host has recorded metric history
    When the top view renders its braille sparklines
    Then the sparklines are drawn from the existing per-host history buffers, not a new collector

  @story @priority:high
  Scenario: The accelerator panel uses NPU terminology instead of ANE
    Given the operator opens the top view
    When the accelerator panel renders
    Then it is labelled "NPU" rather than "ANE"

  # --- Screen-size awareness (reuses the existing responsive pattern) ---

  @story @priority:high
  Scenario: A wide terminal shows the two-column panel grid with full graphs
    Given a focused host shown in the top view
    When the terminal is at least 100 columns wide
    Then panels render in a two-column grid with full sparklines and a multi-column per-core matrix

  @story @priority:high
  Scenario: A medium terminal stacks panels in a single column
    Given a focused host shown in the top view
    When the terminal is between 60 and 99 columns wide
    Then panels stack in a single column, sparklines are kept, and per-core bars wrap to fewer columns

  @story @priority:high
  Scenario: A narrow terminal collapses per-core to an aggregate
    Given a focused host shown in the top view
    When the terminal is between 40 and 59 columns wide
    Then panels stack, sparklines are shortened, and per-core collapses to an aggregate bar with a core-count summary

  @story @priority:high
  Scenario: A tiny terminal shows key numbers only
    Given a focused host shown in the top view
    When the terminal is narrower than 40 columns
    Then only key numbers (cpu%, mem%, power, temp) are shown, one value per line, with no graphs and no per-core bars

  @story @priority:high
  Scenario: No rendered line exceeds the terminal width
    Given a focused host shown in the top view at any width tier
    When the top view renders
    Then every rendered line fits within the terminal width

  @story @priority:high
  Scenario: Content taller than the terminal scrolls within a fixed header and footer
    Given the top view content is taller than the terminal height
    When the operator scrolls with the up and down keys
    Then the header and footer stay fixed, the body scrolls, and the frame never exceeds the terminal height

  @story @priority:high
  Scenario: The top view stays usable at iPhone-portrait width in Termius
    Given the operator is connected over SSH from Termius on a phone in portrait
    When the top view renders at a width narrower than 40 columns
    Then it shows the key-numbers-only layout with no clipped lines

  # --- Graceful degradation: unavailable metrics render a dash ---

  @story @priority:high
  Scenario Outline: A metric the platform cannot supply renders as a dash
    Given a "<platform>" host focused in the top view
    When the top view renders "<metric>"
    Then it shows "—" rather than an error or a fabricated 0

    Examples:
      | platform | metric    |
      | windows  | cpu.load  |
      | linux    | mem.bw    |
      | linux    | npu.util  |
      | windows  | cpu.freq  |

  @story @priority:high
  Scenario Outline: A metric the platform supports renders a real value
    Given a "<platform>" host focused in the top view
    When the top view renders "<metric>"
    Then it shows a real value rather than a dash

    Examples:
      | platform | metric    |
      | macos    | cpu.load  |
      | linux    | mem.swap  |
      | macos    | mem.bw    |
      | macos    | npu.util  |

  # --- NPU rename, backward compatible ---

  @story @priority:high
  Scenario: The legacy power.ane key is normalised to power.npu at ingest
    Given a daemon that reports accelerator power under the legacy key "power.ane"
    When the dashboard ingests that metric
    Then it is stored under the canonical key "power.npu"
    And the top view renders it under the NPU label

  @story @priority:high
  Scenario: A canonical power.npu key is rendered directly
    Given a daemon that reports accelerator power under "power.npu"
    When the dashboard ingests that metric
    Then it is rendered under the NPU label without further translation

  # --- Backward compatibility in a mixed fleet ---

  @story @priority:high
  Scenario: The top view works against an older daemon that lacks the new keys
    Given a mixed fleet where the focused host runs an older daemon without the new metric keys
    When the operator opens the top view for that host
    Then the panels render, the missing metrics show "—", and no error is raised

  @story @priority:high
  Scenario: New metric keys are additive on the wire
    Given a v2.2 daemon emitting the new metric keys
    When an older hub or dashboard receives the stream
    Then it ignores the unknown keys and keeps working
