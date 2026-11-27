# 第 13 章：常用标准库（下）—— JSON、IO、HTTP、SQL
> **阶段**：阶段三 · 并发与工程基础
> **建议时间**：2–3 天
> **前置知识**：第 03–08、11–12 章（函数、结构体与标签、接口、错误、`context`、`time`）
> **代码基准**：Go 1.22 及以上

---

## 本章目标
学完本章，你应当能够：

1. 熟练使用 `encoding/json`，解释标签、`omitempty`、导出字段与数字 `float64` 问题；
2. 掌握 `io`、`bufio`、`os`、`filepath` 的常见用法；
3. 理解 `net/http` 的 Handler 模型，能写路由、中间件、超时与优雅关闭；
4. 会用 `database/sql` 完成 CRUD、预编译、连接池与 `context` 取消；
5. 会用 `sql.NullString`、`driver.Valuer`、`sql.Scanner`；
6. 说清 Gin 与 `net/http`、GORM 与 `database/sql` 的封装关系。

---

## 13.1 本章地图
`encoding/json` 对应 `c.JSON` / `ShouldBindJSON`；`io`/`bufio` 对应请求体与文件流；`os`/`filepath` 对应配置与静态文件；`net/http` 是 Gin 的底座；`database/sql` 是 GORM 的底座。

一句话：**Gin 是在 `net/http` 上增强路由与上下文；GORM 是在 `database/sql` 上增强模型映射与链式 API。**

| 主题 | TypeScript | Go |
|------|------------|-----|
| JSON | `JSON.stringify` / `JSON.parse` | `json.Marshal` / `json.Unmarshal` |
| 流 | Node Stream | `io.Reader` / `io.Writer` |
| HTTP 中间件 | Express `(req, res, next)` | `func(http.Handler) http.Handler` |
| 数据库 | `mysql2` / `pg` + 连接池 | `database/sql` + `sql.DB` |

---

## 13.2 `encoding/json`
### 13.2.1 `Marshal`：Go 值 → JSON
```go
type User struct {
    ID       int    `json:"id"`
    Name     string `json:"name"`
    Email    string `json:"email,omitempty"`
    Password string `json:"-"`
    age      int
}

u := User{ID: 1, Name: "张三", Password: "secret", age: 18}
b, err := json.Marshal(u)
fmt.Println(string(b), err)
```
输出：
```
{"id":1,"name":"张三"} <nil>
```
规则：

- **只有导出字段（首字母大写）会被序列化**；`age` 被静默忽略，不报错；
- 没有标签时默认用字段名；`json:"id"` 指定键名；
- `json:"-"` 永不输出；`omitempty` 在零值时省略；
- `Password` 被排除，`Email` 为空字符串且带 `omitempty`，也被省略。

| 标签 | 作用 |
|------|------|
| `json:"name"` | 指定键名 |
| `json:"name,omitempty"` | 零值省略 |
| `json:"-"` | 完全忽略 |
| `json:"age,string"` | 数字/布尔编码成 JSON 字符串 |

`omitempty` 视为空的值：`false`、`0`、`""`、`nil` 指针/接口、长度为 0 的 slice/map/array。**结构体零值不算空**，例如 `time.Time{}` 仍输出 `"0001-01-01T00:00:00Z"`。

> GORM 的 JSON 字段、Gin 的 `c.JSON` 都走这套规则。字段忘大写，接口就返回 `{}`。

### 13.2.2 `Unmarshal`：JSON → Go 值
```go
data := []byte(`{"id":2,"name":"李四","unknown":true}`)

var u User
err := json.Unmarshal(data, &u)
fmt.Println(u.ID, u.Name, u.Email, err)
```
输出：
```
2 李四  <nil>
```
- 第二个参数必须是**指针**，否则报 `json: Unmarshal(non-pointer main.User)`；
- 未知字段默认忽略；类型不匹配返回错误，但已匹配字段可能已被部分写入；
- 严格接口用 `Decoder.DisallowUnknownFields()`。

更新请求常用指针区分「没传」与「传了零值」：
```go
type UpdateUserReq struct {
    Name *string `json:"name,omitempty"`
    Age  *int    `json:"age,omitempty"`
}

name, age := "王五", 0
b, _ := json.Marshal(UpdateUserReq{Name: &name, Age: &age})
fmt.Println(string(b)) // {"name":"王五","age":0}
```
没有 `age` 时 `Age == nil`；传 `"age":0` 时 `Age != nil` 且 `*Age == 0`。

### 13.2.3 `interface{}` 里的数字是 `float64`
```go
var v any
_ = json.Unmarshal([]byte(`{"age":18,"score":95.5}`), &v)
m := v.(map[string]any)
fmt.Printf("%T %v\n", m["age"], m["age"])
```
输出：
```
float64 18
```
JSON 只有一个 `number` 类型，默认解码为 `float64`。后果：大整数（订单号、雪花 ID）可能丢精度，`m["age"].(int)` 会 panic。两种修复：
```go
// 1. 解析到具体结构体（首选）
type Payload struct{ Age int `json:"age"` }

// 2. 使用 UseNumber
dec := json.NewDecoder(strings.NewReader(`{"age":18}`))
dec.UseNumber()
var m2 map[string]any
_ = dec.Decode(&m2)
fmt.Printf("%T %v\n", m2["age"], m2["age"]) // json.Number 18
```
`json.Number` 是字符串类型，需要时调用 `Int64()` 或 `Float64()`。金额不要用 `float64`，用「分」的 `int64` 或字符串。

