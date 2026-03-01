package sfuzz

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type Target struct {
	Verb string
	URL  URL
	Body []byte
}

func (t Target) ToHTTPRequest(ctx context.Context) *http.Request {
	req, err := http.NewRequestWithContext(ctx, t.Verb, t.URL.String(), bytes.NewReader(t.Body))
	if err != nil {
		panic(fmt.Sprintf("cannot create request (%s): %s", t.String(), err))
	}
	req.Header.Set("Content-Type", "application/json")
	return req
}

func (t Target) String() string { return fmt.Sprintf("%s %s", t.Verb, t.URL.String()) }

type URL struct {
	u *url.URL
}

func (u URL) String() string { return u.u.String() }

func ParseURL(s string) (URL, error) {
	u, err := url.Parse(s)
	return URL{u}, err
}

func MustParseURL(s string) URL {
	u, err := ParseURL(s)
	if err != nil {
		panic(err)
	}
	return u
}
