package checkend

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Worker handles asynchronous sending of notices.
type Worker struct {
	config    *Configuration
	client    *Client
	queue     chan *Notice
	done      chan struct{}
	wg        sync.WaitGroup
	flushCh   chan chan struct{}
	running   bool
	runningMu sync.Mutex
}

// NewWorker creates a new Worker.
func NewWorker(config *Configuration) *Worker {
	return &Worker{
		config:  config,
		client:  NewClient(config),
		queue:   make(chan *Notice, config.MaxQueueSize),
		done:    make(chan struct{}),
		flushCh: make(chan chan struct{}),
	}
}

// Start starts the worker goroutine.
func (w *Worker) Start() {
	w.runningMu.Lock()
	defer w.runningMu.Unlock()

	if w.running {
		return
	}

	w.running = true
	w.wg.Add(1)
	go w.run()
}

// Stop stops the worker and waits for pending notices with timeout.
func (w *Worker) Stop() {
	w.runningMu.Lock()
	if !w.running {
		w.runningMu.Unlock()
		return
	}
	w.running = false
	w.runningMu.Unlock()

	close(w.done)

	// Wait with timeout
	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Graceful shutdown completed
	case <-time.After(w.config.ShutdownTimeout):
		// Timeout reached
		if w.config.Debug {
			fmt.Println("[Checkend] [warning] Shutdown timeout reached, some notices may not have been sent")
		}
	}
}

// Push adds a notice to the queue.
func (w *Worker) Push(notice *Notice) bool {
	w.runningMu.Lock()
	running := w.running
	w.runningMu.Unlock()

	if !running {
		return false
	}

	select {
	case w.queue <- notice:
		return true
	default:
		// Queue full
		return false
	}
}

// Flush waits for all queued notices to be sent.
func (w *Worker) Flush() {
	w.runningMu.Lock()
	running := w.running
	w.runningMu.Unlock()

	if !running || len(w.queue) == 0 {
		return
	}

	done := make(chan struct{})
	w.flushCh <- done
	<-done
}

func (w *Worker) run() {
	defer w.wg.Done()

	for {
		select {
		case <-w.done:
			w.drain()
			return

		case notice := <-w.queue:
			w.sendWithRetry(notice, 3)

		case done := <-w.flushCh:
			// Drain the queue for flush
			for len(w.queue) > 0 {
				select {
				case notice := <-w.queue:
					w.sendWithRetry(notice, 3)
				default:
					break
				}
			}
			close(done)
		}
	}
}

func (w *Worker) sendWithRetry(notice *Notice, maxRetries int) {
	for attempt := 0; attempt < maxRetries; attempt++ {
		resp := w.client.Send(notice)
		if resp != nil {
			return
		}

		if attempt < maxRetries-1 {
			delay := time.Duration(1<<uint(attempt)) * 100 * time.Millisecond
			time.Sleep(delay)
		}
	}
}

func (w *Worker) drain() {
	ctx, cancel := context.WithTimeout(context.Background(), w.config.ShutdownTimeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return
		case notice := <-w.queue:
			w.client.Send(notice)
		default:
			return
		}
	}
}
