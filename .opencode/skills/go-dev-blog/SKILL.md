---
name: go-dev-blog
description: Modern Go features and patterns extracted from go.dev/blog articles (Go 1.21–1.26+). Load when writing Go code that uses newer language features, benchmarks, concurrency testing, JSON v2, GC primitives, generics, structured logging, routing, PGO, iterators, or compiler/runtime optimizations.
---

# Go Dev Blog — Modern Features Reference

Curated guidance from official Go blog articles covering new language features, APIs, and compiler improvements from Go 1.21 through 1.26+.

Source: [go.dev/blog](https://go.dev/blog/)

## Articles

| Article | Go Version | Blog Link |
|---|---|---|
| Benchmark Loop (`testing.B.Loop`) | 1.24+ | [go.dev/blog/testing-b-loop](https://go.dev/blog/testing-b-loop) |
| Cleanups and Weak Pointers (`runtime.AddCleanup`, `weak.Pointer`) | 1.24+ | [go.dev/blog/cleanups-and-weak](https://go.dev/blog/cleanups-and-weak) |
| Go Fix (`go fix` modernizer) | 1.26+ | [go.dev/blog/gofix](https://go.dev/blog/gofix) |
| Generic Interfaces (self-referential constraints, pointer receivers) | 1.18+ | [go.dev/blog/generic-interfaces](https://go.dev/blog/generic-interfaces) |
| Source-Level Inliner (`//go:fix inline`) | 1.26+ | [go.dev/blog/inliner](https://go.dev/blog/inliner) |
| JSON v2 (`encoding/json/v2`, experimental) | 1.25 | [go.dev/blog/jsonv2-exp](https://go.dev/blog/jsonv2-exp) |
| OS Root (`os.Root`, traversal-resistant file access) | 1.24+ | [go.dev/blog/osroot](https://go.dev/blog/osroot) |
| Stack Allocation of Slices | 1.25+ | [go.dev/blog/allocation-optimizations](https://go.dev/blog/allocation-optimizations) |
| Synctest (`testing/synctest`, deterministic concurrency testing) | 1.25+ | [go.dev/blog/testing-time](https://go.dev/blog/testing-time) |
| Container-Aware GOMAXPROCS | 1.25+ | [go.dev/blog/container-aware-gomaxprocs](https://go.dev/blog/container-aware-gomaxprocs) |
| Flight Recorder (`runtime/trace` ring buffer) | 1.25+ | [go.dev/blog/flight-recorder](https://go.dev/blog/flight-recorder) |
| Generic Slice Functions (`slices` package) | 1.22+ | [go.dev/blog/generic-slice-functions](https://go.dev/blog/generic-slice-functions) |
| Green Tea GC (green-thread-aware garbage collector) | 1.25/1.26 | [go.dev/blog/greenteagc](https://go.dev/blog/greenteagc) |
| Profile-Guided Optimization (PGO) | 1.21+ | [go.dev/blog/pgo](https://go.dev/blog/pgo) |
| Range Functions (iterators) | 1.23+ | [go.dev/blog/range-functions](https://go.dev/blog/range-functions) |
| Routing Enhancements (`net/http.ServeMux` patterns) | 1.22+ | [go.dev/blog/routing-enhancements](https://go.dev/blog/routing-enhancements) |
| Structured Logging (`log/slog`) | 1.21+ | [go.dev/blog/slog](https://go.dev/blog/slog) |
| Swiss Tables (map implementation) | 1.24+ | [go.dev/blog/swisstable](https://go.dev/blog/swisstable) |
| Type Construction and Cycle Detection | 1.26 | [go.dev/blog/type-construction-and-cycle-detection](https://go.dev/blog/type-construction-and-cycle-detection) |
| Unique (`unique.Handle` for interning) | 1.23+ | [go.dev/blog/unique](https://go.dev/blog/unique) |

## Full Content

!articles/benchmark-loop.md

!articles/cleanups-and-weak.md

!articles/fix.md

!articles/generic-interfaces.md

!articles/inliner.md

!articles/jsonv2.md

!articles/osroot.md

!articles/stack-allocation.md

!articles/synctest.md

!articles/container-gomaxprocs.md

!articles/flight-recorder.md

!articles/generic-slices.md

!articles/green-tea-gc.md

!articles/pgo.md

!articles/range-functions.md

!articles/routing-enhancements.md

!articles/slog.md

!articles/swiss-tables.md

!articles/type-construction.md

!articles/unique.md

---

## Maintaining This Skill

How to discover, assess, and add new blog articles.

### 1. Scan for new articles

Fetch the blog index at [go.dev/blog](https://go.dev/blog/) and compare against the Articles table above. Look for posts about new language features, standard library additions, runtime/compiler changes, or significant API patterns. Ignore release announcements, community/event posts, and opinion pieces.

### 2. Assess relevance

An article is worth adding if it meets **any** of these criteria:
- Introduces a **new API or language feature** an agent would need to use correctly (e.g., `unique.Handle`, range-over-func iterators, `os.Root`)
- Documents **changed behavior** that would cause bugs if the agent uses outdated patterns (e.g., Swiss Tables changing map iteration, new GOMAXPROCS defaults in containers)
- Covers **compiler/runtime internals** that affect how an agent should write code for performance (e.g., stack allocation heuristics, PGO, inlining pragmas)

Skip articles that are purely explanatory history, community updates, or duplicates of content already covered by `go.instructions.md` or the `modern-go-guidelines` skill.

### 3. Distill the article

Fetch the full blog post and condense it into **actionable coding guidance** — not a copy of the blog. The article file should contain:
- A `#` heading with the feature name and a reference link to the original blog post
- **When to use** / **key rules** — the practical guidance an agent needs
- **Code examples** — minimal, showing correct usage and common pitfalls
- **Tables** for enumerating options, flags, or before/after comparisons
- Target length: 50–150 lines. Cut background motivation and historical context.

### 4. Add to the skill

1. Write the distilled article to `articles/<slug>.md`
2. Add a row to the Articles table in this file (keep alphabetical order by article name)
3. Add a `!articles/<slug>.md` include at the bottom of the Full Content section
4. Update the `description` field in the frontmatter if the new article covers a topic area not yet mentioned

### Gotcha: blog URL slugs

Blog URL slugs don't always match the article title or the filename you'd expect. Always verify the actual URL by fetching it before adding the link. Known examples: `greenteagc` (not `green-tea-gc`), `swisstable` (not `swiss-tables`), `generic-slice-functions` (not `generic-slices`).
