# Go 语言系统教程（TypeScript 对照版）

> **课程定位**：面向有 JS/TS 基础、已经学过一遍 Go/Gin/GORM 的学习者，系统补齐 Go 语言，最终达到「无障碍阅读 Gin / GORM 官方文档与常见源码，并能独立编写日常业务代码」的水平。
>
> **本文件是课程总纲**：包含全部阶段、章节、知识点清单、练习与验收标准。详细讲义按学习进度逐章补齐。

---

## 课程信息

| 项目 | 内容 |
|------|------|
| 课程名称 | Go 语言系统教程（TS 对照版） |
| 目标读者 | 有 JS/TS 基础，学过一遍 Go/Gin/GORM，需要系统补全 |
| 最终目标 | 无障碍读懂 Gin v1.9+ / GORM v2 文档与常见源码；能写日常业务代码 |
| 语言基准 | Go 1.22 及以上（涉及版本差异会单独标注） |
| 框架基准 | Gin v1.9+、GORM v2 |
| 文档位置 | `go/docs/go-course/` |
| 建议节奏 | 约 4 周，每天 1–2 小时，可按实际情况调整 |
| 当前状态 | 总纲 + 第 01–19 章 + 附录 A–E 已全部完成 |

---

## 一、这门课解决什么问题

你可能已经遇到过这些情况：

- 看 Gin/GORM 文档时，方法签名里的 `*gin.Context`、`...HandlerFunc`、`*gorm.DB`、`interface{}` 看得似懂非懂；
- 知道 `db.Where(...).Find(&users)` 能跑，但说不清为什么能链式调用、为什么必须传指针；
- 遇到 `nil`、零值、切片共享、goroutine、context 时，只能凭感觉写；
- 学过语法，但知识是散的，遇到新 API 不知道怎么查、怎么读。

本课程按「语言基础 → 类型系统 → 并发与标准库 → 测试与实战 → 读框架文档」的顺序，把这些点串成一条完整的线。

---

## 二、学完你能做到什么

完成全部章节并通过验收后，你应当能够：

1. **读懂类型签名**：看到任意 Go 函数/方法签名，能说出参数、返回值、指针、接口、可变参数的含义；
2. **读懂 Gin 文档**：理解 `Engine`、`RouterGroup`、`Context`、`HandlerFunc`、中间件链的用法与关系；
3. **读懂 GORM 文档**：理解 `gorm.DB`、链式方法、`Statement`、模型标签、事务、钩子、`Valuer/Scanner`；
4. **独立写业务代码**：HTTP 接口、JSON 编解码、数据库访问、错误处理、并发控制、测试；
5. **自己查文档**：会用 `go doc`、pkg.go.dev、源码目录定位一个陌生 API 的用法；
6. **避开常见坑**：零值、nil 接口、切片共享、map 并发、goroutine 泄漏、循环变量捕获等。

---

## 三、课程边界（重要）

### 3.1 本课程覆盖

- Go 全部日常语法与类型系统（变量、常量、流程控制、函数、结构体、指针、接口、错误处理、泛型）；
- 复合类型与字符串原理（数组、切片、map、UTF-8、rune/byte）；
- 包与模块管理（go.mod、go.sum、版本、replace、vendor、go work）；
- Web 开发常用并发能力（goroutine、channel、select、context、sync）；
- 与 Gin/GORM 直接相关的标准库（`encoding/json`、`net/http`、`database/sql`、`io`、`time`、`log/slog` 等）；
- 测试、调试、工程化习惯；
- Gin/GORM 文档与源码的**阅读方法**（只读层面）。

### 3.2 本课程不覆盖（避免浪费时间）

| 不覆盖内容 | 原因 |
|------------|------|
| runtime、GC、GMP 调度器源码 | 属于底层原理，日常读框架文档用不到；讲过也难记住 |
| 汇编、编译器实现、链接器细节 | 与本课程目标无关 |
| cgo、unsafe 深水区 | 业务开发几乎不用，风险高 |
| 反射框架开发 | 只讲「够用版」：理解结构体标签、JSON、GORM 为什么能读标签 |
| 分布式、微服务、消息队列 | 同级已有 `rpc` 教程，本课程不重复 |
| 性能调优的完整方法论 | 只在测试章节讲 pprof/benchmark 的入门用法 |

> **说明**：不覆盖不等于永远不学。先把「读文档 + 写业务」这条主线打通，后续需要时再按需深入，效率更高。

---

## 四、学习方式

1. **按章节顺序学**：后面的章节依赖前面的知识，跳章容易卡住。
2. **每段代码必须手敲并运行**：看懂 ≠ 会写，运行结果和报错都是学习材料。
3. **先理解再记忆**：重点理解「类型、签名、值/指针、接口」四件事，API 名字可以查。
4. **每章做完验收清单**：全部打勾再进入下一章；不通过就回看对应知识点。
5. **练习题必须做**：基础题巩固语法，进阶题训练组合能力，读文档题训练真实工作能力。
6. **遇到不懂的，用固定模板提问**（见第十一节），描述越具体，解答越准确。

---

## 五、文档结构与章节状态

```
go/docs/go-course/
├── README.md                    # 本文件：课程总纲
├── 01-从TS到Go-心智模型与工具链.md
├── 02-变量常量与类型系统.md
├── 03-流程控制与函数.md
├── 04-复合类型-数组切片map字符串.md
├── 05-结构体方法与嵌入.md
├── 06-指针-够用版.md
├── 07-接口与多态.md
├── 08-错误处理.md
├── 09-包模块与依赖管理.md
├── 10-泛型-够用版.md
├── 11-并发编程-Web常用.md
├── 12-常用标准库-上.md
├── 13-常用标准库-下-JSON-HTTP-SQL.md
├── 14-测试与调试.md
├── 15-工程化习惯.md
├── 16-实战-最小WebAPI.md
├── 17-读懂Gin文档与源码.md
├── 18-读懂GORM文档与源码.md
├── 19-总复习与能力验收.md
├── 附录A-Go与TS对照速查表.md
├── 附录B-常见错误排查手册.md
├── 附录C-文档阅读索引.md
├── 附录D-练习与验收标准汇总.md
└── 附录E-术语表.md
```

### 章节状态

| 章节 | 状态 |
|------|------|
| README 总纲 | ✅ 已完成 |
| 第 01–19 章 | ✅ 已完成 |
| 附录 A–E | ✅ 已完成 |

---

## 六、阶段总览

| 阶段 | 章节 | 核心目标 | 建议时间 |
|------|------|----------|----------|
| 阶段一 | 第 01–04 章 | 建立 Go 心智模型，掌握基础语法与复合类型 | 5 天 |
| 阶段二 | 第 05–08 章 | 掌握结构体、指针、接口、错误处理（读框架的核心） | 6 天 |
| 阶段三 | 第 09–13 章 | 包与模块、泛型、并发、常用标准库 | 7 天 |
| 阶段四 | 第 14–16 章 | 测试、工程化、最小 Web API 实战 | 5 天 |
| 阶段五 | 第 17–19 章 | 读懂 Gin/GORM 文档与源码，总复习 | 4 天 |

> 时间仅为建议。以「通过每章验收」为准，不以天数为准。

---

## 七、各章详细大纲

> 每章详细讲义会按统一模板展开：本章目标 → 前置知识 → TS 对照 → 知识点逐条讲解（含可运行代码）→ 常见错误 → 与 Gin/GORM 的关联 → 练习题 → 验收清单。

---

### 阶段一：基础语法（与 TS 逐点对照）

---

#### 第 01 章：从 TS 到 Go —— 心智模型与工具链

