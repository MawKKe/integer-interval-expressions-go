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
//
// Expressions of this kind are commonly seen in user-facing application contexts
// such as page selectors in print dialogs, field selector in the CLI `cut` tool,
// and so on. This library provides support for parsing and utilizing such
// expressions in wide variety of application contexts.
//
// Internally, the library parses an input string into an abstract logical
// expression, which can be then evaluated with integer values to determine
// whether those values lie in any of the specified intervals. The parsed
// expressions do not contain any actual integer sequences, which allows for
// small memory usage and support for infinite ranges
package integerintervalexpressions

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// subExpression represents a single continuous interval
type subExpression struct {
	start    int
	count    int
	matchAll bool
}

func (se subExpression) String() string {
	if se.matchAll {
		return "*"
	}

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
// by ParseExpression() from a valid expression string.
//
// The Expression only has one useful method: Matches(int), which tells you whether
// the given value lies inside any of the intervals contained within the expression.
type Expression struct {
	intervals []subExpression
	opts      ParseOptions // original options used for parsing this Expression
}

// MatchesNone determines whether the Expression will ever match anything.
// i.e if MatchesNone() == true; then Matches(x) == false for all x.
//
// Such Expressions are the result of input expression strings that contain no
// actual subexpressions. You may instruct the expression parser to reject
// such input expressions by setting ParseOptions.AllowEmptyExpression to
// false; the current default options also have this field set as false (see
// DefaultParseOptions()).
func (e Expression) MatchesNone() bool {
	return len(e.intervals) == 0
}

// MatchesAll determines whether the Expression will match every possible input
// i.e if MatchesAll() == true; then Matches(x) == true for all x.
func (e Expression) MatchesAll() bool {
	for _, sub := range e.intervals {
		if sub.matchAll {
			return true
		}
	}
	return false
}

// Matches determines whether an integer is contained within the intervals expression
//
// For example, given
//   expr, _ := ParseExpression("1,3-5,7-")
// the expressions
//   expr.Matches(1)
//   expr.Matches(4)
//   expr.Matches(9)
// evaluate to true, while
//   expr.Matches(2)
//   expr.Matches(6)
// evaluate to false
//
// This method does not require the Expression to be normalized, although
// normalized instances *should* allow for quicker evaluation due to reduced
// number of interval elements in the Expression; see .Normalize().
func (e Expression) Matches(val int) bool {
	for _, itv := range e.intervals {
		if itv.matchAll {
			return true
		}
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

	// Allow parsing of empty input expressions strings (e.g "" or "   ")?
	// If false, parser will return error on empty input.
	// If true, empty input will result in Expression that will match nothing.
	AllowEmptyExpression bool

	//openEnd bool // 1-3 stands for 1,2,3 or 1,2?
	//greedy  bool // 2-4,2,2- -> which is actually dominant?
}

// DefaultParseOptions returns some sensible set of options for default usage.
func DefaultParseOptions() ParseOptions {
	return ParseOptions{
		Delimiter:            ",",
		PostProcessNormalize: false,
		// Do not allow empty expressions by default; empty expressions
		// match nothing, and likely confuse users.
		AllowEmptyExpression: false,
	}
}

// Normalize reduces overlapping expressions to minimum set of intervals;
// some new interval elements may be totally new, while others are dropped.
// For example, expression '1-4,2-5' should normalize to '1-5'.
// The method returns a new normalized Expression derived from the current
// one.
func (e Expression) Normalize() Expression {
	// short-circuit by empty expression
	// no need to do anything, just return the existing expression
	if e.MatchesNone() {
		return e
	}

	// short-circuit by "*"
	// TODO MatchesAll() loops the intervals, but so do we below. Figure out a
	// way to integrate this check into the main loop below?
	if e.MatchesAll() {
		return Expression{intervals: []subExpression{{matchAll: true}}, opts: e.opts}
	}

	// this code assumes that now intervals are ordered by start value
	sort.Slice(e.intervals, func(a int, b int) bool {
		return e.intervals[a].start < e.intervals[b].start
	})

	var norm []subExpression

	current := e.intervals[0]

	for i := 1; i < len(e.intervals); i++ {
		next := e.intervals[i]
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
//
// Consider the following situation
//
//   // Assume Input is valid for brevity
//   Expr, _ := ParseExpression(Input)
//   Norm    := Expr.Normalize()
//
// Now, the result of Expr.String() should resemble Input. However, if Expr !=
// Norm, then Norm.String() likely differs greatly from Input. That is, a
// normalized Expression is unlikely to serialize back to the original input
// string (unless the input was written in normalized form to begin with).
func (e Expression) String() string {
	var ivs []string
	for _, itv := range e.intervals {
		ivs = append(ivs, itv.String())
	}
	return strings.Join(ivs, e.opts.Delimiter)
}

// ParseExpression calls ParseExpressionWithOptions() with default options (see DefaultParseOptions())
func ParseExpression(input string) (Expression, error) {
	return ParseExpressionWithOptions(input, DefaultParseOptions())
}

// ParseExpressionWithOptions attempts to extract intervals expressions from input.
//
// ---
//
// An intervals expression consists of sequence of individual subexpressions.
//
// A subexpression describes a continuous range of integral values (i.e an
// interval).  A single subexpression string contains one of the following:
//
// - an single integer, for example "1": only the value 1.
//
// - an integer, a dash, and another integer, for example "3-5": values 3,4 and 5.
//
// - an integer and a dash, for example "7-": denotes all integers from 7 to
// infinity (i.e 7,8,9,...)
//
// Currently the parser supports only positive integer values in subexpressions.
//
// Additionally, the parser recognizes a subexpressions equal to "*" and
// interprets them as "match everything". Note that such subexpression will
// dominate over all others, short-circuiting the whole expression to "true".
//
// The intervals expression is consists of subexpressions joined by a delimiter
// character.  By default, a comma (",") is used as the delimiter (although a
// custom delimiter can be specified via the "ParseOptions" structure). For
// example, the expression "1,3-5,7-" can be understood to contain three
// subexpressions: "1", "3-5" and "7-".
//
// Note that the interval expression need not contain any subexpressions, which
// means that "" and ",,,," are valid inputs. However, both of these parse into
// an Expression structure containing 0 subexpressions and are, as such,
// rather useless.
//
// Semantically, a single subexpression is a predicate, and combining multiple
// predicates denotes a logical disjunction. The above expression thus states that
// we have three predicates and an overall expression:
//
//     func a(x int) { return x == 1 }             // "1"
//     func b(x int) { return x >= 3 && x <= 5 }   // "3-5"
//     func c(x int) { return x >= 7 }             // "7-"
//     func expr(x int) { return a(x) || b(x) || c(x) }
//
// (However note that in the library internals the expressions are not actually
// represented this way.)
//
// Note that the library does not support parsing expressions with spaces
// inside subexpressions, or between the subexpressions and delimiters. This may
// change in future version.
//
// ---
//
// Return values:
//
// In case of invalid/malformed input, the function returns an error and an
// empty Expression{}. The errors are constructed with fmt.Errorf, and should
// contain description of what exactly is wrong with the given input.
//
// A valid input string is parsed into a populated Expression, which
// can then be evaluated using the associated methods.
//
// NOTE: The resulting Expression is not guaranteed to be normalized, unless
// you set opts.PostProcessNormalize=true, or manually call .Normalize() on the result.
func ParseExpressionWithOptions(input string, opts ParseOptions) (Expression, error) {
	if opts.Delimiter == "" {
		return Expression{}, fmt.Errorf("ParseOptions.Delimiter is empty")
	}
	intervalsRaw := strings.Split(input, opts.Delimiter)
	var intervals []subExpression
	for _, intervalStr := range intervalsRaw {
		if intervalStr != "" {
			// empty expression '1,,3'.. not very pretty but not invalid
			interval, err := parseSubExpression(intervalStr)
			if err != nil {
				return Expression{}, err
			}
			intervals = append(intervals, interval)
		}
	}

	e := Expression{intervals: intervals, opts: opts}

	if e.MatchesNone() && !opts.AllowEmptyExpression {
		return Expression{}, fmt.Errorf("current options prohibit empty expressions")
	}

	if opts.PostProcessNormalize {
		return e.Normalize(), nil
	}
	return e, nil
}

var subRegexMatchall = regexp.MustCompile(`^\s*\*\s*$`)
var subRegexSingle = regexp.MustCompile(`^\s*(?P<start>\d+)\s*$`)
var subRegexDual = regexp.MustCompile(`^\s*(?P<start>\d+)-(?P<end>\d+)$`)
var subRegexHalfOpen = regexp.MustCompile(`^\s*(?P<start>\d+)(-)$`)

func parseSubExpression(subInput string) (subExpression, error) {
	if subRegexMatchall.MatchString(subInput) {
		return subExpression{matchAll: true}, nil
	}

	if m := subRegexSingle.FindStringSubmatch(subInput); m != nil {
		start := m[subRegexSingle.SubexpIndex("start")]
		if v, err := strconv.ParseInt(start, 10, 0); err != nil {
			return subExpression{}, fmt.Errorf("invalid value for interval start: %w", err)
		} else {
			return subExpression{start: int(v), count: 1}, nil
		}
	}

	if m := subRegexHalfOpen.FindStringSubmatch(subInput); m != nil {
		start := m[subRegexHalfOpen.SubexpIndex("start")]
		if v, err := strconv.ParseInt(start, 10, 0); err != nil {
			return subExpression{}, fmt.Errorf("invalid value for interval start: %w", err)
		} else {
			return subExpression{start: int(v), count: 0}, nil
		}
	}

	if m := subRegexDual.FindStringSubmatch(subInput); m != nil {
		start := m[subRegexDual.SubexpIndex("start")]
		end := m[subRegexDual.SubexpIndex("end")]
		var vStart, vEnd int64
		var err error
		if vStart, err = strconv.ParseInt(start, 10, 0); err != nil {
			return subExpression{}, fmt.Errorf("invalid value for interval start: %w", err)
		}
		if vEnd, err = strconv.ParseInt(end, 10, 0); err != nil {
			return subExpression{}, fmt.Errorf("invalid value for interval end: %w", err)
		}
		if vEnd < vStart {
			return subExpression{}, fmt.Errorf("invalid interval 'a-b' where a > b: %q", subInput)
		}
		a, b := int(vStart), int(vEnd)
		c := b - a + 1
		return subExpression{start: a, count: c}, nil
	}

	return subExpression{}, fmt.Errorf("invalid syntax: %q", subInput)
}
