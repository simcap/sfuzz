package spec

import (
	"fmt"
	"strings"
	"time"

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
	return ""
}

func exampleStringFromParamInfo(name string) any {
	if strings.Contains(name, "date") {
		return time.Now().Format(time.RFC3339)
	}
	if strings.Contains(name, "mail") {
		return "fuzz@testing"
	}
	if name == "sort" || name == "sort_by" {
		return "desc"
	}
	return fmt.Sprintf("%s_example", name)
}