### 13.2.4 时间序列化
`time.Time` 默认使用 RFC 3339：
```go
type Event struct {
    Name string    `json:"name"`
    At   time.Time `json:"at"`
}

e := Event{
    Name: "上线",
    At:   time.Date(2026, 10, 18, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
}
b, _ := json.Marshal(e)
fmt.Println(string(b))
```
输出：
```
{"name":"上线","at":"2026-10-18T12:00:00+08:00"}
```
反序列化同样要求 RFC 3339；传 `"2026-10-18 12:00:00"` 会报解析错误。自定义格式需要定义新类型并实现 `MarshalJSON` / `UnmarshalJSON`：前者用 `json.Marshal(t.Format("2006-01-02"))`，后者先解码成 `string` 再 `time.Parse`。

### 13.2.5 `json.RawMessage`：延迟解析
```go
type Envelope struct {
    Type string          `json:"type"`
    Data json.RawMessage `json:"data"`
}

var env Envelope
_ = json.Unmarshal([]byte(`{"type":"user","data":{"id":1,"name":"张三"}}`), &env)
fmt.Println(env.Type, string(env.Data))
```
输出：
```
user {"id":1,"name":"张三"}
```
之后按 `Type` 再解析 `Data`（`json.Unmarshal(env.Data, &u)`）。典型用途：消息队列事件信封、Webhook、插件协议。

### 13.2.6 `Decoder` / `Encoder` 与 HTML 转义
`Decoder` 直接读 `io.Reader`，适合请求体和 NDJSON：
```go
const stream = `{"id":1,"name":"A"}
{"id":2,"name":"B"}
`
dec := json.NewDecoder(strings.NewReader(stream))
for {
    var u User
    err := dec.Decode(&u)
    if err == io.EOF {
        break
    }
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(u.ID, u.Name)
}
```
输出：
```
1 A
2 B
```
`Encoder` 直接写 `io.Writer`，`Encode` 会自动补换行：
```go
var buf bytes.Buffer
enc := json.NewEncoder(&buf)
enc.SetEscapeHTML(false)
_ = enc.Encode(map[string]string{"url": "https://example.com?a=1&b=2"})
fmt.Print(buf.String())
```
输出：
```
{"url":"https://example.com?a=1&b=2"}
```
默认会转义 `<`、`>`、`&` 为 `\u003c`、`\u003e`、`\u0026`，防止把 JSON 嵌进 HTML 时产生 XSS。纯 API 可以用 `SetEscapeHTML(false)`。
```go
dec.DisallowUnknownFields() // 拒绝未知字段
dec.UseNumber()             // 数字用 json.Number
enc.SetIndent("", "  ")     // 美化输出，调试用
```
HTTP 响应常用写法：
```go
w.Header().Set("Content-Type", "application/json; charset=utf-8")
json.NewEncoder(w).Encode(resp)
```

---

## 13.3 `io`：一切皆流
### 13.3.1 四个核心接口
```go
type Reader interface{ Read(p []byte) (n int, err error) }
type Writer interface{ Write(p []byte) (n int, err error) }
type Closer interface{ Close() error }
type Seeker interface{ Seek(offset int64, whence int) (int64, error) }
```
| 接口 | 含义 | 常见实现 |
|------|------|----------|
| `io.Reader` | 能读出字节 | 文件、`strings.Reader`、请求体 |
| `io.Writer` | 能写入字节 | 文件、`bytes.Buffer`、响应 |
| `io.Closer` | 能关闭释放资源 | 文件、连接、`sql.Rows` |
| `io.Seeker` | 能移动读写位置 | 文件、`bytes.Reader` |

组合接口：`io.ReadWriter`、`io.ReadCloser`、`io.WriteCloser`、`io.ReadSeekCloser`。Go 接口隐式实现，方法匹配就能用。

### 13.3.2 `Copy`、`ReadAll`、`LimitReader`
```go
src := strings.NewReader("hello, go")
var buf bytes.Buffer
n, err := io.Copy(&buf, src)
fmt.Println(n, err, buf.String())
```
输出：
```
9 <nil> hello, go
```
- `io.Copy(dst, src)` 用固定缓冲区拷贝，适合大文件、网络转发；
- `io.ReadAll(r)` 一次读完，适合小数据；
- `io.LimitReader(r, n)` 最多读 n 字节；
- `io.CopyN`、`io.ReadFull`、`io.MultiWriter`、`io.TeeReader` 也常用。

不要对不可信的大请求直接 `ReadAll`，先用 `http.MaxBytesReader` 限制大小。

