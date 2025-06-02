// SPDX-License-Identifier: Apache-2.0
// Copyright 2024 Authors of API-Speculator

package core

import (
	"encoding/json"
	"os"

	"github.com/pb33f/libopenapi/datamodel/high/base"
)

type API struct {
	ClusterName   string `json:"clusterName,omitempty"`
	ServiceName   string `json:"serviceName,omitempty"`
	RequestMethod string `json:"requestMethod"`
	RequestPath   string `json:"requestPath"`
	Occurrences   int    `json:"occurrences,omitempty"`
}

type apiReport struct {
	ClusterId   int    `json:"clusterId"`
	TenantId    int    `json:"tenantId"`
	SpecTitle   string `json:"specTitle"`
	SpecVersion string `json:"specVersion"`
	OASVersion  string `json:"oasVersion"`
	ShadowAPIs  []API  `json:"shadowApis,omitempty"`
	ZombieAPIs  []API  `json:"zombieApis,omitempty"`
	OrphanAPIs  []API  `json:"orphanApis,omitempty"`
}

func (m *Manager) exportJsonReport(reportFilePath string, shadowApis, zombieApis, orphanApis []API, specInfo *base.Info, openApiVersion string) error {
	report := apiReport{
		ClusterId:   m.Cfg.Environment.ClusterId,
		TenantId:    m.Cfg.Environment.TenantId,
		SpecTitle:   specInfo.Title,
		SpecVersion: specInfo.Version,
		OASVersion:  openApiVersion,
		ShadowAPIs:  shadowApis,
		ZombieAPIs:  zombieApis,
		OrphanAPIs:  orphanApis,
	}

	f, err := os.OpenFile(reportFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o666)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		if err := f.Close(); err != nil {
			m.Logger.Error("failed to close file", err)
		}
	}(f)

	bytesToWrite, err := json.MarshalIndent(report, "", " ")
	if err != nil {
		return err
	}
	if _, err = f.Write(bytesToWrite); err != nil {
		return err
	}

	return nil
}
