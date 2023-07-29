package pool

import (
	"strings"
	"sync"
	"sync/atomic"
)

// 选项模式：用于设置任务池的属性
type PoolOptions struct {
	// 任务池的大小（启动多少goroutine）
	poolSize int32
	// pool 的名字
	name string
	// 选择mutex 或者 spinLock
	lockName string
}

// 选项模式：用于设置任务池的 Goroutine 数量
func WithPoolSize(poolSize int32) func(*PoolOptions) {
	return func(options *PoolOptions) {
		options.poolSize = poolSize
	}
}

// 选项模式：用于设置任务池的名字
func WithName(name string) func(*PoolOptions) {
	return func(options *PoolOptions) {
		options.name = name
	}
}

// 选项模式：用于设置任务池的锁为mutex
func WithMutexLocker() func(*PoolOptions) {
	return func(options *PoolOptions) {
		options.lockName = "mutex"
	}
}

// 选项模式：用于设置任务池的锁为spinLock
func WithSpinLocker() func(*PoolOptions) {
	return func(options *PoolOptions) {
		options.lockName = "spinlock"
	}
}

type pool struct {

	// pool 选项配置
	options *PoolOptions
	//	任务队列
	// 可以使用chan，但是chan的缓冲区是有限的，如果任务队列中的任务数量超过了缓冲区的大小，那么就会阻塞
	// 所以这里使用链表来实现任务队列，更加灵活
	taskHead  *task
	taskTail  *task
	taskCount int32
	taskLock  sync.Locker

	// worker的数量
	workerCount int32
}

func NewPool(options ...func(*PoolOptions)) pool {
	// 初始化默认配置
	poolOptions := &PoolOptions{}
	for _, option := range options {
		option(poolOptions)
	}

	p := pool{
		options: poolOptions,
	}

	p.options.lockName = strings.ToLower(p.options.lockName)
	if p.options.lockName == "" {
		// 默认使用mutex
		p.options.lockName = "mutex"
	}

	if p.options.lockName == "mutex" {
		p.taskLock = NewMutexLock()
	} else {
		p.taskLock = NewSpinLock()
	}

	return p
}

func (p *pool) Name() string {
	return p.options.name
}

// 设置任务池的大小
func (p *pool) SetPoolSize(poolSize int32) {
	atomic.StoreInt32(&p.options.poolSize, poolSize)
}

// 获取任务池的大小
func (p *pool) Cap() int32 {
	return atomic.LoadInt32(&p.options.poolSize)
}

// 执行一个任务
func (p *pool) Go(f func()) {
	p.goFunc(f)
}

// 减少worker的数量
func (p *pool) decWorkerCount() {
	atomic.AddInt32(&p.workerCount, -1)
}

// 增加worker的数量
func (p *pool) incWorkerCount() {
	atomic.AddInt32(&p.workerCount, 1)
}

// 获取worker的数量
func (p *pool) getWorkerCount() int32 {
	return atomic.LoadInt32(&p.workerCount)
}

func (p *pool) goFunc(f func()) {
	// task Pool中获取一个task
	task := taskPool.Get().(*task)
	task.f = f

	// 将task放入任务队列
	p.taskLock.Lock()
	if p.taskHead == nil {
		p.taskHead = task
		p.taskTail = task
	} else {
		p.taskTail.next = task
		p.taskTail = task
	}
	p.taskLock.Unlock()

	// 判断是否需要启动一个新的worker
	if p.getWorkerCount() < p.Cap() {
		// 启动一个新的worker
		p.incWorkerCount()
		worker := workerPool.Get().(*worker)
		// worker 与 pool 之间的关联
		worker.pool = p
		worker.run()
	}
}
