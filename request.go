package sfuzz

import (
	"encoding/json"
)

type FuzzRequest struct {
	Methods []string
	URL     string
	Body    json.RawMessage
}

func (r FuzzRequest) GenerateTargets() []Target {
	return nil
}
