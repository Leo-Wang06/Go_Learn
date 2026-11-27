# 第 04 章：复合类型 —— 数组、切片、map、字符串

> **阶段**：阶段一 · 基础语法
> **建议时间**：1–2 天
> **前置知识**：第 01–03 章（变量、类型、零值、`range`、值传递）
> **代码基准**：Go 1.22 及以上

---
## 本章目标

学完本章，你应当能够：

1. 说清数组与切片的区别，知道什么时候用哪个；
2. 熟练使用 `make`、`append`、`copy`、`len`、`cap`；
3. 解释切片共享底层数组的现象，以及扩容后为什么不共享；
4. 会用三索引切片 `s[low:high:max]` 控制容量；
5. 熟练使用 map，理解 comma-ok、遍历无序与并发写 panic；
6. 理解 `string` 是只读字节序列，分清 `byte`、`rune` 与 UTF-8；
7. 能在 `string`、`[]byte`、`[]rune` 之间正确转换；
8. 会用 `strings.Builder` 高效拼接字符串；
9. 看懂 Gin/GORM 文档中的 `[]User`、`map[string]any`、`[]byte`。

---

## 4.0 先建立一张地图

| 类型 | 定位 | 长度 | 零值 | 赋值/传参 | 典型用途 |
|------|------|------|------|-----------|----------|
| 数组 `[N]T` | 固定长度序列 | 编译期固定 | 全零元素 | 整体拷贝 | 固定尺寸数据 |
| 切片 `[]T` | 数组视图 | 运行时可变 | `nil` | 拷贝切片头，共享底层数组 | 列表、查询结果 |
| map `map[K]V` | 键值对集合 | 运行时可变 | `nil` | 拷贝引用，共享同一张表 | 字典、计数、set |
| string | 只读字节序列 | 固定 | `""` | 拷贝字符串头，共享字节 | 文本、参数 |

与 TS 的对应：`Array<T>` ↔ `[]T`，`Record<string,T>` ↔ `map[string]T`，`string` ↔ `string`。关键差异：Go 切片有 `len/cap` 且 `append` 返回新切片；Go map 遍历无序；Go 字符串按 UTF-8 字节计数，JS 按 UTF-16 码元计数。

---

## 4.1 数组：固定长度、值类型

### 4.1.1 定义、零值与长度

数组长度写在类型里，`[3]int` 和 `[4]int` 是**两个不同类型**：
```go
var a [3]int        // [0 0 0]
fmt.Println(len(a)) // 3
fmt.Println(cap(a)) // 3，数组的 cap 永远等于 len

var b [4]int
a = b // 编译错误：cannot use b (variable of type [4]int) as [3]int value
```
TS 的 `number[]` 长度不是类型的一部分；Go 数组长度是。

### 4.1.2 初始化与遍历

```go
a := [3]int{1, 2, 3}      // [1 2 3]
b := [...]int{1, 2, 3, 4} // b 是 [4]int，长度由编译器推断
c := [5]int{1: 10, 3: 30} // [0 10 0 30 0]

for i, v := range a {
    fmt.Println(i, v) // 0 1 / 1 2 / 2 3
}
for i := range a {
    a[i] *= 2 // 正确：通过下标修改
}
```
`[...]T{...}` 只用于数组，切片不能写 `[...]T`。

### 4.1.3 比较与值语义

只要元素可比较，数组就可以用 `==`、`!=` 逐元素比较，例如 `[2]int{1,2} == [2]int{1,2}` 为 `true`。长度不同的数组类型不同，不能比较。

赋值和传参都会拷贝整个数组：

```go
func modify(a [3]int)     { a[0] = 100 }
func modifyPtr(a *[3]int) { a[0] = 100 }

func main() {
    a := [3]int{1, 2, 3}
    b := a
    b[0] = 100
    fmt.Println(a, b) // [1 2 3] [100 2 3]

    modify(a)
    fmt.Println(a) // [1 2 3]，传的是副本
    modifyPtr(&a)
    fmt.Println(a) // [100 2 3]
}
```
数组越大拷贝成本越高，所以日常「列表」基本都用切片。数组常见于固定长度缓冲区、map key、切片底层存储。

---

## 4.2 切片：可变长度的数组视图

### 4.2.1 切片头

切片变量是一个**切片头**，包含三部分：指向底层数组的指针、长度 `len`、容量 `cap`。

```
底层数组: [ 0 | 1 | 2 | 3 | 4 | 5 ]
              ^
切片 s:       ptr 指向这里，len=3，cap=5
```
切片赋值、传参拷贝的是这个头（三个字段），底层数组仍然共享。这是所有共享陷阱的根源。

### 4.2.2 声明、字面量与 `make`

