// SPDX-License-Identifier: Apache-2.0
// Copyright 2024 Authors of API-Speculator

package core

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"

	"github.com/5gsec/api-speculator/internal/apievent"
	"github.com/5gsec/api-speculator/internal/config"
	"github.com/5gsec/api-speculator/internal/database"
	"github.com/5gsec/api-speculator/internal/util"
)

type Manager struct {
	Ctx       context.Context
	Logger    *zap.SugaredLogger
	DBHandler *database.Handler
	Cfg       config.Configuration
}

func (m *Manager) findDocuments(collectionName string, clusterId int) ([]apievent.ApiEvent, error) {
	filter := bson.D{
		bson.E{
			Key:   "operation",
			Value: "Api",
		},
		bson.E{
			Key:   "cluster_id",
			Value: clusterId,
		},
	}
	projection := bson.D{
		{Key: "_id", Value: 0},
		{Key: "api_event.http.request.method", Value: 1},
		{Key: "api_event.http.request.path", Value: 1},
		// Fixme: Discrepancy bw status_code and statuscode
		//{Key: "api_event.http.response.status_code", Value: 1},
		{Key: "api_event.http.response.statuscode", Value: 1},
	}

	cursor, err := m.DBHandler.Database.
		Collection(collectionName).
		Find(m.Ctx, &filter, &options.FindOptions{
			Projection: &projection,
		})
	if err != nil {
		return nil, fmt.Errorf("failed to find documents: %w", err)
	}

	var apiEvents []apievent.ApiEvent
	for cursor.Next(m.Ctx) {
		var document bson.M
		if err := cursor.Decode(&document); err != nil {
			m.Logger.Error(err)
			continue
		}

		responseCode := document["api_event"].(bson.M)["http"].(bson.M)["response"].(bson.M)["statuscode"]
		if responseCode == nil {
			continue
		}
		apiEvents = append(apiEvents, apievent.ApiEvent{
			RequestMethod: document["api_event"].(bson.M)["http"].(bson.M)["request"].(bson.M)["method"].(string),
			RequestPath:   document["api_event"].(bson.M)["http"].(bson.M)["request"].(bson.M)["path"].(string),
			ResponseCode:  int(responseCode.(int64)),
		})
	}
	if len(apiEvents) == 0 {
		m.Logger.Warnf("no documents found in `%s` collection of clusterID: `%d`", collectionName, clusterId)
		return apiEvents, nil
	}

	if err := cursor.Close(m.Ctx); err != nil {
		m.Logger.Errorf("failed to close cursor: %v", err)
		return nil, err
	}

	return apiEvents, nil
}

func (m *Manager) close() {
	if m.DBHandler != nil {
		if err := m.DBHandler.Disconnect(); err != nil {
			m.Logger.Errorf("failed to disconnect to database: %v", err)
		}
	}
}

func Run(ctx context.Context, configFilePath string) {
	mgr := &Manager{
		Ctx:    ctx,
		Logger: util.GetLogger(),
	}
	defer mgr.close()

	mgr.Logger.Info("starting speculator")

	cfg, err := config.New(configFilePath, mgr.Logger)
	if err != nil {
		mgr.Logger.Error(err)
		return
	}
	mgr.Cfg = cfg

	dbHandler, err := database.New(mgr.Ctx, mgr.Cfg.Database)
	if err != nil {
		mgr.Logger.Error(err)
		return
	}
	mgr.DBHandler = dbHandler

	collectionName := mgr.Cfg.Database.Collection
	clusterId := mgr.Cfg.Environment.ClusterId

	events, err := mgr.findDocuments(collectionName, clusterId)
	if err != nil {
		mgr.Logger.Errorf("failed to find documents: %v", err)
		return
	}
	if len(events) == 0 {
		return
	}

	model, _ := mgr.buildModel(mgr.Cfg.OpenAPISpec)
	trie := mgr.buildTrie(model)

	shadowApis, zombieApis := mgr.findShadowAndZombieApi(trie, events, model)
	if err := mgr.exportJsonReport(mgr.Cfg.Exporter.JsonReportFilePath, shadowApis, zombieApis, model.Model.Info, model.Index.GetConfig().SpecInfo.Version); err != nil {
		mgr.Logger.Error(err)
		return
	}
	mgr.Logger.Infof("successfully generated `%s` JSON report", mgr.Cfg.Exporter.JsonReportFilePath)
}
