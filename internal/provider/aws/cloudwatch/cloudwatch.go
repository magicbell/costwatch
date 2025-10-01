package cloudwatch

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/tailbits/costwatch/internal/costwatch"
)

var _ costwatch.Service = (*Service)(nil)

type Service struct {
	mtrcs []costwatch.Metric
}

func NewService(cfg aws.Config) *Service {
	return &Service{}
}

// =============================================================================
func (s *Service) NewMetric(mtr costwatch.Metric) {
	s.mtrcs = append(s.mtrcs, mtr)
}

func (s *Service) Label() string {
	return "aws.CloudWatch"
}

func (s *Service) Metrics() []costwatch.Metric {
	return s.mtrcs
}
