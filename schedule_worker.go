package schedule_worker

import (
	"sync/atomic"
	"time"
)

type ScheduleWorker struct {
	Task    func() error
	OnError func(error)

	done            atomic.Bool
	until           time.Time
	approximation   time.Duration
	stopChan        chan bool
	immediatelyChan chan bool
}

// ExtNewScheduleWorker note: unlike NewScheduleWorker returned worker has to be started manually
func ExtNewScheduleWorker(task func() error, onError func(error), approximation time.Duration) *ScheduleWorker {
	worker := &ScheduleWorker{
		Task:            task,
		OnError:         onError,
		done:            atomic.Bool{},
		until:           time.Time{},
		approximation:   approximation,
		stopChan:        make(chan bool),
		immediatelyChan: make(chan bool),
	}
	return worker
}
func NewScheduleWorker(task func() error, onError ...func(error)) *ScheduleWorker {
	var errorHandler func(error)
	if len(onError) != 0 {
		errorHandler = onError[0]
	} else {
		errorHandler = func(error) {}
	}
	worker := ExtNewScheduleWorker(task, errorHandler, time.Second)
	go worker.Start()
	return worker
}

func (worker *ScheduleWorker) Until(t time.Time) *ScheduleWorker {
	if t.After(time.Now()) {
		worker.startNewScheduleRoutine(t)
	} else {
		worker.DoImmediately()
	}
	return worker
}

func (worker *ScheduleWorker) For(d time.Duration) *ScheduleWorker {
	return worker.Until(time.Now().Add(d))
}

func (worker *ScheduleWorker) Add(d time.Duration) *ScheduleWorker {
	if worker.until.Equal(time.Time{}) {
		return worker.Until(time.Now().Add(d))
	} else {
		return worker.Until(worker.until.Add(d))
	}
}

func (worker *ScheduleWorker) DoImmediately() {
	worker.immediatelyChan <- true
}

func (worker *ScheduleWorker) Cancel() {
	worker.stopChan <- true
}

func (worker *ScheduleWorker) GetTime() time.Time {
	return worker.until
}

func (worker *ScheduleWorker) IsDone() bool {
	return worker.done.Load()
}

func (worker *ScheduleWorker) Start() {
	for {
		select {
		case _ = <-worker.immediatelyChan:
			worker.OnError(worker.Task())
			worker.done.Store(true)
			return
		case _ = <-worker.stopChan:
			return
		}
	}
}

func (worker *ScheduleWorker) isApproximatelyEqualToSchedule(t time.Time) bool {
	return worker.until.Sub(t).Abs().Nanoseconds() <= worker.approximation.Nanoseconds()
}

func (worker *ScheduleWorker) startNewScheduleRoutine(t time.Time) {
	if !worker.isApproximatelyEqualToSchedule(t) {
		worker.until = t
		go func() {
			time.Sleep(time.Until(t))
			if worker.isApproximatelyEqualToSchedule(t) {
				worker.DoImmediately()
			}
		}()
		return
	}
}
