# 附录 C：文档阅读索引

> **用途**：知道去哪查、怎么查、先读什么；把「看文档」变成可重复的流程。
> **版本基准**：Go 1.22+、Gin v1.9+、GORM v2。
> **配套章节**：第 01、09、12、13、17、18 章。

---

## C.1 读文档的固定流程

1. **确认版本**：先看文档对应的版本号，不要拿 v1.11 的行为套 v1.9。
2. **找索引**：在 pkg.go.dev 的 Index 或 `go doc` 输出里定位类型/函数。
3. **读签名**：参数类型、返回值、指针、接口、可变参数各是什么。
4. **读说明与示例**：注意「是否写响应」「是否返回新对象」「失败时是否 panic」。
5. **跳源码确认**：签名和注释说不清的行为，直接看实现与测试。

读签名的顺序：**函数名 → 参数 → 返回值 → 接收者 → 文档第一段 → Example**。

---

## C.2 `go doc` 用法

```bash
go doc fmt                         # 包概览
go doc fmt.Println                 # 单个函数
go doc net/http.Handler            # 类型
go doc net/http.Request.Clone      # 方法
go doc -all strings                # 包的全部导出 API
go doc -src errors.Is              # 直接看实现源码
go doc -u github.com/gin-gonic/gin.Context  # 包含未导出成员
go doc github.com/gin-gonic/gin.Context.ShouldBindJSON
```

第三方库需要当前目录在某个模块内（有 `go.mod`），否则先 `go get`。

| 参数 | 作用 |
|------|------|
| `-all` | 显示包的全部文档 |
| `-src` | 显示源码而不是摘要 |
| `-u` | 包含未导出标识符 |
| `-short` | 只显示一行摘要 |
| `-cmd` | 显示命令（main 包）文档 |

配套命令：

```bash
go list -m -versions github.com/gin-gonic/gin   # 查可用版本
go list -m all                                  # 当前依赖树
go env GOMODCACHE                               # 模块缓存位置
go mod download -json github.com/gin-gonic/gin  # 模块下载信息
```

---

## C.3 pkg.go.dev 页面怎么读

| 页面区域 | 内容 | 阅读重点 |
|----------|------|----------|
| 顶部 | 导入路径、版本、发布时间 | 确认版本；用 `@v1.9.1` 固定版本 |
| Documentation / Index | 常量、变量、函数、类型索引 | 先扫一遍有哪些类型 |
| Overview | 包级说明 | 包的设计意图、使用前提 |
| Constants / Variables | 哨兵错误、默认值 | `ErrXxx` 常用于 `errors.Is` |
| Functions | 包级函数 | 参数与返回值 |
| Types | 类型及其方法 | 方法的接收者、返回值 |
| Examples | 可运行示例 | 最接近真实用法 |
| Source Files | 源文件列表 | 跳转到源码 |
| Directories | 子包 | 驱动、中间件等拆分 |

固定版本 URL 示例：

```
https://pkg.go.dev/github.com/gin-gonic/gin@v1.9.1
https://pkg.go.dev/gorm.io/gorm@v1.25.12
```

页面上的 `[查看源码]` 通常跳到 `cs.opensource.google` 或 GitHub；本地读源码更快，见 C.14。

---

## C.4 常用标准库索引

