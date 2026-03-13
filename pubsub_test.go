package sfuzz_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/simcap/sfuzz"
	"github.com/simcap/sfuzz/assert"
)

func TestPubSub(t *testing.T) {
	file := fmt.Sprintf(`
https://nice.fr/one?id=FUZZ123STR&age=FUZZDTE
POST https://nice.fr/two/FUZZu8uUID {"name": "FUZZjohnSTR"}
`)

	requests, err := sfuzz.Parse(strings.NewReader(file))
	assert.Equal(t, err, nil)

	ps := sfuzz.NewPubSub()
	for _, request := range requests {
		candidates, err := request.BuildFuzzCandidates()
		if err != nil {
			t.Fatal(err)
		}
		for _, candidate := range candidates {
			iter := sfuzz.NewIterator(sfuzz.CounterGenerator(5))
			ps.AddPublisher(iter, candidate.Hash())
			ps.AddSubscribers(candidate)
		}
	}

	out := make(chan sfuzz.Target)
	go func() {
		for {
			select {
			case <-t.Context().Done():
				return
			default:
				if ps.Publish(t.Context(), out) {
					close(out)
					return
				}
			}
		}
	}()
	for target := range out {
		fmt.Println(target)
	}
}
