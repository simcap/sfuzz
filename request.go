package sfuzz

import (
	"cmp"
	"encoding/json"
	"net/url"
	"slices"
)

// Request represents a parsed line from usually a fuzz file or any io.Reader
type Request struct {
	Verb           string
	URL            url.URL
	Body           json.RawMessage
	ParsedKeywords []FuzzKeyword
}

func (r Request) BuildFuzzCandidates() ([]FuzzCandidate, error) {
	slices.SortFunc(r.ParsedKeywords, func(k1 FuzzKeyword, k2 FuzzKeyword) int {
		return cmp.Compare(k2.Start, k1.Start)
	})

	var targets []FuzzCandidate
	for index := range r.ParsedKeywords {
		u, body := r.URL, r.Body

		for current, k := range r.ParsedKeywords {
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
		targets = append(targets, FuzzCandidate{Request: Request{URL: u, Verb: r.Verb, Body: body}, Keyword: r.ParsedKeywords[index]})
	}
	return resetKeywordIndices(targets)
}

func resetKeywordIndices(targets []FuzzCandidate) (out []FuzzCandidate, err error) {
	for _, target := range targets {
		switch target.Keyword.Location {
		case PathKeyword:
			target.Keyword, err = ParseUniqueKeyword(target.URL.Path)
			if err != nil {
				return targets, err
			}
			target.Keyword.Location = PathKeyword
		case QueryKeyword:
			target.Keyword, err = ParseUniqueKeyword(target.URL.RawQuery)
			if err != nil {
				return targets, err
			}
			target.Keyword.Location = QueryKeyword
		case BodyKeyword:
			target.Keyword, err = ParseUniqueKeyword(string(target.Body))
			if err != nil {
				return targets, err
			}
			target.Keyword.Location = BodyKeyword
		}

		out = append(out, target)
	}
	return
}
