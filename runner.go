package sfuzz

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type runner struct {
	report   *Report
	log      *slog.Logger
	client   *http.Client
	rps      uint
	selector Selector
}

func NewRunner(report *Report, opts ...option) *runner {
	r := &runner{
		report: report,
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

	for _, request := range r.report.requests {
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
			r.report.AddError(e)
			l.Error(e.Error(), "val", target.Value)
			continue
		}

		resp, err := r.client.Do(target.R)
		if err != nil {
			r.report.AddError(err)
			l.Error(err.Error())
			continue
		}
		r.report.FuzzedCount++
		r.report.AddRoundTrip(resp)
		l = logWithResponse(l, resp)
		l.Info("called target")
	}

	r.report.elapsed = time.Since(start)
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

type Target struct {
	Candidate FuzzCandidate
	R         *http.Request
	Value     any
	Err       error
}
