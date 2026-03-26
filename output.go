package sfuzz

import (
	"bytes"
	"embed"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"
)

type output interface {
	Write(io.Writer) error
}

func NewJSONLOutput(r *RunData) output {
	return &jsonL{data: r}
}

type jsonL struct {
	data *RunData
}

func (o jsonL) Write(w io.Writer) error {
	for _, roundTrips := range o.data.RoundTripsPerStatus {
		for _, roundTrip := range roundTrips {
			content, err := json.MarshalIndent(roundTrip, " ", " ")
			if err != nil {
				return err
			}
			if _, err = fmt.Fprintln(w, string(content)); err != nil {
				return err
			}
		}
	}
	return nil
}

//go:embed html
var htmlFS embed.FS

func NewHTMLOutput(r *RunData) (output, error) {
	pages := template.Must(template.New("main").Funcs(funcs).ParseFS(htmlFS, "html/*.html"))
	return &htmlOutput{pages: pages, report: r}, nil
}

type htmlOutput struct {
	report *RunData
	pages  *template.Template
}

func (h *htmlOutput) Write(w io.Writer) error { return h.pages.Execute(w, h.report) }

type noopOutput struct{}

func (noopOutput) Write(io.Writer) error { return nil }

func getRequestAndResponseBytes(r *http.Response) ([]byte, []byte, error) {
	req, err := httputil.DumpRequest(r.Request, true)
	if err != nil {
		return nil, nil, err
	}
	resp, err := httputil.DumpResponse(r, true)
	if err != nil {
		return nil, nil, err
	}
	return bytes.Trim(req, "\r\n"), bytes.Trim(resp, "\r\n"), nil
}

var funcs = template.FuncMap{
	"displayRequestBody": displayRequestBody,
	"joinList":           joinList,
}

func joinList(all []string) string {
	return strings.Join(all, ", ")
}

func displayRequestBody(r *http.Request) string {
	body, err := r.GetBody()
	if err != nil {
		return err.Error()
	}

	var v map[string]any
	dec := json.NewDecoder(body)
	if err = dec.Decode(&v); err != nil {
		content, err := io.ReadAll(r.Body)
		if err != nil {
			return err.Error()
		}
		return string(content)
	}
	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	enc.SetIndent("", " ")
	_ = enc.Encode(&v)
	return out.String()
}
