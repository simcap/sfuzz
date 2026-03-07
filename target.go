package sfuzz

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

type Target struct {
	Verb    string
	URL     url.URL
	Body    []byte
	Keyword FuzzKeyword
}

func (t Target) Replace(v any) (Target, error) {
	switch t.Keyword.Location {
	case PathKeyword:
		return t.replacePathKeyword(v)
	case QueryKeyword:
		return t.replaceQueryKeyword(v)
	case BodyKeyword:
		return t.replaceBodyKeyword(v)
	default:
		return Target{}, errors.New("cannot replace: invalid keyword location")
	}
}

func (t Target) ToHTTPRequest(ctx context.Context) *http.Request {
	req, err := http.NewRequestWithContext(ctx, t.Verb, t.URL.String(), bytes.NewReader(t.Body))
	if err != nil {
		panic(fmt.Sprintf("cannot create request (%s): %s", t.String(), err))
	}
	req.Header.Set("Content-Type", "application/json")
	return req
}

func (t Target) String() string { return fmt.Sprintf("%s %v", t.Verb, t.URL) }

func (t Target) replacePathKeyword(v any) (Target, error) {
	u := t.URL
	u.Path = fmt.Sprintf("%s%s%s", u.Path[:t.Keyword.Start], v, u.Path[t.Keyword.End:])
	target := Target{Verb: t.Verb, URL: u, Body: t.Body, Keyword: t.Keyword}
	return target, nil
}

func (t Target) replaceQueryKeyword(v any) (Target, error) {
	u := t.URL
	escaped := url.QueryEscape(fmt.Sprintf("%v", v))
	u.RawQuery = fmt.Sprintf("%s%s%s", u.RawQuery[:t.Keyword.Start], escaped, u.RawQuery[t.Keyword.End:])
	target := Target{Verb: t.Verb, URL: u, Body: t.Body, Keyword: t.Keyword}
	return target, nil
}

func (t Target) replaceBodyKeyword(v any) (Target, error) {
	body := fmt.Sprintf("%s%s%s", t.Body[:t.Keyword.Start], v, t.Body[t.Keyword.End:])
	target := Target{Verb: t.Verb, URL: t.URL, Body: []byte(body), Keyword: t.Keyword}
	return target, nil
}
