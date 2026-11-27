# 附录 A：Go 与 TypeScript 对照速查表

> **用途**：写 Go 时快速查「这件事在 TS 里怎么写，在 Go 里怎么写」。
> **约定**：代码只保留关键部分，省略 `import` 与完整 `main`；Go 基准为 1.22+。
> **配套章节**：第 01–14 章。

---

## A.1 总表

| 维度 | TypeScript | Go |
|------|------------|-----|
| 执行 | 编译成 JS，由 Node/浏览器运行 | 编译成机器码直接运行 |
| 类型 | 编译期检查，运行时擦除 | 编译期强类型，参与内存布局 |
| 转换 | 大量隐式兼容，`as` 是断言 | 几乎无隐式转换，必须 `T(v)` |
| 空值 | `undefined` / `null` | 零值；`nil` 只用于特定类型 |
| 变量 | `let` / `const` | `var` / `:=` / `const` |
| 类 | `class` | `struct` + 方法 |
| 继承 | `extends` | 嵌入 + 组合 |
| 接口 | `implements` 或结构化匹配 | 隐式实现，只看方法集 |
| 错误 | `throw` / `try/catch` | `error` 返回值 |
| 并发 | `Promise` / `async/await` | `goroutine` / `channel` |
| 可见性 | `export` | 首字母大小写 |
| 包管理 | `npm` / `package.json` | `go mod` / `go.mod` |
| 依赖校验 | `package-lock.json` | `go.sum` |
| 入口 | `package.json` 的 `main` | `package main` + `func main()` |
| 测试 | Vitest / Jest | `testing` + `go test` |
| 泛型 | `<T>`，运行时擦除 | `[T any]`，编译期检查 |

---

## A.2 类型与转换

| TypeScript | Go | 说明 |
|------------|-----|------|
| `number` | `int` / `int64` / `float64` | 区分整数/浮点、位数、符号 |
| `string` | `string` | 不可变字节序列，UTF-8 |
| `boolean` | `bool` | 没有真值转换 |
| `bigint` | `int64` / `math/big.Int` | 超出 64 位用 `math/big` |
| `any` | `any` | 空接口，可装任意值 |
| `unknown` | 无直接对应 | Go 没有强制收窄 |
| `void` | 无返回值 | 函数可以不返回 |
| `never` | 无直接对应 | 用 `panic` 或无限循环表达 |
| `T[]` | `[]T` | 切片；`[N]T` 才是定长数组 |
| `[T, U]` | `struct{ A T; B U }` | 没有元组 |
| `Record<string, T>` | `map[string]T` | key 必须可比较 |
| `Date` | `time.Time` | 标准库 `time` |
| `() => void` | `func()` | 函数类型 |
| `(a: T) => U` | `func(T) U` | 参数在前，返回值在后 |
| `T \| null` | `*T` / `sql.NullString` 等 | 没有联合类型 |
| `'a' \| 'b'` | 自定义类型 + 常量 | 没有字面量联合类型 |

```ts
let age = 25                 // number
const n = Number('42')
const u = obj as User        // 仅编译期断言
```

```go
age := 25                    // int
n, err := strconv.Atoi("42") // 可能失败
u, ok := obj.(User)          // 运行时检查

type ID = string             // 别名，与 string 完全等价
type UserID int64            // 新类型，需显式转换
```

Go 没有联合类型；有限取值用「自定义类型 + `const` + 校验」表达。`type A = B` 是别名，`type A B` 是新类型。

---

## A.3 变量、常量与空值

| 场景 | TypeScript | Go |
|------|------------|-----|
| 可变变量 | `let x = 1` | `x := 1` / `var x = 1` |
| 不可变绑定 | `const x = 1` | `const x = 1`（仅编译期常量） |
| 只声明 | `let x: number`（`undefined`） | `var x int`（`0`） |
| 包级变量 | 模块顶层 | 函数外 `var x = 1` |
| 解构/展开 | `{ ...obj }`、`[...arr]` | 无；手动复制或 `slices.Clone` |

| 类型 | 零值 |
|------|------|
| 数值 | `0` |
| `string` | `""` |
| `bool` | `false` |
| 指针、切片、map、channel、函数、接口 | `nil` |
| 结构体 | 每个字段各自取零值 |

```ts
let name: string | undefined
interface User { age?: number }
```

```go
type User struct {
    Age *int   // 表达「未设置」
}

var name string  // ""，不是 nil
var age int      // 0
// var n int = nil // 编译错误
```

要点：Go 没有 `undefined`，声明即初始化；可选值用 `*T`、`sql.NullInt64` 或额外布尔字段；GORM 中 `Age int` 为 `0` 时 `Updates(struct)` 会忽略它；条件必须显式写 `name != ""`。

---

## A.4 函数

```ts
function add(a: number, b: number): number { return a + b }
const add2 = (a: number, b: number) => a + b
```

```go
func add(a, b int) int { return a + b }
var add2 = func(a, b int) int { return a + b }
```

