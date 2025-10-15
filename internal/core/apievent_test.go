package core

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestGetNestedAndGetNestedString(t *testing.T) {
	doc := bson.M{
		"level1": bson.M{
			"level2": bson.M{
				"str": "value",
				"num": int32(42),
			},
			"plain": "top",
		},
	}

	// existing nested
	if v, ok := getNested(doc, "level1", "level2", "str"); !ok || v != "value" {
		t.Fatalf("expected nested string 'value', got %#v (ok=%v)", v, ok)
	}

	// getNestedString existing
	if s, ok := getNestedString(doc, "level1", "plain"); !ok || s != "top" {
		t.Fatalf("expected nested string 'top', got %#v (ok=%v)", s, ok)
	}

	// non-existing path
	if _, ok := getNested(doc, "level1", "missing"); ok {
		t.Fatalf("expected missing path to be false")
	}

	// wrong type for getNestedString
	if s, ok := getNestedString(doc, "level1", "level2", "num"); ok || s != "" {
		t.Fatalf("expected non-string to return false and empty string, got s=%q ok=%v", s, ok)
	}
}

func TestToIntVariousTypes(t *testing.T) {
	cases := []struct {
		name   string
		input  interface{}
		want   int
		wantOk bool
	}{
		{"int", int(7), 7, true},
		{"int32", int32(8), 8, true},
		{"int64", int64(9), 9, true},
		{"float64", float64(10.0), 10, true},
		{"numericString", "123", 123, true},
		{"badString", "abc", 0, false},
		{"nil", nil, 0, false},
		{"unsupported", bson.M{}, 0, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := toInt(tc.input)
			if ok != tc.wantOk || got != tc.want {
				t.Fatalf("toInt(%#v) = (%d, %v), want (%d, %v)", tc.input, got, ok, tc.want, tc.wantOk)
			}
		})
	}
}

func TestValidateStringOperator(t *testing.T) {
	okCases := []StringOperators{
		{Eq: []string{"a"}},
		{Neq: []string{"a"}},
		{RegEx: []string{"^foo"}},
		{},
	}
	for i, c := range okCases {
		if !validateStringOperator(c) {
			t.Fatalf("expected case %d to be valid: %#v", i, c)
		}
	}

	badCases := []StringOperators{
		{Eq: []string{"a"}, Neq: []string{"b"}},
		{Eq: []string{"a"}, RegEx: []string{"x"}},
		{Neq: []string{"a"}, RegEx: []string{"x"}},
		{Eq: []string{"a"}, Neq: []string{"b"}, RegEx: []string{"x"}},
	}
	for i, c := range badCases {
		if validateStringOperator(c) {
			t.Fatalf("expected case %d to be invalid: %#v", i, c)
		}
	}
}

func TestBuildMongoFilterCriteria_SingleHostnameEq(t *testing.T) {
	// FilterCriteria to match hostname eq ["host1"]
	fc := []FilterCriteria{
		{
			Operator: "",
			Condition: Condition{
				Field: "hostname",
				Value: StringOperators{Eq: []string{"host1"}},
			},
		},
	}

	expr, err := buildMongoFilterCriteria(fc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expect expression like: bson.D{{"api_event.http.request.hostname", bson.D{{"$in", []string{"host1"}}}}}
	if len(expr) != 1 {
		t.Fatalf("expected single expression element, got %#v", expr)
	}

	elem := expr[0]
	if elem.Key != "api_event.http.request.hostname" {
		t.Fatalf("expected key 'api_event.http.request.hostname', got %q", elem.Key)
	}

	// value is a bson.D with one element: $in -> []string{"host1"}
	valD, ok := elem.Value.(bson.D)
	if !ok {
		t.Fatalf("expected element value to be bson.D, got %#v", elem.Value)
	}
	if len(valD) != 1 || valD[0].Key != "$in" {
		t.Fatalf("expected $in operator, got %#v", valD)
	}

	// Value of $in should be []string or []interface{}
	switch v := valD[0].Value.(type) {
	case []string:
		if !reflect.DeepEqual(v, []string{"host1"}) {
			t.Fatalf("$in value mismatch, got %#v", v)
		}
	case []interface{}:
		if len(v) != 1 || v[0] != "host1" {
			t.Fatalf("$in value mismatch, got %#v", v)
		}
	default:
		t.Fatalf("unexpected $in value type: %T %#v", v, v)
	}
}
