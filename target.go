package sfuzz

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
)

type Target struct {
	Verb string
	URL  string
	Body []byte
}

func (t Target) ToHTTPRequest(ctx context.Context) *http.Request {
	req, err := http.NewRequestWithContext(ctx, t.Verb, t.URL, bytes.NewReader(t.Body))
	if err != nil {
		panic(fmt.Sprintf("cannot create request (%s): %s", t.String(), err))
	}
	req.Header.Set("Content-Type", "application/json")
	return req
}

func (t Target) String() string { return fmt.Sprintf("%s %s", t.Verb, t.URL) }
