package jpath

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"
)

// Parser parses JSONPath query strings into an inspectable syntax tree
type Parser struct {
	src  []rune
	text string
	pos  int
}

var (
	// ErrInvalidPath is raised when a JSONPath query cannot be parsed
	ErrInvalidPath = errors.New("invalid JSONPath query")

	// ErrExpectedRoot indicates the query does not start with `$`
	ErrExpectedRoot = errors.New("expected root selector '$'")

	// ErrUnexpectedToken indicates invalid token ordering or placement
	ErrUnexpectedToken = errors.New("unexpected token")

	// ErrUnterminatedString indicates a quoted string was not closed
	ErrUnterminatedString = errors.New("unterminated string")

	// ErrBadEscape indicates an invalid escape sequence in a string literal
	ErrBadEscape = errors.New("invalid escape sequence")

	// ErrBadNumber indicates an invalid numeric literal
	ErrBadNumber = errors.New("invalid number")

	// ErrBadSlice indicates an invalid array slice selector
	ErrBadSlice = errors.New("invalid slice selector")

	// ErrBadFunc indicates an invalid function invocation in a filter
	ErrBadFunc = errors.New("invalid function call")
)

const maxJSONInt = int64(9007199254740991)

// Parse parses a JSONPath query into a PathExpr syntax tree
func (p *Parser) Parse(query string) (*PathExpr, error) {
	if query == "" || strings.TrimSpace(query) != query {
		return nil, wrapPathError(query, 0, ErrExpectedRoot)
	}
	p.src = []rune(query)
	p.text = query
	p.pos = 0

	res, err := p.parseFullPath()
	if err != nil {
		return nil, err
	}
	if !p.eof() {
		return nil, wrapPathError(query, p.pos, ErrUnexpectedToken)
	}
	return res, nil
}

func (p *Parser) parseFullPath() (*PathExpr, error) {
	if !p.consume('$') {
		return nil, wrapPathError(p.text, p.pos, ErrExpectedRoot)
	}
	var segments []*SegmentExpr
	for !p.eof() {
		p.skipWS()
		if p.eof() {
			break
		}
		sg, ok, err := p.parseSegment()
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, wrapPathError(p.text, p.pos, ErrUnexpectedToken)
		}
		segments = append(segments, sg)
	}
	return &PathExpr{Segments: segments}, nil
}

func (p *Parser) parseRelativePath() (*PathExpr, error) {
	var segments []*SegmentExpr
	for {
		p.skipWS()
		sg, ok, err := p.parseSegment()
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		segments = append(segments, sg)
	}
	return &PathExpr{Segments: segments}, nil
}

func (p *Parser) parseSegment() (*SegmentExpr, bool, error) {
	if p.eof() {
		return nil, false, nil
	}
	if p.peek() == '.' {
		p.pos++
		if p.consume('.') {
			sels, err := p.parseDescendantSelectors()
			if err != nil {
				return nil, false, err
			}
			return &SegmentExpr{
				Descendant: true,
				Selectors:  sels,
			}, true, nil
		}
		sel, err := p.parseDotSelector()
		if err != nil {
			return nil, false, err
		}
		return &SegmentExpr{
			Selectors: []*SelectorExpr{sel},
		}, true, nil
	}
	if p.peek() == '[' {
		sels, err := p.parseBracketSelectors()
		if err != nil {
			return nil, false, err
		}
		return &SegmentExpr{
			Selectors: sels,
		}, true, nil
	}
	return nil, false, nil
}

func (p *Parser) parseDescendantSelectors() ([]*SelectorExpr, error) {
	if p.eof() {
		return nil, wrapPathError(p.text, p.pos, ErrUnexpectedToken)
	}
	if p.peek() == '*' {
		p.pos++
		return []*SelectorExpr{
			{Kind: SelectorWildcard},
		}, nil
	}
	if p.peek() == '[' {
		return p.parseBracketSelectors()
	}
	nm, ok := p.parseMemberName()
	if !ok {
		return nil, wrapPathError(p.text, p.pos, ErrUnexpectedToken)
	}
	return []*SelectorExpr{
		{Kind: SelectorName, Name: nm},
	}, nil
}

func (p *Parser) parseDotSelector() (*SelectorExpr, error) {
	if p.eof() {
		return nil, wrapPathError(p.text, p.pos, ErrUnexpectedToken)
	}
	if p.peek() == '*' {
		p.pos++
		return &SelectorExpr{Kind: SelectorWildcard}, nil
	}
	nm, ok := p.parseMemberName()
	if !ok {
		return nil, wrapPathError(p.text, p.pos, ErrUnexpectedToken)
	}
	return &SelectorExpr{
		Kind: SelectorName,
		Name: nm,
	}, nil
}

