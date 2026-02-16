# AGENTS.md

This file defines repository-specific conventions for agents changing code and markdown in `jpath`.

## Baseline

- Target Go version: `1.25`.
- Keep changes idiomatic and minimal.
- Follow existing local patterns over generic style rules.

## File And Declaration Layout

Use this top-level order unless a file has a clear local reason to differ:

1. `package`
2. `import` block
3. `type` declarations
4. `const` declarations
5. `var` declarations
6. functions and methods

For `type` declarations, order by containership and usage priority:

- Put the primary API type first in the block.
- Put types contained by that type next.
- Put dependencies of those types after their users.
- Put unexported types after exported types, still following containership.

For grouped declarations:

- Use `type ( ... )`, `const ( ... )`, or `var ( ... )` only when there are multiple related declarations.
- Put exported types first, then unexported types, preserving containership and dependency order.
- Error sentinels must use `Err`-prefixed names in a `var` block.
- Keep compile-time interface assertions in `var` blocks.

## Function And Method Ordering

Within a file, order callables as follows:

1. Constructors (`New...`)
2. Exported methods of exported types
3. Unexported methods of exported types
4. Methods of unexported types
5. Exported functions
6. Unexported helper functions

For exported receiver types:

- Put exported methods before unexported methods.
- Keep related methods together by functionality.
- Within a group, order by call chain or first use.

## Naming

- Receiver names: short, lowercase, and derived from type name (`e *Engine`, `c *Compiler`).
- Prefer short local names in tight scope (`i`, `n`, `ok`, `err`, `ctx`).
- Use `ok` for map/type assertion booleans.
- Constructor names must use `New`.
- Acronyms stay uppercase (`ID`, `URL`, `HTTP`).

## Imports

- Run `goimports` on all files.
- Keep imports grouped and sorted per `goimports` output.

## Wrapping And Formatting

- Wrap Go source to 80 columns max (tabs count as 4 columns).
- Keep signatures and calls on one line when they fit.
- When wrapping is required, pack as many args per line as fit before wrapping.
- For wrapped signatures and calls, break after `(` and keep trailing commas.
- For wrapped calls where first arg is `t *testing.T`, keep `t` on the first line before wrapping remaining args.
- In `switch` statements, leave one blank line after each non-empty `case` block when another `case` or `default` block follows.

## Markdown Style

Imported from `../argyll/.claude/rules/markdown-style.md`:

- Do not hard-wrap Markdown text.
- Let lines run long in `.md` files.
- Exception: do not reformat `LICENSE.md`.

## Implementation Style

- Prefer early returns and guard clauses.
- Avoid deep nesting; one conditional nesting level max unless duplication avoidance clearly justifies more.
- Keep helpers near where they are used.
- Do not use panic in production code paths.

## Error Handling

- Publicly relevant errors must be typed sentinels (`Err...`) and returned or wrapped.
- Wrap with `%w` first, then context (`fmt.Errorf("%w: %s", ErrX, detail)`).
- Do not return plain ad-hoc error strings from production paths when a typed sentinel applies.

## Testing Conventions

- Tests are black-box only: use external test packages (`<pkg>_test`).
- Use `testing` package with `TestXxx` naming.
- Keep tests deterministic and explicit.
- Verify both happy paths and error paths.

## Change Checklist For Agents

Before finishing a change:

1. Ensure formatting is clean (`gofmt` and `goimports` as needed).
2. Ensure declaration and method order match this file.
3. Ensure typed errors and wrapping style are correct.
4. Ensure markdown style is respected (no hard-wrapped markdown text).
5. Ensure tests are black-box by default.
6. Run relevant tests (at minimum affected packages; ideally `go test ./...`).
