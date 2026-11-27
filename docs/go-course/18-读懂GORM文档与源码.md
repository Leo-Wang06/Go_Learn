# 第 18 章：如何读懂 GORM 文档与源码
> **阶段**：阶段五 · 读文档与源码
> **建议时间**：2 天
> **前置知识**：第 05–08 章（结构体、指针、接口、错误处理）、第 13 章（`database/sql`）、第 17 章（Gin 文档阅读方法）
> **代码基准**：Go 1.22 及以上，GORM v2（`gorm.io/gorm` v1.25+），示例数据库使用 SQLite

---
## 本章目标
学完本章，你应当能够：

1. 找到 GORM v2 的文档、API 索引和源码入口，分清 v1 与 v2；
2. 画出 `gorm.DB`、`gorm.Statement`、`gorm.Session`、`gorm.Config` 的关系；
3. 解释链式方法为什么返回 `*gorm.DB`，错误为什么在 `Error` 字段；
4. 说清链式对象何时复用、何时克隆，避免条件串味；
5. 读懂模型标签与 CRUD 签名，理解指针、零值、`RowsAffected`；
6. 读懂四种关联与 `Preload`，知道 N+1 怎么产生；
7. 会写事务、钩子、`Valuer/Scanner`，能用 `WithContext`；
8. 从 `gorm.go`、`chainable_api.go`、`statement.go`、`callbacks/` 定位源码；
9. 避开零值更新、软删除、预加载、连接池等高频坑。
---
## 18.1 文档入口与 v2 版本
### 18.1.1 三个入口
| 入口 | 地址 | 用途 |
|------|------|------|
| 官方指南 | `https://gorm.io/docs/` | 按主题讲解 |
| API 文档 | `https://pkg.go.dev/gorm.io/gorm` | 查类型、签名、字段 |
| 源码仓库 | `https://github.com/go-gorm/gorm` | 看实现与 release notes |

推荐阅读顺序：CRUD 四页 → 模型与关联 → 事务/钩子/Session/Context → 高级与安全。不要从第一页硬读到最后一页。
### 18.1.2 v1 与 v2
| 项目 | GORM v1 | GORM v2 |
|------|---------|---------|
| 导入路径 | `github.com/jinzhu/gorm` | `gorm.io/gorm` |
| 驱动 | `.../gorm/dialects/mysql` | `gorm.io/driver/mysql` |
| 事务 | `db.Begin()` 为主 | `db.Transaction(func(tx *gorm.DB) error)` |
| 关联预加载 | 能力有限 | 支持嵌套、条件、排序 |
| 批量写入 | 支持有限 | `CreateInBatches`、`OnConflict` |

看到教程导入 `github.com/jinzhu/gorm`，可以直接关掉——那是 v1，API 大量不兼容。驱动是独立模块，用哪个数据库就引入哪个：
```go
import (
    "gorm.io/driver/mysql"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)
```
### 18.1.3 实验环境
```powershell
go mod init gormlab
go get gorm.io/gorm gorm.io/driver/sqlite
```
`gorm.io/driver/sqlite` 依赖 CGO。Windows 没有 GCC 时会报 `cgo: C compiler "gcc" not found`：安装 MinGW-w64/TDM-GCC，或改用纯 Go 驱动 `github.com/glebarez/sqlite`（`Open` 用法相同，但非官方驱动）。

本章统一使用官方 SQLite 驱动。核心初始化代码：
```go
type User struct {
    gorm.Model
    Name  string `gorm:"size:64;not null"`
    Age   int    `gorm:"default:18"`
    Email string `gorm:"uniqueIndex;size:128"`
}

func openDB() *gorm.DB {
    db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
    if err != nil {
        log.Fatal(err)
    }
    sqlDB, _ := db.DB()
    sqlDB.SetMaxOpenConns(1) // 内存库必须单连接，否则连接间看不到同一份数据
    if err := db.AutoMigrate(&User{}); err != nil {
        log.Fatal(err)
    }
    return db
}
```
`db.NamingStrategy.TableName("User")` 返回 `users`，`OrderItem` 返回 `order_items`：默认策略是蛇形 + 复数。
---
## 18.2 核心类型地图
```text
gorm.Config ──────────────┐ 被所有克隆共享
   ▲                      │
   │ 嵌入                 │
gorm.DB ──────────────────┘
   ├── Error        error      最近一次操作的错误
   ├── RowsAffected int64      受影响行数
   └── Statement    *Statement 本次 SQL 的构建状态
                          ├── Table / Model / Dest
                          ├── Clauses  子句集合
                          ├── Vars     占位符参数
                          ├── SQL      构建出的 SQL
                          └── Context  请求上下文
```
`gorm.Session` 是传给 `db.Session()` 的**配置结构体**，不是数据库连接；它返回的仍然是 `*gorm.DB`。
### 18.2.1 `gorm.DB`
源码位置：`gorm.go`。
```go
type DB struct {
    *Config
    Error        error
    RowsAffected int64
    Statement    *Statement
    clone        int
}
```
| 字段 | 含义 | 何时看 |
|------|------|--------|
| `Config` | 全局配置 | 初始化时 |
| `Error` | 最近一次操作的错误 | 每次执行完必看 |
| `RowsAffected` | 影响行数 | 判断更新/删除是否命中 |
| `Statement` | SQL 构建状态 | 调试 SQL、写插件 |
| `clone` | 内部克隆标记 | 理解链式复用 |

`Error` 和 `RowsAffected` 是字段，不是方法：写 `result.Error`，不是 `result.Error()`。
### 18.2.2 `gorm.Statement`
源码位置：`statement.go`。它保存「一条 SQL 正在被拼成什么样」：

