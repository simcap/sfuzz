package sfuzz_test

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/simcap/sfuzz"
	"github.com/simcap/sfuzz/assert"
)

func TestHTMLOutput(t *testing.T) {
	codes := []int{http.StatusAccepted, http.StatusOK, http.StatusNotFound, http.StatusInternalServerError}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /any", func(w http.ResponseWriter, r *http.Request) {
		rand.Shuffle(len(codes), func(i, j int) { codes[i], codes[j] = codes[j], codes[i] })
		w.WriteHeader(codes[0])
		w.Write(fmt.Appendf(nil, `{"code": %d}`, codes[0]))
	})

	server := httptest.NewServer(mux)
	file := fmt.Sprintf("%s/any?id=FUZZSTR\nPOST %[1]s/any?id=FUZZSTR\nPUT %[1]s/any?id=FUZZSTR", server.URL)

	requests, err := sfuzz.Parse(strings.NewReader(file))
	assert.Equal(t, err, nil)

	runner := sfuzz.NewRunner(requests,
		sfuzz.WithSelector(func(sfuzz.FuzzKeyword) sfuzz.Generator {
			return sfuzz.CounterGenerator(1)
		}),
	)
	runner.Run(t.Context())

	out, err := sfuzz.NewHTMLOutput(runner.Results())
	assert.Equal(t, err, nil)

	err = out.Write(os.Stdout)
	assert.Equal(t, err, nil)
}
