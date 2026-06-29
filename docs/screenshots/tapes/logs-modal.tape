# VHS tape — the in-dashboard log modal (l), with search (/).
# Render: vhs docs/screenshots/tapes/logs-modal.tape  (writes gif/logs-modal.gif)
# Requires: vhs (github.com/charmbracelet/vhs). Driven against the --demo fleet,
# so it needs no hub.
Output gif/logs-modal.gif
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
Type "l"           # open the logs modal
Sleep 1500ms
Type "/"           # search within the buffered + live lines
Sleep 500ms
Type "error"
Sleep 2s
Escape@500ms       # clear search
Escape@500ms       # back to detail
Escape@500ms       # back to grid
Sleep 1s