| 包 | 用途 | 必看 API | 章节 |
|----|------|----------|------|
| `fmt` | 格式化输出 | `Printf`、`Sprintf`、`Errorf`、`%v/%+v/%#v/%T/%w` | 03、08、12 |
| `strings` | 字符串处理 | `Contains`、`Split`、`Join`、`TrimSpace`、`Builder` | 04、12 |
| `strconv` | 字符串与数字转换 | `Atoi`、`Itoa`、`ParseInt`、`ParseFloat` | 02、12 |
| `bytes` | 字节切片 | `Buffer`、`Equal`、`Contains` | 04、12 |
| `unicode/utf8` | UTF-8 | `RuneCountInString`、`DecodeRuneInString` | 04 |
| `sort` / `slices` | 排序 | `sort.Slice`、`slices.Sort`、`SortFunc`、`Clone` | 10、12 |
| `maps` | map 工具 | `Clone`、`Equal`、`DeleteFunc` | 10、12 |
| `time` | 时间 | `Now`、`Parse`、`Format`、`Duration`、`Ticker` | 12 |
| `errors` | 错误 | `New`、`Is`、`As`、`Join` | 08 |
| `context` | 取消与超时 | `WithCancel`、`WithTimeout`、`WithValue` | 11、13 |
| `sync` | 同步 | `Mutex`、`RWMutex`、`WaitGroup`、`Once` | 11 |
| `sync/atomic` | 原子操作 | `AddInt64`、`LoadInt64`、`CompareAndSwapInt64` | 11 |
| `io` | 流式读写 | `Reader`、`Writer`、`Copy`、`ReadAll`、`EOF` | 13 |
| `bufio` | 缓冲 IO | `Reader`、`Writer`、`Scanner` | 13 |
| `os` | 文件与环境 | `ReadFile`、`WriteFile`、`Getenv`、`Args` | 13 |
| `path/filepath` | 路径 | `Join`、`Base`、`Dir`、`Ext`、`Glob` | 13 |
| `encoding/json` | JSON | `Marshal`、`Unmarshal`、`Decoder`、`RawMessage` | 13 |
| `net/http` | HTTP | `Handler`、`HandlerFunc`、`Server`、`Client`、`Shutdown` | 13、16 |
| `net/http/httptest` | HTTP 测试 | `NewRecorder`、`NewServer` | 14 |
| `database/sql` | 数据库 | `Open`、`Query`、`QueryRow`、`Exec`、`Scan`、连接池方法 | 13 |
| `log/slog` | 结构化日志 | `Info`、`Error`、`With`、`Handler` | 12、15 |
| `flag` | 命令行参数 | `String`、`Int`、`Parse` | 15 |
| `os/signal` | 信号 | `NotifyContext` | 15 |
| `reflect` | 反射 | `TypeOf`、`ValueOf`、`Kind`、`Tag.Get` | 05 |
| `testing` | 测试 | `T`、`B`、`Run`、`Errorf`、`Fatalf`、`TempDir` | 14 |

> 版本提示：`maps.Keys` / `maps.Values` 在 Go 1.23 起返回 `iter.Seq`，Go 1.22 的 `maps` 包没有这两个函数；1.22 用 `for k := range m` 或自行收集。`slices.Collect` 也是 1.23 起才有。

---

## C.5 Gin 文档阅读路线

1. **README Quickstart**：先跑通最小示例，认识 `gin.Default()`、`r.GET`、`c.JSON`。
2. **`docs/doc.md`**（模块内）或 GitHub 仓库 `docs/`：按主题查路由、参数、绑定、中间件。
3. **pkg.go.dev**：查准确签名，重点看 `Engine`、`RouterGroup`、`Context`。
4. **Examples**：Gin 仓库 `examples/` 下有大量可运行项目。
5. **源码**：最后读 `gin.go` → `routergroup.go` → `context.go`。

阅读时始终问三个问题：

- 这个方法返回什么？`*Context`、`IRoutes`、`*RouterGroup` 还是 `error`？
- 它会不会自动写响应/状态码？
- 它会不会中止后续 handler（`Abort`）？

---

## C.6 Gin 核心类型地图

| 类型 | 定义 | 源文件 | 阅读要点 |
|------|------|--------|----------|
| `Engine` | `struct` | `gin.go` | 路由树、配置、`Run`、`ServeHTTP`；嵌入 `RouterGroup` |
| `RouterGroup` | `struct` | `routergroup.go` | `Handlers`、`basePath`、`engine`；`GET/POST/Group/Use` |
| `Context` | `struct` | `context.go` | `Request`、`Writer`、`Params`、`handlers`、`index`、`Keys` |
| `HandlerFunc` | `func(*Context)` | `gin.go` | Gin 的核心函数类型；中间件也是它 |
| `HandlersChain` | `[]HandlerFunc` | `gin.go` | 一条请求的处理器链；`Last()` 取最后一个 |
| `IRoutes` | `interface` | `routergroup.go` | 路由注册方法的集合 |
| `IRouter` | `interface` | `routergroup.go` | `IRoutes` + `Group` |
| `H` | `map[string]any` | `utils.go` | 构造 JSON 的快捷方式 |
| `ResponseWriter` | `interface` | `response_writer.go` | 包装 `http.ResponseWriter`，记录状态码与响应大小 |
| `Binding` | `interface` | `binding/binding.go` | 请求绑定策略（JSON、form、query、uri） |
| `Render` | `interface` | `render/render.go` | 响应渲染策略（JSON、HTML、text、data） |
| `Error` | `struct` | `errors.go` | `c.Error(err)` 收集的错误项 |

