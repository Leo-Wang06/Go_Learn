# 第 17 章：如何读懂 Gin 文档与源码

> **阶段**：阶段五 · 读文档与源码
> **建议时间**：1–2 天
> **前置知识**：第 03、07、13、16 章（函数、接口、`net/http`、最小 Web API）
> **代码基准**：Go 1.22+，Gin v1.9+（本章源码行号以 v1.10.0 为例）

---

## 本章目标

学完本章，你应当能够：

1. 找到 Gin 的官方文档、pkg.go.dev 和源码入口；
2. 画出 `Engine`、`RouterGroup`、`Context`、`HandlerFunc`、`HandlersChain` 的类型关系；
3. 读懂路由注册、`Group`、`Use` 的签名和行为；
4. 区分 `Param`、`Query`、`ShouldBindJSON`、`GetHeader` 的用途；
5. 判断 `JSON`、`String`、`Data`、`HTML`、`Status` 是否会写响应；
6. 解释中间件 `Next`、`Abort`、`Set`、`Get` 的执行顺序；
7. 说清 `Engine` 为什么能作为 `http.Handler`，并顺着 `gin.go`、`routergroup.go`、`context.go` 定位关键实现；
8. 看到任意 Gin 方法签名，能判断它返回什么、是否写响应、是否中止链；
9. 把同样的阅读方法迁移到 GORM。

---

## 17.1 文档入口与阅读顺序

| 入口 | 地址 | 用途 |
|------|------|------|
| 官方文档 | `https://gin-gonic.com/docs/` | 教程、示例、FAQ |
| GitHub README | `https://github.com/gin-gonic/gin` | 快速开始、版本、生态 |
| pkg.go.dev | `https://pkg.go.dev/github.com/gin-gonic/gin` | 精确签名、类型、方法列表 |

日常查 API 优先用 pkg.go.dev，它展示的就是编译器看到的签名；遇到「怎么用」再看官方文档。

推荐顺序：README Quick Start → 官方文档 Routing/Parameters/Binding/Middleware → pkg.go.dev 的 `Engine`/`RouterGroup`/`Context` → 源码 `gin.go`/`routergroup.go`/`context.go` → `middleware/` 和测试。不要从源码第一行开始读，先建立类型地图。

`go doc github.com/gin-gonic/gin.Context.ShouldBindJSON` 这类命令比网页更快，适合确认签名。v1.9 与 v1.10 核心类型基本一致；v1.10 把 `interface{}` 换成了 `any`，两者是别名。本章行号以 v1.10.0 为例，其他版本用函数名搜索。

---

## 17.2 类型地图：先认清核心类型

### 17.2.1 一张关系图

```
Engine
 ├─ embeds RouterGroup ── Handlers HandlersChain
 │                         basePath string
 │                         engine *Engine
 │                         Group() *RouterGroup
 │                         Use() IRoutes
 ├─ trees methodTrees
 └─ ServeHTTP(w http.ResponseWriter, r *http.Request)

HandlersChain = []HandlerFunc
HandlerFunc   = func(*Context)

Context
 ├─ Request *http.Request
 ├─ Writer  ResponseWriter
 ├─ Params  Params
 ├─ handlers HandlersChain / index int8
 └─ Keys map[string]any
```

### 17.2.2 `HandlerFunc`

```go
type HandlerFunc func(*Context)
```

它就是一个普通函数类型：参数是 `*gin.Context`，没有返回值，错误通过 `c.JSON` / `c.Abort` / `c.Error` 表达。所以路由注册本质上是在保存函数：`r.GET("/ping", ping)`。

### 17.2.3 `HandlersChain`

```go
type HandlersChain []HandlerFunc
```

一次请求要执行的函数列表，顺序是「全局中间件 → 分组中间件 → 路由中间件 → 最终 handler」。`Last()` 返回最后一个 handler，`c.Handler()` 和 `c.HandlerName()` 都依赖它。

### 17.2.4 `Engine`

```go
type Engine struct {
    RouterGroup
    RedirectTrailingSlash  bool
    RedirectFixedPath      bool
    HandleMethodNotAllowed bool
    trees methodTrees
}
```

- `Engine` **嵌入** `RouterGroup`，所以 `r.GET`、`r.Use`、`r.Group` 是提升自 `RouterGroup` 的方法；
- `Engine` 自己持有路由树 `trees`；
- `Engine` 实现 `ServeHTTP`，因此可以直接作为 `http.Handler`。

### 17.2.5 `RouterGroup`

```go
type RouterGroup struct {
    Handlers HandlersChain
    basePath string
    engine   *Engine
    root     bool
}
```

`Handlers` 是当前累积的中间件；`basePath` 是分组前缀；`engine` 指回根引擎；`root` 表示它是不是 `Engine` 内嵌的根分组。

