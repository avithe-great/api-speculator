// SPDX-License-Identifier: Apache-2.0
// Copyright 2024 Authors of API-Speculator

package apievent

type ApiEvent struct {
	RequestMethod string `json:"request_method"`
	RequestPath   string `json:"request_path"`
	ResponseCode  int    `json:"response_code"`
}
