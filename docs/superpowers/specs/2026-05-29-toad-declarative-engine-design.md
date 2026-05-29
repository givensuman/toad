# Toad: Declarative Container Engine Design

## Overview

Toad is a CLI tool for creating and managing Podman-based development containers,
forked from Toolbx. Its key differentiator is **declarative-first workflows**:
containers defined by `toad.yaml` are immutable, reproducible, and treat the
YAML file as the single source of truth.

## Architecture

```
cmd/                  # Cobra command definitions (thin wrappers)
  ├── root.go         # global flags, preRun, migration
  ├── create.go       # ad-hoc container creation (enhanced Toolbx flow)
  ├── enter.go        # enter existing container
  ├── run.go          # run command in container
  ├── list.go         # list containers
  ├── rm.go / rmi.go  # remove containers/images
  ├── up.go           # NEW: declarative container lifecycle
  └── down.go         # NEW: stop + remove declarative container
pkg/
  ├── declaration/    # NEW: declaration engine
  │   ├── types.go    # Declaration, Mount, Hook structs
  │   ├── yaml.go     # parse + validate toad.yaml
  │   └── engine.go   # lifecycle orchestrator (up/down)
  ├── pkgmanager/     # NEW: package manager abstraction
  │   ├── manager.go  # Manager interface
  │   ├── dnf.go      # DNF implementation
  │   ├── apt.go      # APT implementation
  │   └── pacman.go   # Pacman implementation
  ├── podman/         # existing: Podman interaction layer
  ├── shell/          # existing: shell command execution
  ├── utils/          # existing: utilities (modified)
  └── ...             # nvidia, skopeo, term, version (unchanged)
```

## Config System

No OS auto-detection. Configuration via Viper (TOML):

**File:** `~/.config/toad/toad.conf` (user only, no system-wide config)

```toml
[general]
distro = "fedora"
release = "42"

[packages.dnf]
extra = ["fish", "vim", "golang"]

[packages.apt]
extra = ["fish", "vim", "golang"]

[packages.pacman]
extra = ["fish", "vim", "go"]

[podman.flags]
extra = ["--shm-size=2g"]
```

- `distro` / `release` set defaults (no host OS detection)
- CLI flags (`--distro`, `--release`) override config values
- Config is optional — all fields have CLI flag equivalents

## Declaration Files (toad.yaml)

Discovered by walking up from cwd. Single source of truth for declarative containers.

```yaml
# image OR distro — mutually exclusive
# image: registry.fedoraproject.org/fedora-toolbox:42
distro: fedora
release: "42"

# optional: explicit name, else random via namesgenerator
# container: my-dev-env

with-pkgs:
  dnf: [fish, vim, golang, libseccomp-devel]
  apt: [fish, vim, golang, libseccomp-dev]
  pacman: [fish, vim, go, libseccomp]

with-flags:
  - --shm-size=2g

mounts:
  - source: .
    target: /workspace
    readonly: false

env:
  EDITOR: vim
  FOO: bar

init-hooks:
  post-create: ["sh", "-c", "echo 'first boot'"]
  post-start: ["sh", "-c", "echo 'every boot'"]
```

### Validation Rules

- `image` and `distro` are mutually exclusive (validated at parse time)
- `release` is required when `distro` is specified and the distro requires it
- All paths are validated for existence before container creation
- `source: .` resolves to the directory containing toad.yaml

### Container Labels (Declarative Mode)

```
com.github.givensuman.toad=true
com.github.givensuman.toad.declaration=<sha256 of toad.yaml>
com.github.givensuman.toad.workdir=<path to toad.yaml directory>
```

## Package Manager Abstraction

```go
type Manager interface {
    Name() string
    Install(pkgs []string) []string
    UpdateDB() []string
    ListInstalled() []string
    Query(pkg string) []string
}
```

### Resolution

