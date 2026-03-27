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
	for param := range op.pathValues() {
		name := param.Name()
		keyword := o.generateKeyword(param)
		out = strings.Replace(out, fmt.Sprintf("{%s}", name), keyword, 1)
	}
	return out
}

func (o *OAPI) queryWithFuzzKeywords(op PathOperation) string {
	out := make(url.Values)
	for param := range op.queryValues() {
		out.Set(param.Name(), o.generateKeyword(param))
	}
	return out.Encode()
}

func (o *OAPI) bodyWithFuzzKeywords(op PathOperation) map[string]any {
	out := make(map[string]any)
	for value := range op.bodyValues() {
		switch {
		case value.IsJSONArray():
			out[value.Name()] = []string{}
		case value.IsJSONObject():
			out[value.Name()] = make(map[string]any)
		default:
			out[value.Name()] = o.generateKeyword(value)
		}
	}
	return out
}

func (o *OAPI) generateKeyword(v *Value) (keyword string) {
	keyword = "FUZZSTR"
	switch {
	case v.NoSchema():
		return
	case v.IsNumber():
		keyword = "FUZZNUM"
	case v.IsUUID():
		keyword = "FUZZUID"
	case v.IsDate():
		keyword = "FUZZDTE"
	case v.IsBoolean():
		keyword = "FUZZBOL"
	}

	if !o.noExamples {
		if example := GenerateExample(v); example != nil {
			keyword = keyword[:4] + fmt.Sprintf("%v", example) + keyword[4:]
		}
	}

	return
}

func (op *PathOperation) queryValues() iter.Seq[*Value] {
	return func(yield func(*Value) bool) {
		for _, param := range op.Operation.Parameters {
			if v := param.Value; v != nil {
				if v.In == "query" {
					p := NewValue(v)
					if ref := v.Schema; ref != nil {
						p.WithSchema(ref.Value)
					}
					if !yield(p) {
						return
					}
				}
			}
		}
	}
}

func (op *PathOperation) pathValues() iter.Seq[*Value] {
	return func(yield func(*Value) bool) {
		for _, param := range op.Operation.Parameters {
			if v := param.Value; v != nil {
				if v.In == "path" {
					p := NewValue(v)
					if ref := v.Schema; ref != nil {
						p.WithSchema(ref.Value)
					}
					if !yield(p) {
						return
					}
				}
			}
		}
	}
}

func (op *PathOperation) bodyValues() iter.Seq[*Value] {
	return func(yield func(*Value) bool) {
		if ref := op.Operation.RequestBody; ref != nil {
			if val := ref.Value; val != nil {
				for _, media := range val.Content {
					if schemaRef := media.Schema; schemaRef != nil {
						if schema := schemaRef.Value; schema != nil {
							for k, v := range schema.Properties {
								if !yield(NewJSONValue(k, v.Value)) {
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
