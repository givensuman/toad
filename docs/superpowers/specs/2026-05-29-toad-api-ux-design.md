# Toad API & UX Design

> Design-it-fresh approach. Optimize for declarative-first workflow while keeping backward compatibility for ad-hoc operations.

## CLI Command Tree

```
toad create [--name=] [--distro|--image] [--release] [--with-pkgs] [--with-flags] [container]
    Create an ad-hoc container. Container name: positional, then --name, then random.

toad up [--path=.] [--open=<command>]
    Declarative: find toad.yaml, create container, start it, enter it.
    --open runs a command instead of opening a shell.

toad down [--path=.] [--rmi]
    Declarative: find toad.yaml, stop + remove container. Optionally remove image.

toad enter [--name=] [container]
    Enter a container for interactive use. Container: positional, then --name, then default.

toad run [--name=] [container] <command>
    Run a command in a container. Container: positional, then --name.

toad ls [--containers|--images]
    List containers and/or images. Renamed from list for brevity.

toad rm [--all] [--force] <container...>
    Remove containers.

toad rmi [--all] [--force] <image...>
    Remove images.

toad inspect <container>
    Show detailed container info (ip, mounts, env, labels, etc.).

toad init [--path=.]
    Scaffold a starter toad.yaml in the target directory.

toad completion [bash|fish|zsh]
    Shell completion.

toad help [command]
    Help (Cobra built-in).
```

## Flag Standardization

| Short | Long | Commands | Notes |
|-------|------|----------|-------|
| `-n` | `--name` | create, enter, run | Container name (replaces `-c`) |
| `-d` | `--distro` | create | Distribution |
| `-i` | `--image` | create | Image (removed from `ls`) |
| `-r` | `--release` | create | OS release |
| `-p` | `--path` | up, down, init | Path to toad.yaml directory |
| `-a` | `--all` | rm, rmi | Remove all |
| `-f` | `--force` | rm, rmi | Force removal |
| `-c` | `--containers` | ls | List only containers (now conflict-free) |
| | `--images` | ls | List only images (no short flag) |
| | `--rmi` | down | Also remove image |
| | `--with-pkgs` | create | Packages to install (unchanged) |
| | `--with-flags` | create | Extra podman create flags (unchanged) |

## Workflow UX

### Progress Indicators
Spinner visibility decoupled from log level — always show spinners unless explicitly quieted. Use `--quiet`/`-q` global flag to suppress all non-error output.

### Download Prompt Simplification
Replace the two-phase async prompt (200+ lines, goroutines, eventfd, poll, raw terminal, cursor save/restore) with a single blocking prompt:
```
Image required to create Toad container.
Download registry.fedoraproject.org/fedora-toolbox:42? [y/N]:
```
No image size pre-fetch. No terminal gymnastics. Async complexity removed.

### `toad init` UX
```
$ toad init
Created toad.yaml in /home/user/project
Edit it, then run 'toad up' to create your dev container.
```

Default toad.yaml:
```yaml
distro: fedora
release: "42"
```

### `toad up` Status Clarity
Step-by-step feedback:
```
$ toad up
Found toad.yaml in /home/user/project
Pulling registry.fedoraproject.org/fedora-toolbox:42... done
Creating container happy_mclean... done
Starting container happy_mclean... done
Connecting to happy_mclean...
```

### `toad.yaml` Discovery
Show which file was used: `Found toad.yaml in /home/user/project`.

## Config & Declarations

### Config (~/.config/toad/toad.conf)
TOML format (unchanged). Simple key-value, user-only, no system-wide config.
```toml
[general]
distro = "fedora"
release = "42"
```

### Declarations (toad.yaml)
YAML format (unchanged). Complex nested structures for packages, mounts, hooks, env vars.

## Error Messages

Standard format:
```
Error: <message>
Run 'toad <command> --help' for usage.
```

Consistent across all commands. All "toolbox" references removed from user-facing strings. Error types specific: container not found, image not found, distro unsupported, etc.

## Implementation Priorities

1. **Flag standardization** — Rename `-c` → `-n`, remove `-i` from `ls`, add `--color=none` to `ls` output
2. **`toad inspect`** — New command wrapping `podman inspect`
3. **`toad init`** — New command, writes starter toad.yaml
4. **`list` → `ls`** — Rename, add alias for backward compat
5. **Download prompt simplification** — Replace async prompt with blocking prompt
6. **Progress indicators** — Decouple spinner from log level
7. **Status clarity** — Step-by-step output for up/down
8. **Error message standardization** — Consistent format across all commands
