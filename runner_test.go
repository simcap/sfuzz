package sfuzz_test

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/simcap/sfuzz"
	"github.com/simcap/sfuzz/assert"
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
	assert.Equal(t, err, nil)

	log := sfuzz.NewLogger(t.Output())

	var fuzzCount = rand.Intn(5)
	runner := sfuzz.NewRunner(requests,
		sfuzz.WithLogger(log),
		sfuzz.WithSelector(func(sfuzz.FuzzKeyword) sfuzz.Fuzzer {

			return sfuzz.CounterFuzzer(fuzzCount)
		}),
	)
	runner.Run(t.Context())

	assert.Equal(t, len(actualGets), 2*fuzzCount)
	assert.Equal(t, len(actualPosts), 2*fuzzCount)
}
