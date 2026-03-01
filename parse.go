package sfuzz

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// Parse parses a given file (or io.Reader) containing fuzz requests on each line.
//
// A fuzz request is a one-liner from a file with the following definition:
// [GET|POST|PUT|DELETE] URL [JSON_BODY|@FILENAME_WITH_BODY]
//
// Example:
// POST https://example.com/customers/FUZZ_NUM?id=FUZZ_STR {"age": FUZZ_NUM, "name": "john"}
func Parse(input io.Reader) (out []FuzzRequest, err error) {
	var count int
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		count++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var request FuzzRequest
		request.Methods, err = parseVerbs(line)

		index := strings.Index(line, "http")
		if index < 0 {
			return nil, fmt.Errorf("line %d: no http keyword found anywhere", count)
		}

		if err = parseURLAndBody(line[index:], &request); err != nil {
			return nil, fmt.Errorf("line %d: %w", count, err)
		}
		out = append(out, request)
	}
	return out, nil
}

func parseVerbs(line string) ([]string, error) {
	head, _, found := strings.Cut(line, " ")
	if !found {
		return nil, errors.New("no space separator found in line")
	}
	if strings.HasPrefix(head, "http") {
		return []string{http.MethodGet}, nil
	}
	if strings.Contains(head, "|") {
		return strings.Split(head, "|"), nil
	}
	return []string{head}, nil
}

func parseURLAndBody(line string, r *FuzzRequest) error {
	uri, body, hasBody := strings.Cut(strings.TrimSpace(line), " ")
	parsed, err := url.Parse(uri)
	if err != nil {
		return err
	}
	r.URL = parsed.String()

	if hasBody {
		r.Body, err = parseBody(body)
		if err != nil {
			return err
		}
	}
	return nil
}

func parseBody(text string) (out []byte, err error) {
	switch {
	case strings.HasPrefix(text, "{"):
		out = []byte(text)
	case strings.HasPrefix(text, "@"):
		out, err = os.ReadFile(text[1:])
	default:
		return nil, fmt.Errorf("body must be JSON or a @filepath")
	}
	return out, nil
}
