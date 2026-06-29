# VHS tape — the in-dashboard process view (t), with the sort picker (s).
# Render: vhs docs/screenshots/tapes/top-modal.tape  (writes gif/top-modal.gif)
# Requires: vhs (github.com/charmbracelet/vhs). Driven against the --demo fleet.
Output gif/top-modal.gif
Set Theme "Catppuccin Mocha"
Set FontSize 16
Set Width 1100
Set Height 640
Set Padding 20

Hide
Type "bin/heimdall-dashboard --demo"
Enter
Sleep 2s
Show

Enter@500ms        # open the host detail view
Sleep 1s
Type "t"           # open the top (process) modal — CPU-desc by default
Sleep 1500ms
Type "s"           # open the sort picker
Sleep 800ms
Down@400ms         # choose a different sort key (persisted to config)
Down@400ms
Enter
Sleep 2s
Escape@500ms       # back to detail
Escape@500ms       # back to grid
Sleep 1s
