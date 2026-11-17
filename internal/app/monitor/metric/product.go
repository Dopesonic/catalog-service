package mmetric

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/Dopesonic/catalog-service/internal/app/monitor"
)

const (
	namespace = "catalog"
	subsystem = "product"

	labelResult = "result"
	labelReason = "reason"

	createResultSuccess = "success"
	createResultError   = "error"
	createReasonNone    = "none"

	createTotalName = "create_total"
	createTotalHelp = "Total completed product creation attempts."

	createDurationName = "create_duration_seconds"
	createDurationHelp = "Product creation duration in seconds."

	createdTotalName = "created_total"
	createdTotalHelp = "Total successfully created products."

	createdAmountName = "created_amount_cents_total"
	createdAmountHelp = "Total price of successfully created products in cents."
)

type Product struct {
	createTotal           *prometheus.CounterVec
	createDurationSeconds *prometheus.HistogramVec
	createdTotal          prometheus.Counter
	createdAmountTotal    prometheus.Counter
}

var _ monitor.ProductMetrics = (*Product)(nil)

func NewPrometheusProduct(registerer prometheus.Registerer) (*Product, error) {
	if registerer == nil {
		registerer = prometheus.DefaultRegisterer
	}

	total, err := counterVec(
		registerer,
		createTotalName,
		createTotalHelp,
		labelResult,
		labelReason,
	)
	if err != nil {
		return nil, fmt.Errorf("register %s: %w", createTotalName, err)
	}

	duration, err := histogramVec(
		registerer,
		createDurationName,
		createDurationHelp,
		prometheus.DefBuckets,
		labelResult,
		labelReason,
	)
	if err != nil {
		return nil, fmt.Errorf("register %s: %w", createDurationName, err)
	}

	created, err := counter(
		registerer,
		createdTotalName,
		createdTotalHelp,
	)
	if err != nil {
		return nil, fmt.Errorf("register %s: %w", createdTotalName, err)
	}

	amount, err := counter(
		registerer,
		createdAmountName,
		createdAmountHelp,
	)
	if err != nil {
		return nil, fmt.Errorf("register %s: %w", createdAmountName, err)
	}

	return &Product{
		createTotal:           total,
		createDurationSeconds: duration,
		createdTotal:          created,
		createdAmountTotal:    amount,
	}, nil
}

type productCreateMetric struct {
	parent  *Product
	started time.Time
}

func (m *Product) Create() monitor.ProductCreateMetric {
	return &productCreateMetric{
		parent:  m,
		started: time.Now(),
	}
}

func (m *productCreateMetric) Success(price int64) {
	m.parent.createTotal.WithLabelValues(createResultSuccess, createReasonNone).Inc()

	m.parent.createDurationSeconds.WithLabelValues(
		createResultSuccess,
		createReasonNone,
	).Observe(time.Since(m.started).Seconds())

	m.parent.createdTotal.Inc()

	if price > 0 {
		m.parent.createdAmountTotal.Add(float64(price))
	}
}

func (m *productCreateMetric) Failed(err error) {
	reason := createReason(err)

	m.parent.createTotal.WithLabelValues(createResultError, reason).Inc()

	m.parent.createDurationSeconds.WithLabelValues(
		createResultError,
		reason,
	).Observe(time.Since(m.started).Seconds())
}
