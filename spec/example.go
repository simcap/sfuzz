package spec

import (
	"fmt"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/google/uuid"
)

func GenerateExample(param Param) any {
	schema, value := param.Schema, param.Value
	if schema == nil {
		return ""
	}
	if schema.Example != nil {
		return schema.Example
	}
	if schema.Format == "uuid" {
		return uuid.NewString()
	}
	if schema.Type.Is(openapi3.TypeString) && value != nil && value.In == "path" {
		return uuid.NewString()
	}
	if schema.Type.Is(openapi3.TypeString) {
		return exampleStringFromParamInfo(value)
	}
	if schema.Type.Is(openapi3.TypeNumber) || schema.Type.Is(openapi3.TypeInteger) {
		return 12345
	}
	if schema.Type.Is(openapi3.TypeString) {
		return "anystring"
	}

	for _, s := range extractSchemas(schema.AnyOf) {
		if ex := GenerateExample(Param{Schema: &s, Value: value}); ex != nil {
			return ex
		}
	}
	return ""
}

func exampleStringFromParamInfo(param *openapi3.Parameter) any {
	if param == nil {
		return ""
	}
	if strings.Contains(param.Name, "date") {
		return time.Now().Format(time.RFC3339)
	}
	return fmt.Sprintf("example_for_%s", param.Name)
}

func extractSchemas(refs openapi3.SchemaRefs) (out []openapi3.Schema) {
	if refs == nil {
		return
	}
	for _, ref := range refs {
		if ref.Value != nil {
			out = append(out, *ref.Value)
		}
	}
	return
}
