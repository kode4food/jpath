# jpath

![Build Status](https://github.com/kode4food/jpath/actions/workflows/build.yml/badge.svg) [![Code Coverage](https://qlty.sh/gh/kode4food/projects/jpath/coverage.svg)](https://qlty.sh/gh/kode4food/projects/jpath) [![Maintainability](https://qlty.sh/gh/kode4food/projects/jpath/maintainability.svg)](https://qlty.sh/gh/kode4food/projects/jpath) [![GitHub](https://img.shields.io/github/license/kode4food/jpath)](https://github.com/kode4food/jpath)

jpath is a JSONPath parser/compiler for Go. It is built around a two-stage pipeline: parse into an inspectable AST, then compile into a composed function chain. That separation lets you validate or rewrite paths before execution, while keeping runtime evaluation fast and predictable. Extension functions are scoped through registries so custom behavior can be sandboxed.

## Features

- Parse JSONPath expressions into AST nodes
- Compile AST into a fast runnable function chain
- Register extension filter functions per registry instance
- Keep registries isolated for sandboxed behavior
- Run JSONPath compliance suite fixtures in tests

## Core API

- `Parse(query string) (*PathExpr, error)`
- `MustParse(query string) *PathExpr`
- `Compile(path *PathExpr) (Path, error)`
- `MustCompile(path *PathExpr) Path`
- `Query(query string, document any) ([]any, error)`
- `MustQuery(query string, document any) []any`

### Registry Management

- `NewRegistry() *Registry`
- `(*Registry).Parse(query string) (*PathExpr, error)`
- `(*Registry).Compile(path *PathExpr) (Path, error)`
- `(*Registry).Query(query string, document any) ([]any, error)`
- `(*Registry).RegisterFunction(name string, def *FunctionDefinition) error`
- `(*Registry).Clone() *Registry`

Top-level functions use a default registry. Use explicit `Registry` instances when you need sandboxed extension registration.

## Usage

### Parse, compile, and run

```go
registry := jpath.NewRegistry()
pathExpr, err := registry.Parse("$.store.book[*].title")
if err != nil {
	panic(err)
}
path, err := registry.Compile(pathExpr)
if err != nil {
	panic(err)
}
matches := path.Query(document)
```

### One-step query

```go
matches := jpath.MustQuery("$.store.book[*].title", document)
```

### Compose a Path manually

```go
path := jpath.ComposePath(
	jpath.ChildSegment(jpath.SelectName("store")),
	jpath.ChildSegment(jpath.SelectName("book")),
	jpath.ChildSegment(jpath.SelectWildcard()),
	jpath.ChildSegment(jpath.SelectName("title")),
)
matches := path.Query(document)
```

### Register an extension function

```go
registry := jpath.NewRegistry()
registry.MustRegisterFunction("startsWith", &jpath.FunctionDefinition{
	Validate: func(args []jpath.FilterExpr, use jpath.FunctionUse, inComparison bool) error {
		if len(args) != 2 {
			return fmt.Errorf("invalid function arity")
		}
		if inComparison {
			return fmt.Errorf("function result must not be compared")
		}
		return nil
	},
	Eval: func(args []*jpath.FunctionValue) *jpath.FunctionValue {
		left, ok := args[0].Scalar.(string)
		if !ok {
			return &jpath.FunctionValue{Scalar: false}
		}
		right, ok := args[1].Scalar.(string)
		if !ok {
			return &jpath.FunctionValue{Scalar: false}
		}
		return &jpath.FunctionValue{Scalar: strings.HasPrefix(left, right)}
	},
})
```

## Errors

`(*Registry).Query` wraps parse and compile failures with `ErrInvalidPath`.

Parser errors for `errors.Is` checks:

- `ErrExpectedRoot`
- `ErrUnexpectedToken`
- `ErrUnterminatedString`
- `ErrBadEscape`
- `ErrBadNumber`
- `ErrBadSlice`
- `ErrBadFunc`

Registry function errors:

- `ErrUnknownFunc`
- `ErrBadFuncName`
- `ErrBadFuncDefinition`
- `ErrFuncExists`

Function validation errors:

- `ErrInvalidFuncArity`
- `ErrFuncResultMustBeCompared`
- `ErrFuncResultMustNotBeCompared`
- `ErrFuncRequiresSingularQuery`
- `ErrFuncRequiresQueryArgument`

## Status

- Implements RFC 9535 (JSONPath)
- Passes the JSONPath Compliance Test Suite
- Pre-v1.0.0: API may change
