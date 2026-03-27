package spec

import (
	"cmp"

	"github.com/getkin/kin-openapi/openapi3"
)

type Value struct {
	v      string
	json   bool
	key    string
	schema *openapi3.Schema
	param  *openapi3.Parameter
}

func NewValue(p *openapi3.Parameter) *Value {
	return &Value{param: p}
}

func NewJSONValue(key string, s *openapi3.Schema) *Value {
	return &Value{key: key, json: true, schema: s}
}

func FromValue(v *Value) *Value {
	if v == nil {
		return new(Value)
	}
	if v.json {
		return NewJSONValue(v.key, v.schema)
	}
	return NewValue(v.param).WithSchema(v.schema)
}

func (v *Value) GetSchema() *openapi3.Schema {
	if v.schema == nil {
		return openapi3.NewSchema()
	}
	if v.schema.AnyOf != nil {
		for _, value := range extractSchemas(v.schema.AnyOf) {
			if value.Type.Is(openapi3.TypeNull) {
				continue
			}
			return cmp.Or(value, openapi3.NewSchema())
		}
	}
	return v.schema
}

func (v *Value) IsNumber() bool {
	return v.GetSchema().Type.Is(openapi3.TypeNumber) || v.GetSchema().Type.Is(openapi3.TypeInteger)
}

func (v *Value) IsBoolean() bool {
	return v.GetSchema().Type.Is(openapi3.TypeBoolean)
}

func (v *Value) IsString() bool {
	return v.GetSchema().Type.Is(openapi3.TypeString)
}
func (v *Value) IsInPath() bool {
	return v.param != nil && v.param.In == "path"
}

func (v *Value) IsDate() bool {
	switch v.GetSchema().Format {
	case "date", "date-time", "datetime":
		return true
	default:
		return false
	}
}

func (v *Value) IsUUID() bool {
	return v.GetSchema().Format == "uuid"
}

func (v *Value) IsJSONArray() bool {
	if !v.json {
		return false
	}
	return v.getJSONType() == jsonArray
}

func (v *Value) IsJSONObject() bool {
	if !v.json {
		return false
	}
	return v.getJSONType() == jsonObject
}

func (v *Value) IsJSONBoolean() bool {
	if !v.json {
		return false
	}
	return v.getJSONType() == jsonBoolean
}

func (v *Value) IsJSONNumber() bool {
	if !v.json {
		return false
	}
	return v.getJSONType() == jsonNumber
}

func (v *Value) Example() any {
	return v.GetSchema().Example
}

func (v *Value) WithSchema(s *openapi3.Schema) *Value {
	v.schema = s
	return v
}

func (v *Value) NoSchema() bool { return v.schema == nil }

func (v *Value) Name() string {
	if v.json {
		return v.key
	}
	if v.param != nil {
		return v.param.Name
	}
	return ""
}

func (v *Value) getJSONType() jsonType {
	switch {
	case v.GetSchema().Type.Is(openapi3.TypeString):
		return jsonString
	case v.GetSchema().Type.Is(openapi3.TypeNumber) || v.schema.Type.Is(openapi3.TypeInteger):
		return jsonNumber
	case v.GetSchema().Type.Is(openapi3.TypeArray):
		return jsonArray
	case v.GetSchema().Type.Is(openapi3.TypeBoolean):
		return jsonBoolean
	case v.GetSchema().Type.Is(openapi3.TypeNull):
		return jsonNull
	case v.GetSchema().Type.Is(openapi3.TypeObject):
		return jsonObject
	default:
		return jsonString
	}
}

type jsonType uint

const (
	jsonString = iota
	jsonNumber
	jsonArray
	jsonBoolean
	jsonObject
	jsonNull
)