| 特性 | TypeScript | Go |
|------|------------|-----|
| 默认参数 | `function f(x = 1)` | 不支持；用 Options |
| 可选参数 | `function f(x?: number)` | 不支持；用指针/零值/Options |
| 剩余参数 | `...xs: number[]` | `xs ...int` |
| 展开调用 | `f(...arr)` | `f(arr...)` |
| 重载 | 支持 | 不支持，改函数名或参数结构 |
| 多返回值 | 返回对象/元组 | 原生多返回值 |

```go
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("除数不能为 0")
    }
    return a / b, nil
}

result, err := divide(10, 2)
if err != nil {
    return err
}
```

闭包捕获的是变量本身：`func counter() func() int { n := 0; return func() int { n++; return n } }`。

---

## A.5 结构体、类与方法

```ts
class User {
  constructor(public id: number, public name: string) {}
  greet() { return `你好，${this.name}` }
}
```

```go
type User struct {
    ID   int
    Name string
}

func NewUser(id int, name string) *User { return &User{ID: id, Name: name} }
func (u User) Greet() string            { return "你好，" + u.Name }
```

| TypeScript | Go |
|------------|-----|
| `class` / 字段 | `struct` / 结构体字段 |
| `constructor` | 惯例 `NewXxx` |
| `this.name` | `u.Name`（显式接收者） |
| `private` / `public` | 小写不导出 / 大写导出 |
| `static` 方法 | 普通包级函数 |
| `readonly` | 无直接对应 |

**值语义是最大差异**：TS 对象赋值共享引用；Go 结构体赋值复制。

```go
a := User{Name: "张三"}
b := a
b.Name = "李四"
fmt.Println(a.Name) // 张三
```

需要共享修改时传 `*User`。

| 接收者 | 写法 | 能改原值 | 方法集 |
|--------|------|----------|--------|
| 值 | `func (u User) M()` | 否 | `T` 和 `*T` |
| 指针 | `func (u *User) M()` | 是 | 只有 `*T` |

```go
func (c *Counter) Inc() { c.Count++ } // 有效
func (c Counter) Bad()  { c.Count++ } // 改的是副本
```

---

## A.6 接口与多态

```ts
interface Shape { area(): number }
class Circle implements Shape {
  constructor(private r: number) {}
  area() { return Math.PI * this.r ** 2 }
}
```

```go
type Shape interface { Area() float64 }

type Circle struct{ R float64 }
func (c Circle) Area() float64 { return math.Pi * c.R * c.R }

var s Shape = Circle{R: 2} // 无需 implements
```

| TypeScript | Go |
|------------|-----|
| `interface` | `interface` |
| `implements` | 方法签名匹配即实现 |
| `instanceof` | 类型断言 / `type switch` |
| `as T` | `x.(T)`（运行时检查） |
| `any` | `any` |
| `unknown` | 无直接对应 |

```go
s, ok := v.(string) // 安全断言；v.(string) 失败会 panic
```

**nil 接口陷阱**：接口值 = 动态类型 + 动态值。

```go
var p *MyError = nil
var err error = p
fmt.Println(err == nil) // false
```

错误处理中直接 `return nil`，不要把 nil 具体指针装进接口。

---

## A.7 继承与嵌入

```ts
class Animal {
  constructor(public name: string) {}
  speak() { return `${this.name} 发出声音` }
}
class Dog extends Animal {
  speak() { return `${this.name} 汪汪` }
}
```

```go
type Animal struct{ Name string }
func (a Animal) Speak() string { return a.Name + " 发出声音" }

type Dog struct{ Animal } // 匿名字段：嵌入
func (d Dog) Fetch() string { return "捡球" }

d := Dog{Animal{Name: "旺财"}}
d.Speak()          // 方法提升
d.Animal.Speak()   // 显式调用被遮蔽的方法
```

| TypeScript | Go |
|------------|-----|
| `extends` | 嵌入匿名字段 |
| `super.m()` | `d.Animal.m()` |
| 方法重写/动态分派 | 同名方法遮蔽；无虚函数 |
| `instanceof` | 类型断言 |
| 多继承 | 可嵌入多个类型，注意同名冲突 |

嵌入是组合，不是继承；`Dog` 不能赋给 `Animal`，多态要靠接口。

---

## A.8 错误处理

```ts
try {
  const user = await getUser(id)
} catch (err) {
  console.error('查询失败', err)
}
```

```go
user, err := getUser(ctx, id)
if err != nil {
    return fmt.Errorf("查询用户 %d: %w", id, err)
}
```

| 场景 | TypeScript | Go |
|------|------------|-----|
| 抛出 | `throw` | `return err` |
| 捕获 | `try/catch` | `if err != nil` |
| 自定义错误 | `class X extends Error` | 实现 `Error() string` |
| 类型判断 | `instanceof` | `errors.As` |
| 哨兵判断 | `err === ErrX` | `errors.Is` |
| 包装 | `new Error(msg, { cause })` | `fmt.Errorf("...: %w", err)` |
| 清理 | `finally` | `defer` |
| 顶层兜底 | `uncaughtException` | `recover`（同 goroutine） |

