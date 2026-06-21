// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package openapi

import _ "embed"

// embeddedSpec is the OpenAPI specification compiled into the binary at build
// time. It is the default source for auto-generated `service` commands, so the
// CLI works out of the box with no external spec file required.
//
// To refresh it, copy the latest spec over internal/openapi/spec/openapi.json
// and rebuild. Users can override the embedded copy at runtime via the
// LINGTONG_OPENAPI environment variable or ~/.lingtong-cli/openapi.json.
//
//go:embed spec/openapi.json
var embeddedSpec []byte

// HasEmbeddedSpec reports whether a spec was compiled into the binary.
func HasEmbeddedSpec() bool {
	return len(embeddedSpec) > 0
}

// EmbeddedSpec parses and returns the spec embedded at build time.
func EmbeddedSpec() (*Spec, error) {
	return ParseBytes(embeddedSpec)
}
