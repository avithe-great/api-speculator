// SPDX-License-Identifier: Apache-2.0
// Copyright 2024 Authors of API-Speculator

package apispec

import (
	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

func BuildOASV3Model(specBytes []byte) (*libopenapi.DocumentModel[v3.Document], error) {
	document, err := libopenapi.NewDocument(specBytes)
	if err != nil {
		return nil, err
	}

	docModel, errors := document.BuildV3Model()
	if len(errors) > 0 {
		return nil, errors[0]
	}
	return docModel, nil
}
