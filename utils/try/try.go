package try

import "time"

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
		time.Sleep(delay)
	}
	return result, err
}
