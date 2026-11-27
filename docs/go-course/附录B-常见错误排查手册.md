# 附录 B：常见错误排查手册

> **用途**：按报错关键字快速定位原因与修复方法。
> **说明**：Go 版本、操作系统、驱动不同，错误文字可能略有差异，按关键字匹配即可。
> **配套章节**：第 01–19 章。

---

## B.1 通用排查流程

1. **从下往上看**：panic 栈最后一行是根因，上面是调用链。
2. **找自己的代码**：跳过标准库/框架栈帧，定位第一个项目路径。
3. **最小复现**：缩到 20 行以内，多数问题会在这个过程中暴露。
4. **三道检查**：`go build ./...` → `go vet ./...` → `go test ./... -race`。
5. **不要猜**：用 `fmt.Printf("%T %#v\n", v, v)`、`go doc`、Delve 确认实际类型和值。

---

## B.2 编译错误

| 错误信息 | 原因 | 修复 |
|----------|------|------|
| `undefined: fmt` / `undefined: userName` | 未导入、拼写错误、作用域不对，或跨包用了未导出标识符 | 补 `import`、核对拼写；跨包首字母必须大写 |
| `declared and not used: x` | 局部变量声明后未使用（包级无此限制） | 删除变量，或 `_ = x` |
| `"fmt" imported and not used` | 导入未使用 | 删除导入；仅需副作用时用 `import _ "..."` |
| `invalid operation: a + b (mismatched types int and int64)` | Go 无隐式数值转换 | 显式写 `int(b)`；字符串数字用 `strconv` |
| `cannot use b (variable of type int64) as int value in assignment` | 同上 | 同上；注意溢出与精度 |
| `missing return` | 有执行路径没有返回值 | 保证所有分支返回，末尾加兜底 `return` |
| `assignment mismatch: 2 variables but f returns 1 value` | 左侧变量数与返回值数不一致 | 补齐变量或调整函数签名 |
| `no new variables on left side of :=` | `:=` 左侧至少要有一个新变量 | 改用 `=`，或引入新变量 |
| `syntax error: non-declaration statement outside function body` | 在包级用了 `:=` 或普通语句 | 包级只能写 `var` / `const` / `type` / `func` |
| `syntax error: unexpected newline, expected }` | 左大括号没和 `if` / `for` / `func` 同行 | 写成 `if ok {`，`{` 留在同一行 |
| `cannot assign to m["k"]` / `cannot take address of m["k"]` | map 元素不可寻址 | 取出到临时变量改完放回，或 map 存指针 |
| `Circle does not implement Shape (method Area has pointer receiver)` | 指针接收者的方法只属于 `*Circle` 方法集 | 传 `&c` / `&Circle{}`，或改值接收者 |
| `cannot use []string{...} as []any value` | 具体类型切片与 `[]any` 不通用 | 逐元素转换，或直接使用具体类型 |

**map 元素修改示例**：

```go
u := m["k"]
u.Age = 18
m["k"] = u

// 或让 map 存指针
m2 := map[string]*User{"k": {Age: 18}}
m2["k"].Age = 20
```

> `string(65)` 得到的是 `"A"` 而不是 `"65"`；数字转字符串必须用 `strconv.Itoa` / `FormatInt`。

---

## B.3 运行时 panic

