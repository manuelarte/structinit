package analyzer

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		patterns string
		options  map[string]string
	}{
		"simple": {
			patterns: "simple",
		},
		"comments": {
			patterns: "comments",
		},
		"side effects": {
			patterns: "side-effects",
		},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			a := New()

			for k, v := range test.options {
				err := a.Flags.Set(k, v)
				if err != nil {
					t.Fatal(err)
				}
			}

			analysistest.Run(t, analysistest.TestData(), a, test.patterns)
		})
	}
}

func TestAnalyzerWithFix(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		patterns string
		options  map[string]string
	}{
		"simple fix": {
			patterns: "simple",
		},
		"comments fix": {
			patterns: "comments",
		},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			a := New()

			for k, v := range test.options {
				err := a.Flags.Set(k, v)
				if err != nil {
					t.Fatal(err)
				}
			}

			analysistest.RunWithSuggestedFixes(t, analysistest.TestData(), a, test.patterns)
		})
	}
}
