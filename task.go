package pool

import (
	"sync"
)

// 任务

var taskPool sync.Pool

func init() {
	taskPool.New = newTask
}

type task struct {
	// 任务执行的上下文
	f func()
	// 链表的下一个任务
	next *task
}

func (t *task) Recycle() {
	t.f = nil
	t.next = nil
	// 将 task 放入 sync.Pool 中
	taskPool.Put(t)
}

func newTask() interface{} {
	return &task{}
}
