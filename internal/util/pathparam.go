// SPDX-License-Identifier: Apache-2.0
// Copyright 2024 Authors of API-Speculator

package util

import (
	"strings"
)

const (
	ParamPrefix = "{"
	ParamSuffix = "}"
)

func IsPathParam(segment string) bool {
	return strings.HasPrefix(segment, ParamPrefix) &&
		strings.HasSuffix(segment, ParamSuffix)
}
