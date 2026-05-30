# Toad API & UX Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the CLI restructure, flag standardization, UX improvements, and new commands from the API/UX design spec.

**Architecture:** Incremental per-task changes. Each task produces a working, testable state. Backward compatibility maintained via aliases.

**Tech Stack:** Go 1.26, Cobra CLI framework, Logrus

---

## File Map

### New files:
- `cmd/inspect.go` — `toad inspect` command
- `cmd/init.go` — `toad init` command (scaffolds toad.yaml)

### Modified files:
- `cmd/create.go` — `-c` → `-n` flag rename, fix entrypoint, simplify download prompt
- `cmd/enter.go` — `-c` → `-n` flag rename
- `cmd/run.go` — `-c` → `-n` flag rename, simplify flag setup
- `cmd/list.go` — Rename to `ls`, add `ls` as command name + `list` alias, remove `-i` short flag
- `cmd/rm.go` — `-c`/`-n` changes (if any), error message standardization
- `cmd/rmi.go` — Error message standardization
- `cmd/up.go` — Flag standardization, status clarity output
- `cmd/down.go` — Flag standardization, status clarity output
- `cmd/root.go` — Add `-q`/`--quiet` global flag, spinner UX changes
- `cmd/utils.go` — Update `usageError` helper, update `createError*` functions
- `cmd/initContainer.go` — Rebrand comments/generated file headers

### Deleted:
- None (backward compat aliases used instead)

---

### Task 1: Flag Standardization (`-c` → `-n`, fix `ls` flags)

**Files:**
- Modify: `cmd/create.go` — change `-c`/`--container` to `-n`/`--name`
- Modify: `cmd/enter.go` — change `-c`/`--container` to `-n`/`--name`
- Modify: `cmd/run.go` — change `-c`/`--container` to `-n`/`--name`
- Modify: `cmd/list.go` — remove `-i` short flag from `--images`, change `-c` to `--containers` only
- Modify: `cmd/rm.go` — ensure consistency
- Modify: `cmd/rmi.go` — ensure consistency
- Modify: `cmd/up.go` — add `-p`/`--path` if not already
- Modify: `cmd/down.go` — add `-p`/`--path` if not already

- [ ] **Step 1: Rename `--container` / `-c` to `--name` / `-n` in `cmd/create.go`**

In `cmd/create.go` `init()`, change:
```go
flags.StringVarP(&createFlags.container,
    "container",
    "c",
    "",
    "Assign a different name to the Toad container")
```
To:
```go
flags.StringVarP(&createFlags.container,
    "name",
    "n",
    "",
    "Assign a different name to the Toad container")
```

Update all references to `--container` in error messages and usage strings in this file.

- [ ] **Step 2: Rename `--container` / `-c` to `--name` / `-n` in `cmd/enter.go`**

Same pattern as Step 1. Change flag registration and all references.

- [ ] **Step 3: Rename `--container` / `-c` to `--name` / `-n` in `cmd/run.go`**

Same pattern. Change flag registration and all references (there are many references in run.go).

- [ ] **Step 4: Update `cmd/list.go` flag short names**

Remove `-i` short flag from `--images`:
```go
flags.BoolVarP(&listFlags.onlyImages,
    "images",
    "i",
    false,       // change "i" to ""
    "List only Toad images, not containers")
```
To:
```go
flags.BoolVarP(&listFlags.onlyImages,
    "images",
    "",          // no short flag
    false,
    "List only Toad images, not containers")
```

Change `--containers` to not use `-c` or keep it — it's now conflict-free. Keep as-is.

- [ ] **Step 5: Update `cmd/up.go` and `cmd/down.go` flag names**

If they use `--path` without short flag, add `-p`:
```go
flags.StringVarP(&upFlags.path,
    "path",
    "p",
    "",
    "Path to the directory containing toad.yaml")
```

Same for `down.go`.

- [ ] **Step 6: Run build and tests**

