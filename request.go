package sfuzz

import (
	"encoding/json"
	"net/url"
)

type FuzzRequest struct {
	Methods []string
	URL     *url.URL
	Body    json.RawMessage

	Keywords []FuzzKeyword
}

func (r FuzzRequest) GenerateTargets() ([]Target, error) {
	var targets []Target
	for i := 0; i < len(r.Keywords); i++ {
		u := *r.URL
		body := r.Body
		for current, k := range r.Keywords {
			if current == i {
				continue
			}
			switch k.Location {
			case PathKeyword:
				s := u.Path
				s = s[0:k.Start] + k.Spec.GenerateExample() + s[k.End:]
				u.Path = s
			case QueryKeyword:
				s := u.RawQuery
				s = s[0:k.Start] + k.Spec.GenerateExample() + s[k.End:]
				u.RawQuery = s
			case BodyKeyword:
				s := string(body)
				s = s[0:k.Start] + k.Spec.GenerateExample() + s[k.End:]
				body = []byte(s)
			}
		}
		targets = append(targets, Target{URL: u.String(), Body: body})
	}
	return targets, nil
}
