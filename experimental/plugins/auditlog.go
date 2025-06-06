// Copyright 2023 Juan Pablo Tosso and the OWASP Coraza contributors
// SPDX-License-Identifier: Apache-2.0

package plugins

import (
	"github.com/ad3n/seclang/experimental/plugins/plugintypes"
	"github.com/ad3n/seclang/internal/auditlog"
)

// RegisterAuditLogWriter registers a new audit log writer.
func RegisterAuditLogWriter(name string, writerFactory func() plugintypes.AuditLogWriter) {
	auditlog.RegisterWriter(name, writerFactory)
}

// RegisterAuditLogFormatter registers a new audit log formatter.
func RegisterAuditLogFormatter(name string, format plugintypes.AuditLogFormatter) {
	auditlog.RegisterFormatter(name, format)
}