### 13.3.3 `io.EOF`
```go
r := strings.NewReader("abc")
buf := make([]byte, 2)
for {
    n, err := r.Read(buf)
    fmt.Printf("n=%d data=%q err=%v\n", n, buf[:n], err)
    if err != nil { break }
}
// n=2 data="ab" err=<nil>; n=1 data="c" err=<nil>; n=0 data="" err=EOF
```
- 先处理本次读到的 `n` 字节，再判断 `err`；
- 最后一次读取可能同时返回数据和 `io.EOF`；
- `io.EOF` 表示正常结束，不是异常；
- 只有 `err != nil && err != io.EOF` 才是真错误。

### 13.3.4 `Closer` 与 `defer`
```go
f, err := os.Open("data.txt")
if err != nil {
    return err
}
defer f.Close()
```
写文件时关闭错误可能代表数据未落盘，严谨写法是在 `defer` 里检查 `f.Close()` 的返回值。HTTP 服务端会自动关闭 `Request.Body`；客户端拿到的 `Response.Body` 必须自己关闭，否则连接无法复用。

---

## 13.4 `bufio`：带缓冲的读写
```go
sc := bufio.NewScanner(strings.NewReader("Go\nGin\nGORM\n"))
for sc.Scan() {
    fmt.Println(sc.Text())
}
if err := sc.Err(); err != nil {
    fmt.Println("scan error:", err)
}
```
输出：
```
Go
Gin
GORM
```
- `Scan()` 返回 `false` 表示结束或出错，必须用 `Err()` 区分；
- 默认按行切分，去掉 `\n`/`\r\n`；
- 默认单个 token 最大 64 KB，超长报 `bufio.Scanner: token too long`；
- 扩大缓冲：`sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)`；
- 其他切分函数：`ScanWords`、`ScanRunes`、`ScanBytes`。

`bufio.Reader` 提供 `ReadString`、`ReadBytes`、`ReadRune`、`Peek`：
```go
r := bufio.NewReader(strings.NewReader("第一行\n第二行\n"))
line, _ := r.ReadString('\n')
fmt.Printf("%q\n", line) // "第一行\n"
```
`bufio.Writer` 会缓冲，**必须 `Flush`**：
```go
var buf bytes.Buffer
w := bufio.NewWriter(&buf)
w.WriteString("hello\n")
w.WriteString("world\n")
fmt.Println("buffered len:", buf.Len()) // 0
w.Flush()
fmt.Print(buf.String())
```
输出：
```
buffered len: 0
hello
world
```
忘记 `Flush` 是经典错误：数据还在缓冲区，文件或网络里什么都没有。
```go
defer func() {
    if err := w.Flush(); err != nil {
        log.Printf("flush: %v", err)
    }
}()
```
HTTP 流式响应还需要 `http.Flusher`：`if f, ok := w.(http.Flusher); ok { f.Flush() }`。

---

## 13.5 `os`：文件、环境变量与进程
### 13.5.1 文件读写
```go
dir, _ := os.MkdirTemp("", "go-demo-*")
defer os.RemoveAll(dir)

path := filepath.Join(dir, "hello.txt")
if err := os.WriteFile(path, []byte("你好，Go\n"), 0644); err != nil {
    log.Fatal(err)
}

data, err := os.ReadFile(path)
if err != nil {
    log.Fatal(err)
}
fmt.Print(string(data))
```
输出：
```
你好，Go
```
`os.WriteFile` 会创建或截断；权限 `0644` 只在创建时生效并受 umask 影响；`os.ReadFile` 一次读入内存，适合小文件。

`os.Open` 只读；`os.Create` 等于读写 + 创建 + 截断（0666）；追加写用 `os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)`。常用 flag：`O_RDONLY`、`O_WRONLY`、`O_RDWR`、`O_CREATE`、`O_TRUNC`、`O_APPEND`，用 `|` 组合。

### 13.5.2 环境变量
```go
os.Setenv("APP_ENV", "dev")
fmt.Println(os.Getenv("APP_ENV")) // dev

v, ok := os.LookupEnv("NOT_SET")
fmt.Printf("%q %v\n", v, ok) // "" false
```
`Getenv` 无法区分「未设置」和「设置为空」；`LookupEnv` 可以。全部变量用 `os.Environ()`。

### 13.5.3 `os.Args` 与 `os.Exit`
```go
// go run . -port 8080
fmt.Println(os.Args) // [C:\...\app.exe -port 8080]
```
`os.Args[0]` 是程序名，后面的元素是参数；复杂解析用 `flag`（第 15 章）。
```go
if len(os.Args) < 2 {
    fmt.Fprintln(os.Stderr, "usage: app <name>")
    os.Exit(1)
}
```
**`os.Exit` 不会执行 `defer`**，也不等待其他 goroutine，只适合启动阶段的最后退出。正常流程应 `return`。标准流 `os.Stdin`、`os.Stdout`、`os.Stderr` 都实现了 `io.Reader` / `io.Writer`。

---

