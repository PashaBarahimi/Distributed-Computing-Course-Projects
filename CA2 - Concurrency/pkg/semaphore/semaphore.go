package semaphore

type Semaphore struct {
	sem chan struct{}
}

func New(n int) *Semaphore {
	return &Semaphore{sem: make(chan struct{}, n)}
}

func (s *Semaphore) Acquire() {
	s.sem <- struct{}{}
}

func (s *Semaphore) Release() {
	<-s.sem
}

func (s *Semaphore) TryAcquire() bool {
	select {
	case s.sem <- struct{}{}:
		return true
	default:
		return false
	}
}