```go
var a []int         // nil 切片，len=0 cap=0
b := []int{}        // 空切片，非 nil，len=0 cap=0
c := []int{1, 2, 3} // len=3 cap=3

fmt.Println(a == nil, b == nil) // true false
fmt.Println(len(c), cap(c))     // 3 3

s1 := make([]int, 3)    // [0 0 0]，len=3 cap=3
s2 := make([]int, 3, 5) // [0 0 0]，len=3 cap=5
fmt.Println(s1, s2)     // [0 0 0] [0 0 0]
```
规则：`make([]T, len)` 容量等于长度；`make([]T, len, cap)` 要求 `cap >= len`；长度内元素初始化为零值。`make([]T, 0, 5)` 才是「预留 5 个位置、当前 0 个元素」；`make([]int, 5)` 是「已经有 5 个 0」。

### 4.2.3 `len` 与 `cap`

```go
s := make([]int, 2, 5)
fmt.Println(len(s), cap(s)) // 2 5
```
`len` 是当前可读写元素个数，合法下标 `0` 到 `len-1`；`cap` 是从起始位置到底层数组末尾的元素个数；`cap >= len` 永远成立。访问 `s[len(s)]` 会 panic，即使还有容量。要使用剩余容量必须 `append`。

### 4.2.4 `append` 与扩容

`append` 返回新切片，**必须接住返回值**：

```go
s := []int{1, 2}
s = append(s, 3)       // [1 2 3]
s = append(s, 4, 5, 6) // [1 2 3 4 5 6]

b := []int{7, 8}
s = append(s, b...) // 展开另一个切片

buf := []byte("hello")
buf = append(buf, " world"...)
fmt.Println(string(buf)) // hello world
```
扩容规则（只记结论）：

- 容量够时直接写底层数组，返回的切片仍与原切片共享；
- 容量不够时分配更大的数组，拷贝旧元素，返回的切片指向新数组；
- 小切片大致按 2 倍增长，大切片约 1.25 倍，最后按内存规格取整；
- 不要依赖具体容量数字，只依赖 `cap >= len`。

```go
s := []int{1, 2, 3}
fmt.Println(len(s), cap(s)) // 3 3
s = append(s, 4)
fmt.Println(len(s), cap(s)) // 4 6
s = append(s, 5)
fmt.Println(len(s), cap(s)) // 5 6
s = append(s, 6)
fmt.Println(len(s), cap(s)) // 6 6
s = append(s, 7)
fmt.Println(len(s), cap(s)) // 7 12
```
具体数字可能随版本和元素类型变化，但「不够就扩、扩了就不共享」不变。

### 4.2.5 `copy`

```go
src := []int{1, 2, 3}
dst := make([]int, 5)

n := copy(dst, src)
fmt.Println(n, dst) // 3 [1 2 3 0 0]
```
规则：复制 `min(len(dst), len(src))` 个元素；返回实际复制个数；不会自动扩容 `dst`；源和目标可以重叠，`copy` 会正确处理。克隆用 `clone := make([]int, len(src)); copy(clone, src)`；`copy` 的源也可以是字符串，例如 `copy(buf, "hello")`。

### 4.2.6 nil 切片 vs 空切片

```go
var a []int  // nil
b := []int{} // 非 nil，空

fmt.Println(a == nil, b == nil) // true false
fmt.Println(len(a), len(b))     // 0 0
fmt.Println(a, b)               // [] []
```
两者在 `len`、`range`、`append`、`copy` 上行为一致，关键差异在 JSON：

```go
var a []int
b := []int{}
ja, _ := json.Marshal(a)
jb, _ := json.Marshal(b)
fmt.Println(string(ja)) // null
fmt.Println(string(jb)) // []
```
这是 API 高频坑：数据库没查到时，`var users []User` 直接 `c.JSON(200, users)` 会返回 `null`。应初始化 `users := make([]User, 0)`。

### 4.2.7 切片不能直接比较

```go
a := []int{1, 2}
b := []int{1, 2}
fmt.Println(a == nil) // false
// fmt.Println(a == b) // 编译错误：slice can only be compared to nil
```
比较内容用循环或 `slices.Equal(a, b)`（Go 1.21+）。

---

## 4.3 切片表达式与共享底层数组

### 4.3.1 `s[low:high]`

左闭右开 `[low, high)`：

```go
s := []int{0, 1, 2, 3, 4, 5}
a := s[1:4]

fmt.Println(a)      // [1 2 3]
fmt.Println(len(a)) // 3
fmt.Println(cap(a)) // 5 = cap(s) - low
```
省略形式：`s[:3]`、`s[3:]`、`s[:]`。`s[:]` 不是复制，仍共享。边界：对切片 `0 <= low <= high <= cap(s)`；`high` 可以超过 `len(s)`，只要不超过 `cap(s)`；越界 panic。例如 `make([]int, 2, 5)[0:4]` 合法，`[0:6]` 会 panic。

### 4.3.2 共享底层数组

```go
s := []int{1, 2, 3, 4, 5}
a := s[1:3] // [2 3]
b := s[2:4] // [3 4]

a[0] = 20
fmt.Println(s) // [1 20 3 4 5]
fmt.Println(a) // [20 3]
fmt.Println(b) // [20 4]
```
`a[0]` 和 `s[1]` 是同一个内存位置。业务中的「莫名其妙被改了」多半来自这里，例如函数内 `first := nums[:2]; first[0] = 0` 会无意中修改调用方的数据。不想影响原数据就先复制：`first := make([]int, 2); copy(first, nums[:2])`。

