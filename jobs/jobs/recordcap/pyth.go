package recordcap

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type PythJob struct {
	repository *Repository
	logger     *zap.Logger
}

func NewPythJob(repository *Repository, logger *zap.Logger) *PythJob {
	return &PythJob{
		repository: repository,
		logger:     logger,
	}
}

func (j *PythJob) Run(ctx context.Context) error {
	maxTime := time.Now().AddDate(0, 0, -7)
	err := j.repository.DeletePyth(ctx, maxTime)
	if err != nil {
		j.logger.Error("error delete pyth record", zap.Error(err))
		return err
	}
	return nil
}