| 字段 | 含义 |
|------|------|
| `Table` / `Model` | 表名、模型 |
| `Dest` | 查询结果写到哪里 |
| `Clauses` | `WHERE`、`ORDER BY`、`LIMIT` 等子句 |
| `Vars` | `?` 对应的参数 |
| `SQL` | 最终 SQL |
| `Schema` | 模型元数据（表、字段、关联） |
| `Context` | 上下文 |
| `Unscoped` | 是否跳过软删除条件 |
| `Preloads` | 待执行的预加载 |

链式方法本质上是在往 `Statement` 里塞东西；真正执行发生在 `Find`、`Create` 这类终结方法里。
### 18.2.3 `gorm.Session`
| 字段 | 作用 |
|------|------|
| `NewDB` | 丢弃已有条件，开全新 Statement |
| `DryRun` | 只生成 SQL，不执行 |
| `PrepareStmt` | 预编译语句缓存 |
| `SkipHooks` | 跳过钩子 |
| `SkipDefaultTransaction` | 跳过默认事务 |
| `AllowGlobalUpdate` | 允许无 `WHERE` 的全局更新/删除 |
| `FullSaveAssociations` | 全量保存关联 |
| `Context` | 上下文 |
| `Logger` / `NowFunc` | 本次会话的日志器、时间函数 |
| `CreateBatchSize` | 批量插入分批大小 |
### 18.2.4 `gorm.Config`
在 `gorm.Open` 时传入，之后被所有克隆共享：

| 字段 | 作用 |
|------|------|
| `SkipDefaultTransaction` | 单条写操作不自动包事务 |
| `NamingStrategy` | 表名/列名策略 |
| `Logger` | 日志器 |
| `NowFunc` | 时间来源（测试可固定） |
| `DryRun` | 只生成 SQL |
| `PrepareStmt` | 预编译缓存 |
| `CreateBatchSize` | 批量插入大小 |
| `TranslateError` | 把驱动错误翻译成 `gorm.ErrDuplicatedKey` 等 |
```go
db, err := gorm.Open(sqlite.Open("app.db"), &gorm.Config{
    SkipDefaultTransaction: true,
    NamingStrategy: schema.NamingStrategy{TablePrefix: "t_", SingularTable: true},
    Logger:         logger.Default.LogMode(logger.Info),
})
```
`Config` 在 `Open` 后应视为只读。不要在请求中改 `db.Config.Logger`，并发下会出问题。
### 18.2.5 `gorm.Model`
```go
type Model struct {
    ID        uint `gorm:"primarykey"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt DeletedAt `gorm:"index"`
}
```
嵌入后自动拥有自增主键、创建时间、更新时间、软删除字段。不需要这些字段时，完全可以自定义模型。
---
## 18.3 链式方法：为什么能一直点
### 18.3.1 签名
```go
func (db *DB) Where(query interface{}, args ...interface{}) (tx *DB)
func (db *DB) Order(value interface{}) (tx *DB)
func (db *DB) Limit(limit int) (tx *DB)
func (db *DB) Find(dest interface{}, conds ...interface{}) (tx *DB)
```
每个方法返回 `*gorm.DB`，所以能继续点：
```go
var users []User
result := db.Where("age > ?", 18).Order("age DESC").Limit(10).Find(&users)
```
每个链式方法大致做两件事：`tx := db.getInstance()` 决定克隆还是复用；把条件写进 `tx.Statement`，返回 `tx`。
### 18.3.2 错误为什么在 `Error` 字段
链式调用无法每一步都返回 `(*gorm.DB, error)`，否则要写成：
```go
db, err := db.Where(...)
db, err = db.Order(...)
```
GORM 的选择是「错误先存起来，最后统一取」：
```go
result := db.Where("age > ?", 18).Find(&users)
if result.Error != nil {
    log.Println("查询失败:", result.Error)
    return
}
fmt.Println("查到", result.RowsAffected, "条")
```
- 链式方法通常不返回错误；
- 终结方法执行后把错误写进 `Error`；
- `Error` 非 nil 时，后续终结方法通常直接返回，不再执行 SQL；
- 永远检查 `result.Error`，很多 GORM 方法根本没有 `err` 返回值。
### 18.3.3 克隆与复用
`getInstance()` 的规则可简化为：

| `clone` | 行为 |
|---------|------|
| `1` | 新建空 `Statement`（干净起点） |
| `2` | 复制当前 `Statement`（带上已有条件） |
| `0` | 直接返回自己 |

三条经验：

**一、链式方法不修改原对象。**
```go
base := db.Model(&User{}).Where("age > ?", 18)

var adults []User
base.Find(&adults) // age > 18