### 17.2.6 `IRoutes` 与 `IRouter`

```go
type IRoutes interface {
    Use(...HandlerFunc) IRoutes
    GET(string, ...HandlerFunc) IRoutes
    POST(string, ...HandlerFunc) IRoutes
    Handle(string, string, ...HandlerFunc) IRoutes
    Any(string, ...HandlerFunc) IRoutes
    Static(string, string) IRoutes
    // ... DELETE/PATCH/PUT/OPTIONS/HEAD/StaticFS
}

type IRouter interface {
    IRoutes
    Group(string, ...HandlerFunc) *RouterGroup
}
```

`GET/POST/Use` 返回 `IRoutes`，方便链式调用；`Group` 返回具体的 `*RouterGroup`，因为分组后还要继续注册。`Engine` 和 `RouterGroup` 都实现 `IRouter`。

根分组的 `returnObj()` 返回 `*Engine`，普通分组返回 `*RouterGroup`，所以 `r.Use(...)` 之后还能继续 `r.GET(...)`。

### 17.2.7 `Context`

```go
type Context struct {
    Request *http.Request
    Writer  ResponseWriter
    Params  Params

    handlers HandlersChain
    index    int8
    fullPath string

    mu   sync.RWMutex
    Keys map[string]any
}
```

`Request` 是原始请求；`Writer` 是包装后的响应写入器；`Params` 保存路径参数；`handlers` 和 `index` 控制中间件链；`Keys` 保存中间件之间共享的值，读写有锁保护。大部分字段不导出，只能通过方法访问。

### 17.2.8 `ResponseWriter`

```go
type ResponseWriter interface {
    http.ResponseWriter
    http.Hijacker
    http.Flusher
    Status() int
    Size() int
    Written() bool
    WriteHeaderNow()
}
```

它比 `http.ResponseWriter` 多了状态码、字节数、是否已写出等能力。`c.Writer` 的类型就是它。

### 17.2.9 `gin.H`

`type H map[string]any`，只是 JSON 的快捷写法，没有魔法：`c.JSON(200, gin.H{"message": "pong"})`。

---

## 17.3 路由注册：GET、Group、Use

### 17.3.1 基本注册

```go
r := gin.New()

r.GET("/ping", func(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"message": "pong"})
})

r.POST("/users", createUser)
r.GET("/users/:id", getUser)
```

签名：

```go
func (group *RouterGroup) GET(relativePath string, handlers ...HandlerFunc) IRoutes
```

- `relativePath` 是相对当前分组的前缀；
- `handlers ...HandlerFunc` 是可变参数，最后一个通常是业务 handler，前面的是路由级中间件；
- 返回 `IRoutes`，所以可以链式调用。

### 17.3.2 路径参数

Gin 使用冒号语法：

```go
r.GET("/users/:id", func(c *gin.Context) {
    id := c.Param("id")
    c.String(200, "id=%s", id)
})
```

也支持通配：

```go
r.GET("/files/*filepath", func(c *gin.Context) {
    path := c.Param("filepath")
    c.String(200, "path=%s", path)
})
```

注意：Gin 的路径参数是 `:id`，不是 Go 1.22 `ServeMux` 的 `{id}`。

### 17.3.3 `Group`

```go
v1 := r.Group("/api/v1")
{
    v1.POST("/users", createUser)
    v1.GET("/users/:id", getUser)
}
```

签名：

```go
func (group *RouterGroup) Group(relativePath string, handlers ...HandlerFunc) *RouterGroup
```

`Group` 会创建一个新的 `RouterGroup`：

- `basePath` = 父分组路径 + `relativePath`；
- `Handlers` = 父分组 handlers + 本次传入的 handlers；
- 返回具体类型，方便继续 `Group` 或注册路由。

### 17.3.4 `Use`

```go
r.Use(gin.Logger(), gin.Recovery())
v1 := r.Group("/api/v1")
v1.Use(Auth())
```

签名：

```go
func (group *RouterGroup) Use(middleware ...HandlerFunc) IRoutes
```

`Use` 把中间件追加到当前分组的 `Handlers`。**它只影响之后注册的路由**，因为路由在注册时就把当时的 handlers 组合好了。这是最常见的「中间件不生效」原因之一。

### 17.3.5 注册时发生了什么

`RouterGroup.handle` 计算绝对路径，调用 `combineHandlers` 把「分组 Handlers + 路由 handlers」拼成一个链，再调用 `engine.addRoute` 插入路由树。最终链是「分组中间件在前，路由 handler 在后」，这也解释了为什么 `Use` 必须在路由注册前调用。

### 17.3.6 `Any` 与 `Handle`

