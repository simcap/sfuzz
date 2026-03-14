package sfuzz

import (
	"cmp"
	"context"
	"fmt"

	"golang.org/x/time/rate"
)

type pubsub struct {
	limiter     *rate.Limiter
	subscribers map[string][]FuzzCandidate
	publishers  map[string]Iterator
}

func NewPubSub(frequency int) *pubsub {
	f := min(frequency, 1000)
	return &pubsub{
		limiter:     rate.NewLimiter(rate.Limit(f), 1),
		subscribers: make(map[string][]FuzzCandidate),
		publishers:  make(map[string]Iterator),
	}
}

func (p *pubsub) String() string {
	return fmt.Sprintf("Value producers (generators): %d, Candidate (subscribers): %d", len(p.subscribers), len(p.publishers))
}

func (p *pubsub) PublishLoop(ctx context.Context, out chan<- Target) error {
	defer close(out)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		targets, stop := p.publish(ctx)
		for _, target := range targets {
			if err := p.limiter.Wait(ctx); err != nil {
				return err
			}
			out <- target
		}
		if stop {
			return nil
		}
	}
}

func (p *pubsub) publish(ctx context.Context) ([]Target, bool) {
	var (
		status []bool
		out    []Target
	)
	for channel, pub := range p.publishers {
		val, more := pub.Next(ctx)
		status = append(status, more)
		if !more {
			continue
		}
		for _, sub := range p.subscribers[channel] {
			c, err := sub.Replace(val)
			if err != nil {
				out = append(out, Target{Value: val, Candidate: c, Err: err})
			} else {
				out = append(out, Target{Value: val, Candidate: c, R: c.ToHTTPRequest(ctx)})
			}
		}
	}

	return out, !cmp.Or(status...)
}

func (p *pubsub) AddSubscribers(all ...FuzzCandidate) {
	for _, c := range all {
		unique := c.Hash()
		p.subscribers[unique] = append(p.subscribers[unique], c)
	}
}

func (p *pubsub) AddPublisher(iter Iterator, candidate FuzzCandidate) {
	unique := candidate.Hash()
	if _, ok := p.publishers[unique]; !ok {
		p.publishers[unique] = iter
	}
}
