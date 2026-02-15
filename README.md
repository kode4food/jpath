# jpath

![Build Status](https://github.com/kode4food/jpath/actions/workflows/build.yml/badge.svg) [![Code Coverage](https://qlty.sh/gh/kode4food/projects/jpath/coverage.svg)](https://qlty.sh/gh/kode4food/projects/jpath) [![Maintainability](https://qlty.sh/gh/kode4food/projects/jpath/maintainability.svg)](https://qlty.sh/gh/kode4food/projects/jpath) [![GitHub](https://img.shields.io/github/license/kode4food/jpath)](https://github.com/kode4food/jpath)

jpath is a focused JSONPath parser/compiler for Go. The package exposes six top-level convenience functions backed by a global default registry, and you can create isolated registries for sandboxed extension-function behavior.

## Features

- Parse JSONPath expressions into exported AST nodes
- Compile AST into a fast runnable instruction stream
- Register extension filter functions per registry instance
- Keep registries isolated for sandboxed behavior
- Run JSONPath compliance suite fixtures in tests

## Core API

- `Parse(query string) (PathExpr, error)`
- `MustParse(query string) PathExpr`
- `Compile(path PathExpr) (Runnable, error)`
- `MustCompile(path PathExpr) Runnable`
- `Query(query string) (*Path, error)`
- `MustQuery(query string) *Path`
- `NewRegistry() *Registry`
- `(*Registry).Parse(query string) (PathExpr, error)`
- `(*Registry).Compile(path PathExpr) (Runnable, error)`
- `(*Registry).Query(query string) (*Path, error)`
- `(*Registry).RegisterFunction(name string, def FunctionDefinition) error`
- `(*Registry).Clone() *Registry`

Top-level functions use `DefaultRegistry`. Use explicit `Registry` instances when you need sandboxed extension registration.

## Usage

### Parse, compile, and run

```go
reg := jpath.NewRegistry()
ast, err := reg.Parse("$.store.book[*].title")
if err != nil {
	panic(err)
}
run, err := reg.Compile(ast)
if err != nil {
	panic(err)
}
result := run.Run(doc)
_ = result
```

### One-step query

```go
reg := jpath.NewRegistry()
path := reg.MustQuery("$.store.book[*].title")
result := path.Query(doc)
_ = result
```

### Register an extension function

```go
reg := jpath.NewRegistry()
reg.MustRegisterFunction("startsWith", jpath.FunctionDefinition{
	Validate: func(args []jpath.FilterExpr, use jpath.FunctionUse, inComparison bool) error {
		if len(args) != 2 {
			return fmt.Errorf("invalid function arity")
		}
		if inComparison {
			return fmt.Errorf("function result must not be compared")
		}
		return nil
	},
	Eval: func(args []jpath.FunctionValue) jpath.FunctionValue {
		left, ok := args[0].Scalar.(string)
		if !ok {
			return jpath.FunctionValue{Scalar: false}
		}
		right, ok := args[1].Scalar.(string)
		if !ok {
			return jpath.FunctionValue{Scalar: false}
		}
		return jpath.FunctionValue{Scalar: strings.HasPrefix(left, right)}
	},
})
```

## Errors

`(*Registry).Query` wraps parse and compile failures with `ErrInvalidPath`.

Parser errors are exported for `errors.Is` checks:

- `ErrExpectedRoot`
- `ErrUnexpectedToken`
- `ErrUnterminatedString`
- `ErrBadEscape`
- `ErrBadNumber`
- `ErrBadSlice`
- `ErrBadFunction`

Registry function errors are exported:

- `ErrUnknownFunction`
- `ErrBadFunctionName`
- `ErrBadFunctionDefinition`
- `ErrFunctionExists`

## Status

Work in progress. Not ready for production use.