```go
r.Any("/health", healthHandler)          // 注册所有常见方法
r.Handle("GET", "/custom", handler)      // 自定义方法
r.Match([]string{"GET", "POST"}, "/x", h) // 多个方法
```

源码中 `Any` 会遍历 `anyMethods`，对每个方法调用 `Handle`。

---

## 17.4 请求参数：Param、Query、Bind、Header

### 17.4.1 `c.Param`

```go
r.GET("/users/:id", func(c *gin.Context) {
    id := c.Param("id")
    c.String(200, "id=%s", id)
})
```

`c.Param` 从 `c.Params` 里按 key 查找，找不到返回空字符串。它是路径参数，不是查询参数。

### 17.4.2 `c.Query`、`DefaultQuery`、`GetQuery`

```go
q := c.Query("q")                   // 不存在或为空都返回 ""
page := c.DefaultQuery("page", "1") // 不存在返回 "1"
value, ok := c.GetQuery("q")        // 不存在返回 ("", false)
```

`Query` 分不清「没传」和「传了空值」；需要区分时用 `GetQuery`，参数存在但为空时返回 `("", true)`。

### 17.4.3 `c.GetHeader`

`token := c.GetHeader("Authorization")` 等价于 `c.Request.Header.Get(key)`，只读请求头，不区分大小写。

### 17.4.4 `ShouldBindJSON`

```go
type CreateUserInput struct {
    Name  string `json:"name" binding:"required"`
    Email string `json:"email" binding:"required,email"`
}

func createUser(c *gin.Context) {
    var in CreateUserInput
    if err := c.ShouldBindJSON(&in); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusCreated, in)
}
```

要点：

- 参数必须传指针；
- `json` 标签对应 JSON 字段；
- `binding:"required"` 做非空校验；
- 返回 `error`，**不会**自动写响应，也**不会** abort；
- 失败后要自己决定状态码并 `return`。

### 17.4.5 `ShouldBindQuery` 与 `ShouldBindUri`

```go
type ListQuery struct {
    Limit  int `form:"limit"`
    Offset int `form:"offset"`
}

var q ListQuery
if err := c.ShouldBindQuery(&q); err != nil { ... }
```

- `ShouldBindQuery` 用 `form` 标签，不是 `json`；
- `ShouldBindUri` 用 `uri` 标签，绑定 `:id` 等路径参数；
- `ShouldBind` 根据请求方法和 `Content-Type` 自动选择绑定器。

### 17.4.6 `BindJSON` 与 `ShouldBindJSON` 的关键区别

源码：

```go
func (c *Context) BindJSON(obj any) error {
    return c.MustBindWith(obj, binding.JSON)
}

func (c *Context) MustBindWith(obj any, b binding.Binding) error {
    if err := c.ShouldBindWith(obj, b); err != nil {
        c.AbortWithError(http.StatusBadRequest, err).SetType(ErrorTypeBind)
        return err
    }
    return nil
}
```

实测行为（Gin v1.10.0，请求体 `{}`，字段 `binding:"required"`）：

```go
// ShouldBindJSON：返回错误，不写响应
c.ShouldBindJSON(&in) // 之后自己 c.JSON(422, ...)
// 响应：422 {"error":"Key: 'Name' Error:Field validation for 'Name' failed on the 'required' tag"}
```

```go
// BindJSON：自动写 400 并 abort
c.BindJSON(&in)
// 响应：400，body 为空
```

结论：

| 方法 | 返回 | 失败时写响应 | 失败时 abort |
|------|------|--------------|--------------|
| `ShouldBindJSON` | `error` | 否 | 否 |
| `BindJSON` | `error` | 是（400） | 是 |

业务代码通常用 `ShouldBindJSON`，因为你可以返回统一错误格式。

### 17.4.7 `Content-Type`

- `ShouldBindJSON` 强制按 JSON 解析，不检查 `Content-Type`；
- `ShouldBind` 会根据 `Content-Type` 选择 JSON、XML、Form 等绑定器；
- 客户端发送 JSON 时应设置 `Content-Type: application/json`，否则 `ShouldBind` 可能走错分支。

---

## 17.5 响应：JSON、String、Data、HTML、Status

### 17.5.1 `c.JSON`

```go
c.JSON(http.StatusOK, gin.H{"message": "pong"})
c.JSON(http.StatusCreated, user)
```

- 设置 `Content-Type: application/json`；
- 写入状态码和 JSON body；
- 返回值是 `void`，调用即写响应。

### 17.5.2 `c.String`

```go
c.String(http.StatusOK, "hello %s", "gin")
```

按 `format` 和参数格式化，设置 `Content-Type: text/plain; charset=utf-8`，适合简单文本和健康检查。

### 17.5.3 `c.Data`

```go
c.Data(http.StatusOK, "application/octet-stream", []byte{0x89, 0x50, 0x4E, 0x47})
```

