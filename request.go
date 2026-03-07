package sfuzz

import (
	"cmp"
	"encoding/json"
	"fmt"
	"net/url"
	"slices"
)

type FuzzRequest struct {
	Verb     string
	URL      *url.URL
	Body     json.RawMessage
	Keywords []FuzzKeyword
}

func (r FuzzRequest) BuildTargets() ([]Target, error) {
	slices.SortFunc(r.Keywords, func(k1 FuzzKeyword, k2 FuzzKeyword) int {
		return cmp.Compare(k2.Start, k1.Start)
	})

	var targets []Target
	for index := range r.Keywords {
		u, body := *r.URL, r.Body

		for current, k := range r.Keywords {
			if current == index {
				continue
			}
			switch k.Location {
			case PathKeyword:
				u.Path = u.Path[0:k.Start] + k.Example + u.Path[k.End:]
			case QueryKeyword:
				u.RawQuery = u.RawQuery[0:k.Start] + k.Example + u.RawQuery[k.End:]
			case BodyKeyword:
				body = slices.Concat(body[0:k.Start], []byte(k.Example), body[k.End:])
			}
		}
		newTarget := Target{URL: u, Verb: r.Verb, Body: body, Keyword: r.Keywords[index]}
		targets = append(targets, newTarget)
	}
	return resetKeywordIndices(targets)
}

func resetKeywordIndices(targets []Target) (out []Target, err error) {
	for _, target := range targets {
		switch target.Keyword.Location {
		case PathKeyword:
			target.Keyword, err = parseUniqueKeywords(target.URL.Path)
			if err != nil {
				return targets, err
			}
			target.Keyword.Location = PathKeyword
		case QueryKeyword:
			target.Keyword, err = parseUniqueKeywords(target.URL.RawQuery)
			if err != nil {
				return targets, err
			}
			target.Keyword.Location = QueryKeyword
		case BodyKeyword:
			target.Keyword, err = parseUniqueKeywords(string(target.Body))
			if err != nil {
				return targets, err
			}
			target.Keyword.Location = BodyKeyword
		}

		out = append(out, target)
	}
	return
}

func parseUniqueKeywords(s string) (FuzzKeyword, error) {
	keywords, err := ParseKeywords(s)
	if err != nil {
		return FuzzKeyword{}, err
	}
	if len(keywords) != 1 {
		return FuzzKeyword{}, fmt.Errorf("expected 1 keyword, got %d (%v)", len(keywords), s)
	}
	return keywords[0], nil
}
