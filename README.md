## Schedule Worker

### Introduction

I needed something both simple and more functional than a familiar `go func() { time.Sleep(...); foo(); }`

So this package was created. All code review/additions would be very appreciated.

### Functions Overview
`workerFoo := NewScheduleWorker(foo) // without error handler`  
`workerBar := NewScheduleWorker(bar, OnErrorFunc) // with error handler`  
Main 3 functions of a worker: `For`, `Until`, `Add`
- `workerFoo.For(time.Minute)` -- schedule for a minute from now
- `workerBar.Until(time.Parse("2006-01-02", "2077-01-01"))` -- schedule until 2077
- `workerBar.For(time.Minute).Add(time.Minute * 5)` -- `worker.Add` allows to increase/decrease datetime scheduled

`DoImmediately`, `Cancel`, `GetTime`, `IsDone`, `ExtNewScheduleWorker` -- could be useful depend on your use case
### Example

```go
package main

import (
	"log"
	"os"
	"time"

	_ "github.com/dontsellfish/Schedule-Worker"
)

func DeleteExampleFile() error {
	return os.Remove("example.jpg")
}

func OnError(err error) {
	log.Println(err.Error())
}

func main() {
	unneededFileDeletionWorker := NewScheduleWorker(DeleteExampleFile, OnError).For(time.Minute)
	// some initialization
	for {
		select {
		case _ = <-FileAccessRequested:
			unneededFileDeletionWorker.For(time.Minute)
			// some work with the file
		case _ = <-DeleteRightNow:
			unneededFileDeletionWorker.Immediately()
		case _ = <-NeverDeleteTheFile:
			unneededFileDeletionWorker.Cancel()
		}
	}
}
```