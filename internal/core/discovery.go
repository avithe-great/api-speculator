// SPDX-License-Identifier: Apache-2.0
// Copyright 2024 Authors of API-Speculator

package core

import (
	"slices"
	"strings"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/v3"

	"github.com/5gsec/api-speculator/internal/apievent"
	"github.com/5gsec/api-speculator/internal/apispec"
	"github.com/5gsec/api-speculator/internal/pathtrie"
)

func (m *Manager) findShadowAndZombieApi(trie pathtrie.PathTrie, events []apievent.ApiEvent, model *libopenapi.DocumentModel[v3.Document]) ([]string, []string) {
	var shadowApis []string
	var zombieApis []string

	for _, event := range events {
		requestPath, _ := apispec.GetPathAndQuery(event.RequestPath)

		// Skip static assets and root endpoint
		if requestPath == "/" ||
			strings.HasPrefix(requestPath, "/assets") ||
			strings.HasPrefix(requestPath, "/site") ||
			strings.HasPrefix(requestPath, "/sites") ||
			strings.HasPrefix(requestPath, "env") ||
			strings.HasSuffix(requestPath, "env") ||
			strings.HasSuffix(requestPath, "png") ||
			strings.HasSuffix(requestPath, "svg") ||
			strings.HasSuffix(requestPath, "gif") ||
			strings.HasSuffix(requestPath, "js") {
			continue
		}

		_, _, found := trie.GetPathAndValue(requestPath)
		if !found {
			if !slices.Contains(shadowApis, requestPath) {
				shadowApis = append(shadowApis, requestPath)
			}
		}

		currPathValue, found := model.Model.Paths.PathItems.Get(requestPath)
		if found {
			for operation := currPathValue.GetOperations().First(); operation != nil; operation = operation.Next() {
				if operation.Value().Deprecated != nil && *operation.Value().Deprecated {
					if !slices.Contains(zombieApis, requestPath) {
						zombieApis = append(zombieApis, requestPath)
					}
				}
			}
		}
	}

	return shadowApis, zombieApis
}
