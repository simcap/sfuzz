package sfuzz

import (
	"maps"
	"net/http"
	"slices"
	"sync"
)

type Report struct {
	requests   []Request
	servers    []string
	RoundTrips map[int][]*http.Response
}

func NewReport(requests []Request) *Report {
	return &Report{requests: requests, RoundTrips: make(map[int][]*http.Response)}
}

func (r *Report) AddRoundTrip(resp *http.Response) {
	r.RoundTrips[resp.StatusCode] = append(r.RoundTrips[resp.StatusCode], resp)
}

func (r *Report) Servers() []string {
	r.computeServersOnce()()
	return r.servers
}

func (r *Report) computeServersOnce() func() {
	return sync.OnceFunc(func() {
		unique := make(map[string]struct{})
		for _, req := range r.requests {
			unique[req.URL.Host] = struct{}{}
		}
		r.servers = slices.Collect(maps.Keys(unique))
	})
}