直接写原始字节，需要自己指定 `Content-Type`，适合图片、文件片段、protobuf。

### 17.5.4 `c.HTML`

```go
r.LoadHTMLGlob("templates/*")
c.HTML(http.StatusOK, "index.html", gin.H{"title": "Home"})
```

需要先加载模板；`name` 是模板文件名；设置 `Content-Type: text/html; charset=utf-8`。没有加载模板会 panic 或返回错误，取决于版本。

### 17.5.5 `c.Status`

```go
c.Status(http.StatusNoContent)
```

只写状态码，不写 body，常与 `c.Abort()` 搭配。如果后面又调用 `c.JSON`，会尝试写第二次状态码。

### 17.5.6 是否会自动写响应

| 方法 | 写响应 | 备注 |
|------|--------|------|
| `c.JSON` / `c.String` / `c.Data` / `c.HTML` | 是 | 状态码 + body |
| `c.Status` | 只写状态码 | 无 body |
| `c.Set/Get` / `c.Param/Query` | 否 | 只操作 context 或读请求 |
| `c.ShouldBindJSON` | 否 | 只解析 |
| `c.Abort` | 否 | 只改 index |

`c.JSON` 默认转义 HTML 字符；需要原样输出时用 `c.PureJSON`。普通 API 用 `c.JSON` 更安全。

---

## 17.6 中间件：Next、Abort、Set、Get

### 17.6.1 中间件签名

```go
func Auth() gin.HandlerFunc {
    return func(c *gin.Context) {
        // before
        c.Next()
        // after
    }
}
```

`gin.HandlerFunc` 就是 `func(*gin.Context)`。中间件通常是「返回 HandlerFunc 的函数」，这样可以接收配置参数。TS/Express 的 `(req, res, next)` 对应 `HandlerFunc(c)`，`next()` 对应 `c.Next()`，`res.json` 对应 `c.JSON`，`app.use` 对应 `r.Use`。

### 17.6.2 `c.Next` 的执行顺序

```go
r := gin.New()

r.Use(func(c *gin.Context) {
    fmt.Println("A before")
    c.Next()
    fmt.Println("A after")
})

r.Use(func(c *gin.Context) {
    fmt.Println("B before")
    c.Next()
    fmt.Println("B after")
})

r.GET("/", func(c *gin.Context) {
    fmt.Println("handler")
    c.String(200, "ok")
})
```

实测输出：

```
A before
B before
handler
B after
A after
```

`c.Next()` 的源码：

```go
func (c *Context) Next() {
    c.index++
    for c.index < int8(len(c.handlers)) {
        c.handlers[c.index](c)
        c.index++
    }
}
```

它把 `index` 往后移，依次执行剩余 handler。因此：

- `c.Next()` 之前的代码按注册顺序执行；
- 最终 handler 执行完后，`c.Next()` 返回；
- `c.Next()` 之后的代码按相反顺序执行。

### 17.6.3 不调用 `c.Next` 会怎样

很多教程说「中间件必须调用 `c.Next()` 才会继续」，这不准确。真正的循环在 `Engine.handleHTTPRequest` 里：它设置 `c.handlers` 后调用一次 `c.Next()`，中间件返回后外层循环会继续递增 `index`。所以不调用 `Next`，handler 仍会执行；但中间件里 `Next` 之后的收尾代码会提前执行。需要中断时用 `c.Abort()`，不是「不调用 Next」。

### 17.6.4 `c.Abort`

```go
func Auth() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.AbortWithStatusJSON(401, gin.H{"error": "missing token"})
            return
        }
        c.Set("user_id", int64(42))
        c.Next()
    }
}
```

`Abort` 源码：

```go
func (c *Context) Abort() {
    c.index = abortIndex
}
```

`abortIndex` 是 `math.MaxInt8 >> 1`，即 63。把 index 直接跳到链尾，后续 handler 不再执行。

实测：

```go
r.Use(func(c *gin.Context) {
    fmt.Println("auth before")
    c.AbortWithStatusJSON(401, gin.H{"error": "missing token"})
    fmt.Println("auth after abort")
})
r.GET("/", func(c *gin.Context) {
    fmt.Println("SHOULD NOT RUN")
})
```

实测输出 `auth before`、`auth after abort`，状态 401。注意：`Abort` **不会**中断当前函数；写完响应后要 `return`；`AbortWithStatusJSON` = `Abort()` + `JSON()`，不会自动 return。

### 17.6.5 `c.Set` 与 `c.Get`

```go
c.Set("user_id", int64(42))
value, ok := c.Get("user_id")
userID := value.(int64)
```