类型关系（简图）：

```
http.Request ──┐
               ├──> gin.Context ──> HandlersChain ──> HandlerFunc
http.ResponseWriter ─┘                     │
                                           └── RouterGroup/Engine 注册
Engine ──嵌入──> RouterGroup
Engine ──实现──> http.Handler
```

---

## C.7 一次 Gin 请求的调用链

1. `net/http.Server` 接收连接，每个请求调用 `Engine.ServeHTTP`。
2. `ServeHTTP` 从 `sync.Pool` 取 `Context`，重置 `Request` / `Writer`。
3. `handleHTTPRequest` 按 method + path 在路由树中匹配，写入 `c.Params`、`c.handlers`、`c.fullPath`。
4. 调用 `c.Next()`：`index++`，依次执行 `handlers`。
5. 中间件里的 `c.Next()` 会继续向内执行；`c.Abort()` 把 `index` 设为 `abortIndex`，后续 handler 不再执行。
6. handler 通过 `c.JSON` 等写响应；`Context` 放回池中复用。

关键结论：

- `c.Next()` 不是「跳到下一个中间件」那么简单，它会在当前函数内递归式地把剩余链跑完；
- `Abort()` 不会中断当前函数，只影响后续 handler；
- `Engine` 实现 `http.Handler`，所以能直接交给 `http.Server`。

---

## C.8 Gin 源码文件索引

| 文件 | 内容 |
|------|------|
| `gin.go` | `Engine`、`HandlerFunc`、`HandlersChain`、`New`、`Default`、`Run`、`ServeHTTP` |
| `routergroup.go` | `RouterGroup`、`IRoutes`、`IRouter`、`GET/POST/Group/Use/Static` |
| `context.go` | `Context`、参数读取、绑定、响应、`Next`、`Abort`、`Set/Get` |
| `tree.go` | 路由树（radix tree）与路径匹配；不必通读，查参数匹配时看 |
| `response_writer.go` | `ResponseWriter`、`responseWriter`、状态码与大小记录 |
| `recovery.go` | `Recovery` 中间件：`defer` + `recover` |
| `logger.go` | `Logger` 中间件与日志格式 |
| `mode.go` | `DebugMode` / `ReleaseMode` / `TestMode` |
| `auth.go` | Basic Auth 中间件 |
| `errors.go` | `Error`、`ErrorType`、错误收集 |
| `fs.go` | `Static` / `StaticFS` 的文件系统实现 |
| `binding/` | `binding.go`、`json.go`、`form.go`、`query.go`、`uri.go`、`default_validator.go` |
| `render/` | `render.go`、`json.go`、`html.go`、`text.go`、`data.go`、`redirect.go` |
| `internal/bytesconv/` | 字符串与字节转换的内部工具 |

> v1.9 与 v1.12 的路由树内部结构有差异（如 `trees` 与 `routeTrees`），但公开 API 与文件职责稳定；读行为时以你锁定的版本为准。

---

## C.9 GORM 文档阅读路线

官方文档站：`https://gorm.io/docs/`。建议顺序：

1. **Models**：结构体、`gorm.Model`、字段标签、主键、表名。
2. **Connecting to Database**：DSN、驱动、连接池、日志。
3. **Create / Query / Update / Delete**：CRUD 与零值、软删除。
4. **Advanced Query**：`Where`、`Joins`、`Group`、`Having`、子查询。
5. **Associations / Preload**：关系与 N+1。
6. **Transactions / Hooks**：事务回调、生命周期钩子。
7. **Migrations**：`AutoMigrate` 与 Migrator。
8. **Context / Error Handling / Logger / Performance**：工程化收尾。

