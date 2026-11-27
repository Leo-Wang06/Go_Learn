# 第 11 章：并发编程（Web 常用）

> **阶段**：阶段三 · 并发与工程基础
> **建议时间**：2 天
> **前置知识**：第 03 章（函数、闭包、defer）、第 04 章（切片、map）、第 07 章（接口）、第 08 章（错误处理）
> **代码基准**：Go 1.22 及以上

---

## 本章目标

学完本章，你应当能够：

1. 说清并发与并行的区别；
2. 正确创建、等待 goroutine，知道它为什么没有返回值；
3. 熟练使用 channel 的缓冲、方向、关闭、遍历和 comma-ok；
4. 用 `select` 实现多路等待、非阻塞检查和超时；
5. 用 `context` 传递取消信号、超时和请求级数据；
6. 用 `sync`、`atomic` 保护共享状态并做简单计数；
7. 识别数据竞争、goroutine 泄漏和循环变量陷阱；
8. 写出 worker pool、fan-in/fan-out，并理解 Gin/GORM 中 context 的传递方式。

> 不展开 GMP 调度器、内存模型细节和调度器源码。重点是「写对业务并发」和「会排查问题」。未写 `package main` 的代码块是可放进函数的片段，省略了 import 和函数包装。

---

## 11.1 并发 vs 并行

| 概念 | 含义 | 类比 |
|------|------|------|
| 并发 | 多个任务在**同一时间段**内交替推进 | 一个厨师同时照看几口锅 |
| 并行 | 多个任务在**同一时刻**真正同时执行 | 几个厨师各看一口锅 |

并发是一种结构，把任务拆成可独立推进的单元；并行是一种执行方式，需要多核 CPU。写了 `go` 不等于一定并行：单核上多个 goroutine 仍然可以并发，只是交替执行；多核上运行时可能把它们分配到不同核心。具体调度由 runtime 决定，业务代码不应依赖某个 goroutine 何时运行。

> 本章只保证「并发正确」，不保证「执行顺序」。涉及多个 goroutine 的输出，顺序都可能变化。

---

## 11.2 goroutine

### 11.2.1 用 `go` 启动

```go
func worker(id int) {
    fmt.Println("worker", id, "start")
    time.Sleep(100 * time.Millisecond)
    fmt.Println("worker", id, "done")
}

func main() {
    go worker(1)
    go worker(2)
    time.Sleep(300 * time.Millisecond)
    fmt.Println("main done")
}
```

可能输出 `worker 1 start`、`worker 2 start`、`worker 1 done`、`worker 2 done`、`main done`（顺序不保证）。

- `go worker(1)` 立即返回，主 goroutine 继续执行；
- goroutine 没有返回值，函数的返回值会被丢弃；
- 参数在 `go` 语句处求值，和 `defer` 类似。

### 11.2.2 主 goroutine 退出，程序就结束

```go
func main() {
    go fmt.Println("hello")
}
```

这段程序**可能什么都不输出**。`main` 返回后进程立即退出，不会等待其他 goroutine。正确做法：

```go
var wg sync.WaitGroup
wg.Add(1)
go func() {
    defer wg.Done()
    fmt.Println("hello")
}()
wg.Wait()
```

输出 `hello`。记住：**`main` 不等人**。

### 11.2.3 goroutine 没有返回值

需要结果时，把结果写进 channel：

```go
type result struct{ data string; err error }

ch := make(chan result, 1)
go func() {
    data, err := fetch(1) // fetch 返回 (string, error)
    ch <- result{data: data, err: err}
}()

r := <-ch
if r.err != nil {
    fmt.Println("failed:", r.err)
    return
}
fmt.Println(r.data) // user-1
```

也可以用 `WaitGroup` + 共享变量，但共享变量必须加锁；channel 通常更直观。

### 11.2.4 匿名函数与参数

```go
for i := 1; i <= 3; i++ {
    go func(id int) { fmt.Println(id) }(i)
}
```

显式传参避免闭包捕获问题。Go 1.22 起循环变量每次迭代新建，直接捕获也安全（见 11.10），但显式传参更清晰，也兼容旧语言版本。

### 11.2.5 goroutine 里的 panic

一个 goroutine panic 且没有 recover，会导致**整个程序崩溃**。任务可能 panic 时，必须在 goroutine 内部 `defer + recover`：

