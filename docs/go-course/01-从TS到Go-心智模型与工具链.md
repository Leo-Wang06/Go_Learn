# 第 01 章：从 TS 到 Go —— 心智模型与工具链

> **阶段**：阶段一 · 基础语法
> **建议时间**：1 天
> **前置知识**：会用命令行；有 JS/TS 基础
> **代码基准**：Go 1.22 及以上，命令以 Windows PowerShell 为主

---

## 本章目标

学完本章，你应当能够：

1. 说清 Go 与 TypeScript 在编译、类型、运行方式上的本质差异；
2. 独立安装并验证 Go 环境；
3. 创建、运行、构建一个 Go 模块；
4. 熟练使用 `go run`、`go build`、`go test`、`go fmt`、`go vet`、`go doc`；
5. 看懂 `go.mod`、`go.sum` 和包的目录规则；
6. 读懂最常见的编译错误，并知道怎么修；
7. 理解 Gin/GORM 的示例为什么也是普通的 `package main` + `import`。

---

## 1.1 先建立心智模型：Go 和 TS 到底哪里不一样

### 1.1.1 一张总表

| 维度 | TypeScript | Go |
|------|------------|-----|
| 执行方式 | 编译成 JS，由 Node/浏览器执行 | 编译成机器码，直接运行 |
| 类型检查 | 编译期检查，类型被擦除 | 编译期强类型检查，类型参与编译 |
| 类型转换 | 很多场景可隐式兼容 | 基本没有隐式转换，必须显式转换 |
| 空值 | `undefined` / `null` | 每种类型都有「零值」；`nil` 只用于特定类型 |
| 类 | `class` | `struct` + 方法 |
| 继承 | `extends` | 嵌入（embedding）+ 组合 |
| 异常 | `try/catch/finally` | `error` 返回值；`panic/recover` 是最后手段 |
| 并发 | `async/await`、`Promise` | `goroutine`、`channel` |
| 可见性 | `export` 关键字 | 首字母大小写 |
| 包管理 | `npm` / `package.json` | `go mod` / `go.mod` |
| 入口 | 任意模块文件 | `package main` + `func main()` |
| 分号 | 可选 | 不写；编译器在行尾自动插入 |
| 运行时依赖 | 需要 Node | 通常单个可执行文件，无需额外运行时 |

这张表不需要背，先有个印象。下面逐条解释。

---

### 1.1.2 TypeScript 的类型会被擦除，Go 的类型参与编译

TypeScript：

```ts
function add(a: number, b: number): number {
  return a + b
}
```

编译成 JavaScript 后，类型信息被擦掉，运行时看到的是：

```js
function add(a, b) {
  return a + b
}
```

Go：

```go
func add(a int, b int) int {
    return a + b
}
```

Go 在编译阶段就检查类型。下面这段代码**编译不过**：

```go
var a int = 1
var b int64 = 2

a = b // 编译错误：不能把 int64 赋给 int
```

必须显式转换：

```go
a = int(b) // 正确
```

> **为什么重要**：Go 连 `int` 和 `int64` 都分得很清楚。数据库、JSON、时间戳经常涉及不同整数类型，类型搞错就编译不过。这是好事——错误在编译期就暴露，而不是运行时才炸。

---

### 1.1.3 Go 没有 class，但有 struct + 方法

TypeScript：

```ts
class User {
  constructor(public name: string, public age: number) {}

  sayHello() {
    return `我是 ${this.name}`
  }
}
```

Go：

```go
type User struct {
    Name string
    Age  int
}

func (u User) SayHello() string {
    return "我是 " + u.Name
}
```

对应关系：

| TypeScript | Go |
|------------|-----|
| `class` | `struct` |
| 字段 | 结构体字段 |
| 方法 | 带接收者的函数 |
| `this` | 接收者 `u` |
| `constructor` | 惯例函数 `NewUser(...)` |
| `extends` | 嵌入（embedding） |

本章只需要看懂「Go 用 struct 和方法替代 class」这一点，细节在第 05 章。

---

### 1.1.4 TypeScript 有 undefined/null，Go 有零值和 nil

TypeScript 里，变量没赋值是 `undefined`：

```ts
let name: string
console.log(name) // undefined
```

Go 里，变量声明后一定有一个「零值」：

```go
var name string
fmt.Println(name) // 输出空行，name 是 ""

var age int
fmt.Println(age) // 0

var ok bool
fmt.Println(ok) // false
```

常见零值：

| 类型 | 零值 |
|------|------|
| `int`、`int64`、`float64` 等数值 | `0` |
| `string` | `""`（空字符串） |
| `bool` | `false` |
| 指针、切片、map、channel、函数、接口 | `nil` |

`nil` 不是「什么都能赋值」的万能空值。它只能用于指针、切片、map、channel、函数、接口这些类型。下面这样写是错的：

```go
var n int = nil // 编译错误
```

> **为什么重要**：GORM 最有名的「零值陷阱」就来自这里。`Age int` 的零值是 `0`，GORM 无法区分「年龄是 0」和「没有设置年龄」。这个坑会在第 05、06、18 章反复出现。

---

### 1.1.5 TypeScript 用 try/catch，Go 用 error 返回值

TypeScript：

