package args

import (
    "path/filepath"
    "strings"

    "iasi-dev/internal/cli"
    "iasi-dev/internal/structures"
)

// EmitTargetDiagnostics reports discovery diagnostics belonging to a target.
// A missing configuration under a known target is attributed to that target.
func EmitTargetDiagnostics(Context structures.Context, target string) {
    target = filepath.Clean(target)
    for i := range Context.Diagnostics {
        d := &Context.Diagnostics[i]
        if d.Emitted { continue }
        relative, err := filepath.Rel(target, d.Path)
        if err != nil { continue }
        if relative != "." && (relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator))) { continue }
        cli.Attention(Context, "%s", d.Message)
        d.Emitted = true
    }
}

// EmitRemainingDiagnostics reports issues not reached by a target runner.
func EmitRemainingDiagnostics(Context structures.Context) {
    for i := range Context.Diagnostics {
        d := &Context.Diagnostics[i]
        if d.Emitted { continue }
        cli.Attention(Context, "%s", d.Message)
        d.Emitted = true
    }
}
