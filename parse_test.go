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
	filename := createFileWithContent(t, []byte(`{"town": "Paris"}`))

	inputs := fmt.Sprintf(`
POST|PUT https://example.com/customers/FUZZ_NUM?id=FUZZ_STR {"age": FUZZ_NUM, "name": "john"}
https://example.com/customers/FUZZ_NUM?id=FUZZ_UID @%s
`, filename)

	requests, err := sfuzz.Parse(strings.NewReader(inputs))
	Equal(t, err, nil)
	Equal(t, len(requests), 2)

	one := requests[0]
	Equal(t, one.Methods[0], "POST")
	Equal(t, one.Methods[1], "PUT")
	Equal(t, one.URL, "https://example.com/customers/FUZZ_NUM?id=FUZZ_STR")
	EqualBytes(t, one.Body, []byte(`{"age": FUZZ_NUM, "name": "john"}`))

	two := requests[1]
	Equal(t, two.Methods[0], "GET")
	Equal(t, two.URL, "https://example.com/customers/FUZZ_NUM?id=FUZZ_UID")
	EqualBytes(t, two.Body, []byte(`{"town": "Paris"}`))
}

func createFileWithContent(t *testing.T, data []byte) string {
	t.Helper()
	filename := filepath.Join(t.ArtifactDir(), "sfuzz-test.json")
	if err := os.WriteFile(filename, data, 0666); err != nil {
		t.Fatal(err)
	}
	return filename
}
