package spec

import (
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
	doc *openapi3.T
}

func NewOAPISpec(r io.Reader) (*OAPI, error) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromIoReader(r)
	return &OAPI{doc: doc}, err
}

func (o *OAPI) GenerateFuzzFile(w io.Writer) error {
	server := o.server()
	for op := range o.operationsIter() {
		uri := fmt.Sprintf("%s%s", server, op.pathWithFuzzKeywords())
		if q := op.queryWithFuzzKeywords(); q != "" {
			uri = fmt.Sprintf("%s?%s", uri, q)
		}
		fmt.Fprintf(w, "%s %s\n", op.Method, uri)
	}

	return nil
}

func (o *OAPI) server() string {
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

func (op *PathOperation) pathWithFuzzKeywords() string {
	out := op.Path
	for param := range op.pathParams() {
		name := param.Value.Name
		keyword := matchKeyword(param)
		out = strings.Replace(out, fmt.Sprintf("{%s}", name), keyword, 1)
	}
	return out
}

func (op *PathOperation) queryWithFuzzKeywords() string {
	out := make(url.Values)
	for param := range op.queryParams() {
		out.Set(param.Value.Name, matchKeyword(param))
	}
	return out.Encode()
}

func matchKeyword(param Param) (keyword string) {
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
	return
}

type Param struct {
	Value  *openapi3.Parameter
	Schema *openapi3.Schema
}

func (op *PathOperation) queryParams() iter.Seq[Param] {
	return func(yield func(Param) bool) {
		for _, param := range op.Operation.Parameters {
			if v := param.Value; v != nil {
				if v.In == "query" {
					p := Param{Value: v}
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
					p := Param{Value: v}
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
