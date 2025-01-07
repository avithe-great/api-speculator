// SPDX-License-Identifier: Apache-2.0
// Copyright 2024 Authors of API-Speculator

package util

import (
	"reflect"
)

func IsNil(a any) bool {
	return a == nil || (reflect.ValueOf(a).Kind() == reflect.Ptr && reflect.ValueOf(a).IsNil())
}
