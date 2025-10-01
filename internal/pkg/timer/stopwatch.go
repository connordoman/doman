package timer

import "time"

type Stopwatch struct {
	StartTime time.Time
	EndTime   time.Time
	Running   bool
}

func NewStopwatch(start bool) *Stopwatch {
	return &Stopwatch{
		StartTime: time.Now(),
		Running:   start,
	}
}

func (s *Stopwatch) Start() {
	s.StartTime = time.Now()
	s.Running = true
}

func (s *Stopwatch) Stop() {
	s.EndTime = time.Now()
	s.Running = false
}

func (s Stopwatch) Elapsed() time.Duration {
	if s.Running {
		return time.Since(s.StartTime)
	}

	return s.EndTime.Sub(s.StartTime)
}

func (s Stopwatch) String() string {
	return s.Elapsed().Round(time.Millisecond).String()
}