## 13.6 `filepath`：跨平台路径
```go
p := filepath.Join("data", "users", "1.json")
fmt.Println(p)                   // Windows: data\users\1.json
fmt.Println(filepath.ToSlash(p)) // data/users/1.json
fmt.Println(filepath.Base(p))    // 1.json
fmt.Println(filepath.Dir(p))     // data\users
fmt.Println(filepath.Ext(p))     // .json
```
| 函数 | 作用 |
|------|------|
| `Join` | 拼接并清理路径 |
| `Base` / `Dir` | 最后一段 / 去掉最后一段 |
| `Ext` | 扩展名（含点），`a.tar.gz` → `.gz` |
| `Abs` / `Rel` | 绝对 / 相对路径 |
| `Clean` / `IsAbs` | 清理 / 是否绝对路径 |
| `ToSlash` / `FromSlash` | 分隔符转换 |

注意：

- 不要手写 `"a/" + "b"`，用 `filepath.Join`；
- **URL 路径用 `path` 包**，URL 永远用 `/`，不要用 `filepath`；
- 展示给前端时可用 `filepath.ToSlash`。

`Glob`：
```go
matches, err := filepath.Glob(filepath.Join(dir, "*.txt"))
```
- 没有匹配时返回空切片和 `nil`，不是错误；
- 结果按字典序排序；
- 只支持 `*`、`?`、`[...]`，不支持 `**`。

`filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {...})` 递归遍历，按字典序；回调返回 `filepath.SkipDir` 可跳过目录，返回 `err` 终止遍历。

---

## 13.7 `net/http`
### 13.7.1 `Handler` 与 `HandlerFunc`
```go
type Handler interface {
    ServeHTTP(w ResponseWriter, r *Request)
}

type HandlerFunc func(ResponseWriter, *Request)

func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) { f(w, r) }
```
任何实现了 `ServeHTTP` 的类型都是 Handler。普通函数通过 `http.HandlerFunc` 适配：
```go
mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, "hello")
})
```
### 13.7.2 `ServeMux`：路由
Go 1.22 起支持方法、路径参数和通配符：
```go
mux := http.NewServeMux()

mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "user %s", r.PathValue("id"))
})
mux.HandleFunc("POST /users", createUser)
mux.HandleFunc("GET /files/{path...}", serveFile)
```
- `{id}` 匹配一段路径，用 `PathValue` 取值；
- `{path...}` 匹配剩余全部路径；
- 路径匹配但方法不匹配时返回 405；
- 更具体的模式优先；以 `/` 结尾的模式匹配子树。

用 `httptest` 验证：
```go
req := httptest.NewRequest(http.MethodGet, "/users/42", nil)
rec := httptest.NewRecorder()
mux.ServeHTTP(rec, req)
fmt.Printf("%d %q\n", rec.Code, rec.Body.String())
```
输出：
```
200 "user 42"
```
Go 1.22 之前的项目只能写 `mux.HandleFunc("/users", ...)`，方法判断在 handler 里做。

### 13.7.3 `Server` 与超时
生产环境不要直接用没有超时的 `http.ListenAndServe`：
```go
srv := &http.Server{
    Addr:              ":8080",
    Handler:           mux,
    ReadHeaderTimeout: 5 * time.Second,
    ReadTimeout:       10 * time.Second,
    WriteTimeout:      10 * time.Second,
    IdleTimeout:       60 * time.Second,
}
if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
    log.Fatal(err)
}
```
| 超时 | 控制什么 | 防什么问题 |
|------|----------|------------|
| `ReadHeaderTimeout` | 读请求头的最长时间 | Slowloris 攻击 |
| `ReadTimeout` | 读整个请求（含 body） | 慢速上传 |
| `WriteTimeout` | 写响应的最长时间 | 慢客户端占连接 |
| `IdleTimeout` | keep-alive 空闲时间 | 连接长期不释放 |

### 13.7.4 `Request` 与 `ResponseWriter`
常用请求字段：`r.Method`、`r.URL.Path`、`r.URL.Query()`、`r.Header.Get("Authorization")`、`r.Body`、`r.Context()`、`r.PathValue("id")`、`r.RemoteAddr`。

读取 JSON 请求体：
```go
func createUser(w http.ResponseWriter, r *http.Request) {
    r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 最大 1 MB
    defer r.Body.Close()

    var req struct {
        Name string `json:"name"`
    }
    dec := json.NewDecoder(r.Body)
    dec.DisallowUnknownFields()
    if err := dec.Decode(&req); err != nil {
        http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"name": req.Name})
}
```
`ResponseWriter`：
```go
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusCreated)
w.Write([]byte(`{"ok":true}`))
```
**顺序非常重要**：

1. 先改 `Header()`；
2. 再 `WriteHeader`（可省略，默认 200）；
3. 最后 `Write`；
4. 响应头一旦写出，再改 Header 或再 `WriteHeader` 都无效；
5. 不设置 `Content-Type` 时，Go 根据前 512 字节猜测。

### 13.7.5 中间件本质：包装 Handler
```go
func logging(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
    })
}
```
假设 `recoverer`、`auth` 也是同样的中间件：
```go
handler := recoverer(logging(auth(mux)))
```
请求顺序：`recoverer → logging → auth → mux`；响应阶段反向执行 `next` 之后的代码。