func (p *Parser) parseBracketSelectors() ([]*SelectorExpr, error) {
	if !p.consume('[') {
		return nil, wrapPathError(p.text, p.pos, ErrUnexpectedToken)
	}
	p.skipWS()
	var sels []*SelectorExpr
	for {
		sel, err := p.parseBracketSelector()
		if err != nil {
			return nil, err
		}
		sels = append(sels, sel)
		p.skipWS()
		if p.consume(']') {
			if len(sels) == 0 {
				return nil, wrapPathError(
					p.text,
					p.pos,
					ErrUnexpectedToken,
				)
			}
			return sels, nil
		}
		if !p.consume(',') {
			return nil, wrapPathError(
				p.text,
				p.pos,
				ErrUnexpectedToken,
			)
		}
		p.skipWS()
	}
}

func (p *Parser) parseBracketSelector() (*SelectorExpr, error) {
	if p.eof() {
		return nil, wrapPathError(p.text, p.pos, ErrUnexpectedToken)
	}
	switch p.peek() {
	case '*':
		p.pos++
		return &SelectorExpr{Kind: SelectorWildcard}, nil
	case '?':
		p.pos++
		p.skipWS()
		fl, err := p.parseFilter()
		if err != nil {
			return nil, err
		}
		return &SelectorExpr{Kind: SelectorFilter, Filter: fl}, nil
	case '\'', '"':
		s, err := p.parseString()
		if err != nil {
			return nil, err
		}
		return &SelectorExpr{Kind: SelectorName, Name: s}, nil
	default:
		return p.parseIndexOrSlice()
	}
}

func (p *Parser) parseIndexOrSlice() (*SelectorExpr, error) {
	startPos := p.pos
	p.skipWS()
	hasStart := false
	start := 0
	if p.peek() != ':' {
		n, ok := p.parseIntLiteral()
		if !ok {
			return nil, wrapPathError(
				p.text,
				p.pos,
				ErrUnexpectedToken,
			)
		}
		hasStart = true
		start = n
	}
	p.skipWS()
	if p.peek() != ':' {
		if !hasStart {
			return nil, wrapPathError(p.text, startPos, ErrBadSlice)
		}
		return &SelectorExpr{Kind: SelectorIndex, Index: start}, nil
	}
	p.pos++
	p.skipWS()
	hasEnd := false
	end := 0
	if !p.eof() && p.peek() != ':' && p.peek() != ',' && p.peek() != ']' {
		n, ok := p.parseIntLiteral()
		if !ok {
			return nil, wrapPathError(p.text, p.pos, ErrBadSlice)
		}
		hasEnd = true
		end = n
	}
	p.skipWS()
	step := 1
	if p.consume(':') {
		p.skipWS()
		if !p.eof() && p.peek() != ',' && p.peek() != ']' {
			n, ok := p.parseIntLiteral()
			if !ok {
				return nil, wrapPathError(p.text, p.pos, ErrBadSlice)
			}
			step = n
		}
	}
	return &SelectorExpr{
		Kind: SelectorSlice,
		Slice: &SliceExpr{
			HasStart: hasStart,
			Start:    start,
			HasEnd:   hasEnd,
			End:      end,
			Step:     step,
		},
	}, nil
}

func (p *Parser) parseFilter() (FilterExpr, error) {
	return p.parseExpr()
}

func (p *Parser) parseExpr() (FilterExpr, error) {
	return p.parseOr()
}

func (p *Parser) parseOr() (FilterExpr, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for {
		p.skipWS()
		if !p.consumeString("||") {
			return left, nil
		}
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Op: "||", Left: left, Right: right}
	}
}

func (p *Parser) parseAnd() (FilterExpr, error) {
	left, err := p.parseCompare()
	if err != nil {
		return nil, err
	}
	for {
		p.skipWS()
		if !p.consumeString("&&") {
			return left, nil
		}
		right, err := p.parseCompare()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Op: "&&", Left: left, Right: right}
	}
}

func (p *Parser) parseCompare() (FilterExpr, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for {
		p.skipWS()
		op := ""
		switch {
		case p.consumeString("=="):
			op = "=="
		case p.consumeString("!="):
			op = "!="
		case p.consumeString("<="):
			op = "<="
		case p.consumeString(">="):
			op = ">="
		case p.consume('<'):
			op = "<"
		case p.consume('>'):
			op = ">"
		}
		if op == "" {
			return left, nil
		}
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Op: op, Left: left, Right: right}
	}
}

func (p *Parser) parseUnary() (FilterExpr, error) {
	p.skipWS()
	if p.consume('!') {
		ex, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &UnaryExpr{Op: "!", Expr: ex}, nil
	}
	return p.parsePrimary()
}

