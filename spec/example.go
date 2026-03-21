package spec

import (
	"fmt"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/google/uuid"
)

func GenerateExample(param Param) any {
	schema := param.Schema
	if schema == nil {
		return ""
	}
	if schema.Example != nil {
		return schema.Example
	}
	if schema.Format == "uuid" {
		return uuid.NewString()
	}
	if schema.Type.Is(openapi3.TypeString) && param.Location == "path" {
		return uuid.NewString()
	}
	if schema.Type.Is(openapi3.TypeString) {
		return exampleStringFromParamInfo(param.Name)
	}
	if schema.Type.Is(openapi3.TypeNumber) || schema.Type.Is(openapi3.TypeInteger) {
		return 12345
	}
	if schema.Type.Is(openapi3.TypeString) {
		return "anystring"
	}

	for _, s := range extractSchemas(schema.AnyOf) {
		if ex := GenerateExample(Param{Schema: s, Name: param.Name, Location: param.Location}); ex != nil {
			return ex
		}
	}
	return ""
}

func exampleStringFromParamInfo(name string) any {
	if strings.Contains(name, "date") {
		return time.Now().Format(time.RFC3339)
	}
	if strings.Contains(name, "mail") {
		return "fuzz@testing"
	}
	return fmt.Sprintf("%s_example", name)
}

func extractSchemas(refs openapi3.SchemaRefs) (out []*openapi3.Schema) {
	if refs == nil {
		return
	}
	for _, ref := range refs {
		if ref.Value != nil {
			out = append(out, ref.Value)
		}
	}
	return
}
