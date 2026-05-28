package schema_compliance

import (
	"fmt"
	"strings"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

const (
	schemaCacheSize = 128
	schemaResource  = "schema.json"
)

var compiledSchemaCache = mustNewSchemaCache()

func mustNewSchemaCache() *lru.Cache[string, *jsonschema.Schema] {
	cache, err := lru.New[string, *jsonschema.Schema](schemaCacheSize)
	if err != nil {
		panic(fmt.Sprintf("create schema cache: %v", err))
	}
	return cache
}

func compileSchema(schemaJSON string) (*jsonschema.Schema, error) {
	if schema, ok := compiledSchemaCache.Get(schemaJSON); ok {
		return schema, nil
	}

	schema, err := compileSchemaUncached(schemaJSON)
	if err != nil {
		return nil, err
	}

	compiledSchemaCache.Add(schemaJSON, schema)
	return schema, nil
}

func compileSchemaUncached(schemaJSON string) (*jsonschema.Schema, error) {
	doc, err := jsonschema.UnmarshalJSON(strings.NewReader(schemaJSON))
	if err != nil {
		return nil, err
	}

	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource(schemaResource, doc); err != nil {
		return nil, err
	}
	return compiler.Compile(schemaResource)
}
