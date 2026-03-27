package sfuzz

import (
	"cmp"
	"encoding/json"
	"fmt"
	"maps"
	"net/url"
	"slices"
	"strconv"
	"strings"
)

var (
	FuzzPrefix     = "FUZZ"
	TypeSuffixSize = 3

	GenericString Kind = "STR"
	Numeral       Kind = "NUM"
	UniversalID   Kind = "UID"
	Date          Kind = "DTE"
	Time          Kind = "TME"
	Boolean       Kind = "BOL"
)

type Kind string

type Keyword interface {
	Kind() Kind
	Example() string
	Ref() string
	Replace(Request, any) (Request, error)
}

type QueryKeyword struct {
	name    string
	kind    Kind
	example string
}

func (p QueryKeyword) Example() string { return p.example }
func (p QueryKeyword) Ref() string     { return p.name }
func (p QueryKeyword) Kind() Kind      { return p.kind }

func (p QueryKeyword) Replace(r Request, v any) (Request, error) {
	out := make(url.Values)
	for k, values := range r.URL.Query() {
		if p.Ref() == k {
			out[k] = []string{fmt.Sprintf("%v", v)}
		} else {
			out[k] = values
		}
	}
	if len(out) > 0 {
		r.URL.RawQuery = out.Encode()
	}
	return r, nil
}

type JSONBodyKeyword struct {
	kind          Kind
	example, name string
}

func (k JSONBodyKeyword) Example() string { return k.example }
func (k JSONBodyKeyword) Kind() Kind      { return k.kind }
func (k JSONBodyKeyword) Ref() string     { return k.name }
func (k JSONBodyKeyword) Replace(r Request, exotic any) (Request, error) {
	var v map[string]any
	if err := json.Unmarshal(r.Body, &v); err != nil {
		return r, err
	}
	for key := range v {
		if k.Ref() == key {
			switch k.Kind() {
			case Boolean:
				switch vv := exotic.(type) {
				case bool:
					v[key] = exotic
				case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
					v[key] = exotic
				case string:
					b, err := strconv.ParseBool(vv)
					if err == nil {
						v[key] = b
					}
				}
			case Numeral:
				if num, err := json.Number(fmt.Sprintf("%v", exotic)).Int64(); err == nil {
					v[key] = num
				} else if f, err := json.Number(fmt.Sprintf("%v", exotic)).Float64(); err == nil {
					v[key] = f
				} else {
					v[key] = exotic
				}
			default:
				v[key] = exotic
			}
		}
	}
	body, err := json.Marshal(v)
	if err != nil {
		return r, err
	}
	r.Body = body
	return r, nil
}

type PathKeyword struct {
	index   int
	name    string
	kind    Kind
	example string
}

func (k PathKeyword) Example() string { return k.example }
func (k PathKeyword) Ref() string     { return k.name }
func (k PathKeyword) Kind() Kind      { return k.kind }
func (k PathKeyword) Replace(r Request, v any) (Request, error) {
	segments := strings.Split(strings.TrimLeft(r.URL.Path, "/"), "/")
	segments[k.index] = fmt.Sprintf("%v", v)
	u := r.URL
	u.Path = strings.Join(segments, "/")
	r.URL.Path = u.Path
	return r, nil
}

func ParseQuery(u url.URL) ([]Keyword, error) {
	var holders []Keyword
	for _, key := range slices.Sorted(maps.Keys(u.Query())) {
		value := u.Query().Get(key)
		kind, example, found := parseKeywordString(value)
		if found {
			holders = append(holders, QueryKeyword{kind: kind, example: example, name: key})
		}
	}
	return holders, nil
}

func ParsePath(u url.URL) ([]Keyword, error) {
	var keywords []Keyword
	for i, segment := range strings.Split(strings.TrimLeft(u.Path, "/"), "/") {
		kind, example, found := parseKeywordString(segment)
		if found {
			keywords = append(keywords, PathKeyword{index: i, kind: kind, example: example, name: fmt.Sprintf("path_%d", i)})
		}
	}
	return keywords, nil
}

func ParseJSONBody(body []byte) ([]Keyword, error) {
	if len(body) == 0 {
		return nil, nil
	}
	var (
		keywords []Keyword
		v        map[string]any
	)
	if err := json.Unmarshal(body, &v); err != nil {
		return nil, err
	}
	for key, value := range v {
		kind, example, found := parseKeywordString(fmt.Sprintf("%v", value))
		if found {
			keywords = append(keywords, JSONBodyKeyword{kind: kind, example: example, name: key})
		}
	}
	return sortKeywords(keywords), nil
}

func parseKeywordString(s string) (Kind, string, bool) {
	if len(s) < len(FuzzPrefix)+TypeSuffixSize {
		return GenericString, "", false
	}
	if !strings.HasPrefix(s, FuzzPrefix) {
		return GenericString, "", false
	}
	end := len(s) - TypeSuffixSize
	return Kind(s[end:]), s[len(FuzzPrefix):end], true
}

func sortKeywords(keywords []Keyword) []Keyword {
	slices.SortFunc(keywords, func(k1 Keyword, k2 Keyword) int {
		return cmp.Compare(k1.Ref(), k2.Ref())
	})
	return keywords
}