### 4.3.3 `append` 在容量内会覆盖原数组

```go
s := []int{1, 2, 3}
a := s[:2] // len=2 cap=3

a = append(a, 99) // 容量够，直接写入 s[2]
fmt.Println(s)    // [1 2 99]
fmt.Println(a)    // [1 2 99]

a = append(a, 100) // cap 不够，分配新数组
a[0] = 1000
fmt.Println(s) // [1 2 99]，不再受影响
fmt.Println(a) // [1000 2 99 100]
```
`s` 的长度仍是 3，但第三个元素已被第一次 `append` 改掉。记忆口诀：**容量够，共享；容量不够，分家。**

### 4.3.4 三索引切片 `s[low:high:max]`

```go
s := []int{1, 2, 3, 4, 5}
a := s[1:3:3] // len=2 cap=2

fmt.Println(len(a), cap(a)) // 2 2

a = append(a, 99) // cap 不够，强制分配新数组
fmt.Println(s)    // [1 2 3 4 5]
fmt.Println(a)    // [2 3 99]
```
规则：`low <= high <= max <= cap(s)`；新切片 `len = high - low`，`cap = max - low`；两索引形式等价于 `max = cap(s)`。典型用途：把切片交给别人时，禁止对方通过 `append` 覆盖你后面的数据，例如 `safeView := s[1:3:3]`，对方 `append` 必定扩容，不会改到 `s[3]`。

### 4.3.5 字符串切片取的是字节

```go
s := "Go中文"

fmt.Println(s[0:2]) // Go
fmt.Println(s[2:5]) // 中

fmt.Println(utf8.ValidString(s[2:5])) // true
fmt.Println(utf8.ValidString(s[2:4])) // false，只取了 2 个字节
```
字符串切片不复制底层字节，但字符串不可变，所以不存在「改一个影响另一个」。

---

## 4.4 切片作为函数参数

函数拿到的是切片头的副本（指针、len、cap 各一份），指针指向同一个底层数组。

```go
func modify(s []int) {
    s[0] = 100
}

func main() {
    nums := []int{1, 2, 3}
    modify(nums)
    fmt.Println(nums) // [100 2 3]
}
```
修改元素会影响调用方，但函数内 `append` 不一定：

```go
func appendOne(s []int) {
    s = append(s, 4)
    fmt.Println("函数内:", s, len(s), cap(s))
}

func main() {
    nums := []int{1, 2, 3}
    appendOne(nums)
    fmt.Println("函数外:", nums, len(nums), cap(nums))
}
```
输出（容量足够时）：函数内 `[1 2 3 4] 4 6`，函数外 `[1 2 3] 3 3`。

原因有两层：`append` 返回新的切片头，函数内重新赋值，调用方的头不受影响；即使没有扩容，函数内 `len` 变成 4，调用方的 `len` 仍是 3。希望调用方拿到结果，就返回新切片：

```go
func appendOne(s []int) []int { return append(s, 4) }
nums = appendOne(nums)
```
---

## 4.5 二维切片

二维切片是「元素为切片的切片」：`[][]int`。

```go
grid := make([][]int, 3) // 3 行，每行是 nil
for i := range grid {
    grid[i] = make([]int, 4) // 每行 4 列
}

grid[0][0] = 1
fmt.Println(grid) // [[1 0 0 0] [0 0 0 0] [0 0 0 0]]
```
只写 `make([][]int, 3)` 时，`grid[0]` 是 nil 切片，直接 `grid[0][0] = 1` 会 panic（index out of range）。每行长度也可以不同，例如 `[][]int{{1}, {1, 1}, {1, 2, 1}}`；内层类型在字面量里可以省略。二维切片仍遵循共享规则：内层切片是从别处切出来的，就会共享。

---

## 4.6 map：键值对集合

### 4.6.1 声明、`make` 与增删查改

```go
m := make(map[string]int) // 空 map，可读写
m["apple"] = 3
m["banana"] = 5

m["a"] = 1           // 增
m["a"] = 2           // 改
fmt.Println(m["a"])  // 2
delete(m, "a")       // 删
delete(m, "missing") // 删不存在的 key 是安全的
fmt.Println(len(m))  // 2
```
nil map 可以读、`len`、`range`，但**不能写**：

```go
var m map[string]int
fmt.Println(m["a"]) // 0，读不存在的 key 返回零值
fmt.Println(len(m)) // 0
m["a"] = 1          // panic: assignment to entry in nil map
```
### 4.6.2 comma-ok：判断 key 是否存在

直接读不存在的 key 返回零值，无法区分「不存在」和「值就是零值」：

```go
m := map[string]int{"zero": 0}
fmt.Println(m["zero"], m["missing"]) // 0 0，看起来一样

v, ok := m["zero"]
fmt.Println(v, ok) // 0 true
v, ok = m["missing"]
fmt.Println(v, ok) // 0 false
```
只关心存在性：`if _, ok := m["key"]; ok { ... }`。