var named []User
base.Where("name = ?", "张三").Find(&named) // age > 18 AND name = '张三'
// base 仍然只有 age > 18
```
**二、把链式对象存起来重复用，条件会累积。**
```go
query := db.Model(&User{}).Where("age > ?", 18)
query = query.Where("name = ?", "张三")
query = query.Where("status = ?", "active")
query.Find(&users) // 三个条件都在
```
这不是 bug，是设计。但跨请求共享同一个 `query`，条件会互相污染。

**三、需要干净起点时用 `Session(NewDB: true)`。**
```go
clean := db.Session(&gorm.Session{NewDB: true})
clean.Find(&users) // 没有残留条件
```
项目建议：不要把「拼了一半」的 `*gorm.DB` 存进全局变量或单例；每个请求从 `db.WithContext(ctx)` 开始，在本函数内完成拼装和执行。
### 18.3.4 `WithContext` 与 `Session` 的关系
```go
func (db *DB) WithContext(ctx context.Context) *DB {
    return db.Session(&Session{Context: ctx})
}
```
所以 `db.WithContext(c.Request.Context()).Where(...).Find(&users)` 会把 context 一路传到 `database/sql` 的 `QueryContext`。
---
## 18.4 模型与标签
### 18.4.1 默认映射
| Go | 数据库 |
|----|--------|
| `UserProfile` | 表 `user_profiles` |
| `UserName` | 列 `user_name` |
| `ID` | 列 `id` |
| 嵌入 `gorm.Model` | `id`、`created_at`、`updated_at`、`deleted_at` |

覆盖表名：
```go
func (User) TableName() string { return "app_users" }
```
或临时指定：`db.Table("app_users").Find(&users)`。
### 18.4.2 常用标签
| 标签 | 作用 | 示例 |
|------|------|------|
| `primaryKey` | 主键 | `gorm:"primaryKey"` |
| `autoIncrement` | 自增 | `gorm:"autoIncrement"` |
| `column` | 列名 | `gorm:"column:user_name"` |
| `type` | 数据库类型 | `gorm:"type:varchar(100)"` |
| `size` | 长度 | `gorm:"size:64"` |
| `not null` | 非空 | `gorm:"not null"` |
| `default` | 默认值 | `gorm:"default:18"` |
| `uniqueIndex` | 唯一索引 | `gorm:"uniqueIndex"` |
| `index` | 普通索引 | `gorm:"index"` |
| `embedded` | 嵌入结构体 | `gorm:"embedded"` |
| `embeddedPrefix` | 嵌入字段前缀 | `gorm:"embeddedPrefix:addr_"` |
| `serializer` | 序列化方式 | `gorm:"serializer:json"` |
| `many2many` | 多对多中间表 | `gorm:"many2many:user_languages;"` |
| `foreignKey` | 外键 | `gorm:"foreignKey:CompanyID"` |
| `references` | 被引用列 | `gorm:"references:ID"` |
| `constraint` | 外键行为 | `gorm:"constraint:OnDelete:CASCADE;"` |

多个标签用分号：`gorm:"type:varchar(128);uniqueIndex;not null"`。
### 18.4.3 完整模型
```go
type Address struct {
    City    string `gorm:"size:32"`
    Country string `gorm:"size:32"`
}

type Customer struct {
    ID      uint    `gorm:"primaryKey"`
    Name    string  `gorm:"column:customer_name;size:64;not null"`
    Age     int     `gorm:"default:18"`
    Email   string  `gorm:"uniqueIndex;size:128"`
    Address Address `gorm:"embedded;embeddedPrefix:addr_"`
}

func (Customer) TableName() string { return "customers" }
```
迁移后打印列名：
```go
db.AutoMigrate(&Customer{})
columns, _ := db.Migrator().ColumnTypes(&Customer{})
for _, col := range columns {
    fmt.Println(col.Name())
}
```
输出：
```text
id
customer_name
age
email
addr_city
addr_country
```
`embedded` 把 `Address` 平铺进 `customers` 表，`embeddedPrefix` 避免字段重名。

标签只能被**导出字段**读取。小写字段 GORM 完全看不到——这和第 01 章的导出规则、第 05 章的反射、第 13 章的 JSON 标签是同一条知识线。
---
## 18.5 CRUD 签名与零值
### 18.5.1 Create
```go
func (db *DB) Create(value interface{}) (tx *DB)
```
必须传指针，因为 GORM 要把主键、时间戳写回结构体：
```go
user := User{Name: "张三", Age: 25, Email: "zhangsan@example.com"}
result := db.Create(&user)
fmt.Println(result.Error, result.RowsAffected, user.ID) // <nil> 1 1
```
批量插入：
```go
users := []User{
    {Name: "张三", Age: 25, Email: "zhangsan@example.com"},
    {Name: "李四", Age: 17, Email: "lisi@example.com"},
    {Name: "王五", Age: 30, Email: "wangwu@example.com"},
}
db.Create(&users) // ID 依次回填 1、2、3
```
大数据量用 `db.CreateInBatches(&users, 100)` 控制每批大小。
### 18.5.2 First / Take / Last / Find
| 方法 | 排序 | 查不到时 | 用途 |
|------|------|----------|------|
| `First(&u)` | 主键升序 | `gorm.ErrRecordNotFound` | 取第一条 |
| `Take(&u)` | 不指定 | `gorm.ErrRecordNotFound` | 随便取一条 |
| `Last(&u)` | 主键降序 | `gorm.ErrRecordNotFound` | 取最后一条 |
| `Find(&us)` | 不指定 | **不报错**，空切片 | 列表查询 |
```go
var first, last, byID User
db.First(&first)      // 张三
db.Last(&last)        // 王五
db.First(&byID, 2)    // 李四（主键 = 2）

var filtered []User
db.Where("age > ?", 18).Order("age DESC").Find(&filtered)
for _, u := range filtered {
    fmt.Println(u.Name, u.Age) // 王五 30 / 张三 25
}
```
查不到时：`First` 返回 `gorm.ErrRecordNotFound`；`Find` 返回 `nil` 和空切片。
```go
var missing User
err := db.First(&missing, 999).Error
fmt.Println(errors.Is(err, gorm.ErrRecordNotFound)) // true
```
> **高频坑**：列表查询用 `Find`。用 `First` 判断「列表为空」会把正常空结果当成错误。
### 18.5.3 条件写法
```go
db.Where("age > ?", 18).Find(&users)
db.Where(&User{Name: "张三"}).Find(&users)
db.Where(map[string]interface{}{"name": "张三", "age": 25}).Find(&users)
db.Not("age = ?", 18).Find(&users)
db.Where("name LIKE ?", "%张%").Find(&users)
```
用 struct 做条件时，零值字段被忽略：`db.Where(&User{Name: "张三", Age: 0})` 只生成 `name = '张三'`。需要匹配零值时用 map。
### 18.5.4 Updates / Update / Save 与零值
```go
var u User
db.First(&u, 1)

