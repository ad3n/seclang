// Copyright 2022 Juan Pablo Tosso and the OWASP Coraza contributors
// SPDX-License-Identifier: Apache-2.0

//go:build !coraza.disabled_operators.rx

package operators

import (
	"fmt"
	"regexp"
	"strconv"
	"unicode/utf8"

	"rsc.io/binaryregexp"

	"github.com/ad3n/seclang/experimental/plugins/plugintypes"
	"github.com/ad3n/seclang/internal/memoize"
)

type rx struct {
	re *regexp.Regexp
}

var _ plugintypes.Operator = (*rx)(nil)

func newRX(options plugintypes.OperatorOptions) (plugintypes.Operator, error) {
	var data string
	if shouldNotUseMultilineRegexesOperatorByDefault {
		// (?s) enables dotall mode, required by some CRS rules and matching ModSec behavior, see
		// - https://github.com/google/re2/wiki/Syntax
		// - Flag usage: https://groups.google.com/g/golang-nuts/c/jiVdamGFU9E
		data = fmt.Sprintf("(?s)%s", options.Arguments)
	} else {
		// TODO: deprecate multiline modifier set by default in Coraza v4
		// CRS rules will explicitly set the multiline modifier when needed
		// Having it enabled by default can lead to false positives and less performance
		// See https://github.com/corazawaf/coraza/pull/876
		data = fmt.Sprintf("(?sm)%s", options.Arguments)
	}

	if matchesArbitraryBytes(data) {
		// Use binary regex matcher if expression matches non-utf8 bytes. The binary matcher does
		// not match unicode, meaning we cannot support expressions with both unicode and non-utf8
		// matches. This should not be commonly needed.
		return newBinaryRX(options)
	}

	re, err := memoize.Do(data, func() (interface{}, error) { return regexp.Compile(data) })
	if err != nil {
		return nil, err
	}
	return &rx{re: re.(*regexp.Regexp)}, nil
}

func (o *rx) Evaluate(tx plugintypes.TransactionState, value string) bool {
	if tx.Capturing() {
		match := o.re.FindStringSubmatch(value)
		if len(match) == 0 {
			return false
		}
		for i, c := range match {
			if i == 9 {
				return true
			}
			tx.CaptureField(i, c)
		}
		return true
	} else {
		return o.re.MatchString(value)
	}
}

// binaryRx is exactly the same as rx, but using the binaryregexp package for matching
// arbitrary bytes.
type binaryRX struct {
	re *binaryregexp.Regexp
}

var _ plugintypes.Operator = (*binaryRX)(nil)

func newBinaryRX(options plugintypes.OperatorOptions) (plugintypes.Operator, error) {
	data := options.Arguments

	re, err := memoize.Do(data, func() (interface{}, error) { return binaryregexp.Compile(data) })
	if err != nil {
		return nil, err
	}
	return &binaryRX{re: re.(*binaryregexp.Regexp)}, nil
}

func (o *binaryRX) Evaluate(tx plugintypes.TransactionState, value string) bool {
	if tx.Capturing() {
		match := o.re.FindStringSubmatch(value)
		if len(match) == 0 {
			return false
		}
		for i, c := range match {
			if i == 9 {
				return true
			}
			tx.CaptureField(i, c)
		}
		return true
	} else {
		return o.re.MatchString(value)
	}
}

func init() {
	Register("rx", newRX)
}

// matchesArbitraryBytes checks for control sequences for byte matches in the expression.
// If the sequences are not valid utf8, it returns true.
func matchesArbitraryBytes(expr string) bool {
	decoded := make([]byte, 0, len(expr))
	for i := 0; i < len(expr); i++ {
		c := expr[i]
		if c != '\\' {
			decoded = append(decoded, c)
			continue
		}
		if i+3 >= len(expr) {
			decoded = append(decoded, expr[i:]...)
			break
		}
		if expr[i+1] != 'x' {
			decoded = append(decoded, expr[i])
			continue
		}

		v, mb, _, err := strconv.UnquoteChar(expr[i:], 0)
		if err != nil || mb {
			// Wasn't a byte escape sequence, shouldn't happen in practice.
			decoded = append(decoded, expr[i])
			continue
		}

		decoded = append(decoded, byte(v))
		i += 3
	}

	return !utf8.Valid(decoded)
}