### 4.6.3 遍历无序

```go
m := map[string]int{"a": 1, "b": 2, "c": 3}
for k, v := range m {
    fmt.Println(k, v)
}
```
可能输出 `b 2 / a 1 / c 3`，下次可能不同。Go 故意随机化 map 遍历顺序，防止程序依赖「碰巧的顺序」。需要稳定顺序，先收集 key 再排序：

```go
keys := make([]string, 0, len(m))
for k := range m {
    keys = append(keys, k)
}
sort.Strings(keys)
for _, k := range keys {
    fmt.Println(k, m[k]) // a 1 / b 2 / c 3
}
```
### 4.6.4 key 必须可比较，value 任意

key 必须是可比较类型：布尔、数值、字符串、指针、channel、接口，以及由这些组成的数组/结构体。切片、map、函数不能作 key。

```go
type Point struct{ X, Y int }

m := map[Point]string{{1, 2}: "a"} // 合法
fmt.Println(m[Point{1, 2}])        // a

// var bad map[[]int]string // 编译错误：invalid map key type []int
```
数组也可以作 key，例如 `map[[2]int]string`。value 可以是任意类型：

```go
m := map[string][]int{"a": {1, 2, 3}}
m["b"] = []int{4, 5}
fmt.Println(m) // map[a:[1 2 3] b:[4 5]]
```
### 4.6.5 map 元素不可寻址

```go
type User struct{ Name string }
m := map[int]User{1: {Name: "张三"}}

m[1].Name = "李四" // 编译错误：cannot assign to struct field m[1].Name in map
// p := &m[1]      // 编译错误：cannot take address of m[1]
```
正确做法是取出、修改、放回：

```go
u := m[1]
u.Name = "李四"
m[1] = u
fmt.Println(m[1].Name) // 李四
```
map 元素不可寻址，是因为扩容时元素会搬家，地址不稳定（第 06 章讲「可寻址」）。切片作为 value 时也不能直接写 `m["a"][0] = 10`，要先取出切片、修改、再赋回；`m["a"] = append(m["a"], 4)` 是合法的。

### 4.6.6 map 是引用语义

map 变量内部指向运行时哈希表，复制 map 变量复制的是引用。函数内增删改会影响调用方，例如 `func set(m map[string]int) { m["a"] = 1 }` 调用后原 map 能看到 `a`；但 `m = make(...)` 只改函数内的副本。

### 4.6.7 并发读写会 fatal

map 不是并发安全的。多个 goroutine 同时写，或一边读一边写，运行时会报致命错误：

```go
func main() {
    m := make(map[int]int)
    var wg sync.WaitGroup

    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(n int) { defer wg.Done(); m[n] = n }(i)
    }
    wg.Wait()
}
```
可能输出：

```
fatal error: concurrent map writes
```
要点：这是 `fatal error`，不是普通 panic，`recover` 捕获不了；也可能报 `concurrent map read and map write`；不是每次都能复现，但一旦发生程序就崩溃。这里的 `sync.WaitGroup` 只是为了让 goroutine 跑完，第 11 章系统讲。

修复方向：加锁，或用 `sync.Map`。

```go
var mu sync.Mutex
mu.Lock()
m[n] = n
mu.Unlock()
```
`sync.Map` 提供 `Load`、`Store`、`LoadOrStore`、`Delete`、`Range`，不是 `m[k]` 语法，适合读多写少、key 稳定的场景。

### 4.6.8 set 与词频统计

Go 没有内置 `Set`，用 `map[T]struct{}`，`struct{}` 是零大小类型：

```go
func unique(words []string) []string {
    seen := make(map[string]struct{})
    result := make([]string, 0, len(words))
    for _, w := range words {
        if _, ok := seen[w]; ok {
            continue
        }
        seen[w] = struct{}{}
        result = append(result, w)
    }
    return result
}

fmt.Println(unique([]string{"go", "ts", "go", "js", "ts"})) // [go ts js]
```
词频统计：

```go
text := "go is fun go is fast"
counts := make(map[string]int)
for _, w := range strings.Fields(text) {
    counts[w]++
}
fmt.Println(counts) // map[fast:1 fun:1 go:2 is:2]
```
`counts[w]++` 能工作，是因为读不存在的 key 返回 0，加一后写回。

---

## 4.7 字符串：只读的字节序列

### 4.7.1 不可变与 UTF-8

```go
s := "hello"
// s[0] = 'H' // 编译错误：cannot assign to s[0]
```
字符串一旦创建，内容不能修改；所有「修改」都会产生新字符串。Go 字符串字面量默认 UTF-8 编码：

| 字符 | 码点 | UTF-8 字节 | 字节数 |
|------|------|------------|--------|
| `A` | U+0041 | `41` | 1 |
| `中` | U+4E2D | `E4 B8 AD` | 3 |
| `😀` | U+1F600 | `F0 9F 98 80` | 4 |

```go
s := "A中😀"
fmt.Println(len(s))         // 8 = 1 + 3 + 4
fmt.Println(len([]rune(s))) // 3
```
### 4.7.2 `len(s)` 是字节数

