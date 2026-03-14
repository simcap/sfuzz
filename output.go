package sfuzz

import (
	"bytes"
	"embed"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"strings"
)

type output interface {
	Write(io.Writer) error
}

func NewFileOutput(r RunData) (output, error) {
	dir, err := os.MkdirTemp(os.TempDir(), "sfuzz-*")
	if err != nil {
		return nil, err
	}
	return &fileOutput{dir: dir}, nil
}

func NoopOutput() (output, error) { return noopOutput{}, nil }

type fileOutput struct {
	report RunData
	dir    string
}

func (o *fileOutput) Write(w io.Writer) error {
	for status, roundTrips := range o.report.RoundTripsPerStatus {
		path := filepath.Join(o.dir, fmt.Sprintf("%d", status))
		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}

		for _, roundTrip := range roundTrips {
			reqData, respData, err := getRequestAndResponseBytes(roundTrip.Resp)
			if err != nil {
				return errors.Join(err, f.Close())
			}
			var data bytes.Buffer
			data.WriteString(">>> ")
			data.Write(reqData)
			data.WriteString("\n<<< ")
			data.Write(respData)
			data.WriteString("\n\n")
			_, err = f.Write(data.Bytes())
		}

		if err = f.Close(); err != nil {
			return err
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
	"displayBody": displayResponseBody,
	"joinList":    joinList,
}

func joinList(all []string) string {
	return strings.Join(all, ", ")
}

func displayResponseBody(r *http.Response) string {
	defer r.Body.Close()

	var v map[string]any
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&v); err != nil {
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
