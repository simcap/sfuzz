package sfuzz

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
)

type Output interface {
	LogRoundtrip(r *http.Response) error
	Name() string
}

func NewOutput() (Output, error) {
	dir, err := os.MkdirTemp(os.TempDir(), "sfuzz-*")
	if err != nil {
		return nil, err
	}
	return &output{dir: dir}, nil
}

func NoopOutput() (Output, error) {
	return noopOutput{}, nil
}

type output struct {
	dir string
}

func (o *output) Name() string { return o.dir }

func (o *output) LogRoundtrip(r *http.Response) error {
	var data bytes.Buffer

	reqData, err := httputil.DumpRequest(r.Request, true)
	if err != nil {
		return err
	}
	data.WriteString(">>> ")
	data.Write(bytes.TrimRight(reqData, "\r\n"))

	respData, err := httputil.DumpResponse(r, true)
	if err != nil {
		return err
	}
	data.WriteString("\n<<< ")
	data.Write(bytes.Trim(respData, "\r\n"))
	data.WriteString("\n\n")

	path := filepath.Join(o.dir, fmt.Sprintf("%d", r.StatusCode))
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(data.Bytes())
	return err
}

type noopOutput struct{}

func (noopOutput) LogRoundtrip(*http.Response) error { return nil }
func (noopOutput) Name() string                      { return "noop" }
