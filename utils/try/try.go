package try

import (
	"time"

	"github.com/squarefactory/cloud-burster/logger"
	"go.uber.org/zap"
)

func Do[T interface{}](
	fn func() (T, error),
	tries int,
	delay time.Duration,
) (result T, err error) {
	for try := 0; try < tries; try++ {
		result, err = fn()
		if err == nil {
			break
		}
		logger.I.Warn("try failed", zap.Error(err), zap.Int("try", try))
		time.Sleep(delay)
	}
	return result, err
}
