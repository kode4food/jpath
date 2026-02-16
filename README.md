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

| Signature | Description |
| --- | --- |
| `Parse(query string) (*PathExpr, error)` | Parse a query string into an AST (`PathExpr`) |
| `MustParse(query string) *PathExpr` | Parse a query string into an AST and panic on error |
| `Compile(path *PathExpr) (Path, error)` | Compile an AST into an executable `Path` function |
| `MustCompile(path *PathExpr) Path` | Compile an AST into an executable `Path` function and panic on error |
| `Query(query string, document any) ([]any, error)` | Parse, compile, and execute a query against a document with the default registry |
| `MustQuery(query string, document any) []any` | Parse, compile, and execute a query against a document with the default registry, panicking on error |
| `Path func(document any) []any` | Compiled query function returned by `Compile` |

## Registry Management

| Signature | Description |
| --- | --- |
| `NewRegistry() *Registry` | Create an isolated registry preloaded with default JSONPath functions |
| `.Parse(query string) (*PathExpr, error)` | Parse using this registry context |
| `.Compile(path *PathExpr) (Path, error)` | Compile using this registry's function definitions |
| `.Query(query string, document any) ([]any, error)` | Parse, compile, and execute using this registry |
| `.RegisterFunction(name string, arity int, fn Function) error` | Register a scalar extension function with fixed arity |
| `.RegisterDefinition(name string, def *FunctionDefinition) error` | Register a full custom function definition (validation + evaluation) |
| `.Clone() *Registry` | Copy the registry so function registration can diverge safely |

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
matches := path(document)
```

### One-step query

```go
matches := jpath.MustQuery("$.store.book[*].title", document)
```

### Register an extension function

```go
registry := jpath.NewRegistry()
registry.MustRegisterFunction(
	"startsWith", 2, func(args ...any) (any, bool) {
		left, ok := args[0].(string)
		if !ok {
			return nil, false
		}
		right, ok := args[1].(string)
		if !ok {
			return nil, false
		}
		return strings.HasPrefix(left, right), true
	},
)
```

Use `RegisterDefinition` when you need full control over validation rules, node-list arguments, or custom result shapes

## Status

- Implements RFC 9535 (JSONPath)
- Passes the JSONPath Compliance Test Suite
- Pre-v1.0.0: API may change
