package sfuzz

import (
	"context"
	"log/slog"
	"net/http"
)

type runner struct {
	log    *slog.Logger
	client *http.Client
}

func NewRunner(opts ...option) *runner {
	r := &runner{
		log:    slog.New(slog.DiscardHandler),
		client: http.DefaultClient,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *runner) Run(ctx context.Context, targets []Target) {
	for _, target := range targets {
		resp, err := r.client.Do(target.ToHTTPRequest(ctx))
		if err != nil {
			r.log.Error(err.Error(), "target", target)
		} else {
			if err = resp.Body.Close(); err != nil {
				r.log.Error("cannot close body", "target", target, "err", err)
			}
			r.log.Info("called target", "code", resp.StatusCode)
		}
	}
}

type option func(r *runner)

func WithLogger(l *slog.Logger) option {
	return func(r *runner) {
		r.log = l
	}
}
