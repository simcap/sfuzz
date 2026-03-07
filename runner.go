package sfuzz

import (
	"context"
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

func (r *runner) Run(ctx context.Context, targets []Target) {
	for _, t := range targets {
		generator := r.selector(t.Keyword)
		for val := range generator(t.Keyword.Example) {
			target, err := t.Replace(val)
			if err != nil {
				r.log.Error(err.Error(), "val", val)
				continue
			}
			resp, err := r.client.Do(target.ToHTTPRequest(ctx))
			if err != nil {
				r.log.Error(err.Error(), "target", target)
				continue
			}

			if err = resp.Body.Close(); err != nil {
				r.log.Error("cannot close body", "target", target, "err", err)
			}

			r.log.Info("called target", "code", resp.StatusCode,
				slog.Group("req", "method", resp.Request.Method, "path", resp.Request.URL.Path),
			)
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
