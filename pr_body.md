## What changed
- Added a `git-scope.json` Scoop manifest file for Windows installation support, with architecture entries for both 64-bit (amd64) and ARM64, `checkver` for version detection, and `autoupdate` for automatic URL updates on new releases.
- Updated `README.md` to document the Scoop installation method alongside existing Homebrew and source options.
- Added a test to validate the Scoop manifest structure.

## Why
Windows developers currently have to build from source to install git-scope. Adding a Scoop manifest makes installation as simple as `scoop install git-scope`.

Fixes #19

## Testing
- Verified the manifest is valid JSON with all required Scoop fields (`version`, `description`, `homepage`, `license`, `architecture`, `bin`, `checkver`, `autoupdate`).
- Confirmed architecture entries include both `64bit` and `arm64` with correct URL patterns matching the goreleaser archive naming convention.
- Ran `go test ./...` — all tests pass.
- Added `git-scope_test.go` with a test that validates the manifest structure.
