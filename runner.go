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
	output   output
	selector Selector
}

func NewRunner(report *Report, opts ...option) *runner {
	r := &runner{
		report: report,
		log:    slog.New(slog.DiscardHandler),
		client: http.DefaultClient,
		output: noopOutput{},
		selector: func(k FuzzKeyword) Generator {
			switch k.Kind {
			case Numeral:
				return NumGenerator(k.Example)
			case UniversalID:
				return UIDGenerator(k.Example)
			case GenericString:
				return StrGenerator(k.Example)
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
	for _, request := range r.report.requests {
		candidates, err := request.BuildFuzzCandidates()
		if err != nil {
			r.log.Error(fmt.Sprintf("cannot build candidates from request: %v", err))
			return
		}

		for _, candidate := range candidates {
			generator := r.selector(candidate.Keyword)

			for val := range generator(candidate.Keyword.Example) {
				l := logWithTarget(r.log, candidate)

				target, err := candidate.Replace(val)
				if err != nil {
					r.report.AddError(err)
					l.Error(err.Error(), "val", val)
					continue
				}
				resp, err := r.client.Do(target.ToHTTPRequest(ctx))
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
		}
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
func WithOutput(o output) option {
	return func(r *runner) {
		r.output = o
	}
}
