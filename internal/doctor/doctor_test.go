package doctor

import (
	"context"
	"strings"
	"testing"
)

func TestDoctorAndStatus(t *testing.T) {
	ctx := context.Background()
	rep := RunDoctor(ctx, nil, nil, nil)
	if rep == nil {
		t.Fatalf("expected doctor report")
	}

	formatted := rep.Format()
	if !strings.Contains(formatted, "Android-MCP-Go Doctor") {
		t.Fatalf("unexpected doctor format output: %s", formatted)
	}

	statusStr, _ := RunStatus(ctx, nil)
	if !strings.Contains(statusStr, "Android-MCP") {
		t.Fatalf("unexpected status output: %s", statusStr)
	}
}
