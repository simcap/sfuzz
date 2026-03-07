package sfuzz

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
)

type runner struct {
	log      *slog.Logger
	client   *http.Client
	selector Selector
}

func NewRunner(opts ...option) *runner {
	r := &runner{
		log:      slog.New(slog.DiscardHandler),
		client:   http.DefaultClient,
		selector: func(FuzzKeyword) Generator { return NoopGenerator() },
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *runner) Run(ctx context.Context, requests []Request) {
	for _, request := range requests {
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
					l.Error(err.Error(), "val", val)
					continue
				}
				resp, err := r.client.Do(target.ToHTTPRequest(ctx))
				if err != nil {
					l.Error(err.Error())
					continue
				}

				l = logWithResponse(l, resp)
				if err = resp.Body.Close(); err != nil {
					l.Error("cannot close body", "err", err)
				}
				l.Info("called target")
			}
		}
	}
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
