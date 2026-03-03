package sfuzz_test

import (
	"strings"
	"testing"

	"github.com/simcap/sfuzz"
)

func TestTargetGeneration(t *testing.T) {
	s := `GET https://example.com/FUZZ12345UID?id=FUZZabcSTR {"age": FUZZ35NUM}`
	all, err := sfuzz.Parse(strings.NewReader(s))
	Equal(t, err, nil)
	Equal(t, len(all), 1)

	request := all[0]

	targets, err := request.GenerateTargets()
	Equal(t, err, nil)
	Equal(t, len(targets), 3)

	one := targets[0]
	Equal(t, one.URL, "https://example.com/FUZZ12345UID?id=abc")
	EqualBytes(t, one.Body, []byte(`{"age": 35}`))

	two := targets[1]
	Equal(t, two.URL, "https://example.com/12345?id=FUZZabcSTR")
	EqualBytes(t, two.Body, []byte(`{"age": 35}`))

	three := targets[2]
	Equal(t, three.URL, "https://example.com/12345?id=abc")
	EqualBytes(t, three.Body, []byte(`{"age": FUZZ35NUM}`))
}
