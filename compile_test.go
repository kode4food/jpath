package jpath_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kode4food/jpath"
)

type (
	ctsSuite struct {
		Tests []ctsCase `json:"tests"`
	}

	ctsCase struct {
		Name            string `json:"name"`
		Selector        string `json:"selector"`
		InvalidSelector bool   `json:"invalid_selector"`
	}
)

func TestNewCompiler(t *testing.T) {
	c := jpath.NewCompiler()
	assert.NotNil(t, c)
}

func TestCompileSelectors(t *testing.T) {
	ctsPath := filepath.Join(
		"testdata", "jsonpath-compliance-test-suite", "cts.json",
	)
	buf, err := os.ReadFile(ctsPath)
	if err != nil {
		t.Fatalf("compliance suite unavailable at %s: %v", ctsPath, err)
	}
	var suite ctsSuite
	if !assert.NoError(t, json.Unmarshal(buf, &suite)) {
		return
	}

	reg := jpath.NewRegistry()
	for _, tc := range suite.Tests {
		if tc.InvalidSelector {
			continue
		}
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			ast, err := reg.Parse(tc.Selector)
			if !assert.NoError(t, err, tc.Selector) {
				return
			}
			path, err := reg.Compile(ast)
			if !assert.NoError(t, err, tc.Selector) {
				return
			}
			assert.NotNil(t, path)
			assert.NotPanics(t, func() {
				_ = path(nil)
			})
		})
	}
}
