package pool

import (
	"fmt"
	"log"
	"runtime/debug"
	"sync"
	"sync/atomic"
)

var workerPool sync.Pool

func init() {
	workerPool.New = newWorker
}

type worker struct {
	pool *Pool
}

func newWorker() interface{} {
	return &worker{}
}

func (w *worker) run() {
	go func() {
		for {
			var t *task
			// task 队列上锁
			w.pool.taskLock.Lock()
			// 从 task 队列中取出一个 task
			if w.pool.taskHead != nil {
				t = w.pool.taskHead
				// 将 task 队列的头指针指向下一个 task
				w.pool.taskHead = w.pool.taskHead.next
				// 如果 task 队列的头指针为空，则将尾指针也置空
				atomic.AddInt32(&w.pool.taskCount, -1)
				//w.pool.taskCount--
			}

			if t == nil {
				// 没有 task 了，退出
				// close worker = count--
				w.close()
				// 释放 task 队列锁
				w.pool.taskLock.Unlock()
				// 使用sync.Pool 回收 worker
				w.Recycle()
				return
			}
			// 释放 task 队列锁
			w.pool.taskLock.Unlock()

			// 执行 task
			func() {
				defer func() {
					if r := recover(); r != nil {
						msg := fmt.Sprintf("gopool name is %s,  panic errors is %v, stack is %s", w.pool.options.name, r, debug.Stack())
						log.Println(msg)
					}
				}()
				// 执行 task
				t.f()
			}()
			// 回收 task
			t.Recycle()
		}
	}()
}

func (w *worker) close() {
	// worker 数量减一
	w.pool.decWorkerCount()
}
func (w *worker) Recycle() {
	// 将 worker 的 pool 置空
	w.pool = nil
	// 将 worker 放入 sync.Pool 中
	workerPool.Put(w)
}
