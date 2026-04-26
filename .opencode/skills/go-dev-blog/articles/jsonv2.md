# `encoding/json/v2` — Experimental JSON API (Go 1.25)

Go 1.25 adds experimental `encoding/json/v2` and `encoding/json/jsontext` packages that fix longstanding `encoding/json` problems: stricter syntax, better defaults, streaming support, and options-first design.

Reference: [go.dev/blog/jsonv2-exp](https://go.dev/blog/jsonv2-exp)

**Experimental:** Requires `GOEXPERIMENT=jsonv2`. Not covered by the Go compatibility promise. API may change.

```bash
GOEXPERIMENT=jsonv2 go build ./...
GOEXPERIMENT=jsonv2 go test ./...   # test v1 compatibility under new implementation
```

## Package Layout

| Package | Role |
|---|---|
| `encoding/json/jsontext` | Syntactic: streaming `Encoder`/`Decoder`, `Token`, `Value` types. No reflection. |
| `encoding/json/v2` | Semantic: `Marshal`/`Unmarshal` with options. Builds on `jsontext`. |
| `encoding/json` (v1) | Re-implemented in terms of v2 under the experiment. Behavior unchanged within Go 1 compat. |

## v2 API

All functions accept variadic `Options` for configurable behavior:

```go
import json "encoding/json/v2"

// []byte <-> any
json.Marshal(v, opts...)                   // -> ([]byte, error)
json.Unmarshal(data, v, opts...)           // -> error

// io.Writer / io.Reader <-> any
json.MarshalWrite(w, v, opts...)           // -> error
json.UnmarshalRead(r, v, opts...)          // -> error

// jsontext.Encoder / jsontext.Decoder <-> any (streaming)
json.MarshalEncode(enc, v, opts...)        // -> error
json.UnmarshalDecode(dec, v, opts...)      // -> error
```

## Streaming Interfaces

New interfaces for performance-critical types (avoid allocating `[]byte` per value):

```go
// Preferred — streaming, no intermediate allocation
type MarshalerTo interface {
    MarshalJSONTo(*jsontext.Encoder) error
}
type UnmarshalerFrom interface {
    UnmarshalJSONFrom(*jsontext.Decoder) error
}

// Still supported — same as v1
type Marshaler interface {
    MarshalJSON() ([]byte, error)
}
type Unmarshaler interface {
    UnmarshalJSON([]byte) error
}
```

`MarshalerTo`/`UnmarshalerFrom` eliminate quadratic behavior when methods recursively call `Marshal`/`Unmarshal`, and can yield order-of-magnitude speedups.

## Caller-Specified Customization

Override serialization of any type at the call site using generics:

```go
opts := json.WithMarshalers(
    json.MarshalToFunc(func(enc *jsontext.Encoder, v time.Time) error {
        return enc.WriteValue(jsontext.Value(`"` + v.Format(time.RFC3339) + `"`))
    }),
)
data, err := json.Marshal(myStruct, opts)
```

Symmetric for unmarshaling: `json.WithUnmarshalers(json.UnmarshalFromFunc(...))`.

## Behavior Changes from v1

| Behavior | v1 | v2 |
|---|---|---|
| Invalid UTF-8 | Silently accepted | **Error** |
| Duplicate object keys | Silently accepted | **Error** |
| Nil slice/map | `null` | **`[]` / `{}`** |
| Struct field matching | Case-insensitive | **Case-sensitive** |
| `omitempty` semantics | Zero value of Go type | **Empty JSON value** (`null`, `""`, `[]`, `{}`) |
| `time.Duration` | Marshals as number | **Error** (caller must choose format via options) |
| Pointer receiver `MarshalJSON` | Inconsistently called | **Consistently called** |

Each change has a corresponding option to revert to v1 behavior for gradual migration.

## `jsontext` Basics

Low-level streaming encoder/decoder (no reflection):

```go
import "encoding/json/jsontext"

// Encoding
enc := jsontext.NewEncoder(w, opts...)
enc.WriteToken(jsontext.ObjectStart)
enc.WriteToken(jsontext.String("key"))
enc.WriteValue(jsontext.Value(`"value"`))
enc.WriteToken(jsontext.ObjectEnd)

// Decoding
dec := jsontext.NewDecoder(r, opts...)
for {
    tok, err := dec.ReadToken()
    if err != nil { break }
    // process tok.Kind(), tok.String(), tok.Int64(), etc.
}
```

## Struct Tags

v2 inherits v1 tag syntax and adds new options:

```go
type Example struct {
    Name    string `json:"name"`              // same as v1
    Addr    Addr   `json:"addr,inline"`       // inline struct fields (like XML)
    Created string `json:"created,format:RFC3339"` // format option
    Secret  string `json:"-"`                 // omit (same as v1)
    Empty   string `json:"empty,omitempty"`   // v2: omits "", null, [], {}
}
```

## Migration Strategy

1. **Test compatibility**: Run `GOEXPERIMENT=jsonv2 go test ./...` — v1 is reimplemented in terms of v2, so existing code should pass.
2. **Adopt incrementally**: Use v2 functions with options that preserve v1 semantics where needed.
3. **Upgrade hot paths**: Implement `MarshalerTo`/`UnmarshalerFrom` on types where JSON performance matters.
4. **Report regressions**: [go.dev/issue/71497](https://go.dev/issue/71497).
