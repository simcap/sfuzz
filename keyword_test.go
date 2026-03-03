package sfuzz_test

import (
	"testing"

	"github.com/simcap/sfuzz"
)

func TestParseFuzzKeywords(t *testing.T) {
	s := "https://example.com/customers/FUZZUID/order?age=FUZZ1234NUM&name=FUZZjohnSTR"

	keywords, err := sfuzz.ParseKeywords(s)
	Equal(t, err, nil)
	Equal(t, len(keywords), 3)

	first := keywords[0]
	Equal(t, first.Kind, sfuzz.UniversalID)
	Equal(t, first.Start, 30)
	Equal(t, first.End, 37)
	Equal(t, first.Spec.Example, "")

	second := keywords[1]
	Equal(t, second.Kind, sfuzz.Numeral)
	Equal(t, second.Start, 48)
	Equal(t, second.End, 59)
	Equal(t, second.Spec.Example, "1234")

	third := keywords[2]
	Equal(t, third.Kind, sfuzz.GenericString)
	Equal(t, third.Start, 65)
	Equal(t, third.End, 76)
	Equal(t, third.Spec.Example, "john")

	s = `{"date": FUZZ2024-02-04DTE, "stamp": FUZZTME", "time": FUZZ12:34TME}`
	keywords, err = sfuzz.ParseKeywords(s)
	Equal(t, err, nil)
	Equal(t, len(keywords), 3)

	first = keywords[0]
	Equal(t, first.Kind, sfuzz.Date)
	Equal(t, first.Start, 9)
	Equal(t, first.End, 26)
	Equal(t, first.Spec.Example, "2024-02-04")

	second = keywords[1]
	Equal(t, second.Kind, sfuzz.Time)
	Equal(t, second.Start, 37)
	Equal(t, second.End, 44)
	Equal(t, second.Spec.Example, "")

	third = keywords[2]
	Equal(t, third.Kind, sfuzz.Time)
	Equal(t, third.Start, 55)
	Equal(t, third.End, 67)
	Equal(t, third.Spec.Example, "12:34")
}
