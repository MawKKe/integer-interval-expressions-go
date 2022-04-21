// Copyright 2022 Markus Holmström (MawKKe)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package integerintervalexpressions

import (
	"fmt"
	"reflect"
	"testing"
)

func ExampleParseExpression() {
	input := "1,3-5,7-"
	myExpr, err := ParseExpression(input)
	if err != nil {
		fmt.Println(err)
		return
	}
	for i := 0; i < 10; i++ {
		fmt.Printf("%d: %v\n", i, myExpr.Matches(i))
	}
	// Output:
	// 0: false
	// 1: true
	// 2: false
	// 3: true
	// 4: true
	// 5: true
	// 6: false
	// 7: true
	// 8: true
	// 9: true
}

type testCase struct {
	name      string
	input     string
	shouldErr bool
	expected  Expression
}

var testCases []testCase = []testCase{
	{
		// note: current default options prohibit empty expressions.
		// TODO test with AllowEmptyExpressions=true ?
		name:      "empty",
		input:     "",
		shouldErr: true,
		expected:  Expression{}, // note: not same as Expression{intervals: []subExpression}
	},
	{
		// note: current default options prohibit empty expressions.
		// TODO test with AllowEmptyExpressions=true ?
		name:      "empty-commas",
		input:     ",,,,",
		shouldErr: true,
		expected:  Expression{}, // note: not same as Expression{intervals: []subExpression}
	},
	{
		name:      "single-interval-single-digit-exact-0",
		input:     "0",
		shouldErr: false,
		expected: Expression{
			intervals: []subExpression{
				subExpression{start: 0, count: 1},
			},
		},
	},
	{
		name:      "single-interval-single-digit-open-1",
		input:     "1-",
		shouldErr: false,
		expected: Expression{
			intervals: []subExpression{
				subExpression{start: 1, count: 0}, // 0 count means inf
			},
		},
	},
	{
		name:      "single-interval-double-digit-5-7",
		input:     "5-7",
		shouldErr: false,
		expected: Expression{
			intervals: []subExpression{
				subExpression{start: 5, count: 3},
			},
		},
	},
	{
		name:      "multiple-interval-double-digit-5-7",
		input:     "5-7,9-10",
		shouldErr: false,
		expected: Expression{
			intervals: []subExpression{
				subExpression{start: 5, count: 3},
				subExpression{start: 9, count: 2},
			},
		},
	},
	{
		name:      "multiple-interval-with-gaps",
		input:     ",1,,5-7,,9-10,,17-",
		shouldErr: false,
		expected: Expression{
			intervals: []subExpression{
				subExpression{start: 1, count: 1},
				subExpression{start: 5, count: 3},
				subExpression{start: 9, count: 2},
				subExpression{start: 17, count: 0},
			},
		},
	},
	{
		name:      "multiple-interval-with-gaps-and-open-interval-in-the-middle",
		input:     ",1,,5-7,,2-,9-10,,17-",
		shouldErr: false,
		expected: Expression{
			intervals: []subExpression{
				subExpression{start: 1, count: 1},
				subExpression{start: 5, count: 3},
				subExpression{start: 2, count: 0}, // in normalization this would eliminate all except first
				subExpression{start: 9, count: 2},
				subExpression{start: 17, count: 0},
			},
		},
	},
	{
		name:      "single-interval-invalid-single-value",
		input:     "a",
		shouldErr: true,
		expected:  Expression{},
	},
	{
		name:      "single-interval-invalid-start",
		input:     "a-3",
		shouldErr: true,
		expected:  Expression{},
	},
	{
		name:      "single-interval-invalid-end",
		input:     "1-b",
		shouldErr: true,
		expected:  Expression{},
	},
	{
		name:      "single-interval-invalid-range-start-nodash",
		input:     "1@",
		shouldErr: true,
		expected:  Expression{},
	},
	{
		name:      "single-interval-invalid-range",
		input:     "1@3",
		shouldErr: true,
		expected:  Expression{},
	},
	{
		name:      "single-interval-invalid-range-start-missing",
		input:     "-1",
		shouldErr: true,
		expected:  Expression{},
	},
	{
		name:      "single-interval-invalid-range-start-missing-with-non-integer",
		input:     "-x",
		shouldErr: true,
		expected:  Expression{},
	},
	{
		name:      "multiple-interval-invalid-single-valid-otherwise",
		input:     "x,3-5,7-9,10",
		shouldErr: true,
		expected:  Expression{},
	},
	{
		name:      "multiple-interval-valid-first-invalid-second",
		input:     "1-3,6-x",
		shouldErr: true,
		expected:  Expression{},
	},
	{
		name:      "multiple-interval-valid-until-last-open-invalid",
		input:     "1-3,6-8,x-",
		shouldErr: true,
		expected:  Expression{},
	},
}