捕获 panic 的中间件同理：在 `defer` 里 `recover()`，记录日志并 `http.Error(w, "internal server error", 500)`。如果响应已开始写出，`http.Error` 只能追加内容，状态码改不了。

### 13.7.6 优雅关闭
```go
func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("pong"))
    })
    srv := &http.Server{Addr: ":8080", Handler: mux, ReadHeaderTimeout: 5 * time.Second}

    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    go func() {
        if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
            log.Fatalf("server: %v", err)
        }
    }()

    <-ctx.Done()
    stop()
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := srv.Shutdown(shutdownCtx); err != nil {
        log.Printf("shutdown: %v", err)
    }
    log.Println("server exited")
}
```
按 `Ctrl+C` 后会打印 `server exited`。

`Shutdown` 停止接收新连接，等待正在处理的请求完成，超时后强制关闭。`ListenAndServe` 返回 `http.ErrServerClosed` 是正常现象。

> 客户端补充：默认 `http.Get` 没有超时，生产环境用 `&http.Client{Timeout: 5 * time.Second}`；必须 `defer resp.Body.Close()`；4xx/5xx 不会让 `err != nil`，要检查 `resp.StatusCode`。

---

## 13.8 `database/sql`
### 13.8.1 `Open` 不会真正连接
```go
import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
)

dsn := "root:123456@tcp(127.0.0.1:3306)/demo?parseTime=true&loc=Local"
db, err := sql.Open("mysql", dsn)
if err != nil {
    log.Fatal(err)
}
defer db.Close()

ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()
if err := db.PingContext(ctx); err != nil {
    log.Fatal(err)
}
```
- 驱动通过空白导入 `_` 注册；`sql.Open` 只创建连接池对象，**通常不连接数据库**；
- 验证连接用 `PingContext`；
- `sql.DB` 是连接池，不是单个连接，可并发使用；
- 不要每次请求 `sql.Open`，全局初始化一次。

DSN 的 `parseTime=true` 让 MySQL 把时间列扫描成 `time.Time`，`loc=Local` 指定时区。

### 13.8.2 `Query`、`QueryRow`、`Exec`、`Scan`
多行查询：
```go
rows, err := db.QueryContext(ctx,
    "SELECT id, name FROM users WHERE age > ? ORDER BY id", 18)
if err != nil {
    return err
}
defer rows.Close()

for rows.Next() {
    var id int64
    var name string
    if err := rows.Scan(&id, &name); err != nil {
        return err
    }
    fmt.Println(id, name)
}
if err := rows.Err(); err != nil {
    return err
}
```
单行查询：
```go
var name string
err := db.QueryRowContext(ctx,
    "SELECT name FROM users WHERE id = ?", 1).Scan(&name)

if errors.Is(err, sql.ErrNoRows) {
    fmt.Println("user not found")
} else if err != nil {
    return err
}
```
写操作：
```go
res, err := db.ExecContext(ctx,
    "UPDATE users SET name = ? WHERE id = ?", "李四", 1)
if err != nil {
    return err
}
affected, _ := res.RowsAffected()
fmt.Println("affected:", affected)
```
`Scan` 的规则：参数必须是指针；数量、顺序、类型要与列匹配；`NULL` 扫描到 `string`/`int` 会报错，需要 `sql.NullString` 或指针。**`rows.Err()` 不能省**，`Next()` 返回 `false` 可能是正常结束，也可能是中途出错。

### 13.8.3 预编译语句与占位符
```go
stmt, err := db.PrepareContext(ctx, "SELECT name FROM users WHERE id = ?")
if err != nil {
    return err
}
defer stmt.Close()

var name string
err = stmt.QueryRowContext(ctx, 1).Scan(&name)
```
优点：参数与 SQL 分离，防注入；数据库可复用执行计划；适合高频固定 SQL。

| 数据库 | 占位符 |
|--------|--------|
| MySQL / SQLite | `?` |
| PostgreSQL | `$1`、`$2` |
| SQL Server | `@p1` |

**永远不要拼接用户输入**：
```go
// 错误：SQL 注入
query := "SELECT * FROM users WHERE name = '" + name + "'"
// 正确
db.QueryContext(ctx, "SELECT * FROM users WHERE name = ?", name)
```
### 13.8.4 事务
```go
tx, err := db.BeginTx(ctx, nil)
if err != nil { return err }
defer tx.Rollback() // 已 Commit 时返回 sql.ErrTxDone，可忽略

if _, err := tx.ExecContext(ctx,
    "UPDATE accounts SET balance = balance - ? WHERE id = ?", 100, 1); err != nil {
    return err
}
return tx.Commit() // 两笔 UPDATE 之间同理
```
`defer tx.Rollback()` 是保险写法：任何提前 `return` 都会回滚。

