// SPDX-License-Identifier: Apache-2.0
// Copyright 2024 Authors of API-Speculator

package apispec

import (
	"net/url"
	"strings"
)

func ExtractQueryAndParams(path string) (string, url.Values) {
	_, query := GetPathAndQuery(path)
	if query == "" {
		return "", nil
	}

	values, err := url.ParseQuery(query)
	if err != nil {
		return "", nil
	}
	return query, values
}

// GetPathAndQuery splits a URL into path and query components.
func GetPathAndQuery(fullPath string) (string, string) {
	if idx := strings.IndexByte(fullPath, '?'); idx != -1 {
		if idx == len(fullPath)-1 {
			return fullPath[:idx], ""
		}
		return fullPath[:idx], fullPath[idx+1:]
	}
	return fullPath, ""
}