```go
go func() {
    defer func() {
        if r := recover(); r != nil { fmt.Println("recovered:", r) }
    }()
    panic("boom")
}()
```

Gin 的 `Recovery` 中间件就是在这个层面保护每个请求的 handler。

---

## 11.3 channel

```go
unbuffered := make(chan int)     // 无缓冲
buffered := make(chan int, 10)   // 有缓冲，容量 10
```

### 11.3.1 无缓冲 channel

无缓冲 channel 的发送和接收必须**同时就绪**，否则阻塞。

```go
ch := make(chan string)
go func() { ch <- "hello" }()

msg := <-ch
fmt.Println(msg) // hello
```

主 goroutine 执行 `<-ch` 时没有数据，阻塞；子 goroutine 发送，与接收方配对；双方继续执行。

只有发送、没有接收：

```go
ch := make(chan int)
ch <- 1
```

运行时报 `fatal error: all goroutines are asleep - deadlock!`。

### 11.3.2 有缓冲 channel

缓冲区未满时发送不阻塞；缓冲区为空时接收阻塞。

```go
ch := make(chan int, 2)
ch <- 1
ch <- 2
fmt.Println(len(ch), cap(ch)) // 2 2
fmt.Println(<-ch, <-ch)       // 1 2
```

`len(ch)` 是当前元素个数，`cap(ch)` 是容量；缓冲区满时继续发送阻塞，空时继续接收阻塞。缓冲大小应该有明确理由，不要用它掩盖设计问题。

### 11.3.3 阻塞与死锁

| 操作 | 无缓冲 | 有缓冲（未满/未空） |
|------|--------|---------------------|
| 发送 | 等接收方 | 有空间则立即返回 |
| 接收 | 等发送方 | 有数据则立即返回 |

常见死锁原因：同一 goroutine 先发送后接收且没有其他 goroutine；所有 goroutine 都在等一个永远不会发生的事件；等待一个已不可能再发送的 channel。

```go
func main() {
    ch := make(chan int)
    ch <- 1 // 永远等不到接收方
}
```

同样报 `fatal error: all goroutines are asleep - deadlock!`。

### 11.3.4 channel 方向

```go
func produce(out chan<- int) {
    for i := 1; i <= 3; i++ {
        out <- i
    }
    close(out)
}

func consume(in <-chan int) {
    for v := range in {
        fmt.Println(v)
    }
}

func main() {
    ch := make(chan int)
    go produce(ch)
    consume(ch) // 输出 1 2 3
}
```

| 类型 | 含义 |
|------|------|
| `chan T` | 可读可写 |
| `chan<- T` | 只能发送 |
| `<-chan T` | 只能接收 |

方向是编译期检查。`chan int` 可以传给 `chan<- int` 或 `<-chan int` 参数，反之不行。

### 11.3.5 关闭 channel

```go
ch := make(chan int, 2)
ch <- 10
close(ch)

v, ok := <-ch
fmt.Println(v, ok) // 10 true
v, ok = <-ch
fmt.Println(v, ok) // 0 false
```

关闭后接收方仍能取完缓冲区数据，取完后立即返回零值和 `false`；向已关闭 channel 发送会 panic：`send on closed channel`；重复关闭会 panic：`close of closed channel`；关闭 nil channel 也 panic。通常由**发送方**关闭，接收方不要关闭。

### 11.3.6 `range` 与 `ok`

```go
ch := make(chan int, 3)
ch <- 1
ch <- 2
ch <- 3
close(ch)

for v := range ch {
    fmt.Println(v) // 1 2 3
}
```

如果 channel 永远不关闭，`range` 会一直阻塞，可能导致 goroutine 泄漏。

comma-ok 区分「收到零值」和「channel 已关闭」：

```go
v, ok := <-ch
if !ok {
    fmt.Println("channel closed")
}
```

### 11.3.7 nil channel 与完成信号

未初始化的 channel 是 `nil`，发送和接收都会永久阻塞。它在 `select` 里可以动态禁用分支。

只通知完成、不传数据时，用 `struct{}`：

```go
done := make(chan struct{})
go func() {
    // 做一些工作
    close(done)
}()
<-done
```

