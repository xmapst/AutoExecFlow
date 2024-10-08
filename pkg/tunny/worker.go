package tunny

import "sync/atomic"

type Handler func() error

// workRequest is a struct containing context representing a workers intention
// to receive a work payload.
type workRequest struct {
	// jobChan is used to send the payload to this worker.
	jobChan chan<- Handler

	// asyncJobChan is used to send the payload to this worker.
	asyncJobChan chan<- Handler

	// retChan is used to read the result from this worker.
	retChan <-chan error

	// interruptFunc can be called to cancel a running job. When called it is no
	// longer necessary to read from retChan.
	interruptFunc func()
}

// workerWrapper takes a Worker implementation and wraps it within a goroutine
// and channel arrangement. The workerWrapper is responsible for managing the
// lifetime of both the Worker and the goroutine.
type workerWrapper struct {
	worker        Worker
	interruptChan chan struct{}

	// reqChan is NOT owned by this type, it is used to send requests for work.
	reqChan chan<- workRequest

	// closeChan can be closed in order to cleanly shutdown this worker.
	closeChan chan struct{}

	// closedChan is closed by the run() goroutine when it exits.
	closedChan chan struct{}
}

func newWorkerWrapper(
	reqChan chan<- workRequest,
	worker Worker,
) *workerWrapper {
	w := workerWrapper{
		worker:        worker,
		interruptChan: make(chan struct{}),
		reqChan:       reqChan,
		closeChan:     make(chan struct{}),
		closedChan:    make(chan struct{}),
	}

	go w.run()

	return &w
}

func (w *workerWrapper) interrupt() {
	close(w.interruptChan)
	w.worker.Interrupt()
}

func (w *workerWrapper) run() {
	jobChan, retChan, asyncJobChan := make(chan Handler), make(chan error), make(chan Handler)
	defer func() {
		w.worker.Terminate()
		close(retChan)
		close(w.closedChan)
	}()

	for {
		// NOTE: Blocking here will prevent the worker from closing down.
		w.worker.BlockUntilReady()
		select {
		// give more priority to closeChan. Because select will randomly select one when multiple channel available,
		// when many closeChan and reqChan come at the same time in extreme cases, giving high priority to closeChan
		// to prevent closeChan to be ignored all the time
		case <-w.closeChan:
			return
		default:
			nextReq := workRequest{
				jobChan:       jobChan,
				retChan:       retChan,
				asyncJobChan:  asyncJobChan,
				interruptFunc: w.interrupt,
			}

			select {
			case w.reqChan <- nextReq:
				select {
				case payload := <-jobChan:
					result := w.worker.Process(payload)
					select {
					case retChan <- result:
					case <-w.interruptChan:
						w.interruptChan = make(chan struct{})
					}
				case payload := <-asyncJobChan:
					_ = w.worker.Process(payload)
					atomic.AddInt64(&w.worker.(*callbackWorker).pool.queuedJobs, -1)
				case <-w.interruptChan:
					w.interruptChan = make(chan struct{})
				}
			case <-w.closeChan:
				return
			}
		}
	}
}

func (w *workerWrapper) stop() {
	close(w.closeChan)
}

func (w *workerWrapper) join() {
	<-w.closedChan
}
