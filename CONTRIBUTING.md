# Contributing to git-scope

First off, thanks for taking the time to contribute! 🎉

## How Can I Contribute?

### Reporting Bugs

- **Check existing issues** first to avoid duplicates.
- Use a **clear, descriptive title**.
- Include steps to reproduce, expected vs actual behavior, and your OS/Go version.

### Suggesting Features

Open an issue with the `enhancement` label. Describe:
- The problem you're trying to solve
- Your proposed solution
- Any alternatives you've considered

### Pull Requests

1. **Fork** the repo and create your branch from `main`.
2. Run `go build ./...` and `go test ./...` to ensure everything works.
3. Keep PRs focused—one feature or fix per PR.
4. Update documentation if needed.
5. Write a clear PR description.

## Development Setup

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/git-scope.git
cd git-scope

# Build
go build -o git-scope ./cmd/git-scope

# Run
./git-scope
```

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`).
- Keep functions small and focused.
- Add comments for non-obvious logic.

## Questions?

Open a [Discussion](https://github.com/Bharath-code/git-scope/discussions) or reach out on [Twitter/X](https://x.com/iam_pbk).

Thanks again! 🙏
