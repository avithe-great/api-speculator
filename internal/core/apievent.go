package core

import (
	"fmt"

	"github.com/emirpasic/gods/sets/hashset"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/5gsec/api-speculator/internal/apievent"
)

func (m *Manager) findDocuments(collectionName string, clusterId int) (*hashset.Set, error) {
	filter := bson.D{{Key: "operation", Value: "Api"}}
	if clusterId != 0 {
		filter = append(filter, bson.E{Key: "cluster_id", Value: clusterId})
	}

	projection := bson.D{
		{Key: "_id", Value: 0},
		{Key: "cluster_name", Value: 1},
		{Key: "api_event.http.request.headers.:authority", Value: 1},
		{Key: "api_event.http.request.headers.host", Value: 1},
		{Key: "api_event.http.request.method", Value: 1},
		{Key: "api_event.http.request.path", Value: 1},
		{Key: "api_event.http.response.status_code", Value: 1},
		{Key: "api_event.count", Value: 1},
	}

	cursor, err := m.DBHandler.Database.Collection(collectionName).Find(
		m.Ctx,
		filter,
		&options.FindOptions{Projection: &projection},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find documents: %w", err)
	}
	defer func() {
		if cerr := cursor.Close(m.Ctx); cerr != nil {
			m.Logger.Errorf("failed to close cursor: %v", cerr)
		}
	}()

	apiEvents := hashset.New()

	for cursor.Next(m.Ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			m.Logger.Warnf("failed to decode document: %v", err)
			continue
		}

		clusterName, _ := doc["cluster_name"].(string)
		apiEvent, _ := doc["api_event"].(bson.M)
		if apiEvent == nil {
			continue
		}

		httpEvent, _ := apiEvent["http"].(bson.M)
		if httpEvent == nil {
			continue
		}

		req, _ := httpEvent["request"].(bson.M)
		resp, _ := httpEvent["response"].(bson.M)

		var responseCode int
		if rc, ok := resp["status_code"].(int32); ok {
			responseCode = int(rc)
		} else if rc, ok := resp["status_code"].(int64); ok {
			responseCode = int(rc)
		}

		if responseCode == 0 {
			continue
		}

		headers, _ := req["headers"].(bson.M)
		serviceName := ""
		if v, ok := headers[":authority"].(string); ok {
			serviceName = v
		} else if v, ok := headers["host"].(string); ok {
			serviceName = v
		}

		requestMethod, _ := req["method"].(string)
		requestPath, _ := req["path"].(string)

		var occurrences int
		if cnt, ok := apiEvent["count"].(int32); ok {
			occurrences = int(cnt)
		} else if cnt, ok := apiEvent["count"].(int64); ok {
			occurrences = int(cnt)
		}

		apiEvents.Add(apievent.ApiEvent{
			ClusterName:   clusterName,
			ServiceName:   serviceName,
			RequestMethod: requestMethod,
			RequestPath:   requestPath,
			ResponseCode:  responseCode,
			Occurrences:   occurrences,
		})
	}

	if err := cursor.Err(); err != nil {
		m.Logger.Errorf("cursor iteration error: %v", err)
	}

	if apiEvents.Size() == 0 {
		clusterInfo := fmt.Sprintf("clusterID: `%d`", clusterId)
		if clusterId == 0 {
			clusterInfo = "all clusters"
		}
		m.Logger.Warnf("no API event documents found in `%s` for %s", collectionName, clusterInfo)
		return nil, nil
	}

	return apiEvents, nil
}
