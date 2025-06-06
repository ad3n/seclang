// Copyright 2022 Juan Pablo Tosso and the OWASP Coraza contributors
// SPDX-License-Identifier: Apache-2.0

package macro

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ad3n/seclang/experimental/plugins/plugintypes"
	"github.com/corazawaf/coraza/v3/collection"
	"github.com/corazawaf/coraza/v3/types/variables"
)

type Macro interface {
	Expand(tx plugintypes.TransactionState) string
	String() string
}

var errEmptyData = errors.New("empty data")

func NewMacro(data string) (Macro, error) {
	if len(data) == 0 {
		return nil, errEmptyData
	}

	macro := &macro{
		tokens: []macroToken{},
	}
	if err := macro.compile(data); err != nil {
		return nil, err
	}
	return macro, nil
}

type macroToken struct {
	text     string
	variable variables.RuleVariable
	key      string
}

// macro is used to create tokenized strings that can be
// "expanded" at high speed and concurrent-safe.
// A macro contains tokens for strings and expansions
// For example: some string %{tx.var} some string
// The previous example would create 3 tokens:
// - String token: some string
// - Variable token: Variable: TX, key: var
// - String token: some string
type macro struct {
	original string
	tokens   []macroToken
}

// Expand the pre-compiled macro expression into a string
func (m *macro) Expand(tx plugintypes.TransactionState) string {
	if len(m.tokens) == 1 {
		return expandToken(tx, m.tokens[0])
	}
	res := strings.Builder{}
	for _, token := range m.tokens {
		res.WriteString(expandToken(tx, token))
	}
	return res.String()
}

func expandToken(tx plugintypes.TransactionState, token macroToken) string {
	if token.variable == variables.Unknown {
		return token.text
	}
	switch col := tx.Collection(token.variable).(type) {
	case collection.Keyed:
		if c := col.Get(token.key); len(c) > 0 {
			return c[0]
		}
	case collection.Single:
		return col.Get()
	default:
		if c := col.FindAll(); len(c) > 0 {
			return c[0].Value()
		}
	}

	// If the variable is known (e.g. TX) but the key is not found, we return the original text
	tx.DebugLogger().Warn().Str("variable", token.variable.Name()).Str("key", token.key).Msg("key not found in collection, returning the original text")
	return token.text
}

// compile is used to parse the input and generate the corresponding token
// Example input: %{var.foo} and %{var.bar}
// expected result:
// [0] macroToken{text: "%{var.foo}", variable: &variables.Var, key: "foo"},
// [1] macroToken{text: " and ", variable: nil, key: ""}
// [2] macroToken{text: "%{var.bar}", variable: &variables.Var, key: "bar"}
func (m *macro) compile(input string) error {
	l := len(input)
	if l == 0 {
		return fmt.Errorf("empty macro")
	}

	m.original = input
	var currentToken strings.Builder
	isMacro := false

	for i := 0; i < l; i++ {
		c := input[i]

		if c == '%' && i+1 < l && input[i+1] == '{' {
			if currentToken.Len() > 0 {
				m.tokens = append(m.tokens, macroToken{
					text:     currentToken.String(),
					variable: variables.Unknown,
					key:      "",
				})
				currentToken.Reset()
			}
			isMacro = true
			i++ // Skip '{'
			continue
		}

		if isMacro {
			if c == '}' {
				isMacro = false
				if input[i-1] == '.' {
					return fmt.Errorf("empty variable name")
				}
				varName, key, _ := strings.Cut(currentToken.String(), ".")
				v, err := variables.Parse(varName)
				if err != nil {
					return fmt.Errorf("unknown variable %q", varName)
				}
				m.tokens = append(m.tokens, macroToken{
					text:     currentToken.String(),
					variable: v,
					key:      strings.ToLower(key),
				})
				currentToken.Reset()
				continue
			}

			if !isValidMacroChar(c) {
				return fmt.Errorf("malformed variable starting with %q", "%{"+currentToken.String())
			}

			currentToken.WriteByte(c)

			if i+1 == l {
				return errors.New("malformed variable: no closing braces")
			}
			continue
		}

		currentToken.WriteByte(c)
	}

	if currentToken.Len() > 0 {
		m.tokens = append(m.tokens, macroToken{
			text:     currentToken.String(),
			variable: variables.Unknown,
			key:      "",
		})
	}
	return nil
}

func isValidMacroChar(c byte) bool {
	return c == '[' || c == ']' || c == '.' || c == '_' || c == '-' || (c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}

// String returns the original string
func (m *macro) String() string {
	return m.original
}

// IsExpandable return true if there are macro expanadable tokens
// TODO(jcchavezs): this is used only in a commented out section
func (m *macro) IsExpandable() bool {
	return len(m.tokens) > 1
}
