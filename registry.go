package jpath

import (
	"errors"
	"fmt"
	"maps"
)

type (
	// Registry stores function definitions and owns parse/compile/query methods
	Registry struct {
		functions map[string]*FunctionDefinition
	}

	// FunctionDefinition describes a filter function implementation
	FunctionDefinition struct {
		Validate Validator
		Eval     Evaluator
	}

	// Validator validates function arguments for a call site
	Validator func(args []FilterExpr, use FunctionUse, inComparison bool) error

	// Evaluator evaluates wrapped arguments and returns a wrapped result
	Evaluator func(args []*Value) *Value

	// Function evaluates scalar arguments and returns a scalar result
	Function func(args ...any) (any, bool)

	// FunctionUse describes where a function appears in filter validation
	FunctionUse uint8
)

const (
	FunctionUseLogical FunctionUse = iota
	FunctionUseComparisonOperand
	FunctionUseArgument
)

var (
	// ErrUnknownFunc indicates a function name is not registered
	ErrUnknownFunc = errors.New("unknown function")

	// ErrBadFuncName indicates a function name is invalid
	ErrBadFuncName = errors.New("invalid function name")

	// ErrBadFuncDefinition indicates a function definition is invalid
	ErrBadFuncDefinition = errors.New("invalid function definition")

	// ErrFuncExists indicates a function was already registered
	ErrFuncExists = errors.New("function already registered")
)

// NewRegistry creates a registry with default JSONPath functions
func NewRegistry() *Registry {
	res := &Registry{
		functions: map[string]*FunctionDefinition{},
	}
	registerDefaultFunctions(res)
	return res
}

// Clone makes an isolated copy of this registry
func (r *Registry) Clone() *Registry {
	res := *r
	res.functions = maps.Clone(r.functions)
	return &res
}

// RegisterDefinition registers a named function definition in this registry
func (r *Registry) RegisterDefinition(
	name string, def *FunctionDefinition,
) error {
	if !isValidFunctionName(name) {
		return fmt.Errorf("%w: %s", ErrBadFuncName, name)
	}
	if def.Eval == nil {
		return fmt.Errorf("%w: %s", ErrBadFuncDefinition, name)
	}
	if r.functions == nil {
		r.functions = map[string]*FunctionDefinition{}
		registerDefaultFunctions(r)
	}
	if _, ok := r.functions[name]; ok {
		return fmt.Errorf("%w: %s", ErrFuncExists, name)
	}
	r.functions[name] = def
	return nil
}

// RegisterFunction registers a singular-arg scalar function
func (r *Registry) RegisterFunction(name string, arity int, fn Function) error {
	if arity < 0 {
		return fmt.Errorf("%w: %s", ErrBadFuncDefinition, name)
	}
	if fn == nil {
		return fmt.Errorf("%w: %s", ErrBadFuncDefinition, name)
	}
	return r.RegisterDefinition(name, &FunctionDefinition{
		Validate: func(
			args []FilterExpr, _ FunctionUse, _ bool,
		) error {
			if len(args) == arity {
				return nil
			}
			return fmt.Errorf("%w: %s", ErrInvalidFuncArity, name)
		},
		Eval: WrapFunction(fn),
	})
}

// MustRegisterFunction registers a scalar function or panics
func (r *Registry) MustRegisterFunction(
	name string, arity int, fn Function,
) *Registry {
	if err := r.RegisterFunction(name, arity, fn); err != nil {
		panic(err)
	}
	return r
}

// MustRegisterDefinition registers a function definition or panics
func (r *Registry) MustRegisterDefinition(
	name string, def *FunctionDefinition,
) *Registry {
	if err := r.RegisterDefinition(name, def); err != nil {
		panic(err)
	}
	return r
}

// Parse parses a query string into a syntax tree
func (r *Registry) Parse(query string) (*PathExpr, error) {
	var p Parser
	return p.Parse(query)
}

// MustParse parses a query string or panics
func (r *Registry) MustParse(query string) *PathExpr {
	res, err := r.Parse(query)
	if err != nil {
		panic(err)
	}
	return res
}

// Compile compiles a parsed syntax tree into an executable Path
func (r *Registry) Compile(path *PathExpr) (Path, error) {
	c := &Compiler{registry: r}
	return c.Compile(path)
}

// MustCompile compiles a parsed syntax tree or panics
func (r *Registry) MustCompile(path *PathExpr) Path {
	res, err := r.Compile(path)
	if err != nil {
		panic(err)
	}
	return res
}

// Query parses and compiles a query string, then runs it on a document
func (r *Registry) Query(query string, document any) ([]any, error) {
	ast, err := r.Parse(query)
	if err != nil {
		return nil, err
	}
	run, err := r.Compile(ast)
	if err != nil {
		return nil, wrapPathError(query, 0, err)
	}
	return run(document), nil
}

// MustQuery parses and compiles a query string, then runs it or panics
func (r *Registry) MustQuery(query string, document any) []any {
	res, err := r.Query(query, document)
	if err != nil {
		panic(err)
	}
	return res
}

func (r *Registry) function(name string) (*FunctionDefinition, bool) {
	def, ok := r.functions[name]
	return def, ok
}

// WrapFunction wraps a scalar function as an evaluator
func WrapFunction(fn Function) Evaluator {
	return func(args []*Value) *Value {
		values := make([]any, len(args))
		for idx, arg := range args {
			val, ok := arg.singularValue()
			if !ok {
				return ScalarValue(nothing)
			}
			values[idx] = val
		}
		res, ok := fn(values...)
		if !ok {
			return ScalarValue(nothing)
		}
		return ScalarValue(res)
	}
}

func isValidFunctionName(name string) bool {
	if name == "" {
		return false
	}
	for idx, r := range name {
		if idx == 0 {
			if !isNameStart(r) {
				return false
			}
			continue
		}
		if !isNamePart(r) {
			return false
		}
	}
	return true
}
