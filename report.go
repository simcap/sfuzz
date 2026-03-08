package sfuzz

import (
	"fmt"
	"maps"
	"net/http"
	"slices"
	"sync"
	"time"
)

type Report struct {
	requests    []Request
	servers     []string
	elapsed     time.Duration
	FuzzedCount int
	Errors      []error
	RoundTrips  map[int][]*http.Response
}

func NewReport(requests []Request) *Report {
	return &Report{requests: requests, RoundTrips: make(map[int][]*http.Response)}
}

func (r *Report) AddRoundTrip(resp *http.Response) {
	r.RoundTrips[resp.StatusCode] = append(r.RoundTrips[resp.StatusCode], resp)
}
func (r *Report) AddError(err error) {
	if err != nil {
		r.Errors = append(r.Errors, err)
	}
}

func (r *Report) Statuses() []int    { return slices.Sorted(maps.Keys(r.RoundTrips)) }
func (r *Report) Duration() string   { return fmt.Sprintf("%s", r.elapsed) }
func (r *Report) RequestsCount() int { return len(r.requests) }
func (r *Report) KeywordsCount() (count int) {
	for _, req := range r.requests {
		count = count + len(req.ParsedKeywords)
	}
	return
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
