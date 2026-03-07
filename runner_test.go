package sfuzz_test

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/simcap/sfuzz"
)

func TestSimpleRun(t *testing.T) {
	var getCalled, postCalled bool
	mux := http.NewServeMux()
	mux.HandleFunc("/one", func(w http.ResponseWriter, r *http.Request) {
		getCalled = true
	})
	mux.HandleFunc("POST /two", func(w http.ResponseWriter, r *http.Request) {
		postCalled = true
	})

	server := httptest.NewServer(mux)
	requests := []sfuzz.Target{
		{Verb: "GET", URL: MustParseURL(t, fmt.Sprintf("%s/one", server.URL))},
		{Verb: "POST", URL: MustParseURL(t, fmt.Sprintf("%s/two", server.URL))},
	}

	log := slog.New(slog.NewTextHandler(t.Output(), nil))
	runner := sfuzz.NewRunner(sfuzz.WithLogger(log))
	runner.Run(t.Context(), requests)

	Equal(t, getCalled, true)
	Equal(t, postCalled, true)
}

func Equal[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Fatalf("\n got: %v\nwant: %v\n", got, want)
	}
}

func EqualBytes(t *testing.T, got, want []byte) {
	t.Helper()
	if !bytes.Equal(got, want) {
		t.Fatalf("\n got: %q\nwant: %q\n", got, want)
	}
}

func MustParseURL(t *testing.T, s string) url.URL {
	t.Helper()
	u, err := url.Parse(s)
	if err != nil {
		t.Fatal(err)
	}
	return *u
}
