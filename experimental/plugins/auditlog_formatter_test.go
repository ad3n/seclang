// Copyright 2022 Juan Pablo Tosso and the OWASP Coraza contributors
// SPDX-License-Identifier: Apache-2.0

//go:build !tinygo
// +build !tinygo

// Not aimed to tinygo as serial writer is a noop writer

package plugins_test

import (
	"github.com/ad3n/seclang/experimental/plugins"
	"github.com/ad3n/seclang/experimental/plugins/plugintypes"
	"github.com/corazawaf/coraza/v3"
)

type testFormatter struct{}

func (testFormatter) Format(al plugintypes.AuditLog) ([]byte, error) {
	return []byte(al.Transaction().ID()), nil
}

func (testFormatter) MIME() string {
	return "sample"
}

// ExampleRegisterAuditLogFormatter shows how to register a custom audit log formatter
// and tests the output of the formatter.
func ExampleRegisterAuditLogFormatter() {

	plugins.RegisterAuditLogFormatter("txid", &testFormatter{})

	w, err := coraza.NewWAF(
		coraza.NewWAFConfig().
			WithDirectives(`
				SecAuditEngine On
				SecAuditLogParts ABCFHZ
				SecAuditLog /dev/stdout
				SecAuditLogFormat txid
				SecAuditLogType serial
			`),
	)
	if err != nil {
		panic(err)
	}

	tx := w.NewTransactionWithID("abc123")
	tx.ProcessLogging()
	tx.Close()

	// Output: abc123
}
