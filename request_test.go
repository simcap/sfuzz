package sfuzz_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/simcap/sfuzz"
	"github.com/simcap/sfuzz/assert"
)

func TestBuildCandidates(t *testing.T) {
	s := `GET https://example.com/FUZZjohnSTR/FUZZ12345UID?id=FUZZabcSTR&city=FUZZParisSTR {"age": "FUZZ35NUM", "date": "FUZZ2024-09-08DTE"}`
	all, err := sfuzz.Parse(strings.NewReader(s))
	assert.Equal(t, err, nil)
	assert.Equal(t, len(all), 1)

	request := all[0]

	targets, err := request.BuildFuzzCandidates()
	assert.Equal(t, err, nil)
	assert.Equal(t, len(targets), 6)

	one := targets[0]
	assert.Equal(t, one.URL.String(), "https://example.com/FUZZjohnSTR/12345?city=Paris&id=abc")
	assert.EqualBytes(t, one.Body, []byte(`{"age":35,"date":"2024-09-08"}`))

	two := targets[1]
	assert.Equal(t, two.URL.String(), "https://example.com/john/FUZZ12345UID?city=Paris&id=abc")
	assert.EqualBytes(t, two.Body, []byte(`{"age":35,"date":"2024-09-08"}`))

	three := targets[3]
	assert.Equal(t, three.URL.String(), "https://example.com/john/12345?city=Paris&id=FUZZabcSTR")
	assert.EqualBytes(t, three.Body, []byte(`{"age":35,"date":"2024-09-08"}`))

	four := targets[2]
	assert.Equal(t, four.URL.String(), "https://example.com/john/12345?city=FUZZParisSTR&id=abc")
	assert.EqualBytes(t, four.Body, []byte(`{"age":35,"date":"2024-09-08"}`))

	five := targets[4]
	assert.Equal(t, five.URL.String(), "https://example.com/john/12345?city=Paris&id=abc")
	assert.EqualBytes(t, five.Body, []byte(`{"age":"FUZZ35NUM","date":"2024-09-08"}`))

	six := targets[5]
	assert.Equal(t, six.URL.String(), "https://example.com/john/12345?city=Paris&id=abc")
	assert.EqualBytes(t, six.Body, []byte(`{"age":35,"date":"FUZZ2024-09-08DTE"}`))
}

func TestBuildRequest(t *testing.T) {
	s := "POST https://example.com/john/71660f06-c8cd-4b8f-b169-ad9454c7d053?id=1234&city=Paris"
	requests, err := sfuzz.Parse(strings.NewReader(s))
	assert.Equal(t, err, nil)
	assert.Equal(t, len(requests), 1)
	request := requests[0]
	actual, err := request.AutoGenerateKeywords()
	assert.Equal(t, err, nil)
	assert.Equal(t, actual.URL.String(), "https://example.com/john/FUZZ71660f06-c8cd-4b8f-b169-ad9454c7d053UID?city=FUZZParisSTR&id=FUZZ1234NUM")
	assert.Equal(t, actual.Verb, http.MethodPost)
}
