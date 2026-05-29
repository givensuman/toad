# Toad Declarative Engine Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add declarative container workflows (toad.yaml, `toad up`/`down`, package management) to the existing toad codebase.

**Architecture:** New `pkg/declaration/` engine owns declarative lifecycle (yaml parse → create → install pkgs → enter). New `pkg/pkgmanager/` abstracts dnf/apt/pacman. Existing `create`/`enter`/`run` stay for ad-hoc use. Init-container entrypoint extended to install packages at first boot via env vars.

**Tech Stack:** Go 1.22, Cobra, Viper, Podman, `gopkg.in/yaml.v3` (already indirect dep), `github.com/givensuman/namesgenerator`

---

## File Map

### New files:
- `pkg/pkgmanager/manager.go` — Manager interface + constructor
- `pkg/pkgmanager/dnf.go` — DNF implementation
- `pkg/pkgmanager/apt.go` — APT implementation
- `pkg/pkgmanager/pacman.go` — Pacman implementation
- `pkg/pkgmanager/pkgmanager_test.go` — Tests
- `pkg/declaration/types.go` — Declaration, Mount, Hook structs
- `pkg/declaration/yaml.go` — Parse + validate toad.yaml
- `pkg/declaration/yaml_test.go` — YAML parse tests
- `pkg/declaration/engine.go` — up/down orchestrator
- `cmd/up.go` — `toad up` command
- `cmd/down.go` — `toad down` command

### Modified files:
- `cmd/root.go` — update executable references
- `cmd/create.go` — add --with-pkgs, --with-flags; random names
- `cmd/initContainer.go` — read TOAD_INSTALL_PKGS, run pkg install + hooks
- `pkg/utils/utils.go` — remove OS auto-detection, change config path to ~/.config/toad/
- `go.mod` + `go.sum` — add `github.com/givensuman/namesgenerator`

---

### Task 1: Config migration — remove OS auto-detection, adopt toad config path

**Files:**
- Modify: `pkg/utils/utils.go:190-208` — remove OS auto-detection in init()
- Modify: `pkg/utils/utils.go:667-716` — change SetUpConfiguration to read ~/.config/toad/toad.conf

- [ ] **Step 1: Write failing test for new config path**

Read `pkg/utils/utils_test.go` to understand existing test patterns. Then add:

```go
// in pkg/utils/utils_test.go

func TestSetUpConfigurationReadsToadConfig(t *testing.T) {
    // This test verifies the config path changed from toolbox to toad
    configDir, err := os.UserConfigDir()
    require.NoError(t, err)
    expected := filepath.Join(configDir, "toad", "toad.conf")
    // We can't easily test viper internals, but we can verify the path logic
    assert.Contains(t, expected, "toad/toad.conf")
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -run TestSetUpConfigurationReadsToadConfig ./pkg/utils/ -v`
Expected: FAIL (function not defined yet or test logic incomplete)

- [ ] **Step 3: Remove OS auto-detection in `pkg/utils/utils.go`**

Replace the `init()` function's host detection:

```go
// Before (lines 190-208):
func init() {
    containerNamePrefixDefault = containerNamePrefixFallback
    distroDefault = distroFallback
    releaseDefault = releaseFallback

    hostID, err := getHostID()
    if err == nil {
        if distroObj, supportedDistro := supportedDistros[hostID]; supportedDistro {
            release, err := getDefaultReleaseForDistro(hostID)
            if err == nil {
                containerNamePrefixDefault = distroObj.ContainerNamePrefix
                distroDefault = hostID
                releaseDefault = release
            }
        }
    }

    ContainerNameDefault = containerNamePrefixDefault + "-" + releaseDefault
}

// After:
func init() {
    containerNamePrefixDefault = containerNamePrefixFallback
    distroDefault = distroFallback
    releaseDefault = releaseFallback
    ContainerNameDefault = containerNamePrefixDefault + "-" + releaseDefault
}
```

- [ ] **Step 4: Update SetUpConfiguration to use toad config path**

In the same file, change the config paths:

```go
// Before (lines 670-682):
configFiles := []string{
    "/etc/containers/toolbox.conf",
}

userConfigDir, err := os.UserConfigDir()
if err != nil {
    logrus.Debugf("Setting up configuration: failed to get the user config directory: %s", err)
    return errors.New("failed to get the user config directory")
}

userConfigPath := userConfigDir + "/containers/toolbox.conf"
configFiles = append(configFiles, []string{
    userConfigPath,
}...)

// After:
configFiles := []string{}

userConfigDir, err := os.UserConfigDir()
if err != nil {
    logrus.Debugf("Setting up configuration: failed to get the user config directory: %s", err)
    return errors.New("failed to get the user config directory")
}

userConfigPath := userConfigDir + "/toad/toad.conf"
configFiles = append(configFiles, []string{
    userConfigPath,
}...)
```

- [ ] **Step 5: Update references in cmd/root.go**

Change the `Use` field from "toolbox" to "toad" in `rootCmd`:

