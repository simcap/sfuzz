package sfuzz

import (
	"cmp"
	"context"
	"fmt"
)

type pubsub struct {
	subscribers map[string][]FuzzCandidate
	publishers  map[string]Iterator
}

func NewPubSub() *pubsub {
	return &pubsub{
		subscribers: make(map[string][]FuzzCandidate),
		publishers:  make(map[string]Iterator),
	}
}

func (p *pubsub) String() string {
	return fmt.Sprintf("Value producers (generators): %d, Candidate (subscribers): %d", len(p.subscribers), len(p.publishers))
}

func (p *pubsub) Publish(ctx context.Context, out chan<- Target) bool {
	var status []bool
	for channel, pub := range p.publishers {
		val, more := pub.Next(ctx)
		status = append(status, more)
		if !more {
			continue
		}
		for _, sub := range p.subscribers[channel] {
			c, err := sub.Replace(val)
			if err != nil {
				out <- Target{Value: val, Candidate: c, Err: err}
			} else {
				out <- Target{Value: val, Candidate: c, R: c.ToHTTPRequest(ctx)}
			}
		}
	}

	return !cmp.Or(status...)
}

func (p *pubsub) AddSubscribers(all ...FuzzCandidate) {
	for _, c := range all {
		unique := c.Hash()
		p.subscribers[unique] = append(p.subscribers[unique], c)
	}
}

func (p *pubsub) AddPublisher(iter Iterator, hash string) {
	if _, ok := p.publishers[hash]; !ok {
		p.publishers[hash] = iter
	}
}
