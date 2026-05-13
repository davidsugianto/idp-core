package redislock

import (
	"time"

	logs "github.com/davidsugianto/go-pkgs/logger"
)

// GetLock for the keys,
// error will be returned if fail to acquire the lock
func (m *mutex) GetLock() error {
	if m.mutex == nil {
		return ErrMutexEmpty
	}

	err := m.mutex.Lock()
	if err != nil {
		return err
	}

	m.acquired = true

	// create ticker for auto extend the lock time,
	// in case the process have not finished yet, but the lock time is over
	if m.AutoExtend {
		ticker := time.NewTicker(m.CheckTTLTime)
		m.extendRelease = make(chan bool)

		go func() {
			for {
				select {
				case <-m.extendRelease:
					return
				case <-ticker.C:
					var successExtend bool

					// try to extend
					// if failed, retry as many times as user set to maxRetryAutoExtend
					for i := 0; i < m.maxRetryAutoExtend; i++ {
						isSuccess, err := m.Extend()
						if err != nil && !isSuccess {
							logs.Error().Msgf("[Distlockredis][GetLock] Error when extend the lock %v", err)

							continue
						}

						successExtend = true

						// break if success extend
						break
					}

					if !successExtend {
						logs.Error().Msgf("[Distlockredis][GetLock] Fail extend the lock")

						m.extendRelease <- true
					}
				}
			}
		}()
	}

	return err
}

// Release the lock, to release the lock, need to acquire the lock first
func (m *mutex) Release() (bool, error) {
	if !m.acquired {
		return false, ErrFailReleaseNotAcquired
	}

	// stop the ticker for autoExtend
	if m.AutoExtend && m.extendRelease != nil {
		m.extendRelease <- true
	}

	m.acquired = false

	return m.mutex.Unlock()
}

// Extend the lock
func (m *mutex) Extend() (bool, error) {
	if !m.acquired {
		return false, ErrExtendNotAcquired
	}

	return m.mutex.Extend()
}
