package jpath_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kode4food/jpath"
)

type benchmarkCase struct {
	document any
	path     jpath.Path
}

var benchmarkComplianceSink []any

func BenchmarkComplianceSuite(b *testing.B) {
	reg := jpath.NewRegistry()
	cases := benchmarkLoadComplianceCases(b)

	b.Run("parse-compile-run", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, tc := range cases {
				result, err := reg.Query(tc.Selector, tc.Document)
				if err != nil {
					b.Fatalf("query failed for %q: %v", tc.Selector, err)
				}
				benchmarkComplianceSink = result
			}
		}
	})

	compiled := benchmarkCompileCases(b, reg, cases)

	b.Run("run-precompiled", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, tc := range compiled {
				benchmarkComplianceSink = tc.path(tc.document)
			}
		}
	})
}

func benchmarkCompileCases(
	b *testing.B, reg *jpath.Registry, cases []complianceCase,
) []benchmarkCase {
	b.Helper()
	result := make([]benchmarkCase, 0, len(cases))
	for _, tc := range cases {
		ast, err := reg.Parse(tc.Selector)
		if err != nil {
			b.Fatalf("parse failed for %q: %v", tc.Selector, err)
		}
		path, err := reg.Compile(ast)
		if err != nil {
			b.Fatalf("compile failed for %q: %v", tc.Selector, err)
		}
		result = append(result, benchmarkCase{
			document: tc.Document,
			path:     path,
		})
	}
	return result
}

func benchmarkLoadComplianceCases(b *testing.B) []complianceCase {
	b.Helper()
	suitePath := filepath.Join(
		"testdata", "jsonpath-compliance-test-suite", "cts.json",
	)
	if env := os.Getenv("JSONPATH_CTS_FILE"); env != "" {
		suitePath = env
	}
	buf, err := os.ReadFile(suitePath)
	if err != nil {
		b.Fatalf("compliance suite unavailable at %s: %v", suitePath, err)
	}
	var suite complianceSuite
	if err := json.Unmarshal(buf, &suite); err != nil {
		b.Fatalf("invalid compliance suite %s: %v", suitePath, err)
	}
	result := make([]complianceCase, 0, len(suite.Tests))
	for _, tc := range suite.Tests {
		if tc.InvalidSelector {
			continue
		}
		result = append(result, tc)
	}
	if len(result) == 0 {
		b.Fatal("compliance suite does not contain valid selectors")
	}
	return result
}
