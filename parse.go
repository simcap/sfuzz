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
		if err = collectKeywords(&request); err != nil {
			return nil, fmt.Errorf("line %d: %w", count, err)
		}
		out = append(out, request)
	}
	return out, nil
}

func collectKeywords(r *FuzzRequest) error {
	pathKeywords, err := ParseKeywords(r.URL.Path)
	if err != nil {
		return err
	}
	r.URLKeywords = append(r.URLKeywords, pathKeywords...)

	queryKeywords, err := ParseKeywords(r.URL.RawQuery)
	if err != nil {
		return err
	}
	r.QueryKeywords = append(r.QueryKeywords, queryKeywords...)

	bodyKeywords, err := ParseKeywords(string(r.Body))
	if err != nil {
		return err
	}
	r.BodyKeywords = append(r.BodyKeywords, bodyKeywords...)

	return nil
}

func parseVerbs(s string) ([]string, error) {
	head, _, found := strings.Cut(s, " ")
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

func parseURLAndBody(s string, r *FuzzRequest) (err error) {
	uri, body, hasBody := strings.Cut(strings.TrimSpace(s), " ")
	r.URL, err = url.Parse(uri)
	if err != nil {
		return
	}

	if hasBody {
		r.Body, err = parseBody(body)
		if err != nil {
			return
		}
	}
	return
}

func parseBody(s string) (out []byte, err error) {
	switch {
	case strings.HasPrefix(s, "{"):
		out = []byte(s)
	case strings.HasPrefix(s, "@"):
		out, err = os.ReadFile(s[1:])
	default:
		return nil, fmt.Errorf("body must be JSON or a @filepath")
	}
	return out, nil
}
