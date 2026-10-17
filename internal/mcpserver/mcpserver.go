// Package mcpserver exposes git-scope's repo state to AI coding agents over a
// stdio MCP server. Every tool is read-only and reuses the same scan → report
// path as `git-scope status --json`; nothing here touches the network.
package mcpserver

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/Bharath-code/git-scope/internal/attention"
	"github.com/Bharath-code/git-scope/internal/report"
	"github.com/Bharath-code/git-scope/internal/scan"
)

// Repos is the tool output wrapper; MCP clients expect structured content to
// be a JSON object, so the array is nested under "repos".
type Repos struct {
	Repos []report.Repo `json:"repos"`
}

// PathInput selects one repository for repo_status.
type PathInput struct {
	Path string `json:"path" jsonschema:"absolute path of a repository under the configured roots"`
}

// Run serves the MCP tools on stdin/stdout until the client disconnects.
func Run(ctx context.Context, roots, ignore []string, version string) error {
	return New(roots, ignore, version).Run(ctx, &mcp.StdioTransport{})
}

// New builds the server; split from Run so tests can use an in-memory transport.
func New(roots, ignore []string, version string) *mcp.Server {
	statusFn := func() ([]report.Repo, error) {
		repos, err := scan.ScanRoots(roots, ignore)
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		return report.Build(repos, time.Now()), nil
	}

	s := mcp.NewServer(&mcp.Implementation{Name: "git-scope", Version: version}, nil)
	readOnly := &mcp.ToolAnnotations{ReadOnlyHint: true, OpenWorldHint: new(bool)}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_repos",
		Description: "List every git repository under the configured roots with branch, uncommitted file count, ahead/behind, unpushed commits, last commit time and an attention tier (ACTION/WATCH/CLEAN). Sorted most urgent first.",
		Annotations: readOnly,
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, Repos, error) {
		all, err := statusFn()
		return nil, Repos{Repos: all}, err
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "needs_attention",
		Description: "List only repositories with uncommitted, unpushed or out-of-sync work (tier ACTION or WATCH). Use this before ending a session to check nothing is left behind.",
		Annotations: readOnly,
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, Repos, error) {
		all, err := statusFn()
		if err != nil {
			return nil, Repos{}, err
		}
		out := []report.Repo{}
		for _, r := range all {
			if r.Tier != attention.Clean.String() {
				out = append(out, r)
			}
		}
		return nil, Repos{Repos: out}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "repo_status",
		Description: "Get the status of one repository by absolute path. The path must be a repository under git-scope's configured roots.",
		Annotations: readOnly,
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in PathInput) (*mcp.CallToolResult, report.Repo, error) {
		if in.Path == "" {
			return nil, report.Repo{}, fmt.Errorf("path is required")
		}
		want := canonical(in.Path)
		all, err := statusFn()
		if err != nil {
			return nil, report.Repo{}, err
		}
		// ponytail: full rescan per call; scan only the matching root if it gets slow.
		for _, r := range all {
			if canonical(r.Path) == want {
				return nil, r, nil
			}
		}
		return nil, report.Repo{}, fmt.Errorf("%s is not a repository under the configured roots", want)
	})

	return s
}

// canonical makes agent-supplied and scanned paths comparable: it expands
// ~/, resolves relative paths and symlinks (macOS /tmp → /private/tmp), and
// strips trailing slashes.
func canonical(p string) string {
	if rest, ok := strings.CutPrefix(p, "~/"); ok {
		if home, err := os.UserHomeDir(); err == nil {
			p = filepath.Join(home, rest)
		}
	}
	if abs, err := filepath.Abs(p); err == nil {
		p = abs
	}
	if real, err := filepath.EvalSymlinks(p); err == nil {
		p = real
	}
	return filepath.Clean(p)
}