**本章目标**

- 说清 Go 与 TypeScript 在编译、类型、运行方式上的本质差异；
- 独立安装并验证 Go 环境，创建、运行、构建一个模块；
- 熟练使用 `go run/build/test/fmt/vet/doc` 等常用命令；
- 能读懂最常见的编译错误。

**知识点清单**

1. Go 是什么：编译型、静态强类型、垃圾回收、原生并发、单一可执行文件；
2. Go 与 TS 对照：
   - TS 类型擦除 vs Go 编译期强类型；
   - `npm` / `package.json` vs `go mod` / `go.mod`；
   - `node` / `ts-node` vs `go run`；
   - `tsconfig.json` vs `go.mod` 的 `go` 指令；
   - TS 的 `any` vs Go 的 `any`（`interface{}`）；
3. 安装与验证：`go version`、`go env`；
4. 关键环境变量：`GOROOT`、`GOPATH`、`GOMODCACHE`、`GOPROXY`、`GOOS`/`GOARCH`、`CGO_ENABLED`（简介）；
5. 模块基础：
   - `go mod init`；
   - `go.mod` 的 `module` / `go` / `require`；
   - `go.sum` 的作用；
   - `go mod tidy`；
6. 常用命令：
   - `go run`、`go build`、`go install`；
   - `go test`；
   - `go fmt` / `gofmt`；
   - `go vet`；
   - `go doc`；
   - `go list`、`go clean`；
7. 包与目录规则：
   - 一个目录下的 `.go` 文件属于同一个包；
   - `package main` 与 `func main()`；
   - 导入路径与模块路径的关系；
8. 第一个程序逐行解析：`package`、`import`、`func main`、`fmt.Println`；
9. `gofmt` 与基础命名规范；
10. 读懂编译错误：文件名:行:列、`undefined`、`declared and not used`、`imported and not used`、`missing return`；
11. 项目目录雏形：`main.go` 与子包；
12. 查文档的两种方式：`go doc fmt.Println` 与 pkg.go.dev。

**与 Gin/GORM 的关联**

- Gin/GORM 都是普通 Go 模块，理解 `go.mod`、包导入、`main` 包是读它们文档的第一步；
- 文档示例中的 `package main`、`import "github.com/gin-gonic/gin"` 都建立在本章知识上。

**练习**

1. 安装 Go，运行 `go version` 与 `go env`，记录 `GOROOT`、`GOPATH`、`GOPROXY`；
2. 新建 `hello-go` 模块，输出「Hello, Go!」，用 `go run` 和 `go build` 各运行一次；
3. 用 `go doc` 查看 `fmt.Println`、`strings.Contains` 的签名并抄写；
4. 故意制造 3 个编译错误（未定义变量、变量未使用、导入未使用），记录报错原文并解释；
5. 创建 `main.go` + `mathutil/mathutil.go` 两个包，在 `main` 中调用子包函数。

**验收标准**

- [ ] 能独立创建模块并运行程序；
- [ ] 能解释 `go.mod` 和 `go.sum` 各自的作用；
- [ ] 能说出 `go run` 与 `go build` 的区别；
- [ ] 能读懂本章列出的 4 类常见编译错误；
- [ ] 能用 `go doc` 查任意标准库函数签名。

---

#### 第 02 章：变量、常量与类型系统

**本章目标**

- 掌握 Go 的变量声明方式与作用域规则；
- 理解零值、`nil` 与 TS `undefined/null` 的区别；
- 掌握基本类型、类型转换、类型别名与新类型；
- 能解释「为什么 Go 不允许隐式类型转换」。

**知识点清单**

1. 变量声明：
   - `var name type = value`；
   - `var name = value`（类型推断）；
   - `name := value`（短声明，只能函数内）；
   - 先声明后赋值；
   - 包级变量 vs 函数级变量；
2. 多重赋值、交换变量、空白标识符 `_`；
3. 常量：
   - `const`；
   - 无类型常量；
   - 常量表达式；
   - `iota` 与枚举写法；
4. 基本类型：
   - `bool`；
   - 有符号整数：`int8/16/32/64`、`int`；
   - 无符号整数：`uint8/16/32/64`、`uint`、`uintptr`；
   - 浮点：`float32/64`；
   - 复数：`complex64/128`（了解）；
   - `byte`（=`uint8`）、`rune`（=`int32`）；
   - `string`；
5. 零值：
   - 数值 `0`、字符串 `""`、布尔 `false`、指针/切片/map/chan/func/interface 为 `nil`；
   - 零值可用性；
   - 与 TS `undefined` / `null` 的对照；
6. `nil` 能用于哪些类型；`nil` 不能赋给 `int`、`string` 等；
7. 类型转换：
   - `T(v)`；
   - 无隐式转换；
   - 字符串与数字用 `strconv`；
   - 溢出与截断；
8. 类型别名 vs 新类型：
   - `type A = B`（别名）；
   - `type A B`（新类型，底层类型相同但不直接赋值）；
9. 作用域：
   - 包级、函数级、块级；
   - `if` / `for` / `switch` 中的短声明作用域；
   - 变量遮蔽（shadowing）与排查方法；
10. 命名规范：驼峰、常见缩写（`ID`、`URL`、`HTTP`）、包名小写；
11. 常量与枚举在业务中的常见用法（状态码、错误码）。

**与 Gin/GORM 的关联**

- GORM 的零值陷阱（`0`、`""`、`false` 被当作「未设置」）直接来自本章的零值概念；
- Gin 文档中的状态码常量、GORM 的字段默认值都涉及常量与类型。

**练习**

1. 用四种方式声明变量并打印；
2. 用 `iota` 定义一组订单状态常量；
3. 声明各类型变量并打印零值，观察输出；
4. 把 `int` 转 `int8` 制造溢出，解释结果；
5. 写一段代码演示变量遮蔽，并说明如何避免；
6. 用 `strconv.Atoi` 与 `strconv.Itoa` 完成字符串/整数互转。

**验收标准**

- [ ] 能说清 `var`、`:=`、`=` 的使用场景与限制；
- [ ] 能列出至少 5 种类型的零值；
- [ ] 能解释 `nil` 与 TS `null/undefined` 的区别；
- [ ] 能正确使用 `strconv` 做字符串与数字转换；
- [ ] 能识别并修复变量遮蔽问题。

---

#### 第 03 章：流程控制与函数

**本章目标**

- 掌握 Go 的 `if`、`for`、`switch`、`range` 写法；
- 掌握函数的各种形态：多返回值、可变参数、闭包；
- 理解 `defer` 的执行顺序与用途；
- 理解 `panic/recover` 的边界。

**知识点清单**

1. `if`：
   - 条件必须是 `bool`；
   - 不需要括号；
   - 支持初始化语句：`if err := f(); err != nil {}`；
2. `for` 的三种形态：
   - 经典三段式；
   - 条件式（替代 `while`）；
   - 无限循环；
   - Go 没有 `while` 关键字；
3. `range`：
   - 遍历 slice / array / map / string / channel；
   - 索引与值；
   - `_` 忽略不需要的值；
   - Go 1.22 起循环变量每次迭代新建（与旧版本差异）；
4. `switch`：
   - 表达式 switch；
   - 无表达式 switch（替代长 `if-else`）；
   - 不需要 `break`；
   - `fallthrough`；
   - `type switch`（第 07 章展开）；
5. `goto` 与标签（了解，不推荐使用）；
6. 函数定义：
   - 参数、返回值；
   - 相邻同类型参数简写；
   - 无返回值；
