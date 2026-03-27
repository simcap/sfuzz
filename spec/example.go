package spec

import (
	"fmt"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/google/uuid"
)

func GenerateExample(value *Value) any {
	if value.NoSchema() {
		return ""
	}
	if ex := value.Example(); ex != nil {
		return ex
	}
	if value.IsUUID() || (value.IsString() && value.IsInPath()) {
		return uuid.NewString()
	}
	if value.IsString() {
		return exampleStringFromParamInfo(value.Name())
	}
	if value.IsNumber() {
		return 12345
	}
	if value.IsBoolean() {
		return true
	}
	if value.IsString() {
		return "any_string"
	}

	for _, s := range extractSchemas(value.schema.AnyOf) {
		if ex := GenerateExample(FromValue(value).WithSchema(s)); ex != nil {
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