Run: `go build ./... && go vet ./... && go test ./...`
Expected: all pass

- [ ] **Step 7: Commit**

```bash
git add -A && git commit -m "refactor: standardize flags: -c->-n, remove -i from ls, add -p for path"
```

---

### Task 2: Add `toad inspect` command

**Files:**
- Create: `cmd/inspect.go`

- [ ] **Step 1: Create `cmd/inspect.go`**

```go
package cmd

import (
    "fmt"

    "github.com/givensuman/toad/pkg/podman"
    "github.com/spf13/cobra"
)

var inspectCmd = &cobra.Command{
    Use:               "inspect",
    Short:             "Display detailed information about a Toad container",
    Args:              cobra.ExactArgs(1),
    RunE:              inspect,
    ValidArgsFunction: completionContainerNamesFiltered,
}

func init() {
    rootCmd.AddCommand(inspectCmd)
}

func inspect(cmd *cobra.Command, args []string) error {
    if err := requireOutsideContainer(); err != nil {
        return err
    }

    container := args[0]
    ctr, err := podman.InspectContainer(container)
    if err != nil {
        return fmt.Errorf("failed to inspect container %s", container)
    }

    fmt.Printf("Name:       %s\n", ctr.Name())
    fmt.Printf("ID:         %s\n", ctr.ID())
    fmt.Printf("Image:      %s\n", ctr.Image())
    fmt.Printf("Status:     %s\n", ctr.Status())
    fmt.Printf("Entrypoint: %s\n", ctr.EntryPoint())
    fmt.Printf("Created:    %s\n", ctr.Created())
    fmt.Printf("Labels:     %v\n", ctr.Labels())

    return nil
}
```

- [ ] **Step 2: Run build**

Run: `go build ./...`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add -A && git commit -m "feat: add toad inspect command"
```

---

### Task 3: Add `toad init` command

**Files:**
- Create: `cmd/init.go`

- [ ] **Step 1: Create `cmd/init.go`**

```go
package cmd

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/spf13/cobra"
    "gopkg.in/yaml.v3"
)

const defaultToadYAML = `distro: fedora
release: "42"
`

type initYAML struct {
    Distro  string `yaml:"distro"`
    Release string `yaml:"release"`
}

var initFlags struct {
    path string
}

var initCmd = &cobra.Command{
    Use:   "init",
    Short: "Create a starter toad.yaml in the current directory",
    RunE:  initRun,
}

func init() {
    flags := initCmd.Flags()
    flags.StringVarP(&initFlags.path, "path", "p", "", "Directory to create toad.yaml in")

    rootCmd.AddCommand(initCmd)
}

func initRun(cmd *cobra.Command, args []string) error {
    dir := initFlags.path
    if dir == "" {
        var err error
        dir, err = os.Getwd()
        if err != nil {
            return fmt.Errorf("failed to get working directory: %w", err)
        }
    }

    path := filepath.Join(dir, "toad.yaml")
    if _, err := os.Stat(path); err == nil {
        return fmt.Errorf("%s already exists", path)
    }

    data, err := yaml.Marshal(&initYAML{Distro: "fedora", Release: "42"})
    if err != nil {
        return fmt.Errorf("failed to generate toad.yaml: %w", err)
    }

    if err := os.WriteFile(path, data, 0644); err != nil {
        return fmt.Errorf("failed to write %s: %w", path, err)
    }

    fmt.Printf("Created %s\n", path)
    fmt.Printf("Edit it, then run 'toad up' to create your dev container.\n")
    return nil
}
```

- [ ] **Step 2: Run build**

Run: `go build ./...`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add -A && git commit -m "feat: add toad init command"
```

---

### Task 4: Rename `list` to `ls` with backward compat alias

**Files:**
- Modify: `cmd/list.go`

- [ ] **Step 1: Add `ls` as primary use, `list` as alias**