db.Model(&u).Updates(User{Age: 0})                        // 零值被忽略
db.Model(&u).Updates(map[string]interface{}{"age": 0})    // 写入
db.Model(&u).Update("age", 30)                            // 单列，写入
db.Model(&u).UpdateColumn("age", 0)                       // 跳过钩子，不更新时间
db.Save(&u)                                               // 全量保存
```
| 写法 | 零值 | 钩子 | 更新时间 | 场景 |
|------|------|------|----------|------|
| `Updates(User{...})` | 忽略 | 触发 | 更新 | 部分字段 |
| `Updates(map{...})` | 写入 | 触发 | 更新 | 需要零值 |
| `Update` / `UpdateColumn` | 写入 | 后者跳过 | 后者不更新 | 单列 |
| `Save(&u)` | 全部写入 | 触发 | 更新 | 全量保存 |

`Save` 的规则：主键为零时插入，主键非零时更新**所有字段**（包括零值）。
### 18.5.5 Delete 与软删除
模型含 `gorm.DeletedAt` 时，`Delete` 默认软删除：
```go
db.Delete(&u)

var count int64
db.Model(&User{}).Count(&count)            // 2（不含已删除）
db.Unscoped().Model(&User{}).Count(&count) // 3（含已删除）
db.Unscoped().Delete(&u)                   // 物理删除
```
批量删除必须带条件，否则返回 `gorm.ErrMissingWhereClause`：
```go
db.Where("age < ?", 18).Delete(&User{})
```
确实要清空表，显式允许：
```go
db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&User{})
```
`db.Where("1 = 1").Delete(&User{})` 能绕过检查，但这是危险写法，代码评审应直接拒绝。
---
## 18.6 关联
### 18.6.1 四种关系
| 关系 | 外键位置 | 例子 |
|------|----------|------|
| BelongsTo | 当前表 | Employee 属于 Company，`Employee.CompanyID` |
| HasOne | 对方表 | Employee 有一个 Badge，`Badge.EmployeeID` |
| HasMany | 对方表 | Employee 有多个 Order，`Order.EmployeeID` |
| Many2Many | 中间表 | Employee 会多种 Language，`employee_languages` |
```go
type Company struct {
    ID   uint
    Name string
}

type Badge struct {
    ID         uint
    EmployeeID uint
    Number     string
}

type Order struct {
    ID         uint
    EmployeeID uint
    Amount     float64
}

type Language struct {
    ID   uint
    Name string
}

type Employee struct {
    ID        uint
    Name      string
    CompanyID uint
    Company   Company `gorm:"foreignKey:CompanyID"`
    Badge     Badge
    Orders    []Order
    Languages []Language `gorm:"many2many:employee_languages;"`
}

db.AutoMigrate(&Company{}, &Badge{}, &Order{}, &Language{}, &Employee{})
```
### 18.6.2 创建带关联的数据
```go
emp := Employee{
    Name:      "张三",
    Company:   Company{Name: "Acme"},
    Badge:     Badge{Number: "A-001"},
    Orders:    []Order{{Amount: 99.5}, {Amount: 200}},
    Languages: []Language{{Name: "Go"}, {Name: "TypeScript"}},
}
db.Create(&emp)
```
GORM 会按顺序创建关联，自动回填外键和中间表记录。
### 18.6.3 Preload
```go
var employees []Employee
db.Preload("Company").Preload("Badge").
    Preload("Orders").Preload("Languages").
    Find(&employees)

fmt.Println(employees[0].Company.Name)                    // Acme
fmt.Println(len(employees[0].Orders))                     // 2
fmt.Println(len(employees[0].Languages))                  // 2
```
带条件与排序：
```go
db.Preload("Orders", "amount > ?", 100).Find(&employees)

db.Preload("Orders", func(db *gorm.DB) *gorm.DB {
    return db.Order("orders.amount DESC")
}).Find(&employees)

db.Preload("Orders.OrderItems").Find(&employees) // 嵌套
```
### 18.6.4 N+1
错误写法：1 次查员工，N 次查订单。
```go
var employees []Employee
db.Find(&employees)
for _, emp := range employees {
    var orders []Order
    db.Where("employee_id = ?", emp.ID).Find(&orders)
    fmt.Println(emp.Name, len(orders))
}
```
正确写法：
```go
db.Preload("Orders").Find(&employees)
for _, emp := range employees {
    fmt.Println(emp.Name, len(emp.Orders))
}
```
### 18.6.5 Preload 与 Joins
| 对比项 | `Preload` | `Joins` |
|--------|-----------|---------|
| 查询次数 | 1 + 关联数 | 通常 1 次 |
| 适用关系 | 全部 | BelongsTo、HasOne 更合适 |
| 填充关联字段 | 会 | 会（字段选择有限制） |
| 常见用途 | 列表页、详情页 | 按关联表字段过滤 |
```go
db.Joins("Company").Find(&employees) // BelongsTo
db.Preload("Orders").Find(&employees) // HasMany 更稳
```
### 18.6.6 Association 模式
```go
var emp Employee
db.First(&emp, 1)

