package costwatch

import (
	"math"
	"sync"
)

type serviceRegistry struct {
	mu   sync.RWMutex
	data map[string]Service // key: service label
}

var registry = &serviceRegistry{
	data: make(map[string]Service),
}

// RegisterService registers or replaces a service in the global registry.
func RegisterService(svc Service) {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	if svc == nil {
		return
	}

	registry.data[svc.Label()] = svc
}

// ListServices returns a snapshot of all registered services.
func ListServices() []Service {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	res := make([]Service, 0, len(registry.data))
	for _, s := range registry.data {
		res = append(res, s)
	}

	return res
}

// FindService returns a registered service by name.
func FindService(name string) (Service, bool) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	s, ok := registry.data[name]

	return s, ok
}

// FindMetric returns a metric by service+metric name if present.
func FindMetric(serviceName, metricName string) (Metric, bool) {
	s, ok := FindService(serviceName)
	if !ok {
		return nil, false
	}

	for _, m := range s.Metrics() {
		if m.Label() == metricName {
			return m, true
		}
	}

	return nil, false
}

// ComputeCost computes cost via Metric interface defined pricing.
func ComputeCost(service, metric string, units float64) (float64, bool) {
	m, ok := FindMetric(service, metric)
	if !ok {
		return 0, false
	}

	upp := m.UnitsPerPrice()
	if upp == 0 {
		return 0, false
	}

	res := math.Round((units/upp)*m.Price()) / 100

	return res, true
}
