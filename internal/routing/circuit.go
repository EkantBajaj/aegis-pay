package routing

import (
	"errors"
	"log"
	"time"

	"github.com/sony/gobreaker"
)

// ProviderBreaker wraps a payment provider with circuit breaking logic
type ProviderBreaker struct {
	cb *gobreaker.CircuitBreaker
}

// NewProviderBreaker initializes a new breaker for a specific provider
func NewProviderBreaker(name string) *ProviderBreaker {
	settings := gobreaker.Settings{
		Name:        name,
		MaxRequests: 1,               // Allow only 1 request in Half-Open state
		Timeout:     30 * time.Second, // Time to stay in Open state
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			log.Printf("CIRCUIT BREAKER: [%s] State changed from %v to %v", name, from, to)
		},
	}

	return &ProviderBreaker{
		cb: gobreaker.NewCircuitBreaker(settings),
	}
}

// Execute wraps a function call with the circuit breaker
func (p *ProviderBreaker) Execute(req func() (interface{}, error)) (interface{}, error) {
	result, err := p.cb.Execute(req)
	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) {
			return nil, errors.New("provider circuit is open (failing fast)")
		}
		return nil, err
	}
	return result, nil
}
