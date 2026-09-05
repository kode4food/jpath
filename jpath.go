package jpath

// Parse parses a JSONPath query into a PathExpr syntax tree
func Parse(query string) (*PathExpr, error) {
	var p Parser
	return p.Parse(query)
}

// MustParse parses a JSONPath query or panics
func MustParse(query string) *PathExpr {
	path, err := Parse(query)
	if err != nil {
		panic(err)
	}
	return path
}

// Compile compiles a parsed PathExpr into an executable Path
func Compile(path *PathExpr) (Path, error) {
	return NewRegistry().Compile(path)
}

// MustCompile compiles a parsed PathExpr or panics
func MustCompile(path *PathExpr) Path {
	return NewRegistry().MustCompile(path)
}

// Query parses and compiles a JSONPath query, then runs it on a document
func Query(query string, document any) ([]any, error) {
	return NewRegistry().Query(query, document)
}

// MustQuery parses and compiles a JSONPath query, then runs it or panics
func MustQuery(query string, document any) []any {
	return NewRegistry().MustQuery(query, document)
}