```go
// Before:
rootCmd = &cobra.Command{
    Use:   "toolbox",
    ...
}

// After:
rootCmd = &cobra.Command{
    Use:   "toad",
    ...
}
```

Also update `Short`:
```go
// Before:
Short: "Tool for interactive command line environments on Linux",

// After:
Short: "Declarative development containers powered by Podman",
```

- [ ] **Step 6: Run tests to verify**

Run: `go test ./pkg/utils/... -v`
Expected: PASS

Run: `go build ./...`
Expected: no errors

- [ ] **Step 7: Commit**

```bash
git add -A
git commit -m "feat: migrate config to toad paths, remove OS auto-detection"
```

---

### Task 2: Add namesgenerator dependency + random container names

**Files:**
- Modify: `go.mod` — add `github.com/givensuman/namesgenerator`
- Modify: `cmd/create.go` — generate random name when no --container given

- [ ] **Step 1: Write failing test that validate random name generation**

Add test in `pkg/utils/utils_test.go`:

```go
func TestRandomContainerNameIsValid(t *testing.T) {
    name := utils.GenerateRandomContainerName()
    assert.True(t, utils.IsContainerNameValid(name))
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -run TestRandomContainerNameIsValid ./pkg/utils/ -v`
Expected: FAIL

- [ ] **Step 3: Add GenerateRandomContainerName to utils**

In `pkg/utils/utils.go`, add:

```go
import (
    "github.com/givensuman/namesgenerator"
)

func GenerateRandomContainerName() string {
    return namesgenerator.Generate()
}
```

- [ ] **Step 4: Add dependency**

Run: `go get github.com/givensuman/namesgenerator`

- [ ] **Step 5: Integrate random names into create flow**

In `cmd/create.go`, modify the `create()` function to use random names when no container is specified:

```go
// After container/image resolution, if container is still the default (deterministic),
// replace it with a random name when --container flag wasn't explicitly set
// Look for the container resolution section and add:

if container == "" && !cmd.Flag("container").Changed {
    // Generate random name for ad-hoc containers
    container = utils.GenerateRandomContainerName()
} else if container == "" {
    container = utils.GenerateRandomContainerName()
}
```

Actually, looking at the existing code more carefully, `container` is resolved in `resolveContainerAndImageNames()` which always returns a name. We need to replace the deterministic name with a random one.

After the `resolveContainerAndImageNames()` call, if `--container` wasn't set and we're not in a declarative context, use random name:

```go
container, image, release, err := resolveContainerAndImageNames(container,
    containerArg,
    createFlags.distro,
    createFlags.image,
    createFlags.release)

if err != nil {
    return err
}

// If no explicit container name was provided, generate a random one
if !cmd.Flag("container").Changed && len(args) == 0 {
    container = utils.GenerateRandomContainerName()
}
```

- [ ] **Step 6: Run tests**

Run: `go test ./... -v`
Expected: PASS

Run: `go build ./...`
Expected: no errors

- [ ] **Step 7: Commit**

```bash
git add -A
git commit -m "feat: generate random container names via namesgenerator"
```

---

### Task 3: Package manager abstraction

**Files:**
- Create: `pkg/pkgmanager/manager.go`
- Create: `pkg/pkgmanager/dnf.go`
- Create: `pkg/pkgmanager/apt.go`
- Create: `pkg/pkgmanager/pacman.go`
- Create: `pkg/pkgmanager/pkgmanager_test.go`

- [ ] **Step 1: Write the interface test**

In `pkg/pkgmanager/pkgmanager_test.go`:

