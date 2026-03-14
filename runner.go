package sfuzz

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"net/http"
	"slices"
	"sync"
	"time"
)

type runner struct {
	data     *RunData
	log      *slog.Logger
	client   *http.Client
	rps      uint
	selector Selector
}

func NewRunner(requests []Request, opts ...option) *runner {
	r := &runner{
		data:   newRunData(requests),
		log:    slog.New(slog.DiscardHandler),
		client: http.DefaultClient,
		rps:    100,
		selector: func(k FuzzKeyword) Generator {
			switch k.Kind {
			case Numeral:
				return NumGenerator()
			case UniversalID:
				return UIDGenerator()
			case GenericString:
				return StrGenerator()
			}
			return NoopGenerator()
		},
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *runner) Run(ctx context.Context) {
	start := time.Now()
	ps := NewPubSub(int(r.rps))

	for _, request := range r.data.requests {
		candidates, err := request.BuildFuzzCandidates()
		if err != nil {
			r.log.Error(fmt.Sprintf("cannot build candidates from request: %v", err))
			return
		}

		for _, candidate := range candidates {
			ps.AddSubscribers(candidate)
			generator := r.selector(candidate.Keyword)
			ps.AddPublisher(NewIterator(generator), candidate)
		}
	}

	r.log.Info(fmt.Sprintf("pubsub: %s", ps.String()))

	targets := make(chan Target)

	go func() {
		if err := ps.PublishLoop(ctx, targets); err != nil {
			r.log.Error(fmt.Sprintf("cannot publish to pubsub: %v", err))
		}
	}()

	for target := range targets {
		l := logWithTarget(r.log, target.Candidate)

		if e := target.Err; e != nil {
			r.data.AddError(e)
			l.Error(e.Error(), "val", target.Value)
			continue
		}

		resp, err := r.client.Do(target.R)
		if err != nil {
			r.data.AddError(err)
			l.Error(err.Error())
			continue
		}
		r.data.FuzzedCount++
		r.data.AddRoundTrip(RoundTrip{Resp: resp, Target: target})
		l = logWithResponse(l, resp)
		l.Info("called target")
	}

	r.data.elapsed = time.Since(start)
}

func (r *runner) Results() *RunData { return r.data }

type Target struct {
	Candidate FuzzCandidate
	R         *http.Request
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
func (r RoundTrip) Keyword() string   { return r.Target.Candidate.Keyword.String() }
func (r RoundTrip) Error() error      { return r.Target.Err }

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

func (r *RunData) Statuses() []int    { return slices.Sorted(maps.Keys(r.RoundTripsPerStatus)) }
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

type option func(r *runner)

func WithLogger(l *slog.Logger) option {
	return func(r *runner) {
		r.log = l
	}
}
func WithSelector(s Selector) option {
	return func(r *runner) {
		r.selector = s
	}
}
func WithMaxRPS(rps uint) option {
	return func(r *runner) {
		r.rps = rps
	}
}