db.Model(&emp).Association("Languages").Append(&Language{Name: "Rust"})
db.Model(&emp).Association("Languages").Count()
db.Model(&emp).Association("Languages").Clear()
```
`Association` 适合「已知主体，操作它的关联集合」；`Preload` 适合「查询主体时顺带加载」。
---
## 18.7 事务
### 18.7.1 `db.Transaction`
```go
func (db *DB) Transaction(fc func(tx *DB) error, opts ...*sql.TxOptions) error
```
规则：

- 返回 `nil`：提交；
- 返回 `error`：回滚；
- `panic`：回滚，并把 panic 继续抛出。

转账示例：
```go
type Account struct {
    ID      uint
    Name    string
    Balance int
}

func transfer(db *gorm.DB, fromID, toID uint, amount int) error {
    return db.Transaction(func(tx *gorm.DB) error {
        var from, to Account
        if err := tx.First(&from, fromID).Error; err != nil {
            return err
        }
        if err := tx.First(&to, toID).Error; err != nil {
            return err
        }
        if from.Balance < amount {
            return errors.New("余额不足")
        }

        if err := tx.Model(&from).
            Update("balance", gorm.Expr("balance - ?", amount)).Error; err != nil {
            return err
        }
        if err := tx.Model(&to).
            Update("balance", gorm.Expr("balance + ?", amount)).Error; err != nil {
            return err
        }
        return nil
    })
}
```
余额 A=100、B=0 时：转账 200 返回「余额不足」，余额不变；转账 50 成功，最终 A=50、B=50。
### 18.7.2 行锁
MySQL/PostgreSQL 下读取余额应加锁，避免并发读旧值：
```go
import "gorm.io/gorm/clause"

tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&from, fromID)
```
SQLite 不支持 `SELECT ... FOR UPDATE`，学习示例可先不加；生产 MySQL 转账必须加。
### 18.7.3 嵌套与默认事务
`db.Transaction` 内部再开 `db.Transaction`，GORM 默认用 `SAVEPOINT`。需要手动控制时可 `db.Begin()`、`tx.Commit()`、`tx.Rollback()`，但初学者优先用 `db.Transaction`，手动事务容易漏掉 `Rollback`。禁用嵌套用 `gorm.Config{DisableNestedTransaction: true}`。

GORM 默认把单条 `Create/Update/Delete` 包在事务里。高并发写入可关闭：
```go
gorm.Config{SkipDefaultTransaction: true}
```
关闭后，单条写不再自动包事务；多条写仍应显式使用 `db.Transaction`。
---
## 18.8 钩子
| 钩子 | 触发时机 |
|------|----------|
| `BeforeSave` | 创建/更新之前 |
| `BeforeCreate` / `AfterCreate` | 创建前/后 |
| `BeforeUpdate` / `AfterUpdate` | 更新前/后 |
| `BeforeDelete` / `AfterDelete` | 删除前/后 |
| `AfterFind` | 查询之后 |

签名统一：
```go
func (u *User) BeforeCreate(tx *gorm.DB) error {
    if u.Email == "" {
        return errors.New("email 不能为空")
    }
    return nil
}

func (u *User) AfterCreate(tx *gorm.DB) error {
    fmt.Println("AfterCreate:", u.ID)
    return nil
}
```
钩子返回 `error` 会中断后续操作并回滚。`AfterCreate` 里 `u.ID` 已回填。

边界：

- 钩子里用参数 `tx`，不要用外层 `db`，否则可能绕开当前事务；
- 不要在 `BeforeSave` 里再次 `Save` 同一个对象，容易递归；
- 批量操作会逐条触发钩子，注意性能；
- `UpdateColumn`、`UpdateColumns` 跳过钩子；
- 需要时用 `db.Session(&gorm.Session{SkipHooks: true})`。
---
## 18.9 回调与 Clause（只读层面）
### 18.9.1 回调
每个操作都由一串回调组成：
```text
Create: begin_transaction → before_create → save_before_associations
        → create → after_create → save_after_associations → commit_or_rollback
Query:  before_query → query → after_query
Update: begin_transaction → before_update → save_before_associations
        → update → after_update → save_after_associations → commit_or_rollback
Delete: begin_transaction → before_delete → delete → after_delete → commit_or_rollback
```
源码在 `callbacks/callbacks.go`，各操作在 `callbacks/create.go`、`query.go`、`update.go`、`delete.go`。

只读观察示例：
```go
db.Callback().Create().Before("gorm:create").Register("demo:before_create", func(tx *gorm.DB) {
    fmt.Println("callback: before create")
})
// 删除：db.Callback().Create().Remove("demo:before_create")
```
回调无返回值，出错要 `tx.AddError(err)`。

> 业务代码不应把核心逻辑写进回调。回调适合审计、埋点、多租户过滤等横切关注点；读源码能看懂顺序即可。
### 18.9.2 Clause
`clause` 包把 SQL 片段抽象成结构体：

| 类型 | 对应 SQL |
|------|----------|
| `clause.Where` / `clause.Eq` | `WHERE` / `column = value` |
| `clause.OrderBy` | `ORDER BY` |
| `clause.Locking` | `FOR UPDATE` |
| `clause.OnConflict` | `ON CONFLICT` / `ON DUPLICATE KEY` |
| `clause.Returning` | `RETURNING` |
```go
import "gorm.io/gorm/clause"