7. 多返回值与错误返回惯例；
8. 命名返回值与裸 `return`（慎用）；
9. 可变参数 `...T` 与切片展开 `f(s...)`；
10. 函数是一等公民：
    - 函数类型；
    - 函数作为参数/返回值；
    - 匿名函数；
    - 闭包与变量捕获；
11. `defer`：
    - 执行时机（函数返回前）；
    - 后进先出（LIFO）；
    - 参数在 `defer` 语句处求值；
    - 与命名返回值的关系；
    - 常见用途：关闭文件、解锁、`recover`；
12. `panic` / `recover`：
    - 触发方式；
    - `recover` 只在 `defer` 中有效；
    - 库代码不应随意 `panic`；
13. 值传递语义：
    - Go 所有参数都是值拷贝；
    - slice / map / chan 的「引用语义」来自其内部结构；
14. 递归、立即执行函数（IIFE）。

**与 Gin/GORM 的关联**

- Gin 中间件依赖 `defer`、闭包和函数作为值；
- Gin 的 `gin.Recovery()` 本质是 `defer + recover`；
- GORM 的钩子、回调大量使用函数类型。

**练习**

1. 用 `for` 输出 1–100 中的偶数，分别用三种写法；
2. 用无表达式 `switch` 实现成绩等级判断；
3. 写一个返回 `(结果, error)` 的除法函数；
4. 写一个闭包计数器；
5. 用 `defer` 演示 LIFO，并解释输出顺序；
6. 写一个函数，用 `defer + recover` 捕获 panic 并打印错误；
7. 解释：为什么 `defer` 中修改命名返回值会影响最终返回结果？

**验收标准**

- [ ] 能熟练写出三种 `for` 循环；
- [ ] 能解释 `defer` 的执行顺序与参数求值时机；
- [ ] 能正确使用多返回值与错误返回；
- [ ] 能写出闭包并解释捕获的是变量还是值；
- [ ] 能说明 `panic/recover` 的适用场景。

---

#### 第 04 章：复合类型 —— 数组、切片、map、字符串

**本章目标**

- 分清数组与切片的区别；
- 掌握切片的 `len/cap/append/copy` 与共享底层数组陷阱；
- 熟练使用 map，并理解其遍历无序与并发限制；
- 理解 `string`、`byte`、`rune` 与 UTF-8 的关系。

**知识点清单**

1. 数组：
   - 固定长度、值类型；
   - `[3]int` 与 `[4]int` 是不同类型；
   - 初始化、遍历、比较；
2. 切片：
   - 定义与字面量；
   - `make([]T, len, cap)`；
   - `append`；
   - `copy`；
   - `len` / `cap`；
   - 扩容策略（只讲结论，不展开源码）；
   - `nil` 切片 vs 空切片 `[]T{}`；
3. 切片表达式：
   - `s[low:high]`；
   - `s[low:high:max]`（三索引）；
   - 共享底层数组；
   - 修改互相影响；
   - `append` 可能触发扩容，导致不再共享；
4. 切片作为函数参数：传的是切片头（指针、长度、容量）的拷贝；
5. 二维切片与切片的切片；
6. map：
   - 声明与 `make`；
   - 字面量；
   - 增删查改；
   - `delete`；
   - comma-ok：`v, ok := m[k]`；
   - 遍历无序；
   - `len`；
7. map 的 key 必须可比较；value 可以是任意类型；
8. map 并发读写会 panic；`sync.Map` 简介（第 11 章展开）；
9. 用 map 实现 set；
10. 字符串：
    - 不可变字节序列；
    - UTF-8 编码；
    - `len(s)` 是字节数；
    - `rune` 是 Unicode 码点；
    - `for range` 按 rune 遍历；
    - 索引 `s[i]` 得到 `byte`；
11. `string` / `[]byte` / `[]rune` 的转换与开销；
12. `strings.Builder` 与字符串拼接性能；
13. 与 TS 对照：
    - 数组/切片 ↔ `Array`；
    - map ↔ `Record` / `Map`；
    - 字符串不可变性。

**与 Gin/GORM 的关联**

- `c.Param` / `c.Query` 返回 `string`；GORM 的 `[]byte` 字段映射 `BLOB`；
- JSON 数组对应 Go 切片，JSON 对象对应 `map[string]any` 或结构体；
- 分页查询返回的 `[]User` 就是切片。

**练习**

1. 创建长度为 3、容量为 5 的切片，`append` 后观察 `len/cap`；
2. 演示两个切片共享底层数组，修改一个影响另一个；
3. 演示 `append` 扩容后不再共享；
4. 用 map 统计一段文本中每个单词出现次数；
5. 用 map 实现字符串去重（set）；
6. 遍历一个包含中文的字符串，分别用字节索引和 `range`，解释差异；
7. 用 `strings.Builder` 拼接 10000 次字符串，与 `+=` 对比。

**验收标准**

- [ ] 能说清数组与切片的区别；
- [ ] 能解释切片共享底层数组与扩容行为；
- [ ] 能正确使用 comma-ok 判断 map key 是否存在；
- [ ] 能解释 `len("中文")` 与 `len([]rune("中文"))` 的差异；
- [ ] 能根据场景选择 `string`、`[]byte`、`[]rune`。

---

### 阶段二：类型系统与抽象（读 Gin/GORM 的核心）

---

#### 第 05 章：结构体、方法与嵌入

**本章目标**

- 掌握 struct 的定义、初始化与字段导出规则；
- 分清值接收者与指针接收者，理解方法集；
- 掌握嵌入与组合；
- 理解结构体标签，以及反射如何读取标签。

**知识点清单**

1. struct 定义与字段；
2. 字段导出规则：首字母大写才能跨包访问；
3. 初始化：
   - 按字段名；
   - 按位置（不推荐）；
   - 零值 + 后续赋值；
   - `new(T)` 与 `&T{}`；
4. 匿名结构体；
5. 结构体的可比较性；
6. 方法：
   - 接收者语法；
   - 值接收者 vs 指针接收者；
   - 自动取地址与自动解引用；
7. 方法集：
   - `T` 的方法集 vs `*T` 的方法集；
   - 对接口实现的影响；
8. 方法与函数的区别；
9. 嵌入：
   - 匿名字段；
   - 字段与方法提升；
   - 同名冲突与选择；
   - 组合优于继承；
10. 结构体标签：
    - 语法：反引号 + `key:"value"`；
    - `json`、`gorm`、`form`、`binding`、`validate` 示例；
11. 反射够用版：
    - `reflect.TypeOf` / `reflect.ValueOf`；
    - `Kind`；
    - `Field.Tag.Get("json")`；
    - 为什么 `encoding/json`、GORM 能读标签；
    - 不要求写反射框架；
12. 构造函数惯例 `NewXxx`；
13. Functional Options 选项模式（简介）；
14. 内存对齐（了解，不展开）。

**与 Gin/GORM 的关联**

- GORM 的模型就是 struct，字段必须导出，标签决定列名/主键/索引；
- Gin 的参数绑定（`ShouldBindJSON`）依赖 struct 标签；
- GORM 的 `gorm.Model` 通过嵌入提供 `ID`、`CreatedAt`、`UpdatedAt`、`DeletedAt`。

**练习**

1. 定义 `User` 结构体，用三种方式初始化；
2. 给 `User` 添加值接收者方法和指针接收者方法，观察修改是否生效；
3. 用嵌入实现 `Admin` 复用 `User` 的字段和方法；
4. 给字段添加 `json` 与 `gorm` 标签，用反射打印标签值；
5. 写一个 `NewUser(name string) *User` 构造函数；
6. 用 Functional Options 实现可配置的 `NewServer`。

**验收标准**

