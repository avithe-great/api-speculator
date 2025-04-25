package core

import (
	"fmt"

	"github.com/emirpasic/gods/sets/hashset"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/5gsec/api-speculator/internal/apievent"
)

func (m *Manager) findDocuments(collectionName string, clusterId int) (*hashset.Set, error) {
	// Todo: Process in batch of maybe 2000
	filter := bson.D{
		bson.E{Key: "operation", Value: "Api"},
		bson.E{Key: "cluster_id", Value: clusterId},
	}
	projection := bson.D{
		{Key: "_id", Value: 0},
		{Key: "cluster_name", Value: 1},
		{Key: "api_event.http.request.headers.:authority", Value: 1},
		{Key: "api_event.http.request.method", Value: 1},
		{Key: "api_event.http.request.path", Value: 1},
		{Key: "api_event.http.response.status_code", Value: 1},
		{Key: "api_event.count", Value: 1},
	}

	cursor, err := m.DBHandler.Database.
		Collection(collectionName).
		Find(m.Ctx, &filter, &options.FindOptions{
			Projection: &projection,
		})
	if err != nil {
		return nil, fmt.Errorf("failed to find documents: %w", err)
	}
	defer func() {
		if err := cursor.Close(m.Ctx); err != nil {
			m.Logger.Errorf("failed to close cursor: %v", err)
		}
	}()

	apiEvents := hashset.New()
	for cursor.Next(m.Ctx) {
		var document bson.M
		if err := cursor.Decode(&document); err != nil {
			m.Logger.Error(err)
			continue
		}

		responseCode := document["api_event"].(bson.M)["http"].(bson.M)["response"].(bson.M)["status_code"]
		if responseCode == nil {
			continue
		}
		apiEvents.Add(apievent.ApiEvent{
			ClusterName:   document["cluster_name"].(string),
			ServiceName:   document["api_event"].(bson.M)["http"].(bson.M)["request"].(bson.M)["headers"].(bson.M)[":authority"].(string),
			RequestMethod: document["api_event"].(bson.M)["http"].(bson.M)["request"].(bson.M)["method"].(string),
			RequestPath:   document["api_event"].(bson.M)["http"].(bson.M)["request"].(bson.M)["path"].(string),
			ResponseCode:  int(responseCode.(int64)),
			Occurrences:   int(document["api_event"].(bson.M)["count"].(int32)),
		})
	}
	if apiEvents.Size() == 0 {
		m.Logger.Warnf("no documents found in `%s` collection of clusterID: `%d`", collectionName, clusterId)
		return nil, nil
	}

	return apiEvents, nil
}
