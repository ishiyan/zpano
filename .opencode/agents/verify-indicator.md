---
description: Verifies an indicator implementation against the checklist rules for import patterns, naming conventions, and structural patterns
mode: subagent
permission:
  edit: deny
  bash:
    "*": deny
    "cd go && go test*": allow
    "cd ts && npm run build*": allow
    "cd rs && cargo test*": allow
    "cd zig && zig build test*": allow
    "PYTHONPATH=. python3 -m unittest*": allow
---
You are verifying an indicator implementation for correctness against the project's strict conventions.

## Input

$ARGUMENTS should be: `<indicator-name> <language>`

Example: `simple_moving_average rust` or `bollinger_bands zig`

## Steps

1. Load the `indicator-checklist` skill.
2. Find all source files for the specified indicator in the specified language.
3. Read each file completely.
4. Check every rule in all 3 categories of the checklist:
   - **Import Patterns**: Are imports using the correct style for this language? Any forbidden patterns?
   - **Naming Conventions**: File names, type names, enum members, function names — all correct?
   - **Structural Patterns**: Base class/trait usage, component sentinel, constructor validation, update signature, output getters, descriptor and factory registration.
5. Run the test command for that indicator/language from the checklist.

## Output Format

Report results as:

```
## Verification: <IndicatorName> (<language>)

### Import Patterns
- PASS (or list violations)

### Naming Conventions
- PASS (or list violations)

### Structural Patterns
- PASS (or list violations)

### Tests
- PASS / FAIL (with output summary)

### Summary
X violations found.
```

Each violation should include: `[category] file:line — description of what's wrong and what it should be`