`Set` 写入 `c.Keys`，有读写锁；`Get` 返回 `(any, bool)`，必须类型断言。便捷方法有 `GetString`、`GetInt64`、`GetBool`、`MustGet`；`MustGet` 在 key 不存在时 panic。

### 17.6.6 常见中间件写法

计时：`start := time.Now(); c.Next(); log.Println(time.Since(start))`。

TraceID：

```go
func TraceID() gin.HandlerFunc {
    return func(c *gin.Context) {
        id := c.GetHeader("X-Request-ID")
        if id == "" {
            id = strconv.FormatInt(time.Now().UnixNano(), 10)
        }
        c.Set("trace_id", id)
        c.Header("X-Request-ID", id)
        c.Next()
    }
}
```

恢复：在 `defer` 中 `recover()`，然后 `c.AbortWithStatusJSON(500, ...)`。Gin 自带 `gin.Recovery()`，生产项目通常直接用它或在此基础上加日志。

### 17.6.7 中间件顺序与注册时机

```go
r := gin.New()
r.Use(Logger(), Recovery()) // 全局，注册路由前
v1 := r.Group("/api/v1")
v1.Use(Auth())              // 分组，注册该组路由前
v1.GET("/users", listUsers)
```

最终链：

```
Logger → Recovery → Auth → listUsers
```

规则：

- `Use` 只影响之后注册的路由；
- 分组中间件在全局中间件之后；
- Recovery 应尽量靠外层，才能捕获内层 panic；
- 日志中间件应在 Recovery 外层，才能记录 500。

### 17.6.8 goroutine 中要用 `c.Copy()`

`gin.Context` 会被放回对象池复用，handler 里启动 goroutine 直接使用 `c` 是危险的。正确做法是 `cCopy := c.Copy()`，goroutine 里只用 `cCopy`。注意 `Copy()` 不复制 `Request.Body`，异步任务也不应该依赖请求上下文。

---

## 17.7 与 `net/http` 的关系

### 17.7.1 `Engine` 实现 `http.Handler`

```go
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    c := engine.pool.Get().(*Context)
    c.writermem.reset(w)
    c.Request = req
    c.reset()

    engine.handleHTTPRequest(c)

    engine.pool.Put(c)
}
```

因为实现了 `ServeHTTP(http.ResponseWriter, *http.Request)`，`*gin.Engine` 就是 `http.Handler`：

```go
var _ http.Handler = (*gin.Engine)(nil)

http.ListenAndServe(":8080", r) // 直接传 r
```

`r.Run(":8080")` 内部也是 `http.ListenAndServe(address, engine.Handler())`。

### 17.7.2 `ServeHTTP` 做了什么

1. 从对象池取一个 `Context`；
2. 重置 `Writer`、`Request`、`Params`、`Keys`、`index`；
3. 调用 `handleHTTPRequest` 查路由树、设置 handlers、执行 `c.Next()`；
4. 把 `Context` 放回池里。

对象池意味着 `Context` 不能跨请求保存，也不能在 goroutine 里裸用。

### 17.7.3 `ResponseWriter` 被包装

Gin 的 `c.Writer` 不是原始 `http.ResponseWriter`，而是实现了更多方法的包装器：

```go
type ResponseWriter interface {
    http.ResponseWriter
    Status() int
    Size() int
    Written() bool
    WriteHeaderNow()
    // ...
}
```

所以中间件可以读取 `c.Writer.Status()` 记录状态码，`c.Writer.Size()` 记录响应大小。

### 17.7.4 `c.Request.Context`

```go
ctx := c.Request.Context()
db.WithContext(ctx).Find(&users)
```

`c.Request.Context()` 在客户端断开或请求结束时取消。数据库、下游 HTTP 调用都应该用它。

不要用 `context.Background()` 替代，否则请求取消了数据库还在跑。

### 17.7.5 混用 `net/http` 中间件

Gin 的 `Engine` 是 `http.Handler`，所以标准库中间件可以包在外面：

```go
handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // 标准库中间件
    r = r.WithContext(context.WithValue(r.Context(), "key", "value"))
    r.Handler.ServeHTTP(w, r)
})
```

更常见的做法是：

```go
srv := &http.Server{
    Addr:    ":8080",
    Handler: someNetHTTPMiddleware(r),
}
```

注意：标准库中间件拿不到 `*gin.Context`，只能操作 `http.ResponseWriter` 和 `*http.Request`。

### 17.7.6 优雅关闭

Gin 不提供 `Shutdown`。用 `http.Server` 包住 `Engine`：

```go
srv := &http.Server{Addr: ":8080", Handler: r}

go func() {
    if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
        errCh <- err
    }
}()

<-ctx.Done()

shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
_ = srv.Shutdown(shutdownCtx)
```

第 15 章的优雅关闭模式可以直接复用。

---

