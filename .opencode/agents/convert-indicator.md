---
description: Converts/ports an indicator from its reference implementation to a target language following strict conventions
mode: subagent
permission:
  edit: allow
  bash:
    "*": deny
    "cd go && go test*": allow
    "cd ts && npm*": allow
    "cd rs && cargo test*": allow
    "cd zig && zig build*": allow
    "PYTHONPATH=. python3 -m unittest*": allow
    "git diff*": allow
    "git status*": allow
---
You are converting/porting an indicator to a target language. You must follow the project's strict conventions exactly.

## Input

$ARGUMENTS should be: `<indicator-name> <target-language>`

Example: `bollinger_bands rust` or `keltner_channel zig`

## Steps

1. Load BOTH the `indicator-checklist` and `indicator-conversion` skills.
2. Identify the reference implementation:
   - For Go/TS indicators: the other of the pair is the reference (Go references TS, TS references Go)
   - For Python/Zig/Rust: use BOTH Go and TS as references
3. Read the reference implementation completely (all files: main, params, output enum, test).
4. Read any existing implementation in the target language (if partially done).
5. Implement the indicator in the target language following:
   - The `indicator-checklist` skill for exact patterns (imports, naming, structure)
   - The `indicator-conversion` skill for the full conversion workflow
6. Create/update all required files: implementation, params (if separate), test, descriptor entry, factory entry.
7. Run the tests for the new implementation.
8. Self-verify against the checklist (all 3 categories).
9. If any violations found, fix them and re-run tests.

## Critical Rules

- Match the reference output to 13+ decimal places
- Use EXACTLY the import patterns specified in the checklist for the target language
- Never deviate from the naming conventions table
- Always register in both descriptor and factory
- Component sentinels must use the language-idiomatic null check pattern
- Output enum values start at 1 (all languages)

## Output Format

When done, report:
```
## Conversion: <IndicatorName> → <language>

### Files created/modified
- list of files

### Tests
- PASS / FAIL

### Self-verification
- Import Patterns: PASS
- Naming Conventions: PASS  
- Structural Patterns: PASS

### Notes
- Any deviations or decisions made
```
