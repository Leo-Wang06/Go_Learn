# 附录 E：术语表

> **用途**：中英对照 + 一句话解释，读文档/源码卡住时快速查。
> **排序**：先按英文术语字母序；最后单列 Gin/GORM 术语与易混术语。
> **配套章节**：全书。

---

## E.1 A–C

| 术语 | 中文 | 一句话解释 |
|------|------|------------|
| addressable | 可寻址 | 能取地址的表达式（变量、指针解引用、切片元素等）；map 元素不可寻址 |
| alias | 类型别名 | `type A = B`，A 与 B 完全等价 |
| anonymous field | 匿名字段 | 结构体中只写类型不写字段名的字段，用于嵌入 |
| any | 任意类型 | `interface{}` 的别名，Go 1.18+ |
| array | 数组 | 长度固定、值语义的同类元素序列，类型为 `[N]T` |
| assertion | 类型断言 | 从接口值提取具体类型的运行时操作，写作 `x.(T)` |
| atomic | 原子操作 | `sync/atomic` 提供的不可分割读写/修改，避免加锁 |
| benchmark | 基准测试 | `BenchmarkXxx(b *testing.B)`，测量函数性能与内存分配 |
| blank identifier | 空白标识符 | `_`，表示丢弃该值 |
| blocking | 阻塞 | goroutine 因 channel、锁或 IO 无法继续而等待 |
| buffer | 缓冲区 | 有缓冲 channel 中暂存元素的队列 |
| byte | 字节 | `uint8` 的别名；字符串索引得到的就是 byte |
| channel | 通道 | goroutine 之间传递值的类型化管道 |
| closure | 闭包 | 捕获外部变量的函数值 |
| comma-ok | 逗号-ok 惯用法 | `v, ok := m[k]` / `x.(T)` 的双返回值写法 |
| comparable | 可比较 | 可用 `==` / `!=` 比较、可作 map key 的类型约束 |
| composite literal | 复合字面量 | `T{...}` 形式的构造表达式 |
| concurrency | 并发 | 多个任务在重叠时间段内推进，不等于并行 |
| constraint | 约束 | 泛型类型参数允许的类型集合 |
| context | 上下文 | 传递取消信号、超时和请求范围值的接口 |
| copy | 复制 | 内建函数，把元素从源切片复制到目标切片，返回复制数量 |

---

## E.2 D–F

| 术语 | 中文 | 一句话解释 |
|------|------|------------|
| deadlock | 死锁 | 所有 goroutine 互相等待，程序无法继续推进 |
| defer | 延迟调用 | 注册在函数返回前执行的调用，后进先出 |
| defined type | 定义类型（新类型） | `type A B` 创建的新类型，与 B 不同但底层类型相同 |
| dependency injection | 依赖注入 | 由外部把依赖传给组件，而不是组件内部自行创建 |
| dereference | 解引用 | 通过 `*p` 访问指针指向的值 |
| dot import | 点导入 | `import . "pkg"`，把导出标识符直接放入当前命名空间，不推荐 |
| driver.Valuer | 数据库值接口 | 定义 `Value() (driver.Value, error)`，把自定义类型写入数据库 |
| embedding | 嵌入 | 把类型作为匿名字段放入结构体，获得字段与方法提升 |
| error | 错误 | 内建接口 `Error() string`；Go 用返回值传递可预期失败 |
| error wrapping | 错误包装 | 用 `%w` 把底层错误放进新错误链 |
| escape analysis | 逃逸分析 | 编译器判断变量能否留在栈上的分析 |
| exported | 导出的 | 首字母大写的标识符，可被其他包访问 |
| fallthrough | 贯穿 | `switch` 中继续执行下一个 case 的关键字 |
| field tag | 字段标签 | 结构体字段后的反引号元数据，供 JSON/GORM 等读取 |
| function value | 函数值 | 把函数当作值赋给变量、传参或返回 |

---

## E.3 G–I

| 术语 | 中文 | 一句话解释 |
|------|------|------------|
| generic | 泛型 | 用类型参数编写可复用于多种类型的代码 |
| goroutine | 协程（Go 运行时调度） | Go 管理的轻量级并发执行单元，`go f()` 启动 |
| go.mod | 模块定义文件 | 记录模块路径、Go 版本与依赖版本 |
| go.sum | 依赖校验文件 | 记录依赖内容的校验和，保证下载一致 |
| go.work | 工作区文件 | 把多个模块作为整体开发的多模块工作区配置 |
| handler | 处理器 | 处理请求的函数或对象；Gin 中是 `func(*Context)` |
| interface | 接口 | 一组方法签名的集合；Go 通过方法集隐式实现 |
| interface value | 接口值 | 由「动态类型 + 动态值」组成的二元组 |
| internal package | 内部包 | 放在 `internal/` 下的包，只能被其父目录树内的代码导入 |
| iota | 常量计数器 | `const` 块中的自增计数器，用于枚举 |
| init function | 初始化函数 | 包初始化时自动执行、不能被显式调用的函数 |

