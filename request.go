package sfuzz

import (
	"encoding/json"
	"net/url"
)

type FuzzRequest struct {
	Methods []string
	URL     *url.URL
	Body    json.RawMessage

	URLKeywords   []FuzzKeyword
	BodyKeywords  []FuzzKeyword
	QueryKeywords []FuzzKeyword
}

func (r FuzzRequest) GenerateTargets() []Target {
	return nil
}
