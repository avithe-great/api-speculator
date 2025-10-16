package core

import (
	"context"
	"fmt"
	"strings"

	"github.com/5gsec/api-speculator/internal/apievent"
	"github.com/emirpasic/gods/sets/hashset"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// findApiOperationDocuments fetches API documents based on collectionName, optional clusterId,
// and optional collectionCriteria. Returns a set of unique apievent.
func (m *Manager) findApiOperationDocuments(eventCollectionName, apiCollectionName string, clusterId int, nameList []string) (*hashset.Set, error) {
	// base filter: only Api operation documents
	filter := bson.D{{Key: "operation", Value: "Api"}}

	if clusterId != 0 {
		filter = append(filter, bson.E{Key: "cluster_id", Value: clusterId})
	}
	// if apiCollectionName and nameList are provided, fetch criteria and build filter
	if apiCollectionName != "" && len(nameList) > 0 {
		criteriaMap, err := m.GetCriteriaByCollections(m.Ctx, apiCollectionName, nameList)
		if err != nil {
			m.Logger.Errorf("failed to get criteria by collections: %v", err)
			return nil, fmt.Errorf("failed to get criteria by collections: %w", err)
		}

		var allCriteria []FilterCriteria
		for _, criteria := range criteriaMap {
			allCriteria = append(allCriteria, criteria...)
		}

		if len(allCriteria) > 0 {
			criteriaFilter, err := buildMongoFilterCriteria(allCriteria)
			if err != nil {
				m.Logger.Errorf("failed to build mongo query for collection filter criteria: %v", err)
				return nil, fmt.Errorf("failed to build mongo query for collection filter criteria: %w", err)
			}
			filter = append(filter, bson.E{Key: "$and", Value: criteriaFilter})
		}
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

	findOpts := &options.FindOptions{
		Projection: &projection,
	}

	cursor, err := m.DBHandler.Database.Collection(eventCollectionName).Find(m.Ctx, filter, findOpts)
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
			m.Logger.Errorf("failed to decode document: %v", err)
			continue
		}

		rcVal, ok := getNested(doc, "api_event", "http", "response", "status_code")
		if !ok {
			continue
		}
		responseCode, ok := toInt(rcVal)
		if !ok {
			// can't interpret response code as int -> skip
			continue
		}

		clusterName, _ := getString(doc, "cluster_name")
		serviceName, _ := getNestedString(doc, "api_event", "http", "request", "headers", ":authority")
		requestMethod, _ := getNestedString(doc, "api_event", "http", "request", "method")
		requestPath, _ := getNestedString(doc, "api_event", "http", "request", "path")

		occVal, _ := getNested(doc, "api_event", "count")
		occurrences, _ := toInt(occVal)

		apiEvents.Add(apievent.ApiEvent{
			ClusterName:   clusterName,
			ServiceName:   serviceName,
			RequestMethod: requestMethod,
			RequestPath:   requestPath,
			ResponseCode:  responseCode,
			Occurrences:   occurrences,
		})
	}

	if apiEvents.Size() == 0 {
		clusterInfo := fmt.Sprintf("clusterID: `%d`", clusterId)
		if clusterId == 0 {
			clusterInfo = "all clusters"
		}
		m.Logger.Warnf("no documents found in `%s` collection for %s", eventCollectionName, clusterInfo)
		return apiEvents, nil
	}

	return apiEvents, nil
}

// helper: safely get a top-level string value from bson.M
func getString(m bson.M, key string) (string, bool) {
	if v, ok := m[key]; ok && v != nil {
		if s, ok := v.(string); ok {
			return s, true
		}
	}
	return "", false
}

// helper: navigate nested maps and return value if exists
func getNested(m bson.M, path ...string) (interface{}, bool) {
	var cur interface{} = m
	for _, p := range path {
		if curMap, ok := cur.(bson.M); ok {
			if v, exists := curMap[p]; exists {
				cur = v
				continue
			}
			// try map[string]interface{} as some decoders may produce that
			if altMap, ok := interface{}(curMap).(map[string]interface{}); ok {
				if v, exists := altMap[p]; exists {
					cur = v
					continue
				}
			}
			return nil, false
		}

		// If current isn't a map, cannot descend further
		return nil, false
	}
	return cur, true
}

