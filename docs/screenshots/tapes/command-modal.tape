# VHS tape — the on-demand command modal (c), gated on the _cmd capability.
# Render: vhs docs/screenshots/tapes/command-modal.tape (writes gif/command-modal.gif)
# Requires: vhs (github.com/charmbracelet/vhs). Driven against the --demo fleet,
# which advertises a command-enabled host so the `c` affordance is present.
Output gif/command-modal.gif
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
Type "c"           # open the command modal (allow-listed, read-only)
Sleep 1500ms
Down@400ms         # pick a command from the allow-list
Enter
Sleep 2500ms       # result returns inline, correlated by request id
Escape@500ms       # back to detail
Escape@500ms       # back to grid
Sleep 1s
