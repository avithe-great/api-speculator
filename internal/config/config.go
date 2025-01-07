// SPDX-License-Identifier: Apache-2.0
// Copyright 2024 Authors of API-Speculator

package config

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const defaultConfigFilePath = "config/default.yaml"
const defaultJSONReportFilePath = "findings.json"

type Database struct {
	Uri        string `json:"uri"`
	User       string `json:"user"`
	Password   string `json:"password"`
	Name       string `json:"name"`
	Collection string `json:"collection"`
}

type Environment struct {
	ClusterId int `json:"clusterId,omitempty"`
	TenantId  int `json:"tenantId,omitempty"`
}

type Exporter struct {
	JsonReportFilePath string `json:"jsonReportFilePath,omitempty"`
}

type Configuration struct {
	Database    Database    `json:"database"`
	Environment Environment `json:"environment"`
	OpenAPISpec string      `json:"openAPISpec"`
	Exporter    Exporter    `json:"exporter,omitempty"`
}

func (c *Configuration) validate() error {
	if c.Database.Uri == "" {
		return fmt.Errorf("configuration does not contain a valid database URI")
	}
	if c.Database.User == "" {
		return fmt.Errorf("configuration does not contain a valid database user")
	}
	if c.Database.Password == "" {
		return fmt.Errorf("configuration does not contain a valid database password")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("configuration does not contain a valid database name")
	}
	if c.Database.Collection == "" {
		return fmt.Errorf("configuration does not contain a valid database collection name")
	}

	if c.OpenAPISpec == "" {
		return fmt.Errorf("configuration does not contain a valid OpenAPI Specification filepath or URL")
	}

	if c.Environment.ClusterId == 0 {
		return fmt.Errorf("please provide a valid cluster ID")
	}

	if c.Exporter.JsonReportFilePath == "" {
		return fmt.Errorf("configuration does not contain a valid JSON reports file path")
	}

	return nil
}

func New(configFilePath string, logger *zap.SugaredLogger) (Configuration, error) {
	if configFilePath == "" {
		configFilePath = defaultConfigFilePath
		logger.Warn("using default config file path: ", configFilePath)
	}

	viper.SetConfigFile(configFilePath)
	if err := viper.ReadInConfig(); err != nil {
		return Configuration{}, fmt.Errorf("failed to read config file: %w", err)
	}

	config := Configuration{}
	if err := viper.Unmarshal(&config); err != nil {
		return Configuration{}, fmt.Errorf("failed to unmarshal config file: %w", err)
	}
	if config.Exporter.JsonReportFilePath == "" {
		config.Exporter.JsonReportFilePath = defaultJSONReportFilePath
		logger.Warn("using default JSON report file path: ", defaultJSONReportFilePath)
	}

	if err := config.validate(); err != nil {
		return Configuration{}, err
	}

	dbUser := config.Database.User
	dbPassword := config.Database.Password

	config.Database.User = ""
	config.Database.Password = ""

	if logger.Level().String() == "debug" {
		bytes, err := json.Marshal(config)
		if err != nil {
			logger.Errorf("failed to marshal config file: %v", err)
		}
		logger.Debugf("configuration: %s", string(bytes))
	}

	config.Database.User = dbUser
	config.Database.Password = dbPassword

	return config, nil
}
