package observability

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	RequestsTotal   *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
	TokensTotal     *prometheus.CounterVec
	CostTotal       *prometheus.CounterVec
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{}
	m.RequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "glide_requests_total",
			Help: "Total number of LLM requests routed",
		},
		[]string{"provider", "status"},
	)
	reg.MustRegister(m.RequestsTotal)

	m.RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "glide_requests_duration",
			Help:    "Duration of every LLM request",
			Buckets: prometheus.DefBuckets,
		},

		[]string{"provider"},
	)
	reg.MustRegister(m.RequestDuration)

	m.TokensTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "glide_tokens_total",
			Help: "Total number of tokens spent",
		},
		[]string{"provider", "token_type"},
	)
	reg.MustRegister(m.TokensTotal)

	m.CostTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "glide_cost_total",
			Help: "Total cost spent on tokens",
		},
		[]string{"provider"},
	)
	reg.MustRegister(m.CostTotal)
	return m
}