1. Inspect container image → determine distro
2. Map distro → package manager (fedora→dnf, ubuntu→apt, arch→pacman)
3. Merge packages: config pkgs + toad.yaml pkgs + `--with-pkgs` flag pkgs

### Supported Managers (initial)

| Distro | Manager | Install Command |
|--------|---------|-----------------|
| Fedora/RHEL | dnf | `dnf install -y <pkgs>` |
| Debian/Ubuntu | apt | `apt-get install -y <pkgs>` |
| Arch | pacman | `pacman -S --noconfirm <pkgs>` |

## Declaration Engine

### `toad up` Lifecycle

1. Walk up from cwd → find nearest `toad.yaml`
2. Parse & validate declaration
3. Determine container name: from yaml `container` field, or generate random via `github.com/givensuman/namesgenerator`
4. Create container with `podman create`:
   - Resolved image (from `image` or `distro`+`release`)
   - Custom mounts from declaration (project dir mounted rw)
   - Env vars from declaration
   - Extra Podman flags from `with-flags`
   - Labels marking it as declarative
   - **$HOME and system dirs mounted read-only**
   - Only the project directory (target of `source: .`) is rw
5. Start container
6. Pass install/hook args to init-container via env vars:
   - `TOAD_INSTALL_PKGS` — serialized manager+pkg list
   - `TOAD_POST_CREATE_HOOK` — command for first-boot hook
   - `TOAD_POST_START_HOOK` — command for every-boot hook
7. Init-container runs (enhanced):
   - Reads `TOAD_INSTALL_PKGS` → `UpdateDB()` → `Install(pkgs)`
   - Runs `post-create` hook on first boot
   - Runs `post-start` hook on every boot
   - Creates initialization stamp (existing flow)
8. Enter container (like `enter`)

### `toad down` Lifecycle

1. Find container matching current directory's toad.yaml
2. Stop container
3. Remove container
4. Optional: remove image (flag)

### Ad-hoc vs Declarative Behavior

| Aspect | Ad-hoc (`create`) | Declarative (`up`) |
|--------|------------------|---------------------|
| Name | deterministic (distro-release) | random (namesgenerator) or explicit |
| $HOME | rw (current Toolbx behavior) | ro |
| System dirs | rw | ro |
| Project dir | not mounted | mounted rw |
| Packages | via --with-pkgs flag | from toad.yaml |
| Source of truth | CLI flags | toad.yaml |

## Init-Container Enhancements

The existing `init-container` entrypoint is extended to support:

- **`TOAD_INSTALL_PKGS`** env var — contains base64-encoded JSON of `{manager: string, packages: []string}`. If set, runs package installation after standard init steps but before creating the initialization stamp.
- **`TOAD_POST_CREATE_HOOK`** — runs only on first container start. Persisted via a sentinel file in the container.
- **`TOAD_POST_START_HOOK`** — runs on every container start.

## Command Reference

```
toad create [--distro|--image] [--release] [--with-pkgs] [--with-flags]
    Create an ad-hoc container (no toad.yaml required).
    Enhanced with --with-pkgs and --with-flags flags.

toad enter [container]
    Enter an existing container. Unchanged from Toolbx.

toad run [--container] <command>
    Run a command in an existing container. Unchanged from Toolbx.

toad list
    List containers. Unchanged from Toolbx.

toad rm <container>
    Remove a container. Unchanged from Toolbx.

toad rmi <image>
    Remove an image. Unchanged from Toolbx.

toad up [--path=.]
    Declarative: find toad.yaml, create+enter container.

toad down [--path=.]
    Declarative: stop+remove container from toad.yaml.
```

## Migration from Toolbx

- Existing Toolbx config (`~/.config/containers/toolbox.conf`) is **not** migrated
- Users set up `~/.config/toad/toad.conf` independently
- Existing Toolbx containers still work via `enter` / `run` / `rm`
- `create` retains Toolbx-compatible behavior (deterministic names, rw $HOME)