```ts
try {
  const data = JSON.parse(text)
  console.log(data)
} catch (err) {
  console.error('解析失败', err)
}
```

Go 的常见写法：

```go
data, err := parse(text)
if err != nil {
    fmt.Println("解析失败：", err)
    return
}
fmt.Println(data)
```

Go 把错误当作普通返回值，调用方必须显式检查。本章只需要记住这个模式，第 08 章会系统讲。

---

### 1.1.6 TypeScript 用 async/await，Go 用 goroutine

TypeScript：

```ts
async function main() {
  const result = await fetchData()
  console.log(result)
}
```

Go：

```go
go doSomething() // 启动一个 goroutine，立即继续往下执行
```

Go 没有 `await`，多个 goroutine 之间用 channel、`context`、`sync.WaitGroup` 协调。第 11 章会系统讲。

---

### 1.1.7 可见性靠首字母大小写

TypeScript 用 `export`：

```ts
export class User {}
export function Add() {}
```

Go 没有 `export` 关键字，靠首字母：

```go
type User struct {   // 大写 U：其他包可以访问
    Name string      // 大写 N：其他包可以访问
    age  int         // 小写 a：只有本包可以访问
}

func Add() {}        // 大写 A：导出
func subtract() {}   // 小写 s：不导出
```

这条规则适用于：包级变量、常量、类型、函数、方法、结构体字段。

> **为什么重要**：GORM 只能映射**大写开头**的结构体字段。小写字段它直接忽略。Gin 的参数绑定也一样，JSON 字段要能写入结构体，字段必须导出。

---

### 1.1.8 包管理是 go mod，不是 npm

| TypeScript / Node | Go |
|-------------------|-----|
| `package.json` | `go.mod` |
| `package-lock.json` | `go.sum` |
| `node_modules/` | 全局模块缓存（`GOMODCACHE`） |
| `npm install` | `go get` / `go mod tidy` |
| `npm run build` | `go build` |
| `npx ts-node index.ts` | `go run .` |

Go 项目目录里通常**没有** `node_modules` 这样的依赖目录。下载的依赖放在全局模块缓存里，`go.mod` 记录版本，`go.sum` 记录校验值。

---

## 1.2 安装 Go 并验证

### 1.2.1 Windows 安装步骤

1. 打开官网下载页：<https://go.dev/dl/>
2. 下载 Windows 安装包（通常是 `go1.xx.x.windows-amd64.msi`）；
3. 双击安装，默认安装到 `C:\Program Files\Go`；
4. 安装程序会自动把 `C:\Program Files\Go\bin` 加入 PATH；
5. **重新打开一个 PowerShell 窗口**（旧窗口不会自动刷新 PATH）；
6. 执行：

```powershell
go version
```

看到类似输出就说明安装成功：

```
go version go1.22.x windows/amd64
```

版本号和你实际安装的一致即可，不必追求最新。

> **路径建议**：Go 本身支持中文和空格路径，但少数工具（尤其是 cgo、代码生成器）对特殊字符支持不好。如果练习时遇到奇怪的路径问题，可以把练习项目放到 `E:\gopractice` 这类纯英文、无空格的目录。

---

### 1.2.2 `go version` 和 `go env`

```powershell
go version   # 查看版本
go env       # 查看所有 Go 环境变量
```

`go env` 输出很长，重点看这几个：

| 变量 | 含义 | 示例 |
|------|------|------|
| `GOROOT` | Go 安装目录 | `C:\Program Files\Go` |
| `GOPATH` | 工作目录（放缓存、安装的二进制） | `C:\Users\lenovo\go` |
| `GOMODCACHE` | 下载的第三方模块缓存 | `C:\Users\lenovo\go\pkg\mod` |
| `GOBIN` | `go install` 安装二进制的目录 | 默认 `GOPATH\bin` |
| `GOPROXY` | 模块代理地址 | `https://proxy.golang.org,direct` |
| `GOOS` | 目标操作系统 | `windows` |
| `GOARCH` | 目标 CPU 架构 | `amd64` |
| `CGO_ENABLED` | 是否启用 cgo | `1`（有 C 编译器时） |

也可以单独查：

```powershell
go env GOROOT
go env GOPATH
go env GOMODCACHE
go env GOPROXY
```

> **Go Modules 时代**：项目可以放在任意目录，**不需要**放进 `GOPATH/src`。`GOPATH` 现在主要用来放模块缓存和 `go install` 的二进制。

---

### 1.2.3 配置 GOPROXY（国内网络）

如果下载依赖很慢或超时，可以把代理设置为国内镜像：

```powershell
go env -w GOPROXY=https://goproxy.cn,direct
```

解释：

- `https://goproxy.cn`：代理服务器；
- `,direct`：代理找不到时，回退到直接下载；
- `go env -w`：把配置写入 Go 的用户配置文件，永久生效。

查看配置文件位置：

```powershell
go env GOENV
```

撤销配置：

```powershell
go env -u GOPROXY
```

> 如果你所在网络访问官方代理正常，可以不改。改与不改都不影响语法学习。

---

### 1.2.4 常见安装问题

