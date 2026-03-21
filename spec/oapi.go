package spec

import (
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"maps"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

type OAPI struct {
	doc        *openapi3.T
	server     string
	noExamples bool
}

func NewOAPISpec(r io.Reader, opts ...option) (*OAPI, error) {
	doc, err := openapi3.NewLoader().LoadFromIoReader(r)
	oapi := &OAPI{doc: doc}
	for _, opt := range opts {
		opt(oapi)
	}
	return oapi, err
}

func (o *OAPI) GenerateFuzzFile(w io.Writer) error {
	server := o.Server()
	for op := range o.operationsIter() {
		uri := fmt.Sprintf("%s%s", server, o.pathWithFuzzKeywords(op))
		if q := o.queryWithFuzzKeywords(op); q != "" {
			uri = fmt.Sprintf("%s?%s", uri, q)
		}
		fmt.Fprintf(w, "%s %s", op.Method, uri)

		if object := o.bodyWithFuzzKeywords(op); len(object) > 0 {
			v, err := json.Marshal(object)
			if err != nil {
				return err
			}
			fmt.Fprintf(w, " %s", string(v))
		}
		fmt.Fprintf(w, "\n")
	}

	return nil
}

func (o *OAPI) Server() string {
	if o.server != "" {
		return o.server
	}
	for _, s := range o.doc.Servers {
		return s.URL
	}
	return ""
}

func (o *OAPI) operationsIter() iter.Seq[PathOperation] {
	return func(yield func(PathOperation) bool) {
		paths := o.doc.Paths.Map()
		for _, key := range slices.Sorted(maps.Keys(paths)) {
			path, item := key, paths[key]

			op := PathOperation{Path: path, Item: item}
			if item.Get != nil {
				op.Method, op.Operation = http.MethodGet, item.Get
				if !yield(op) {
					return
				}
			}
			if item.Post != nil {
				op.Method, op.Operation = http.MethodPost, item.Post
				if !yield(op) {
					return
				}
			}
			if item.Put != nil {
				op.Method, op.Operation = http.MethodPut, item.Put
				if !yield(op) {
					return
				}
			}
			if item.Delete != nil {
				op.Method, op.Operation = http.MethodDelete, item.Delete
				if !yield(op) {
					return
				}
			}
			if item.Patch != nil {
				op.Method, op.Operation = http.MethodPatch, item.Patch
				if !yield(op) {
					return
				}
			}
		}
	}
}

type PathOperation struct {
	Method    string
	Path      string
	Item      *openapi3.PathItem
	Operation *openapi3.Operation
}

func (o *OAPI) pathWithFuzzKeywords(op PathOperation) string {
	out := op.Path
	for param := range op.pathParams() {
		name := param.Name
		keyword := o.generateKeyword(param)
		out = strings.Replace(out, fmt.Sprintf("{%s}", name), keyword, 1)
	}
	return out
}

func (o *OAPI) queryWithFuzzKeywords(op PathOperation) string {
	out := make(url.Values)
	for param := range op.queryParams() {
		out.Set(param.Name, o.generateKeyword(param))
	}
	return out.Encode()
}

func (o *OAPI) bodyWithFuzzKeywords(op PathOperation) map[string]any {
	out := make(map[string]any)
	for key, param := range op.bodyProperties() {
		switch key.Type {
		case JSONArray:
			out[key.Val] = []string{}
		case JSONObject:
			out[key.Val] = make(map[string]any)
		default:
			out[key.Val] = o.generateKeyword(param)
		}
	}
	return out
}

func (o *OAPI) generateKeyword(param Param) (keyword string) {
	keyword = "FUZZSTR"
	if param.Schema == nil {
		return
	}

	switch {
	case param.Schema.Type.Is("number"):
		keyword = "FUZZNUM"
	}
	switch param.Schema.Format {
	case "uuid":
		keyword = "FUZZUID"
	case "date", "date-time", "datetime":
		keyword = "FUZZDTE"
	}

	if !o.noExamples {
		if example := GenerateExample(param); example != nil {
			keyword = keyword[:4] + fmt.Sprintf("%v", example) + keyword[4:]
		}
	}

	return
}

type Param struct {
	Name     string
	Location string
	Schema   *openapi3.Schema
}

func fromOAPIParam(value *openapi3.Parameter) Param {
	return Param{Name: value.Name, Location: value.In}
}

func (op *PathOperation) queryParams() iter.Seq[Param] {
	return func(yield func(Param) bool) {
		for _, param := range op.Operation.Parameters {
			if v := param.Value; v != nil {
				if v.In == "query" {
					p := fromOAPIParam(v)
					if ref := v.Schema; ref != nil {
						p.Schema = ref.Value
					}
					if !yield(p) {
						return
					}
				}
			}
		}
	}
}

func (op *PathOperation) pathParams() iter.Seq[Param] {
	return func(yield func(Param) bool) {
		for _, param := range op.Operation.Parameters {
			if v := param.Value; v != nil {
				if v.In == "path" {
					p := fromOAPIParam(v)
					if ref := v.Schema; ref != nil {
						p.Schema = ref.Value
					}
					if !yield(p) {
						return
					}
				}
			}
		}
	}
}

type JSONType uint

const (
	JSONString = iota
	JSONNumber
	JSONArray
	JSONBoolean
	JSONObject
)

type jsonKey struct {
	Val  string
	Type JSONType
}

func (op *PathOperation) bodyProperties() iter.Seq2[jsonKey, Param] {
	return func(yield func(jsonKey, Param) bool) {
		if ref := op.Operation.RequestBody; ref != nil {
			if val := ref.Value; val != nil {
				for _, media := range val.Content {
					if schemaRef := media.Schema; schemaRef != nil {
						if schema := schemaRef.Value; schema != nil {
							for k, v := range schema.Properties {
								key := jsonKey{
									Val:  k,
									Type: getJSONType(v.Value),
								}
								if !yield(key, Param{Schema: v.Value, Name: k, Location: "body"}) {
									return
								}
							}
						}
					}
				}
			}
		}
	}
}

func getJSONType(s *openapi3.Schema) JSONType {
	if s == nil {
		return JSONString
	}

	if s.AnyOf != nil {
		for _, ss := range extractSchemas(s.AnyOf) {
			return getJSONType(ss)
		}
	}

	switch {
	case s.Type.Is("string"):
		return JSONString
	case s.Type.Is("number"):
		return JSONNumber
	case s.Type.Is("array"):
		return JSONArray
	case s.Type.Is("object"):
		return JSONObject
	default:
		return JSONString
	}
}

type option func(*OAPI)

func WithNoExamples() option {
	return func(o *OAPI) {
		o.noExamples = true
	}
}

func WithServer(s string) option {
	return func(o *OAPI) {
		o.server = s
	}
}
