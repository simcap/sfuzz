package sfuzz_test

import (
	"testing"

	"github.com/simcap/sfuzz"
	"github.com/simcap/sfuzz/assert"
)

func TestParseFuzzKeywords(t *testing.T) {
	s := "https://example.com/customers/FUZZUID/order?age=FUZZ1234NUM&name=FUZZjohnSTR"

	keywords, err := sfuzz.ParseKeywords(s)
	assert.Equal(t, err, nil)
	assert.Equal(t, len(keywords), 3)

	first := keywords[0]
	assert.Equal(t, first.Kind, sfuzz.UniversalID)
	assert.Equal(t, first.Start, 30)
	assert.Equal(t, first.End, 37)
	assert.Equal(t, first.Example, "")

	second := keywords[1]
	assert.Equal(t, second.Kind, sfuzz.Numeral)
	assert.Equal(t, second.Start, 48)
	assert.Equal(t, second.End, 59)
	assert.Equal(t, second.Example, "1234")

	third := keywords[2]
	assert.Equal(t, third.Kind, sfuzz.GenericString)
	assert.Equal(t, third.Start, 65)
	assert.Equal(t, third.End, 76)
	assert.Equal(t, third.Example, "john")

	s = `{"date": FUZZ2024-02-04DTE, "stamp": FUZZTME", "time": FUZZ12:34TME}`
	keywords, err = sfuzz.ParseKeywords(s)
	assert.Equal(t, err, nil)
	assert.Equal(t, len(keywords), 3)

	first = keywords[0]
	assert.Equal(t, first.Kind, sfuzz.Date)
	assert.Equal(t, first.Start, 9)
	assert.Equal(t, first.End, 26)
	assert.Equal(t, first.Example, "2024-02-04")

	second = keywords[1]
	assert.Equal(t, second.Kind, sfuzz.Time)
	assert.Equal(t, second.Start, 37)
	assert.Equal(t, second.End, 44)
	assert.Equal(t, second.Example, "")

	third = keywords[2]
	assert.Equal(t, third.Kind, sfuzz.Time)
	assert.Equal(t, third.Start, 55)
	assert.Equal(t, third.End, 67)
	assert.Equal(t, third.Example, "12:34")
}
