package backoff

import (
	"math"
	"time"
)

// Strategy defines how to calculate the next delay for a retry.
type Strategy interface {
	Next(attempt int) time.Duration
}

// Exponential implements an exponential backoff strategy.
type Exponential struct {
	Base  time.Duration
	Max   time.Duration
	Limit int
}

// Next returns the duration to wait for the next attempt.
func (e *Exponential) Next(attempt int) time.Duration {
	if attempt <= 0 {
		return e.Base
	}
	
	// Calc: Base * 2^attempt
	exp := math.Pow(2, float64(attempt))
	delay := time.Duration(float64(e.Base) * exp)
	
	if e.Max > 0 && delay > e.Max {
		return e.Max
	}
	
	return delay
}

// NewExponential creates a default exponential backoff.
func NewExponential(base, max time.Duration) *Exponential {
	return &Exponential{
		Base: base,
		Max:  max,
	}
}