func (p *Parser) parsePrimary() (FilterExpr, error) {
	p.skipWS()
	if p.eof() {
		return nil, wrapPathError(p.text, p.pos, ErrUnexpectedToken)
	}
	switch p.peek() {
	case '(':
		p.pos++
		ex, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		p.skipWS()
		if !p.consume(')') {
			return nil, wrapPathError(p.text, p.pos, ErrUnexpectedToken)
		}
		return ex, nil
	case '\'', '"':
		s, err := p.parseString()
		if err != nil {
			return nil, err
		}
		return &LiteralExpr{Value: s}, nil
	case '$':
		p.pos++
		path, err := p.parseRelativePath()
		if err != nil {
			return nil, err
		}
		return &PathValueExpr{Absolute: true, Path: path}, nil
	case '@':
		p.pos++
		path, err := p.parseRelativePath()
		if err != nil {
			return nil, err
		}
		return &PathValueExpr{Absolute: false, Path: path}, nil
	default:
		if isNumberStart(p.peek()) {
			n, ok := p.parseNumberLiteral()
			if !ok {
				return nil, wrapPathError(p.text, p.pos, ErrBadNumber)
			}
			return &LiteralExpr{Value: n}, nil
		}
		if p.consumeString("true") {
			return &LiteralExpr{Value: true}, nil
		}
		if p.consumeString("false") {
			return &LiteralExpr{Value: false}, nil
		}
		if p.consumeString("null") {
			return &LiteralExpr{Value: nil}, nil
		}
		if ident, ok := p.parseIdentifier(); ok {
			if !p.consume('(') {
				return nil, wrapPathError(p.text, p.pos, ErrUnexpectedToken)
			}
			args, err := p.parseCallArgs()
			if err != nil {
				return nil, err
			}
			return &FuncExpr{Name: ident, Args: args}, nil
		}
	}
	return nil, wrapPathError(p.text, p.pos, ErrUnexpectedToken)
}

func (p *Parser) parseCallArgs() ([]FilterExpr, error) {
	p.skipWS()
	if p.consume(')') {
		return nil, nil
	}
	var res []FilterExpr
	for {
		ex, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		res = append(res, ex)
		p.skipWS()
		if p.consume(')') {
			return res, nil
		}
		if !p.consume(',') {
			return nil, wrapPathError(p.text, p.pos, ErrBadFunc)
		}
		p.skipWS()
	}
}

func (p *Parser) parseString() (string, error) {
	if p.eof() {
		return "", wrapPathError(p.text, p.pos, ErrUnterminatedString)
	}
	q := p.peek()
	if q != '\'' && q != '"' {
		return "", wrapPathError(p.text, p.pos, ErrUnterminatedString)
	}
	p.pos++
	var b strings.Builder
	for !p.eof() {
		r := p.peek()
		p.pos++
		if r == q {
			return b.String(), nil
		}
		if r < 0x20 {
			return "", wrapPathError(p.text, p.pos, ErrBadEscape)
		}
		if r != '\\' {
			b.WriteRune(r)
			continue
		}
		if p.eof() {
			return "", wrapPathError(p.text, p.pos, ErrBadEscape)
		}
		esc := p.peek()
		p.pos++
		switch esc {
		case '\\', '/':
			b.WriteRune(esc)
		case '\'':
			if q != '\'' {
				return "", wrapPathError(p.text, p.pos, ErrBadEscape)
			}
			b.WriteRune(esc)
		case '"':
			if q != '"' {
				return "", wrapPathError(p.text, p.pos, ErrBadEscape)
			}
			b.WriteRune(esc)
		case 'b':
			b.WriteByte('\b')
		case 'f':
			b.WriteByte('\f')
		case 'n':
			b.WriteByte('\n')
		case 'r':
			b.WriteByte('\r')
		case 't':
			b.WriteByte('\t')
		case 'u':
			if p.pos+4 > len(p.src) {
				return "", wrapPathError(p.text, p.pos, ErrBadEscape)
			}
			hex := string(p.src[p.pos : p.pos+4])
			if !isHexWord(hex) {
				return "", wrapPathError(p.text, p.pos, ErrBadEscape)
			}
			p.pos += 4
			v, err := strconv.ParseInt(hex, 16, 32)
			if err != nil {
				return "", wrapPathError(p.text, p.pos, ErrBadEscape)
			}
			r := rune(v)
			if r >= 0xD800 && r <= 0xDBFF {
				if p.pos+6 > len(p.src) {
					return "", wrapPathError(p.text, p.pos, ErrBadEscape)
				}
				if p.src[p.pos] != '\\' || p.src[p.pos+1] != 'u' {
					return "", wrapPathError(p.text, p.pos, ErrBadEscape)
				}
				p.pos += 2
				hex = string(p.src[p.pos : p.pos+4])
				if !isHexWord(hex) {
					return "", wrapPathError(p.text, p.pos, ErrBadEscape)
				}
				p.pos += 4
				lo, err := strconv.ParseInt(hex, 16, 32)
				if err != nil {
					return "", wrapPathError(p.text, p.pos, ErrBadEscape)
				}
				rr := rune(lo)
				if rr < 0xDC00 || rr > 0xDFFF {
					return "", wrapPathError(p.text, p.pos, ErrBadEscape)
				}
				b.WriteRune(utf16.DecodeRune(r, rr))
				continue
			}
			if r >= 0xDC00 && r <= 0xDFFF {
				return "", wrapPathError(p.text, p.pos, ErrBadEscape)
			}
			b.WriteRune(r)
		default:
			return "", wrapPathError(p.text, p.pos, ErrBadEscape)
		}
	}
	return "", wrapPathError(p.text, p.pos, ErrUnterminatedString)
}