### 13.8.5 连接池
```go
db.SetMaxOpenConns(20)                  // 最大打开连接数
db.SetMaxIdleConns(10)                  // 最大空闲连接数
db.SetConnMaxLifetime(time.Hour)        // 连接最长寿命
db.SetConnMaxIdleTime(10 * time.Minute) // 连接最长空闲时间
```
| 设置 | 默认 |
|------|------|
| `MaxOpenConns` | 0（不限制） |
| `MaxIdleConns` | 2 |
| `ConnMaxLifetime` | 0（不限制） |
| `ConnMaxIdleTime` | 0（不限制） |

经验：`MaxIdleConns` 不大于 `MaxOpenConns`；`MaxOpenConns` 小于数据库 `max_connections` 并留余量；寿命小于负载均衡的空闲断连时间；池子不是越大越好。

### 13.8.6 `context` 版本方法
Web 服务中优先使用带 `Context` 的方法：
```go
db.PingContext(ctx)
db.QueryContext(ctx, ...)
db.QueryRowContext(ctx, ...)
db.ExecContext(ctx, ...)
db.PrepareContext(ctx, ...)
db.BeginTx(ctx, nil)
```
把 `r.Context()` 传下去，客户端断开或超时后数据库操作会被取消。GORM 对应 `db.WithContext(ctx)`。

### 13.8.7 `sql.NullString` 等 NULL 类型
```go
// type NullString struct { String string; Valid bool }
var email sql.NullString
err := db.QueryRowContext(ctx,
    "SELECT email FROM users WHERE id = ?", 1).Scan(&email)
if err != nil {
    return err
}
if email.Valid {
    fmt.Println("email:", email.String)
} else {
    fmt.Println("email is NULL")
}
```
同类：`NullInt64`、`NullFloat64`、`NullBool`、`NullTime`。

`NullString` 直接 JSON 序列化会得到 `{"String":"a@b.com","Valid":true}`，通常不是前端想要的。API 层应转换成 `*string` 或自定义 `MarshalJSON`。

### 13.8.8 `driver.Valuer` 与 `sql.Scanner`
```go
type Valuer interface{ Value() (driver.Value, error) }
type Scanner interface{ Scan(src any) error }
```
示例：限制状态枚举。
```go
type Status string

const (
    StatusActive Status = "active"
    StatusBanned Status = "banned"
)

func (s Status) Value() (driver.Value, error) {
    switch s {
    case StatusActive, StatusBanned:
        return string(s), nil
    default:
        return nil, fmt.Errorf("invalid status %q", s)
    }
}

func (s *Status) Scan(src any) error {
    if src == nil {
        *s = ""
        return nil
    }
    switch v := src.(type) {
    case []byte:
        *s = Status(v)
    case string:
        *s = Status(v)
    default:
        return fmt.Errorf("cannot scan %T into Status", src)
    }
    return nil
}
```
- `Value` 用值接收者；`Scan` 用指针接收者；
- `driver.Value` 只允许 `int64`、`float64`、`bool`、`[]byte`、`string`、`time.Time`、`nil`；
- GORM 的 JSON 字段、加密字段、枚举同样依赖这两个接口。

### 13.8.9 完整操作流程（MySQL）
建表：
```sql
CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(50) NOT NULL,
    email VARCHAR(100) NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```
安装驱动后：
```go
// go get github.com/go-sql-driver/mysql
func run(ctx context.Context, db *sql.DB) error {
    res, err := db.ExecContext(ctx,
        "INSERT INTO users(name, email) VALUES(?, ?)", "张三", nil)
    if err != nil { return err }
    id, _ := res.LastInsertId()

    var u struct {
        ID        int64
        Name      string
        Email     sql.NullString
        CreatedAt time.Time
    }
    err = db.QueryRowContext(ctx,
        "SELECT id, name, email, created_at FROM users WHERE id = ?", id).
        Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
    if errors.Is(err, sql.ErrNoRows) { return fmt.Errorf("user %d not found", id) }
    if err != nil { return err }

    fmt.Println(u.ID, u.Name, u.Email.Valid,
        u.CreatedAt.Format("2006-01-02 15:04:05"))
    return nil
}
```
数据库为空、自增从 1 开始时，`id` 为 1，`u.Email.Valid` 为 false，`u.CreatedAt` 可用 `Format("2006-01-02 15:04:05")` 输出。

换成 SQLite 时驱动名、DSN、方言不同。`modernc.org/sqlite` 是纯 Go 驱动（驱动名 `"sqlite"`，可用 `":memory:"`）；GORM 的 `gorm.io/driver/sqlite` 在 Windows 上通常需要 CGO。

---

## 13.9 与 Gin/GORM 的关联
### 13.9.1 Gin 建立在 `net/http` 上
| Gin | 标准库对应 |
|-----|------------|
| `*gin.Engine` | 实现 `http.Handler` |
| `*gin.Context` | 包装 `*http.Request` 与响应写入器 |
| `gin.HandlerFunc` | `func(*gin.Context)`，思路同 `http.HandlerFunc` |
| `c.Request` | `*http.Request` |
| `c.Writer` | 扩展的 `http.ResponseWriter` |
| `c.JSON` | 设置类型 + `encoding/json` |
| `c.ShouldBindJSON` | 读 `Request.Body` + JSON 解码 |
| `r.Run(":8080")` | 底层 `http.ListenAndServe` |

