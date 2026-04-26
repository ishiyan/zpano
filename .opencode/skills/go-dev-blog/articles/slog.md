# Structured Logging with `log/slog` (Go 1.21+)

`log/slog` brings structured, leveled logging to the standard library. It provides a common framework that third-party logging packages can share via the `Handler` interface.

Reference: [go.dev/blog/slog](https://go.dev/blog/slog)

## Quick Start

```go
slog.Info("request handled", "method", r.Method, "path", r.URL.Path, "status", 200)
// 2023/08/04 16:09:19 INFO request handled method=GET path=/ status=200
```

## Levels

| Level | Int | Usage |
|---|---|---|
| `slog.LevelDebug` | -4 | Verbose development info |
| `slog.LevelInfo` | 0 | Normal events |
| `slog.LevelWarn` | 4 | Potential issues |
| `slog.LevelError` | 8 | Errors requiring attention |

Levels are integers — use any value between named levels for custom levels.

## Handlers

| Handler | Output format |
|---|---|
| Default (via `log`) | `2023/08/04 INFO msg key=value` |
| `slog.NewTextHandler(w, opts)` | `time=... level=INFO msg=... key=value` |
| `slog.NewJSONHandler(w, opts)` | `{"time":"...","level":"INFO","msg":"...","key":"value"}` |

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))
slog.SetDefault(logger)
```

## Key Patterns

### Key-value pairs (convenient)

```go
slog.Info("msg", "user", name, "id", id)
```

### Typed attrs (performant, allocation-free)

```go
slog.LogAttrs(ctx, slog.LevelInfo, "msg",
    slog.String("user", name),
    slog.Int("id", id),
)
```

### Logger.With (factor out common attrs)

```go
logger := slog.Default().With("service", "auth", "version", "1.2")
logger.Info("started")  // includes service= and version= automatically
```

### Groups (structured nesting)

```go
slog.Info("request", slog.Group("req", "method", "GET", "path", "/"))
// JSON: {"req":{"method":"GET","path":"/"}}
```

### LogValuer (custom type rendering)

Implement `slog.LogValuer` on your types to control how they appear in logs (e.g., redacting secrets, logging struct fields as groups).

## Writing a Custom Handler

Implement `slog.Handler`:

```go
type Handler interface {
    Enabled(context.Context, Level) bool
    Handle(context.Context, Record) error
    WithAttrs([]Attr) Handler
    WithGroup(string) Handler
}
```

- `Enabled` — fast-path to skip disabled levels.
- `WithAttrs`/`WithGroup` — pre-format attrs added via `Logger.With` for performance.

## Testing

Use `testing/slogtest` (Go 1.21+) to verify custom handler correctness.

## Pitfalls

1. **Mismatched key-value pairs** — odd number of args after message creates a `!BADKEY`. Use `go vet` to catch.
2. **Context parameter** — pass `context.Context` to `slog.InfoContext` etc. for trace ID propagation. Canceling ctx does NOT prevent logging.
3. **Don't ignore `Enabled`** — expensive attribute computation should be guarded by checking the level first.