func (p *Parser) parseNumberLiteral() (float64, bool) {
	start := p.pos
	_ = p.consume('-')
	if !p.parseIntegerPart(start) {
		p.pos = start
		return 0, false
	}
	if p.consume('.') {
		if !p.consumeDigits() {
			p.pos = start
			return 0, false
		}
	}
	if p.consume('e') || p.consume('E') {
		_ = p.consume('+') || p.consume('-')
		if !p.consumeDigits() {
			p.pos = start
			return 0, false
		}
	}
	n, err := strconv.ParseFloat(string(p.src[start:p.pos]), 64)
	if err != nil {
		p.pos = start
		return 0, false
	}
	return n, true
}

func (p *Parser) parseIntLiteral() (int, bool) {
	start := p.pos
	neg := p.consume('-')
	if !p.parseIntegerPart(start) {
		p.pos = start
		return 0, false
	}
	if neg && p.pos == start+2 && p.src[start+1] == '0' {
		p.pos = start
		return 0, false
	}
	raw := string(p.src[start:p.pos])
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		p.pos = start
		return 0, false
	}
	if n < -maxJSONInt || n > maxJSONInt {
		p.pos = start
		return 0, false
	}
	return int(n), true
}

func (p *Parser) parseIntegerPart(start int) bool {
	if p.eof() || !isDigit(p.peek()) {
		p.pos = start
		return false
	}
	if p.peek() == '0' {
		p.pos++
		if !p.eof() && isDigit(p.peek()) {
			p.pos = start
			return false
		}
		return true
	}
	if !p.consumeDigits() {
		p.pos = start
		return false
	}
	return true
}

func (p *Parser) parseMemberName() (string, bool) {
	if p.eof() || !isNameStart(p.peek()) {
		return "", false
	}
	start := p.pos
	p.pos++
	for !p.eof() && isNamePart(p.peek()) {
		p.pos++
	}
	return string(p.src[start:p.pos]), true
}

func (p *Parser) parseIdentifier() (string, bool) {
	return p.parseMemberName()
}

func (p *Parser) consumeDigits() bool {
	if p.eof() || !isDigit(p.peek()) {
		return false
	}
	for !p.eof() && isDigit(p.peek()) {
		p.pos++
	}
	return true
}

func (p *Parser) skipWS() {
	for !p.eof() && unicode.IsSpace(p.peek()) {
		p.pos++
	}
}

func (p *Parser) eof() bool {
	return p.pos >= len(p.src)
}

func (p *Parser) peek() rune {
	if p.eof() {
		return 0
	}
	return p.src[p.pos]
}

func (p *Parser) consume(r rune) bool {
	if p.eof() || p.src[p.pos] != r {
		return false
	}
	p.pos++
	return true
}

func (p *Parser) consumeString(v string) bool {
	r := []rune(v)
	if p.pos+len(r) > len(p.src) {
		return false
	}
	for i := range r {
		if p.src[p.pos+i] != r[i] {
			return false
		}
	}
	p.pos += len(r)
	return true
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isNumberStart(r rune) bool {
	return r == '-' || isDigit(r)
}

func isHexWord(v string) bool {
	if len(v) != 4 {
		return false
	}
	for _, r := range v {
		switch {
		case r >= '0' && r <= '9':
		case r >= 'a' && r <= 'f':
		case r >= 'A' && r <= 'F':
		default:
			return false
		}
	}
	return true
}

func isNameStart(r rune) bool {
	switch {
	case r == '_':
		return true
	case r >= 0x80:
		return true
	default:
		return unicode.IsLetter(r)
	}
}

func isNamePart(r rune) bool {
	if isNameStart(r) {
		return true
	}
	return unicode.IsDigit(r)
}

func wrapPathError(query string, pos int, err error) error {
	return fmt.Errorf(
		"%w at offset %d in %q: %w", ErrInvalidPath, pos, query, err,
	)
}