```go
fmt.Println(len("Go"))   // 2
fmt.Println(len("中文")) // 6
fmt.Println(len("😀"))   // 4

fmt.Println(utf8.RuneCountInString("Go中文")) // 4，按码点数
fmt.Println(len([]rune("Go中文")))            // 4
```
不要用 `len(s)` 判断「有几个字」，中文字符串的字节数通常大于字符数。

### 4.7.3 索引得到 `byte`

```go
s := "Go中文"

fmt.Printf("%T %v\n", s[0], s[0]) // uint8 71
fmt.Printf("%c\n", s[0])          // G
fmt.Println(s[2])                 // 228，「中」的第一个字节
```
`s[i]` 的类型是 `byte`（`uint8`），不是 `rune`，也不是「一个字符」。按字节遍历 `for i := 0; i < len(s); i++` 会得到 `71 111 228 184 173 230 150 135`，这是 UTF-8 原始字节，不是可读字符。

### 4.7.4 `rune` 与 `for range`

`rune` 是 `int32` 的别名，表示一个 Unicode 码点：

```go
var r rune = '中'
fmt.Println(r)        // 20013
fmt.Printf("%c\n", r) // 中
fmt.Printf("%T\n", r) // int32
```
单引号是 rune 字面量，双引号是字符串字面量：`'中'` 是 rune，`"中"` 是 3 字节字符串。

```go
s := "Go中文"
for i, r := range s {
    fmt.Printf("字节索引=%d 字符=%c 码点=%d\n", i, r, r)
}
```
输出：

```
字节索引=0 字符=G 码点=71
字节索引=1 字符=o 码点=111
字节索引=2 字符=中 码点=20013
字节索引=5 字符=文 码点=25991
```
索引是每个 rune 的**起始字节索引**，所以是 0、1、2、5。字符串含非法 UTF-8 字节时，`range` 会产出 `unicode.ReplacementChar`（U+FFFD，显示为 `�`），宽度按 1 继续。

### 4.7.5 常用字符串操作

```go
s := "  Go is fun  "

fmt.Println(strings.Contains(s, "Go"))       // true
fmt.Println(strings.HasPrefix(s, "  Go"))    // true
fmt.Println(strings.Index(s, "is"))          // 5
fmt.Println(strings.TrimSpace(s))            // Go is fun
fmt.Println(strings.ReplaceAll(s, " ", "-")) // --Go-is-fun--
fmt.Println(strings.ToUpper("go"))           // GO
fmt.Println(strings.Repeat("ab", 3))         // ababab

parts := strings.Split("a,b,c", ",")
fmt.Println(parts)                             // [a b c]
fmt.Println(strings.Join(parts, "-"))          // a-b-c
fmt.Println(strings.Fields("  go   is  fun ")) // [go is fun]
```
`strings.Fields` 按任意空白拆分并去掉多余空白，比 `Split(s, " ")` 实用。字符串比较是逐字节字典序：`"apple" < "banana"` 为 `true`，`==` 比较内容，`strings.EqualFold("Go", "GO")` 忽略大小写。

---

## 4.8 `string`、`[]byte`、`[]rune` 的转换

### 4.8.1 `[]byte(s)` 与 `[]rune(s)`

```go
s := "Go中文"
b := []byte(s)
fmt.Println(len(b))    // 8
fmt.Printf("% x\n", b) // 47 6f e4 b8 ad e6 96 87

b[0] = 'g'
fmt.Println(string(b)) // go中文
fmt.Println(s)         // Go中文，原串不变

r := []rune(s)
fmt.Println(len(r)) // 4
fmt.Println(r)      // [71 111 20013 25991]

r[2] = '语'
fmt.Println(string(r)) // Go语文
```
`[]byte(s)` 复制字节，`[]rune(s)` 解码 UTF-8 并复制成 `int32` 序列；修改结果都不影响原字符串。

### 4.8.2 修改字符串

```go
// 按字节修改，适合 ASCII 或明确知道字节位置
b := []byte("hello")
b[0] = 'H'
fmt.Println(string(b)) // Hello

// 按字符修改，适合中文
r := []rune("Go中文")
r[2] = '语'
fmt.Println(string(r)) // Go语文
```
### 4.8.3 转换开销与选择

三种转换都会分配新内存：`[]byte(s)` 复制字节，`[]rune(s)` 解码 UTF-8，`string(b)` / `string(r)` 复制或编码，均为 O(n)。热路径里避免反复转换。只读遍历时 `for _, r := range s` 按字符不分配，`for i := 0; i < len(s); i++` 按字节不分配。

| 需求 | 用法 |
|------|------|
| 当只读字节处理 | `[]byte(s)`（复制） |
| 按字符处理、修改 | `[]rune(s)`（解码） |
| 字节切片转字符串 | `string(b)` |
| rune 切片转字符串 | `string(r)` |
| 数字转字符串 | `strconv.Itoa(n)` |
| 字符串转数字 | `strconv.Atoi(s)` |

### 4.8.4 `string(int)` 的陷阱

