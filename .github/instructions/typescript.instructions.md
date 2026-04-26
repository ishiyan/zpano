---
description: 'TypeScript programming language coding conventions and best practices'
applyTo: '**/*.ts'
---

<!-- No community upstream — locally maintained, based on official TypeScript docs and project conventions. -->
<!-- To update: review against https://www.typescriptlang.org/docs/ and AGENTS.md periodically. -->

# TypeScript Coding Conventions and Best Practices

Follow idiomatic TypeScript practices based on the [TypeScript Handbook](https://www.typescriptlang.org/docs/handbook/) and project-specific conventions.

## General Instructions

- Target ES2020+ with strict mode enabled.
- Prioritize readability, type safety, and correctness.
- Zero runtime dependencies — use only the standard library and language features.
- Ensure code compiles without warnings under strict mode.

## Naming Conventions

- `camelCase` for functions, methods, variables, and properties.
- `PascalCase` for classes, enums, interfaces, and type aliases.
- `UPPER_SNAKE_CASE` for enum members and module-level constants.
- Leading `_` for private fields (`_sampleCount`, `_logretSum`).
- `kebab-case` for filenames (`bar-component.ts`, `quote-component.spec.ts`).
- Avoid stuttering (e.g., `Bar.barOpen` — prefer `Bar.open`).

## Code Style

- 4-space indent. No tabs.
- Keep lines under 100 characters when practical.
- Use guard clauses and early returns to reduce nesting.
- Use template literals for string interpolation.
- Prefer `const` over `let`; never use `var`.

## Types and Null Handling

- Use `number | null` for nullable computation results (impossible computations like division by zero).
- Return `null` explicitly for mathematically impossible results; never return `undefined` for this purpose.
- `throw new Error(...)` for invalid inputs with descriptive messages.
- Use definite assignment assertion (`!:`) for fields set via constructor data-copy pattern.
- Prefer specific types over `any` — use `any` only for legacy constructor patterns.
- Use `Record<K, V>` for dictionary types.

## Imports and Exports

- Relative imports within modules (`import { Bar } from './bar';`).
- Named exports (`export class`, `export enum`, `export function`) — avoid default exports.
- Barrel re-exports via `index.ts` files for module public APIs.
- Spec files may include `import { } from 'jasmine';` for type awareness.

## Classes and Enums

- Use numeric enums starting at 0, with `UPPER_SNAKE_CASE` members.
- Use `readonly` for immutable public fields.
- Use `/** JSDoc */` comments on classes, public methods, and enum members.
- Keep classes focused — one primary responsibility per class.

## Error Handling

- `throw new Error(...)` for invalid inputs — validate early with guard clauses.
- Return `null` for computations that are mathematically impossible (zero denominators, empty arrays).
- Never silently swallow errors.

## Testing

- Jasmine 5 with spec files named `*.spec.ts`, co-located with source files.
- Structure: nested `describe`/`it` blocks with descriptive names.
- `toBeCloseTo(expected, 13)` for floating-point comparisons (13 decimal places).
- `toBeNull()` for null/impossible-computation checks.
- `toBe()` for exact values and booleans.
- `toThrowError()` for input validation tests.
- Helper functions and test data arrays defined within `describe` blocks or at module level.

## Patterns to Avoid

- Don't use `any` where a specific type is possible.
- Don't use `undefined` as a return value for impossible computations — use `null`.
- Don't use default exports.
- Avoid mutable module-level state.
- Avoid deeply nested callbacks — extract to named functions.
- Don't ignore TypeScript compiler errors with `@ts-ignore` unless absolutely necessary (and document why).