func getNestedString(m bson.M, path ...string) (string, bool) {
	v, ok := getNested(m, path...)
	if !ok || v == nil {
		return "", false
	}
	if s, ok := v.(string); ok {
		return s, true
	}
	return "", false
}

// toInt attempts to convert various numeric BSON types to an int (int64/int32/float64/string)
func toInt(v interface{}) (int, bool) {
	if v == nil {
		return 0, false
	}
	switch val := v.(type) {
	case int:
		return val, true
	case int32:
		return int(val), true
	case int64:
		return int(val), true
	case float64:
		return int(val), true
	case string:
		var n int
		_, err := fmt.Sscanf(val, "%d", &n)
		if err == nil {
			return n, true
		}
		return 0, false
	default:
		return 0, false
	}
}

// FilterCriteria defines a condition and operator for Mongo query filtering.
type FilterCriteria struct {
	Operator  string    `bson:"operator,omitempty" json:"operator,omitempty"`
	Condition Condition `bson:"condition,omitempty" json:"condition,omitempty"`
}

// Condition represents a filter condition containing a field and its value.
type Condition struct {
	Field string          `bson:"field,omitempty" json:"field,omitempty"`
	Value StringOperators `bson:"value,omitempty" json:"value,omitempty"`
}

// StringOperators represents string-based comparison filters.
type StringOperators struct {
	Eq    []string `json:"eq,omitempty" bson:"eq,omitempty"`
	Neq   []string `json:"neq,omitempty" bson:"neq,omitempty"`
	RegEx []string `json:"regex,omitempty" bson:"regex,omitempty"`
}

// FieldMeta contains metadata describing a MongoDB field mapping.
type FieldMeta struct {
	DisplayName map[string]string
	BsonKey     string
	UnwindPath  string
	DefaultSort string
	FieldType   string // e.g., "string", "number", "boolean"
}

// fieldMetadata maps logical field names to BSON key paths and metadata.
func fieldMetadata(field string) (*FieldMeta, error) {
	fieldMap := map[string]*FieldMeta{
		"api_type":            {BsonKey: "api_event.metadata.api_type", FieldType: "string"},
		"auth_type":           {BsonKey: "api_event.metadata.is_authenticated", DisplayName: map[string]string{"true": "Authenticated", "false": "Non-authenticated"}, FieldType: "boolean"},
		"hostname":            {BsonKey: "api_event.http.request.hostname", FieldType: "string"},
		"method":              {BsonKey: "api_event.http.request.method", FieldType: "string"},
		"path":                {BsonKey: "api_event.http.request.path", FieldType: "string"},
		"response_code":       {BsonKey: "api_event.http.response.status_code", FieldType: "number"},
		"access_type":         {BsonKey: "api_event.metadata.access_type", FieldType: "string"},
		"count":               {BsonKey: "api_event.count"},
		"destination_ip":      {BsonKey: "destination", FieldType: "string"},
		"destination_name":    {BsonKey: "api_event.network.destination.metadata.name", FieldType: "string"},
		"destination_type":    {BsonKey: "api_event.network.destination.type", FieldType: "string"},
		"risk_score":          {BsonKey: "api_event.overall_risk_score", FieldType: "number"},
		"severity":            {BsonKey: "api_event.overall_severity", FieldType: "number"},
		"sensitive_data_type": {UnwindPath: "api_event.sensitive_data", BsonKey: "api_event.sensitive_data.name", FieldType: "string"},
	}

	if details, ok := fieldMap[field]; ok {
		return details, nil
	}
	return nil, fmt.Errorf("unknown group_by key: %s", field)
}

