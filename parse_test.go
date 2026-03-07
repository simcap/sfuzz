package sfuzz_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/simcap/sfuzz"
)

func TestParseFuzzRequest(t *testing.T) {
	filename := createFileWithContent(t, []byte(`{"town": "Paris", "code": "FUZZSTR"}`))

	inputs := fmt.Sprintf(`
POST https://example.com/customers/FUZZ1234NUM?id=FUZZSTR {"age": FUZZNUM, "name": "john"}
https://example.com/customers/123456?id=FUZZUID @%s
`, filename)

	requests, err := sfuzz.Parse(strings.NewReader(inputs))
	Equal(t, err, nil)
	Equal(t, len(requests), 2)

	one := requests[0]
	Equal(t, one.Verb, "POST")
	Equal(t, one.URL.String(), "https://example.com/customers/FUZZ1234NUM?id=FUZZSTR")
	EqualBytes(t, one.Body, []byte(`{"age": FUZZNUM, "name": "john"}`))
	Equal(t, len(one.ParsedKeywords), 3)
	Equal(t, one.ParsedKeywords[0].Location, sfuzz.PathKeyword)
	Equal(t, one.ParsedKeywords[1].Location, sfuzz.QueryKeyword)
	Equal(t, one.ParsedKeywords[2].Location, sfuzz.BodyKeyword)

	two := requests[1]
	Equal(t, two.Verb, "GET")
	Equal(t, two.URL.String(), "https://example.com/customers/123456?id=FUZZUID")
	EqualBytes(t, two.Body, []byte(`{"town": "Paris", "code": "FUZZSTR"}`))
	Equal(t, len(two.ParsedKeywords), 2)
	Equal(t, two.ParsedKeywords[0].Location, sfuzz.QueryKeyword)
	Equal(t, two.ParsedKeywords[1].Location, sfuzz.BodyKeyword)
}

func createFileWithContent(t *testing.T, data []byte) string {
	t.Helper()
	filename := filepath.Join(t.ArtifactDir(), "sfuzz-test.json")
	if err := os.WriteFile(filename, data, 0666); err != nil {
		t.Fatal(err)
	}
	return filename
}