`struct{}` 不占空间，`close` 会唤醒所有等待者。

---


## 11.4 select

### 11.4.1 多路等待

```go
ch1 := make(chan string)
ch2 := make(chan string)

go func() { time.Sleep(20 * time.Millisecond); ch1 <- "one" }()
go func() { time.Sleep(10 * time.Millisecond); ch2 <- "two" }()

select {
case v := <-ch1:
    fmt.Println("ch1:", v)
case v := <-ch2:
    fmt.Println("ch2:", v)
}
```

通常输出 `ch2: two`，因为 `ch2` 先就绪。但不要依赖顺序：两个 case 同时就绪时，Go 会随机选一个。

### 11.4.2 `default`：非阻塞

```go
select {
case v := <-ch:
    fmt.Println("received:", v)
default:
    fmt.Println("no data")
}
```

有就绪 case 就执行；没有就绪 case 且有 `default`，立即执行 `default`；没有 `default`，阻塞等待。常见用途：非阻塞检查数据、尝试发送而不阻塞。

### 11.4.3 超时

```go
go func() { time.Sleep(200 * time.Millisecond); ch <- "data" }()

select {
case v := <-ch:
    fmt.Println("received:", v)
case <-time.After(50 * time.Millisecond):
    fmt.Println("timeout") // 输出 timeout
}
```

`time.After(d)` 返回一个 channel，d 时间后收到值，适合一次性超时。在**热循环**里反复使用会创建大量定时器；更好的做法是 `context.WithTimeout` 或复用 `time.Timer`（第 12 章）。

### 11.4.4 空 select 与 nil 分支

`select {}` 没有任何 case，会永久阻塞，通常导致 deadlock。nil channel 的分支永远不会就绪：

```go
var ch chan int
select {
case v := <-ch:
    fmt.Println(v) // 永远不会执行
default:
    fmt.Println("default") // 输出 default
}
```

### 11.4.5 在循环中使用 select

```go
for {
    select {
    case v := <-ch:
        fmt.Println("data:", v)
    case <-ctx.Done():
        fmt.Println("exit:", ctx.Err())
        return
    }
}
```

这是 worker 的常见写法：有数据就处理，收到取消信号就退出。`return` 很重要，否则循环不会结束。

---

## 11.5 context

`context` 用于在 goroutine 之间传递**取消信号、超时和请求级数据**。一个 HTTP 请求触发数据库查询，客户端断开时，我们希望查询被取消、下游 goroutine 收到信号、超时自动生效。只靠 channel 也能做，但每个函数都要多传一个 channel；`context` 把这些能力统一成一个接口。

### 11.5.1 `Background` 与 `TODO`

```go
ctx := context.Background()
```

- `Background()` 返回非 nil、永远不会被取消的根 context；
- `TODO()` 也返回非 nil 的根 context，语义是「暂未决定」；
- 不要把 `nil` 传给需要 context 的函数，很多函数会 panic；
- 确定是根 context 用 `Background()`，还没想好用 `TODO()` 占位。

### 11.5.2 `WithCancel`

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

done := make(chan struct{})
go func() {
    <-ctx.Done()
    fmt.Println("stopped:", ctx.Err())
    close(done)
}()

cancel()
<-done
fmt.Println("root:", ctx.Err())
```

输出：

```
stopped: context canceled
root: context canceled
```

`cancel()` 关闭 `ctx.Done()`，所有等待者被唤醒；`ctx.Err()` 返回 `context.Canceled`。cancel 必须调用（通常 `defer cancel()`），否则可能造成 context 泄漏；多次调用是安全的。

### 11.5.3 `WithTimeout`

```go
ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
defer cancel()

select {
case <-time.After(200 * time.Millisecond):
    fmt.Println("finished")
case <-ctx.Done():
    fmt.Println(ctx.Err()) // context deadline exceeded
}
```

`WithTimeout(parent, d)` 等价于 `WithDeadline(parent, time.Now().Add(d))`。超时后 `ctx.Done()` 关闭，`ctx.Err()` 返回 `context.DeadlineExceeded`，所有子 context 一并取消。

Web 请求常见写法：

```go
ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
defer cancel()
```

### 11.5.4 `WithDeadline` 与取消传播

```go
deadline := time.Now().Add(50 * time.Millisecond)
ctx, cancel := context.WithDeadline(context.Background(), deadline)
defer cancel()
<-ctx.Done()
fmt.Println(ctx.Err()) // context deadline exceeded
```

`WithTimeout` 给的是时长，`WithDeadline` 给的是绝对时间点；父 context 截止更早时，子 context 以更早的为准。取消会从父传播到所有子：

```go
parent, cancelParent := context.WithCancel(context.Background())
child, cancelChild := context.WithTimeout(parent, time.Second)
defer cancelChild()

