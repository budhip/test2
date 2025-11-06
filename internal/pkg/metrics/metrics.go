package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics interface {
	PrometheusRegisterer() prometheus.Registerer
}

type metrics struct {
	reg prometheus.Registerer
}

func New() Metrics {
	reg := prometheus.DefaultRegisterer

	return &metrics{
		reg: reg,
	}
}

func (m *metrics) PrometheusRegisterer() prometheus.Registerer {
	return m.reg
}
