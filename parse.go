package sfuzz

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
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
// A fuzz request is a one-liner respecting the syntax:
// [GET|POST|PUT|DELETE] URL [JSON_BODY|@FILENAME_WITH_BODY]
//
// Example:
// POST https://example.com/customers/FUZZNUM?id=FUZZSTR {"age": "FUZZNUM", "name": "john"}
func Parse(input io.Reader) (out []Request, err error) {
	var count int
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		count++
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		var request Request
		request.Verb, err = parseVerb(line)

		index := strings.Index(line, "http")
		if index < 0 {
			return nil, fmt.Errorf("line %d: no http prefix found", count)
		}

		if err = parseURLAndBody(line[index:], &request); err != nil {
			return nil, fmt.Errorf("line %d: %w", count, err)
		}
		hasher := sha256.New()
		hasher.Write([]byte(request.Verb))
		hasher.Write([]byte(fmt.Sprintf("%s%s%s", request.URL.Scheme, request.URL.Host, request.URL.Path)))
		request.Signature = hex.EncodeToString(hasher.Sum(nil))

		if err = collectKeywords(&request); err != nil {
			return nil, fmt.Errorf("line %d: %w", count, err)
		}
		out = append(out, request)
	}
	return
}

func collectKeywords(r *Request) error {
	pathKeywords, err := ParsePath(r.URL)
	if err != nil {
		return err
	}
	r.ParsedKeywords = append(r.ParsedKeywords, sortKeywords(pathKeywords)...)

	queryKeywords, err := ParseQuery(r.URL)
	if err != nil {
		return err
	}
	r.ParsedKeywords = append(r.ParsedKeywords, sortKeywords(queryKeywords)...)

	bodyKeywords, err := ParseJSONBody(r.Body)
	if err != nil {
		return err
	}
	r.ParsedKeywords = append(r.ParsedKeywords, sortKeywords(bodyKeywords)...)

	return nil
}

func parseVerb(s string) (string, error) {
	head, _, found := strings.Cut(s, " ")
	if !found {
		return "", errors.New("no space separator found in line")
	}
	if strings.HasPrefix(head, "http") {
		return http.MethodGet, nil
	}
	return head, nil
}

func parseURLAndBody(s string, r *Request) (err error) {
	uri, body, hasBody := strings.Cut(strings.TrimSpace(s), " ")
	parsed, err := url.Parse(uri)
	if err != nil {
		return
	}
	r.URL = *parsed

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