cancelParent()
<-child.Done()
fmt.Println(child.Err()) // context canceled
```

即使 child 的超时还没到，父取消后 child 也会取消。一个请求的 context 取消，所有下游数据库查询、RPC、goroutine 都能收到信号。

### 11.5.5 `WithValue`（慎用）

```go
type ctxKey string

const requestIDKey ctxKey = "requestID"

ctx := context.WithValue(context.Background(), requestIDKey, "req-123")
fmt.Println(ctx.Value(requestIDKey)) // req-123
```

- key 用**自定义类型**，避免不同包用同名字符串冲突；
- 只放请求级元数据：TraceID、用户 ID、租户 ID、语言等；
- 不要放业务参数、数据库连接、logger 等可以通过参数或依赖注入传递的东西；
- 值应不可变且并发安全；`Value` 返回 `any`，取出时类型断言：

```go
if id, ok := ctx.Value(requestIDKey).(string); ok {
    fmt.Println(id)
}
```

> 把 context 当「万能参数包」是反模式。函数需要什么就显式传什么。

### 11.5.6 context 使用规则

1. 作为函数的**第一个参数**，命名 `ctx`；
2. 不要存进结构体字段，除非生命周期与结构体一致；
3. 不要传 `nil`，用 `Background()` 或 `TODO()`；
4. 创建带取消的 context 后**一定调用 cancel**，通常 `defer cancel()`；
5. 不要用 context 传递可选参数或业务依赖；
6. 下游函数接收 ctx 后应继续往下传，不要新建 `Background()` 切断取消链。

---

## 11.6 sync

### 11.6.1 `Mutex`：互斥锁

多个 goroutine 同时修改共享变量会产生数据竞争。用 `sync.Mutex` 保护临界区：

```go
var (
    mu    sync.Mutex
    count int
    wg    sync.WaitGroup
)

for i := 0; i < 1000; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        mu.Lock()
        count++
        mu.Unlock()
    }()
}
wg.Wait()
fmt.Println(count) // 1000
```

`Lock()` 后立即 `defer mu.Unlock()`，防止 panic 或提前 return 导致死锁；临界区尽量小，不要把网络、文件 IO 放在锁里；不要复制包含 `Mutex` 的结构体；同一个 goroutine 不能重复 `Lock` 非重入锁。

### 11.6.2 `RWMutex`：读写锁

读多写少时用 `sync.RWMutex`：多个读可同时进行；写与读、写与写互斥。

```go
type Cache struct {
    mu   sync.RWMutex
    data map[string]string
}

func (c *Cache) Get(key string) (string, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    v, ok := c.data[key]
    return v, ok
}

func (c *Cache) Set(key, value string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.data[key] = value
}

// 使用前初始化：cache := &Cache{data: make(map[string]string)}
```

读锁里不能修改 map；`RWMutex` 不一定比 `Mutex` 快，读操作很短时普通 `Mutex` 可能更简单；不要嵌套获取同一把读写锁。

### 11.6.3 `WaitGroup`：等待一组 goroutine

```go
var wg sync.WaitGroup

for i := 1; i <= 3; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        fmt.Println("worker", id)
    }(i)
}
wg.Wait()
fmt.Println("all done")
```

`Add(n)` 必须在启动 goroutine **之前**调用；每个 `Add(1)` 对应一个 `Done()`；`Wait()` 阻塞到计数归零；计数不能变成负数，否则 panic：`sync: negative WaitGroup counter`。

### 11.6.4 `Once`：只执行一次

```go
var once sync.Once

