package costwatch

import (
	"math"
	"sync"
)

type serviceRegistry struct {
	mu   sync.RWMutex
	data map[string]Service // key: service label
}

var globalSvcReg = &serviceRegistry{data: make(map[string]Service)}

// RegisterGlobalService registers or replaces a service in the global registry.
func RegisterGlobalService(svc Service) {
	if svc == nil {
		return
	}
	globalSvcReg.mu.Lock()
	defer globalSvcReg.mu.Unlock()
	globalSvcReg.data[svc.Label()] = svc
}

// FindService returns a registered service by name.
func FindService(name string) (Service, bool) {
	globalSvcReg.mu.RLock()
	defer globalSvcReg.mu.RUnlock()
	s, ok := globalSvcReg.data[name]
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
func ComputeCost(serviceName, metricName string, units float64) (float64, bool) {
	m, ok := FindMetric(serviceName, metricName)
	if !ok {
		return 0, false
	}
	upp := m.UnitsPerPrice()
	if upp == 0 {
		return 0, false
	}
	return math.Round((units/upp)*m.Price()*100) / 100, true
}