读文档时重点确认：

- 方法是**链式方法**（返回 `*gorm.DB`）还是**终止方法**（执行 SQL）；
- 结果从 `result.Error`、`result.RowsAffected` 取；
- 传的是结构体、指针还是切片指针；
- 零值、软删除、关联是否会改变生成的 SQL。

---

## C.10 GORM 核心类型地图

| 类型 | 定义 | 源文件 | 阅读要点 |
|------|------|--------|----------|
| `gorm.DB` | `struct` | `gorm.go` | 嵌入 `*Config`；`Error`、`RowsAffected`、`Statement`、`clone` |
| `gorm.Statement` | `struct` | `statement.go` | 一次操作的上下文：`Table`、`Model`、`Dest`、`Clauses`、`Vars`、`SQL`、`Schema`、`Context` |
| `gorm.Session` | `struct` | `gorm.go` | 会话级配置：`DryRun`、`NewDB`、`SkipHooks`、`Context` 等 |
| `gorm.Config` | `struct` | `gorm.go` | 全局配置：命名策略、日志、`PrepareStmt`、`NowFunc` |
| `gorm.Model` | `struct` | `model.go` | `ID`、`CreatedAt`、`UpdatedAt`、`DeletedAt` |
| `gorm.Dialector` | `interface` | `interfaces.go` | 数据库方言（MySQL、SQLite、Postgres） |
| `gorm.Plugin` | `interface` | `interfaces.go` | 插件生命周期 |
| `gorm.Valuer` | `interface` | `interfaces.go` | 自定义值写入数据库 |
| `sql.Scanner` | 标准库接口 | `database/sql` | 从数据库读取到自定义类型 |
| `clause.Expression` | `interface` | `clause/clause.go` | SQL 片段构建 |
| `schema.Schema` | `struct` | `schema/schema.go` | 模型元数据：表、字段、关系、索引 |
| `schema.Field` | `struct` | `schema/field.go` | 字段元数据：列名、类型、标签 |
| `logger.Interface` | `interface` | `logger/logger.go` | 日志接口 |
| `gorm.Migrator` | `interface` | `migrator.go` | 迁移 API |

关系（简图）：

```
DB ──Config
 │
 ├── Statement ──Schema ──Field / Relationship
 │      │
 │      ├── Clauses (WHERE / SELECT / ORDER ...)
 │      └── Vars / SQL
 │
 └── Callbacks (create / query / update / delete / preload ...)
```

---

## C.11 GORM 链式调用怎么读

```go
var users []User
err := db.WithContext(ctx).
    Where("age > ?", 18).
    Order("created_at DESC").
    Limit(10).
    Find(&users).Error
```

执行过程（概念上）：

1. `WithContext`、`Where`、`Order`、`Limit` 都是链式方法，返回新的/克隆的 `*gorm.DB`，把条件写入 `Statement.Clauses` 或 `Statement` 字段。
2. `Find(&users)` 是终止方法，触发 `query` 回调：根据 `Statement` 生成 SQL、执行、把结果扫描进 `users`。
3. 错误写到返回的 `tx.Error`，影响行数写到 `tx.RowsAffected`。

阅读要点：

- 链式方法通常返回 `tx`，**不要忽略返回值**；`db.Where(...)` 不修改原 `db` 的条件（底层通过 clone 实现）。
- 需要复用「干净」的会话时用 `Session(&gorm.Session{NewDB: true})`。
- `Find` 到切片时查不到是空切片 + `nil` 错误；`First` 查不到是 `gorm.ErrRecordNotFound`。

---

## C.12 GORM 源码文件索引