for i := 0; i < 3; i++ {
    once.Do(func() { fmt.Println("init once") }) // 只输出一次
}
```

`Once.Do` 保证函数在多个 goroutine 中只执行一次，常用于初始化全局配置、加载一次数据、注册一次回调。如果 `Do` 里的函数 panic，`Once` 会认为它已经执行过，后续调用不会再执行，所以初始化逻辑要自己保证不 panic。

---

## 11.7 atomic

`sync/atomic` 提供原子操作，适合简单计数器、标志位。

```go
var counter atomic.Int64
var wg sync.WaitGroup

for i := 0; i < 1000; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        counter.Add(1)
    }()
}
wg.Wait()
fmt.Println(counter.Load()) // 1000
```

`Store` 写入、`Load` 读取、`CompareAndSwap(old, new)` 比较并交换；旧式 API 是 `atomic.AddInt64(&n, 1)`、`atomic.LoadInt64(&n)`。

边界：原子操作适合**单个变量**的读改写；多个变量需要保持一致（例如「余额 + 流水」）时，用 Mutex 或数据库事务；不要用 atomic 保护 map、切片等复合结构；`atomic.Value` 可以存任意类型，但写入类型必须一致。

---

## 11.8 数据竞争与 `-race`

### 11.8.1 什么是数据竞争

两个 goroutine 同时访问同一个变量，至少一个是写，且没有同步，就是数据竞争。

```go
count := 0
var wg sync.WaitGroup

for i := 0; i < 1000; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        count++ // 数据竞争
    }()
}
wg.Wait()
fmt.Println(count) // 结果不确定，通常小于 1000
```

`count++` 不是原子操作，底层大致是「读 - 加 - 写」，多个 goroutine 会互相覆盖。

### 11.8.2 用 `-race` 检测

```powershell
go run -race .
go test -race ./...
```

发现竞争时会打印 `WARNING: DATA RACE`，列出读写位置和 goroutine。注意：`-race` 只能发现**实际执行到**的竞争，不能证明没有竞争；开启后程序变慢、占用更多内存，适合测试和排查；Windows 上若提示 `-race requires cgo`，需要安装 gcc/mingw 并设置 `CGO_ENABLED=1`。CI 里建议至少跑一次 `go test -race ./...`。

### 11.8.3 map 的并发访问

内置 map **不支持并发读写**，同时读写会直接崩溃：

```
fatal error: concurrent map read and map write
```

即使「一个 goroutine 只读、一个只写」也不行。解决方案：用 `sync.Mutex`/`RWMutex` 保护；用 `sync.Map`（适合读多写少、键集合相对稳定）；或每个 goroutine 操作自己的 map，最后合并。

```go
var mu sync.RWMutex
m := make(map[string]int)

mu.Lock()
m["a"] = 1
mu.Unlock()

mu.RLock()
fmt.Println(m["a"])
mu.RUnlock()
```

`sync.Map` 的 API 是 `Load`、`Store`、`Delete`、`Range`。不要把它当「并发版 map」无脑替换普通 map；多数场景用 `RWMutex` 更清晰。

---

## 11.9 goroutine 泄漏

goroutine 泄漏是指 goroutine 永远阻塞、无法退出，长期占用内存和栈。

```go
func leak() {
    ch := make(chan int)
    go func() {
        ch <- 1 // 没有接收方，永远阻塞
    }()
}
```

`leak()` 返回后，子 goroutine 仍阻塞在发送上，永远不会被回收。修复方式：

```go
// 方式一：缓冲 channel，发送不会阻塞
ch := make(chan int, 1)
go func() { ch <- 1 }()

