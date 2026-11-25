// Copyright 2026 High Five Labs
// SPDX-License-Identifier: Apache-2.0

package extensions

import (
	"net/http"

	"github.com/highfivelabs/krm-rest-runtime/rest"
)

// Vendor is how the manager reaches GitHub.
var Vendor = &rest.Vendor{
	Name:           "github",
	DefaultBaseURL: "https://api.github.com",
	CredentialKeys: []string{"token"},
	Authorize:      func(r *http.Request, c map[string]string) { r.Header.Set("Authorization", "Bearer "+c["token"]) },
	Headers:        map[string]string{"Accept": "application/vnd.github+json", "X-GitHub-Api-Version": "2022-11-28"},
	// The GraphQL kinds' endpoint is /graphql on api.github.com, and
	// /api/graphql on GitHub Enterprise Server, whose REST API is /api/v3.
	RESTPath: "/v3",
}