| 错误信息 | 原因 | 修复 |
|----------|------|------|
| `panic: runtime error: invalid memory address or nil pointer dereference` | 对 nil 指针取字段/调方法 | 解引用前判空；先检查 error 再使用返回值 |
| `panic: runtime error: index out of range [3] with length 3` | 索引 ≥ 长度；空切片取 `s[0]` | 先判断 `len(s) > i` |
| `panic: runtime error: slice bounds out of range [:5] with capacity 3` | 切片表达式越界 | 满足 `0 <= low <= high <= cap(s)` |
| `panic: assignment to entry in nil map` | map 未 `make` 就写入（读取 nil map 安全） | `m = make(map[string]int)` 或字面量初始化 |
| `panic: runtime error: integer divide by zero` | 整数除数为 0（浮点除零得到 `Inf`） | 先判断 `b == 0` 并返回 error |
| `panic: close of closed channel` / `send on closed channel` | 重复关闭，或向已关闭 channel 发送 | 只由发送方关闭一次；接收方不关闭 |
| `fatal error: all goroutines are asleep - deadlock!` | 所有 goroutine 都在等 channel/锁 | 检查发送/接收配对、`Add`/`Done`；加缓冲或超时 |
| `panic: interface conversion: interface {} is string, not int` | 类型断言失败 | 用 `n, ok := v.(int)` 或 `type switch` |
| `fatal error: concurrent map writes` | 多 goroutine 同时读写同一 map，进程直接终止 | `Mutex` / `RWMutex` 保护；`recover` 无效 |
| `panic: runtime error: comparing uncomparable type []int` | 对含切片/map/函数的接口值做 `==`，或用作 map key | 切片用 `slices.Equal`，map 用 `maps.Equal` |

**并发写 map 的标准修复**：

```go
var mu sync.Mutex

mu.Lock()
m[k] = v
mu.Unlock()
```

---

## B.4 切片、map、字符串

| 现象 | 原因 | 修复 |
|------|------|------|
| 函数内 `append` 后调用方长度不变 | 切片是值传递，传的是切片头（指针、len、cap）的拷贝 | 返回新切片 `return append(s, x)`，或传 `*[]int` |
| 修改子切片影响原切片 | `b := a[:2]` 与原切片共享底层数组 | 需要独立副本用 `slices.Clone(a)` 或 `copy` |
| `append` 后有时共享、有时不共享 | 取决于 `cap` 是否足够，扩容后会指向新数组 | 不要依赖共享行为，需要独立就显式复制 |
| nil 切片 JSON 为 `null`，空切片为 `[]` | `var a []T` 是 nil；`[]T{}` 已初始化 | API 要求数组时用 `make([]T, 0)` |
| `len("中文") == 6`，按字符却想得 2 | 字符串是 UTF-8 字节序列 | 用 `len([]rune(s))` 或 `for range` |
| `s[0]` 得到 228，不是「中」 | 索引取的是 `byte` | 需要字符用 `[]rune(s)`；`for range` 按 rune |
| `cannot assign to s[0]` | 字符串不可变 | 转 `[]byte` / `[]rune` 修改后转回 |
| map 遍历顺序每次不同 | Go 故意随机化遍历起点 | 需要顺序时取 key 后 `slices.Sort` |
| `invalid map key type []int` | 切片、map、函数不可比较 | 用 string/整数/数组/全字段可比较的结构体作 key |

**append 不影响调用方的示例**：

```go
func add(s []int) { s = append(s, 1) }

s := []int{}
add(s)
fmt.Println(len(s)) // 0

// 正确写法
func add2(s []int) []int { return append(s, 1) }
s = add2(s)
```

---

## B.5 指针与 nil

| 现象 | 原因 | 修复 |
|------|------|------|
| 解引用 nil 指针 panic | 指针未初始化，或查询失败后继续使用 | 使用前判空；`if u == nil { ... }` |
| 循环变量取地址后全是同一个值 | Go 1.21 及以下复用循环变量 | 升级 `go 1.22+`，或循环内 `i := i` |
| GORM `Create(user)` 后 ID 没回填 | 传值无法可靠回填字段 | 统一传 `&user`；查询用 `First(&user, id)` |

**`new(T)` 与 `&T{}`**：两者都返回 `*T` 且字段为零值；`&User{Name: "张三"}` 更常用。返回局部变量地址是安全的，Go 会自动做逃逸分析。

---

## B.6 接口

