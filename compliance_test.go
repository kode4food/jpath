package jpath_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kode4food/jpath"
)

type (
	complianceSuite struct {
		Tests []complianceCase `json:"tests"`
	}

	complianceCase struct {
		Name            string   `json:"name"`
		Selector        string   `json:"selector"`
		Document        any      `json:"document"`
		Result          []any    `json:"result"`
		Results         [][]any  `json:"results"`
		InvalidSelector bool     `json:"invalid_selector"`
		Tags            []string `json:"tags"`
	}
)

func TestComplianceSuite(t *testing.T) {
	reg := jpath.NewRegistry()
	path := filepath.Join(
		"testdata", "jsonpath-compliance-test-suite", "cts.json",
	)
	if env := os.Getenv("JSONPATH_CTS_FILE"); env != "" {
		path = env
	}
	buf, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("compliance suite unavailable at %s: %v", path, err)
	}
	var suite complianceSuite
	if !assert.NoError(t, json.Unmarshal(buf, &suite)) {
		return
	}
	for _, tc := range suite.Tests {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			p, err := reg.Query(tc.Selector)
			if tc.InvalidSelector {
				assert.Error(t, err)
				return
			}
			if !assert.NoError(t, err) {
				return
			}
			got := p.Query(tc.Document)
			if got == nil {
				got = []any{}
			}
			if len(tc.Results) > 0 {
				for _, expected := range tc.Results {
					if reflect.DeepEqual(expected, got) {
						return
					}
				}
				assert.Failf(t, "unexpected result", "%#v", got)
				return
			}
			assert.Equal(t, tc.Result, got)
		})
	}
}
