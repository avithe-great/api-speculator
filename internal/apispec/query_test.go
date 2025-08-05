// SPDX-License-Identifier: Apache-2.0
// Copyright 2024 Authors of API-Speculator

package apispec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPathAndQuery(t *testing.T) {
	tests := []struct {
		input     string
		wantPath  string
		wantQuery string
	}{
		{"/api/v1/users?id=123", "/api/v1/users", "id=123"},
		{"/api/v1/users?", "/api/v1/users", ""},
		{"/api/v1/users", "/api/v1/users", ""},
	}

	for _, tt := range tests {
		path, query := GetPathAndQuery(tt.input)
		assert.Equal(t, tt.wantPath, path)
		assert.Equal(t, tt.wantQuery, query)
	}
}
