package jpath

// DefaultRegistry is the package-level default registry
var DefaultRegistry = NewRegistry()

// Parse parses a JSONPath query into a PathExpr syntax tree
func Parse(query string) (PathExpr, error) {
	return DefaultRegistry.Parse(query)
}

// MustParse parses a JSONPath query or panics
func MustParse(query string) PathExpr {
	return DefaultRegistry.MustParse(query)
}

// Compile compiles a parsed PathExpr into a Runnable program
func Compile(path PathExpr) (Runnable, error) {
	return DefaultRegistry.Compile(path)
}

// MustCompile compiles a parsed PathExpr or panics
func MustCompile(path PathExpr) Runnable {
	return DefaultRegistry.MustCompile(path)
}

// Query parses and compiles a JSONPath query into an executable Path
func Query(query string) (*Path, error) {
	return DefaultRegistry.Query(query)
}

// MustQuery parses and compiles a JSONPath query or panics
func MustQuery(query string) *Path {
	return DefaultRegistry.MustQuery(query)
}