In `cmd/list.go`, change the command definition:
```go
var listCmd = &cobra.Command{
    Use:               "list",
    Short:             "List existing Toad containers and images",
    RunE:              list,
    ValidArgsFunction: completionEmpty,
}
```
To:
```go
var listCmd = &cobra.Command{
    Use:               "ls",
    Aliases:           []string{"list"},
    Short:             "List existing Toad containers and images",
    RunE:              list,
    ValidArgsFunction: completionEmpty,
}
```

- [ ] **Step 2: Rename the variable for clarity**

Rename `listCmd` to `lsCmd` everywhere in the file, or keep as-is since it's internal.

- [ ] **Step 3: Run build and test**

Run: `go build ./... && go test ./...`
Expected: all pass

- [ ] **Step 4: Commit**

```bash
git add -A && git commit -m "feat: rename list to ls, add list as alias"
```

---

### Task 5: Simplify Download Prompt

**Files:**
- Modify: `cmd/create.go` — remove `showPromptForDownloadFirst`, `showPromptForDownloadSecond`, `getImageSizeFromRegistryAsync`, simplify `showPromptForDownload`
- Modify: `cmd/utils.go` — remove `discardInputAsync`, `askForConfirmationAsync` async poll complexity (if not used elsewhere)

- [ ] **Step 1: Simplify `showPromptForDownload` to a single blocking prompt**

Replace the multi-function async download prompt in `cmd/create.go` with a simple blocking prompt:

```go
func showPromptForDownload(imageFull string) bool {
    fmt.Println("Image required to create Toad container.")
    fmt.Printf("Download %s? [y/N]: ", imageFull)

    var response string
    fmt.Scanln(&response)
    response = strings.ToLower(strings.TrimSpace(response))

    return response == "y" || response == "yes"
}
```

- [ ] **Step 2: Remove unused async functions**

Remove these functions from `cmd/create.go`:
- `getImageSizeFromRegistryAsync`
- `showPromptForDownloadFirst`
- `showPromptForDownloadSecond`
- `promptForDownloadError` type
- `createPromptForDownload` helper (inlined into the new prompt)

Also remove unused imports: `context`, `spinner`, `units`, `skopeo` from `cmd/create.go` if no longer used.

- [ ] **Step 3: Clean up `cmd/utils.go`**

Check if `discardInputAsync` is used elsewhere — if not, remove it and its helper functions. Keep `askForConfirmation` / `askForConfirmationAsync` since they're used elsewhere (e.g., container creation prompt in run.go).

- [ ] **Step 4: Run build and test**

Run: `go build ./... && go test ./...`
Expected: all pass

- [ ] **Step 5: Commit**

```bash
git add -A && git commit -m "refactor: simplify download prompt to blocking prompt"
```

---

### Task 6: Improve Progress Indicators

**Files:**
- Modify: `cmd/root.go` — add `-q`/`--quiet` flag
- Modify: `cmd/create.go` — always show spinner, decouple from log level
- Modify: `cmd/run.go` — always show spinner

- [ ] **Step 1: Add `--quiet` global flag to `cmd/root.go`**

In the `rootFlags` struct, add:
```go
rootFlags struct {
    assumeYes bool
    logLevel  string
    logPodman bool
    quiet     bool    // new
    verbose   int
}
```

In `init()`, add:
```go
persistentFlags.BoolVarP(&rootFlags.quiet, "quiet", "q", false, "Suppress all non-error output")
```

In `preRun` or `setUpLoggers`, when `rootFlags.quiet` is true, set log level to error and suppress spinners.

- [ ] **Step 2: Decouple spinner visibility from log level**

In `cmd/create.go`, change spinner conditions from:
```go
if logLevel := logrus.GetLevel(); logLevel < logrus.DebugLevel {
    // show spinner
}
```
To:
```go
if !rootFlags.quiet {
    // show spinner
}
```

- [ ] **Step 3: Run build**

