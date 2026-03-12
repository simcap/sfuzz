package spec

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/google/uuid"
)

func GenerateExample(schema *openapi3.Schema) any {
	if schema == nil {
		return ""
	}
	if schema.Example != nil {
		return schema.Example
	}
	if schema.Format == "uuid" {
		return uuid.NewString()
	}
	if schema.Type.Is(openapi3.TypeNumber) || schema.Type.Is(openapi3.TypeInteger) {
		return 12345
	}

	return ""
}
