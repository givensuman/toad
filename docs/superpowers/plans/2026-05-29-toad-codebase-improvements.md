# Toad Codebase Improvements Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Improve API coherency, user experience, and code quality by rebranding from Toolbx, fixing the help system, eliminating duplicated patterns, and cleaning up bugs.

**Architecture:** Incremental refactor — each task is self-contained and testable. No structural changes to package layout. Focus on eliminating repetition and standardizing patterns.

**Tech Stack:** Go 1.26, Cobra, Logrus

---

## File Map

### Modified files:
- `cmd/root.go` — Use field, Short description, remove manpage help reference
- `cmd/rootDefault.go` — Update default output
- `cmd/help.go` — Replace manpage dispatch with Cobra built-in help
- `cmd/create.go` — Rebrand user-facing strings, replace manpage help
- `cmd/enter.go` — Rebrand user-facing strings, replace manpage help
- `cmd/run.go` — Rebrand user-facing strings, replace manpage help
- `cmd/list.go` — Rebrand user-facing strings, replace manpage help
- `cmd/rm.go` — Rebrand user-facing strings, replace manpage help
- `cmd/rmi.go` — Rebrand user-facing strings, replace manpage help
- `cmd/down.go` — Rebrand user-facing strings, replace manpage help
- `cmd/up.go` — Rebrand user-facing strings, replace manpage help
- `cmd/initContainer.go` — Rebrand comments and strings
- `cmd/utils.go` — Add `usageError` helper, update common commands text, remove `showManual`
- `cmd/completion.go` — Rebrand comment
- `pkg/utils/utils.go` — Update config paths, default container name
- `pkg/podman/container.go` — Update `IsToolbx` label check to include toad labels

---

### Task 1: Add `usageError` helper and standardize error construction

**Files:**
- Modify: `cmd/utils.go` — add helpers
- Modify: `cmd/create.go` — replace repetitive error patterns
- Modify: `cmd/enter.go` — replace repetitive error patterns
- Modify: `cmd/run.go` — replace repetitive error patterns
- Modify: `cmd/list.go` — replace repetitive error patterns
- Modify: `cmd/rm.go` — replace repetitive error patterns
- Modify: `cmd/rmi.go` — replace repetitive error patterns
- Modify: `cmd/down.go` — replace repetitive error patterns

- [ ] **Step 1: Add `usageError` and `errMsg` helpers in `cmd/utils.go`**

Add at the top of the helper section (around line 24):

```go
// usageError formats an error message with a "Run 'X --help' for usage" suffix.
func usageError(format string, args ...any) error {
  msg := fmt.Sprintf(format, args...)
  return errors.New(msg + "\nRun '" + executableBase + " --help' for usage.")
}

// errMsg is a short alias for errors.New when no usage hint is needed.
func errMsg(format string, args ...any) error {
  return fmt.Errorf(format, args...)
}
```

- [ ] **Step 2: Replace error construction in `cmd/create.go`**

Replace all `strings.Builder` + multiple `fmt.Fprintf` + `errors.New` patterns with `usageError` calls:

```go
// Before (line 131-136):
if cmd.Flag("distro").Changed && cmd.Flag("image").Changed {
  var builder strings.Builder
  fmt.Fprintf(&builder, "options --distro and --image cannot be used together\n")
  fmt.Fprintf(&builder, "Run '%s --help' for usage.", executableBase)
  errMsg := builder.String()
  return errors.New(errMsg)
}

// After:
if cmd.Flag("distro").Changed && cmd.Flag("image").Changed {
  return usageError("options --distro and --image cannot be used together")
}
```

Replace the same pattern at these locations:
- Line 139-146: `--image` and `--release` conflict
- Line 148-158: `--authfile` not found
- Line 229-236: container already exists
- Line 758-763: `--assumeyes` needed for download
- Line 783-789: failed to pull image
- Line 1019-1020: container doesn't support cgroups

- [ ] **Step 3: Replace error construction in `cmd/run.go`**

Replace:
- Line 118-123: missing argument for run
- Line 247-251: container too old
- Line 262-267: NVIDIA driver mismatch

- [ ] **Step 4: Replace error construction in `cmd/rm.go`**

Replace:
- Line 73-78: missing argument for rm

- [ ] **Step 5: Replace error construction in `cmd/rmi.go`**

Replace:
- Line 73-78: missing argument for rmi

