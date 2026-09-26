package tui

import (
	"reflect"
	"testing"
)

func TestEditorArgs(t *testing.T) {
	const p = "/repos/app"
	tests := []struct {
		name   string
		fields []string
		want   []string
	}{
		{"plain editor appends path", []string{"code"}, []string{"code", p}},
		{"flags kept before path", []string{"nvim", "-R"}, []string{"nvim", "-R", p}},
		{"lazygit uses cwd only", []string{"lazygit"}, []string{"lazygit"}},
		{"lazygit.exe on windows", []string{`C:\bin\LazyGit.exe`}, []string{`C:\bin\LazyGit.exe`}},
		{"path token substituted", []string{"lazygit", "--path", "{path}"}, []string{"lazygit", "--path", p}},
		{"path token inside arg", []string{"tool", "--dir={path}"}, []string{"tool", "--dir=" + p}},
		{"legacy trailing flag still works", []string{"lazygit", "--path"}, []string{"lazygit", "--path", p}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := editorArgs(tt.fields, p); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("editorArgs(%q) = %q, want %q", tt.fields, got, tt.want)
			}
		})
	}
}

func TestEditorCmdSetsDir(t *testing.T) {
	if c := editorCmd([]string{"vim"}, "/repos/app"); c.Dir != "/repos/app" {
		t.Errorf("Dir = %q, want repo path", c.Dir)
	}
}
