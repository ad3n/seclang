// Copyright 2023 Juan Pablo Tosso and the OWASP Coraza contributors
// SPDX-License-Identifier: Apache-2.0

package plugins_test

import (
	"testing"

	"github.com/ad3n/seclang/experimental/plugins"
	"github.com/ad3n/seclang/experimental/plugins/plugintypes"
	"github.com/ad3n/seclang/internal/operators"
)

func TestGetOperator(t *testing.T) {
	t.Run("get existing operator", func(t *testing.T) {
		operator := func(options plugintypes.OperatorOptions) (plugintypes.Operator, error) {
			return nil, nil
		}

		plugins.RegisterOperator("custom_operator", operator)
		_, err := operators.Get("custom_operator", plugintypes.OperatorOptions{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}
