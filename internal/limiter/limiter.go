package limiter

import (
	"sync"
	"time"

	"github.com/pranavnallari/glide/internal/config"
)

type bucket struct {
	tokens     float64    // current token count (float because refill is fractional)
	lastRefill time.Time  // when we last calculated refill
	rate       float64    // tokens per second
	maxTokens  float64    // bucket capacity
	mu         sync.Mutex // one lock per bucket
}

func (b *bucket) allow() bool {
	b.mu.Lock()

	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens = min(b.tokens+elapsed*b.rate, b.maxTokens)
	b.lastRefill = now

	if b.tokens >= 1 {
		b.tokens -= 1
		return true
	}
	return false
}

func newBucket(rpm int) *bucket {
	if rpm == 0 {
		return nil
	}
	return &bucket{
		tokens:     float64(rpm),
		rate:       float64(rpm) / 60,
		maxTokens:  float64(rpm),
		lastRefill: time.Now(),
	}
}

type Limiter struct {
	globalBucket    *bucket
	providerBuckets map[string]*bucket
	keyBuckets      map[string]*bucket
}

func NewLimiter(conf *config.LocalConfig) *Limiter {
	var limiter Limiter
	limiter.providerBuckets = make(map[string]*bucket)
	limiter.keyBuckets = make(map[string]*bucket)
	limiter.globalBucket = newBucket(conf.Limits.Global.RPM)
	for i, p := range conf.Limits.PerProvider {
		limiter.providerBuckets[i] = newBucket(p.RPM)
	}

	for i, p := range conf.Keys {
		limiter.keyBuckets[i] = newBucket(p.RPM)
	}
	return &limiter
}

func (l *Limiter) Allow(providerName, keyName string) bool {
	if l.globalBucket != nil && !l.globalBucket.allow() {
		return false
	}

	if (l.providerBuckets[providerName] != nil) && !l.providerBuckets[providerName].allow() {
		return false
	}

	if (l.keyBuckets[keyName] != nil) && (!l.keyBuckets[keyName].allow()) {
		return false
	}

	return true

}
