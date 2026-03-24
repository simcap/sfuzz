package sfuzz_test

import (
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/simcap/sfuzz"
	"github.com/simcap/sfuzz/assert"
)

func TestParseFuzzKeywords(t *testing.T) {
	s := "https://example.com/customers/FUZZUID/order?age=FUZZ1234NUM&name=FUZZjohnSTR"

	requests, err := sfuzz.Parse(strings.NewReader(s))
	assert.Equal(t, err, nil)
	assert.Equal(t, len(requests), 1)
	keywords := requests[0].ParsedKeywords
	assert.Equal(t, len(keywords), 3)

	first := keywords[0]
	assert.Equal(t, first.Kind(), sfuzz.UniversalID)
	assert.Equal(t, first.Example(), "")

	second := keywords[1]
	assert.Equal(t, second.Kind(), sfuzz.Numeral)
	assert.Equal(t, second.Example(), "1234")

	third := keywords[2]
	assert.Equal(t, third.Kind(), sfuzz.GenericString)
	assert.Equal(t, third.Example(), "john")

	s = `{"date": "FUZZ2024-02-04DTE", "stamp": "FUZZTME", "time": "FUZZ12:34TME"}`
	keywords, err = sfuzz.ParseJSONBody([]byte(s))
	assert.Equal(t, err, nil)
	assert.Equal(t, len(keywords), 3)

	first = keywords[0]
	assert.Equal(t, first.Kind(), sfuzz.Date)
	assert.Equal(t, first.Example(), "2024-02-04")

	second = keywords[1]
	assert.Equal(t, second.Kind(), sfuzz.Time)
	assert.Equal(t, second.Example(), "")

	third = keywords[2]
	assert.Equal(t, third.Kind(), sfuzz.Time)
	assert.Equal(t, third.Example(), "12:34")
}

func TestParsePath(t *testing.T) {
	s := "https://example.com/customers/FUZZ456NUM/books/FUZZMobyDickSTR"
	u := mustParseURL(t, s)
	uu := mustParseURL(t, fmt.Sprintf("%s/", s))
	for _, uri := range []url.URL{u, uu} {
		keywords, err := sfuzz.ParsePath(uri)
		assert.Equal(t, err, nil)
		assert.Equal(t, len(keywords), 2)
		first, second := keywords[0], keywords[1]
		assert.Equal(t, first.Example(), "456")
		assert.Equal(t, strings.Contains(first.Ref(), "1"), true)
		assert.Equal(t, second.Example(), "MobyDick")
		assert.Equal(t, strings.Contains(second.Ref(), "3"), true)
	}
}

func TestQueryReplace(t *testing.T) {
	u := mustParseURL(t, "https://example.com?id=FUZZ123NUM")
	holder, err := sfuzz.ParseQuery(u)
	assert.Equal(t, err, nil)
	r := sfuzz.Request{URL: u}
	fuzzed, err := holder[0].Replace(r, "fuzzed")
	assert.Equal(t, err, nil)
	result := fuzzed.URL.String()
	assert.Equal(t, strings.Contains(result, "fuzzed"), true)
	assert.Equal(t, fuzzed.URL.Query().Get("id"), "fuzzed")
}

func TestQueryParsing(t *testing.T) {
	verify := func(t *testing.T, u url.URL) {
		t.Helper()
		holders, err := sfuzz.ParseQuery(u)
		assert.Equal(t, err, nil)
		assert.Equal(t, len(holders), 3)
		one, second, third := holders[0], holders[1], holders[2]
		assert.Equal(t, one.Ref(), "dob")
		assert.Equal(t, second.Ref(), "id")
		assert.Equal(t, third.Ref(), "name")
		assert.Equal(t, one.Example(), "2024-03-09")
		assert.Equal(t, second.Example(), "123")
		assert.Equal(t, third.Example(), "john")
	}

	u, err := url.Parse("https://localhost:8080?id=FUZZ123NUM&name=FUZZjohnSTR&dob=FUZZ2024-03-09DTE")
	assert.Equal(t, err, nil)
	verify(t, *u)

	t.Run("auto parsing", func(t *testing.T) {
		t.Skip()
		u, err = url.Parse("https://localhost:8080?id=123&name=john&dob=2024-03-09")
		assert.Equal(t, err, nil)
		verify(t, *u)
	})
}

func mustParseURL(t *testing.T, s string) url.URL {
	t.Helper()
	u, err := url.Parse(s)
	if err != nil {
		t.Fatal(err)
	}
	return *u
}
