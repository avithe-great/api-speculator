// SPDX-License-Identifier: Apache-2.0
// Copyright 2024 Authors of API-Speculator

package database

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/5gsec/api-speculator/internal/config"
	"github.com/5gsec/api-speculator/internal/util"
)

type Handler struct {
	Database   *mongo.Database
	Disconnect func() error
}

func New(ctx context.Context, dbConfig config.Database) (*Handler, error) {
	logger := util.GetLogger()

	opts := options.Client().
		ApplyURI(dbConfig.Uri).
		SetAuth(
			options.Credential{
				Username: dbConfig.User,
				Password: dbConfig.Password,
			},
		)
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	logger.Infof("connecting to %s database", dbConfig.Name)
	if err := client.Ping(ctx, readpref.PrimaryPreferred()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	logger.Info("connected to database")

	return &Handler{
		Database: client.Database(dbConfig.Name),
		Disconnect: func() error {
			return client.Disconnect(ctx)
		},
	}, nil
}