- [ ] 能解释值接收者与指针接收者的区别；
- [ ] 能说清方法集规则及其对接口的影响；
- [ ] 能用嵌入实现组合复用；
- [ ] 能写出常见 struct 标签；
- [ ] 能用反射读取字段标签。

---

#### 第 06 章：指针（够用版）

**本章目标**

- 掌握 `&`、`*`、指针类型与 nil 指针；
- 理解什么时候必须用指针；
- 能解释 GORM 为什么要求传指针；
- 不深挖逃逸分析细节。

**知识点清单**

1. 变量与内存地址；
2. `&` 取地址、`*` 解引用；
3. 指针类型 `*T`；
4. 指针的零值是 `nil`；nil 解引用会 panic；
5. 指针作为参数：修改调用方变量；
6. 指针作为返回值：返回局部变量地址是安全的（只给结论）；
7. `new(T)` vs `&T{}`；
8. 结构体指针与字段访问的自动解引用；
9. 什么时候用指针：
   - 需要修改原值；
   - 避免大对象拷贝；
   - 需要表达「可选 / 未设置」（nil）；
   - 方法集需要指针接收者；
10. 什么时候不用指针：小对象、不可变数据、追求简洁；
11. 指针与 GORM：
    - `db.Create(&user)` 为什么必须传指针；
    - `db.First(&user, id)` 如何回填结果；
    - 指针字段区分零值与 NULL；
12. 指针与 JSON：`omitempty`、nil 指针字段；
13. 多级指针（了解）；
14. `unsafe.Pointer`（不展开）；
15. 常见错误：nil 指针、误以为返回局部变量地址不安全、指针比较。

**与 Gin/GORM 的关联**

- GORM 的 `Create/First/Find/Update/Delete` 几乎都需要指针；
- Gin 的 `c.ShouldBindJSON(&req)` 需要传结构体指针才能写入；
- 指针字段是表达数据库 NULL 的常用方式。

**练习**

1. 写函数通过指针修改传入的整数；
2. 写函数通过指针修改结构体字段；
3. 演示 nil 指针解引用 panic，并用判断避免；
4. 对比 `new(User)` 与 `&User{}`；
5. 定义 `Age *int` 字段，分别赋 nil 和值，打印结果；
6. 解释 `db.Create(user)` 与 `db.Create(&user)` 的差别。

**验收标准**

- [ ] 能正确使用 `&` 和 `*`；
- [ ] 能判断什么时候必须用指针；
- [ ] 能解释 GORM 为什么要求传指针；
- [ ] 能识别并避免 nil 指针 panic；
- [ ] 能用指针字段表达「未设置」。

---

#### 第 07 章：接口与多态

**本章目标**

- 理解 Go 的隐式接口实现；
- 掌握接口值、类型断言、`type switch`；
- 认识 Gin/GORM 文档中高频出现的标准接口；
- 避开 nil 接口陷阱。

**知识点清单**

1. 接口定义：方法集合；
2. 隐式实现：不需要 `implements`；
3. 接口值：动态类型 + 动态值；
4. nil 接口 vs 接口内包含 nil 指针（经典陷阱）；
5. 空接口：
   - `interface{}` 与 `any`（Go 1.18+）；
   - 何时使用，何时避免；
6. 类型断言：
   - `v := x.(T)`（失败 panic）；
   - `v, ok := x.(T)`（安全）；
7. `type switch`；
8. 接口组合；
9. 小接口原则：接口越少方法越好（`io.Reader`/`io.Writer`）；
10. 常见标准接口：
    - `error`；
    - `fmt.Stringer`；
    - `io.Reader`、`io.Writer`、`io.Closer`、`io.ReadWriter`；
    - `http.Handler`；
    - `driver.Valuer`；
    - `sql.Scanner`；
11. 接口与指针接收者的方法集规则；
12. 接口的常见用途：依赖倒置、mock、多态；
13. 接口底层表示（iface/eface）只给概念；
14. 常见陷阱：
    - nil 接口判断；
    - 接口比较；
    - `[]interface{}` 与具体类型切片不通用；
15. Gin/GORM 中的接口：
    - `http.Handler`；
    - `gin.HandlerFunc`；
    - `gorm.Plugin`；
    - `driver.Valuer` / `sql.Scanner`。

**与 Gin/GORM 的关联**

- Gin 的 `Engine` 实现了 `http.Handler`；
- 中间件类型 `gin.HandlerFunc` 本质是函数类型；
- GORM 自定义类型通过实现 `driver.Valuer` 和 `sql.Scanner` 与数据库互转；
- 测试中常用接口做 mock。

**练习**

1. 定义 `Shape` 接口，让 `Circle`、`Rect` 隐式实现；
2. 写函数接收 `Shape`，输出面积；
3. 用类型断言从 `[]any` 中筛选出字符串；
4. 用 `type switch` 处理 `int`、`string`、`bool`；
5. 演示「接口内包 nil 指针」的陷阱；
6. 写一个类型实现 `fmt.Stringer`；
7. 写一个类型实现 `driver.Valuer` 与 `sql.Scanner`（伪代码即可）。

**验收标准**

- [ ] 能解释什么是隐式接口实现；
- [ ] 能正确使用类型断言与 `type switch`；
- [ ] 能解释 nil 接口陷阱；
- [ ] 能说出至少 5 个常用标准接口及其方法；
- [ ] 能理解 Gin/GORM 文档中的接口类型。

---

#### 第 08 章：错误处理

**本章目标**

- 掌握 Go 的错误返回惯例；
- 掌握错误包装、`errors.Is`、`errors.As`；
- 能设计可维护的错误信息；
- 理解 GORM 的错误处理方式。

**知识点清单**

1. `error` 接口：`Error() string`；
2. `errors.New` 与 `fmt.Errorf`；
3. 错误返回惯例：最后一个返回值、`if err != nil`；
4. 哨兵错误（sentinel error）：
   - `io.EOF`；
   - `sql.ErrNoRows`；
   - `gorm.ErrRecordNotFound`；
5. 错误包装：
   - `fmt.Errorf("...: %w", err)`；
   - `Unwrap`；
6. `errors.Is`：判断错误链中是否包含目标错误；
7. `errors.As`：把错误链中的某类错误提取到目标变量；
8. 自定义错误类型：实现 `error`；
9. 错误信息设计：
   - 带上下文；
   - 小写开头；
   - 不加句号；
   - 不重复记录；
10. 错误与日志的边界：底层包装、上层记录；
11. `panic/recover` 的使用边界；
12. `errors.Join`（Go 1.20+）合并多个错误；
13. GORM 错误处理：
    - `result.Error`；
    - `result.RowsAffected`；
    - `gorm.ErrRecordNotFound`；
14. HTTP 错误映射（预告，第 16 章展开）。

**与 Gin/GORM 的关联**

- GORM 每个操作返回 `*gorm.DB`，错误在 `Error` 字段；
- 查询不到记录返回 `gorm.ErrRecordNotFound`，需要用 `errors.Is` 判断；
- Gin 的错误处理中间件需要区分业务错误与系统错误。

**练习**

1. 写一个返回 `(int, error)` 的函数并处理错误；
2. 用 `%w` 包装错误，再用 `errors.Is` 判断原始错误；
3. 定义自定义错误类型，用 `errors.As` 提取；
4. 模拟 GORM 查询：判断 `errors.Is(err, gorm.ErrRecordNotFound)`；
5. 设计一组业务错误码与错误信息；
6. 用 `errors.Join` 合并两个错误。

**验收标准**