---

## E.4 J–L

| 术语 | 中文 | 一句话解释 |
|------|------|------------|
| JSON tag | JSON 标签 | `json:"name,omitempty"`，控制 JSON 字段名与省略行为 |
| key | 键 | map 中用于查找的键，必须可比较 |
| len | 长度 | 内建函数：字符串字节数、切片/数组/map/channel 元素数 |
| cap | 容量 | 内建函数：切片容量或 channel 缓冲容量 |
| lexical scope | 词法作用域 | 标识符在源码块结构中的可见范围 |
| lock | 锁 | 保护共享状态的同步原语，如 `sync.Mutex` |

---

## E.5 M–O

| 术语 | 中文 | 一句话解释 |
|------|------|------------|
| map | 映射 | 无序键值集合，key 必须可比较 |
| method | 方法 | 带接收者的函数，接收者类似 TS 的 `this` |
| method set | 方法集 | 一个类型可调用的方法集合，决定它实现哪些接口 |
| middleware | 中间件 | 在请求处理前后执行的函数；Gin 中类型为 `HandlerFunc` |
| module | 模块 | 由 `go.mod` 定义的一组包的版本化单元 |
| mutex | 互斥锁 | `sync.Mutex`，保证同一时刻只有一个 goroutine 进入临界区 |
| named return | 命名返回值 | 在签名中给返回值命名，可直接 `return` |
| nil | 空值 | 指针、切片、map、channel、函数、接口的零值，表示「没有值」 |
| nil interface trap | nil 接口陷阱 | 接口中装入 nil 具体指针后，接口本身不等于 nil |
| overflow | 溢出 | 数值超出目标类型范围后被截断或回绕 |
| optional value | 可选值 | Go 用 `*T`、`sql.NullString` 等类型或额外布尔字段表达「可能没有」 |

---

## E.6 P–R

| 术语 | 中文 | 一句话解释 |
|------|------|------------|
| package | 包 | 同一目录下的一组 Go 文件，共享包名与命名空间 |
| panic | 运行时恐慌 | 触发崩溃的内建函数，通常表示不可恢复错误 |
| pointer | 指针 | 保存变量地址的值，类型为 `*T` |
| pointer receiver | 指针接收者 | `func (t *T) M()`，可修改原值；只有 `*T` 的方法集包含它 |
| pseudo-version | 伪版本 | 依赖未打标签时生成的版本号，形如 `v0.0.0-20240101120000-abcdef123456` |
| race condition | 数据竞争 | 多 goroutine 无同步访问同一内存且至少一个写 |
| receive | 接收 | 从 channel 取值，写作 `<-ch` |
| receiver | 接收者 | 方法名前的参数，类似 TS 的 `this` |
| recover | 恢复 | 在 `defer` 中捕获 panic 的内建函数，只对当前 goroutine 有效 |
| reflection | 反射 | 运行时检查类型与值的能力，对应 `reflect` 包 |
| replace directive | 替换指令 | `go.mod` 中用本地路径或另一版本替换依赖 |
| rune | Unicode 码点 | `int32` 的别名，表示一个字符 |
| RWMutex | 读写锁 | 允许多个读或一个写的锁 |

---

## E.7 S–U

| 术语 | 中文 | 一句话解释 |
|------|------|------------|
| select | 多路选择 | 同时等待多个 channel 操作的控制结构 |
| semver | 语义化版本 | `vMAJOR.MINOR.PATCH`，Go 模块用它表达兼容性 |
| send | 发送 | 向 channel 写值，写作 `ch <- v` |
| sentinel error | 哨兵错误 | 预定义的包级错误值，如 `io.EOF`，用 `errors.Is` 判断 |
| shadowing | 遮蔽 | 内层作用域声明同名变量，遮住外层变量 |
| slice | 切片 | 指向底层数组的「指针 + len + cap」视图 |
| slice header | 切片头 | 切片的三元组：底层数组指针、长度、容量 |
| slice expression | 切片表达式 | `s[low:high]` 或 `s[low:high:max]` |
| struct | 结构体 | 一组命名字段的聚合类型，值语义 |
| struct tag | 结构体标签 | 字段后的元数据字符串，见 field tag |
| sync.Map | 并发 map | 并发安全的 map，适合读多写少、key 集合稳定的场景 |
| sync.WaitGroup | 等待组 | 等待一组 goroutine 完成的计数器 |
| sql.Scanner | 数据库扫描接口 | 定义 `Scan(src any) error`，把数据库值读入自定义类型 |
| standard library | 标准库 | Go 安装自带的包集合 |

---

## E.8 V–Z