| 文件 | 内容 |
|------|------|
| `gorm.go` | `DB`、`Config`、`Session`、`Open`、`WithContext`、连接初始化 |
| `chainable_api.go` | `Where`、`Select`、`Order`、`Limit`、`Offset`、`Preload`、`Joins`、`Model`、`Table`、`Scopes` |
| `finisher_api.go` | `Create`、`First`、`Take`、`Last`、`Find`、`Save`、`Updates`、`Delete`、`Transaction` |
| `statement.go` | `Statement`、SQL 构建、`AddClause`、`SetColumn`、`Parse` |
| `callbacks.go` | 回调注册与执行入口 |
| `callbacks/create.go` | 创建流程与自增 ID 回填 |
| `callbacks/query.go` | 查询流程 |
| `callbacks/update.go` | 更新流程与零值处理 |
| `callbacks/delete.go` | 删除流程与软删除 |
| `callbacks/transaction.go` | 事务回调 |
| `callbacks/preload.go` | 预加载 |
| `callbacks/associations.go` | 关联保存 |
| `schema/schema.go` | 模型解析入口 |
| `schema/field.go` | 字段解析、标签处理 |
| `schema/relationship.go` | 关系解析 |
| `schema/naming.go` | 表名/列名命名策略 |
| `schema/serializer.go` | 序列化字段 |
| `clause/clause.go` | `Clause`、`Expression`、`Builder` |
| `clause/where.go`、`select.go`、`insert.go`、`update.go`、`delete.go` | 各类 SQL 片段 |
| `clause/joins.go`、`order_by.go`、`limit.go`、`group_by.go` | JOIN、排序、分页、分组 |
| `model.go` | `gorm.Model` |
| `soft_delete.go` | `DeletedAt` 与软删除 |
| `errors.go` | `ErrRecordNotFound` 等哨兵错误 |
| `interfaces.go` | `Dialector`、`Plugin`、`Valuer` 等 |
| `association.go` | 关联操作 |
| `migrator.go`、`migrator/` | 迁移 |
| `logger/` | 日志接口与默认实现 |
| `prepare_stmt.go` | 预编译语句缓存 |
| `scan.go` | 行扫描 |

---

## C.13 推荐阅读顺序

**第一轮：标准库（第 12–13 章后）**

`fmt` → `strings` / `strconv` → `time` → `encoding/json` → `net/http` → `context` → `database/sql`。

**第二轮：Gin（第 17 章）**

1. README 最小示例；
2. `routergroup.go`：路由怎么注册、`Group` / `Use`；
3. `gin.go`：`Engine` 结构与 `ServeHTTP`；
4. `context.go`：`Next` / `Abort`、绑定、响应；
5. `response_writer.go`：状态码与响应包装；
6. `recovery.go` / `logger.go`：中间件写法；
7. `binding/` / `render/`：按需查。

**第三轮：GORM（第 18 章）**

1. 文档 Models / CRUD；
2. `gorm.go`：`DB` / `Config` / `Session`；
3. `chainable_api.go` / `finisher_api.go`：链式与终止方法；
4. `statement.go`：一次操作的状态；
5. `callbacks/`：SQL 何时执行；
6. `schema/`：标签如何变成列；
7. `clause/`：SQL 如何拼出来。

**第四轮：带着问题读**

不要从第一行读到最后一行。先提一个具体问题（例如「`Updates` 为什么忽略零值？」），再沿调用链定位，读完写三句话总结。

---

## C.14 源码阅读工具与技巧

| 工具/技巧 | 用法 |
|-----------|------|
| `go doc -src` | 直接看函数实现，不必打开仓库 |
| `go env GOMODCACHE` | 依赖源码在 `$GOMODCACHE/github.com/gin-gonic/gin@vX.Y.Z` |
| VS Code | `Ctrl+点击` / `F12` 跳转定义；依赖源码是只读的 |
| `grep` / `rg` | 在模块目录内搜符号：`rg "func \(c \*Context\) Next"` |
| `go list -f '{{.Dir}}' <pkg>` | 直接打印某个包的本地源码目录 |
| 测试文件 | `*_test.go` 是最好的用法示例，常比文档更完整 |
| `git log -p <file>` | 看某段实现为何改成现在这样（需克隆仓库） |
| `go doc -all <pkg>` | 快速建立包的 API 全貌 |

**读源码的三个问题**：

1. 输入从哪来（参数、`Statement`、`Context`）？
2. 状态存在哪（字段、`Clauses`、`Keys`）？
3. 输出到哪（返回值、`Error`、`ResponseWriter`）？

回答完这三个问题，一个函数就读懂了。
