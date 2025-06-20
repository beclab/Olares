package state

import (
	"math"
	"math/rand"
	"time"
)

const (
	retryBaseDelay     = 5 * time.Second
	retryMaxDelay      = 10 * time.Minute
	retryBackoffFactor = 2.0
)

func calculateNextRetryDelay(retryNum int) time.Duration {
	backoffDelay := float64(retryBaseDelay) * math.Pow(retryBackoffFactor, float64(retryNum))

	if backoffDelay > float64(retryMaxDelay) {
		backoffDelay = float64(retryMaxDelay)
	}

	delay := time.Duration(backoffDelay)

	jitter := float64(delay) * 0.25 * (rand.Float64()*2 - 1)
	delay = time.Duration(float64(delay) + jitter)

	if delay < 0 {
		delay = retryBaseDelay
	}

	return delay
}

func calculateNextRetryTime(retryNum int) time.Time {
	delay := calculateNextRetryDelay(retryNum)
	return time.Now().Add(delay)
}

func CalculateNextRetryTime(retryNum int) time.Time {
	return calculateNextRetryTime(retryNum)
}