`string(65)` 不是 `"65"`，而是把 65 当作 Unicode 码点：

```go
fmt.Println(string(65))       // A
fmt.Println(string(0x4E2D))   // 中
fmt.Println(string(228))      // ä
fmt.Println(strconv.Itoa(65)) // 65
```
`string(byte(228))` 也是码点 U+00E4（`ä`），不是字节 `0xE4`。单个字节转字符串用 `string([]byte{b})`。

---

## 4.9 `strings.Builder`

### 4.9.1 为什么不用循环 `+=`

字符串不可变，每次 `+=` 都可能分配新字符串并复制旧内容，循环 n 次总复制量约为 O(n²)：

```go
s := ""
for i := 0; i < 10000; i++ {
    s += "a"
}
```
### 4.9.2 Builder 用法

```go
var b strings.Builder

b.WriteString("Go")
b.WriteString(" is ")
b.WriteString("fun")

fmt.Println(b.String()) // Go is fun
fmt.Println(b.Len())    // 9，字节数

b.Grow(10000) // 预分配，减少扩容
```
| 方法 | 作用 |
|------|------|
| `WriteString(s)` | 追加字符串 |
| `WriteByte(b)` | 追加一个字节 |
| `WriteRune(r)` | 追加一个 rune（UTF-8 编码） |
| `String()` | 取出结果 |
| `Len()` | 当前字节数 |
| `Reset()` | 清空复用 |
| `Grow(n)` | 预分配至少 n 字节 |

`Builder` 实现了 `io.Writer`，可以配合 `fmt.Fprintf(&b, "user=%s age=%d", "张三", 25)`。

### 4.9.3 性能对比

可以用 `time.Now()` 和 `time.Since(start)` 分别给 `+=` 与 `Builder` 计时：

```go
const n = 10000

start := time.Now()
s := ""
for i := 0; i < n; i++ { s += "a" }
fmt.Println("+= 耗时:", time.Since(start))

start = time.Now()
var b strings.Builder
b.Grow(n)
for i := 0; i < n; i++ { b.WriteString("a") }
fmt.Println("Builder 耗时:", time.Since(start))
```
具体耗时因机器而异，但通常 `Builder` 明显更快，差距随 n 增大而扩大；两者长度都应为 n。

### 4.9.4 注意事项

**不要复制已经使用过的 Builder：**

```go
var b strings.Builder
b.WriteString("a")

// c := b // 错误：复制了已使用的 Builder
```
`Builder` 内部有 `noCopy` 保护，复制后继续写可能 panic 或数据错乱。需要传递时传指针：`func write(b *strings.Builder) { b.WriteString("x") }`。

**拼接已有切片优先用 `strings.Join`**，例如 `strings.Join([]string{"go", "is", "fun"}, "-")` 得到 `go-is-fun`。`Join` 先算总长度，一次分配。只拼两三个短字符串时 `+` 更清晰；拼接字节用 `bytes.Buffer`。

---

## 4.10 与 TypeScript 的对照

| 操作 | TypeScript | Go |
|------|------------|-----|
| 创建数组 | `const a = [1,2,3]` | `a := []int{1,2,3}` |
| 追加 | `a.push(4)` | `a = append(a, 4)` |
| 长度 / 容量 | `a.length` | `len(a)` / `cap(a)` |
| 切片 | `a.slice(1,3)` 复制 | `a[1:3]` 共享 |
| map 判断存在 | `key in obj` / `map.has` | `v, ok := m[k]` |
| 遍历顺序 | 对象有插入顺序 | 无序 |
| 字符串长度 | `"中".length === 1` | `len("中") == 3` |
| 取第 i 个 | `s[i]` 返回字符串 | `s[i]` 返回 `byte` |
| 遍历字符 | `for (const ch of s)` | `for _, r := range s` |

`append` 必须接返回值，这是 TS 开发者最容易忘的一点。JS 字符串按 UTF-16 码元计数，Go 按 UTF-8 字节计数，这是两边最容易对不上的地方。

---

## 4.11 常见错误与陷阱

### 4.11.1 速查表

| 错误/现象 | 原因 | 修复 |
|-----------|------|------|
| `cannot use [4]int as [3]int` | 数组长度是类型的一部分 | 统一长度或改用切片 |
| `append` 后原切片没变化 | 没接返回值 | `s = append(s, x)` |
| `append` 改了别的切片 | 容量够，共享底层数组 | 三索引或先复制 |
| `make([]int, 5)` 得到 5 个 0 | 第二参数是长度 | 用 `make([]int, 0, 5)` |
| `s[2]` panic | 下标超过 `len` | 用 `append` 扩容 |
| `copy` 没复制全 | `dst` 太短 | `make` 足够长度 |
| `a == b` 编译失败 | 切片只能和 nil 比 | 循环比较或 `slices.Equal` |
| nil 切片 JSON 是 `null` | nil 与空切片语义不同 | 初始化 `make([]T, 0)` |
| `assignment to entry in nil map` | map 没 `make` | `m := make(map[K]V)` |
| 读 map 得到 0 分不清是否存在 | 缺 comma-ok | `v, ok := m[k]` |
| `cannot assign to struct field m[k].X` | map 元素不可寻址 | 取出、修改、放回 |
| `concurrent map writes` | map 并发写 | 加锁或 `sync.Map` |
| `len("中文")` 得到 6 | 返回字节数 | `utf8.RuneCountInString` |
| `s[0]` 打印出数字 | `byte` 是整数 | 用 `%c` 或 `string(s[0])` |
| `string(65)` 得到 `A` | 按码点转换 | `strconv.Itoa(65)` |
| 修改 `[]byte(s)` 后原串没变 | 转换复制了 | 改完再 `string(b)` 赋回 |
| `range` 字符串索引不连续 | 索引是字节偏移 | 用 `r` 或自己计数 |
| Builder 复制后 panic | Builder 不可复制 | 传 `*strings.Builder` |
| `+=` 循环很慢 | 反复分配复制 | `strings.Builder` / `strings.Join` |