`Engine` 的关键方法：
```go
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request)
```
所以可以把 Gin 挂到标准库服务器上：`srv := &http.Server{Addr: ":8080", Handler: r, ReadHeaderTimeout: 5 * time.Second}`，其中 `r` 是 `*gin.Engine`。

**重要实践**：`r.Run()` 内部使用 `http.ListenAndServe`，没有超时，也没有优雅关闭。生产环境应自己创建 `http.Server`，把 `Engine` 作为 `Handler`，再调用 `Shutdown`（同 13.7.6）。

中间件对比：

| 标准库 | Gin |
|--------|-----|
| `func(http.Handler) http.Handler` | `func(*gin.Context)` |
| 包装整个 Handler | 共享 `Context`，`c.Next()` 推进 |
| 每个中间件独立对象 | 所有 handler 在一条 `HandlersChain` 中 |

本质相同：都是在目标处理逻辑前后插入代码。`c.Next()` / `c.Abort()` 在第 17 章展开。

### 13.9.2 GORM 建立在 `database/sql` 上
| GORM | `database/sql` 对应 |
|------|---------------------|
| `gorm.Open` | `sql.Open` + 方言 |
| `*gorm.DB` | 持有连接池，底层 `*sql.DB` |
| `Create` / `Update` / `Delete` | `ExecContext` |
| `First` / `Take` | `QueryRowContext` + `Scan` |
| `Find` | `QueryContext` + 循环 `Scan` |
| `WithContext(ctx)` | 带 `Context` 的方法 |
| `gorm.ErrRecordNotFound` | 对应 `sql.ErrNoRows` 的封装 |
| 自定义类型 | `driver.Valuer` + `sql.Scanner` |
| `PrepareStmt: true` | 底层预编译 |

取底层连接池：
```go
sqlDB, err := db.DB() // *gorm.DB -> *sql.DB
if err != nil {
    return err
}
sqlDB.SetMaxOpenConns(20)
sqlDB.SetMaxIdleConns(10)
sqlDB.SetConnMaxLifetime(time.Hour)
```
传请求上下文：
```go
var u User
err := db.WithContext(ctx).First(&u, id).Error
if errors.Is(err, gorm.ErrRecordNotFound) { return nil, ErrUserNotFound }
if err != nil { return nil, err }
return &u, nil
```
读 GORM 文档时就能对应上：「连接池」是底层 `*sql.DB`；「自定义类型」是 `Scanner`/`Valuer`；「`WithContext`」是向下传 `ctx`；「`ErrRecordNotFound`」是「没有行」的封装。

---

## 13.10 常见错误与陷阱
| 现象 | 原因 | 修复 |
|------|------|------|
| `json.Marshal` 得到 `{}` | 字段未导出 | 首字母大写 |
| `json: Unmarshal(non-pointer ...)` | 传值而非指针 | 传 `&v` |
| 数字断言 `int` 失败 | `any` 中默认 `float64` | 结构体或 `UseNumber` |
| `omitempty` 没省略时间 | 结构体零值不算空 | `*time.Time` 或自定义类型 |
| 读文件把 EOF 当错误 | EOF 是正常结束 | 单独判断 |
| 文件内容为空 | `bufio.Writer` 没 `Flush` | 调用 `Flush` |
| `superfluous response.WriteHeader` | 重复写状态码 | 只写一次 |
| Header 不生效 | 在 `Write` 后才设置 | 先 Header 后 Write |
| `Open` 成功但连不上 | Open 不建立连接 | `PingContext` |
| `converting NULL to string` | 列可为 NULL | `NullString` / 指针 |
| MySQL 用 `$1` 报错 | 占位符方言不匹配 | MySQL 用 `?` |
| `rows.Next()` 后未查 `Err` | 忽略中途错误 | 检查 `rows.Err()` |
| 请求超时 SQL 还在跑 | 没传 `context` | `QueryContext(r.Context(), ...)` |
| URL 路径出现反斜杠 | 用了 `filepath` | URL 用 `path` |
| 并发写响应崩溃 | `ResponseWriter` 非并发安全 | 每请求一个 goroutine 写 |

```go
db.QueryRow("SELECT ...")                     // 错误：取消不传播
db.QueryRowContext(r.Context(), "SELECT ...") // 正确
db.WithContext(r.Context()).First(&user, id)  // GORM
```

---

## 13.11 练习题
### 基础题
1. 定义 `User`（`ID`、`Name`、`Email`、`Password`），要求 `Password` 不输出、`Email` 为空时省略，并打印 JSON。
2. 把第 1 题 JSON 解析回结构体，打印字段。
3. 用 `*int` 区分「没传 age」和「age 为 0」，写出两种请求体并打印结果。
4. 用 `json.Decoder` 解析包含两个 JSON 对象的字符串，循环打印。
5. 用 `os.WriteFile` 写三行文本，再用 `bufio.Scanner` 逐行读取并打印行号。
6. 用 `filepath` 处理 `data/users/1.json`，打印 `Join`、`Base`、`Dir`、`Ext`、`ToSlash`。
7. 写 `GET /hello` 返回 `{"message":"hello"}`，用 `httptest` 验证状态码和响应体。
8. 写打印方法和耗时的中间件，包装第 7 题的 Handler。
9. 用 `database/sql` 按 id 查询用户，处理 `sql.ErrNoRows`。
10. 写插入、更新、删除 SQL，并打印 `RowsAffected`。

