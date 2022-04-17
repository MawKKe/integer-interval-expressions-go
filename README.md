# integer-interval-expressions-go
A Go library for parsing integer interval expressions of the form `1,3-5,7-`

Expressions of this kind are commonly seen in user-facing application contexts
such as page selectors in print dialogs, field selector in the CLI `cut` tool,
and so on. This library provides support for parsing and utilizing such
expressions in wide variety of application contexts.

Internally, the library parses an input string into an abstract logical expression,
which can be then evaluated with integer values to determine whether those
values lie in any of the specified intervals. The parsed expressions do not contain
any actual integer sequences, which allows for small memory usage and support for
infinite ranges

[![Go](https://github.com/MawKKe/integer-interval-expressions-go/workflows/Go/badge.svg)](https://github.com/MawKKe/integer-interval-expressions-go/actions/workflows/go.yml)

# Install

Add the library into your project:

```Shell
$ go get github.com/MawKKe/integer-interval-expressions-go@latest
```

then import the library in your application:

```Go
import (
    ...
    intervals "github.com/MawKKe/integer-interval-expressions-go"
    ...
)
```

# Usage

The library is quite simple to use. The primary function is `ParseExpression`
which takes a string and returns the parsed intervals expression as an abstract
`Expression` object:

```Go
inputString := "1,3-5,7-"
myExpr, err := intervals.ParseExpression(inputString)
if err != nil {
    fmt.Println("error:", err)
    return
}
fmt.Println("got valid expression:", myExpr)
```

Now you can evaluate the expression with various values:

```Go
myExpr.Matches(1) // == true
myExpr.Matches(2) // == false
myExpr.Matches(3) // == true
myExpr.Matches(4) // == true
myExpr.Matches(6) // == false
myExpr.Matches(7) // == true
myExpr.Matches(8) // == true
myExpr.Matches(9) // == true, is so for all values >=7
```
As you see, the expression evaluates to true only on the specified integer
intervals.

**NOTE**: The `ParseExpression()` is merely a convenience function, while the
actual work is performed by `ParseExpressionWithOptions()`. The difference is
that the latter accepts a `ParseOptions` struct in addition to the input
string; the options can be used for adjusting the operation of the parser to
suit your needs. See the `go doc` documentation for more information. You may
also be interested in the example and test functions in `expr_test.go`

Thats pretty much all there is to it. How you actually use this functionality
is up to you.  A typical use case is to iterate your application data entries
and check each one against the expression:

```Go
for _, page := range MyDocument.Pages {
    if myExpr.Matches(page.Number){
        PrintPage(page)
    }
}
```
etc.

## Syntax

See the documentation for function `ParseExpressionWithOptions` for description
of supported intervals expression syntax.

## Optimization
The expression parser will happily process an input containing duplicate or
overlapping subexpressions. Likewise, the order of subexpressions is irrelevant
to the parser. However, poorly constructed expressions may result in
unsatisfactory `Expression.Match` performance. To overcome such issues, the
`Expression` instance can be simplified via its `Normalize()` method. The
method sorts the subexpression ranges and merges overlapping ones, producing a
minimal set of subexpressions. The resulting new `Expression` should be
semantically equivalent to the original, while being more performant from practical
perspective. 

**NOTE**: The parser does not perform the normalization automatically, unless
`ParseOptions::PostProcessNormalize` is set to true.

**NOTE**: Normalization may significantly change how an `Expression` is 
represented in text form. See the documentation for `Expression.String()`

## De-serialization
An `Expression` object can be converted back to string form via the `String()`
method.  For non-normalized expressions the resulting string should match
closely the original expression string (omitting any superfluous whitespace).
With normalized expressions, the resulting string is unlikely to be anything
like the original input expression unless the original expression was already
in normal form (this is not a bug).

# Dependencies

The program is written in Go, version 1.18. It may compile with older compiler
versions.  The program does not have any third party dependencies.

# License

Copyright 2022 Markus Holmström (MawKKe)

The works under this repository are licenced under Apache License 2.0.
See file `LICENSE` for more information.

# Contributing

This project is hosted at https://github.com/MawKKe/integer-interval-expressions-go

You are welcome to leave bug reports, fixes and feature requests. Thanks!

