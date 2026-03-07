package sfuzz_test

import (
	"io"
	"strings"
	"testing"

	"github.com/simcap/sfuzz"
)

func TestTargetReplace(t *testing.T) {

	t.Run("replace in path", func(t *testing.T) {
		target := generateUniqueTarget(t, "https://example.com/books/FUZZSTR?expiry=23526")
		t.Run("is idempotent", func(t *testing.T) {
			value := "anything"
			for range 2 {
				replaced, err := target.Replace(value)
				Equal(t, err, nil)
				req := replaced.ToHTTPRequest(t.Context())
				Equal(t, req.URL.String(), "https://example.com/books/anything?expiry=23526")
			}
		})
	})

	t.Run("replace in query", func(t *testing.T) {
		target := generateUniqueTarget(t, "https://example.com/books/1234?expiry=FUZZDTE")
		t.Run("is idempotent", func(t *testing.T) {
			value := "to encode ++"
			for range 2 {
				replaced, err := target.Replace(value)
				Equal(t, err, nil)
				req := replaced.ToHTTPRequest(t.Context())
				Equal(t, req.URL.String(), "https://example.com/books/1234?expiry=to+encode+%2B%2B")
			}
		})
	})

	t.Run("replace in body", func(t *testing.T) {
		target := generateUniqueTarget(t, `https://example.com/books/1234?expiry=2023-02-01 {"stamp": "FUZZTME"}`)
		t.Run("is idempotent", func(t *testing.T) {
			value := "some json value"
			for range 2 {
				replaced, err := target.Replace(value)
				Equal(t, err, nil)
				req := replaced.ToHTTPRequest(t.Context())
				body, err := io.ReadAll(req.Body)
				Equal(t, err, nil)
				Equal(t, string(body), `{"stamp": "some json value"}`)
			}
		})
	})
}

func generateUniqueTarget(t *testing.T, s string) sfuzz.Target {
	requests, err := sfuzz.Parse(strings.NewReader(s))
	Equal(t, err, nil)
	Equal(t, len(requests), 1)
	targets, err := requests[0].BuildTargets()
	Equal(t, err, nil)
	Equal(t, len(targets), 1)
	return targets[0]
}
