package pool

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// 测试任务池
func TestTaskPool(t *testing.T) {

	for i := 0; i < 5; i++ {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			var (
				mu    = new(sync.RWMutex)
				count = 0
			)

			// 定义任务函数
			Task := func() {
				mu.Lock()
				count++
				mu.Unlock()
			}

			// 创建一个任务池
			pool := NewPool(
				WithPoolSize(10),
				WithName("test"),
				WithMutexLocker())
			var n = 100000
			// 创建n个任务
			for i := 0; i < n; i++ {
				pool.Go(Task)
			}
			time.Sleep(time.Second * 10)
			fmt.Printf("任务池执行完毕，结果为：%d\n", count)
		})
	}

}

// 任务数量
const benchmarkTimes = 10000 * 100

// 任务池大小
const poolSize = 100

const cpu = 8

// Benchmark results
// benchmarksTimes = 10000 * 1000 (1000w) poolSize = 8
// BenchmarkPoolWithMutex-8   	       1	5363986800 ns/op	14904072 B/op	  853594 allocs/op 5.3s
// BenchmarkPoolWithSpinLock-8   	       1	5606164000 ns/op	16582256 B/op	  992754 allocs/op 5.6s

// benchmarksTimes = 10000 * 1000 * 5 (5000w) poolSize = 8
// BenchmarkPoolWithMutex-8   	       1	29066707400 ns/op	86936424 B/op	 3836621 allocs/op 29s
// BenchmarkPoolWithSpinLock-8   	       1	29925896800 ns/op	87047072 B/op	 5239898 allocs/op 29s

// benchmarksTimes = 10000 * 10 (10w) poolSize = 8
// BenchmarkPoolWithMutex-8   	      21	  53854957 ns/op	  123188 B/op	    6826 allocs/op 0.5s
// BenchmarkPoolWithSpinLock-8   	      20	  64976005 ns/op	  202441 B/op	   12182 allocs/op 0.6s

// benchmarksTimes = 10000 * 10 (10w) cpu = 8 loop = 10000 * 100
// BenchmarkPoolWithMutex-8   	       1	1028093200 ns/op	  434896 B/op	   10152 allocs/op 1s
// BenchmarkPoolWithSpinLock-8   	       2	 997625500 ns/op	  268260 B/op	     104 allocs/op 1s

// benchmarksTimes = 10000 * 10 (10w) cpu = 8 loop = 10000 * 1000
// BenchmarkPoolWithMutex-8   	       1	9482649000 ns/op	 3703528 B/op	   99935 allocs/op 9s
// BenchmarkPoolWithSpinLock-8   	       1	10449792500 ns/op	 3621224 B/op	   94786 allocs/op 10s

// benchmarksTimes = 10000 * 10 * 4(40w) cpu = 8 loop = 10000 * 10000
//BenchmarkPoolWithMutex-8   	       1	42029635600 ns/op	14787752 B/op	  399029 allocs/op 42s
//BenchmarkPoolWithSpinLock-8   	       1	40625008400 ns/op	14488296 B/op	  378489 allocs/op

// BenchmarkPool-8   	       2	 631636550 ns/op	18278100 B/op	 1121325 allocs/op 0.6s
// BenchmarkPoolWithMutex-8   	       2	 647160350 ns/op	18202408 B/op	 1103015 allocs/op 0.6s
// BenchmarkPoolWithSpinLock-8   	       2	 724062450 ns/op	19186344 B/op	 1175579 allocs/op 0.7s
// BenchmarkPoolWithSpinLock-8   	       2	 784044400 ns/op	19493856 B/op	 1205073 allocs/op 0.8s

// 使用mutex 的任务池
func BenchmarkPoolWithMutex(b *testing.B) {
	// 创建一个任务池
	var wg sync.WaitGroup
	pool := NewPool(
		WithPoolSize(cpu),
		WithName("test"),
		WithMutexLocker())

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		wg.Add(benchmarkTimes)
		for j := 0; j < benchmarkTimes; j++ {
			pool.Go(func() {
				testFunc()
				wg.Done()
			})
		}
		wg.Wait()
	}
}

func testFunc() {
	DoCopyStack(0, 0)
}

// 使用spinLock 的任务池
func BenchmarkPoolWithSpinLock(b *testing.B) {
	// 创建一个任务池
	var wg sync.WaitGroup
	pool := NewPool(
		WithPoolSize(cpu),
		WithName("test"),
		WithSpinLocker())

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		wg.Add(benchmarkTimes)
		for j := 0; j < benchmarkTimes; j++ {
			pool.Go(func() {
				testFunc()
				wg.Done()
			})
		}
		wg.Wait()
	}
}

func DoCopyStack(a, b int) int {
	if b < 100 {
		return DoCopyStack(0, b+1)
	}
	return 0
}

// 测试锁的性能
// goroutine 数量
var goCount = 8

// 循环次数
var loop = 10000 * 10000

// goCount = 1 loop = 10000 * 10000
// BenchmarkMutex/Mutex-8         	       1	53894899100 ns/op 53.8s
// BenchmarkSpinLock/SpinLock-8         	       1	20043964100 ns/op

// goCount = 8 loop = 10000 * 1000
// BenchmarkMutex/Mutex-8         	       1	4459657800 ns/op 4.4s
// BenchmarkSpinLock/SpinLock-8         	       1	1786316600 ns/op 1.7s

// goCount = 8 loop = 10000 * 10000 (10000w)
// BenchmarkMutex/Mutex-8         	       1	53354728700 ns/op 53.3s
// BenchmarkSpinLock/SpinLock-8         	       1	18465041200 ns/op 18.4s

func BenchmarkMutex(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.Run("Mutex", func(b *testing.B) {
			benchmarkMutex()
		})
	}
}
func benchmarkMutex() {
	var count int
	var mu sync.Mutex
	wg := sync.WaitGroup{}
	wg.Add(goCount)
	for i := 0; i < goCount; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < loop; j++ {
				mu.Lock()
				count++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	fmt.Println(count)
}

func BenchmarkSpinLock(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.Run("SpinLock", func(b *testing.B) {
			benchmarkSpinLock()
		})
	}
}

func benchmarkSpinLock() {
	var count int
	var sl spinLock
	wg := sync.WaitGroup{}
	wg.Add(goCount)
	for i := 0; i < goCount; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < loop; j++ {
				sl.Lock()
				count++
				sl.Unlock()
			}
		}()
	}
	wg.Wait()
	fmt.Println(count)
}

// 测试锁是否可用
func TestMutex(t *testing.T) {
	t.Run("Mutex", func(t *testing.T) {
		lock := newMutexLock()
		var count int
		var wg sync.WaitGroup
		wg.Add(100)
		for i := 0; i < 100; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < 10000; j++ {
					lock.Lock()
					count++
					lock.Unlock()
				}
			}()
		}
		wg.Wait()
		fmt.Println(count)
	})
}

func TestSpinLock(t *testing.T) {
	t.Run("SpinLock", func(t *testing.T) {
		lock := newSpinLock()
		var count int
		var wg sync.WaitGroup
		wg.Add(100)
		for i := 0; i < 100; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < 10000; j++ {
					lock.Lock()
					count++
					lock.Unlock()
				}
			}()
		}
		wg.Wait()
		fmt.Println(count)
	})
}
