package jpath

var defaultRegistry = NewRegistry()

// Parse parses a JSONPath query into a PathExpr syntax tree
func Parse(query string) (PathExpr, error) {
	return defaultRegistry.Parse(query)
}

// MustParse parses a JSONPath query or panics
func MustParse(query string) PathExpr {
	return defaultRegistry.MustParse(query)
}

// Compile compiles a parsed PathExpr into an executable Path
func Compile(path PathExpr) (Path, error) {
	return defaultRegistry.Compile(path)
}

// MustCompile compiles a parsed PathExpr or panics
func MustCompile(path PathExpr) Path {
	return defaultRegistry.MustCompile(path)
}

// Query parses and compiles a JSONPath query, then runs it on a document
func Query(query string, document any) ([]any, error) {
	return defaultRegistry.Query(query, document)
}

// MustQuery parses and compiles a JSONPath query, then runs it or panics
func MustQuery(query string, document any) []any {
	return defaultRegistry.MustQuery(query, document)
}
