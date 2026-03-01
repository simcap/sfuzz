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

func (r *runner) Run(ctx context.Context, requests []Target) {
	for _, req := range requests {
		resp, err := r.client.Do(req.ToHTTPRequest(ctx))
		if err != nil {
			r.log.Error(err.Error(), "request", req)
		} else {
			resp.Body.Close()
			r.log.Info("call done", "code", resp.StatusCode)
		}
	}
}

type option func(r *runner)

func WithLogger(l *slog.Logger) option {
	return func(r *runner) {
		r.log = l
	}
}
