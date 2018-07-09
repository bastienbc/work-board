package worker

import (
	"sync"
)

//Worker is a structure taht can consumes Work posted on a WorkBoard
//It uses resources that need to be closed, so use close after you don't need it
type Worker struct {
	workQueue   chan Work
	stopTrigger chan interface{}
	cond        *sync.Cond
	stopped     bool
}

//NewWorker builds a worker, and allocate ressources
func NewWorker() *Worker {
	return &Worker{
		workQueue:   make(chan Work),
		stopTrigger: make(chan interface{}),
		cond:        sync.NewCond(new(sync.Mutex)),
		stopped:     true,
	}
}

//Work will tell the worker to consume Work posted on a WorkBoard
//In fact it post a receiver on the WorkBoard, and when that receiver is triggered
//It consumes the Work and run it
//When it is finished, it post its receiver again, telling it is available again
//This is non blocking, so use Stop to tell a worker to stop postng its receiver
//Worker should not be working on another WorkBoard in parallel
func (w *Worker) Work(wb *WorkBoard) {
	go func() {
		w.stopped = false
		for {
			wb.queue <- w.workQueue

			select {
			case work := <-w.workQueue:
				work.Run()
			case <-w.stopTrigger:
				w.cond.L.Lock()
				defer w.cond.L.Unlock()
				w.stopped = true
				w.cond.Signal()
				return
			}
		}
	}()
}

//Stop will tell the worker to not post its receiver to the current WorkBoard
//The worker actually stops as soon as it is not working on anything
func (w *Worker) Stop() {
	go func() {
		w.stopTrigger <- true
	}()
}

//Close deallocate resources created with NewWorker
func (w *Worker) Close() {
	w.cond.L.Lock()
	defer w.cond.L.Unlock()
	for !w.stopped {
		w.cond.Wait()
	}
	close(w.stopTrigger)
	close(w.workQueue)
}