// buildMongoFilterCriteria converts a list of FilterCriteria into a MongoDB filter (bson.D).
func buildMongoFilterCriteria(filterCriteria []FilterCriteria) (bson.D, error) {
	if len(filterCriteria) == 0 {
		return bson.D{}, fmt.Errorf("filter criteria is empty")
	}

	expr := bson.D{}

	first := filterCriteria[0]
	fieldMeta, err := fieldMetadata(first.Condition.Field)
	if err != nil {
		return bson.D{}, fmt.Errorf("error getting bson key for field '%s': %v", first.Condition.Field, err)
	}
	if err := addToFilter(&expr, fieldMeta.BsonKey, first.Condition.Value); err != nil {
		return bson.D{}, fmt.Errorf("error creating bson object for field '%s': %v", fieldMeta.BsonKey, err)
	}

	for i := 1; i < len(filterCriteria); i++ {
		next := filterCriteria[i]
		nextExpr := bson.D{}
		fieldMeta, err := fieldMetadata(next.Condition.Field)
		if err != nil {
			return bson.D{}, fmt.Errorf("error getting bson key for field '%s': %v", next.Condition.Field, err)
		}
		if err := addToFilter(&nextExpr, fieldMeta.BsonKey, next.Condition.Value); err != nil {
			return bson.D{}, fmt.Errorf("error creating bson object for field '%s': %v", fieldMeta.BsonKey, err)
		}

		op := strings.ToUpper(next.Operator)
		switch op {
		case "OR":
			expr = bson.D{{Key: "$or", Value: bson.A{expr, nextExpr}}}
		case "AND":
			expr = bson.D{{Key: "$and", Value: bson.A{expr, nextExpr}}}
		default:
			return bson.D{}, fmt.Errorf("unknown operator '%s' in filter criteria", op)
		}
	}

	return expr, nil
}

// appendStringFilter adds string comparison filters to the provided bson.D.
func appendStringFilter(key string, op StringOperators, filters *bson.D) {
	if !validateStringOperator(op) {
		return
	}

	switch {
	case len(op.Eq) > 0:
		*filters = append(*filters, bson.E{Key: key, Value: bson.D{{Key: "$in", Value: op.Eq}}})
	case len(op.Neq) > 0:
		*filters = append(*filters, bson.E{Key: key, Value: bson.D{{Key: "$nin", Value: op.Neq}}})
	case len(op.RegEx) > 0:
		*filters = append(*filters, bson.E{Key: key, Value: bson.D{{Key: "$regex", Value: primitive.Regex{Pattern: op.RegEx[0], Options: "i"}}}})
	}
}

// validateStringOperator ensures only one type of string operator is applied.
func validateStringOperator(filter StringOperators) bool {
	if len(filter.Eq) != 0 && len(filter.Neq) != 0 && len(filter.RegEx) != 0 {
		return false
	}
	if len(filter.Eq) != 0 && (len(filter.Neq) != 0 || len(filter.RegEx) != 0) {
		return false
	}
	if len(filter.Eq) == 0 && (len(filter.Neq) != 0 && len(filter.RegEx) != 0) {
		return false
	}
	return true
}

// addToFilter appends a filter condition to the provided bson.D based on field type.
func addToFilter(filter *bson.D, fieldName string, fieldValue interface{}) error {
	switch v := fieldValue.(type) {
	case StringOperators:
		if !validateStringOperator(v) {
			return fmt.Errorf("invalid string operator for field '%s': %v", fieldName, v)
		}
		appendStringFilter(fieldName, v, filter)
	default:
		return fmt.Errorf("unsupported field value type '%T' for field '%s'", v, fieldName)
	}
	return nil
}

// GetCriteriaByCollections fetches the "criteria" field for the given api_collection names.
// It returns a map of collectionName → []FilterCriteria.
func (m *Manager) GetCriteriaByCollections(ctx context.Context, apiCollectionName string, names []string) (map[string][]FilterCriteria, error) {
	if len(names) == 0 {
		return nil, fmt.Errorf("no collection names provided")
	}

	filter := bson.M{
		"name": bson.M{"$in": names},
	}

	cursor, err := m.DBHandler.Database.Collection(apiCollectionName).Find(ctx, filter, &options.FindOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to find collections: %w", err)
	}
	defer cursor.Close(ctx)

	result := make(map[string][]FilterCriteria)
	for cursor.Next(ctx) {
		var doc struct {
			Name     string           `bson:"name"`
			Criteria []FilterCriteria `bson:"criteria"`
		}
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode collection document: %w", err)
		}

		result[doc.Name] = doc.Criteria
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