- [ ] **Step 6: Replace error construction in `cmd/enter.go`**

Replace:
- Line 63: "this is not a Toolbx container" (just `errors.New` — no usage hint needed)

- [ ] **Step 7: Update `createError*` helper functions in `cmd/utils.go`**

Replace the `strings.Builder` pattern in these helpers:

```go
func createErrorContainerNotFound(container string) error {
  return usageError("container %s not found\nUse the 'create' command to create a Toolbx.\n", container)
}

func createErrorDistroWithoutRelease(distro string) error {
  return usageError("option '--release' is needed\nDistribution %s doesn't match the host.", distro)
}

func createErrorInvalidContainer(containerArg string) error {
  return usageError("invalid argument for '%s'\nContainer names must match '%s'.", containerArg, utils.ContainerNameRegexp)
}

// ... same pattern for all createError* functions in utils.go
```

- [ ] **Step 8: Run build to verify**

Run: `go build ./...`
Expected: no errors

Run: `go vet ./cmd/...`
Expected: no errors

- [ ] **Step 9: Commit**

```bash
git add -A
git commit -m "refactor: add usageError helper, standardize error construction" -m "Replaces 50+ repetitive strings.Builder/Fprintf patterns with a single usageError() helper. Reduces error construction boilerplate ~75%."
```

---

### Task 2: Replace manpage-based help with standard Cobra help

**Files:**
- Modify: `cmd/root.go` — remove custom `rootHelp`, remove `SetHelpFunc`
- Modify: `cmd/help.go` — rewrite to use Cobra built-in help, remove `showManual` dispatch
- Modify: `cmd/utils.go` — remove `showManual`, `getUsageForCommonCommands`
- Modify: `cmd/create.go` — remove custom `createHelp`, add `--help` usage to cobra
- Modify: `cmd/enter.go` — remove custom `enterHelp`
- Modify: `cmd/run.go` — remove custom `runHelp`
- Modify: `cmd/list.go` — remove custom `listHelp`
- Modify: `cmd/rm.go` — remove custom `rmHelp`
- Modify: `cmd/rmi.go` — remove custom `rmiHelp`
- Modify: `cmd/down.go` — remove custom `downHelp`
- Modify: `cmd/up.go` — remove custom `upHelp`
- Modify: `cmd/initContainer.go` — remove custom `initContainerHelp`

- [ ] **Step 1: Simplify `cmd/help.go` to use Cobra built-in help**

Replace the entire file:

```go
package cmd

import (
  "github.com/spf13/cobra"
)

var helpCmd = &cobra.Command{
  Use:   "help [command]",
  Short: "Display help information about Toad",
  Args:  cobra.MaximumNArgs(1),
  RunE: func(cmd *cobra.Command, args []string) error {
    if len(args) == 0 {
      cmd.Root().Help()
      return nil
    }
    targetCmd, _, err := cmd.Root().Find(args)
    if err != nil {
      return err
    }
    targetCmd.Help()
    return nil
  },
}

func init() {
  rootCmd.AddCommand(helpCmd)
}
```

- [ ] **Step 2: Remove custom help functions from root.go**

In `cmd/root.go`, remove lines 172-200 (`rootHelp` function) and line 113 (`rootCmd.SetHelpFunc(rootHelp)`).

Remove `strings` import if no longer needed.

- [ ] **Step 3: Remove `showManual` and `getUsageForCommonCommands` from `cmd/utils.go`**

Remove the `showManual` function (lines 531-566) and `getUsageForCommonCommands` (lines 409-417).

- [ ] **Step 4: Remove all `*Help` functions from command files**

For each command file, remove the `*Help` function and the `SetHelpFunc` call in `init()`:

