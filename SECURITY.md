# Security Policy

## Supported Versions

Only the latest minor release receives security fixes.

| Version | Supported          |
| ------- | ------------------ |
| 1.4.x   | :white_check_mark: |
| < 1.4   | :x:                |

## Reporting a Vulnerability

**Do not open a public issue.** Report privately through either:

1. **GitHub private advisory (preferred):** [Report a vulnerability](https://github.com/Bharath-code/git-scope/security/advisories/new)
2. **Email:** kumarbharath63@gmail.com, with "git-scope security" in the subject

Please include the affected version, OS, and steps to reproduce.

What to expect:
- Acknowledgment within 48 hours
- An assessment and fix timeline within 7 days
- Credit in the release notes and advisory (unless you prefer to stay anonymous)

## Scope

git-scope promises to be **read-only, local-first, and safe at work**. These are in scope:

- Any way scanning a repository can **modify it** (git-scope must never write to a repo)
- Any way a malicious repository can make git-scope **execute code**, for example through
  repo-local git config such as `core.fsmonitor` (see `internal/gitstatus`)
- Any **network call or data leaving the machine**, from the CLI, TUI, or `git-scope mcp`
- The MCP server exposing repositories **outside the configured roots**
- Config, cache, or state files written with permissions readable by other users

## Verifying Releases

Release archives ship with `checksums.txt`, an SPDX SBOM per archive, and signed
build provenance. Verify an archive with:

```bash
gh attestation verify git-scope_<version>_<os>_<arch>.tar.gz --repo Bharath-code/git-scope
```