db.Clauses(clause.Where{
    Exprs: []clause.Expression{
        clause.Eq{Column: clause.Column{Name: "age"}, Value: 18},
    },
}).Find(&users)
```
`ToSQL` 不执行 SQL，只走一遍构建流程：
```go
sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
    return tx.Model(&User{}).Where("age > ?", 18).
        Order("age DESC").Limit(3).Find(&[]User{})
})
fmt.Println(sql)
```
输出（不同版本反引号风格可能不同）：
```text
SELECT * FROM `users` WHERE age > 18 AND `users`.`deleted_at` IS NULL ORDER BY age DESC LIMIT 3
```
Upsert：
```go
db.Clauses(clause.OnConflict{
    Columns:   []clause.Column{{Name: "id"}},
    DoUpdates: clause.AssignmentColumns([]string{"name", "age"}),
}).Create(&users)
```
`ToSQL` 是学习 GORM 最有用的调试工具：看不懂某个链式调用生成了什么，就打印出来。
---
## 18.10 自定义类型：Valuer / Scanner
数据库只认识有限类型。`[]string` 想存成 JSON 字段，需要：

- 写入：Go 类型 → 数据库值，实现 `driver.Valuer`；
- 读取：数据库值 → Go 类型，实现 `sql.Scanner`。
```go
type Tags []string

func (t Tags) Value() (driver.Value, error) {
    if t == nil {
        return nil, nil
    }
    data, err := json.Marshal(t)
    if err != nil {
        return nil, err
    }
    return string(data), nil
}

func (t *Tags) Scan(src any) error {
    if src == nil {
        *t = nil
        return nil
    }
    switch v := src.(type) {
    case []byte:
        return json.Unmarshal(v, t)
    case string:
        return json.Unmarshal([]byte(v), t)
    default:
        return fmt.Errorf("Tags.Scan: 不支持的类型 %T", src)
    }
}

type Article struct {
    ID   uint
    Tags Tags `gorm:"type:text"`
}
```
使用：
```go
db.AutoMigrate(&Article{})
article := Article{Tags: Tags{"go", "gorm"}}
db.Create(&article)

var got Article
db.First(&got, article.ID)
fmt.Println(got.Tags) // [go gorm]
```
三个细节：`Value` 用值接收者，`Scan` 用指针接收者（要改接收者）；必须处理 `nil`，数据库 `NULL` 传给 `Scan` 时是 `nil`；`driver.Value` 只允许 `int64`、`float64`、`bool`、`[]byte`、`string`、`time.Time`、`nil`，不要返回 `[]string`。

简单替代：`Tags []string `gorm:"serializer:json"``。但很多第三方类型（枚举、金额、加密字段）仍用 `Valuer/Scanner`，读文档要认识。

NULL 的三种表达：

| 写法 | 能表达 NULL | 读取后 |
|------|-------------|--------|
| `string` | 否（零值 `""`） | `""` |
| `*string` | 是 | `nil` 或指针 |
| `sql.NullString` | 是 | `.Valid` / `.String` |
---
## 18.11 Context、错误与日志
### 18.11.1 WithContext
```go
func GetUser(db *gorm.DB, c *gin.Context, id uint) (*User, error) {
    var user User
    err := db.WithContext(c.Request.Context()).First(&user, id).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}
```
好处：客户端断开时查询可取消；可设置超时；TraceID 能挂到 context 上。用 `context.WithTimeout` 创建带超时的 ctx，传给 `WithContext`；超时后错误链里通常包含 `context deadline exceeded`，用 `errors.Is(err, context.DeadlineExceeded)` 判断。
### 18.11.2 ErrRecordNotFound
```go
var user User
err := db.First(&user, 999).Error
if errors.Is(err, gorm.ErrRecordNotFound) {
    // 404
}
```
- `First`、`Take`、`Last` 查不到返回该错误；
- `Find` 查不到返回 `nil`，切片为空；
- 用 `errors.Is`，不要用 `==`，错误可能被包装。
### 18.11.3 RowsAffected
```go
result := db.Model(&User{}).Where("id = ?", 999).Update("age", 18)
fmt.Println(result.Error == nil, result.RowsAffected) // true 0
```
更新执行成功但没有命中行，`Error` 仍是 `nil`。判断「用户不存在」必须看 `RowsAffected`。
### 18.11.4 日志
```go
import "gorm.io/gorm/logger"

db, err := gorm.Open(sqlite.Open("app.db"), &gorm.Config{
    Logger: logger.Default.LogMode(logger.Info),
})
```
| 级别 | 输出 |
|------|------|
| `logger.Silent` | 不输出 |
| `logger.Error` | 只输出错误 |
| `logger.Warn` | 错误 + 慢 SQL |
| `logger.Info` | 所有 SQL |

临时打开：`db.Debug().Find(&users)`。`Debug()` 返回使用 `Info` 级别的新会话，不影响全局。
---
## 18.12 源码阅读路径
### 18.12.1 目录地图
| 路径 | 内容 |
|------|------|
| `gorm.go` | `DB`、`Config`、`Session`、`Open`、`Transaction`、`WithContext`、`ToSQL` |
| `chainable_api.go` | `Model`、`Where`、`Order`、`Limit`、`Preload`、`Joins`、`Select`、`Unscoped` |
| `finisher_api.go` | `Create`、`First`、`Find`、`Updates`、`Delete`、`Count`、`Scan` |
| `statement.go` | `Statement`、`clone`、`AddClause`、SQL 构建 |
| `callbacks/` | 各操作的回调实现 |
| `schema/`、`clause/`、`migrator/`、`logger/` | 模型解析、SQL 子句、迁移、日志 |
| `association.go` | `Association` 模式 |
### 18.12.2 一次 Create 的调用链
```text
db.Create(&user)
  → finisher_api.go: Create
  → callbacks.Create().Execute(tx)
  → callbacks/callbacks.go: processor.Execute
  → callbacks/create.go: create 回调
  → 构建 INSERT → ConnPool.ExecContext
  → 回填主键、时间戳 → 返回 *gorm.DB
