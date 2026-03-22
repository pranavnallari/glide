package limiter

import (
	"testing"
	"time"
)

func TestBucketAllowsWhenTokensAvailable(t *testing.T) {
	b := newBucket(60)
	if !b.allow() {
		t.Fatal("expected allow when bucket is full")
	}
}

func TestBucketRejectsWhenEmpty(t *testing.T) {
	b := newBucket(1)
	b.allow() // drain the one token
	if b.allow() {
		t.Fatal("expected reject when bucket is empty")
	}
}

func TestBucketRefillsOverTime(t *testing.T) {
	b := newBucket(60) // 1 token per second
	b.tokens = 0
	b.lastRefill = time.Now().Add(-2 * time.Second) // fake 2 seconds passing
	if !b.allow() {
		t.Fatal("expected allow after refill time has passed")
	}
}

func TestNewBucketWithZeroRPMReturnsNil(t *testing.T) {
	b := newBucket(0)
	if b != nil {
		t.Fatal("expected nil bucket for zero rpm")
	}
}

func TestLimiterAllowsWhenNoBucketsConfigured(t *testing.T) {
	l := &Limiter{
		globalBucket:    nil,
		providerBuckets: make(map[string]*bucket),
		keyBuckets:      make(map[string]*bucket),
	}
	if !l.Allow("openai", "my-key") {
		t.Fatal("expected allow when no limits configured")
	}
}

func TestLimiterRejectsOnGlobalLimit(t *testing.T) {
	b := newBucket(1)
	b.allow() // drain it
	l := &Limiter{
		globalBucket:    b,
		providerBuckets: make(map[string]*bucket),
		keyBuckets:      make(map[string]*bucket),
	}
	if l.Allow("openai", "") {
		t.Fatal("expected reject when global bucket is empty")
	}
}

func TestLimiterRejectsOnProviderLimit(t *testing.T) {
	b := newBucket(1)
	b.allow() // drain it
	l := &Limiter{
		globalBucket:    nil,
		providerBuckets: map[string]*bucket{"openai": b},
		keyBuckets:      make(map[string]*bucket),
	}
	if l.Allow("openai", "") {
		t.Fatal("expected reject when provider bucket is empty")
	}
}
