// SPDX-License-Identifier: Apache-2.0
// Copyright 2024 Authors of API-Speculator

package core

import (
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"

	"github.com/5gsec/api-speculator/internal/apispec"
	"github.com/5gsec/api-speculator/internal/pathtrie"
)

func (m *Manager) buildModel(oasCfg string) (*libopenapi.DocumentModel[v3.Document], error) {
	var err error
	var specBytes []byte

	if strings.HasPrefix(oasCfg, "http://") || strings.HasPrefix(oasCfg, "https://") {
		specBytes, err = downloadSpec(oasCfg)
		if err != nil {
			return nil, err
		}
	} else {
		specBytes, err = os.ReadFile(oasCfg)
		if err != nil {
			return nil, err
		}
	}

	model, err := apispec.BuildOASV3Model(specBytes)
	if err != nil {
		m.Logger.Error(err)
		return nil, nil
	}
	return model, nil

}

func downloadSpec(oasCfg string) ([]byte, error) {
	response, err := http.Get(oasCfg)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (m *Manager) buildTrie(model *libopenapi.DocumentModel[v3.Document]) pathtrie.PathTrie {
	trie := pathtrie.New()
	for paths := model.Model.Paths.PathItems.First(); paths != nil; paths = paths.Next() {
		_ = trie.Insert(paths.Key(), paths.Value())
	}
	return trie
}
