package spec

import "github.com/getkin/kin-openapi/openapi3"

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

func (v *Value) IsNumber() bool {
	if v.schema == nil {
		return false
	}
	return v.schema.Type.Is(openapi3.TypeNumber) || v.schema.Type.Is(openapi3.TypeInteger)
}

func (v *Value) IsString() bool {
	if v.schema == nil {
		return false
	}
	return v.schema.Type.Is(openapi3.TypeString)
}
func (v *Value) IsInPath() bool {
	return v.param != nil && v.param.In == "path"
}

func (v *Value) IsDate() bool {
	if v.schema == nil {
		return false
	}
	switch v.schema.Format {
	case "date", "date-time", "datetime":
		return true
	default:
		return false
	}
}

func (v *Value) IsUUID() bool {
	if v.schema == nil {
		return false
	}
	return v.schema.Format == "uuid"
}

func (v *Value) IsJSONArray() bool {
	if v.schema == nil {
		return false
	}
	if !v.json {
		return false
	}
	return v.jsonType() == jsonArray
}

func (v *Value) IsJSONObject() bool {
	if v.schema == nil || !v.json {
		return false
	}
	return v.jsonType() == jsonObject
}

func (v *Value) IsJSONNumber() bool {
	if v.schema == nil || !v.json {
		return false
	}
	return v.jsonType() == jsonNumber
}

func (v *Value) Example() any {
	if v.schema == nil {
		return ""
	}
	return v.schema.Example
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

func (v *Value) jsonType() jsonType {
	if v.schema == nil {
		return jsonNull
	}
	if v.schema.AnyOf != nil {
		for _, ss := range extractSchemas(v.schema.AnyOf) {
			return NewValue(nil).WithSchema(ss).jsonType()
		}
	}
	switch {
	case v.schema.Type.Is(openapi3.TypeString):
		return jsonString
	case v.schema.Type.Is(openapi3.TypeNumber) || v.schema.Type.Is(openapi3.TypeInteger):
		return jsonNumber
	case v.schema.Type.Is(openapi3.TypeArray):
		return jsonArray
	case v.schema.Type.Is(openapi3.TypeBoolean):
		return jsonBoolean
	case v.schema.Type.Is(openapi3.TypeNull):
		return jsonNull
	case v.schema.Type.Is(openapi3.TypeObject):
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
