# Toad Codebase Issues & Remediation Plan

Current status: `go build ./...` passes, `go vet ./...` passes, all unit tests pass.

---

## Critical Issues

### 1. Non-functional Declarative Engine (the project's flagship feature)
**Files:** `pkg/declaration/engine.go`, `cmd/up.go`, `cmd/down.go`

`toad up` and `toad down` are stubs. `Up()` finds a `toad.yaml` and prints "Container 'X' is ready" — but never calls `podman create`, `podman start`, or `podman run/exec`. `Down()` resolves the container name but never calls `podman stop` or `podman rm`. The entire declarative lifecycle is unimplemented.

**Impact:** The project's primary differentiator from Toolbx doesn't work. `toad up` is misleading — it claims the container is ready when nothing was created.

**Fix:** Implement the full lifecycle:
- `Up()`: `podman create` with image, mounts, env vars, labels, flags from toad.yaml → `podman start` → wait for init → `podman exec` to enter
- `Down()`: `podman stop` → `podman rm` → optional `podman rmi`

---

## High-Priority Issues

### 2. Incomplete Rebranding from Toolbx
**Files:** multiple

The fork from Toolbx is incomplete. Key remnants:
- **`cmd/create.go:412`**: Entry point is hardcoded `"toolbox"` — should use `executableBase`
- **`cmd/run.go:227`**: Entry point check expects `"toolbox"` — won't match `"toad"` containers
- **`cmd/run.go:394`**: Fallback command is `"toolbox"`
- **`cmd/create.go:446-448`**: Labels use both `com.github.containers.toolbox=true` and `com.github.givensuman.toad=true`
- **`cmd/root.go:210`**: Config dir is `~/.config/toolbox` — should be `~/.config/toad`
- **CI workflows**: Push to `quay.io/toolbx` instead of a toad registry
- **`pkg/utils/utils.go:499`**: Runtime directory uses `"toad"` (correct) but comment references `toolbox`

**Impact:** Confuses users, breaks entry point detection, references old project.

### 3. Declarative Engine `UpOptions`/`DownOptions` Not Connected to Podman
**Files:** `cmd/up.go`, `cmd/down.go`, `pkg/declaration/engine.go`

Even the stub interface is incomplete. The `UpOptions` and `DownOptions` don't propagate podman create args, mounts, env vars, or packages from the declaration to actual podman commands. The `MergePackages`, `EnvVars` methods on `Declaration` exist but are never called by `Up()`/`Down()`.

### 4. Hardcoded Container Name in `create.go` Entry Point
**File:** `cmd/create.go:412`

```go
entryPoint := []string{
    "toolbox", "--log-level", "debug",
    "init-container",
```
This should be `executableBase` (which resolves to `"toad"` at runtime). Using a hardcoded `"toolbox"` means the binary must be named `toolbox` or symlinked for init-container to work.

---

## Medium-Priority Issues

### 5. Race Condition Window in Container Initialization
**File:** `cmd/run.go:544-693`

`ensureContainerIsInitialized` has a window between when the watcher is set up and when the initialization stamp is checked. If the container initializes in this gap, the function may time out with `"failed to initialize container"` even though initialization succeeded. The `PathExists` check at line 634 partially mitigates this, but the watcher setup races with the container's init process.

### 6. Potential Goroutine Leak in `ensureContainerIsInitialized`
**File:** `cmd/run.go:642-690`

The `followEntryPointLogsAsync` goroutine may leak if the ticker timeout fires (`initializedTimeout.C`) without proper cancellation of the logs goroutine. The `logsCancel` is called but the goroutine reading from `reader` may block.

### 7. Complex Terminal Prompt Logic with Race Conditions
**File:** `cmd/create.go:751-946`

`showPromptForDownload` has a two-phase async design (`showPromptForDownloadFirst` → `showPromptForDownloadSecond`) with concurrent goroutines for image size fetching and user input. The context cancellation chain is fragile and could leak goroutines or miss cleanup on error paths. The terminal state management (raw mode toggle, cursor save/restore) adds complexity.

### 8. `configurePKCS11` Redundant Code Paths
**File:** `cmd/initContainer.go:603-666`

