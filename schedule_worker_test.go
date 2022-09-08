package schedule_worker

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func createDumbUtilities() (func() error, func(err error), *atomic.Bool) {
	dumbErr := errors.New("dumb error")
	isDumbErrHandled := &atomic.Bool{}
	dumbErrHandler := func(err error) {
		if err.Error() == dumbErr.Error() {
			isDumbErrHandled.Store(true)
		}
	}
	dumbFuncErr := func() error {
		return dumbErr
	}
	return dumbFuncErr, dumbErrHandler, isDumbErrHandled
}

func TestNewScheduleWorker(t *testing.T) {
	t.Parallel()

	t.Run("without handler", func(t *testing.T) {
		t.Parallel()

		fun, handler, hasRun := createDumbUtilities()

		NewScheduleWorker(func() error { handler(fun()); return nil }).For(time.Second)
		time.Sleep(time.Second * 5)
		if !hasRun.Load() {
			t.Fail()
		}
	})

	t.Run("with handler", func(t *testing.T) {
		t.Parallel()

		fun, handler, hasRun := createDumbUtilities()

		NewScheduleWorker(fun, handler).For(time.Second)
		time.Sleep(time.Second * 5)
		if !hasRun.Load() {
			t.Fail()
		}
	})
}

func TestExtNewScheduleWorker(t *testing.T) {
	t.Parallel()
	t.Run("custom approximation", func(t *testing.T) {
		t.Parallel()
		fun, handler, hasRun := createDumbUtilities()
		w := ExtNewScheduleWorker(fun, handler, time.Second*8)
		go w.Start()
		//log.Println(w.GetTime())
		w.For(time.Second * 6)
		//log.Println("6", w.GetTime())
		w.For(time.Second)
		//log.Println(w.GetTime())
		time.Sleep(time.Second * 2)
		//log.Println(w.GetTime())
		if hasRun.Load() {
			//log.Println(w.GetTime())
			t.Error("ran tho shouldn't")
		}
		time.Sleep(time.Second * 10)
		if !hasRun.Load() || !w.IsDone() {
			t.Error("didn't ran at all :c")
		}
	})
}

func TestScheduleWorker_For(t *testing.T) {
	t.Parallel()

	t.Run("positive", func(t *testing.T) {
		t.Parallel()

		fun, handler, hasRun := createDumbUtilities()

		w := NewScheduleWorker(fun, handler).For(time.Second * 5)
		time.Sleep(time.Second * 3)
		if hasRun.Load() || w.IsDone() {
			t.Fail()
		}
		w.For(time.Second * 5)
		time.Sleep(time.Second * 4)
		if hasRun.Load() || w.IsDone() {
			t.Fail()
		}
		time.Sleep(time.Second * 2)
		if !hasRun.Load() || !w.IsDone() {
			t.Fail()
		}
	})

	t.Run("negative", func(t *testing.T) {
		t.Parallel()

		fun, handler, hasRun := createDumbUtilities()

		NewScheduleWorker(fun, handler).For(-time.Second * 5)
		time.Sleep(time.Second * 1)
		if !hasRun.Load() {
			t.Fail()
		}
	})
}

func TestScheduleWorker_Add(t *testing.T) {
	t.Parallel()

	fun, handler, hasRun := createDumbUtilities()

	w := NewScheduleWorker(fun, handler)
	//log.Println(w.GetTime().String())
	for i := 0; i < 20; i++ {
		w.Add(time.Millisecond * 500)
		//log.Println(w.GetTime().String())
	}
	time.Sleep(time.Second * 3)
	if hasRun.Load() || w.IsDone() {
		t.Error("did ran, tho hadn't")
	}
	time.Sleep(time.Second * 8)
	if !hasRun.Load() || !w.IsDone() {
		t.Error("didn't run at all :c")
	}
}

func TestScheduleWorker_Until(t *testing.T) {
	t.Parallel()

	fun, handler, hasRun := createDumbUtilities()

	w := NewScheduleWorker(fun, handler).Until(time.Now().Add(time.Second * 5))
	time.Sleep(time.Second * 3)
	if hasRun.Load() || w.IsDone() {
		t.Fail()
	}
	w.Until(time.Now().Add(time.Second * 5))
	time.Sleep(time.Second * 4)
	if hasRun.Load() || w.IsDone() {
		t.Fail()
	}
	time.Sleep(time.Second * 3)
	if !hasRun.Load() || !w.IsDone() {
		t.Fail()
	}
}

func TestScheduleWorker_Immediately(t *testing.T) {
	fun, handler, hasRun := createDumbUtilities()

	w := NewScheduleWorker(fun, handler).For(time.Second * 5)
	w.DoImmediately()
	time.Sleep(time.Second)
	if !hasRun.Load() || !w.IsDone() {
		t.Fail()
	}
}

func TestScheduleWorker_Cancel(t *testing.T) {
	t.Parallel()

	fun, handler, hasRun := createDumbUtilities()

	w := NewScheduleWorker(fun, handler).For(time.Second * 2)
	w.Cancel()
	time.Sleep(time.Second * 5)
	if hasRun.Load() {
		t.Fail()
	}
	w.For(time.Second)
	time.Sleep(time.Second * 5)
	if hasRun.Load() || w.IsDone() {
		t.Fail()
	}
}
