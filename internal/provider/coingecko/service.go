package coingecko

import "github.com/costwatchai/costwatch/internal/costwatch"

// Ensure Service implements costwatch.Service
var _ costwatch.Service = (*Service)(nil)

type Service struct {
	mtrcs []costwatch.Metric
}

func NewService() *Service { return &Service{} }

func (s *Service) Label() string { return "coingecko" }

func (s *Service) Metrics() []costwatch.Metric { return s.mtrcs }

func (s *Service) NewMetric(mtr costwatch.Metric) { s.mtrcs = append(s.mtrcs, mtr) }