- [ ] 能写出符合规范的错误返回代码；
- [ ] 能区分 `errors.Is` 与 `errors.As` 的使用场景；
- [ ] 能设计带上下文的错误信息；
- [ ] 能正确处理 `gorm.ErrRecordNotFound`；
- [ ] 能说明什么时候该 panic、什么时候该返回 error。

---

### 阶段三：并发与工程基础

---

#### 第 09 章：包、模块与依赖管理

**本章目标**

- 掌握包、导入、导出、`init`、`internal`；
- 掌握 `go.mod`/`go.sum` 与版本管理；
- 会使用 `replace`、`vendor`、`go work`；
- 会阅读第三方库文档与源码。

**知识点清单**

1. `package` 声明与包名；
2. 一个目录一个包；
3. `import`：
   - 普通导入；
   - 别名导入；
   - 空白导入 `_`（副作用）；
   - 点导入（不推荐）；
4. 导出规则：首字母大写；
5. `init` 函数：执行时机与用途；
6. `internal` 包：限制外部导入；
7. `go.mod`：
   - `module`；
   - `go`；
   - `require`；
   - `replace`；
   - `exclude`、`retract`（了解）；
8. `go.sum`：校验依赖完整性；
9. 语义化版本：
   - `v0.x`、`v1.x`、`v2+`；
   - 主版本与导入路径；
   - 伪版本（pseudo-version）；
10. 依赖命令：
    - `go get`；
    - `go mod tidy`；
    - `go mod download`；
    - `go list -m all`；
11. `replace` 本地替换与调试；
12. `vendor` 目录；
13. `go work` 多模块工作区；
14. `GOPROXY` 与国内代理配置；
15. 阅读第三方库：
    - pkg.go.dev；
    - `go doc`；
    - 源码目录结构；
16. 依赖选择原则：维护状态、版本、许可证、社区；
17. 模块缓存 `GOMODCACHE`。

**与 Gin/GORM 的关联**

- `go get github.com/gin-gonic/gin`、`go get gorm.io/gorm` 的版本规则；
- GORM 的驱动按数据库拆分模块（如 `gorm.io/driver/mysql`）；
- 读 Gin/GORM 源码时需要理解包结构。

**练习**

1. 创建包含 `main` 包和两个子包的项目；
2. 用 `internal` 限制一个包只能内部导入；
3. 添加一个第三方依赖并查看 `go.mod` 变化；
4. 用 `replace` 把依赖替换为本地目录；
5. 用 `go list -m all` 查看依赖树；
6. 用 `go doc` 查看 `github.com/gin-gonic/gin` 的核心类型。

**验收标准**

- [ ] 能解释 `go.mod` 与 `go.sum` 的作用；
- [ ] 能读懂语义化版本与主版本规则；
- [ ] 会使用 `go mod tidy`、`replace`；
- [ ] 能用 `go doc` 查第三方库；
- [ ] 能说出 `internal` 的作用。

---

#### 第 10 章：泛型（够用版）

**本章目标**

- 理解类型参数与类型约束；
- 能读懂 Gin/GORM 中的泛型签名；
- 知道什么时候不该用泛型。

**知识点清单**

1. 为什么需要泛型；与 TS 泛型对照；
2. 类型参数：`func F[T any](x T) T`；
3. 类型约束：
   - `any`；
   - `comparable`；
   - 联合类型 `int | int64`；
   - `~T`（底层类型）；
   - 接口约束；
4. 泛型类型：`type Stack[T any] struct`；
5. 类型推断；
6. 泛型方法的限制（方法不能额外声明类型参数）；
7. 常用示例：`Map`、`Filter`、`Reduce`（教学用）；
8. 标准库泛型 API：
   - `slices` 包；
   - `maps` 包；
9. 如何阅读 Gin/GORM 中的泛型签名；
10. 何时不用泛型：过度抽象、只有一两种类型时。

**与 Gin/GORM 的关联**

- 标准库 `slices`/`maps` 的泛型 API 在业务中经常用到；
- GORM 的部分扩展/生成器使用泛型；
- 读文档时能认出 `[T any]` 的含义即可。

**练习**

1. 写泛型函数 `Max[T int | float64](a, b T) T`；
2. 写泛型函数 `Map[T, U any](s []T, f func(T) U) []U`；
3. 用 `slices.Sort` 排序 `[]int` 与 `[]string`；
4. 定义一个泛型 `Stack[T]`；
5. 解释 `~int` 与 `int` 作为约束的区别。

**验收标准**

- [ ] 能读懂 `[T any]`、`[T comparable]` 等签名；
- [ ] 能写简单的泛型函数与泛型类型；
- [ ] 会用 `slices`/`maps` 的常用泛型函数；
- [ ] 能判断什么时候不需要泛型。

---

#### 第 11 章：并发编程（Web 常用）

**本章目标**

- 掌握 goroutine 与 channel 的基本用法；
- 掌握 `select`、`context`、`sync`；
- 能识别常见并发陷阱；
- 不深挖 GMP 调度器。

**知识点清单**

1. 并发 vs 并行；
2. goroutine：
   - `go f()`；
   - 主 goroutine 退出会导致程序结束；
   - goroutine 没有返回值；
3. channel：
   - 无缓冲 / 有缓冲；
   - 发送与接收；
   - 阻塞行为；
   - 方向 channel；
   - `close`；
   - `range` channel；
   - `v, ok := <-ch`；
4. `select`：
   - 多路等待；
   - `default`；
   - 超时；
5. `context`：
   - `Background` / `TODO`；
   - `WithCancel`；
   - `WithTimeout`；
   - `WithDeadline`；
   - `WithValue`（慎用）；
   - 取消传播；
6. `sync`：
   - `Mutex`；
   - `RWMutex`；
   - `WaitGroup`；
   - `Once`；
7. `atomic` 基本用法；
8. 数据竞争：
   - `go test -race`；
   - 常见竞争场景；
9. goroutine 泄漏与避免；
10. 循环变量捕获（Go 1.22 前后的差异）；
11. 并发模式简介：worker pool、fan-in/fan-out；
12. Web 中的并发：
    - 每个请求一个 goroutine；
    - `r.Context()`；
    - Gin/GORM 的 context 传递；
13. 不展开：GMP 调度器、内存模型细节、调度器源码。

**与 Gin/GORM 的关联**

- 每个 HTTP 请求由独立 goroutine 处理；
- Gin 的 `c.Request.Context()` 用于超时与取消；
- GORM 的 `db.WithContext(ctx)` 把 context 传到数据库操作；
- 并发写 map、共享变量竞争是业务常见 bug。

**练习**

1. 启动两个 goroutine 打印数字，用 `WaitGroup` 等待结束；
2. 用无缓冲 channel 在两个 goroutine 间传值；
3. 用 `select` + `time.After` 实现超时；
4. 用 `context.WithTimeout` 取消一个耗时操作；
5. 用 `Mutex` 保护共享计数器，并用 `-race` 验证；
6. 演示 goroutine 泄漏并修复；
7. 解释 Go 1.22 循环变量变化。

**验收标准**

- [ ] 能正确创建、等待 goroutine；
- [ ] 能使用 channel 完成基本通信；
- [ ] 能用 `select` 处理超时；
- [ ] 能用 `context` 传递取消信号；
- [ ] 能识别并修复数据竞争与 goroutine 泄漏。

---

#### 第 12 章：常用标准库（上）

**本章目标**

- 熟练使用字符串、数字转换、排序、时间等高频库；
- 掌握格式化输出动词；
- 会用 `log/slog` 与 `embed`。

**知识点清单**

