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

// Package integerintervalexpressions is a library for parsing integer interval
// expressions of the form '1,3-5,7-'
package integerintervalexpressions

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// subExpression represents a single continuous interval
type subExpression struct {
	start int
	count int
}

func (se subExpression) String() string {
	switch se.count {
	case 0:
		return fmt.Sprintf("%d-", se.start)
	case 1:
		return fmt.Sprintf("%d", se.start)
	default:
		return fmt.Sprintf("%d-%d", se.start, se.start+se.count-1)
	}
}

// Expression is an abstract type containing a sequence of subexpressions
// describing integer intervals. An Expression instance can only be constructed
// by ParseExpression() from a valid expression string; see the README.md for
// information about the syntax.
//
// The Expression only has one useful method: Matches(int), which tells you whether
// the given value lies inside any of the intervals contained within the expression.
type Expression struct {
	intervals []subExpression
	opts      ParseOptions // original options used for parsing this Expression
}

// Matches determines whether an integer is contained within the intervals expression
//
// For example, given si := ParseInterval("1-3,5,9-10,13-"), the expressions
// si.Matches(2) and si.Matches(15) return true, while si.Matches(4) will return false.
//
// This method does not require the Expression to be normalized, although
// normalized instances *should* allow for quicker evaluation due to reduced
// number of interval elements inside Expression (hint: do not begin
// to optimize prematurely, you are unlikely to ever need Expression of
// such sizes that this becomes an issue).
func (si Expression) Matches(val int) bool {
	for _, itv := range si.intervals {
		if val >= itv.start {
			if itv.count == 0 || val <= (itv.start+itv.count-1) {
				return true
			}
		}
	}
	return false
}

// ParseOptions adjusts how the ParseExpression function will interpret the input
type ParseOptions struct {
	Delimiter            string
	PostProcessNormalize bool
	//openEnd bool // 1-3 stands for 1,2,3 or 1,2?
	//greedy  bool // 2-4,2,2- -> which is actually dominant?
}

// DefaultParseOptions returns some sensible set of options for default usage.
func DefaultParseOptions() ParseOptions {
	return ParseOptions{Delimiter: ",", PostProcessNormalize: false}
}

// Normalize reduces overlapping expressions to minimum set of intervals;
// some new interval elements may be totally new, while others are dropped.
// For example, expression '1-4,2-5' should normalize to '1-5'.
func (si Expression) Normalize() Expression {
	if len(si.intervals) <= 1 {
		return si
	}

	// this code assumes that now intervals are ordered by start value
	sort.Slice(si.intervals, func(a int, b int) bool {
		return si.intervals[a].start < si.intervals[b].start
	})

	var norm []subExpression

	current := si.intervals[0]

	for i := 1; i < len(si.intervals); i++ {
		next := si.intervals[i]
		if current.count == 0 {
			// extends to infinity, we can skip
			break
		}
		currentEnd := current.start + current.count - 1

		// next.start is inside interval curr, or next.start is immediately next
		// after last value in curr.
		if (next.start - currentEnd) <= 1 {
			if next.count == 0 {
				// next extends to infinity, we can stop
				current.count = 0
				break
			} else {
				// next is absorbed into current
				nextEnd := next.start + next.count - 1
				current.count = nextEnd - current.start + 1
			}
		} else {
			// next interval is outside/non-adjacent to currentent
			norm = append(norm, current)
			current = next
			if current.count == 0 {
				break
			}
		}
	}
	norm = append(norm, current)
	return Expression{intervals: norm}
}

// Convert Expression back to textual format.
// Note the following case:
//
//   Expr, _ := ParseExpression(Input) // assume input is valid
//   Norm    := Expr.Normalize()
//
// now Expr.String() should resemble Input, HOWEVER if Expr != Norm, then
// Norm.String() likely nothing close to Input. That is, a normalized
// Expression is unlikely to serialize back to the original input string unless
// the input was written in normalized form to begin with.
func (si Expression) String() string {
	var ivs []string
	for _, itv := range si.intervals {
		ivs = append(ivs, itv.String())
	}
	return strings.Join(ivs, si.opts.Delimiter)
}

// ParseExpression calls ParseExpressionWithOptions with default options (see DefaultParseOptions())
func ParseExpression(input string) (Expression, error) {
	return ParseExpressionWithOptions(input, DefaultParseOptions())
}

// ParseExpressionWithOptions attempts to extract list of interval expressions from 'input'.
// A single interval is expressed with:
// - a single integer (e.g '7')
// - a single integer and a dash (e.g. '1-', meaning 1,2,3,4,...) It is basically 'x ... inf'
// - a single integer, a dash, and another integer (e.g. '1-4', meaning 1,2,3,4). It is an error
//   to supply expression 'a-b' where a > b.
// In all expressions the integers are assumed positive or 0.
// NOTE: this library does not support notation '-x' for open-ended interval -inf...x.
//
// Multiple intervals are expressed by placing them between the delimiter character, which is
// by default ",". For example, input '0,1,4-7,9,11-12' means integer values 0,1,4,5,6,9,11,12.
//
// The input string can contain 0 or more interval expressions, which means
// that empty string is valid input. It also is valid to pass empty string
// between two commas; these empty interval expressions are skipped.
//
// In case of problems in parsing, the function returns an error and an empty Expression{}.
// The errors are constructed with fmt.Errorf, and contain description of what exactly is wrong
// with the given input.
//
// Note: the input is not guaranteed to be normalized; for that you should use Normalize()
func ParseExpressionWithOptions(input string, opts ParseOptions) (Expression, error) {
	intervalsRaw := strings.Split(input, opts.Delimiter)
	var intervals []subExpression
	for _, intervalStr := range intervalsRaw {
		if intervalStr == "" {
			// empty expression '1,,3'.. not very pretty but not invalid
			continue
		}
		split := strings.Split(intervalStr, "-")

		n := len(split)

		if n != 1 && n != 2 {
			return Expression{}, fmt.Errorf("invalid interval expression: %q", intervalStr)
		}

		a, err := strconv.ParseInt(split[0], 10, 0)
		if err != nil {
			return Expression{}, fmt.Errorf("invalid value for interval start: %w", err)
		}

		// single digit, interval of length 1
		if n == 1 {
			intervals = append(intervals, subExpression{start: int(a), count: 1})
			continue
		}

		// implicit n == 2 cases

		// digit and a dash: 'x-', interval length infinite (internally denoted with 0)
		if split[1] == "" {
			intervals = append(intervals, subExpression{start: int(a), count: 0})
			continue
		}

		// two digits separated by dash
		if b, err := strconv.ParseInt(split[1], 10, 0); err != nil {
			return Expression{}, fmt.Errorf("invalid value for interval end: %w", err)
		} else if b < a {
			return Expression{}, fmt.Errorf("invalid interval 'a-b' where a > b: %q", intervalStr)
		} else {
			intervals = append(intervals, subExpression{start: int(a), count: (int(b) - int(a)) + 1})
		}
	}
	si := Expression{intervals: intervals, opts: opts}
	if opts.PostProcessNormalize {
		return si.Normalize(), nil
	}
	return si, nil
}
