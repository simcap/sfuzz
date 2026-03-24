package sfuzz

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
)

// FuzzCandidate represents a request with only one fuzz keyword left to fuzz.
type FuzzCandidate struct {
	Request
	Keyword Keyword
}

func (t FuzzCandidate) Replace(v any) (FuzzCandidate, error) {
	request, err := t.Keyword.Replace(t.Request, v)
	if err != nil {
		return FuzzCandidate{}, err
	}
	return FuzzCandidate{Request: request, Keyword: t.Keyword}, nil
}

func (t FuzzCandidate) Hash() string {
	return fmt.Sprintf("%T%s%s%s", t.Keyword, t.Keyword.Ref(), t.Keyword.Kind(), t.Keyword.Example())
}

func (t FuzzCandidate) ToHTTPRequest(ctx context.Context, header http.Header) *http.Request {
	req, err := http.NewRequestWithContext(ctx, t.Verb, t.URL.String(), bytes.NewReader(t.Body))
	if err != nil {
		panic(fmt.Sprintf("cannot create request (%s): %s", t.String(), err))
	}
	for key, values := range header {
		req.Header[key] = values
	}
	req.Header.Set("Content-Type", "application/json")
	return req
}

func (t FuzzCandidate) String() string { return fmt.Sprintf("%s %v", t.Verb, t.URL) }