### 4.11.2 切片共享的经典 bug

```go
data := []int{1, 2, 3, 4, 5}
a := data[:2]
b := append(a, 99)

fmt.Println(data) // [1 2 99 4 5]
fmt.Println(b)    // [1 2 99]
```
`a` 的容量是 5，`append` 没有扩容，直接覆盖 `data[2]`。修复：`b := append(data[:2:2], 99)`，限制 cap 强制扩容。

### 4.11.3 `range` 修改元素无效

`for _, v := range s` 中的 `v` 是副本，`v *= 2` 不会修改切片；要用 `for i := range s { s[i] *= 2 }`。第 03 章已经演示过这个行为。

### 4.11.4 `delete` 不会缩容

`delete` 只删除键值对，不会立即释放哈希表底层内存。大量增删后如果内存敏感，可以重建 map 并复制需要的键值。

### 4.11.5 按字节切片切坏 UTF-8

```go
s := "中"
fmt.Println(s[0:1]) // 非法 UTF-8，终端可能显示 �
```
要么使用完整字符边界，要么转 `[]rune` 操作。

---

## 4.12 与 Gin/GORM 的关联

### 4.12.1 Gin 参数是字符串或字符串切片

```go
id := c.Param("id")         // string
tags := c.QueryArray("tag") // []string，?tag=go&tag=web
fmt.Println(tags)

id, err := strconv.Atoi(c.Param("id"))
if err != nil {
    c.JSON(400, gin.H{"error": "invalid id"})
    return
}
```
`c.Query("page")` 和 `c.DefaultQuery("tag", "all")` 也返回 string，缺省值为 `""` 或指定值。`gin.H` 的定义就是 `map[string]any`。

### 4.12.2 JSON 数组对应切片

```go
type CreatePostRequest struct {
    Title string   `json:"title" binding:"required"`
    Tags  []string `json:"tags"`
}

func CreatePost(c *gin.Context) {
    var req CreatePostRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    c.JSON(201, gin.H{"title": req.Title, "tags": req.Tags})
}
```
请求 `{"title":"Go 入门","tags":["go","backend"]}` 会把 `tags` 解码成 `[]string`。请求没有 `tags` 时 `req.Tags` 是 nil，返回 JSON 是 `null`；需要空数组时初始化 `req.Tags = []string{}`。

### 4.12.3 查询结果就是切片

```go
var users []User
db.Where("age > ?", 18).Order("id DESC").Limit(10).Find(&users)
```
- `[]User` 是切片；
- `Find` 需要 `&users`，因为要修改切片头（第 06 章讲原因）；
- 查不到时 `users` 可能是 nil，返回 JSON 前可以初始化。

分页与 `IN` 查询：

```go
page, size := 1, 10
db.Offset((page - 1) * size).Limit(size).Find(&users)

ids := []int{1, 2, 3}
db.Where("id IN ?", ids).Find(&users) // 展开成 IN (?, ?, ?)
```
`Pluck` 可以只取一列到基础类型切片：`var names []string; db.Model(&User{}).Pluck("name", &names)`。

### 4.12.4 `[]byte` 对应 BLOB

```go
type File struct {
    ID   uint
    Name string
    Data []byte `gorm:"type:blob"`
}
```
`[]byte` 通常映射 `BLOB` / `BYTEA`；`string` 通常映射 `VARCHAR` / `TEXT`。小图标、哈希值适合；大文件不建议直接塞数据库。

### 4.12.5 动态 JSON 与切片字段

```go
var data map[string]any
if err := json.Unmarshal(body, &data); err != nil {
    // ...
}
name, _ := data["name"].(string)
age, _ := data["age"].(float64) // JSON 数字默认 float64
fmt.Println(name, age)

type User struct {
    ID   uint
    Tags []string `gorm:"serializer:json"`
}
```
`Tags` 会以 JSON 文本存进数据库，读出时自动反序列化。适合不需要按标签查询的场景；要按标签筛选应建关联表。

### 4.12.6 返回空数组而不是 null

```go
users := make([]User, 0)
if err := db.Find(&users).Error; err != nil {
    c.JSON(500, gin.H{"error": err.Error()})
    return
}
c.JSON(200, users)
```
这是前后端联调中最常见的细节之一。