## 17.8 源码阅读路径

### 17.8.1 准备源码

```bash
go mod download github.com/gin-gonic/gin@v1.10.0
```

源码在 `$GOMODCACHE/github.com/gin-gonic/gin@v1.10.0/`。在 VS Code 里用 `Ctrl+P` 按文件名跳转，用 `Ctrl+Shift+F` 全局搜索函数名。

### 17.8.2 `gin.go`：引擎与请求入口

按顺序读：

| 位置（v1.10.0 约） | 内容 |
|---------------------|------|
| `gin.go:48` | `HandlerFunc` / `HandlersChain` |
| `gin.go:89` | `type Engine struct` |
| `gin.go:336` | `func (engine *Engine) addRoute` |
| `gin.go:583` | `func (engine *Engine) ServeHTTP` |
| `gin.go:605` | `func (engine *Engine) handleHTTPRequest` |
| `gin.go:389` | `func (engine *Engine) Run` |

阅读重点：

- `Engine` 嵌入 `RouterGroup`，所以路由方法被提升；
- `addRoute` 把 handler 链插入路由树；
- `handleHTTPRequest` 查树、设置 `c.handlers`、调用 `c.Next()`；
- `Run` 最终调用 `http.ListenAndServe`。

### 17.8.3 `routergroup.go`：分组与注册

| 位置（约） | 内容 |
|------------|------|
| `routergroup.go:27` | `type IRouter interface` |
| `routergroup.go:33` | `type IRoutes interface` |
| `routergroup.go:55` | `type RouterGroup struct` |
| `routergroup.go:65` | `func (group *RouterGroup) Use` |
| `routergroup.go:72` | `func (group *RouterGroup) Group` |
| `routergroup.go:86` | `func (group *RouterGroup) handle` |
| `routergroup.go:241` | `func (group *RouterGroup) combineHandlers` |

阅读重点：

- `Group` 复制父分组 handlers；
- `handle` 组合 handlers 并调用 `engine.addRoute`；
- `returnObj` 区分根分组和普通分组。

### 17.8.4 `context.go`：请求上下文

| 位置（约） | 内容 |
|------------|------|
| `context.go:55` | `type Context struct` |
| `context.go:182` | `Next`（index 循环） |
| `context.go:199` | `Abort`（index = 63） |
| `context.go:260` | `Set` / `Get`（带锁的 Keys） |
| `context.go:404` | `Param` / `Query`（读请求） |
| `context.go:711` | `ShouldBindJSON` |
| `context.go:1004` | `JSON` / `String` / `Data`（走 Render） |

阅读重点：`reset` 清空哪些字段决定对象池安全；`Next` 控制链；`Abort` 跳到链尾；`Set/Get` 有锁；`Param/Query` 只是读请求；渲染方法最终都走 `c.Render`。

### 17.8.5 `tree.go`、`middleware/`、`binding/`、`render/`

- `tree.go`：基数树，静态路径、`:param`、`*catchAll` 混合存储，`getValue` 查 handler；
- `middleware/`：`logger.go`、`recovery.go`、`gzip.go`、`basic_auth.go`；Gin 没有内置 CORS；
- `binding/`：`json.go`、`form.go`、`validator.go`；
- `render/`：`json.go`、`string.go`、`html.go`。

这些目录解释了绑定和渲染背后做了什么。日常业务知道「注册时插入、请求时查找」即可，不需要读完整路由树实现。

### 17.8.6 一次请求的完整路径

以 `GET /api/v1/users/42` 为例：

1. `http.ListenAndServe` 调用 `Engine.ServeHTTP`；
2. 从对象池取 `Context` 并重置；
3. `handleHTTPRequest` 查到 `/api/v1/users/:id`，设置 `Params` 和 `handlers`；
4. `c.Next()` 依次执行中间件和 `getUser`；
5. `getUser` 调用 `c.Param("id")`、`c.JSON`，`c.JSON` 走 `c.Render` 写响应；
6. `Context` 放回对象池。

### 17.8.7 阅读源码的技巧

先看类型和接口，再看 `ServeHTTP`/`Run`/`handle` 入口，跟踪一条请求路径，先跳过 redirect、静态文件、HTTP/2 等分支，用 `_test.go` 当用法示例。pkg.go.dev 看签名，源码看实现。

| 技巧 | 做法 |
|------|------|
| 先看类型 | `type X struct` 和接口定义 |
| 再看入口 | `ServeHTTP`、`Run`、`handle` |
| 跟踪一条路径 | 从 `ServeHTTP` 到 handler |
| 忽略次要分支 | 先跳过 redirect、静态文件、HTTP/2 |
| 用测试当示例 | `_test.go` 里有大量用法 |
| 对照文档 | pkg.go.dev 看签名，源码看实现 |