func TestMatchesNone(t *testing.T) {
	optsAllow := ParseOptions{Delimiter: ",", AllowEmptyExpression: true}
	optsDisallow := ParseOptions{Delimiter: ",", AllowEmptyExpression: false}

	cases := []struct {
		name              string
		opts              ParseOptions
		input             string
		shouldErr         bool
		expectMatchesNone bool
	}{
		{
			name:              "empty-allowed",
			opts:              optsAllow,
			input:             "",
			shouldErr:         false,
			expectMatchesNone: true,
		},
		{
			name:              "non-empty-allowed",
			opts:              optsAllow,
			input:             "1-3",
			shouldErr:         false,
			expectMatchesNone: false,
		},
		{
			name:              "empty-disallowed",
			opts:              optsDisallow,
			input:             "",
			shouldErr:         true,
			expectMatchesNone: true,
		},
		{
			name:              "non-empty-disallowed",
			opts:              optsDisallow,
			input:             "1-3",
			shouldErr:         false,
			expectMatchesNone: false,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			expr, err := ParseExpressionWithOptions(test.input, test.opts)
			if test.shouldErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if a, b := test.expectMatchesNone, expr.MatchesNone(); a != b {
				t.Fatalf("expected MatchesNone() == %v, got %v", b, a)
			}
		})
	}
}

func TestMatchesAll(t *testing.T) {
	cases := []struct {
		input          string
		matchAllExpect bool
	}{
		{
			input:          "*",
			matchAllExpect: true,
		},
		{
			input:          "1,3-5,7-",
			matchAllExpect: false,
		},
		{
			input:          "*,1,3-5,7-",
			matchAllExpect: true,
		},
		{
			input:          "1,*,3-5,7-",
			matchAllExpect: true,
		},
		{
			input:          "1,3-5,*,7-",
			matchAllExpect: true,
		},
		{
			input:          "1,3-5,7-,*",
			matchAllExpect: true,
		},
		{
			input:          "*,*,*,*",
			matchAllExpect: true,
		},
	}
	for _, test := range cases {
		expr, err := ParseExpression(test.input)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if a, b := test.matchAllExpect, expr.MatchesAll(); a != b {
			t.Fatalf("expected MatchesAll() == %v, got %v", a, b)
		}
	}
}

func TestParseExpression(t *testing.T) {
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			res, err := ParseExpression(test.input)
			if test.shouldErr && err == nil {
				t.Fatalf("Expected error, got <nil> instead")
			}
			if !test.shouldErr && err != nil {
				t.Fatalf("Got error: %v, expected <nil>", err)
			}
			nGot := len(res.intervals)
			nExpt := len(test.expected.intervals)
			if err == nil && (nGot != nExpt || !reflect.DeepEqual(res.intervals, test.expected.intervals)) {
				t.Fatalf("Expected result (n = %d):\n\t%#v\nGot (n = %d):\n\t%#v", nExpt, test.expected, nGot, res)
			}
		})
	}
}

type normalizeTest struct {
	name   string
	input  Expression
	expect Expression
}

