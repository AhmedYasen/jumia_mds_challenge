package workerpool

import (
	"sync/atomic"
	"time"
)

type WorkerPool[T any] struct {
	maxJobs   uint64
	jobRecBuf <-chan Job[T]
}

func New[T any](maxJobs uint64, jobRecvBuf <-chan Job[T]) *WorkerPool[T] {
	return &WorkerPool[T]{
		maxJobs:   maxJobs,
		jobRecBuf: jobRecvBuf,
	}
}

func (p *WorkerPool[T]) Run(f func(T, uint64)) {
	usedWorkers := int64(0)
	for {
		if uint64(usedWorkers) < p.maxJobs {
			atomic.AddInt64(&usedWorkers, 1)
			job := <-p.jobRecBuf
			go func() {
				f(job.Element, job.Id)
				atomic.AddInt64(&usedWorkers, -1)
			}()
		}
		time.Sleep(1 * time.Second)
	}
}
