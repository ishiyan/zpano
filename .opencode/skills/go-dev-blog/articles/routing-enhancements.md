# HTTP Routing Enhancements (Go 1.22+)

Go 1.22 adds method matching and wildcards to `net/http.ServeMux`, reducing the need for third-party routers.

Reference: [go.dev/blog/routing-enhancements](https://go.dev/blog/routing-enhancements)

## New Pattern Syntax

```go
http.HandleFunc("GET /posts/{id}", handlePost)
http.HandleFunc("POST /posts/", createPost)
http.HandleFunc("/files/{pathname...}", serveFile)
http.HandleFunc("/posts/{$}", listPosts)  // exact match, trailing slash only
```

| Syntax | Meaning |
|---|---|
| `GET /path` | Method prefix — matches GET (and HEAD). Other methods match exactly. |
| `{name}` | Wildcard — matches one path segment |
| `{name...}` | Remaining wildcard — matches all remaining segments |
| `{$}` | End anchor — matches only the exact path with trailing slash |

## Extracting Wildcards

```go
func handlePost(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    // ...
}
```

`r.SetPathValue(name, value)` — lets external routers expose their matches via the same API.

## Precedence Rules

**Most specific pattern wins.** A pattern is more specific if it matches a strict subset of requests.

| Rule | Example |
|---|---|
| Literal beats wildcard | `/posts/latest` > `/posts/{id}` |
| Method beats no-method | `GET /posts/{id}` > `/posts/{id}` |
| Longer literal prefix wins (old behavior preserved) | `/posts/` > `/` |
| Host beats no-host (tie-breaker) | `example.com/` > `/` |

**Conflicts:** if two patterns overlap and neither is more specific, registering both panics.

```go
// These conflict — panics at registration:
// "/posts/{id}"    and   "/{resource}/latest"
```

## 405 Method Not Allowed

If a path matches but the method does not, the server automatically replies `405 Method Not Allowed` with an `Allow` header listing valid methods.

## Compatibility

- Patterns with braces `{}` were previously treated as literals. `GODEBUG=httpmuxgo121` restores old behavior.
- Requires `go 1.22` in `go.mod` to use the new syntax.
