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

func GetPathAndQuery(fullPath string) (path, query string) {
	// Example: "/example-path?param=value" returns "/example-path", "param=value"
	index := strings.IndexByte(fullPath, '?')
	if index == -1 {
		return fullPath, ""
	}

	// Example: "/path?" returns "/path?", ""
	if index == (len(fullPath) - 1) {
		return fullPath, ""
	}

	path = fullPath[:index]
	query = fullPath[index+1:]
	return
}