---

## 17.9 签名判断法：返回值、写响应、abort

### 17.9.1 三个问题

看到任意 Gin 方法，先问：

1. 它返回什么？`void`、`error`、`*Context`、`IRoutes`？
2. 它会不会写响应？
3. 它会不会 abort 后续 handler？

### 17.9.2 常见方法速查

| 方法 | 返回 | 写响应 | abort |
|------|------|--------|-------|
| `c.JSON` | 无 | 是 | 否 |
| `c.String` | 无 | 是 | 否 |
| `c.Data` | 无 | 是 | 否 |
| `c.HTML` | 无 | 是 | 否 |
| `c.Status` | 无 | 只写状态码 | 否 |
| `c.ShouldBindJSON` | `error` | 否 | 否 |
| `c.BindJSON` | `error` | 是（400） | 是 |
| `c.Next` | 无 | 否 | 否 |
| `c.Abort` | 无 | 否 | 是 |
| `c.AbortWithStatus` | 无 | 是 | 是 |
| `c.AbortWithStatusJSON` | 无 | 是 | 是 |
| `c.Set` | 无 | 否 | 否 |
| `c.Get` | `(any, bool)` | 否 | 否 |
| `c.Param` | `string` | 否 | 否 |
| `c.Query` | `string` | 否 | 否 |
| `Group` | `*RouterGroup` | 否 | 否 |
| `Use` | `IRoutes` | 否 | 否 |
| `GET` | `IRoutes` | 否 | 否 |
| `Engine.Run` | `error` | 否 | 否 |

### 17.9.3 文档符号

| 符号 | 含义 |
|------|------|
| `*gin.Context` | 指向 Context 的指针，handler 参数 |
| `gin.H` | `map[string]any`，JSON 快捷写法 |
| `...HandlerFunc` | 可变参数，可传 0 到多个 handler |
| `any` | `interface{}` 的别名 |
| `IRoutes` | 路由注册接口，支持链式调用 |
| `IRouter` | 在 `IRoutes` 基础上增加 `Group` |

### 17.9.4 看到 `...HandlerFunc` 怎么读

```go
func (group *RouterGroup) GET(relativePath string, handlers ...HandlerFunc) IRoutes
```

第一个参数是路径，后面可传任意多个 handler；最后一个通常是业务 handler，前面的作为路由级中间件；返回 `IRoutes` 支持链式调用。`r.GET("/x", Auth(), handler)` 表示 Auth 在该路由上先执行。

---

## 17.10 把方法迁移到 GORM

同样的阅读方法适用于 GORM：

1. **类型地图**：`gorm.DB`、`gorm.Statement`、`gorm.Session`、`gorm.Config`；
2. **入口**：`gorm.Open`、`db.WithContext`、`db.Create/Find/First`；
3. **链式关系**：`Where`、`Order`、`Limit` 返回 `*gorm.DB`，`First/Find` 是 finisher，真正执行 SQL；
4. **状态**：`db.Error`、`db.RowsAffected`；
5. **源码路径**：`gorm.go`、`chainable_api.go`、`finisher_api.go`、`statement.go`、`callbacks/`。

```go
db.WithContext(ctx).Where("email = ?", email).First(&user)
```

`WithContext` 和 `Where` 都返回 `*gorm.DB`；`First` 执行查询；错误在 `db.Error`，记录不存在是 `gorm.ErrRecordNotFound`。第 18 章会展开。

---

## 17.11 常见错误与陷阱

| 错误/现象 | 原因 | 修复 |
|-----------|------|------|
| `superfluous response.WriteHeader` | `c.JSON` 后又写一次 | 写完 `return` |
| 中间件不生效 | `Use` 在路由注册之后 | 先 `Use` 再注册 |
| `Abort` 后仍执行代码 | `Abort` 不中断当前函数 | `Abort` 后 `return` |
| `c.Param` 为空 | 路由写成 `{id}` | Gin 用 `:id` |
| 绑定失败但返回 200 | 忽略 `ShouldBindJSON` 错误 | 检查 err 并返回 400 |
| `form` 标签不生效 | 用了 `ShouldBindJSON` | 查询参数用 `form` + `ShouldBindQuery` |
| 异步 goroutine 拿到旧值 | `Context` 被复用 | 用 `c.Copy()` |
| 子 goroutine panic 导致进程退出 | Recovery 只捕获同 goroutine | 子 goroutine 自己 recover |
| 代理 IP 警告 | 默认信任所有代理 | 配置 `SetTrustedProxies` |

补充三点：

1. `c.Next()` 不是「继续链」的开关，外层引擎循环才是；它的作用是「在这里执行剩余链，再回来做收尾」；
2. `c.Abort()` 只改 index，不写响应、不 return；
3. `Engine` 是 `http.Handler`，所以标准库的 `http.Server`、超时、优雅关闭都能直接用。

