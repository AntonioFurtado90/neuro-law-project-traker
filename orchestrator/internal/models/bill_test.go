package models

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"
)

// schemaPath locates contracts/schemas/bill.schema.json relative to this
// file, independent of the current working directory the test runs from.
func schemaPath(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine caller for schema path resolution")
	}
	// this file: orchestrator/internal/models/bill_test.go -> repo root is 3 levels up.
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..", "..")
	return filepath.Join(repoRoot, "contracts", "schemas", "bill.schema.json")
}

func jsonTagNames(t *testing.T) []string {
	t.Helper()
	var names []string
	typ := reflect.TypeOf(Bill{})
	for i := 0; i < typ.NumField(); i++ {
		tag := typ.Field(i).Tag.Get("json")
		name, _, _ := strings.Cut(tag, ",")
		if name == "" {
			t.Fatalf("field %s has no json tag", typ.Field(i).Name)
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func TestBillFields_MatchSchemaProperties(t *testing.T) {
	raw, err := os.ReadFile(schemaPath(t))
	if err != nil {
		t.Fatalf("reading bill.schema.json: %v", err)
	}

	var schema struct {
		Properties map[string]any `json:"properties"`
		Required   []string       `json:"required"`
	}
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatalf("parsing bill.schema.json: %v", err)
	}

	var schemaFields []string
	for name := range schema.Properties {
		schemaFields = append(schemaFields, name)
	}
	sort.Strings(schemaFields)

	goFields := jsonTagNames(t)

	if !reflect.DeepEqual(schemaFields, goFields) {
		t.Fatalf("Bill json tags %v do not match bill.schema.json properties %v", goFields, schemaFields)
	}

	for _, required := range schema.Required {
		found := false
		for _, name := range goFields {
			if name == required {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("schema requires %q but Bill has no matching json tag", required)
		}
	}
}
