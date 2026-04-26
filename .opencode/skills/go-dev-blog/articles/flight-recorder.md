# Flight Recorder — `runtime/trace` (Go 1.25+)

The flight recorder continuously buffers execution trace data in memory, letting you snapshot the last few seconds when something goes wrong — after it already happened.

Reference: [go.dev/blog/flight-recorder](https://go.dev/blog/flight-recorder)

## API

```go
import "runtime/trace"

fr := trace.NewFlightRecorder(trace.FlightRecorderConfig{
    MinAge:   10 * time.Second,  // retain at least this much trace data
    MaxBytes: 10 << 20,          // cap memory usage (10 MiB)
})
fr.Start()

// Later, when something goes wrong:
f, _ := os.Create("snapshot.trace")
fr.WriteTo(f)
f.Close()
fr.Stop()
```

| Method | Description |
|---|---|
| `NewFlightRecorder(cfg)` | Creates a recorder with the given config |
| `Start()` | Begins recording |
| `Stop()` | Stops recording |
| `Enabled() bool` | Check if currently recording |
| `WriteTo(w io.Writer)` | Snapshots buffered trace data |

## Configuration Guidelines

| Parameter | Guideline |
|---|---|
| `MinAge` | Set to ~2x the event window you're debugging (e.g., 10s for a 5s timeout) |
| `MaxBytes` | A few MB/s of trace data for typical services; 10 MB/s for busy ones |

## Usage Pattern

```go
var once sync.Once

func maybeCaptureTrace(fr *trace.FlightRecorder, duration time.Duration) {
    if !fr.Enabled() || duration < threshold {
        return
    }
    once.Do(func() {
        f, err := os.Create("snapshot.trace")
        if err != nil { return }
        defer f.Close()
        fr.WriteTo(f)
        fr.Stop()
    })
}
```

Trigger on: slow requests, failed health checks, timeouts, panics.

## Analyzing Traces

```bash
go tool trace snapshot.trace
```

Opens a browser-based viewer showing:
- **Goroutine timeline** — when each goroutine ran, blocked, or was scheduled
- **Flow events** — how goroutines interact (channel sends, mutex unlocks)
- **Stack traces** — at key points in time
- **STATS** — thread count, heap size, goroutine count

## When to Use

| Scenario | Tool |
|---|---|
| Short-lived program or test | `runtime/trace.Start` / `Stop` |
| Long-running server, known-bad event | **Flight recorder** (snapshot on error) |
| Fleet-wide random sampling | External trace collection infra |
