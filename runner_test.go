package sfuzz_test

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/simcap/sfuzz"
)

func TestRunner(t *testing.T) {
	var actualGets, actualPosts []*http.Request
	mux := http.NewServeMux()
	mux.HandleFunc("GET /one", func(w http.ResponseWriter, r *http.Request) {
		actualGets = append(actualGets, r)
	})
	mux.HandleFunc("POST /two/{id}", func(w http.ResponseWriter, r *http.Request) {
		actualPosts = append(actualPosts, r)
	})

	server := httptest.NewServer(mux)
	file := fmt.Sprintf(`
%s/one?id=FUZZ123STR&age=FUZZDTE
POST %s/two/FUZZu8uUID {"name": "FUZZjohnSTR"}
`, server.URL, server.URL)

	requests, err := sfuzz.Parse(strings.NewReader(file))
	Equal(t, err, nil)

	log := sfuzz.NewLogger(t.Output())

	var fuzzCount = rand.Intn(5)
	runner := sfuzz.NewRunner(
		sfuzz.WithLogger(log),
		sfuzz.WithSelector(func(sfuzz.FuzzKeyword) sfuzz.Generator {
			return sfuzz.CounterGenerator(fuzzCount)
		}),
	)

	runner.Run(t.Context(), requests)

	Equal(t, len(actualGets), fuzzCount*2)
	Equal(t, len(actualPosts), fuzzCount*2)
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