| 现象 | 原因 | 解决 |
|------|------|------|
| `go : 无法将“go”项识别为 cmdlet...` | PATH 没生效 | 重新打开 PowerShell；检查 `C:\Program Files\Go\bin` 是否在 PATH |
| `go version` 显示旧版本 | 装了多个 Go，PATH 顺序不对 | 检查 PATH，删掉旧版本路径 |
| `go get` 卡住或超时 | 网络访问代理慢 | 配置 `GOPROXY` |
| `cgo: C compiler not found` | 用了需要 cgo 的包，但没有 C 编译器 | 初学者先用纯 Go 包；或安装 MinGW-w64 |
| 项目路径含中文/空格导致奇怪错误 | 少数工具不支持 | 换到纯英文无空格路径 |

---

## 1.3 Go 工具链：每个命令是干什么的

### 1.3.1 命令总表

| 命令 | 作用 | 常用示例 |
|------|------|----------|
| `go version` | 查看版本 | `go version` |
| `go env` | 查看/设置环境变量 | `go env -w GOPROXY=...` |
| `go mod init` | 初始化模块，生成 `go.mod` | `go mod init hello-go` |
| `go mod tidy` | 添加缺失依赖、删除无用依赖 | `go mod tidy` |
| `go get` | 添加/升级依赖 | `go get github.com/gin-gonic/gin@latest` |
| `go mod download` | 下载依赖到缓存 | `go mod download` |
| `go run` | 编译并运行（不留下二进制） | `go run .` |
| `go build` | 编译（main 包会生成可执行文件） | `go build -o app.exe .` |
| `go install` | 编译并安装到 `GOBIN` | `go install .` |
| `go test` | 运行测试 | `go test ./...` |
| `go fmt` | 格式化代码 | `go fmt ./...` |
| `go vet` | 静态检查可疑代码 | `go vet ./...` |
| `go doc` | 查看文档 | `go doc fmt.Println` |
| `go list` | 列出包/模块 | `go list ./...`、`go list -m all` |
| `go clean` | 清理构建产物/缓存 | `go clean -cache`、`go clean -modcache` |

> `./...` 是 Go 的通配写法，意思是「当前目录及所有子目录中的包」。

---

### 1.3.2 `go run`、`go build`、`go install` 的区别

这三个命令最容易混。

**`go run .`**

- 编译当前目录的 main 包；
- 把临时可执行文件放到临时目录并运行；
- 运行结束后临时文件删除；
- 适合开发时快速验证。

```powershell
go run .
```

**`go build .`**

- 编译当前目录的包；
- 如果是 main 包，会在当前目录生成可执行文件；
- 不会自动运行。

```powershell
go build .
```

Windows 下生成 `hello-go.exe`（名字取决于目录名或模块名）。也可以指定输出名：

```powershell
go build -o app.exe .
```

然后手动运行：

```powershell
.\app.exe
```

**`go install .`**

- 编译并安装到 `GOBIN`（默认 `GOPATH\bin`）；
- 安装后的命令可以在任意目录运行（前提是 `GOBIN` 在 PATH 里）；
- 常用于安装命令行工具。

```powershell
go install .
```

简单记：

| 命令 | 生成文件 | 自动运行 | 用途 |
|------|----------|----------|------|
| `go run .` | 临时文件，运行后删除 | 是 | 开发调试 |
| `go build .` | 当前目录的可执行文件 | 否 | 构建产物 |
| `go install .` | `GOBIN` 下的可执行文件 | 否 | 安装工具 |

---

### 1.3.3 `go fmt`、`go vet`、`go doc`

**`go fmt`**

Go 官方格式化工具，统一缩进、空格、换行。团队协作必须用。

```powershell
go fmt ./...      # 格式化当前模块所有包
gofmt -w .        # 格式化当前目录的 .go 文件
```

区别：

- `go fmt ./...` 按包处理，递归所有子目录；
- `gofmt -w .` 只处理当前目录的 `.go` 文件，不递归。

大多数编辑器保存时会自动运行 `gofmt`。

**`go vet`**

静态检查，能发现一些编译器不报、但很可能是 bug 的写法，例如：

- `Printf` 格式串和参数类型不匹配；
- 结构体标签格式错误；
- 不可达代码；
- 锁被复制。

```powershell
go vet ./...
```

**`go doc`**

在终端查看标准库或第三方库文档：

```powershell
go doc fmt            # fmt 包概览
go doc fmt.Println    # 某个函数
go doc -all strings   # strings 包全部导出内容
```

第 1.8 节会详细练习。

---

### 1.3.4 交叉编译（了解）

Go 可以在一台机器上编译另一个平台的可执行文件，靠 `GOOS` 和 `GOARCH`。

PowerShell 示例：编译 Linux 版本

```powershell
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"
go build -o app-linux .
```

编译完成后恢复环境：

```powershell
Remove-Item Env:GOOS
Remove-Item Env:GOARCH
Remove-Item Env:CGO_ENABLED
```

> `CGO_ENABLED=0` 表示不使用 cgo，通常能得到静态链接的二进制，部署更方便。初学阶段不用深入，知道有这个开关即可。

---

## 1.4 模块（module）和包（package）

### 1.4.1 两个概念不要混

| 概念 | 定义 | 类比 |
|------|------|------|
| 模块（module） | 一组包的集合，由根目录的 `go.mod` 定义 | 一个项目 / 一个 npm 包 |
| 包（package） | 一个目录下的 `.go` 文件集合，属于同一个命名空间 | 一个目录 / 一个命名空间 |