---

## 17.12 练习题

### 基础题

1. 用自己的话解释 `gin.HandlerFunc`、`gin.HandlersChain`，并画出核心类型关系图。
2. 写一个 `GET /users/:id`，用 `c.Param` 返回 ID。
3. 用 `c.Query` 和 `c.DefaultQuery` 实现分页参数。
4. 用 `ShouldBindJSON` 接收 `{name,email}`，失败返回 400。
5. 用 `c.JSON`、`c.String`、`c.Data` 各写一个接口。
6. 写一个记录耗时的中间件，打印 `c.Writer.Status()`。

### 进阶题

7. 写一个 Auth 中间件，缺失 token 时 `AbortWithStatusJSON(401)`。
8. 写一个 TraceID 中间件，用 `c.Set` 保存，handler 用 `c.GetString` 读取。
9. 验证 A/B 两个中间件的执行顺序是 A before → B before → handler → B after → A after。
10. 验证「中间件不调用 `c.Next()`，handler 仍会执行」，并解释原因。
11. 验证 `ShouldBindJSON` 失败不会自动写 400，而 `BindJSON` 会。
12. 用 `c.Copy()` 在 goroutine 中安全读取 context 值。
13. 对比 `group.Use` 前后注册的路由，确认中间件只影响之后的路由。
14. 把 `Engine` 传给 `http.Server`，实现优雅关闭，并自定义 `NoRoute`。
15. 用 `r.Static("/assets", "./assets")` 提供静态文件。

### 读文档题

16. 执行 `go doc github.com/gin-gonic/gin.Context.ShouldBindJSON`，回答参数和返回值。
17. 执行 `go doc github.com/gin-gonic/gin.RouterGroup`，回答 `Use` 和 `Group` 的签名。
18. 阅读 `gin.go` 的 `ServeHTTP`，回答 `Context` 从哪里来、到哪里去。
19. 阅读 `context.go` 的 `Next` 和 `Abort`，解释 `index` 的作用。
20. 阅读 `routergroup.go` 的 `combineHandlers`，解释最终 handler 链的顺序。
21. 查阅 pkg.go.dev，回答 `IRoutes` 和 `IRouter` 的区别。

---

## 17.13 验收清单

- [ ] 能画出 Gin 核心类型关系图，解释 `HandlerFunc`、`HandlersChain`、`IRoutes`、`IRouter`；
- [ ] 能解释 `GET/Group/Use` 的注册过程；
- [ ] 能区分 `Param`、`Query`、`ShouldBind*`、`GetHeader`；
- [ ] 能判断 `JSON/String/Data/HTML/Status` 是否写响应；
- [ ] 能解释 `Next/Abort/Set/Get` 的执行顺序；
- [ ] 能说清 `Engine` 为什么能作为 `http.Handler`；
- [ ] 能定位至少 5 个核心类型/方法的源码位置；
- [ ] 能不依赖教程读懂一段 Gin 文档示例，并把方法迁移到 GORM。

---

## 17.14 本章小结

| 主题 | 一句话解释 | 易错点 |
|------|------------|--------|
| `Engine` / `RouterGroup` | 引擎嵌入分组，持有路由树 | `Use` 只影响之后的路由 |
| `Context` | 一次请求的状态与操作入口 | 会被对象池复用 |
| `HandlerFunc` / `HandlersChain` | `func(*Context)` 及其切片 | 没有 error 返回值 |
| `IRoutes` / `IRouter` | 路由注册接口 | 返回它支持链式调用 |
| `c.Next` / `c.Abort` | 执行剩余链 / 跳过后续 handler | Abort 不写响应、不 return |
| `ShouldBindJSON` / `BindJSON` | 只解析 / 失败自动 400 + abort | 统一错误格式用前者 |
| `Engine` 与 `net/http` | 实现 `ServeHTTP` | 可直接传给 `http.Server` |
| 源码入口 | `gin.go`、`routergroup.go`、`context.go` | 先类型后入口 |

---

## 17.15 下一章预告

第 18 章：**如何读懂 GORM 文档与源码**。你会用同一套方法建立 GORM 类型地图：`gorm.DB`、`Statement`、`Session`、`Config`；理解链式方法为什么返回 `*gorm.DB`、`Error` 字段如何传递、`Where/Find` 何时真正执行 SQL；读懂模型标签、事务、钩子、`Valuer/Scanner`；并定位 `gorm.go`、`chainable_api.go`、`finisher_api.go`、`statement.go` 和 `callbacks/` 的关键实现。

本章的「类型地图 → 签名 → 链式关系 → 源码路径」方法，会在 GORM 上再完整用一遍。
