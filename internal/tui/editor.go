package tui

import (
	"os/exec"
	"strings"
)

// cwdOnlyEditors reject a positional repo path (lazygit parses it as a
// subcommand, #23) but open whatever repo they are started in.
var cwdOnlyEditors = map[string]bool{"lazygit": true, "gitui": true, "tig": true}

// editorArgs builds argv for opening repoPath. A {path} token in the config
// is substituted; otherwise the path is appended unless the tool is cwd-only.
func editorArgs(fields []string, repoPath string) []string {
	args := make([]string, 0, len(fields)+1)
	substituted := false
	for _, f := range fields {
		if strings.Contains(f, "{path}") {
			f = strings.ReplaceAll(f, "{path}", repoPath)
			substituted = true
		}
		args = append(args, f)
	}
	name := strings.TrimSuffix(strings.ToLower(fields[0][strings.LastIndexAny(fields[0], `/\`)+1:]), ".exe")
	// ponytail: cwd-only applies to the bare command, so existing
	// "lazygit --path" configs keep getting the path appended.
	cwdOnly := len(fields) == 1 && cwdOnlyEditors[name]
	if !substituted && !cwdOnly {
		args = append(args, repoPath)
	}
	return args
}

func editorCmd(fields []string, repoPath string) *exec.Cmd {
	args := editorArgs(fields, repoPath)
	c := exec.Command(args[0], args[1:]...)
	c.Dir = repoPath
	return c
}