1. `fmt`：
   - `Print/Println/Printf`；
   - `Sprintf`；
   - `Errorf`；
   - 常用动词：`%v`、`%+v`、`%#v`、`%T`、`%d`、`%s`、`%q`、`%w`；
   - `Scan/Scanln`（了解）；
2. `strings`：
   - `Contains`、`HasPrefix`、`HasSuffix`；
   - `Split`、`Join`；
   - `Trim*`；
   - `Replace`、`ReplaceAll`；
   - `Fields`；
   - `Builder`；
3. `strconv`：
   - `Atoi`、`Itoa`；
   - `ParseInt`、`ParseFloat`；
   - `FormatInt`、`FormatFloat`；
4. `bytes`：
   - `Buffer`；
   - 与 `strings` 的对应关系；
5. `sort`：
   - `sort.Slice`；
   - `sort.Sort`；
   - 稳定性；
6. `slices`：
   - `Sort`、`SortFunc`；
   - `Contains`、`Index`；
   - `Insert`、`Delete`；
   - `Clone`、`Equal`；
7. `maps`：
   - `Keys`、`Values`；
   - `Clone`、`Equal`；
8. `time`：
   - `Time`、`Duration`；
   - `Now`；
   - `Parse`；
   - `Format`；
   - 参考时间 `2006-01-02 15:04:05`；
   - `Unix` / `UnixMilli`；
   - UTC 与 Local；
   - `Timer`、`Ticker`；
9. `log/slog`：
   - 结构化日志；
   - 级别；
   - Handler（简介）；
10. `embed`：
    - `//go:embed`；
    - 嵌入文件与目录；
11. 标准库阅读方法：`go doc` + 官方文档。

**与 Gin/GORM 的关联**

- Gin 日志、GORM 日志都涉及格式化与日志级别；
- `time.Time` 是 GORM 模型的高频字段类型；
- `strings`/`strconv` 是参数处理的基础。

**练习**

1. 用 `fmt.Printf` 打印结构体的 `%v`、`%+v`、`%#v`、`%T`；
2. 用 `strings` 拆分并统计句子中的单词；
3. 用 `strconv` 完成字符串与整数/浮点互转；
4. 用 `sort.Slice` 按年龄排序结构体切片；
5. 用 `time.Parse` 解析 `2026-10-17 12:00:00` 并格式化输出；
6. 用 `slog` 输出一条带字段的日志；
7. 用 `embed` 嵌入一个文本文件并读取。

**验收标准**

- [ ] 能熟练使用 `fmt` 常用动词；
- [ ] 能用 `strings`/`strconv` 处理常见字符串与数字；
- [ ] 能用 `sort`/`slices` 排序；
- [ ] 能正确解析与格式化时间；
- [ ] 能用 `slog` 输出结构化日志。

---

#### 第 13 章：常用标准库（下）—— JSON、IO、HTTP、SQL

**本章目标**

- 掌握 JSON 编解码与 struct 标签；
- 掌握 `io`/`bufio`/`os`/`filepath` 基本用法；
- 理解 `net/http` 的 Handler 模型与中间件本质；
- 理解 `database/sql` 的基本用法与连接池；
- 为读 Gin/GORM 文档打基础。

**知识点清单**

1. `encoding/json`：
   - `Marshal` / `Unmarshal`；
   - struct tag；
   - `omitempty`；
   - 字段必须导出；
   - 数字默认解码为 `float64`（`interface{}` 场景）；
   - 时间序列化；
   - `json.RawMessage`；
   - `Decoder` / `Encoder`；
   - HTML 转义与 `SetEscapeHTML(false)`；
2. `io`：
   - `Reader`、`Writer`、`Closer`、`Seeker`；
   - `io.Copy`；
   - `io.ReadAll`；
   - `io.EOF`；
3. `bufio`：
   - `Reader`；
   - `Writer`；
   - `Scanner`；
4. `os`：
   - 文件读写；
   - 环境变量；
   - `os.Args`；
   - `os.Exit`；
5. `filepath`：
   - `Join`、`Base`、`Dir`、`Ext`；
   - `Glob`；
   - 跨平台路径；
6. `net/http`：
   - `Handler` 接口；
   - `HandlerFunc`；
   - `ServeMux`；
   - `Server`；
   - `Request` / `ResponseWriter`；
   - 中间件本质：包装 Handler；
   - 超时配置；
   - 优雅关闭 `Shutdown`；
7. `database/sql`：
   - `Open`、`Ping`；
   - `Query`、`QueryRow`、`Exec`；
   - `Scan`；
   - 预编译语句；
   - 连接池：`SetMaxOpenConns`、`SetMaxIdleConns`、`SetConnMaxLifetime`；
   - `context` 版本方法；
8. `sql.NullString` 等；
9. `driver.Valuer` 与 `sql.Scanner`；
10. 关系理解：
    - Gin 是对 `net/http` 的增强；
    - GORM 是对 `database/sql` 的封装；
    - 读框架文档时，能对应回原始标准库能力。

**与 Gin/GORM 的关联**

- Gin 的 `Context` 封装了 `http.Request` 与 `http.ResponseWriter`；
- Gin 的 `c.JSON` 底层是 `encoding/json`；
- GORM 底层使用 `database/sql`；
- 连接池配置直接影响 GORM 性能；
- `context` 贯穿 HTTP → GORM。

**练习**

1. 定义 struct 并用 `json.Marshal/Unmarshal` 往返；
2. 用 `omitempty` 观察空值字段是否输出；
3. 用 `json.Decoder` 解析请求体；
4. 用 `os` + `filepath` 读取一个文件；
5. 用 `net/http` 写一个返回 JSON 的 Handler；
6. 写一个简单的中间件（打印耗时）包装 Handler；
7. 用 `database/sql` 连接 SQLite/MySQL，完成一次查询与插入；
8. 解释 Gin/GORM 分别封装了哪些标准库能力。

**验收标准**

- [ ] 能正确使用 JSON 编解码与标签；
- [ ] 能用 `io`/`os`/`filepath` 完成文件操作；
- [ ] 能写 `http.Handler` 与简单中间件；
- [ ] 能用 `database/sql` 完成 CRUD 与 `Scan`；
- [ ] 能说清 Gin/GORM 与标准库的关系。

---

### 阶段四：质量与实战

---

#### 第 14 章：测试与调试

**本章目标**

- 会写单元测试、表驱动测试、子测试；
- 会写基准测试与 HTTP 测试；
- 会用 `go vet`、`-race`、Delve 排查问题。

**知识点清单**

1. `go test`；
2. 测试文件 `_test.go`；
3. `TestXxx(t *testing.T)`；
4. 表驱动测试；
5. 子测试 `t.Run`；
6. `t.Helper`、`t.Cleanup`、`t.TempDir`；
7. 覆盖率：
   - `-cover`；
   - `-coverprofile`；
8. 基准测试：
   - `BenchmarkXxx(b *testing.B)`；
   - `-bench`；
   - `-benchmem`；
9. 示例测试 `ExampleXxx`；
10. `httptest`：
    - `NewRecorder`；
    - `NewServer`；
11. mock：
    - 接口 + 手写 fake；
    - `testify/mock` 简介；
12. `go vet`；
13. `staticcheck`（第三方）；
14. Delve：
    - `dlv debug`；
    - `dlv test`；
    - 断点；
15. `go test -race`；
16. 测试组织与命名；
17. Gin/GORM 项目测试思路：handler 测试、repository 测试、集成测试。

**与 Gin/GORM 的关联**

- Gin 提供 `httptest` 测试路由与中间件；
- GORM 测试通常用 SQLite 内存库或事务回滚；
- 读框架源码时，测试文件是很好的用法示例。

