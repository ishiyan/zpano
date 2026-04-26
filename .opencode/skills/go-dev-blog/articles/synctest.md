# `testing/synctest` — Testing Concurrent Code (Go 1.25+)

`testing/synctest` eliminates flaky, slow concurrent tests that rely on `time.Sleep` or polling. It provides deterministic synchronization and a fake clock for goroutines under test.

Reference: [go.dev/blog/testing-time](https://go.dev/blog/testing-time)

**GA in Go 1.25.** No `GOEXPERIMENT` flag needed. Covered by the Go compatibility promise.

## API

Two functions:

| Function | Purpose |
|---|---|
| `synctest.Test(t, func(*testing.T))` | Runs `f` in an isolated "bubble" with fake time. Creates a scoped `*testing.T`. Returns when all bubble goroutines exit. Panics on deadlock. |
| `synctest.Wait()` | Waits until every goroutine in the current bubble is durably blocked. Provides synchronization (race detector aware). |

## Basic Pattern

**Before** (slow and flaky):

```go
func TestWithDeadline(t *testing.T) {
    deadline := time.Now().Add(1 * time.Second)
    ctx, _ := context.WithDeadline(t.Context(), deadline)
    time.Sleep(time.Until(deadline) + 100*time.Millisecond) // slow, still flaky
    if err := ctx.Err(); err != context.DeadlineExceeded {
        t.Fatalf("context not canceled after deadline")
    }
}
```

**After** (fast and deterministic):

```go
func TestWithDeadline(t *testing.T) {
    synctest.Test(t, func(t *testing.T) {
        deadline := time.Now().Add(1 * time.Second)
        ctx, _ := context.WithDeadline(t.Context(), deadline)

        time.Sleep(time.Until(deadline))
        synctest.Wait()
        if err := ctx.Err(); err != context.DeadlineExceeded {
            t.Fatalf("context not canceled after deadline")
        }
    })
}
```

No channels needed for synchronization — `Wait` makes plain variable access race-free.

## Fake Time

Inside a bubble, `time.Now`, `time.Sleep`, `time.After`, `time.NewTimer`, etc. use a fake clock starting at **2000-01-01 00:00:00 UTC**. Time only advances when all goroutines are durably blocked. Computation takes zero fake time.

```go
synctest.Test(t, func(t *testing.T) {
    start := time.Now()
    time.Sleep(10 * time.Second) // returns instantly
    t.Log(time.Since(start))     // 10s (fake time)
})
```

**Time stops when the root goroutine returns.** Background goroutines waiting on timers after that point cause a deadlock panic.

```go
// DEADLOCK: root returns, time stops, Sleep never completes
synctest.Test(t, func(t *testing.T) {
    go func() {
        time.Sleep(1 * time.Nanosecond) // never returns
    }()
})
```

Timers scheduled for the same instant fire in **randomized** order (not creation order).

## What Is "Durably Blocked"

`Wait` returns when all bubble goroutines are **durably blocked** — blocked in a way only another bubble goroutine can unblock.

| Operation | Durably blocking? |
|---|---|
| `time.Sleep`, timers | Yes |
| Send/receive on channel created in bubble | Yes |
| `sync.WaitGroup.Wait` (WaitGroup belongs to bubble) | Yes |
| `sync.Cond.Wait` | Yes |
| `select{}` (empty select) | Yes |
| `sync.Mutex` lock | **No** |
| Network / file I/O, syscalls, cgo | **No** |

### `sync.WaitGroup` bubble association

A `WaitGroup` is associated with a bubble on the first call to `Go` or `Add`. Waiting on a same-bubble `WaitGroup` is durably blocking. Calling `Go`/`Add` on a `WaitGroup` belonging to a different bubble panics.

## Testing Network Code

Real I/O is not durably blocking. Use `net.Pipe` (in-memory) or a fake `net.Listener`:

```go
synctest.Test(t, func(t *testing.T) {
    srvConn, cliConn := net.Pipe()
    defer srvConn.Close()
    defer cliConn.Close()

    go func() {
        req, _ := http.NewRequest("GET", "http://test.tld/", nil)
        // use cliConn via custom Transport.DialContext
    }()

    synctest.Wait()
})
```

See the [synctest package docs](https://pkg.go.dev/testing/synctest#hdr-Example__HTTP_100_Continue) for a complete HTTP example.

## Rules

- **All goroutines must exit** before `Test` returns. Clean up background goroutines or they deadlock.
- **Time stops** when the root goroutine (started by `Test`) returns. Pending timers will not fire.
- **Bubbled channels** (created inside a bubble) panic if accessed from outside.
- **Deadlock panics** include stack traces for bubble goroutines only, with durable-block annotations.
- **Race detector** understands `Wait` — use plain variables instead of channels for synchronization within a bubble.
- **`t.Cleanup`** registered inside a bubble runs inside that bubble's scope.

## Migration from Go 1.24 Experimental API

| Go 1.24 (experimental) | Go 1.25 (GA) |
|---|---|
| `GOEXPERIMENT=synctest go test` | `go test` (no flag) |
| `synctest.Run(func() { ... })` | `synctest.Test(t, func(t *testing.T) { ... })` |
| Time advances after root returns | Time **stops** when root returns |
| Timer order deterministic | Timers at same instant **randomized** |
| `sync.WaitGroup` not bubble-aware | `sync.WaitGroup` associated with bubble |