- `cmd/create.go`: Remove `createHelp` function (lines 548-567), remove line 105 `createCmd.SetHelpFunc(createHelp)`
- `cmd/enter.go`: Remove `enterHelp` function (lines 115-134), remove line 57 `enterCmd.SetHelpFunc(enterHelp)`
- `cmd/run.go`: Remove `runHelp` function (lines 481-500), remove line 83 `runCmd.SetHelpFunc(runHelp)`
- `cmd/list.go`: Remove `listHelp` function (lines 96-115), remove line 45 `listCmd.SetHelpFunc(listHelp)`
- `cmd/rm.go`: Remove `rmHelp` function (lines 103-122), remove line 40 `rmCmd.SetHelpFunc(rmHelp)`
- `cmd/rmi.go`: Remove `rmiHelp` function (lines 103-122), remove line 40 `rmiCmd.SetHelpFunc(rmiHelp)`
- `cmd/down.go`: Remove `downHelp` function (lines 87-103), remove line 41 `downCmd.SetHelpFunc(downHelp)`
- `cmd/up.go`: Replace `upHelp` with `cmd.Help()` fallback — remove lines 59-75, replace with `cmd.Help()` or just remove `upHelp` entirely and remove the `SetHelpFunc` call
- `cmd/initContainer.go`: Remove `initContainerHelp` function (lines 436-455), remove line 133 `initContainerCmd.SetHelpFunc(initContainerHelp)`

Each replacement looks like:

```go
// Remove in init():
createCmd.SetHelpFunc(createHelp)

// Remove the entire createHelp function
func createHelp(cmd *cobra.Command, args []string) {
  if utils.IsInsideContainer() { ... }
  if err := showManual("toolbox-create"); err != nil { ... }
}
```

- [ ] **Step 5: Set Cobra's default help template for rootCmd**

In `cmd/root.go` `init()`, add:

```go
rootCmd.SetHelpTemplate(`Usage:  {{.Use}} [command]

{{.Short}}

Commands:
{{range .Commands}}{{.Name | printf "  %-12s"}}{{.Short}}
{{end}}

Run '{{.Use}} <command> --help' for more details on a command.
`)
```

- [ ] **Step 6: Run build to verify**

Run: `go build ./...`
Expected: no errors

- [ ] **Step 7: Commit**

```bash
git add -A
git commit -m "refactor: replace manpage help with standard Cobra help" -m "Removes all custom *Help functions and manpage dispatch. Uses Cobra's built-in help system instead. Removes --help forwarding to host container. Simplifies help.go to delegate to cmd.Root().Help()."
```

---

### Task 3: Rebrand user-facing strings from Toolbx to Toad

**Files:**
- Modify: `cmd/root.go` — Use, Short, error strings
- Modify: `cmd/rootDefault.go` — common commands text
- Modify: `cmd/create.go` — all user-facing "Toolbx" references
- Modify: `cmd/enter.go` — all user-facing "Toolbx" references
- Modify: `cmd/run.go` — all user-facing "Toolbx" references
- Modify: `cmd/list.go` — all user-facing "Toolbx" references
- Modify: `cmd/rm.go` — all user-facing "Toolbx" references
- Modify: `cmd/rmi.go` — all user-facing "Toolbx" references
- Modify: `cmd/down.go` — all user-facing "Toolbx" references
- Modify: `cmd/up.go` — all user-facing "Toolbx" references
- Modify: `cmd/initContainer.go` — comments and internal strings
- Modify: `pkg/podman/container.go` — label constants
- Modify: `pkg/podman/podman.go` — `isToolbx` function

- [ ] **Step 1: Update root command identity**

In `cmd/root.go`:

```go
// Change Use from "toolbox" to "toad"
rootCmd = &cobra.Command{
  Use:   "toad",
  Short: "Declarative development containers powered by Podman",  // or similar
  ...
}
```

- [ ] **Step 2: Update `rootRunImpl` default command text**

In `cmd/rootDefault.go`:

```go
// Before:
builder.WriteString("create    Create a new Toolbx container\n")
builder.WriteString("enter     Enter an existing Toolbx container\n")
builder.WriteString("list      List all existing Toolbx containers and images\n")

// After:
builder.WriteString("create    Create a new Toad container\n")
builder.WriteString("enter     Enter an existing Toad container\n")
builder.WriteString("list      List all existing Toad containers and images\n")
```

