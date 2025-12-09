# VHS Tape for git-scope
# To run this: vhs demo.tape
# Install vhs: brew install vhs

Output demo.gif

Set FontSize 16
Set Width 1200
Set Height 800
Set Padding 20
Set FontFamily "JetBrains Mono"

# Setup the terminal look
Set Theme "Catppuccin Mocha"

# Hide the prompt setup
Hide
Type "export PS1='> '"
Enter
Ctrl+L
Show

# Start the demo with version check
Type "git-scope --version"
Sleep 500ms
Enter
Sleep 1s

# Run the TUI
Type "git-scope"
Sleep 1s
Enter
Sleep 2s

# Navigate a bit
Type "j"
Sleep 500ms
Type "j"
Sleep 500ms
Type "k"
Sleep 500ms

# Show searching
Type "/"
Sleep 500ms
Type "git"
Sleep 300ms
Type "-"
Sleep 300ms
Type "scope"
Sleep 1s
Enter
Sleep 2s

# Show filtering
Type "c"
Sleep 1s
Type "f"
Sleep 1s
Type "f"
Sleep 1s
Type "f"
Sleep 1s

# Show sorting
Type "s"
Sleep 1s
Type "s"
Sleep 1s

# Quit
Type "q"
Sleep 1s
