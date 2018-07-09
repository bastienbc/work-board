package worker

import (
	"sync"
)

//SimpleWork is here to be able to use a function as a Work
type SimpleWork func()

//Run simply call the function
func (w SimpleWork) Run() {
	w()
}

//Work is the interface that describes a Work.
//It only needs one method
type Work interface {
	Run()
}

//WorkBoard is a board with two parts
//The first parts is where Work request are post
//The second is where a Worker tell he is available
//Once you start dispatching jobs,
//Work on the board are affected to the next available Worker
type WorkBoard struct {
	queue       chan chan Work
	workQueue   chan Work
	stopTrigger chan interface{}
	wg          *sync.WaitGroup
}

//NewWorkBoard sets how many Work you can post without waiting,
//And how many workers max should be working
//You can affect less Worker than what has been defined here
//But if you affect more, there will be only max Workers working in parallel
func NewWorkBoard(maxWaitingJobs int, nbWorkers int) *WorkBoard {
	return &WorkBoard{
		queue:       make(chan chan Work, nbWorkers),
		workQueue:   make(chan Work, maxWaitingJobs),
		stopTrigger: make(chan interface{}),
		wg:          new(sync.WaitGroup),
	}
}

//PostWork pins a Work on the WorkBoard
//It will be executed later, when a Worker will consume it
func (wb *WorkBoard) PostWork(work Work) {
	wb.workQueue <- work
}

//StartDispatch execute the Work dispatch to the Workers in a goroutine
//Close will wait for it to be done before closig ressources
func (wb *WorkBoard) StartDispatch() {
	wb.wg.Add(1)
	go func() {
		defer wb.wg.Done()
		for {
			select {
			case work := <-wb.workQueue:
				go func() {
					worker := <-wb.queue
					worker <- work
				}()
			case <-wb.stopTrigger:
				return
			}
		}
	}()
}

//Close waits for the dispatch to be over before closing ressources
func (wb *WorkBoard) Close() {
	wb.stopTrigger <- true
	wb.wg.Wait()
	close(wb.queue)
	close(wb.workQueue)
	close(wb.stopTrigger)
}