```go
if ok, err := utils.IsP11KitClientPresent(); err != nil {
    // ... if !ok { return nil }
} else {
    // ... if !ok { return nil }  // identical code
}
```
Both the error and success branches handle `!ok` identically. This is confusing and suggests a logic issue — if `err != nil` but `ok == true`, the function falls through with potentially inconsistent state.

### 9. `SetUpConfiguration` Calls Heavy `ResolveContainerAndImageNames`
**File:** `pkg/utils/utils.go:676-683`

During config setup (which runs on every command), `ResolveContainerAndImageNames` is called to determine the default container name. This does distro resolution, image name parsing, and release detection — all for a value that rarely changes. Should cache or lazy-init.

### 10. `Flock` Raw File Descriptor Handling
**File:** `pkg/utils/utils.go:221-236`

Creates a lock file, gets its raw fd, and calls `syscall.Flock`. The `*os.File` finalizer could close the fd. Also doesn't handle `EINTR` retry for `syscall.Flock`.

### 11. `mountBind` Silently Skips Missing Sources
**File:** `cmd/initContainer.go:1021-1071`

When the source path doesn't exist, `mountBind` returns `nil` silently. While intentional for optional host paths, this can mask real configuration errors (e.g., a user-defined mount in toad.yaml with a wrong path).

---

## Low-Priority / Cleanup Issues

### 12. `path` vs `filepath` Inconsistency
**File:** `pkg/utils/utils.go:499`

Uses `path.Join` instead of `filepath.Join`. Works on Linux but inconsistent with the rest of the codebase.

### 13. `init()` Calls Non-Trivial Operations
**File:** `cmd/root.go:92-96`

`init()` calls `setUpGlobals()` which does `user.Current()`, `os.Executable()`, `filepath.EvalSymlinks()`, `os.Getwd()`, `utils.GetCgroupsVersion()`. This runs at package init time, even for `completion` and `--help` commands. Failures cause `os.Exit(1)` instead of clean error handling.

### 14. Selective Error Swallowing in `pullImage`
**File:** `cmd/create.go:657-662`

`ImageReferenceCanBeID` check silently falls through if `ImageExists` fails. Works as designed but could be confusing during debugging.

### 15. Test Coverage Gaps
- `pkg/nvidia` — no tests at all
- `pkg/skopeo` — no tests
- `pkg/version` — no tests
- `cmd/` — only tests `exitError` type
- No integration tests for the declarative engine

---

## Remediation Plan

### Phase 1: Fix Critical & High-Priority Bugs (~2 days)

1. **Fix entry point in `create.go`**: Replace hardcoded `"toolbox"` with `executableBase`
2. **Fix entry point check in `run.go`**: Accept both `"toolbox"` and `"toad"` as valid entry points
3. **Fix fallback command in `run.go`**: Use `executableBase` for fallback
4. **Fix config path in `root.go`**: Use `~/.config/toad` instead of `~/.config/toolbox`

### Phase 2: Implement Declarative Engine (~1 week)

1. Implement `Up()` lifecycle: find toad.yaml → create container with declaration settings → start container → wait for init → enter
2. Implement `Down()` lifecycle: find toad.yaml → stop container → remove container → optional rmi
3. Wire `UpOptions`/`DownOptions` to propagate mounts, env, packages, and flags
4. Add unit tests for `Up()` and `Down()`

### Phase 3: Fix Race Conditions & Goroutine Leaks (~1 day)

1. Fix initialization window race in `ensureContainerIsInitialized`
2. Fix goroutine leak in `followEntryPointLogsAsync` cancellation
3. Add `EINTR` retry to `Flock`
4. Clean up `configurePKCS11` redundant code paths

### Phase 4: Rebranding & Cleanup (~1 day)

1. Update CI workflows from `quay.io/toolbx` to toad registry
2. Update container labels to prefer `com.github.givensuman.toad=true`
3. Fix `path` → `filepath` in `utils.go`
4. Move heavy operations out of `init()`
5. Add tests for untested packages

### Phase 5: Simplify Complex Code (~1 day)

1. Simplify `showPromptForDownload` two-phase terminal prompt
2. Add logging to `mountBind` when source path doesn't exist
3. Cache `ResolveContainerAndImageNames` in config setup