---

## 4.13 练习题

### 基础题

1. 创建长度为 3、容量为 5 的 `[]int`，打印 `len/cap`，`append` 后再打印。
2. 让两个切片共享底层数组，修改一个，打印另一个验证。
3. 演示 `append` 触发扩容后不再共享，打印扩容前后的 `cap`。
4. 用 `copy` 把切片复制到更长的切片，打印复制个数和结果。
5. 用 map 统计一段英文文本中每个单词出现次数。
6. 用 `map[string]struct{}` 对字符串切片去重。
7. 遍历 `"Go语言"`，打印字节索引、rune 值和字符。
8. 用 `strings.Builder` 拼接 1 到 100 的数字，用逗号分隔。

### 进阶题

9. 写 `clone(s []int) []int`，返回完全独立的副本。
10. 写函数接收 `[]int`，返回去掉所有偶数后的新切片，不修改原切片。
11. 用三索引切片保护函数：内部可以 `append`，但不能覆盖调用方后面的数据。
12. 演示「`append` 覆盖原数组」的 bug，并用三索引修复。
13. 用 map 实现简单缓存：`Get(key)` 返回值和是否存在，`Set(key, value)` 写入。
14. 统计字符串中每个字符（按 rune）出现次数。
15. 实现 `reverse(s string) string`，正确处理中文。
16. 实现 `truncate(s string, n int) string`，按字符截断，不能截坏 UTF-8。
17. 用二维切片表示 3x3 矩阵，实现转置。

### 读文档题

18. 执行 `go doc builtin.make`，回答支持哪些类型、第二三参数含义。
19. 执行 `go doc strings.Builder`，列出至少 5 个方法并说明用途。
20. 执行 `go doc encoding/json.Marshal`，结合实验解释 nil 切片与空切片。

---

## 4.14 验收清单

- [ ] 能说清数组与切片的区别：长度、值语义、可比较性；
- [ ] 能解释切片头包含哪三部分，以及为什么传参是「拷贝切片头」；
- [ ] 能正确使用 `make([]T, len, cap)`，不把长度当容量；
- [ ] 能解释 `append` 何时共享、何时扩容；
- [ ] 能演示并修复切片共享底层数组的 bug；
- [ ] 会用三索引切片限制容量；
- [ ] 能说清 nil 切片与空切片的区别，尤其是 JSON 输出；
- [ ] 能熟练对 map 增删查改，使用 comma-ok；
- [ ] 能解释 map 遍历无序、key 必须可比较、元素不可寻址；
- [ ] 知道 map 并发写会 fatal，并知道修复方向；
- [ ] 能解释 `len("中文")` 与 `len([]rune("中文"))` 的差异；
- [ ] 能正确使用 `byte`、`rune`、`[]byte`、`[]rune`；
- [ ] 会用 `strings.Builder` 替代循环 `+=`；
- [ ] 能看懂 Gin 的 `QueryArray`、GORM 的 `[]User` 与 `map[string]any`。

---

## 4.15 本章小结

| 概念 | 一句话解释 | 易错点 |
|------|------------|--------|
| 数组 `[N]T` | 固定长度、值类型 | 长度是类型的一部分；传参整体拷贝 |
| 切片 `[]T` | 指向底层数组的视图 | 赋值/传参共享底层数组 |
| 切片头 | 指针 + len + cap | `append` 返回新头，必须接住 |
| `make` | 创建切片/map/chan | 第二参数是长度，第三参数才是容量 |
| `append` | 追加元素 | 容量够会覆盖共享数组；不够才扩容 |
| `copy` | 复制元素 | 只复制 `min(len)` 个；不扩容 |
| nil 切片 | 未初始化的切片 | JSON 是 `null`，不是 `[]` |
| 三索引 | `s[low:high:max]` | 限制 cap，强制 `append` 扩容 |
| map | 键值对哈希表 | nil 不能写；遍历无序；并发写 fatal |
| comma-ok | `v, ok := m[k]` | 区分「不存在」和「零值」 |
| map 元素 | 不可寻址 | 不能 `&m[k]`，不能直接改结构体字段 |
| string | 只读字节序列 | `len` 是字节数，不是字符数 |
| byte | `uint8` | `s[i]` 得到的是字节 |
| rune | `int32`，Unicode 码点 | `range` 按 rune 遍历 |
| UTF-8 | 变长编码 | 中文通常 3 字节，emoji 4 字节 |
| `[]byte` | 字节切片 | 转换会复制；适合二进制 |
| `[]rune` | 码点切片 | 转换会解码；适合按字符处理 |
| Builder | 高效字符串拼接 | 不可复制；可 `Grow` 预分配 |

---

## 4.16 下一章预告

第 05 章：**结构体、方法与嵌入**

你将学到结构体定义与初始化、字段导出、值接收者与指针接收者、方法集、嵌入与组合、结构体标签、反射够用版、构造函数与 Functional Options。这些内容是读懂 Gin 参数绑定和 GORM 模型定义的直接基础。
