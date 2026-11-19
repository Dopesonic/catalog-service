package mmetric

import (
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

func registerCollector(
	registerer prometheus.Registerer,
	collector prometheus.Collector,
) (prometheus.Collector, error) {
	err := registerer.Register(collector)
	if err == nil {
		return nil, nil
	}

	var alreadyRegistered prometheus.AlreadyRegisteredError
	if errors.As(err, &alreadyRegistered) {
		return alreadyRegistered.ExistingCollector, nil
	}

	return nil, err
}

func counter(reg prometheus.Registerer, name, help string) (prometheus.Counter, error) {
	collector := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
		Help:      help,
	})

	existing, err := registerCollector(reg, collector)
	if err != nil {
		return nil, err
	}

	if existing == nil {
		return collector, nil
	}

	registered, ok := existing.(prometheus.Counter)
	if !ok {
		return nil, fmt.Errorf(
			"metric %q already registered with incompatible collector type %T",
			name,
			existing,
		)
	}

	return registered, nil
}

func counterVec(
	reg prometheus.Registerer,
	name, help string,
	labels ...string,
) (*prometheus.CounterVec, error) {
	collector := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      name,
			Help:      help,
		},
		labels,
	)

	existing, err := registerCollector(reg, collector)
	if err != nil {
		return nil, err
	}

	if existing == nil {
		return collector, nil
	}

	registered, ok := existing.(*prometheus.CounterVec)
	if !ok {
		return nil, fmt.Errorf(
			"metric %q already registered with incompatible collector type %T",
			name,
			existing,
		)
	}

	return registered, nil
}

func histogramVec(
	reg prometheus.Registerer,
	name, help string,
	buckets []float64,
	labels ...string,
) (*prometheus.HistogramVec, error) {
	collector := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      name,
			Help:      help,
			Buckets:   buckets,
		},
		labels,
	)

	existing, err := registerCollector(reg, collector)
	if err != nil {
		return nil, err
	}

	if existing == nil {
		return collector, nil
	}

	registered, ok := existing.(*prometheus.HistogramVec)
	if !ok {
		return nil, fmt.Errorf(
			"metric %q already registered with incompatible collector type %T",
			name,
			existing,
		)
	}

	return registered, nil
}
