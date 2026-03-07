package sfuzz

import (
	"cmp"
	"encoding/json"
	"net/url"
	"slices"
)

type FuzzRequest struct {
	Methods []string
	URL     *url.URL
	Body    json.RawMessage

	Keywords []FuzzKeyword
}

func (r FuzzRequest) BuildTargets() ([]Target, error) {
	slices.SortFunc(r.Keywords, func(k1 FuzzKeyword, k2 FuzzKeyword) int {
		return cmp.Compare(k2.Start, k1.Start)
	})

	var targets []Target
	for index, keyword := range r.Keywords {
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
		targets = append(targets, Target{URL: u, Body: body, Keyword: keyword})
	}
	return targets, nil
}