| 术语 | 中文 | 一句话解释 |
|------|------|------------|
| value receiver | 值接收者 | `func (t T) M()`，操作的是副本 |
| value semantics | 值语义 | 赋值/传参复制整个值 |
| variadic | 可变参数 | `...T`，调用时可传零个或多个 T |
| vendor | 依赖目录 | 把依赖源码放入项目 `vendor/`，构建时不走模块缓存 |
| workspace | 工作区 | 用 `go.work` 把多个模块作为整体开发 |
| zero value | 零值 | 变量声明未赋值时的默认值：`0`、`""`、`false`、`nil` 等 |

---

## E.9 Gin / GORM 术语

| 术语 | 中文 | 一句话解释 |
|------|------|------------|
| Engine | 引擎 | Gin 的顶层对象，持有路由树与配置，实现 `http.Handler` |
| RouterGroup | 路由组 | 带公共前缀与中间件的路由注册单元 |
| Context | 上下文对象 | Gin 对请求/响应的封装，贯穿整条 handler 链 |
| HandlerFunc | 处理器函数 | `func(*gin.Context)`，路由与中间件的统一类型 |
| HandlersChain | 处理器链 | `[]HandlerFunc`，一次请求按顺序执行的函数列表 |
| H | JSON 快捷 map | `gin.H`，即 `map[string]any` |
| Binding | 绑定策略 | 把请求数据写入结构体的接口，如 JSON、form、query |
| Render | 渲染策略 | 把响应写出的接口，如 JSON、HTML、text |
| ResponseWriter | 响应写入器 | Gin 对 `http.ResponseWriter` 的包装，记录状态码与大小 |
| Abort | 中止链 | 把 handler 索引推到末尾，后续 handler 不再执行 |
| gorm.DB | GORM 数据库对象 | 链式调用的入口，持有配置、`Statement`、`Error`、`RowsAffected` |
| Statement | 语句状态 | 一次数据库操作的上下文：表、条件、SQL、变量、Schema |
| Session | 会话 | 带独立配置的 GORM 会话，如 `DryRun`、`NewDB`、`SkipHooks` |
| Dialector | 数据库方言 | 连接与生成 SQL 的数据库适配接口 |
| gorm.Model | GORM 基础模型 | 嵌入后提供 `ID`、`CreatedAt`、`UpdatedAt`、`DeletedAt` |
| clause | SQL 片段 | GORM 内部构建 WHERE/SELECT/ORDER 等子句的表达式 |
| hook | 钩子 | 在创建/更新/删除/查询前后自动调用的方法 |
| association | 关联 | `BelongsTo`、`HasOne`、`HasMany`、`Many2Many` 关系 |
| Preload | 预加载 | 先查主表，再按关联条件批量查关联表，避免 N+1 |
| N+1 | N+1 查询 | 查 1 次列表后又为每条记录查 1 次关联，导致查询爆炸 |
| soft delete | 软删除 | 不物理删除，只写 `deleted_at`，默认查询自动过滤 |
| AutoMigrate | 自动迁移 | 根据结构体创建/补充表与索引，适合开发期 |
| DSN | 数据源名称 | 数据库连接字符串，包含地址、账号、库名、参数 |
| RowsAffected | 影响行数 | 本次操作影响的行数，用于判断更新/删除是否命中 |
| ErrRecordNotFound | 记录不存在 | GORM 的哨兵错误，`First`/`Take` 查不到时返回 |

---

## E.10 易混术语对照

| 易混项 | 区别 |
|--------|------|
| 零值 vs nil | 零值是每种类型的默认值；nil 只是引用类类型的零值 |
| 类型别名 vs 定义类型 | `type A = B` 完全等价；`type A B` 是新类型，需转换 |
| 接口 vs 接口值 | 接口是类型（方法集）；接口值是运行时的「类型 + 值」 |
| 方法 vs 函数 | 方法有接收者，可被接口方法集收录；函数没有 |
| 数组 vs 切片 | 数组长度固定且是值；切片是带 len/cap 的视图 |
| 切片 vs 切片头 | 切片是使用中的值；切片头是底层指针、len、cap 三元组 |
| 模块 vs 包 | 模块是版本化单元（`go.mod`）；包是目录级代码组织 |
| 并发 vs 并行 | 并发是任务重叠推进；并行是同一时刻真正同时执行 |
| panic vs error | error 是可预期失败，应返回并处理；panic 是不可恢复崩溃 |
| defer vs finally | defer 在函数返回前按 LIFO 执行；finally 在 try/catch 后执行 |
| 值接收者 vs 指针接收者 | 值接收者操作副本；指针接收者可改原值，方法集不同 |
| 数据竞争 vs 死锁 | 竞争是无同步的并发访问；死锁是所有人互相等待 |
| Preload vs Joins | Preload 发多条查询并拼装；Joins 用 SQL JOIN 一次查回 |
| Session vs Transaction | Session 是配置/复用上下文；Transaction 是原子提交/回滚边界 |
