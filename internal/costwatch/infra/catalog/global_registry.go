package catalog

import "github.com/tailbits/costwatch/internal/costwatch"

type GlobalRegistryCatalog struct{}

func (GlobalRegistryCatalog) ComputeCost(service, metric string, units float64) (float64, bool) {
	return costwatch.ComputeCost(service, metric, units)
}
