package sfuzz

import (
	"embed"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
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

var funcs = template.FuncMap{
	"joinList": joinList,
}

func joinList(all []string) string {
	return strings.Join(all, ", ")
}
