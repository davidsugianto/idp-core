package redislock

import (
	"time"

	"github.com/go-redsync/redsync/v4"
)

// IMutexDistLock contains interface of distlock
type IMutexDistLock interface {
	GetLock() error
	Release() (bool, error)
	Extend() (bool, error)
}

// MutexOpt is config for initialize mutex
type MutexOpt struct {
	// Redis Key
	Key string
	// Redis Expiry Time (TTL), default 15s
	LockTime time.Duration
	// Extra time for Expiry
	WaitTimeExtra time.Duration
	// Number Of retry when failed to acquire the lock
	RetryCount int
	// Automatic Extend Expiry Time if there is incomplete process
	AutoExtend bool
	// Time to check the TTL periodically, and will extend automatically if auto extend set to true.
	CheckTTLTime time.Duration
	// max retry if fail auto extend, default 5
	MaxRetryAutoExtend int
}

type mutex struct {
	mutex              *redsync.Mutex
	acquired           bool
	key                string
	AutoExtend         bool
	CheckTTLTime       time.Duration
	extendRelease      chan bool
	maxRetryAutoExtend int
}

func (rw *redsyncWrap) NewMutexW(mutexOpt MutexOpt) (IMutexDistLock, error) {
	if rw.rs == nil {
		return nil, ErrClientEmpty
	}

	err := validateRequest(&mutexOpt)
	if err != nil {
		return nil, err
	}

	redMutex := rw.rs.NewMutex(mutexOpt.Key,
		redsync.WithExpiry(mutexOpt.LockTime+mutexOpt.WaitTimeExtra),
		redsync.WithTries(mutexOpt.RetryCount))

	return &mutex{
		mutex:              redMutex,
		key:                mutexOpt.Key,
		AutoExtend:         mutexOpt.AutoExtend,
		CheckTTLTime:       mutexOpt.CheckTTLTime,
		maxRetryAutoExtend: mutexOpt.MaxRetryAutoExtend,
	}, nil
}

func validateRequest(mutexOpt *MutexOpt) error {
	if mutexOpt.Key == "" {
		return ErrKeyEmpty
	}

	// add default value for lock time
	if mutexOpt.LockTime == 0 {
		mutexOpt.LockTime = DefaultLockTime * time.Second
	}

	// add default value for retry count
	if mutexOpt.RetryCount == 0 {
		mutexOpt.RetryCount = DefaultRetryCount
	}

	// add default value for check ttl time
	if mutexOpt.CheckTTLTime == 0 {
		mutexOpt.CheckTTLTime = DefaultCheckTTLTime * time.Second
	}

	// add default max retry count
	if mutexOpt.MaxRetryAutoExtend == 0 {
		mutexOpt.MaxRetryAutoExtend = DefaultMaxRetryAutoExtend
	}

	return nil
}
