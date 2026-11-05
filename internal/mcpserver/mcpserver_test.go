package mcpserver

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestTools(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	root := t.TempDir()
	for _, name := range []string{"clean", "dirty"} {
		dir := filepath.Join(root, name)
		if out, err := exec.Command("git", "init", "-q", dir).CombinedOutput(); err != nil {
			t.Fatalf("git init: %v\n%s", err, out)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "dirty", "f"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(filepath.Join(root, "dirty"), link); err != nil {
		// Windows needs developer mode / admin for symlinks; test the plain path.
		link = filepath.Join(root, "dirty")
	}

	ctx := context.Background()
	st, ct := mcp.NewInMemoryTransports()
	if _, err := New([]string{root}, nil, "test").Connect(ctx, st, nil); err != nil {
		t.Fatal(err)
	}
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "t"}, nil).Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cs.Close()

	call := func(name string, args map[string]any) (*mcp.CallToolResult, map[string]any) {
		t.Helper()
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
		if err != nil {
			t.Fatal(err)
		}
		var out map[string]any
		b, _ := json.Marshal(res.StructuredContent)
		_ = json.Unmarshal(b, &out)
		return res, out
	}

	tests := []struct {
		name    string
		tool    string
		args    map[string]any
		check   func(map[string]any) bool
		wantErr bool
	}{
		{"list_repos returns all", "list_repos", nil, func(o map[string]any) bool { return len(o["repos"].([]any)) == 2 }, false},
		{"needs_attention drops clean", "needs_attention", nil, func(o map[string]any) bool {
			r := o["repos"].([]any)
			return len(r) == 1 && r[0].(map[string]any)["name"] == "dirty"
		}, false},
		{"repo_status via symlink", "repo_status", map[string]any{"path": link + "/"}, func(o map[string]any) bool { return o["tier"] == "WATCH" }, false},
		{"repo_status outside roots", "repo_status", map[string]any{"path": t.TempDir()}, nil, true},
		{"repo_status empty path", "repo_status", map[string]any{"path": ""}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, out := call(tt.tool, tt.args)
			if res.IsError != tt.wantErr {
				t.Fatalf("IsError = %v, want %v: %+v", res.IsError, tt.wantErr, res.Content)
			}
			if tt.check != nil && !tt.check(out) {
				t.Errorf("unexpected output: %v", out)
			}
		})
	}
}