### 进阶题
11. 实现自定义 `Date` 类型，JSON 只接受和输出 `2006-01-02`。
12. 用 `json.RawMessage` 实现事件分发器：`user` 解析成用户，`order` 解析成订单。
13. 用 `io.Copy` + `io.LimitReader` 实现最大 10 MB 的文件复制。
14. 写中间件记录响应状态码和耗时（提示：包装 `ResponseWriter`）。
15. 用 `signal.NotifyContext` + `Shutdown` 改造第 7 题，按 `Ctrl+C` 优雅退出。
16. 用 `database/sql` 实现带 `context` 超时的查询，验证超时返回错误。
17. 给 `Status` 实现 `driver.Valuer` 和 `sql.Scanner`，非法状态写入时报错。
18. 配置连接池（最大 20、空闲 10、寿命 1 小时），解释每个参数。
19. 用 `sql.NullString` 查询可为 NULL 的邮箱，并转换成 `*string`。
20. 用 `http.MaxBytesReader` 限制请求体 1 MB，超出返回 413。

### 读文档题
21. `go doc encoding/json.Marshal`：返回值是什么？哪些类型无法序列化？
22. `go doc io.Copy`：参数和返回值是什么？
23. `go doc net/http.Server`：列出 4 个超时字段及含义。
24. `go doc database/sql.DB.SetMaxOpenConns`：默认值是什么？
25. `go doc database/sql.Scanner`：抄写接口定义，说明 `src` 可能是什么类型。

---

## 13.12 验收清单
- [ ] 能用 `json.Marshal`/`Unmarshal`，解释导出字段、标签、`omitempty`、`float64` 数字，并会用 `RawMessage` 与 `Decoder`/`Encoder`；
- [ ] 能说清 `io.Reader`/`Writer`/`Closer`/`Seeker`，正确处理 `io.EOF`，会用 `bufio` 读写并 `Flush`；
- [ ] 能用 `os` 和 `filepath` 完成文件、环境变量、路径操作，能写 `Handler`、`HandlerFunc`、`ServeMux` 路由和中间件；
- [ ] 能配置 HTTP 超时并实现优雅关闭；
- [ ] 能用 `database/sql` 完成 CRUD、`Scan`、预编译、事务、连接池；
- [ ] 能用 `sql.NullString`、`driver.Valuer`、`sql.Scanner`；
- [ ] 能说清 Gin/GORM 与标准库的对应关系。

---

## 13.13 本章小结
| 主题 | 核心结论 | 易错点 |
|------|----------|--------|
| JSON 序列化 | 只输出导出字段，标签控制键名 | 小写字段静默丢失 |
| `omitempty` | 零值省略，结构体零值不算空 | 时间字段仍输出 |
| JSON 数字 | `any` 中默认 `float64` | 大整数丢精度 |
| 时间 | 默认 RFC 3339 | 自定义格式需实现接口 |
| `RawMessage` | 保留原始 JSON | 解析前不能直接访问字段 |
| `io` | Reader/Writer 是通用流抽象 | 把 EOF 当错误 |
| `bufio` | 减少系统调用，Writer 必须 Flush | 忘记 Flush |
| `os` | 文件、环境变量、进程控制 | `os.Exit` 不执行 defer |
| `filepath` | 跨平台路径 | URL 误用 filepath |
| `net/http` | Handler + ServeMux + Server | 无超时、Header 顺序错误 |
| 中间件 | 包装 Handler | 响应已写后改不了状态码 |
| `database/sql` | `sql.DB` 是连接池 | 以为 Open 会连接 |
| 连接池 | 控制并发与生命周期 | 超过数据库上限 |
| NULL | `NullXxx` 或指针 | 直接 Scan 到 string |
| 自定义类型 | `Valuer` + `Scanner` | `Scan` 用值接收者 |
| Gin | 增强 `net/http` | `r.Run` 无超时、无优雅关闭 |
| GORM | 封装 `database/sql` | 忘记 `WithContext` |

---

## 13.14 下一章预告
第 14 章：**测试与调试**

你将学到：

- `go test`、`_test.go`、`TestXxx`；
- 表驱动测试与 `t.Run` 子测试；
- `t.Helper`、`t.Cleanup`、`t.TempDir`；
- 覆盖率、基准测试、示例测试；
- `httptest` 测试 HTTP Handler；
- 接口 fake 与 `testify/mock`；
- `go vet`、`staticcheck`、Delve、`go test -race`；
- Gin/GORM 项目测试思路。

本章写的 HTTP Handler 和数据库代码，下一章会直接拿来做测试。