| 现象 | 原因 | 修复 |
|------|------|------|
| `err != nil`，但内部指针是 nil | 接口值 = 动态类型 + 动态值，装入了 nil 指针后接口本身不为 nil | 返回错误时直接 `return nil`，不要把 nil 指针赋给接口 |
| `[]any` 不能接收 `[]string` | 切片类型不协变 | 逐元素转换 |
| 断言 `v.(int)` panic | 实际类型不是 int | `n, ok := v.(int)`；多类型用 `type switch` |

**nil 接口陷阱**：

```go
type MyError struct{}
func (e *MyError) Error() string { return "boom" }

var p *MyError = nil
var err error = p
fmt.Println(err != nil) // true
```

---

## B.7 错误处理

| 现象 | 原因 | 修复 |
|------|------|------|
| `err == io.EOF` 判断不到包装错误 | `%w` 包装后原错误在链中，指针不相等 | `errors.Is(err, io.EOF)` |
| `errors.As(err, target)` 不生效甚至 panic | 第二个参数必须是错误类型指针的地址 | 写 `errors.As(err, &target)` |
| `errors.Is` 找不到自定义错误 | 用了 `%v` 而非 `%w`，链断了 | `fmt.Errorf("上下文: %w", err)` |
| `err` 一直是 nil | `:=` 在 `if` 块内遮蔽了外层 err | 外层声明后用 `=`；或直接 `if err := f(); err != nil { return err }` |
| `defer resp.Body.Close()` 时 panic | 在检查 err 之前就 defer，resp 可能为 nil | 先 `if err != nil { return err }` 再 defer |

---

## B.8 并发

| 现象 | 原因 | 修复 |
|------|------|------|
| 闭包打印全是最后一个值 | Go 1.21 及以下循环变量复用 | `go 1.22+`；或 `v := v`；或作为参数传入 |
| goroutine 永远阻塞、泄漏 | 无缓冲 channel 只发不收，或 context 未取消 | 确保有人接收；用缓冲、`select`、`context` |
| `wg.Wait()` 提前返回 | `Add` 写在 goroutine 内，可能晚于 `Wait` | 启动前 `wg.Add(1)` |
| `panic: sync: negative WaitGroup counter` | `Done()` 次数多于 `Add()` | 检查 Add/Done 一一对应 |
| `WARNING: DATA RACE` | 多 goroutine 无同步读写同一变量 | `Mutex` / `atomic` / channel；`go test -race` |
| 子 goroutine panic 导致进程退出 | `recover` 只对当前 goroutine 有效 | 每个 goroutine 自己 `defer recover` |
| 请求结束了下游还在跑 | context 没传递或被 `Background()` 截断 | 全链路传 `ctx`；`defer cancel()` |

---

## B.9 模块与依赖

| 错误信息 | 原因 | 修复 |
|----------|------|------|
| `no required module provides package ...` | 依赖未加入 `go.mod` | `go get github.com/foo/bar`，再 `go mod tidy` |
| `missing go.sum entry ...` | `go.sum` 缺少校验记录 | `go mod tidy` 或 `go mod download`；不要手改 |
| `updates to go.mod needed` | `go.mod` 与代码导入不一致 | `go mod tidy`，提交更新后的两个文件 |
| `module ... found, but does not contain package ...` | 包路径写错，或版本没有该子包 | 核对仓库目录；`go list -m -versions` |
| `verifying ...: checksum mismatch` | 缓存损坏、代理不一致或依赖被篡改 | `go clean -modcache` 后重下；检查 `GOPROXY` |
| `module declares its path as .../v2 but was required as ...` | v2+ 主版本必须写在导入路径里 | 全部改成 `github.com/foo/bar/v2` |
| `inconsistent vendoring` | `vendor/` 与 `go.mod` 不一致 | `go mod vendor` 重新生成 |

---

## B.10 JSON、HTTP 与 SQL

