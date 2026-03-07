package sfuzz_test

import (
	"strings"
	"testing"

	"github.com/simcap/sfuzz"
)

func TestBuildTargets(t *testing.T) {
	s := `GET https://example.com/FUZZjohnSTR/FUZZ12345UID?id=FUZZabcSTR&city=FUZZParisSTR {"age": FUZZ35NUM, "date": "FUZZ2024-09-08DTE"}`
	all, err := sfuzz.Parse(strings.NewReader(s))
	Equal(t, err, nil)
	Equal(t, len(all), 1)

	request := all[0]

	targets, err := request.BuildFuzzCandidates()
	Equal(t, err, nil)
	Equal(t, len(targets), 6)

	one := targets[5]
	Equal(t, one.URL.String(), "https://example.com/FUZZjohnSTR/12345?id=abc&city=Paris")
	EqualBytes(t, one.Body, []byte(`{"age": 35, "date": "2024-09-08"}`))

	two := targets[2]
	Equal(t, two.URL.String(), "https://example.com/john/FUZZ12345UID?id=abc&city=Paris")
	EqualBytes(t, two.Body, []byte(`{"age": 35, "date": "2024-09-08"}`))

	three := targets[4]
	Equal(t, three.URL.String(), "https://example.com/john/12345?id=FUZZabcSTR&city=Paris")
	EqualBytes(t, three.Body, []byte(`{"age": 35, "date": "2024-09-08"}`))

	four := targets[1]
	Equal(t, four.URL.String(), "https://example.com/john/12345?id=abc&city=FUZZParisSTR")
	EqualBytes(t, four.Body, []byte(`{"age": 35, "date": "2024-09-08"}`))

	five := targets[3]
	Equal(t, five.URL.String(), "https://example.com/john/12345?id=abc&city=Paris")
	EqualBytes(t, five.Body, []byte(`{"age": FUZZ35NUM, "date": "2024-09-08"}`))

	six := targets[0]
	Equal(t, six.URL.String(), "https://example.com/john/12345?id=abc&city=Paris")
	EqualBytes(t, six.Body, []byte(`{"age": 35, "date": "FUZZ2024-09-08DTE"}`))
}