一个模块可以包含多个包。一个包对应一个目录。

```
hello-go/                 <- 模块根目录（有 go.mod）
├── go.mod
├── main.go               <- package main
└── mathutil/             <- 子目录
    └── mathutil.go       <- package mathutil
```

---

### 1.4.2 初始化模块：`go mod init`

```powershell
mkdir hello-go
cd hello-go
go mod init hello-go
```

执行后生成 `go.mod`：

```
module hello-go

go 1.22
```

说明：

- `module hello-go`：模块路径。本地练习随便起名即可；
- `go 1.22`：模块使用的 Go 语言版本，也是最低版本要求（你本机安装的版本可能不同，以实际生成内容为准）。

如果以后要把项目发布到 GitHub，模块路径应该写成仓库地址，例如：

```powershell
go mod init github.com/yourname/hello-go
```

这样别人才能通过 `go get github.com/yourname/hello-go` 下载。

> **模块路径不是网址**。它只是一个标识符，Go 用它来定位包。只有需要被远程下载时，才要求它看起来像域名路径。

---

### 1.4.3 `go.mod` 逐行解释

一个引入依赖后的 `go.mod` 可能长这样：

```
module hello-go

go 1.22

require github.com/gin-gonic/gin v1.9.1

require (
    github.com/bytedance/sonic v1.9.1 // indirect
    github.com/gabriel-vasile/mimetype v1.4.2 // indirect
)
```

| 部分 | 含义 |
|------|------|
| `module` | 模块路径 |
| `go` | 语言版本/最低版本 |
| `require` | 直接依赖及版本 |
| `// indirect` | 间接依赖（依赖的依赖） |
| `replace` | 把某个依赖替换成另一个路径/版本（第 09 章） |
| `toolchain` | Go 1.21+ 可能出现的工具链版本声明（了解） |

> 如果 `go.mod` 里声明的 Go 版本比本机新，Go 1.21+ 可能会自动下载对应工具链。想强制使用本机版本，可以设置 `GOTOOLCHAIN=local`（了解即可）。

---

### 1.4.4 `go.sum` 是什么

`go.sum` 记录依赖模块内容的校验值，用来防止依赖被篡改。

特点：

- 由 Go 自动生成和维护；
- **不要手改**；
- 应该和 `go.mod` 一起提交到版本库；
- 只使用标准库的项目可能没有 `go.sum`，因为没有第三方依赖需要校验。

类比：`go.sum` ≈ `package-lock.json` 里的完整性校验部分。

---

### 1.4.5 包与目录规则

规则一：**一个目录下的 `.go` 文件必须属于同一个包**。

```
mathutil/
├── add.go        package mathutil
└── sub.go        package mathutil   ✅
```

```
mathutil/
├── add.go        package mathutil
└── sub.go        package other      ❌ 编译错误
```

规则二：**可执行程序的入口必须是 `package main` + `func main()`**。

```go
package main

func main() {
    // 程序从这里开始执行
}
```

规则三：**包名和目录名通常一致，但不是强制的**。

目录叫 `mathutil`，包名一般也叫 `mathutil`。也可以不同，但强烈建议保持一致，否则读代码会困惑。

规则四：**测试文件可以用不同的包名**。

`xxx_test.go` 可以是 `package foo`，也可以是 `package foo_test`。第 14 章会讲。

---

### 1.4.6 import 路径怎么算

导入路径 = **模块路径 + 包所在子目录**。

假设模块路径是 `hello-go`：

| 包所在目录 | import 路径 |
|------------|-------------|
| 模块根目录 | `hello-go`（通常不这样导入 main 包） |
| `mathutil/` | `hello-go/mathutil` |
| `internal/db/` | `hello-go/internal/db` |

标准库没有模块前缀：

```go
import "fmt"
import "strings"
import "net/http"
```

第三方库用它自己的模块路径：

```go
import "github.com/gin-gonic/gin"
import "gorm.io/gorm"
```

导入多个包时用括号：

```go
import (
    "fmt"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"

    "hello-go/mathutil"
)
```

习惯上分成三组（标准库、第三方、本项目），组间空一行。`goimports` 会自动整理。

---

## 1.5 第一个程序：逐行拆解

### 1.5.1 创建模块

打开 PowerShell：

```powershell
mkdir hello-go
cd hello-go
go mod init hello-go
```

确认当前目录有 `go.mod`：

```powershell
Get-ChildItem
```

---

### 1.5.2 写 `main.go`

