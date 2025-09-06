package port

type Catalog interface {
	ComputeCost(service, metric string, units float64) (float64, bool)
}