**练习**

1. 给字符串工具函数写表驱动测试；
2. 用 `t.Run` 拆分子测试；
3. 写一个 Benchmark 并运行 `-benchmem`；
4. 用 `httptest` 测试一个 HTTP Handler；
5. 用接口 + fake 测试一个依赖数据库的 service；
6. 用 `-race` 检测并发问题；
7. 用 Delve 给一个程序打断点并查看变量。

**验收标准**

- [ ] 能写表驱动测试与子测试；
- [ ] 能写基准测试并读懂内存分配；
- [ ] 能用 `httptest` 测试接口；
- [ ] 能用接口 mock 依赖；
- [ ] 会用 `-race`、`go vet`、Delve。

---

#### 第 15 章：工程化习惯

**本章目标**

- 建立可维护的 Go 项目结构；
- 掌握命名、注释、错误处理、配置、日志、优雅关闭等规范；
- 形成代码评审清单。

**知识点清单**

1. 目录结构：
   - `cmd/`；
   - `internal/`；
   - `pkg/`（谨慎使用）；
   - `api/`、`config/`、`migrations/`；
   - 不过度设计；
2. 命名规范：
   - 包名小写、简短；
   - 变量名；
   - 接口命名（`-er`）；
   - 错误变量 `ErrXxx`；
3. 注释规范：
   - 包注释；
   - 导出标识符注释；
   - `TODO`/`FIXME`；
4. 错误处理规范：
   - 包装；
   - 边界；
   - 日志；
5. 配置：
   - 环境变量；
   - `flag`；
   - 配置文件；
   - 不硬编码；
6. 日志：
   - 结构化；
   - 级别；
   - TraceID；
7. 优雅关闭：
   - `signal.NotifyContext`；
   - `Server.Shutdown`；
8. 依赖注入的朴素做法：
   - 构造函数注入；
   - `wire` 简介（不展开）；
9. 代码评审清单；
10. `gofmt` / `goimports`；
11. `golangci-lint` 简介；
12. 构建时写入版本信息 `-ldflags`；
13. 交叉编译 `GOOS`/`GOARCH`。

**与 Gin/GORM 的关联**

- Gin 项目的中间件、路由、配置组织；
- GORM 的数据库初始化与连接池配置；
- 优雅关闭是 Web 服务上线的必备能力。

**练习**

1. 把第 16 章实战项目按 `cmd/internal` 结构拆分；
2. 用环境变量配置端口与数据库地址；
3. 用 `slog` 输出带 TraceID 的日志；
4. 用 `signal.NotifyContext` + `Shutdown` 实现优雅关闭；
5. 写一份自己的代码评审清单；
6. 用 `-ldflags` 注入版本号并打印。

**验收标准**

- [ ] 能设计不过度复杂的目录结构；
- [ ] 能规范命名与注释；
- [ ] 能用环境变量/flag 配置程序；
- [ ] 能实现优雅关闭；
- [ ] 能说出依赖注入的基本做法。

---

#### 第 16 章：实战 —— 最小 Web API

**本章目标**

- 用 `net/http` + `database/sql` 独立完成一个小型 API；
- 理解 handler/service/repository 分层；
- 完成测试与优雅关闭；
- 为读 Gin/GORM 文档建立对照。

**知识点清单（实践项）**

1. 需求与接口设计：RESTful 基础；
2. 项目初始化与目录结构；
3. `net/http` 路由与 Handler；
4. 中间件：日志、恢复、简单 CORS；
5. JSON 请求/响应；
6. 统一错误响应格式；
7. `database/sql` 接入（SQLite 或 MySQL）；
8. 分层：handler / service / repository；
9. `context` 传递；
10. 优雅关闭；
11. 测试：`httptest` + 表驱动；
12. 对照：同样功能用 Gin/GORM 怎么写（只读对照，不要求实现）。

**与 Gin/GORM 的关联**

- 完成后你会发现：Gin 主要替代路由/上下文/绑定/渲染；GORM 主要替代 `database/sql` 的样板代码；
- 带着这套对照去读第 17、18 章，理解会快很多。

**练习**

1. 实现用户 CRUD API（至少 Create/Get/List）；
2. 加入统一错误响应；
3. 为 handler 写测试；
4. 加入优雅关闭；
5. 写出 Gin/GORM 版本的等价代码思路（伪代码即可）。

**验收标准**

- [ ] 项目可运行、接口可用；
- [ ] 代码分层清晰；
- [ ] 关键路径有测试；
- [ ] 能解释每一层为什么存在；
- [ ] 能说出 Gin/GORM 会简化哪些部分。

---

### 阶段五：读文档与源码

---

#### 第 17 章：如何读懂 Gin 文档与源码

**本章目标**

- 建立 Gin 的类型地图；
- 能读懂文档中的常见签名；
- 能顺着源码定位关键实现；
- 能判断方法是否会中止请求、是否返回链式对象。

**知识点清单**

1. Gin 文档入口：README、docs、pkg.go.dev；
2. 核心类型：
   - `Engine`；
   - `RouterGroup`；
   - `Context`；
   - `HandlerFunc`；
   - `HandlersChain`；
   - `IRoutes` / `IRouter`；
3. 路由注册：
   - `GET/POST/...`；
   - `Group`；
   - `Use`；
4. 请求处理：
   - `c.Param`；
   - `c.Query` / `c.DefaultQuery`；
   - `c.ShouldBindJSON` / `ShouldBindQuery` / `ShouldBindUri`；
   - `c.GetHeader`；
5. 响应：
   - `c.JSON`；
   - `c.String`；
   - `c.Data`；
   - `c.HTML`；
   - `c.Status`；
6. 中间件：
   - `c.Next()`；
   - `c.Abort()`；
   - `c.Set` / `c.Get`；
   - 执行顺序；
7. 与 `net/http` 的关系：
   - `Engine` 实现 `http.Handler`；
   - `HandlerFunc` 是 `func(*Context)`；
   - `ResponseWriter` 被包装；
8. 源码阅读路径：
   - `gin.go`；
   - `routergroup.go`；
   - `context.go`；
   - `middleware/`；
9. 文档符号：
   - `*gin.Context`；
   - `gin.H`；
   - `...HandlerFunc`；
   - `any`；
10. 常见判断：
    - 方法是否返回 `*Context`；
    - 是否会自动写响应；
    - 是否会 abort 后续 handler。

**与 Gin/GORM 的关联**

- 这一章是「读文档能力」的第一次完整训练；
- 后续读 GORM 文档会复用同样的方法：类型地图 → 签名 → 链式关系 → 源码路径。

**练习**

1. 用自己的话解释 `gin.HandlerFunc`、`gin.HandlersChain`；
2. 读文档回答：`c.ShouldBindJSON` 返回什么？失败时会不会自动写 400？
3. 读源码定位 `c.Next()` 与 `c.Abort()` 的实现；
4. 解释 `Engine` 为什么能作为 `http.Handler` 使用；
5. 画出一次请求经过中间件链的执行顺序图。

**验收标准**

- [ ] 能画出 Gin 核心类型关系图；
- [ ] 能解释常见方法签名；
- [ ] 能定位至少 5 个核心类型/方法的源码位置；
- [ ] 能判断中间件与响应的行为；
- [ ] 能不依赖教程读懂一段 Gin 文档示例。

---

#### 第 18 章：如何读懂 GORM 文档与源码

**本章目标**

- 建立 GORM 的类型地图；
- 理解链式调用、`Statement`、`Session`；
- 能读懂模型标签、事务、钩子、`Valuer/Scanner`；
- 能定位核心源码。

**知识点清单**

