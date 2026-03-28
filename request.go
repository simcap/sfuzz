package sfuzz

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// Request represents a parsed line from usually a fuzz file or any io.Reader
type Request struct {
	Verb           string
	URL            url.URL
	Body           json.RawMessage
	ParsedKeywords []Keyword
	Signature      string
}

func (r Request) BuildFuzzCandidates() ([]FuzzCandidate, error) {
	var targets []FuzzCandidate
	for index := range r.ParsedKeywords {
		unique := r.ParsedKeywords[index]

		request := r
		for current, keyword := range r.ParsedKeywords {
			if current == index {
				continue
			}
			var err error
			request, err = keyword.Replace(request, keyword.Example())
			if err != nil {
				return nil, err
			}
		}
		targets = append(targets, FuzzCandidate{Request: request, Keyword: unique})
	}
	return targets, nil
}

type RequestStats struct {
	Count    int
	Keywords int
	servers  map[string]struct{}
}

func (s RequestStats) AddServer(host string) {
	if s.servers == nil {
		s.servers = make(map[string]struct{})
	}
	s.servers[host] = struct{}{}
}

func (s RequestStats) Servers() []string {
	return slices.Collect(maps.Keys(s.servers))
}

func (r Request) AutoGenerateKeywords() (Request, error) {
	u, err := url.Parse(r.URL.String())
	if err != nil {
		return r, err
	}

	segments := strings.Split(u.Path, "/")
	if len(segments) > 1 {
		segments = slices.DeleteFunc(segments, func(s string) bool {
			return s == ""
		})
		var out []string
		for _, segment := range segments[1:] {
			kind := keywordType(segment)
			out = append(out, fmt.Sprintf("%s%s%s", "FUZZ", segment, kind))
		}
		if len(out) > 0 {
			u.Path = fmt.Sprintf("/%s/%s", segments[0], strings.Join(out, "/"))
		}
	}

	values := url.Values{}
	for k, all := range u.Query() {
		for _, v := range all {
			kind := keywordType(v)
			values[k] = append(values[k], fmt.Sprintf("%s%s%s", "FUZZ", v, kind))
		}
	}
	u.RawQuery = values.Encode()

	r.URL = *u
	err = collectKeywords(&r)
	return r, err
}

func keywordType(s string) Kind {
	if _, err := strconv.Atoi(s); err == nil {
		return Numeral
	}
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return Numeral
	}
	if err := uuid.Validate(s); err == nil {
		return UniversalID
	}

	return GenericString
}
