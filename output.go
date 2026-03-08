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

func NewFileOutput(r Report) (output, error) {
	dir, err := os.MkdirTemp(os.TempDir(), "sfuzz-*")
	if err != nil {
		return nil, err
	}
	return &fileOutput{dir: dir}, nil
}

func NoopOutput() (output, error) { return noopOutput{}, nil }

type fileOutput struct {
	report Report
	dir    string
}

func (o *fileOutput) Write(w io.Writer) error {
	for status, responses := range o.report.RoundTrips {
		path := filepath.Join(o.dir, fmt.Sprintf("%d", status))
		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}

		for _, resp := range responses {
			reqData, respData, err := getRequestAndResponseBytes(resp)
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

func NewHTMLOutput(r *Report) (output, error) {
	pages := template.Must(template.New("report").Funcs(funcs).ParseFS(htmlFS, "html/*.html"))
	return &htmlOutput{pages: pages, report: r}, nil
}

type htmlOutput struct {
	report *Report
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
