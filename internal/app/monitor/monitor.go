package monitor

type ProductCreateMetric interface {
	Success(price int64)
	Failed(err error)
}

type ProductMetrics interface {
	Create() ProductCreateMetric
}
