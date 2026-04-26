---
description: 'Python programming language coding conventions and best practices'
applyTo: '**/*.py'
---

<!-- No community upstream — locally maintained, based on PEP 8 and PEP 257. -->
<!-- To update: review against https://peps.python.org/pep-0008/ and https://peps.python.org/pep-0257/ periodically. -->

# Python Coding Conventions and Best Practices

Follow idiomatic Python practices based on [PEP 8](https://peps.python.org/pep-0008/), [PEP 257](https://peps.python.org/pep-0257/), and the [Python documentation](https://docs.python.org/3/).

## General Instructions

- Target Python 3.10+ — use modern syntax and type annotations.
- Prioritize readability, simplicity, and correctness.
- Use built-in generics for type hints: `tuple[int, int]`, `list[float]`, `dict[str, int]` (not `typing.Tuple`, etc.).
- Use `float`, `bool`, `int` directly — not `Optional` wrappers unless needed for `None` semantics.
- Ensure code runs without warnings.

## Naming Conventions

- `snake_case` for functions, methods, variables, and module names.
- `PascalCase` for classes.
- `UPPER_SNAKE_CASE` for module-level constants and enum members.
- Leading `_` for private/internal names.
- Avoid single-letter names except for short loop variables or well-known math symbols.

## Code Style

- 4-space indent. No tabs.
- Use backslash `\` for line continuation where parentheses don't suffice.
- Keep lines under 100 characters when practical.
- Use f-strings for string formatting (not `%` or `.format()`).
- Prefer `is None` / `is not None` over `== None`.
- Use guard clauses and early returns to reduce nesting.

## Imports

- Standard library first, then third-party, then local — separated by blank lines.
- Use relative imports within packages (`from .module import name`).
- Tests use absolute imports (`from package.subpackage import module`).
- Avoid wildcard imports (`from module import *`).

## Error Handling

- Raise `ValueError` for invalid inputs with descriptive messages.
- Return `None` for computations that are mathematically impossible (e.g., division by zero in a ratio).
- Use specific exception types — avoid bare `except:`.
- Guard clauses for zero denominators and edge cases before computation.

## Functions and Methods

- Keep functions short and focused on a single task.
- Use type annotations for all public function signatures.
- Document public functions with docstrings (PEP 257).
- Prefer returning values over mutating arguments.
- Use `@staticmethod` and `@classmethod` appropriately.

## Data Structures

- Use `enum.IntEnum` for integer-valued enumerations with readable names.
- Prefer tuples for fixed-size heterogeneous data, lists for homogeneous sequences.
- Use `dataclasses` or plain classes — not `dict` — for structured data with named fields.
- Use `__slots__` on performance-critical classes to reduce memory overhead.

## Testing

- Use `unittest.TestCase` with descriptive method names (`test_<what>_<condition>`).
- `assertAlmostEqual(x, y, places=13)` for floating-point comparisons.
- `assertIsNone()` for null/impossible-computation checks.
- `assertRaises(ValueError)` for input validation tests.
- Tests live in `test_*.py` files alongside or within the package.

## Patterns to Avoid

- Don't use mutable default arguments (`def f(x=[]):`).
- Don't catch exceptions silently — always log or re-raise.
- Avoid global mutable state.
- Don't use `type()` for type checking — use `isinstance()`.
- Avoid deeply nested comprehensions — extract to functions for readability.