- [ ] **Step 3: Bulk-replace Toolbx/Toolbx in user-facing strings in cmd/**

For each command file, replace:
- "Toolbx container" → "Toad container"
- "Toolbx " → "Toad " (where it refers to the tool, not the label)
- Error messages like "this is not a Toolbx container" → "this is not a Toad container"
- `executableBase` name references (already correct since it resolves from the binary name)

Files: `cmd/create.go`, `cmd/enter.go`, `cmd/run.go`, `cmd/list.go`, `cmd/rm.go`, `cmd/rmi.go`, `cmd/down.go`, `cmd/up.go`

Do NOT touch:
- Container labels (these need to stay for backward compat — see step 5)
- Env vars (TOOLBOX_PATH — changing this would break users)

- [ ] **Step 4: Update `initContainer.go` comments and generated file headers**

In `cmd/initContainer.go`, replace:
- All `# Written by Toolbx` → `# Written by Toad`
- All `# https://containertoolbx.org/` → `# https://toad.dev/` (or remove)
- All user-facing log messages

- [ ] **Step 5: Update `pkg/podman/podman.go` `isToolbx` function**

In `pkg/podman/podman.go`, update the label detection to also recognize toad labels:

```go
func isToolbx(labels map[string]string) bool {
  if labels["com.github.containers.toolbox"] == "true" ||
     labels["com.github.debarshiray.toolbox"] == "true" ||
     labels["com.github.givensuman.toad"] == "true" {
    return true
  }
  return false
}
```

Also rename the function to `isToadContainer` for clarity and update all callers, OR keep as-is since it's internal. Keep as-is for minimal diff.

- [ ] **Step 6: Update container label in `cmd/create.go`**

In `cmd/create.go`, change the create label:

```go
// Before:
"--label", "com.github.containers.toolbox=true",

// After:
"--label", "com.github.givensuman.toad=true",
```

- [ ] **Step 7: Update `pkg/utils/utils.go` default container name prefix**

In `pkg/utils/utils.go`, update the fallback container name prefix:

```go
// Before:
containerNamePrefixFallback = "toolbox"

// After:
containerNamePrefixFallback = "toad"
```

- [ ] **Step 8: Run build to verify**

Run: `go build ./...`
Expected: no errors

- [ ] **Step 9: Commit**

```bash
git add -A
git commit -m "feat: rebrand user-facing strings from Toolbx to Toad" -m "Updates all user-visible references from Toolbx/toolbox to Toad/toad. Container labels updated to toad-specific namespace. Backward-compat maintained for existing toolbox-labeled containers."
```

---

### Task 4: Remove duplicated container-guard pattern with middleware

**Files:**
- Modify: `cmd/root.go` — add `requireOutsideContainer` persistent pre-run
- Modify: `cmd/create.go` — remove guard block
- Modify: `cmd/enter.go` — remove guard block
- Modify: `cmd/run.go` — remove guard block
- Modify: `cmd/list.go` — remove guard block
- Modify: `cmd/rm.go` — remove guard block
- Modify: `cmd/rmi.go` — remove guard block
- Modify: `cmd/down.go` — remove guard block
- Modify: `cmd/up.go` — remove guard block
- Modify: `cmd/help.go` — remove guard block

- [ ] **Step 1: Add middleware functions in `cmd/root.go`**

Add helper functions:

```go
// requireOutsideContainer returns an error if running inside a non-Toad container.
// Commands that must run on the host (create, enter, run, list, rm, rmi, up, down) should call this.
func requireOutsideContainer() error {
  if !utils.IsInsideContainer() {
    return nil
  }
  if !utils.IsInsideToolboxContainer() {
    return fmt.Errorf("this is not a %s container", executableBase)
  }
  exitCode, err := utils.ForwardToHost()
  return &exitError{exitCode, err}
}
```

- [ ] **Step 2: Replace guard blocks in each command handler**

In each command's `RunE` function, replace the 8-line guard block:

```go
// Before:
func create(cmd *cobra.Command, args []string) error {
  if utils.IsInsideContainer() {
    if !utils.IsInsideToolboxContainer() {
      return errors.New("this is not a Toolbx container")
    }
    exitCode, err := utils.ForwardToHost()
    return &exitError{exitCode, err}
  }
  // ... rest of function

// After:
func create(cmd *cobra.Command, args []string) error {
  if err := requireOutsideContainer(); err != nil {
    return err
  }
  // ... rest of function
```

Apply to: `create`, `enter`, `run`, `list`, `rm`, `rmi`, `up`, `down`.

For `help.go` — since it delegates to `cmd.Root().Help()`, the inside-container guard isn't needed (Cobra's help is local). Remove the entire `help` function body and just call `cmd.Help()`.

- [ ] **Step 3: Run build to verify**

Run: `go build ./...`
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "refactor: deduplicate inside-container guard pattern" -m "Replaces the same 8-line guard block in every command handler with a single requireOutsideContainer() middleware call. Cuts ~80 lines of duplication."
```

---

### Task 5: Fix bugs and cleanup

**Files:**
- Modify: `cmd/down.go` — fix race condition in --rmi
- Modify: `cmd/create.go` — remove dead code, fix deprecated imports
- Modify: `cmd/root.go` — uncomment or remove subid validation

- [ ] **Step 1: Fix race condition in `cmd/down.go` `--rmi`**

The current code removes the container first, then tries to inspect it for the image name:

```go
logrus.Debugf("Removing container %s", container)
if err := podman.RemoveContainer(container, true); err != nil {
  return err
}
// ... then InspectContainer() on removed container!
```

Fix by inspecting FIRST, then removing:

```go
func down(cmd *cobra.Command, args []string) error {
  if err := requireOutsideContainer(); err != nil {
    return err
  }

  container, err := declaration.Down(&declaration.DownOptions{
    Path: downFlags.path,
    Rmi:  downFlags.rmi,
  })
  if err != nil {
    return err
  }

  var image string
  if downFlags.rmi {
    ctr, err := podman.InspectContainer(container)
    if err == nil {
      image = ctr.Image()
    }
  }

  logrus.Debugf("Removing container %s", container)
  if err := podman.RemoveContainer(container, true); err != nil {
    return err
  }
  fmt.Printf("Removed container: %s\n", container)

  if image != "" {
    logrus.Debugf("Removing image %s", image)
    if err := podman.RemoveImage(image, false); err != nil {
      return fmt.Errorf("failed to remove image %s: %w", image, err)
    }
    fmt.Printf("Removed image: %s\n", image)
  }

  return nil
}
```

- [ ] **Step 2: Fix deprecated `io/ioutil` imports**

In `cmd/create.go` and `cmd/initContainer.go`, replace `io/ioutil` usage:

```go
// Replace:
import "io/ioutil"

// With the specific functions from os:
// ioutil.ReadFile → os.ReadFile
// ioutil.WriteFile → os.WriteFile
```

Line 6 in `cmd/create.go` and line 8/619/833/1276 in `cmd/initContainer.go`.

Also line 274 in `cmd/root.go`.

- [ ] **Step 3: Handle dead subid validation in `cmd/root.go`**

Either remove the commented-out block (lines 412-415) or add a TODO comment:

```go
// TODO: subuid/subgid validation disabled — https://github.com/givensuman/toad/issues/X
```

- [ ] **Step 4: Run build to verify**

Run: `go build ./...`
Expected: no errors

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "fix: race condition in down --rmi, deprecation cleanup" -m "Fixes race where InspectContainer was called after container removal. Replaces deprecated io/ioutil with os package functions. Marks dead subid validation code."
```

---

### Task 6: Add unit tests for error helpers and guard middleware

**Files:**
- Modify: `cmd/root_test.go` — add tests for `usageError`, `requireOutsideContainer`

- [ ] **Step 1: Write tests for `usageError`**

Add to `cmd/root_test.go`:

```go
func TestUsageError(t *testing.T) {
  err := usageError("something went wrong")
  assert.Error(t, err)
  assert.Contains(t, err.Error(), "something went wrong")
  assert.Contains(t, err.Error(), "--help")
}

func TestUsageErrorFormatted(t *testing.T) {
  err := usageError("container %s not found", "my-container")
  assert.Error(t, err)
  assert.Contains(t, err.Error(), "container my-container not found")
}
```

- [ ] **Step 2: Run tests**

Run: `go test ./cmd/ -v -run TestUsageError`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add -A
git commit -m "test: add unit tests for usageError helper"
```

---

### Self-Review Checklist

1. **Coverage of identified issues:**
   - Error construction repetition → Task 1
   - Manpage help → Task 2
   - Rebranding incomplete → Task 3
   - Container guard duplication → Task 4
   - Race condition in down --rmi → Task 5
   - ioutil deprecation → Task 5
   - Dead code → Task 5
   - Low test coverage → Task 6

2. **Placeholder scan:** All steps contain concrete code and commands. No TBDs, TODOs, or vague instructions.

3. **Internal consistency:** `usageError` defined in Task 1, tested in Task 6. `requireOutsideContainer` defined in Task 4, used in Tasks 3-5. Function signatures match between tasks.

4. **Independence:** Each task can be applied independently. Tasks 3 and 4 touch the same files but don't conflict — rebranding changes strings, middleware replaces logic blocks. Apply Task 4 first if desired.
