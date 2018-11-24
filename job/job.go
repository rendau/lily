package job

import (
	"sync"
)

var (
	stop   bool
	stopMU sync.RWMutex
	wg     sync.WaitGroup
)

func Start() bool {
	stopMU.RLock()
	defer stopMU.RUnlock()

	if stop {
		return false
	}

	wg.Add(1)

	return true
}

func Finish() {
	wg.Done()
}

func GracefulStopAll() {
	stopMU.Lock()
	stop = true
	stopMU.Unlock()
	wg.Wait()
}
