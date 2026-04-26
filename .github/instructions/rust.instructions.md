---
description: 'Rust programming language coding conventions and best practices'
applyTo: '**/*.rs'
---

<!-- Upstream: https://github.com/github/awesome-copilot/blob/main/instructions/rust.instructions.md -->
<!-- Locally trimmed to remove external-crate assumptions (serde, tokio, rayon, async-std) for this zero-deps project. -->
<!-- To update: fetch upstream, diff, merge relevant changes, then re-apply project-specific trimming. -->

# Rust Coding Conventions and Best Practices

Follow idiomatic Rust practices and community standards when writing Rust code.

Based on [The Rust Book](https://doc.rust-lang.org/book/), [Rust API Guidelines](https://rust-lang.github.io/api-guidelines/), and [RFC 430 naming conventions](https://github.com/rust-lang/rfcs/blob/master/text/0430-finalizing-naming-conventions.md).

## General Instructions

- Prioritize readability, safety, and maintainability.
- Use strong typing and leverage Rust's ownership system for memory safety.
- Handle errors gracefully using `Result<T, E>` and provide meaningful error messages.
- Use consistent naming conventions following [RFC 430](https://github.com/rust-lang/rfcs/blob/master/text/0430-finalizing-naming-conventions.md).
- Write idiomatic, safe, and efficient Rust code that follows the borrow checker's rules.
- Ensure code compiles without warnings.

## Patterns to Follow

- Use modules (`mod`) and public interfaces (`pub`) to encapsulate logic.
- Handle errors properly using `?`, `match`, or `if let`.
- Prefer enums over flags and states for type safety.
- Use iterators instead of index-based loops as they're often faster and safer.
- Use `&str` instead of `String` for function parameters when you don't need ownership.
- Prefer borrowing and zero-copy operations to avoid unnecessary allocations.

### Ownership, Borrowing, and Lifetimes

- Prefer borrowing (`&T`) over cloning unless ownership transfer is necessary.
- Use `&mut T` when you need to modify borrowed data.
- Explicitly annotate lifetimes when the compiler cannot infer them.
- Use `Rc<T>` / `Arc<T>` for shared ownership (single-threaded / thread-safe).
- Use `RefCell<T>` / `Mutex<T>` / `RwLock<T>` for interior mutability.

## Patterns to Avoid

- Don't use `unwrap()` or `expect()` unless absolutely necessary—prefer proper error handling.
- Avoid panics in library code—return `Result` instead.
- Don't rely on global mutable state—use dependency injection or thread-safe containers.
- Avoid deeply nested logic—refactor with functions or combinators.
- Avoid `unsafe` unless required and fully documented.
- Don't overuse `clone()`; prefer borrowing.
- Avoid premature `collect()`; keep iterators lazy until you need the collection.

## Code Style and Formatting

- Use `rustfmt` for formatting. Use `cargo clippy` for linting.
- Keep lines under 100 characters when possible.
- Place documentation immediately before the item using `///`.

## Error Handling

- Use `Result<T, E>` for recoverable errors and `panic!` only for unrecoverable errors.
- Prefer `?` operator for error propagation.
- Implement `std::error::Error` for custom error types (or use `thiserror` if available).
- Use `Option<T>` for values that may or may not exist.
- Validate function arguments and return appropriate errors for invalid input.

## API Design

- Eagerly implement common traits: `Debug`, `Clone`, `PartialEq`, `Default`, and others as appropriate.
- Use standard conversion traits: `From`, `AsRef`, `AsMut`.
- Use newtypes for static distinctions; prefer specific types over `bool` parameters.
- Functions with a clear receiver should be methods.
- Structs should have private fields; all public types must implement `Debug`.

## Testing

- Write unit tests in `#[cfg(test)] mod tests { ... }` at the bottom of source files.
- Use `#[test]` annotations; name tests descriptively.
- Test both success and error cases.
- Document all public APIs with rustdoc (`///` comments).

## Quality Checklist

- [ ] Naming follows RFC 430 conventions
- [ ] Common traits implemented (`Debug`, `Clone`, `PartialEq`) where appropriate
- [ ] Error handling uses `Result<T, E>` with meaningful error types
- [ ] No unnecessary `unsafe` code
- [ ] Efficient use of iterators, minimal allocations
- [ ] Code passes `cargo fmt`, `cargo clippy`, and `cargo test`

## Related Skills

- **`rust-skills`** — Comprehensive Rust coding guidelines with 179 rules across 14 categories covering ownership, error handling, async patterns, API design, memory optimization, performance, testing, and common anti-patterns. Load for in-depth guidance when writing or reviewing Rust code.