// 方式二：用 context 提供退出路径
func worker(ctx context.Context, ch chan<- int) {
    select {
    case ch <- 1:
    case <-ctx.Done():
        return
    }
}
```

其他常见泄漏：`for { select { ... } }` 没有退出分支；`range ch` 但 channel 永远不关闭；HTTP 响应体没有 `Close`；`time.Ticker` 没有 `Stop`；context 没有 `cancel`。

发现方法：`runtime.NumGoroutine()` 观察数量是否持续增长；pprof 的 goroutine profile 查看堆栈；测试里反复执行，断言数量回到基线；代码审查重点看「启动 goroutine 的地方，退出条件是什么」。

```go
before := runtime.NumGoroutine()
leak()
time.Sleep(50 * time.Millisecond)
fmt.Println(runtime.NumGoroutine() > before) // true
```

这只是粗略演示；生产排查用 pprof 更可靠。

---

## 11.10 Go 1.22 循环变量

旧版本（Go 1.21 及以下）的 `for` 循环变量在整个循环中只有一个实例：

```go
var wg sync.WaitGroup
for i := 0; i < 3; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        fmt.Println(i)
    }()
}
wg.Wait()
```

Go 1.21 语义下，三个 goroutine 很可能都打印 `3`：闭包捕获同一个 `i`，循环结束时 `i == 3`。

Go 1.22 起，**每次迭代创建新的循环变量**，上面的代码打印 `0`、`1`、`2`（顺序不保证）。

这个变化由模块 `go.mod` 的 `go` 指令决定：`go 1.22` 及以上启用新语义；`for range` 的索引和值也一样；模块还写 `go 1.21` 时，即使工具链是 1.22+，也保持旧语义。显式传参 `go func(id int){...}(i)` 在任何版本都安全。

---

## 11.11 并发模式

### 11.11.1 worker pool

固定数量的 worker 从同一个 channel 取任务：

```go
jobs := make(chan int, 5)
results := make(chan int, 5)
var wg sync.WaitGroup

for w := 1; w <= 3; w++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        for j := range jobs {
            results <- j * j
        }
    }()
}

for i := 1; i <= 5; i++ {
    jobs <- i
}
close(jobs)

wg.Wait()
close(results)

sum := 0
for r := range results {
    sum += r
}
fmt.Println(sum) // 55
```

启动 3 个 worker 阻塞在 `range jobs`；主 goroutine 发送 5 个任务后关闭 jobs；worker 处理完退出；`wg.Wait()` 等所有 worker 结束；关闭 results 后汇总。

优点：限制并发数，避免一次启动成千上万个 goroutine；任务和 worker 解耦。

### 11.11.2 fan-out

fan-out 是「一个输入，多个消费者」：多个 worker 同时从同一个 channel 读任务。上面的 worker pool 就是 fan-out。

### 11.11.3 fan-in

fan-in 是「多个输入，一个输出」：把多个 channel 合并成一个。

```go
func fanIn(chs ...<-chan int) <-chan int {
    out := make(chan int)
    var wg sync.WaitGroup

    for _, ch := range chs {
        wg.Add(1)
        go func(c <-chan int) {
            defer wg.Done()
            for v := range c {
                out <- v
            }
        }(ch)
    }

    go func() {
        wg.Wait()
        close(out)
    }()
    return out
}

// 使用：for v := range fanIn(ch1, ch2) { sum += v }
```

顺序不保证，但总和确定。关键：每个输入由一个 goroutine 读取；用 `WaitGroup` 等它们结束；最后关闭输出 channel。

### 11.11.4 组合

真实业务常见 pipeline：`producer -> jobs -> N 个 worker -> results -> 汇总`。规则是：**谁发送，谁关闭；谁等待，谁协调**。

---

## 11.12 Web 中的并发

### 11.12.1 每个请求一个 goroutine

`net/http` 为每个请求启动一个 goroutine，Gin 建立在 `net/http` 之上。多请求天然并发，handler 的局部变量一般不用加锁；但**共享状态**（全局 map、缓存、计数器）仍需同步。

```go
var visits atomic.Int64

