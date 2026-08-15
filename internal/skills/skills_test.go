package skills

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestSkillsManager(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("ANDROID_MCP_HOME", tempHome)

	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("failed to create skills manager: %v", err)
	}

	if mgr.IsInstalled() {
		t.Fatalf("expected skills not installed initially")
	}

	ctx := context.Background()
	if err := mgr.EnsureSkills(ctx); err != nil {
		t.Fatalf("EnsureSkills failed: %v", err)
	}

	if !mgr.IsInstalled() {
		t.Fatalf("expected skills to be installed after EnsureSkills")
	}

	listOut := mgr.ListSkills()
	if !strings.Contains(listOut, "Android-MCP-go Installed Skills") {
		t.Fatalf("unexpected list skills output: %s", listOut)
	}
	if !strings.Contains(listOut, filepath.Join(tempHome, "skills")) {
		t.Fatalf("missing target skills directory in list output: %s", listOut)
	}
}
