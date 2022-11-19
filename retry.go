package go_redislock

import "time"

// RetryStrategy defines retry strategy
type RetryStrategy interface {

	// Next
	//	first return means the interval of the next retry
	//	second return means need to retry or not
	Next() (time.Duration, bool)
}

// FixIntervalRetry is a fix interval retry strategy
type FixIntervalRetry struct {
	Interval time.Duration // retry interval
	Max      int           // max retry times
	count    int           // the number of retry
}

func (f *FixIntervalRetry) Next() (time.Duration, bool) {
	f.count++
	return f.Interval, f.count <= f.Max
}
