---
description: 'Guidance on os.Root for traversal-resistant file access in Go 1.24+'
applyTo: '**/*.go'
---

# `os.Root` — Traversal-Resistant File Access (Go 1.24+)

Use `os.Root` or `os.OpenInRoot` when opening files in a directory where the filename is untrusted. This prevents path traversal attacks (`../`, symlinks, Windows device names) at the OS level, without TOCTOU races.

Reference: [go.dev/blog/osroot](https://go.dev/blog/osroot)

## Rule of Thumb

Code that calls `filepath.Join` with an externally-provided filename should use `os.Root` instead:

```go
// VULNERABLE — attacker can escape baseDirectory
f, err := os.Open(filepath.Join(baseDirectory, filename))

// SAFE — confined to baseDirectory
f, err := os.OpenInRoot(baseDirectory, filename)
```

## When to Use

Use `os.Root` / `os.OpenInRoot` when:
- Opening a file in a directory, AND
- The filename comes from untrusted input (user input, archive entries, network data)

Do NOT use when:
- The filename is fully trusted (e.g., a user-specified CLI output path that may refer to anywhere)

## API

### Quick: `os.OpenInRoot`

One-shot open of an untrusted filename within a directory:

```go
f, err := os.OpenInRoot("/some/root", untrustedFilename)
```

### Full: `os.Root`

For multiple operations within a directory:

```go
root, err := os.OpenRoot("/some/root")
if err != nil {
    return err
}
defer root.Close()

f, err := root.Open("path/to/file")           // read
f, err  = root.Create("path/to/new")          // create
f, err  = root.OpenFile("path", flags, perm)  // open with flags
info, err := root.Stat("path")                // stat (follows symlinks)
info, err  = root.Lstat("path")               // stat (no follow)
err = root.Mkdir("dir", 0o755)                // mkdir
err = root.Remove("path")                     // remove
sub, err := root.OpenRoot("subdir")           // nested root
```

## What It Prevents

| Attack | Defense |
|---|---|
| `../../etc/passwd` (relative escape) | Blocked — `..` cannot escape root |
| Symlink pointing outside root | Blocked — symlinks resolved within root |
| TOCTOU race (symlink created between check and open) | Blocked — uses `openat` syscalls on Unix |
| Windows device names (`NUL`, `COM1`, `CONOUT$`) | Blocked |

Symlinks and `..` that stay within the root are allowed.

## Path Sanitization (When Root Is Overkill)

When the threat model excludes attackers with filesystem access, validate paths before use:

| Function | Since | Purpose |
|---|---|---|
| `filepath.IsLocal(path)` | Go 1.20 | Reports if path is local (no `..` escape, not absolute, not empty, no Windows reserved names) |
| `filepath.Localize(path)` | Go 1.23 | Converts `/`-separated path to a safe local OS path |

Use these for untrusted filenames when you don't need the full protection of `os.Root`.

## Platform Caveats

| Platform | Notes |
|---|---|
| **Unix** | Uses `openat` syscalls. Defends against symlinks but not bind mounts (requires root to create). `Root` holds an fd — tracks directory across renames. |
| **Windows** | Holds an open handle — prevents root directory rename/deletion. Blocks reserved device names. |
| **WASI** | Relies on WASI preview 1 sandboxing; protection depends on the runtime implementation. |
| **GOOS=js** | Vulnerable to TOCTOU races (no `openat`). References directory by name, not fd. |

## Performance

Paths with many directory components or `..` segments are more expensive under `Root` than plain `os.Open`. For hot paths, pre-clean input with `filepath.Clean` to remove `..` and limit directory depth.
