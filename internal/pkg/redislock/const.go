package redislock

import (
	"fmt"
)

var (
	ErrClientEmpty            = fmt.Errorf("no mutex wrapper client found")
	ErrKeyEmpty               = fmt.Errorf("key empty")
	ErrFailReleaseNotAcquired = fmt.Errorf("release fail because never acquired the lock ")
	ErrExtendNotAcquired      = fmt.Errorf("Extend fail because never acquired the lock ")
	ErrMutexEmpty             = fmt.Errorf("mutex client empty")
)

const (
	DefaultLockTime           = 15
	DefaultRetryCount         = 1
	DefaultCheckTTLTime       = 1 // value in second
	DefaultMaxRetryAutoExtend = 5
)