var normalizeTests []normalizeTest = []normalizeTest{
	normalizeTest{
		// 0 elements
		name:   "simple-empty",
		input:  Expression{},
		expect: Expression{},
	},
	normalizeTest{
		// 1 element
		name: "simple-single-individual",
		input: Expression{intervals: []subExpression{
			subExpression{start: 2, count: 1},
		}},
		expect: Expression{intervals: []subExpression{
			subExpression{start: 2, count: 1},
		}},
	},
	normalizeTest{
		// 3 elements
		name: "simple-individual-consecutive-ordered",
		input: Expression{intervals: []subExpression{
			subExpression{start: 1, count: 1},
			subExpression{start: 2, count: 1},
			subExpression{start: 3, count: 1},
		}},
		expect: Expression{intervals: []subExpression{
			subExpression{start: 1, count: 3},
		}},
	},
	normalizeTest{
		// 3 elements
		name: "simple-individual-consecutive-ordered-with-gaps",
		input: Expression{intervals: []subExpression{
			subExpression{start: 1, count: 1},
			subExpression{start: 3, count: 1},
			subExpression{start: 10, count: 1},
		}},
		expect: Expression{intervals: []subExpression{
			subExpression{start: 1, count: 1},
			subExpression{start: 3, count: 1},
			subExpression{start: 10, count: 1},
		}},
	},
	normalizeTest{
		// 3 elements
		name: "simple-individual-consecutive-non-ordered-with-gaps",
		input: Expression{intervals: []subExpression{
			subExpression{start: 10, count: 1},
			subExpression{start: 3, count: 1},
			subExpression{start: 1, count: 1},
		}},
		expect: Expression{intervals: []subExpression{
			subExpression{start: 1, count: 1},
			subExpression{start: 3, count: 1},
			subExpression{start: 10, count: 1},
		}},
	},
	normalizeTest{
		// 2 elements
		name: "simple-overlapping-ordered",
		input: Expression{intervals: []subExpression{
			subExpression{start: 1, count: 2},
			subExpression{start: 2, count: 2},
		}},
		expect: Expression{intervals: []subExpression{
			subExpression{start: 1, count: 3},
		}},
	},
	normalizeTest{
		// 2 elements
		name: "simple-overlapping-not-ordered",
		input: Expression{intervals: []subExpression{
			subExpression{start: 2, count: 2},
			subExpression{start: 1, count: 2},
		}},
		expect: Expression{intervals: []subExpression{
			subExpression{start: 1, count: 3},
		}},
	},
	normalizeTest{
		// 2 elements
		name: "simple-disjoint-ordered",
		input: Expression{intervals: []subExpression{
			subExpression{start: 1, count: 2},
			subExpression{start: 4, count: 3},
		}},
		expect: Expression{intervals: []subExpression{
			subExpression{start: 1, count: 2},
			subExpression{start: 4, count: 3},
		}},
	},
	normalizeTest{
		// 2 elements
		name: "simple-overlapping-ordered-unbounded",
		input: Expression{intervals: []subExpression{
			subExpression{start: 2, count: 3},
			subExpression{start: 3, count: 0},
		}},
		expect: Expression{intervals: []subExpression{
			subExpression{start: 2, count: 0},
		}},
	},
	normalizeTest{
		// "1,1-" i.e overlapping
		name: "simple-redundant-overlapping-zeros",
		input: Expression{intervals: []subExpression{
			subExpression{start: 1, count: 1},
			subExpression{start: 0, count: 0},
		}},
		expect: Expression{intervals: []subExpression{
			subExpression{start: 0, count: 0},
		}},
	},
	normalizeTest{
		// "1,1-" i.e overlapping
		name: "simple-redundant-overlapping",
		input: Expression{intervals: []subExpression{
			subExpression{start: 2, count: 1},
			subExpression{start: 2, count: 0},
		}},
		expect: Expression{intervals: []subExpression{
			subExpression{start: 2, count: 0},
		}},
	},
	normalizeTest{
		// 1,5-7,2-,9-10,17-
		name: "random-complicated-expression",
		input: Expression{intervals: []subExpression{
			subExpression{start: 1, count: 1},
			subExpression{start: 5, count: 2},
			subExpression{start: 2, count: 0},
			subExpression{start: 9, count: 2},
			subExpression{start: 17, count: 0},
		}},
		expect: Expression{intervals: []subExpression{
			subExpression{start: 1, count: 0},
		}},
	},
	normalizeTest{
		// 1,5-7,2-,9-10,17-
		name: "match-all-1",
		input: Expression{intervals: []subExpression{
			subExpression{start: 1, count: 1},
			subExpression{start: 5, count: 2},
			subExpression{matchAll: true},
			subExpression{start: 2, count: 0},
			subExpression{start: 9, count: 2},
			subExpression{start: 17, count: 0},
		}},
		expect: Expression{intervals: []subExpression{
			subExpression{matchAll: true},
		}},
	},
}

func TestNormalize(t *testing.T) {
	for _, test := range normalizeTests {
		t.Run(test.name, func(t *testing.T) {
			got := test.input.Normalize()
			if !reflect.DeepEqual(test.expect, got) {
				t.Fatalf("\nInput:\n\t%v\nExpect:\n\t%v\nGot:\n\t%v", test.input, test.expect, got)
			}
		})
	}
}

func TestExpressionStringer(t *testing.T) {
	// TODO add better tests

	inputs := []string{"1-3,4,10-", "1,3-5,*,7-"}

	for _, input := range inputs {
		expr, err := ParseExpression(input)

		if err != nil {
			t.Fatal(err)
		}

		str := expr.String()

		if str != input {
			t.Fatalf("expected: %q, got: %q", input, str)
		}

		if out := fmt.Sprintf("%v", expr); out != input {
			t.Fatalf("Expected: %q, got: %q", input, out)
		}
	}
}