| 错误/现象 | 原因 | 修复 |
|-----------|------|------|
| `json: cannot unmarshal string into Go struct field User.age of type int` | JSON 类型与字段类型不匹配 | 调整字段类型，或自定义 `UnmarshalJSON` |
| JSON 字段被静默忽略 | 字段小写未导出 | 改为大写 `Name string` |
| `json: unsupported type: func()` | 序列化了函数/channel 等 | 定义 DTO，或实现 `MarshalJSON` |
| `omitempty` 不生效 | 只省略零值；结构体不是零值 | 区分 0 与未设置用 `*int`；需要 0 就去掉 tag |
| `http: superfluous response.WriteHeader call` | 重复写状态码，如先 `c.Status` 再 `c.JSON` | 每请求只写一次；`c.JSON(code, obj)` 已含状态码 |
| `listen tcp :8080: bind: Only one usage ...` | 端口被占用 | 换端口；`netstat -ano \| findstr :8080` 后结束进程 |
| `context deadline exceeded` | 超时、下游无响应或 context 被取消 | 检查超时配置；把 `ctx` 继续传给下游 |
| `sql: no rows in result set` | `QueryRow.Scan` 没查到记录 | `errors.Is(err, sql.ErrNoRows)`，转 404 |
| `converting NULL to int is unsupported` | 数据库 NULL 扫描进非空字段 | 用 `sql.NullInt64`、`*int` 或设默认值 |
| MySQL `Error 1045` | 账号/密码错误 | 检查 DSN、权限、host |
| MySQL `Error 1062` | 唯一键冲突 | 捕获后转业务错误 |
| MySQL `Error 1054` | 未知列 | 检查字段名与迁移 |
| MySQL `invalid connection` | 连接被服务端断开 | 设置 `SetConnMaxLifetime`，重试 |
| GORM `record not found` | `First` 没查到 | `errors.Is(err, gorm.ErrRecordNotFound)` |
| GORM `WHERE conditions required` | 无条件批量更新/删除被阻止 | 加条件；确需全表用 `AllowGlobalUpdate` |
| GORM 更新零值无效 | `Updates(struct)` 忽略 `0`、`""`、`false` | 用 map，或 `Select("Name", "Age").Updates(...)` |

**GORM 更新零值**：

```go
db.Model(&user).Updates(User{Name: "", Age: 0}) // 零值被忽略

db.Model(&user).Updates(map[string]any{"name": "", "age": 0})
db.Model(&user).Select("Name", "Age").Updates(User{Name: "", Age: 0})
```

---

## B.11 测试

| 现象 | 原因 | 修复 |
|------|------|------|
| `testing: warning: no tests to run` | 文件名不是 `_test.go`、函数不是 `TestXxx(t *testing.T)` | 检查命名；`go test -v ./...` |
| `--- FAIL: TestAdd` 下只有一行 got/want | 断言失败 | 看 `文件:行`；`Errorf` 继续，`Fatalf` 停止当前测试 |
| `panic: test timed out after 10m0s` | 死锁、goroutine 泄漏、等待不会发生的事件 | `go test -timeout 30s`；检查 channel/`WaitGroup`/`context` |
| 子 goroutine 调用 `t.Fatalf` 行为异常 | `FailNow` 只能在运行测试函数的 goroutine 中调用 | 子 goroutine 用 `t.Errorf` 后 `return`，或把错误传回主 goroutine |

---

## B.12 关键字速查

| 关键字 | 小节 |
|--------|------|
| `undefined` / `declared and not used` | B.2 |
| `mismatched types` / `missing return` | B.2 |
| `nil pointer dereference` | B.3 |
| `index out of range` / `slice bounds` | B.3 |
| `nil map` / `deadlock` / `interface conversion` | B.3 |
| `concurrent map` / `DATA RACE` | B.3 / B.8 |
| `no required module` / `missing go.sum entry` | B.9 |
| `cannot unmarshal` / `WriteHeader` | B.10 |
| `no rows` / `record not found` | B.10 |
| `no tests to run` / `timed out` | B.11 |