```go
type NotFoundError struct{ ID int }
func (e *NotFoundError) Error() string { return fmt.Sprintf("用户 %d 不存在", e.ID) }

var target *NotFoundError
if errors.As(err, &target) { fmt.Println(target.ID) }
```

`errors.Is` 判断哨兵错误（`io.EOF`、`gorm.ErrRecordNotFound`），`errors.As` 提取具体错误类型。

---

## A.9 并发

```ts
const [a, b] = await Promise.all([fetchA(), fetchB()])
```

```go
var wg sync.WaitGroup
wg.Add(2)
go func() { defer wg.Done(); a = fetchA() }()
go func() { defer wg.Done(); b = fetchB() }()
wg.Wait()
```

| 概念 | TypeScript | Go |
|------|------------|-----|
| 并发单元 | `Promise` / `async` | `goroutine` |
| 等待结果 | `await` | `<-ch` / `wg.Wait()` |
| 等待全部 | `Promise.all` | `WaitGroup` / `errgroup` |
| 竞速 | `Promise.race` | `select` |
| 超时/取消 | `AbortController` | `context.WithTimeout` |
| 定时器 | `setInterval` | `time.Ticker` |
| 共享状态 | 单线程，通常无需锁 | `Mutex` / `atomic` |
| 竞争检测 | 基本不存在 | `go test -race` |

```go
ch := make(chan int, 2)
ch <- 1
ch <- 2
close(ch)

v, ok := <-ch // 1, true
v, ok = <-ch  // 2, true
v, ok = <-ch  // 0, false：已关闭且取完

ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()
req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
resp, err := http.DefaultClient.Do(req)
```

---

## A.10 包与模块

| 概念 | TypeScript / Node | Go |
|------|-------------------|-----|
| 模块单位 | 一个文件 | 一个目录（同包） |
| 声明 | 无 | 文件顶部 `package xxx` |
| 导入 | `import { x } from './x'` | `import "module/path/pkg"` |
| 默认导出 | `export default` | 无 |
| 私有 | 未导出 | 首字母小写 |
| 依赖清单 | `package.json` | `go.mod` |
| 依赖校验 | `package-lock.json` | `go.sum` |
| 依赖目录 | `node_modules` | 模块缓存 / `vendor` |
| 安装 | `npm install lodash` | `go get github.com/xxx/yyy` |
| 清理 | `npm prune` | `go mod tidy` |
| 工作区 | npm workspaces | `go work` |
| 版本范围 | `^1.2.3` | 最小版本选择，记录精确版本 |

```go
import (
    "os"
    "github.com/gin-gonic/gin"
    _ "github.com/go-sql-driver/mysql" // 仅副作用
)
```

- 导入的是包路径，不是文件路径；未使用的导入是编译错误；
- 主版本 v2+ 写在路径里：`github.com/foo/bar/v2`。

---

## A.11 可见性与命名

| 场景 | TypeScript | Go |
|------|------------|-----|
| 导出类型 | `export class User` | `type User struct` |
| 导出函数 | `export function Add()` | `func Add()` |
| 包内私有 | 不加 `export` | 首字母小写 |
| 类私有字段 | `private age` | `age int` |
| 接口命名 | `UserService` | `UserService` / `Reader` |
| 错误变量 | `const ErrNotFound` | `var ErrNotFound = errors.New(...)` |

命名习惯：`ID`、`URL`、`HTTP` 全大写；不用 `snake_case`；包名小写简短；构造函数 `NewXxx`。

---

## A.12 测试

```ts
import { describe, it, expect } from 'vitest'
describe('add', () => {
  it('两数相加', () => { expect(add(1, 2)).toBe(3) })
})
```

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name string
        a, b int
        want int
    }{
        {"正数", 1, 2, 3},
        {"负数", -1, 1, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := add(tt.a, tt.b); got != tt.want {
                t.Errorf("add(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
            }
        })
    }
}
```

| 概念 | TypeScript | Go |
|------|------------|-----|
| 测试文件 | `*.test.ts` | `*_test.go` |
| 测试函数 | `it(...)` | `func TestXxx(t *testing.T)` |
| 分组 | `describe` | `t.Run` |
| 断言 | `expect(...).toBe(...)` | `if got != want { t.Errorf(...) }` |
| 失败即停 | 抛异常 | `t.Fatalf` |
| 清理 | `afterEach` | `t.Cleanup` |
| 临时目录 | `fs.mkdtemp` | `t.TempDir` |
| HTTP 测试 | supertest | `net/http/httptest` |
| 基准 | 第三方库 | `func BenchmarkXxx(b *testing.B)` |
| 覆盖率 | `--coverage` | `go test -cover` |