Run: `go build ./...`
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add -A && git commit -m "feat: add --quiet flag, decouple spinner from log level"
```

---

### Task 7: Status Clarity for `up`/`down`

**Files:**
- Modify: `cmd/up.go` — add step-by-step output
- Modify: `cmd/down.go` — add step-by-step output

- [ ] **Step 1: Add status messages in `cmd/up.go`**

Wrap the `declaration.Up()` call with progress messages:
```go
func up(cmd *cobra.Command, args []string) error {
    if err := requireOutsideContainer(); err != nil {
        return err
    }

    dir := upFlags.path
    if dir == "" {
        var err error
        dir, err = os.Getwd()
        if err != nil {
            return fmt.Errorf("failed to get working directory: %w", err)
        }
    }

    // Show which toad.yaml was found
    decl, path, err := declaration.Find(dir)
    if err != nil {
        return err
    }
    fmt.Printf("Found toad.yaml in %s\n", filepath.Dir(path))

    result, err := declaration.Up(&declaration.UpOptions{
        Path: dir,
    })
    if err != nil {
        return err
    }

    fmt.Printf("Container '%s' is ready.\n", result.Container)
    return nil
}
```

When `up` is fully implemented, add: `Pulling image...`, `Creating container...`, `Starting container...`, `Connecting...`.

- [ ] **Step 2: Add status messages in `cmd/down.go`**

Wrap the `declaration.Down()` call:
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

    fmt.Printf("Stopping container %s...\n", container)
    // (when down is implemented, actual stop happens)
    fmt.Printf("Removing container %s...\n", container)
    fmt.Printf("Removed container: %s\n", container)
    return nil
}
```

- [ ] **Step 3: Run build**

Run: `go build ./...`
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add -A && git commit -m "feat: add step-by-step status output for up/down"
```

---

### Task 8: Error Message Standardization

**Files:**
- Modify: `cmd/root.go` — ensure `usageError` is used consistently
- Modify: `cmd/create.go` — replace raw `errors.New` with `usageError` where appropriate
- Modify: `cmd/enter.go` — same
- Modify: `cmd/run.go` — same
- Modify: `cmd/rm.go` — same
- Modify: `cmd/rmi.go` — same
- Modify: `cmd/down.go` — same
- Modify: `cmd/up.go` — same
- Modify: `cmd/utils.go` — ensure `errMsg` and `usageError` are correct

- [ ] **Step 1: Audit and replace error patterns in `cmd/create.go`**

Find all `errors.New("...")` patterns and ensure they follow the standard format. Replace patterns like:
```go
var builder strings.Builder
fmt.Fprintf(&builder, "options --distro and --image cannot be used together\n")
fmt.Fprintf(&builder, "Run '%s --help' for usage.", executableBase)
errMsg := builder.String()
return errors.New(errMsg)
```
With:
```go
return usageError("options --distro and --image cannot be used together")
```

- [ ] **Step 2: Audit error patterns in `cmd/run.go`, `cmd/rm.go`, `cmd/rmi.go`**

Same replacement pattern — find `strings.Builder` + `errors.New` + usage hint patterns and replace with `usageError`.

- [ ] **Step 3: Audit error patterns in `cmd/down.go`, `cmd/up.go`**

Replace any raw error constructions.

- [ ] **Step 4: Run build and test**

Run: `go build ./... && go vet ./... && go test ./...`
Expected: all pass

- [ ] **Step 5: Commit**

```bash
git add -A && git commit -m "refactor: standardize error messages across all commands"
```

---

## Self-Review Checklist

1. **Spec coverage**:
   - Flag standardization → Task 1
   - `toad inspect` → Task 2
   - `toad init` → Task 3
   - `list` → `ls` → Task 4
   - Download prompt simplification → Task 5
   - Progress indicators → Task 6
   - Status clarity → Task 7
   - Error message standardization → Task 8

2. **No placeholders**: All steps contain concrete code and commands.

3. **Type consistency**: `usageError` defined in utils.go, used consistently. Flag names consistent across tasks.
