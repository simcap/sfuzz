package sfuzz

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// FuzzCandidate represents a request with only one fuzz keyword left to fuzz.
type FuzzCandidate struct {
	Request
	Keyword FuzzKeyword
}

func (t FuzzCandidate) Replace(v any) (FuzzCandidate, error) {
	switch t.Keyword.Location {
	case PathKeyword:
		return t.replacePathKeyword(v)
	case QueryKeyword:
		return t.replaceQueryKeyword(v)
	case BodyKeyword:
		return t.replaceBodyKeyword(v)
	default:
		return FuzzCandidate{}, errors.New("cannot replace: invalid keyword location")
	}
}

func (t FuzzCandidate) ToHTTPRequest(ctx context.Context) *http.Request {
	req, err := http.NewRequestWithContext(ctx, t.Verb, t.URL.String(), bytes.NewReader(t.Body))
	if err != nil {
		panic(fmt.Sprintf("cannot create request (%s): %s", t.String(), err))
	}
	req.Header.Set("Content-Type", "application/json")
	return req
}

func (t FuzzCandidate) String() string { return fmt.Sprintf("%s %v", t.Verb, t.URL) }

func (t FuzzCandidate) replacePathKeyword(v any) (FuzzCandidate, error) {
	u := t.URL
	u.Path = fmt.Sprintf("%s%s%s", u.Path[:t.Keyword.Start], v, u.Path[t.Keyword.End:])
	target := FuzzCandidate{Request: Request{Verb: t.Verb, URL: u, Body: t.Body}, Keyword: t.Keyword}
	return target, nil
}

func (t FuzzCandidate) replaceQueryKeyword(v any) (FuzzCandidate, error) {
	u := t.URL
	escaped := url.QueryEscape(fmt.Sprintf("%v", v))
	u.RawQuery = fmt.Sprintf("%s%s%s", u.RawQuery[:t.Keyword.Start], escaped, u.RawQuery[t.Keyword.End:])
	target := FuzzCandidate{Request: Request{Verb: t.Verb, URL: u, Body: t.Body}, Keyword: t.Keyword}
	return target, nil
}

func (t FuzzCandidate) replaceBodyKeyword(v any) (FuzzCandidate, error) {
	body := fmt.Sprintf("%s%s%s", t.Body[:t.Keyword.Start], v, t.Body[t.Keyword.End:])
	target := FuzzCandidate{Request: Request{Verb: t.Verb, URL: t.URL, Body: []byte(body)}, Keyword: t.Keyword}
	return target, nil
}