```
`Create` 的源码只有三行：`tx = db.getInstance()`、`tx.Statement.Dest = value`、`return tx.callbacks.Create().Execute(tx)`。
### 18.12.3 一次 Where 的调用链
```text
db.Where("age > ?", 18)
  → chainable_api.go: Where
  → tx := db.getInstance()
  → tx.Statement.BuildCondition(query, args...)
  → tx.Statement.AddClause(clause.Where{Exprs: conds})
  → 返回 tx
```
`AddClause` 会把新 `WHERE` 与已有 `WHERE` 合并，而不是覆盖。这就是多个 `Where` 能叠加的原因。
### 18.12.4 自己定位陌生 API
以 `Preload` 为例：

1. 打开 `pkg.go.dev/gorm.io/gorm`，搜索 `Preload`；
2. 看到签名 `func (db *DB) Preload(query string, args ...interface{}) *DB`；
3. 点 `View Source` 跳到 `chainable_api.go`；
4. 看到参数存进 `tx.Statement.Preloads`；
5. 再搜索 `Preloads`，找到 `callbacks/preload.go`。

方法：**pkg.go.dev 找签名 → 源码找字段 → 回调找执行**。

本地源码在 `go env GOMODCACHE` 下，例如 `...\gorm.io\gorm@v1.31.2\`。也可以用：
```powershell
go doc gorm.io/gorm.DB
go doc gorm.io/gorm.DB.Updates
go doc gorm.io/gorm/clause.OnConflict
```
---
## 18.13 常见错误与陷阱
| 现象 | 原因 | 修复 |
|------|------|------|
| `Updates(User{Age: 0})` 没更新 | struct 零值被忽略 | map、`Update` |
| `Save` 把空字段覆盖了 | 全量更新 | 用 `Updates` |
| 删除后还能查到 | 软删除默认过滤 | `Unscoped()` |
| `Count` 比预期少 | 软删除过滤 | `Unscoped().Count` |
| 全局更新/删除报错 | 缺 `WHERE` | 加条件或 `AllowGlobalUpdate` |
| 列表为空却报错 | 用了 `First` | 用 `Find` |
| 关联字段为空 | 没 `Preload` | `Preload("Orders")` |
| 列表页 SQL 很多 | N+1 | `Preload` / `Joins` |
| 事务部分成功 | 用了外层 `db` | 全程用 `tx` |
| 连接耗尽 | 连接池未配置 | 配 `SetMaxOpenConns` 等 |
| 时间差 8 小时 | 时区/`parseTime` | DSN 加 `parseTime=True&loc=Local` |
| 唯一键错误看不懂 | 驱动错误未翻译 | `TranslateError: true` |
| SQL 注入 | 拼接 SQL | 用 `?` 占位符 |

零值更新再强调一次：**struct 更新只更新非零字段**，这是 GORM 最高频的坑，没有之一。
```go
db.Model(&user).Updates(User{Age: 0})                     // 不生效
db.Model(&user).Updates(map[string]interface{}{"age": 0}) // 正确
db.Model(&user).Update("age", 0)                          // 正确
```
软删除：默认查询自动加 `deleted_at IS NULL`；`Unscoped()` 包含已删除或物理删除。软删除字段必须是 `gorm.DeletedAt` 类型。

预加载：循环里查关联就是 N+1，用 `Preload`。但关联特别多、数据特别大时，要考虑分页、按需加载、缓存，而不是把所有关联塞进一次查询。

连接池（GORM 底层是 `database/sql`）：
```go
sqlDB, err := db.DB()
if err != nil {
    log.Fatal(err)
}
sqlDB.SetMaxOpenConns(100)
sqlDB.SetMaxIdleConns(10)
sqlDB.SetConnMaxLifetime(time.Hour)
sqlDB.SetConnMaxIdleTime(10 * time.Minute)
```
经验：`MaxOpenConns` 不超过数据库 `max_connections`；`MaxIdleConns` 设为最大连接的 10%–30%；`ConnMaxLifetime` 小于数据库/代理的空闲超时；每个进程一套池，不要每个请求 `gorm.Open`。

事务里用错 `db`：
```go
db.Transaction(func(tx *gorm.DB) error {
    db.Create(&user) // 错误：不在事务里
    return tx.Create(&order).Error
})
```
`tx` 才绑定事务连接。混用外层 `db`，那条 SQL 会走另一个连接，回滚带不走它。

SQL 注入：
```go
db.Where(fmt.Sprintf("name = '%s'", name)) // 危险
db.Where("name = ?", name)                 // 正确
```
---
## 18.14 与 Gin 的关联：分层写法
真实项目按 handler → service → repository 分层。Repository 持有 `*gorm.DB`，查询时把请求上下文传下去：
```go
func (r *UserRepo) FindByID(ctx context.Context, id uint) (*User, error) {
    var user User
    if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
        return nil, err
    }
    return &user, nil
}
```
Handler 里用 `strconv.ParseUint(c.Param("id"), 10, 64)` 解析 ID，调用 repository 后：`errors.Is(err, gorm.ErrRecordNotFound)` 映射 404，其他错误记日志并返回 500，成功则 `c.JSON(200, user)`。

要点：`c.Request.Context()` 一路传到 GORM；数据库错误不要原样返回客户端；不要把 `*gorm.DB` 塞进 `gin.Context`，用依赖注入传给 handler。
---
## 18.15 练习题
### 基础题
1. 用 `go doc` 查看 `gorm.DB`、`Statement`、`Session` 的字段，抄写至少 5 个。
2. 创建 `User` 模型（含 `gorm.Model`、`Name`、`Age`、`Email`），迁移后打印列名。
3. 插入 3 条用户，打印回填 ID 和 `RowsAffected`。
4. 分别用 `First`、`Take`、`Last`、`Find` 查询，记录输出差异。
5. 用 `Where`、`Order`、`Limit`、`Offset` 写分页查询。
6. 演示 `Updates(struct)` 与 `Updates(map)` 对零值的不同行为。
7. 给模型加 `DeletedAt`，演示软删除、`Unscoped` 查询、物理删除。
8. 用 `Preload` 加载一对多关联，打印关联数量。
9. 用 `db.Transaction` 写一个失败回滚的例子。
10. 给模型加 `BeforeCreate` 钩子，验证空邮箱无法创建。
### 进阶题
11. 用 `ToSQL` 打印 `Where + Order + Limit + Find` 生成的 SQL。
12. 用 `clause.OnConflict` 实现「存在则更新，不存在则插入」。
13. 用 `driver.Valuer` / `sql.Scanner` 把 `[]string` 存成 JSON 字段。
14. 用 `Association` 给用户追加角色，并统计数量。
15. 用 `Joins("Company")` 查询属于某公司的员工。
16. 演示 N+1，再用 `Preload` 优化，比较 SQL 条数。
17. 用 `Session(NewDB: true)` 解决链式条件串味。
18. 配置连接池并解释四个参数。
19. 用 `WithContext` + 超时取消一个 MySQL 查询。
20. 写一个自定义回调，在每次创建前打印表名。
### 读文档题
21. 读 `Update` 页：`Updates` 支持哪些参数类型？零值如何处理？
22. 读 `Preload` 页：如何给预加载加条件和排序？
23. 读 `Transaction` 页：返回 error 和 panic 分别发生什么？
24. 读 `Hooks` 页：列出所有钩子及触发顺序。
25. 读 `Session` 页：解释 `NewDB`、`DryRun`、`SkipHooks`。
26. 在 pkg.go.dev 找到 `gorm.DB.ToSQL` 的源码位置，说明实现步骤。
---
## 18.16 验收清单
- [ ] 能说出 GORM v1 与 v2 的导入路径区别；
- [ ] 能画出 `gorm.DB`、`Statement`、`Session`、`Config` 的关系图；
- [ ] 能解释链式方法为什么返回 `*gorm.DB`，以及 `Error` 字段的作用；
- [ ] 能说清 `Session(NewDB: true)` 与普通链式复用的区别；
- [ ] 能读懂 `gorm.Model` 和至少 8 个常用标签；
- [ ] 能解释 `Create`、`First`、`Find`、`Updates`、`Delete` 的签名与指针要求；
- [ ] 能说清 struct 更新、map 更新、`Save`、`UpdateColumn` 的差别；
- [ ] 能解释软删除原理和 `Unscoped` 的用途；
- [ ] 能写出四种关联模型，并用 `Preload` 加载；
- [ ] 能解释 N+1，知道 `Preload` 与 `Joins` 的适用场景；
- [ ] 能写 `db.Transaction(func(tx *gorm.DB) error)` 并解释提交/回滚规则；
- [ ] 能写出至少 3 个钩子，知道钩子返回 error 的后果；
- [ ] 能实现 `driver.Valuer` 与 `sql.Scanner`；
- [ ] 能用 `WithContext` 传递请求上下文；
- [ ] 能定位 `gorm.go`、`chainable_api.go`、`statement.go`、`callbacks/` 的关键代码；
- [ ] 能列出至少 5 个 GORM 常见坑并说明原因。
---
## 18.17 本章小结
| 概念 | 一句话解释 | 易错点 |
|------|------------|--------|
| `gorm.DB` | 数据库句柄 + 链式状态 | 错误在 `Error` 字段 |
| `Statement` | 一条 SQL 的构建状态 | 链式方法都在改它 |
| `Session` | 会话级配置 | 不是数据库连接 |
| `Config` | 全局配置 | `Open` 后应视为只读 |
| 链式方法 | 返回 `*gorm.DB` | 条件会累积，注意复用 |
| `gorm.Model` | ID + 时间戳 + 软删除 | 不需要时可不嵌入 |
| 标签 | 表名、列名、索引、关联 | 字段必须导出 |
| `Create` | 插入并回填 | 必须传指针 |
| `First` | 主键升序取一条 | 查不到返回 `ErrRecordNotFound` |
| `Find` | 查询列表 | 查不到不报错 |
| `Updates` | 部分更新 | struct 零值被忽略 |
| `Delete` | 删除 | 有 `DeletedAt` 时软删除 |
| `Preload` | 预加载关联 | 不写就 N+1 |
| `Transaction` | 事务闭包 | 内部必须用 `tx` |
| 钩子 | 生命周期回调 | 返回 error 会回滚 |
| `Valuer/Scanner` | Go 类型与数据库互转 | `Scan` 用指针接收者 |
| `WithContext` | 传递取消与超时 | 传 `c.Request.Context()` |
| `ErrRecordNotFound` | 单条查询未命中 | 用 `errors.Is` 判断 |
---
## 18.18 下一章预告
第 19 章：**总复习与能力验收**。

你将会：用一张知识地图串起 18 章；完成 30 道高频综合题；做 Gin/GORM 文档阅读实战；阅读 5 段源码并说出行为；用能力自评表找出薄弱环节；制定下一步路线（Gin/GORM 教程、项目实战、测试与性能、rpc）。

学完第 18 章，你已经具备「打开 GORM 文档任意一页，独立读懂并写出示例」的能力。第 19 章要做的，是把这种能力固定下来，并验证它真的属于你。
