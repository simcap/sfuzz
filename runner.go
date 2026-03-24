package sfuzz

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
	"time"
)

type runner struct {
	data              *RunData
	log               *slog.Logger
	client            *http.Client
	headers           http.Header
	wordlists         map[Kind][]string
	selector          Selector
	rps               uint
	showProgress      bool
	jsonlOutputWriter io.Writer
}

func NewRunner(requests []Request, opts ...Option) *runner {
	r := &runner{
		data:      newRunData(requests),
		log:       slog.New(slog.DiscardHandler),
		client:    http.DefaultClient,
		wordlists: make(map[Kind][]string),
		rps:       100,
		headers:   make(http.Header),
		selector:  DefaultSelector,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *runner) Run(ctx context.Context) {
	ps, err := r.loadPubsub()
	if err != nil {
		r.log.Error(fmt.Sprintf("cannot load pubsub: %v", err))
		return
	}

	targets := make(chan Target)
	go func() {
		if err = ps.PublishLoop(ctx, targets); err != nil {
			r.log.Error(fmt.Sprintf("cannot publish to pubsub: %v", err))
		}
	}()

	start := time.Now()
	for target := range targets {
		l := logWithTarget(r.log, target.Candidate)

		if e := target.Err; e != nil {
			r.data.AddError(e)
			l.Error(e.Error(), "val", target.Value)
			continue
		}

		resp, err := r.client.Do(target.Candidate.ToHTTPRequest(ctx, r.headers))
		if err != nil {
			r.data.AddError(err)
			l.Error(err.Error())
			continue
		}
		r.data.FuzzedCount++
		r.data.AddRoundTrip(RoundTrip{Resp: resp, Target: target})
		l = logWithResponse(l, resp)
		l.Info("called target")

		if r.showProgress {
			progress(time.Since(start).Seconds(), r.data.FuzzedCount, len(r.data.Errors), r.data.StatusesCount())
		}
	}

	r.data.elapsed = time.Since(start)
}

func (r *runner) loadPubsub() (*pubsub, error) {
	ps := NewPubSub(r.rps)
	for _, request := range r.data.requests {
		candidates, err := request.BuildFuzzCandidates()
		if err != nil {
			return ps, fmt.Errorf("cannot build candidates from request: %v", err)
		}
		for _, candidate := range candidates {
			ps.AddSubscribers(candidate)
			fuzzer := r.selector(candidate.Keyword)
			ps.AddPublisher(fuzzer, candidate)
		}
	}
	return ps, nil
}

func (r *runner) Results() *RunData { return r.data }

func progress(data ...any) {
	fmt.Fprintf(os.Stdout, "%.2fs, fuzzed: %3d, errors: %3d, statuses: %s\r", data...)
}

type Target struct {
	Candidate FuzzCandidate
	Value     any
	Err       error
}

type RoundTrip struct {
	Resp   *http.Response
	Target Target
}

func (r RoundTrip) Status() int {
	if r.Resp != nil {
		return r.Resp.StatusCode
	}
	return http.StatusTeapot
}
func (r RoundTrip) FuzzValue() string { return fmt.Sprintf("%v", r.Target.Value) }
func (r RoundTrip) Keyword() string   { return "TODO" }
func (r RoundTrip) Error() error      { return r.Target.Err }

func (r RoundTrip) MarshalJSON() ([]byte, error) {
	type out struct {
		Status   int    `json:"status"`
		Request  string `json:"request"`
		Location string `json:"location"`
		Ref      string `json:"ref"`
		Value    string `json:"value"`
	}

	var o out
	o.Status = r.Status()
	if r.Resp != nil {
		o.Request = fmt.Sprintf("%s %s", r.Resp.Request.Method, r.Resp.Request.URL)
	}
	switch r.Target.Candidate.Keyword.(type) {
	case QueryKeyword:
		o.Location = "query"
	case PathKeyword:
		o.Location = "path"
	case JSONBodyKeyword:
		o.Location = "body"
	}
	o.Ref = r.Target.Candidate.Keyword.Ref()
	o.Value = r.FuzzValue()
	return json.Marshal(o)
}

type RunData struct {
	requests            []Request
	servers             []string
	elapsed             time.Duration
	FuzzedCount         int
	Errors              []error
	RoundTripsPerStatus map[int][]RoundTrip
}

func newRunData(requests []Request) *RunData {
	return &RunData{requests: requests, RoundTripsPerStatus: make(map[int][]RoundTrip)}
}

func (r *RunData) AddRoundTrip(t RoundTrip) {
	r.RoundTripsPerStatus[t.Status()] = append(r.RoundTripsPerStatus[t.Status()], t)
}
func (r *RunData) AddError(err error) {
	if err != nil {
		r.Errors = append(r.Errors, err)
	}
}

func (r *RunData) Statuses() []int { return slices.Sorted(maps.Keys(r.RoundTripsPerStatus)) }
func (r *RunData) StatusesCount() string {
	var out []string
	for status, all := range r.RoundTripsPerStatus {
		out = append(out, fmt.Sprintf("%d:%d", status, len(all)))
	}
	slices.Sort(out)
	return fmt.Sprintf("[%s]", strings.Join(out, " "))
}
func (r *RunData) Duration() string   { return fmt.Sprintf("%s", r.elapsed) }
func (r *RunData) RequestsCount() int { return len(r.requests) }
func (r *RunData) KeywordsCount() (count int) {
	for _, req := range r.requests {
		count = count + len(req.ParsedKeywords)
	}
	return
}

func (r *RunData) Servers() []string {
	r.computeServersOnce()()
	return r.servers
}

func (r *RunData) computeServersOnce() func() {
	return sync.OnceFunc(func() {
		unique := make(map[string]struct{})
		for _, req := range r.requests {
			unique[req.URL.Host] = struct{}{}
		}
		r.servers = slices.Collect(maps.Keys(unique))
	})
}

type Option func(r *runner)

func WithLogger(l *slog.Logger) Option {
	return func(r *runner) {
		if r.showProgress {
			r.log = discardLogger
		} else {
			r.log = l
		}
	}
}

var discardLogger = slog.New(slog.DiscardHandler)

func WithProgress(show bool) Option {
	return func(r *runner) {
		r.showProgress = show
		if show {
			r.log = discardLogger
		}
	}
}
func WithSelector(s Selector) Option {
	return func(r *runner) {
		r.selector = s
	}
}
func WithMaxRPS(rps uint) Option {
	return func(r *runner) {
		r.rps = rps
	}
}
func WithWordlist(filenames map[Kind]string) (Option, error) {
	if filenames == nil || len(filenames) == 0 {
		return func(*runner) {}, nil
	}

	wordlists := make(map[Kind][]any)
	for kind, filename := range filenames {
		if len(filename) > 0 {
			values, err := parseWordlistFile(filename)
			if err != nil {
				return nil, err
			}
			if len(values) > 0 {
				wordlists[kind] = values
			}
		}
	}
	if len(wordlists) == 0 {
		return func(*runner) {}, nil
	}
	return func(r *runner) {
		r.selector = WordlistSelector(wordlists)
	}, nil
}

func WithHeaders(headers http.Header) Option {
	return func(r *runner) {
		if headers == nil {
			r.headers = headers
		}
	}
}

func parseWordlistFile(filename string) ([]any, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var out []any
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.HasPrefix(text, "#") {
			continue
		}
		out = append(out, text)
	}
	return out, scanner.Err()
}
