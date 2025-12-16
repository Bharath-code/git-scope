package nudge

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Version is the current app version - used to track per-version nudge
const Version = "1.3.0"

// GitHubRepoURL is the URL to open when user presses S
const GitHubRepoURL = "https://github.com/Bharath-code/git-scope"

// NudgeState represents the persistent state of the star nudge
type NudgeState struct {
	SeenVersion string `json:"seenVersion"`
	Dismissed   bool   `json:"dismissed"`
	Completed   bool   `json:"completed"`
}

// getNudgePath returns the path to the nudge state file
func getNudgePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".cache", "git-scope", "nudge.json")
}

// loadState loads the nudge state from disk
func loadState() *NudgeState {
	path := getNudgePath()
	if path == "" {
		return &NudgeState{}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return &NudgeState{}
	}

	var state NudgeState
	if err := json.Unmarshal(data, &state); err != nil {
		return &NudgeState{}
	}

	return &state
}

// saveState saves the nudge state to disk
func saveState(state *NudgeState) error {
	path := getNudgePath()
	if path == "" {
		return nil
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// ShouldShowNudge checks if the star nudge should be shown
// Returns true only if:
// - Not already seen for this version
// - Not dismissed
// - Not completed (user already starred)
func ShouldShowNudge() bool {
	state := loadState()

	// Already seen for this version
	if state.SeenVersion == Version {
		return false
	}

	// User already completed (pressed S)
	if state.Completed {
		return false
	}

	return true
}

// MarkShown marks the nudge as shown for the current version
func MarkShown() {
	state := loadState()
	state.SeenVersion = Version
	state.Dismissed = false
	_ = saveState(state)
}

// MarkDismissed marks the nudge as dismissed (any key pressed)
func MarkDismissed() {
	state := loadState()
	state.SeenVersion = Version
	state.Dismissed = true
	_ = saveState(state)
}

// MarkCompleted marks the nudge as completed (S pressed, GitHub opened)
func MarkCompleted() {
	state := loadState()
	state.SeenVersion = Version
	state.Completed = true
	_ = saveState(state)
}
