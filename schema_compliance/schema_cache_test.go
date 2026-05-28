package schema_compliance

import "testing"

func TestCompileSchemaCachesCompiledSchemas(t *testing.T) {
	const schemaJSON = `{"type":"object"}`

	first, err := compileSchema(schemaJSON)
	if err != nil {
		t.Fatalf("first compileSchema returned error: %v", err)
	}

	second, err := compileSchema(schemaJSON)
	if err != nil {
		t.Fatalf("second compileSchema returned error: %v", err)
	}

	if first != second {
		t.Fatal("compileSchema did not return cached schema for identical schema JSON")
	}
}