新建 `main.go`，内容如下：

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, Go!")
}
```

---

### 1.5.3 运行

```powershell
go run .
```

输出：

```
Hello, Go!
```

再构建：

```powershell
go build .
```

当前目录会出现 `hello-go.exe`，运行：

```powershell
.\hello-go.exe
```

输出同样是：

```
Hello, Go!
```

---

### 1.5.4 逐行解释

```go
package main
```

声明当前文件属于 `main` 包。可执行程序必须有 `main` 包。

```go
import "fmt"
```

导入标准库 `fmt`（format），用于格式化输入输出。

```go
func main() {
```

定义 `main` 函数。它没有参数，也没有返回值，是程序的入口。

```go
    fmt.Println("Hello, Go!")
```

调用 `fmt` 包的 `Println` 函数，打印一行文本并换行。

```go
}
```

结束函数。

> **关键认知**：Go 程序从 `main` 包的 `main` 函数开始执行。一个可执行程序有且只有一个入口。

---

### 1.5.5 为什么没有分号

Go 的语法允许省略分号，编译器会在行尾自动插入。所以下面两种写法等价：

```go
fmt.Println("Hello")
```

```go
fmt.Println("Hello");
```

实际开发中**不写分号**，`gofmt` 也不会加分号。

---

### 1.5.6 大括号不能换行（重要）

Go 在行尾自动补分号，所以下面这样写会报错：

```go
func main()
{
    fmt.Println("Hello")
}
```

编译错误类似：

```
.\main.go:5:1: syntax error: unexpected newline, expecting {
```

原因：编译器在 `func main()` 这一行末尾补了一个分号，变成 `func main();`，于是它认为函数声明已经结束，无法理解下一行的 `{`。

正确写法：**左大括号必须和声明在同一行**。

```go
func main() {
    fmt.Println("Hello")
}
```

这条规则同样适用于 `if`、`for`、`switch`、`struct`、`func` 等所有需要大括号的地方。

---

## 1.6 多包项目实战

这一节创建一个包含两个包的项目，顺便验证「大小写可见性」和 `go run` 的细节。

### 1.6.1 目录结构

在 `hello-go` 模块中创建以下结构：

```
hello-go/
├── go.mod
├── main.go
├── helper.go
└── mathutil/
    └── mathutil.go
```

---

### 1.6.2 编写 `mathutil/mathutil.go`

```go
package mathutil

// Add 返回 a 与 b 的和。
func Add(a, b int) int {
    return a + b
}

// subtract 是包内私有函数，包外不可见。
func subtract(a, b int) int {
    return a - b
}
```

说明：

- 包名是 `mathutil`；
- `Add` 大写开头，可以被其他包调用；
- `subtract` 小写开头，只能在 `mathutil` 包内部使用。

---

### 1.6.3 编写 `main.go` 和 `helper.go`

`helper.go`：

```go
package main

func greet() string {
    return "Hello from helper.go"
}
```

`main.go`：

```go
package main

import (
    "fmt"

    "hello-go/mathutil"
)

func main() {
    sum := mathutil.Add(3, 5)
    fmt.Println("3 + 5 =", sum)
    fmt.Println(greet())
}
```

注意：

- `main.go` 和 `helper.go` 都属于 `package main`；
- 导入路径是 `hello-go/mathutil`，即「模块路径 + 子目录」；
- `greet` 定义在 `helper.go`，但 `main.go` 可以直接调用，因为它们在同一个包。

---

### 1.6.4 运行

在模块根目录执行：

```powershell
go run .
```

输出：

```
3 + 5 = 8
Hello from helper.go
```

也可以构建：

```powershell
go build -o app.exe .
.\app.exe
```

---

### 1.6.5 大小写可见性实验

把 `main.go` 改成：

```go
fmt.Println(mathutil.subtract(3, 5))
```

再运行：

```powershell
go run .
```

会得到类似错误：

```
.\main.go:11:14: cannot refer to unexported name mathutil.subtract
```

这就是「小写开头只能在包内访问」。

如果改成 `mathutil.Add(3, 5)`，则正常。

> **规则总结**：跨包访问时，标识符首字母必须大写。结构体字段也一样。

---

### 1.6.6 `go run .` 和 `go run main.go` 的区别

现在 `package main` 有两个文件：`main.go` 和 `helper.go`。

执行：

```powershell
go run .
```

会编译当前目录下**所有** `.go` 文件，正常运行。

执行：

```powershell
go run main.go
```

只会编译 `main.go` 这一个文件，运行时会报错：

```
.\main.go:13:14: undefined: greet
```

因为 `greet` 定义在 `helper.go`，而 `helper.go` 没有被包含进来。

> **结论**：在模块目录中，优先使用 `go run .`；`go run 文件名` 只适合单文件小练习。

---

## 1.7 读懂常见编译错误

### 1.7.1 错误信息怎么读

Go 的错误格式通常是：

```
文件路径:行号:列号: 错误信息
```

例如：

```
.\main.go:7:2: undefined: name
```

含义：

- `.\main.go`：出错文件；
- `7`：第 7 行；
- `2`：第 2 列；
- `undefined: name`：错误原因。

排查顺序：

1. 先看第一个错误，不要被后面一堆错误吓到；
2. 打开对应文件和行号；
3. 修一个，重新运行；
4. 重复，直到通过；
5. 通过编译后，再运行 `go vet ./...`。

---

### 1.7.2 十类常见错误

> 下面展示的是典型信息，Go 版本不同，措辞可能略有差异，但含义一致。

#### 错误 1：`undefined: xxx`

```
.\main.go:7:2: undefined: name
```

常见原因：

- 变量/函数名拼错；
- 没有导入对应的包；
- 跨包访问了小写开头的标识符；
- 使用了尚未声明的变量。

修复：检查拼写、导入、大小写。

---

#### 错误 2：`declared and not used: x`

```
.\main.go:8:2: declared and not used: x
```

原因：局部变量声明了但没使用。Go 强制要求使用局部变量。

修复：使用它，或者删除声明。

> 注意：**包级变量**未使用不会报错，只有函数内的局部变量会。

---

#### 错误 3：`"os" imported and not used`

```
.\main.go:4:2: "os" imported and not used
```

原因：导入了包但没有使用。

修复：删除导入；如果只是想要副作用，可以用空白导入：

```go
import _ "some/package"
```

初学阶段不要滥用空白导入。

---

#### 错误 4：`missing return`

```
.\main.go:12:1: missing return
```

原因：函数声明了返回值，但存在没有 `return` 的执行路径。

错误示例：

```go
func sign(n int) int {
    if n > 0 {
        return 1
    }
    // n <= 0 时没有 return
}
```

修复：补全所有路径的返回值：

```go
func sign(n int) int {
    if n > 0 {
        return 1
    }
    return -1
}
```

---

#### 错误 5：类型不匹配

```
.\main.go:9:10: cannot use "abc" (untyped string constant) as int value in assignment
```

原因：Go 没有隐式类型转换，`string` 不能直接赋给 `int`。

修复：使用正确的类型，或显式转换（字符串转数字用 `strconv`，第 12 章）。

---

#### 错误 6：缺少依赖

```
no required module provides package github.com/foo/bar; to add it:
    go get github.com/foo/bar
```

原因：导入了第三方包，但 `go.mod` 里没有这个依赖。

修复：

```powershell
go get github.com/foo/bar
go mod tidy
```

---

#### 错误 7：包路径不对

```
package foo is not in std (C:\Program Files\Go\src\foo)
```

常见原因：

- 把标准库包名拼错；
- 想导入第三方包，但写成了标准库形式；
- 模块没有正确初始化。

修复：检查导入路径；第三方包要用完整模块路径。

---

#### 错误 8：没有 `go.mod`

```
go: go.mod file not found in current directory or any parent directory;
see 'go help modules'
```

原因：在模块外执行了 `go run .` / `go build .` 等包级命令。

修复：

```powershell
go mod init 你的模块名
```

---

#### 错误 9：同一目录包名不一致

```
found packages main and mathutil in C:\...\hello-go
```

原因：同一个目录下的 `.go` 文件声明了不同的 `package`。

修复：统一包名。

---

#### 错误 10：大括号换行

```
.\main.go:5:1: syntax error: unexpected newline, expecting {
```

原因：`{` 写到了下一行，触发了自动分号。

修复：把 `{` 放回声明同一行。

---

### 1.7.3 更多错误速查

| 错误信息（示例） | 原因 | 修复 |
|------------------|------|------|
| `cannot refer to unexported name pkg.foo` | 跨包访问小写标识符 | 改成大写导出，或提供导出函数 |
| `command-line-arguments is not a main package` | `go run` 了非 main 包 | 运行 main 包 |
| `no Go files in ...` | 目录里没有 `.go` 文件 | 检查目录和文件名 |
| `multiple-value f() ... in single-value context` | 多返回值当成一个值使用 | 接收全部返回值（第 03 章） |
| `cannot use x (type A) as type B` | 类型不匹配 | 显式转换或改类型 |
| `import cycle not allowed` | 包之间循环导入 | 重新设计包依赖（第 09 章） |
| `undefined: fmt.Printlnn` | 函数名拼错 | 检查拼写，用 `go doc` 确认 |
| `expected declaration, found ...` | 包级位置写了语句 | 语句必须放进函数里 |

---

## 1.8 如何查文档

### 1.8.1 用 `go doc` 查标准库

```powershell
go doc fmt
```

输出 `fmt` 包的概览和导出函数列表。

查具体函数：

```powershell
go doc fmt.Println
```

输出类似：

```
package fmt // import "fmt"

func Println(a ...any) (n int, err error)
```

查另一个：

```powershell
go doc strings.Contains
```

输出：

```
package strings // import "strings"

func Contains(s, substr string) bool
```

查看某个包的全部导出内容：

```powershell
go doc -all strings
```

> 提示：`go doc` 输出的是「声明 + 注释」，非常适合快速确认参数和返回值。

---

### 1.8.2 pkg.go.dev

浏览器打开 <https://pkg.go.dev/>，搜索包名，例如：

- `fmt`
- `net/http`
- `github.com/gin-gonic/gin`
- `gorm.io/gorm`

页面会展示：

- 包说明；
- 所有导出类型、函数、方法；
- 示例代码；
- 版本信息。

读第三方文档时，这是最常用的入口。

---

### 1.8.3 读懂函数签名

以 `strings.Contains` 为例：

```go
func Contains(s, substr string) bool
```

拆解：

| 部分 | 含义 |
|------|------|
| `func` | 这是一个函数 |
| `Contains` | 函数名，大写开头，可导出 |
| `(s, substr string)` | 两个参数，类型都是 `string` |
| `bool` | 返回一个布尔值 |

再看一个：

```go
func Println(a ...any) (n int, err error)
```

拆解：

- `...any`：可变参数，可以传 0 个或多个任意类型的值；
- `(n int, err error)`：返回两个值；
- `err error`：返回一个 `error` 接口（第 08 章展开）。

再看一个带指针和接口的：

```go
func NewReader(rd io.Reader) *bufio.Reader
```

- `io.Reader`：接口类型（第 07 章）；
- `*bufio.Reader`：指向 `bufio.Reader` 的指针（第 06 章）。

> 现在不需要完全理解 `io.Reader` 和 `*`。只要知道：**读文档时，先看参数类型和返回值个数**，再逐步深入。

---

### 1.8.4 练习：查三个函数

执行下面三条命令，并把签名抄到笔记里：

```powershell
go doc fmt.Println
go doc strings.Contains
go doc strconv.Atoi
```

参考答案：

```
func Println(a ...any) (n int, err error)
func Contains(s, substr string) bool
func Atoi(s string) (int, error)
```

---

## 1.9 代码风格与命名

### 1.9.1 永远使用 `gofmt`

Go 的格式非常统一：

- 缩进用 Tab；
- 左大括号不换行；
- 运算符两侧有空格；
- 导入分组排序。

不需要手动调格式，执行：

```powershell
go fmt ./...
```

或者在编辑器里配置保存时自动格式化。

> 统一的格式让所有人的代码看起来一样，评审时只需要关注逻辑，不关注空格。

---

### 1.9.2 命名规范

| 对象 | 规范 | 示例 |
|------|------|------|
| 包名 | 小写、简短、单词 | `mathutil`、`http` |
| 导出标识符 | 首字母大写 | `User`、`Add`、`Parse` |
| 非导出标识符 | 首字母小写 | `user`、`add`、`parse` |
| 变量/函数 | 驼峰 | `userName`、`parseJSON` |
| 缩写 | 全大写或全小写，保持一致 | `userID`、`URL`、`HTTPClient` |
| 错误变量 | `Err` 前缀 | `ErrNotFound`、`ErrTimeout` |
| 接口 | 常以 `-er` 结尾 | `Reader`、`Writer`、`Stringer` |

Go 社区常见写法是 `userID`，而不是 `userId`。因为 `ID` 是缩写，按惯例全大写。

---

### 1.9.3 注释规范

导出标识符应该有注释，并且注释以标识符名开头：

```go
// Add 返回 a 与 b 的和。
func Add(a, b int) int {
    return a + b
}
```

这样 `go doc` 才能正确展示。

包注释放在包声明上方：

```go
// Package mathutil 提供基础数学运算。
package mathutil
```

> 注释写「为什么」和「做什么」，不要写「怎么做的」流水账。

---

## 1.10 与 Gin/GORM 的关联

### 1.10.1 框架也是普通模块

Gin 和 GORM 不是魔法，它们就是普通的 Go 模块：

- Gin 的模块路径：`github.com/gin-gonic/gin`
- GORM 的模块路径：`gorm.io/gorm`
- GORM 的 MySQL 驱动：`gorm.io/driver/mysql`

你在文档里看到的 import，就是模块路径：

```go
import "github.com/gin-gonic/gin"
import "gorm.io/gorm"
import "gorm.io/driver/mysql"
```

所以本章学的「模块路径 + 导入规则」直接适用于它们。

---

### 1.10.2 安装 Gin/GORM 时发生了什么

在模块目录中执行：

```powershell
go get github.com/gin-gonic/gin@latest
go get gorm.io/gorm@latest
go get gorm.io/driver/mysql@latest
```

会发生：

1. 下载模块到 `GOMODCACHE`；
2. 在 `go.mod` 的 `require` 中写入版本；
3. 在 `go.sum` 中写入校验值；
4. 可能会带出很多 `// indirect` 间接依赖。

清理无用依赖：

```powershell
go mod tidy
```

> 从 Go 1.17 起，`go get` 只负责修改依赖，不再安装可执行程序。安装命令行工具用：
>
> ```powershell
> go install 包路径@版本
> ```

---

### 1.10.3 从本章知识看 Gin 的 `main`

Gin 官方示例（现在不需要看懂细节）：

```go
package main

import "github.com/gin-gonic/gin"

func main() {
    r := gin.Default()
    r.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "pong"})
    })
    r.Run(":8080")
}
```

用本章知识观察：

| 代码 | 本章能理解的部分 |
|------|------------------|
| `package main` | 可执行程序入口 |
| `import "github.com/gin-gonic/gin"` | 第三方模块的导入路径 |
| `func main()` | 程序入口函数 |
| `gin.Default()` | 调用 gin 包的函数 |
| `func(c *gin.Context) { ... }` | 匿名函数（第 03 章） |
| `*gin.Context` | 指针类型（第 06 章） |
| `gin.H{...}` | 一个 map 类型（第 04 章） |

> 现在只要知道：**Gin 示例的结构和你刚写的 Hello World 完全一样**，只是调用了更多函数。

---

### 1.10.4 GORM 的导入路径

GORM 的典型导入：

```go
import (
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)
```

对应关系：

| 包 | 作用 |
|----|------|
| `gorm.io/gorm` | GORM 核心 |
| `gorm.io/driver/mysql` | MySQL 驱动 |
| `gorm.io/driver/postgres` | PostgreSQL 驱动 |
| `gorm.io/driver/sqlite` | SQLite 驱动 |

这些路径同样是「模块路径 + 子目录」，和本章的 `hello-go/mathutil` 是一个道理。

---

## 1.11 本章最容易犯的 8 个错误

1. **在空目录直接 `go run .`**
   先 `go mod init`。

2. **多文件 `package main` 却用 `go run main.go`**
   改用 `go run .`。

3. **左大括号换行**
   把 `{` 放到声明同一行。

4. **导入包但没使用**
   删除导入，或先用起来。

5. **局部变量声明但没使用**
   Go 会直接编译失败，使用或删除。

6. **大小写写错**
   `fmt.Println` 不是 `fmt.println`；`mathutil.Add` 不是 `mathutil.add`。

7. **模块路径和导入路径混淆**
   导入子包时是 `模块路径/子目录`。

8. **在错误的目录执行命令**
   先 `Get-Location` / `pwd` 确认在模块根目录。

---

## 1.12 练习题

### 基础题

1. 安装 Go，执行 `go version` 和 `go env`，把 `GOROOT`、`GOPATH`、`GOMODCACHE`、`GOPROXY` 记录到笔记里。

2. 新建模块 `hello-go`，写 `main.go` 输出 `Hello, Go!`，分别用 `go run .` 和 `go build .` 运行。

3. 修改 `main.go`，让它输出你的名字和今天的日期（日期先用字符串写死即可）。

4. 用 `go doc` 查下面三个函数的签名并抄写：
   - `fmt.Println`
   - `strings.Contains`
   - `strconv.Atoi`

### 进阶题

5. 创建多包项目：
   ```
   hello-go/
   ├── go.mod
   ├── main.go
   ├── helper.go
   └── mathutil/
       └── mathutil.go
   ```
   - `mathutil.Add` 返回两数之和；
   - `helper.go` 提供一个 `greet()` 函数；
   - `main.go` 调用两者并输出结果。

6. 把 `mathutil.Add` 改成小写 `add`，观察并记录编译错误；再改回大写。

7. 执行 `go run main.go`（在多文件 `package main` 的情况下），记录错误并解释原因。

8. 故意制造下面 5 个错误，分别记录报错原文和修复方法：
   - 未定义变量；
   - 局部变量未使用；
   - 导入未使用；
   - 类型不匹配；
   - 左大括号换行。

9. 解释 `go.mod` 和 `go.sum` 的区别。只用标准库的项目为什么可能没有 `go.sum`？

### 读文档题

10. 打开 <https://pkg.go.dev/fmt>，找到 `Printf`，回答：
    - 它的参数是什么？
    - 返回值是什么？
    - `%v`、`%d`、`%s` 分别用于什么类型？（大致回答即可，第 12 章会详细讲）

### 选做题

11. 执行 `go get github.com/gin-gonic/gin@latest`，然后：
    - 查看 `go.mod` 新增了哪些内容；
    - 执行 `go list -m all`，观察依赖树；
    - 暂时不需要写 Gin 代码。

---

## 1.13 验收清单

全部打勾后，再进入第 02 章：

- [√ ] 能独立安装 Go，并用 `go version` 验证；
- [ ] 能说出 `GOROOT`、`GOPATH`、`GOMODCACHE`、`GOPROXY` 的作用；
- [√ ] 能独立创建模块并运行程序；
- [√ ] 能解释 `go.mod` 和 `go.sum` 各自的作用；
- [√ ] 能说出 `go run`、`go build`、`go install` 的区别；
- [√ ] 能解释「一个目录一个包」和「模块路径 + 子目录 = 导入路径」；
- [√ ] 能说出 `package main` 和 `func main()` 的作用；
- [√ ] 能解释大小写可见性，并知道 GORM 为什么要求字段大写；
- [√ ] 能读懂本章列出的常见编译错误；
- [√ ] 能用 `go doc` 查任意标准库函数签名；
- [√ ] 能说出 Go 与 TS 在类型、空值、类、错误、并发上的主要差异。

---

## 1.14 本章小结

| 概念 | 一句话解释 | TS 类比 |
|------|------------|---------|
| 模块 | 由 `go.mod` 定义的包集合 | 一个 npm 项目 |
| 包 | 一个目录下的 `.go` 文件集合 | 一个命名空间 |
| 导入路径 | 模块路径 + 子目录 | import 路径 |
| `package main` | 可执行程序入口包 | 入口文件 |
| `func main()` | 程序开始执行的地方 | `main()` |
| 零值 | 变量声明后自动得到的默认值 | 没有直接对应 |
| `nil` | 指针/切片/map/chan/func/接口的空值 | `null`（但限制更多） |
| 大小写可见性 | 大写导出，小写包内私有 | `export` |
| `go run .` | 编译并运行当前包 | `ts-node` |
| `go build .` | 编译当前包，生成可执行文件 | `tsc` / 打包 |
| `go doc` | 终端查文档 | 编辑器的类型提示 |
| `go vet` | 静态检查 | ESLint 的静态规则 |

---

## 1.15 下一章预告

第 02 章：**变量、常量与类型系统**

你将学到：

- `var`、`:=`、`const` 的完整用法；
- 所有基本类型，以及 `int` 和 `int64` 为什么不能混用；
- 零值、`nil` 的细节；
- 类型转换、类型别名与新类型；
- 作用域与变量遮蔽；
- `iota` 枚举。

学完后，你会真正理解 GORM 的「零值陷阱」是怎么来的。