1. GORM 文档入口与 v2 版本；
2. 核心类型：
   - `gorm.DB`；
   - `gorm.Statement`；
   - `gorm.Session`；
   - `gorm.Config`；
3. 链式方法：
   - `db.Where(...).Order(...).Limit(...).Find(&users)`；
   - 为什么返回 `*gorm.DB`；
   - `Error` 字段；
   - 链式方法的「克隆/复用」注意事项；
4. 模型与标签：
   - `gorm.Model`；
   - `primaryKey`；
   - `column`；
   - `type`；
   - `default`；
   - `uniqueIndex` / `index`；
   - `embedded`；
5. CRUD 方法签名：
   - `Create`；
   - `First` / `Take` / `Last`；
   - `Find`；
   - `Updates` / `Update`；
   - `Delete`；
   - 指针与零值；
6. 关联：
   - `BelongsTo`；
   - `HasOne`；
   - `HasMany`；
   - `Many2Many`；
   - `Preload`；
7. 事务：
   - `db.Transaction(func(tx *gorm.DB) error)`；
8. 钩子：
   - `BeforeCreate` / `AfterCreate`；
   - `BeforeUpdate` / `AfterUpdate`；
   - `BeforeDelete` / `AfterDelete`；
   - `AfterFind`；
9. 回调与 `Clause`（只读层面）；
10. 自定义类型：
    - `driver.Valuer`；
    - `sql.Scanner`；
11. `WithContext` 与 `context`；
12. `ErrRecordNotFound`；
13. 源码阅读路径：
    - `gorm.go`；
    - `chainable_api.go`；
    - `callbacks/`；
    - `statement.go`；
14. 文档中的常见「坑」：零值更新、软删除、预加载、连接池。

**与 Gin/GORM 的关联**

- 这一章直接对应你读 GORM 文档的需求；
- 与第 13 章的 `database/sql` 对照，能看清 GORM 封装了什么。

**练习**

1. 解释 `db.Where(...).Find(&users)` 的返回值与链式原理；
2. 读文档回答：`Updates` 与 `Update` 的区别？零值如何更新？
3. 解释 `gorm.Model` 包含哪些字段；
4. 用 `Transaction` 写一个转账示例（伪代码）；
5. 读源码定位 `Create` 的回调执行路径；
6. 解释 `Valuer/Scanner` 在自定义类型中的作用；
7. 列出 5 个 GORM 常见坑并说明原因。

**验收标准**

- [ ] 能画出 GORM 核心类型关系图；
- [ ] 能解释链式调用与 `Error` 字段；
- [ ] 能读懂模型标签与 CRUD 签名；
- [ ] 能解释事务、钩子、`Valuer/Scanner`；
- [ ] 能定位至少 5 个核心类型/方法的源码位置。

---

#### 第 19 章：总复习与能力验收

**本章目标**

- 串起全部知识形成地图；
- 通过综合测试检验「读文档 + 写代码」能力；
- 明确下一步学习路线。

**知识点清单**

1. 知识地图：语法 → 类型系统 → 并发 → 标准库 → 工程 → 框架；
2. 高频问题 30 题（语法、指针、接口、错误、并发、标准库）；
3. 读文档实战：
   - 随机挑 Gin 文档条目解释；
   - 随机挑 GORM 文档条目解释；
4. 代码阅读测试：给一段源码，说出行为；
5. 能力自评表；
6. 常见误区总结；
7. 下一步路线：
   - Gin/GORM 教程；
   - 项目实战；
   - 测试与性能；
   - 分布式与微服务（同级 `rpc` 教程）。

**与 Gin/GORM 的关联**

- 最终验收标准就是：打开 Gin/GORM 文档任意一页，能独立读懂并写出示例。

**练习**

1. 完成 30 题综合测试；
2. 从 Gin 文档任选 3 个 API，写出用法与注意事项；
3. 从 GORM 文档任选 3 个 API，写出用法与注意事项；
4. 阅读一段 Gin/GORM 源码并解释；
5. 制定自己的下一步学习计划。

**验收标准**

- [ ] 综合测试正确率达到 80% 以上；
- [ ] 能独立读懂 Gin 文档示例；
- [ ] 能独立读懂 GORM 文档示例；
- [ ] 能解释常见源码片段；
- [ ] 能写出下一步计划。

---

## 八、附录

| 附录 | 内容 | 编写时机 |
|------|------|----------|
| 附录 A | Go 与 TypeScript 语法/概念对照速查表 | 随章节同步 |
| 附录 B | 常见编译错误、运行错误、并发错误排查手册 | 随章节同步 |
| 附录 C | 标准库、Gin、GORM 文档与源码阅读索引 | 阶段五前完成 |
| 附录 D | 全部练习题与验收标准汇总 | 随章节同步 |
| 附录 E | 术语中英对照表 | 随章节同步 |

---

## 九、推荐学习计划（4 周）

| 周次 | 内容 | 目标 |
|------|------|------|
| 第 1 周 | 第 01–04 章 | 能写基础程序，理解类型、切片、map、字符串 |
| 第 2 周 | 第 05–08 章 | 能读懂并写出结构体、指针、接口、错误处理 |
| 第 3 周 | 第 09–13 章 | 掌握模块、并发、标准库，能读写 JSON/HTTP/SQL |
| 第 4 周 | 第 14–19 章 | 完成实战、测试、工程化，能读 Gin/GORM 文档 |

> 如果某章验收未通过，不要赶进度。宁可多花一天，也不要带着漏洞进入下一章。

---

## 十、每章统一模板与验收规则

每章讲义固定包含以下部分：

1. **本章目标**：学完能做什么；
2. **前置知识**：需要先掌握什么；
3. **TS 对照与心智模型**：用 TypeScript 类比；
4. **知识点逐条讲解**：每个知识点配可运行代码与输出；
5. **常见错误与陷阱**：错误写法 + 正确写法 + 原因；
6. **与 Gin/GORM 的关联**：这部分知识在框架文档中出现在哪；
7. **练习题**：基础题 / 进阶题 / 读文档题；
8. **验收清单**：全部打勾才进入下一章。

**验收规则**

- 验收清单由你自己逐项确认；
- 练习题至少完成 80%；
- 能用自己的话解释本章核心概念；
- 遇到问题先回看对应知识点，再提问。

---

## 十一、提问模板

遇到不懂的地方，按下面格式提问，能得到更准确的解答：

```
【章节】第 XX 章
【知识点】例如：切片共享底层数组
【我的理解】我认为……
【实际现象】代码/报错/输出……
【我的问题】为什么……？/ 这两者有什么区别？
```

---

## 十二、版本说明与更新记录

### 版本基准

- Go：1.22 及以上（示例尽量兼容 1.22；涉及 1.21/1.22 差异会标注）；
- Gin：v1.9 及以上；
- GORM：v2；
- 环境：Windows 11 + PowerShell（命令在 macOS/Linux 下基本通用）。

### 更新记录

| 日期 | 内容 |
|------|------|
| 2026-10-17 | 创建课程总纲，确定 19 章 + 5 附录的结构 |
| 2026-10-17 | 完成第 01 章：从 TS 到 Go —— 心智模型与工具链 |
| 2026-10-18 | 完成第 02 章：变量、常量与类型系统 |
| 2026-10-18 | 完成第 03 章：流程控制与函数 |
| 2026-10-18 | 完成第 04–19 章及附录 A–E，课程文档全部落地 |

---

> **最后提醒**：这门课的目标不是背 API，而是建立「看懂类型和签名 → 查文档 → 写对代码」的能力。每章都要动手，每章都要验收。祝你学有所成。
