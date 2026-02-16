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
		Validate FunctionValidator
		Eval     FunctionEvaluator
	}

	// FunctionValidator validates function arguments for a call site
	FunctionValidator func(
		args []FilterExpr, use FunctionUse, inComparison bool,
	) error

	// FunctionEvaluator evaluates arguments and returns a function result
	FunctionEvaluator func(args []*FunctionValue) *FunctionValue

	// FunctionValue stores either a scalar or node list argument/result
	FunctionValue struct {
		IsNodes bool
		Scalar  any
		Nodes   []any
	}

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

// RegisterFunction registers a named function in this registry
func (r *Registry) RegisterFunction(name string, def *FunctionDefinition) error {
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

// MustRegisterFunction registers a function or panics
func (r *Registry) MustRegisterFunction(
	name string, def *FunctionDefinition,
) *Registry {
	if err := r.RegisterFunction(name, def); err != nil {
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
	return run.Query(document), nil
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