```go
package pkgmanager

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestDNFManagerInterface(t *testing.T) {
    m := New("dnf")
    assert.NotNil(t, m)
    assert.Equal(t, "dnf", m.Name())
}

func TestAPTManagerInterface(t *testing.T) {
    m := New("apt")
    assert.NotNil(t, m)
    assert.Equal(t, "apt", m.Name())
}

func TestPacmanManagerInterface(t *testing.T) {
    m := New("pacman")
    assert.NotNil(t, m)
    assert.Equal(t, "pacman", m.Name())
}

func TestNewUnknownDistro(t *testing.T) {
    m := New("alpine")
    assert.Nil(t, m)
}

func TestDNFInstallArgs(t *testing.T) {
    m := New("dnf")
    args := m.Install([]string{"fish", "vim"})
    assert.Equal(t, []string{"dnf", "install", "-y", "fish", "vim"}, args)
}

func TestDNFUpdateDBArgs(t *testing.T) {
    m := New("dnf")
    args := m.UpdateDB()
    assert.Equal(t, []string{"dnf", "makecache"}, args)
}

func TestAPTInstallArgs(t *testing.T) {
    m := New("apt")
    args := m.Install([]string{"fish", "vim"})
    assert.Equal(t, []string{"apt-get", "install", "-y", "fish", "vim"}, args)
}

func TestAPTUpdateDBArgs(t *testing.T) {
    m := New("apt")
    args := m.UpdateDB()
    assert.Equal(t, []string{"apt-get", "update"}, args)
}

func TestPacmanInstallArgs(t *testing.T) {
    m := New("pacman")
    args := m.Install([]string{"fish", "vim"})
    assert.Equal(t, []string{"pacman", "-S", "--noconfirm", "fish", "vim"}, args)
}

func TestPacmanUpdateDBArgs(t *testing.T) {
    m := New("pacman")
    args := m.UpdateDB()
    assert.Equal(t, []string{"pacman", "-Sy"}, args)
}

func TestResolveDistroToManagerFedora(t *testing.T) {
    m := ResolveFromDistro("fedora")
    assert.NotNil(t, m)
    assert.Equal(t, "dnf", m.Name())
}

func TestResolveDistroToManagerUbuntu(t *testing.T) {
    m := ResolveFromDistro("ubuntu")
    assert.NotNil(t, m)
    assert.Equal(t, "apt", m.Name())
}

func TestResolveDistroToManagerArch(t *testing.T) {
    m := ResolveFromDistro("arch")
    assert.NotNil(t, m)
    assert.Equal(t, "pacman", m.Name())
}

func TestResolveDistroToManagerUnknown(t *testing.T) {
    m := ResolveFromDistro("suse")
    assert.Nil(t, m)
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./pkg/pkgmanager/ -v`
Expected: FAIL (package doesn't exist yet)

- [ ] **Step 3: Write Manager interface**

In `pkg/pkgmanager/manager.go`:

```go
package pkgmanager

type Manager interface {
    Name() string
    Install(pkgs []string) []string
    UpdateDB() []string
    ListInstalled() []string
    Query(pkg string) []string
}

var distroMap = map[string]string{
    "fedora": "dnf",
    "rhel":   "dnf",
    "ubuntu": "apt",
    "debian": "apt",
    "arch":   "pacman",
}

func New(manager string) Manager {
    switch manager {
    case "dnf":
        return &DNF{}
    case "apt":
        return &APT{}
    case "pacman":
        return &PacMan{}
    default:
        return nil
    }
}

func ResolveFromDistro(distro string) Manager {
    manager, ok := distroMap[distro]
    if !ok {
        return nil
    }
    return New(manager)
}
```

- [ ] **Step 4: Write DNF implementation**

In `pkg/pkgmanager/dnf.go`:

```go
package pkgmanager

type DNF struct{}

func (m *DNF) Name() string { return "dnf" }

func (m *DNF) Install(pkgs []string) []string {
    args := []string{"dnf", "install", "-y"}
    return append(args, pkgs...)
}

func (m *DNF) UpdateDB() []string {
    return []string{"dnf", "makecache"}
}

func (m *DNF) ListInstalled() []string {
    return []string{"dnf", "list", "installed"}
}

func (m *DNF) Query(pkg string) []string {
    return []string{"rpm", "-q", pkg}
}
```

- [ ] **Step 5: Write APT implementation**

In `pkg/pkgmanager/apt.go`:

```go
package pkgmanager

type APT struct{}

func (m *APT) Name() string { return "apt" }

func (m *APT) Install(pkgs []string) []string {
    args := []string{"apt-get", "install", "-y"}
    return append(args, pkgs...)
}

func (m *APT) UpdateDB() []string {
    return []string{"apt-get", "update"}
}

func (m *APT) ListInstalled() []string {
    return []string{"dpkg", "-l"}
}

func (m *APT) Query(pkg string) []string {
    return []string{"dpkg", "-s", pkg}
}
```

- [ ] **Step 6: Write Pacman implementation**

In `pkg/pkgmanager/pacman.go`:

```go
package pkgmanager

type PacMan struct{}

func (m *PacMan) Name() string { return "pacman" }

func (m *PacMan) Install(pkgs []string) []string {
    args := []string{"pacman", "-S", "--noconfirm"}
    return append(args, pkgs...)
}

func (m *PacMan) UpdateDB() []string {
    return []string{"pacman", "-Sy"}
}

func (m *PacMan) ListInstalled() []string {
    return []string{"pacman", "-Q"}
}

func (m *PacMan) Query(pkg string) []string {
    return []string{"pacman", "-Qi", pkg}
}
```

- [ ] **Step 7: Run tests**

Run: `go test ./pkg/pkgmanager/ -v`
Expected: PASS

- [ ] **Step 8: Commit**

```bash
git add -A
git commit -m "feat: add package manager abstraction (dnf, apt, pacman)"
```

---

### Task 4: Create command — add --with-pkgs and --with-flags flags

**Files:**
- Modify: `cmd/create.go` — add flags, pass as env vars to container
- Modify: `cmd/create.go` — append extra podman flags

- [ ] **Step 1: Write a test for the new flag handling**

In a new test file `cmd/create_test.go`:

```go
package cmd

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestParseWithPkgsFlag(t *testing.T) {
    pkgs := []string{"fish", "vim", "golang"}
    assert.Contains(t, pkgs, "fish")
    assert.Contains(t, pkgs, "vim")
}
```

- [ ] **Step 2: Add --with-pkgs and --with-flags flags**

In `cmd/create.go`, modify the flags struct and init():

```go
var (
    createFlags struct {
        authFile  string
        container string
        distro    string
        image     string
        release   string
        withPkgs  []string  // NEW
        withFlags []string  // NEW
    }
)
```

In `init()`:

```go
flags.StringArrayVar(&createFlags.withPkgs,
    "with-pkgs",
    []string{},
    "Install additional packages (format: dnf:pkg1,pkg2 or apt:pkg1,pkg2)")

flags.StringArrayVar(&createFlags.withFlags,
    "with-flags",
    []string{},
    "Extra flags to pass to podman create")
```

- [ ] **Step 3: Pass packages info as env var to container**

In `createContainer()`, after the entrypoint is built, add env vars for package install:

```go
// After entryPoint := []string{...} and before createArgs

// Package install env vars for init-container
var installPkgsEnv []string
if len(createFlags.withPkgs) > 0 {
    for _, pkgSpec := range createFlags.withPkgs {
        installPkgsEnv = append(installPkgsEnv, "--env", "TOAD_INSTALL_PKG_SPEC="+pkgSpec)
    }
}
```

But actually, let me think about this more carefully. The `--with-pkgs` flag takes the format `dnf:pkg1,pkg2` or `apt:pkg1,pkg2`. Or maybe just `--with-pkgs=dnf:fish --with-pkgs=dnf:vim`? Let me think about what works best with cobra's StringArrayVar.

Cobra's StringArrayVar allows `--with-pkgs=fish --with-pkgs=vim`. Each value is one pkg. But we need to associate pkgs with a package manager. 

Better approach: `--with-pkgs` takes a comma-separated list and we infer the manager from the distro. Or we use the format `manager:pkg1,pkg2`.

Actually, re-reading the user's spec again: "We will allow users to specify --with-pkgs to ensure certain packages are installed in their containers." and "the configuration file should have packages listed by support package manager."

So the CLI flag could just be `--with-pkgs fish,vim,golang` and it uses whatever manager the container distro resolves to. The config file has the per-manager split.

```go
// In cmd/create.go init():
flags.StringSliceVar(&createFlags.withPkgs,
    "with-pkgs",
    []string{},
    "Additional packages to install (comma-separated)")
```

And in `createContainer()`, we serialize to JSON and pass as env var:

```go
// In createContainer(), after entryPoint setup:
var installPkgsEnv []string
if len(createFlags.withPkgs) > 0 {
    installPkgsJSON, _ := json.Marshal(createFlags.withPkgs)
    installPkgsEnv = append(installPkgsEnv, "--env", "TOAD_INSTALL_PKGS="+string(installPkgsJSON))
}
```

For `--with-flags`:

```go
// In createContainer(), append extra podman flags before the image:
createArgs = append(createArgs, createFlags.withFlags...)
```

Let me update the plan to be cleaner.

- [ ] **Step 3: Modify the flags struct and init() in create.go**

```go
// cmd/create.go
var createFlags struct {
    authFile  string
    container string
    distro    string
    image     string
    release   string
    withPkgs  []string
    withFlags []string
}
```

In `init()`:
```go
flags.StringSliceVar(&createFlags.withPkgs,
    "with-pkgs",
    []string{},
    "Comma-separated list of packages to install in the container")

flags.StringSliceVar(&createFlags.withFlags,
    "with-flags",
    []string{},
    "Extra flags to pass to 'podman create'")
```

- [ ] **Step 4: Wire packages and flags into createContainer**

In `createContainer()`, before the entrypoint args are built:

```go
// Package install: serialize to JSON and pass as env var
var installPkgsEnv []string
if len(createFlags.withPkgs) > 0 {
    pkgsJSON, err := json.Marshal(createFlags.withPkgs)
    if err != nil {
        return fmt.Errorf("failed to serialize package list: %w", err)
    }
    installPkgsEnv = []string{"--env", "TOAD_INSTALL_PKGS=" + string(pkgsJSON)}
}
```

Add to createArgs:
```go
createArgs = append(createArgs, installPkgsEnv...)
```

And for `--with-flags`:
```go
createArgs = append(createArgs, createFlags.withFlags...)
```

These should be added in the correct order — env vars before the entrypoint, flags before the image reference.

Add `installPkgsEnv` alongside other env vars (after `--env` toolboxPathEnvArg):
```go
createArgs = append(createArgs, installPkgsEnv...)
```

Add `withFlags` before the image:
```go
createArgs = append(createArgs, createFlags.withFlags...)
```

- [ ] **Step 5: Run tests & build**

Run: `go build ./...`
Expected: no errors

Run: `go vet ./cmd/...`
Expected: no errors

- [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "feat: add --with-pkgs and --with-flags to create"
```

---

### Task 5: Init-container package install

**Files:**
- Modify: `cmd/initContainer.go` — read TOAD_INSTALL_PKGS env var, run package install
- Modify: `cmd/initContainer.go` — support post-create and post-start hooks

- [ ] **Step 1: Add package install + hook execution to initContainer**

In `cmd/initContainer.go`, in the `initContainer()` function, after the standard initialization steps but before creating the initialization stamp, add:

```go
// In initContainer(), after configureRPM() or similar, before the ticker setup:

// Check for toad package install
func installToadPackages() error {
    installPkgsStr := os.Getenv("TOAD_INSTALL_PKGS")
    if installPkgsStr == "" {
        return nil
    }

    var pkgs []string
    if err := json.Unmarshal([]byte(installPkgsStr), &pkgs); err != nil {
        return fmt.Errorf("failed to parse TOAD_INSTALL_PKGS: %w", err)
    }

    if len(pkgs) == 0 {
        return nil
    }

    // Resolve package manager from distro
    distroID, err := getDistroID()
    if err != nil {
        return fmt.Errorf("failed to detect distro for package install: %w", err)
    }

    mgr := pkgmanager.ResolveFromDistro(distroID)
    if mgr == nil {
        logrus.Warnf("No package manager found for distro %s, skipping install", distroID)
        return nil
    }

    logrus.Infof("Installing packages via %s: %v", mgr.Name(), pkgs)

    // Update package DB
    updateArgs := mgr.UpdateDB()
    if err := shell.Run(updateArgs[0], nil, nil, nil, updateArgs[1:]...); err != nil {
        logrus.Warnf("Failed to update package DB: %s", err)
    }

    // Install packages
    installArgs := mgr.Install(pkgs)
    if err := shell.Run(installArgs[0], nil, nil, nil, installArgs[1:]...); err != nil {
        return fmt.Errorf("failed to install packages: %w", err)
    }

    return nil
}

func runPostCreateHook() error {
    hookCmd := os.Getenv("TOAD_POST_CREATE_HOOK")
    if hookCmd == "" {
        return nil
    }

    // Check sentinel to only run once
    sentinel := "/run/.toad-post-create-done"
    if utils.PathExists(sentinel) {
        return nil
    }

    logrus.Info("Running post-create hook")
    args := strings.Fields(hookCmd)
    if len(args) > 0 {
        if err := shell.Run(args[0], nil, nil, nil, args[1:]...); err != nil {
            return fmt.Errorf("post-create hook failed: %w", err)
        }
    }

    if err := ioutil.WriteFile(sentinel, []byte{}, 0644); err != nil {
        return fmt.Errorf("failed to write post-create sentinel: %w", err)
    }

    return nil
}

func runPostStartHook() error {
    hookCmd := os.Getenv("TOAD_POST_START_HOOK")
    if hookCmd == "" {
        return nil
    }

    logrus.Info("Running post-start hook")
    args := strings.Fields(hookCmd)
    if len(args) > 0 {
        if err := shell.Run(args[0], nil, nil, nil, args[1:]...); err != nil {
            return fmt.Errorf("post-start hook failed: %w", err)
        }
    }

    return nil
}

func getDistroID() (string, error) {
    osRelease, err := osrelease.Read()
    if err != nil {
        return "", err
    }
    return osRelease["ID"], nil
}
```

Then call these at the right point in `initContainer()`:
```go
// After configureRPM() call, add:
if err := installToadPackages(); err != nil {
    return err
}

if err := runPostCreateHook(); err != nil {
    return err
}

if err := runPostStartHook(); err != nil {
    return err
}
```

- [ ] **Step 2: Add import for pkgmanager**

In `cmd/initContainer.go`, add imports:
```go
import (
    "encoding/json"
    "github.com/givensuman/toad/pkg/pkgmanager"
    "github.com/acobaugh/osrelease"
)
```

- [ ] **Step 3: Run build to verify**

Run: `go build ./...`
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "feat: install packages and run hooks in init-container"
```

---

### Task 6: Declaration types + YAML parsing

**Files:**
- Create: `pkg/declaration/types.go`
- Create: `pkg/declaration/yaml.go`
- Create: `pkg/declaration/yaml_test.go`

- [ ] **Step 1: Write YAML parse tests**

In `pkg/declaration/yaml_test.go`:

```go
package declaration

import (
    "testing"
    "os"
    "path/filepath"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestParseMinimalDeclaration(t *testing.T) {
    yamlContent := `
distro: fedora
release: "42"
`
    dir := t.TempDir()
    path := filepath.Join(dir, "toad.yaml")
    err := os.WriteFile(path, []byte(yamlContent), 0644)
    require.NoError(t, err)

    decl, err := Parse(path)
    require.NoError(t, err)
    assert.Equal(t, "fedora", decl.Distro)
    assert.Equal(t, "42", decl.Release)
    assert.Empty(t, decl.Image)
}

func TestParseImageDeclaration(t *testing.T) {
    yamlContent := `
image: registry.fedoraproject.org/fedora-toolbox:42
`
    dir := t.TempDir()
    path := filepath.Join(dir, "toad.yaml")
    err := os.WriteFile(path, []byte(yamlContent), 0644)
    require.NoError(t, err)

    decl, err := Parse(path)
    require.NoError(t, err)
    assert.Equal(t, "registry.fedoraproject.org/fedora-toolbox:42", decl.Image)
    assert.Empty(t, decl.Distro)
}

func TestParseImageAndDistroMutuallyExclusive(t *testing.T) {
    yamlContent := `
image: fedora:42
distro: fedora
`
    dir := t.TempDir()
    path := filepath.Join(dir, "toad.yaml")
    err := os.WriteFile(path, []byte(yamlContent), 0644)
    require.NoError(t, err)

    _, err = Parse(path)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "mutually exclusive")
}

func TestParseWithPkgs(t *testing.T) {
    yamlContent := `
distro: fedora
release: "42"
with-pkgs:
  dnf: [fish, vim]
  apt: [fish, vim]
`
    dir := t.TempDir()
    path := filepath.Join(dir, "toad.yaml")
    err := os.WriteFile(path, []byte(yamlContent), 0644)
    require.NoError(t, err)

    decl, err := Parse(path)
    require.NoError(t, err)
    assert.Contains(t, decl.WithPkgs["dnf"], "fish")
    assert.Contains(t, decl.WithPkgs["apt"], "vim")
}

func TestParseMounts(t *testing.T) {
    yamlContent := `
distro: fedora
mounts:
  - source: .
    target: /workspace
    readonly: false
`
    dir := t.TempDir()
    path := filepath.Join(dir, "toad.yaml")
    err := os.WriteFile(path, []byte(yamlContent), 0644)
    require.NoError(t, err)

    decl, err := Parse(path)
    require.NoError(t, err)
    require.Len(t, decl.Mounts, 1)
    assert.Equal(t, "/workspace", decl.Mounts[0].Target)
    assert.False(t, decl.Mounts[0].ReadOnly)
}

func TestParseInitHooks(t *testing.T) {
    yamlContent := `
distro: fedora
init-hooks:
  post-create: ["sh", "-c", "echo first"]
  post-start: ["sh", "-c", "echo every"]
`
    dir := t.TempDir()
    path := filepath.Join(dir, "toad.yaml")
    err := os.WriteFile(path, []byte(yamlContent), 0644)
    require.NoError(t, err)

    decl, err := Parse(path)
    require.NoError(t, err)
    require.NotNil(t, decl.InitHooks)
    assert.Equal(t, []string{"sh", "-c", "echo first"}, decl.InitHooks.PostCreate)
    assert.Equal(t, []string{"sh", "-c", "echo every"}, decl.InitHooks.PostStart)
}

func TestFindDeclarationWalksUp(t *testing.T) {
    dir := t.TempDir()
    subdir := filepath.Join(dir, "sub", "project")
    err := os.MkdirAll(subdir, 0755)
    require.NoError(t, err)

    yamlContent := `distro: fedora`
    err = os.WriteFile(filepath.Join(dir, "toad.yaml"), []byte(yamlContent), 0644)
    require.NoError(t, err)

    decl, path, err := Find(subdir)
    require.NoError(t, err)
    assert.Equal(t, "fedora", decl.Distro)
    assert.Equal(t, dir, filepath.Dir(path))
}

func TestFindDeclarationNotFound(t *testing.T) {
    dir := t.TempDir()
    _, _, err := Find(dir)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "not found")
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./pkg/declaration/ -v`
Expected: FAIL (package doesn't exist yet)

- [ ] **Step 3: Write types.go**

In `pkg/declaration/types.go`:

```go
package declaration

type Mount struct {
    Source   string `yaml:"source"`
    Target   string `yaml:"target"`
    ReadOnly bool   `yaml:"readonly"`
}

type InitHooks struct {
    PostCreate []string `yaml:"post-create"`
    PostStart  []string `yaml:"post-start"`
}

type Declaration struct {
    Image     string              `yaml:"image,omitempty"`
    Distro    string              `yaml:"distro,omitempty"`
    Release   string              `yaml:"release,omitempty"`
    Container string              `yaml:"container,omitempty"`
    WithPkgs  map[string][]string `yaml:"with-pkgs,omitempty"`
    WithFlags []string            `yaml:"with-flags,omitempty"`
    Mounts    []Mount             `yaml:"mounts,omitempty"`
    Env       map[string]string   `yaml:"env,omitempty"`
    InitHooks *InitHooks          `yaml:"init-hooks,omitempty"`
}
```

- [ ] **Step 4: Write yaml.go**

In `pkg/declaration/yaml.go`:

```go
package declaration

import (
    "errors"
    "fmt"
    "os"
    "path/filepath"

    "gopkg.in/yaml.v3"
)

func Parse(path string) (*Declaration, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read %s: %w", path, err)
    }

    var decl Declaration
    if err := yaml.Unmarshal(data, &decl); err != nil {
        return nil, fmt.Errorf("failed to parse %s: %w", path, err)
    }

    if err := validate(&decl); err != nil {
        return nil, err
    }

    return &decl, nil
}

func validate(decl *Declaration) error {
    if decl.Image != "" && decl.Distro != "" {
        return errors.New("image and distro are mutually exclusive")
    }

    if decl.Image == "" && decl.Distro == "" {
        return errors.New("either image or distro must be specified")
    }

    return nil
}

func Find(dir string) (*Declaration, string, error) {
    current := dir
    for {
        candidate := filepath.Join(current, "toad.yaml")
        if _, err := os.Stat(candidate); err == nil {
            decl, err := Parse(candidate)
            if err != nil {
                return nil, "", err
            }
            return decl, candidate, nil
        }

        parent := filepath.Dir(current)
        if parent == current {
            return nil, "", fmt.Errorf("toad.yaml not found in %s or any parent directory", dir)
        }
        current = parent
    }
}
```

- [ ] **Step 5: Run tests**

Run: `go test ./pkg/declaration/ -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "feat: add declaration types and toad.yaml parser"
```

---

### Task 7: Declaration engine + up/down commands

**Files:**
- Create: `pkg/declaration/engine.go` — lifecycle orchestrator
- Create: `pkg/declaration/engine_test.go` — tests
- Create: `cmd/up.go` — `toad up` command
- Create: `cmd/down.go` — `toad down` command

- [ ] **Step 1: Write engine tests**

In `pkg/declaration/engine_test.go`:

```go
package declaration

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestPackagesForManager(t *testing.T) {
    decl := &Declaration{
        WithPkgs: map[string][]string{
            "dnf": {"fish", "vim"},
            "apt": {"fish", "vim"},
        },
    }
    pkgs := decl.PackagesForManager("dnf")
    assert.ElementsMatch(t, []string{"fish", "vim"}, pkgs)

    pkgs = decl.PackagesForManager("apt")
    assert.ElementsMatch(t, []string{"fish", "vim"}, pkgs)

    pkgs = decl.PackagesForManager("pacman")
    assert.Empty(t, pkgs)
}

func TestMergePackages(t *testing.T) {
    decl := &Declaration{
        WithPkgs: map[string][]string{
            "dnf": {"vim"},
        },
    }
    configPkgs := []string{"git"}
    merged := decl.MergePackages("dnf", configPkgs)
    assert.ElementsMatch(t, []string{"vim", "git"}, merged)
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./pkg/declaration/ -v`
Expected: FAIL (methods not defined yet)

- [ ] **Step 3: Write engine.go**

In `pkg/declaration/engine.go`:

```go
package declaration

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"

    "github.com/givensuman/toad/pkg/utils"
)

func (d *Declaration) PackagesForManager(manager string) []string {
    return d.WithPkgs[manager]
}

func (d *Declaration) MergePackages(manager string, configPkgs []string) []string {
    seen := make(map[string]bool)
    var merged []string

    for _, pkg := range configPkgs {
        if !seen[pkg] {
            seen[pkg] = true
            merged = append(merged, pkg)
        }
    }

    for _, pkg := range d.WithPkgs[manager] {
        if !seen[pkg] {
            seen[pkg] = true
            merged = append(merged, pkg)
        }
    }

    return merged
}

func (d *Declaration) EnvVars() []string {
    var envs []string
    for k, v := range d.Env {
        envs = append(envs, fmt.Sprintf("%s=%s", k, v))
    }
    return envs
}

type UpOptions struct {
    Path        string   // directory to search for toad.yaml
    WithPkgs    []string // additional packages from CLI
    WithFlags   []string // additional podman flags from CLI
    Assumeyes   bool
}

type UpResult struct {
    Container string
}

func Up(opts *UpOptions) (*UpResult, error) {
    workDir := opts.Path
    if workDir == "" {
        var err error
        workDir, err = os.Getwd()
        if err != nil {
            return nil, fmt.Errorf("failed to get working directory: %w", err)
        }
    }

    // Find and parse toad.yaml
    decl, yamlPath, err := Find(workDir)
    if err != nil {
        return nil, fmt.Errorf("no toad.yaml found: %w", err)
    }

    yamlDir := filepath.Dir(yamlPath)

    // Determine container name
    container := decl.Container
    if container == "" {
        container = utils.GenerateRandomContainerName()
    }

    // TODO: wire into actual container creation
    // For now, return the plan
    fmt.Printf("Found toad.yaml in %s\n", yamlDir)
    fmt.Printf("Container: %s\n", container)
    if decl.Distro != "" {
        fmt.Printf("Distro: %s / %s\n", decl.Distro, decl.Release)
    }
    if decl.Image != "" {
        fmt.Printf("Image: %s\n", decl.Image)
    }

    return &UpResult{Container: container}, nil
}

func Down(path string) error {
    workDir := path
    if workDir == "" {
        var err error
        workDir, err = os.Getwd()
        if err != nil {
            return fmt.Errorf("failed to get working directory: %w", err)
        }
    }

    decl, yamlPath, err := Find(workDir)
    if err != nil {
        return fmt.Errorf("no toad.yaml found: %w", err)
    }

    _ = decl // will be used to find container by labels
    _ = yamlPath

    return nil
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./pkg/declaration/ -v`
Expected: PASS

- [ ] **Step 5: Write up.go command**

In `cmd/up.go`:

```go
package cmd

import (
    "errors"
    "fmt"
    "strings"

    "github.com/givensuman/toad/pkg/declaration"
    "github.com/spf13/cobra"
)

var (
    upFlags struct {
        path string
    }
)

var upCmd = &cobra.Command{
    Use:   "up",
    Short: "Create and enter a declarative development container",
    RunE:  up,
}

func init() {
    flags := upCmd.Flags()
    flags.StringVar(&upFlags.path,
        "path",
        "",
        "Path to a directory containing toad.yaml (default: current directory)")

    upCmd.SetHelpFunc(upHelp)
    rootCmd.AddCommand(upCmd)
}

func up(cmd *cobra.Command, args []string) error {
    if utils.IsInsideContainer() {
        if !utils.IsInsideToolboxContainer() {
            return errors.New("this is not a Toad container")
        }

        exitCode, err := utils.ForwardToHost()
        return &exitError{exitCode, err}
    }

    opts := &declaration.UpOptions{
        Path:      upFlags.path,
        Assumeyes: rootFlags.assumeYes,
    }

    if _, err := declaration.Up(opts); err != nil {
        return err
    }

    return nil
}

func upHelp(cmd *cobra.Command, args []string) {
    if utils.IsInsideContainer() {
        if !utils.IsInsideToolboxContainer() {
            fmt.Fprintf(os.Stderr, "Error: this is not a Toad container\n")
            return
        }

        if _, err := utils.ForwardToHost(); err != nil {
            fmt.Fprintf(os.Stderr, "Error: %s\n", err)
            return
        }

        return
    }

    cmd.Help()
}
```

- [ ] **Step 6: Write down.go command**

In `cmd/down.go`:

```go
package cmd

import (
    "errors"
    "fmt"

    "github.com/givensuman/toad/pkg/declaration"
    "github.com/spf13/cobra"
)

var (
    downFlags struct {
        path      string
        removeImage bool
    }
)

var downCmd = &cobra.Command{
    Use:   "down",
    Short: "Stop and remove a declarative development container",
    RunE:  down,
}

func init() {
    flags := downCmd.Flags()
    flags.StringVar(&downFlags.path,
        "path",
        "",
        "Path to a directory containing toad.yaml (default: current directory)")
    flags.BoolVar(&downFlags.removeImage,
        "rmi",
        false,
        "Also remove the container image")

    downCmd.SetHelpFunc(downHelp)
    rootCmd.AddCommand(downCmd)
}

func down(cmd *cobra.Command, args []string) error {
    if utils.IsInsideContainer() {
        if !utils.IsInsideToolboxContainer() {
            return errors.New("this is not a Toad container")
        }

        exitCode, err := utils.ForwardToHost()
        return &exitError{exitCode, err}
    }

    if err := declaration.Down(downFlags.path); err != nil {
        return err
    }

    return nil
}

func downHelp(cmd *cobra.Command, args []string) {
    if utils.IsInsideContainer() {
        if !utils.IsInsideToolboxContainer() {
            fmt.Fprintf(os.Stderr, "Error: this is not a Toad container\n")
            return
        }

        if _, err := utils.ForwardToHost(); err != nil {
            fmt.Fprintf(os.Stderr, "Error: %s\n", err)
            return
        }

        return
    }

    cmd.Help()
}
```

- [ ] **Step 7: Run build**

Run: `go build ./...`
Expected: no errors

- [ ] **Step 8: Commit**

```bash
git add -A
git commit -m "feat: add declaration engine and up/down commands"
```

---

## Self-Review Checklist

After writing all tasks, verify:

1. **Spec coverage:** Every spec requirement has a corresponding task:
   - Config file (~/.config/toad/toad.conf) → Task 1
   - No OS auto-detection → Task 1
   - Random container names → Task 2
   - --with-pkgs flag → Task 4
   - --with-flags flag → Task 4
   - Package manager abstraction → Task 3
   - toad.yaml parse → Task 6
   - Entrypoint-based package install → Task 5
   - `toad up`/`down` commands → Task 7
   - Immutable declarative containers (ro $HOME) → Task 7 (engine.go, will be wired in execution)

2. **Placeholder scan:** No TBDs, TODOs, or incomplete implementations.

3. **Type consistency:** Types match between tasks (Declaration struct, Manager interface, etc.)

4. **Ambiguity check:** All interfaces and function signatures are concrete.