func handler(c *gin.Context) {
    visits.Add(1)
    c.JSON(200, gin.H{"visits": visits.Load()})
}
```

### 11.12.2 `r.Context()`

每个 `http.Request` 都带 context：客户端断开时被取消，服务器关闭、超时中间件也可能取消它。把 ctx 传给下游数据库、RPC、HTTP 调用，取消才能传播。

```go
func Handler(w http.ResponseWriter, r *http.Request) {
    select {
    case <-time.After(3 * time.Second):
        fmt.Fprintln(w, "done")
    case <-r.Context().Done():
        fmt.Println("client gone:", r.Context().Err())
    }
}
```

### 11.12.3 Gin 中的 context

Gin 的 `*gin.Context` 包装了 `http.Request` 和 `http.ResponseWriter`。请求 context 用 `c.Request.Context()` 获取。

| 名称 | 用途 |
|------|------|
| `*gin.Context` | 读参数、写响应、中间件链 |
| `context.Context` | 取消、超时、跨层传递 |
| `c.Request.Context()` | 传给 GORM/RPC/HTTP |

`*gin.Context` 不是并发安全的，不要传给后台 goroutine 长期使用。后台任务可以：

- 用 `c.Copy()` 复制只读的 Gin context；
- 或只取出需要的值，用独立 `context.Background()` + 超时；
- 必须脱离请求取消但保留值时，用 `context.WithoutCancel(c.Request.Context())`（Go 1.21+）。

### 11.12.4 GORM 与 context

```go
db.WithContext(ctx).First(&user, id)                     // 传统 API
user, err := gorm.G[User](db).Where("id = ?", id).First(ctx) // 泛型 API，GORM v1.30+
```

Gin handler 中的完整示例：

```go
func GetUser(c *gin.Context) {
    ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
    defer cancel()

    var user User
    if err := db.WithContext(ctx).First(&user, "id = ?", c.Param("id")).Error; err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            c.JSON(http.StatusGatewayTimeout, gin.H{"error": "query timeout"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, user)
}
```

取消传播链：客户端断开 -> `http.Request.Context()` 取消 -> handler 的 ctx 取消 -> GORM 查询取消 -> 驱动中断查询。任何一层用了 `context.Background()`，链就断了。

### 11.12.5 后台任务

```go
go func() {
    db.WithContext(c.Request.Context()).Create(&log) // 错误
}()
```

请求结束后 ctx 会取消，后台任务可能立刻失败；`c` 也不适合长期使用。应明确后台任务生命周期，用独立 context + 超时，只传必要数据；任务队列通常比裸 goroutine 可靠。

---

## 11.13 常见错误与陷阱

| 错误/现象 | 原因 | 修复 |
|-----------|------|------|
| 程序什么都没输出 | main 退出，没等 goroutine | WaitGroup/channel |
| `all goroutines are asleep - deadlock!` | 发送/接收没有配对 | 检查配对、缓冲、关闭 |
| `send on closed channel` / `close of closed channel` | 关闭后发送/重复关闭 | 由发送方只关闭一次 |
| `concurrent map read and map write` | map 并发读写 | Mutex/RWMutex/sync.Map |
| 计数结果小于预期 | `count++` 数据竞争 | Mutex/atomic，`-race` |
| `sync: negative WaitGroup counter` | Done 多于 Add | 检查 Add/Done 配对 |
| 锁住后不释放 | 忘记 Unlock/提前 return | `defer mu.Unlock()` |

另外：

- **不要用 sleep 当同步**：`time.Sleep` 只保证至少等了这么久，用 `WaitGroup`、channel 或 context。
- **不要在持有锁时做 IO**：慢操作会让所有等锁的 goroutine 一起变慢，还可能因超时长期占用锁。
- **不要关闭接收方的 channel**：关闭应由发送方负责，接收方关闭会让发送方 panic。
- **`select` 的 case 只做 channel 操作**：在 case 里做长时间计算会拖慢整个循环。

---

## 11.14 与 Gin/GORM 的关联

- **Gin**：每个请求一个 goroutine；中间件链在同一 goroutine 中顺序执行；`*gin.Context` 不是并发安全的，不要在多个 goroutine 中同时写响应；请求级数据用 `c.Set`/`c.Get`，但只在同一请求的同步流程中使用。
- **GORM**：`db.WithContext(ctx)` 把取消信号传给底层 `database/sql`；查询超时、客户端断开时数据库操作会被取消；泛型 API `gorm.G[T](db).First(ctx)` 直接接收 ctx；事务中同样要传 ctx。
- **常见场景**：请求超时用 `context.WithTimeout(c.Request.Context(), d)`；客户端断开时监听 `ctx.Done()`；并发调用多个下游用 `errgroup.Group` 或 `WaitGroup`；限流用 worker pool/semaphore；缓存用 `RWMutex`/`sync.Map`；计数器用 `atomic.Int64`；后台任务用独立 context + 超时。
- `golang.org/x/sync/errgroup` 封装了 WaitGroup + 错误传播 + context 取消，读框架代码时常见；先掌握原生 `WaitGroup`。

---

## 11.15 练习题

### 基础题

1. 启动两个 goroutine 分别打印 1–5，用 `WaitGroup` 等待结束。
2. 用无缓冲 channel 在两个 goroutine 之间传一个字符串；再用容量为 3 的缓冲 channel，用 `select + default` 观察第 4 次发送是否阻塞。
3. 用 `close` + `range` 遍历 channel，输出 1、2、3；用 `v, ok` 判断 channel 是否已关闭。
4. 用 `select` + `time.After` 实现 100ms 超时。
5. 用 `context.WithTimeout` 取消一个耗时 1s 的操作。
6. 用 `sync.Mutex` 保护计数器，启动 1000 个 goroutine，最后输出 1000。
7. 用 `sync.Once` 实现只初始化一次的配置加载。

### 进阶题

8. 用 `atomic.Int64` 重写第 6 题，并用 `go run -race` 验证。
9. 写一个 goroutine 泄漏的例子，再用缓冲 channel 修复。
10. 用 `context.WithCancel` 实现「主 goroutine 通知 3 个 worker 退出」。
11. 解释 Go 1.22 前后 `for i := 0; i < 3; i++ { go func() { fmt.Println(i) }() }` 的输出差异。
12. 用 worker pool 计算 1–100 的平方和，限制 5 个 worker；再用 fan-in 合并两个 channel，统计总和。
13. 用 `RWMutex` 实现并发安全的缓存，支持 Get/Set。
14. 用 `context.WithValue` 传递 requestID，并在下游函数中取出。
15. 写一个 Gin handler，用 `c.Request.Context()` + `context.WithTimeout` 查询数据库，超时返回 504；解释为什么 `*gin.Context` 不能传给后台 goroutine 长期使用。

### 读文档题

16. 执行 `go doc context.Context`，回答它有哪些方法。
17. 执行 `go doc sync.WaitGroup`，回答 `Add`、`Done`、`Wait` 的签名与用途。
18. 执行 `go doc sync/atomic.Int64`，回答 `Add`、`Load`、`Store`、`CompareAndSwap` 的用途。
19. 打开 Gin 文档找到 `c.Request.Context()`，再打开 GORM 文档找到 `WithContext`，说明两者如何配合。

---

## 11.16 验收清单

全部打勾后，再进入第 12 章：

- [ ] 能说清并发与并行的区别；
- [ ] 能正确启动、等待 goroutine，知道 main 退出会结束程序；
- [ ] 能使用无缓冲/有缓冲 channel，解释阻塞行为，正确关闭并遍历；
- [ ] 能用 `select` 实现多路等待、`default` 和超时；
- [ ] 能用 `context` 的 Cancel/Timeout/Deadline/Value，并解释取消传播；
- [ ] 能用 `Mutex`、`RWMutex`、`WaitGroup`、`Once`、`atomic`；
- [ ] 能识别数据竞争并会用 `go test -race`，能修复 goroutine 泄漏；
- [ ] 能解释 Go 1.22 循环变量变化，能写出 worker pool 和 fan-in；
- [ ] 能在 Gin handler 中把 context 传给 GORM。

---

## 11.17 本章小结

| 概念 | 一句话解释 | 易错点 |
|------|------------|--------|
| 并发 | 多个任务交替推进 | 不等于并行 |
| goroutine | `go f()` 启动的执行单元 | 没有返回值，main 不等待 |
| channel | goroutine 间通信管道 | 无缓冲要配对，关闭后不能发送 |
| `select` | 多路等待 | 同时就绪随机选，空 select 永久阻塞 |
| `context` | 取消、超时、请求元数据 | 必须 cancel，不能传 nil |
| `sync` | Mutex/RWMutex/WaitGroup/Once | 不要复制锁，Add 在 go 之前 |
| `atomic` | 原子操作 | 只适合单个变量 |
| 数据竞争 | 无同步的并发读写 | 用 `-race` 检测 |
| goroutine 泄漏 | goroutine 永远阻塞 | context/close/缓冲修复 |

---

## 11.18 下一章预告

第 12 章：**常用标准库（上）**。你将学到：`fmt` 常用动词；`strings`、`strconv`；`bytes.Buffer`；`sort`、`slices`、`maps`；`time` 的格式化、解析、时区、Timer/Ticker；`log/slog` 结构化日志；`embed` 嵌入静态资源。这些是日常业务代码和 Gin/GORM 项目中最常打交道的标准库。
