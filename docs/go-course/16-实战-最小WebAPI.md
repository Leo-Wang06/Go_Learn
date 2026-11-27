# 第 16 章：实战 —— 最小 Web API

> **阶段**：阶段四 · 测试、工程化与实战
> **建议时间**：2 天
> **前置知识**：第 09–15 章（模块、`context`、`net/http`、`database/sql`、测试、工程化）
> **代码基准**：Go 1.22 及以上；SQLite 驱动使用纯 Go 的 `modernc.org/sqlite`

---

## 本章目标

学完本章，你应当能够：

1. 把需求拆成 RESTful 接口；
2. 用 `net/http` 的 Go 1.22 路由写出可运行 API；
3. 用 `database/sql` 完成增删查改；
4. 按 handler / service / repository 分层；
5. 写日志、恢复、CORS、TraceID 中间件；
6. 用统一格式返回错误；
7. 把 `context` 从请求传到数据库，并实现优雅关闭；
8. 用 `httptest` + 表驱动测试覆盖 handler；
9. 看懂同样功能的 Gin/GORM 写法。

---

## 16.1 需求与 RESTful 接口设计

### 16.1.1 接口

| 方法 | 路径 | 含义 | 成功状态码 |
|------|------|------|------------|
| `POST` | `/api/v1/users` | 创建用户 | `201 Created` |
| `GET` | `/api/v1/users/{id}` | 查询单个用户 | `200 OK` |
| `GET` | `/api/v1/users?limit=20&offset=0` | 分页列表 | `200 OK` |

REST 的核心是「用名词表示资源，用 HTTP 方法表示动作」。不要写 `/getUser?id=1`、`/createUser`。

### 16.1.2 状态码与响应格式

| 场景 | 状态码 |
|------|--------|
| 创建成功 | `201` |
| 查询/列表成功 | `200` |
| JSON 格式错误、参数非法、校验失败 | `400` |
| 资源不存在 | `404` |
| 邮箱冲突 | `409` |
| 未预期错误 | `500` |

成功时直接返回资源 JSON；失败时统一为：

```json
{"error":{"code":"VALIDATION_ERROR","message":"name: name is required","trace_id":"..."}}
```

`code` 供程序判断，`message` 供人阅读，`trace_id` 用于查日志。不要把 SQL 错误或堆栈暴露给客户端。

### 16.1.3 与 Gin 的预告

Gin 会简化路由、参数绑定和响应渲染，但不会替你决定状态码、错误格式和分层。先写好 `net/http` 版本，第 17 章再看 Gin 会轻松很多。

---

## 16.2 项目初始化与目录结构

```bash
mkdir mini-api && cd mini-api
go mod init example.com/miniapi
go get modernc.org/sqlite@v1.29.10
```

选择 `modernc.org/sqlite` 是因为它纯 Go 实现，`CGO_ENABLED=0` 也能编译和交叉编译。

```
mini-api/
├── cmd/api/main.go
├── internal/{database,model,repository,service,httpapi}/
└── migrations/{embed.go,001_init.sql}
```

分层与请求流转：handler 解析 JSON、取参数、映射状态码；service 校验、规范化、业务规则；repository 执行 SQL、扫描、翻译唯一键错误；最底层是 `database/sql`。依赖方向只能向下。repository 不知道 HTTP，service 不知道 `http.ResponseWriter`，handler 不写 SQL。

与 TS/Express 对照：

| Express | 本章 |
|---------|------|
| `app.post('/users', handler)` | `mux.HandleFunc("POST /api/v1/users", h)` |
| `req.params.id` | `r.PathValue("id")` |
| `res.status(201).json(user)` | `writeJSON(w, 201, user)` |
| `next(err)` | 返回 `error`，由 handler 统一映射 |

---

## 16.3 领域模型与错误

`internal/model/user.go`：

```go
package model

import (
    "errors"
    "fmt"
    "time"
)

var (
    ErrNotFound = errors.New("user not found")
    ErrConflict = errors.New("email already exists")
)

type User struct {
    ID        int64     `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}

type CreateUserInput struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Message)
}
```

`json` 标签决定 API 字段名；`ErrNotFound` / `ErrConflict` 供 `errors.Is`；`ValidationError` 供 `errors.As`。

---

## 16.4 数据库接入

### 16.4.1 连接池

`internal/database/database.go`：

```go
package database

import (
    "context"
    "database/sql"
    "fmt"
    "time"

    _ "modernc.org/sqlite"
)

func Open(dsn string) (*sql.DB, error) {
    db, err := sql.Open("sqlite", dsn)
    if err != nil {
        return nil, fmt.Errorf("open database: %w", err)
    }

    // SQLite 同一时刻只允许一个写事务，限制单连接避免 SQLITE_BUSY。
    db.SetMaxOpenConns(1)
    db.SetMaxIdleConns(1)
    db.SetConnMaxLifetime(0)

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := db.PingContext(ctx); err != nil {
        db.Close()
        return nil, fmt.Errorf("ping database: %w", err)
    }
    return db, nil
}
```

`sql.Open` 不会真正连接，`PingContext` 才会。连接池参数必须配置。

### 16.4.2 迁移

`migrations/embed.go`：

```go
// Package migrations embeds SQL migration files.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
```

`migrations/001_init.sql`：

```sql
CREATE TABLE IF NOT EXISTS users (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT    NOT NULL,
    email      TEXT    NOT NULL UNIQUE,
    created_at INTEGER NOT NULL
);
```

`created_at` 存 Unix 纳秒，读写都转成 `time.Time`，避免驱动对时间格式的差异。

`internal/database/migrate.go`：

```go
package database

import (
    "context"
    "database/sql"
    "fmt"
    "io/fs"
    "sort"

    "example.com/miniapi/migrations"
)

func Migrate(ctx context.Context, db *sql.DB) error {
    names, err := fs.Glob(migrations.FS, "*.sql")
    if err != nil {
        return err
    }
    sort.Strings(names)
    for _, name := range names {
        script, err := migrations.FS.ReadFile(name)
        if err != nil {
            return err
        }
        if _, err := db.ExecContext(ctx, string(script)); err != nil {
            return fmt.Errorf("apply %s: %w", name, err)
        }
    }
    return nil
}
```

这是「启动时执行幂等 SQL」的最小做法。生产项目应使用版本表或 golang-migrate/goose。

### 16.4.3 MySQL 差异

换 MySQL 时驱动名改为 `mysql`，DSN 加 `parseTime=true`，唯一键冲突错误码是 `1062`，连接池可以开多个连接。

---

## 16.5 repository：SQL 与扫描

`internal/repository/user_repository.go`：

```go
package repository

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    "strings"
    "time"

    "example.com/miniapi/internal/model"
)

type UserRepository struct {
    db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
    return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
    result, err := r.db.ExecContext(ctx,
        `INSERT INTO users (name, email, created_at) VALUES (?, ?, ?)`,
        user.Name, user.Email, user.CreatedAt.UnixNano(),
    )
    if err != nil {
        if isUniqueViolation(err) {
            return model.ErrConflict
        }
        return fmt.Errorf("insert user: %w", err)
    }
    id, err := result.LastInsertId()
    if err != nil {
        return fmt.Errorf("read last insert id: %w", err)
    }
    user.ID = id
    return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
    user, err := scanUser(r.db.QueryRowContext(ctx,
        `SELECT id, name, email, created_at FROM users WHERE id = ?`, id))
    if errors.Is(err, sql.ErrNoRows) {
        return nil, model.ErrNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("query user %d: %w", id, err)
    }
    return user, nil
}

func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]model.User, error) {
    rows, err := r.db.QueryContext(ctx,
        `SELECT id, name, email, created_at FROM users ORDER BY id LIMIT ? OFFSET ?`,
        limit, offset,
    )
    if err != nil {
        return nil, fmt.Errorf("list users: %w", err)
    }
    defer rows.Close()

    users := make([]model.User, 0, limit)
    for rows.Next() {
        user, err := scanUser(rows)
        if err != nil {
            return nil, fmt.Errorf("scan user: %w", err)
        }
        users = append(users, *user)
    }
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("iterate users: %w", err)
    }
    return users, nil
}

func scanUser(row interface{ Scan(...any) error }) (*model.User, error) {
    var (
        user      model.User
        createdNS int64
    )
    if err := row.Scan(&user.ID, &user.Name, &user.Email, &createdNS); err != nil {
        return nil, err
    }
    user.CreatedAt = time.Unix(0, createdNS).UTC()
    return &user, nil
}

func isUniqueViolation(err error) bool {
    msg := err.Error()
    return strings.Contains(msg, "UNIQUE constraint failed") ||
        strings.Contains(msg, "Duplicate entry") ||
        strings.Contains(msg, "constraint failed: UNIQUE")
}
```

关键点：全部用 `?` 占位符；所有调用带 `ctx`；`sql.ErrNoRows` 单独判断；`defer rows.Close()` 并检查 `rows.Err()`；空结果用 `make([]T, 0)`，避免 JSON 输出 `null`；唯一键冲突翻译成领域错误。

---

## 16.6 service：业务规则

`internal/service/user_service.go`：

```go
package service

import (
    "context"
    "errors"
    "fmt"
    "net/mail"
    "strings"
    "time"

    "example.com/miniapi/internal/model"
)

type UserRepository interface {
    Create(ctx context.Context, user *model.User) error
    GetByID(ctx context.Context, id int64) (*model.User, error)
    List(ctx context.Context, limit, offset int) ([]model.User, error)
}

type UserService struct {
    repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
    return &UserService{repo: repo}
}

func (s *UserService) Create(ctx context.Context, in model.CreateUserInput) (*model.User, error) {
    name := strings.TrimSpace(in.Name)
    if name == "" {
        return nil, &model.ValidationError{Field: "name", Message: "name is required"}
    }

    email := strings.ToLower(strings.TrimSpace(in.Email))
    if email == "" {
        return nil, &model.ValidationError{Field: "email", Message: "email is required"}
    }
    addr, err := mail.ParseAddress(email)
    if err != nil || addr.Address != email {
        return nil, &model.ValidationError{Field: "email", Message: "email is invalid"}
    }

    user := &model.User{Name: name, Email: email, CreatedAt: time.Now().UTC()}
    if err := s.repo.Create(ctx, user); err != nil {
        if errors.Is(err, model.ErrConflict) {
            return nil, model.ErrConflict
        }
        return nil, fmt.Errorf("create user: %w", err)
    }
    return user, nil
}

func (s *UserService) GetByID(ctx context.Context, id int64) (*model.User, error) {
    if id <= 0 {
        return nil, &model.ValidationError{Field: "id", Message: "id must be positive"}
    }
    return s.repo.GetByID(ctx, id)
}

func (s *UserService) List(ctx context.Context, limit, offset int) ([]model.User, error) {
    if limit <= 0 {
        limit = 20
    }
    if limit > 100 {
        limit = 100
    }
    if offset < 0 {
        offset = 0
    }
    return s.repo.List(ctx, limit, offset)
}
```

接口定义在 service（使用方），repository 隐式实现；校验和规范化属于业务规则；service 返回领域错误，不返回 HTTP 状态码。

---

## 16.7 HTTP 层：路由、Handler、JSON

`internal/httpapi/handler.go`：

```go
package httpapi

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "log/slog"
    "net/http"
    "strconv"

    "example.com/miniapi/internal/model"
)

type UserService interface {
    Create(ctx context.Context, in model.CreateUserInput) (*model.User, error)
    GetByID(ctx context.Context, id int64) (*model.User, error)
    List(ctx context.Context, limit, offset int) ([]model.User, error)
}

type Handler struct {
    users  UserService
    logger *slog.Logger
}

func NewHandler(users UserService, logger *slog.Logger) *Handler {
    return &Handler{users: users, logger: logger}
}

func (h *Handler) Routes() http.Handler {
    mux := http.NewServeMux()
    mux.HandleFunc("POST /api/v1/users", h.createUser)
    mux.HandleFunc("GET /api/v1/users", h.listUsers)
    mux.HandleFunc("GET /api/v1/users/{id}", h.getUser)
    return mux
}

func (h *Handler) createUser(w http.ResponseWriter, r *http.Request) {
    var in model.CreateUserInput
    if err := decodeJSON(w, r, &in); err != nil {
        writeError(w, r, http.StatusBadRequest, "INVALID_JSON", err.Error())
        return
    }

    user, err := h.users.Create(r.Context(), in)
    if err != nil {
        h.writeServiceError(w, r, err)
        return
    }
    w.Header().Set("Location", fmt.Sprintf("/api/v1/users/%d", user.ID))
    writeJSON(w, http.StatusCreated, user)
}

func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
    id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
    if err != nil || id <= 0 {
        writeError(w, r, http.StatusBadRequest, "INVALID_ID", "id must be a positive integer")
        return
    }

    user, err := h.users.GetByID(r.Context(), id)
    if err != nil {
        h.writeServiceError(w, r, err)
        return
    }
    writeJSON(w, http.StatusOK, user)
}

func (h *Handler) listUsers(w http.ResponseWriter, r *http.Request) {
    q := r.URL.Query()
    limit, err := strconv.Atoi(q.Get("limit"))
    if err != nil && q.Get("limit") != "" {
        writeError(w, r, http.StatusBadRequest, "INVALID_QUERY", "limit must be an integer")
        return
    }
    if limit <= 0 {
        limit = 20
    }
    offset, _ := strconv.Atoi(q.Get("offset"))
    if offset < 0 {
        offset = 0
    }

    users, err := h.users.List(r.Context(), limit, offset)
    if err != nil {
        h.writeServiceError(w, r, err)
        return
    }
    writeJSON(w, http.StatusOK, map[string]any{
        "items": users, "limit": limit, "offset": offset,
    })
}

func (h *Handler) writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
    var ve *model.ValidationError
    switch {
    case errors.As(err, &ve):
        writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", ve.Error())
    case errors.Is(err, model.ErrNotFound):
        writeError(w, r, http.StatusNotFound, "NOT_FOUND", "user not found")
    case errors.Is(err, model.ErrConflict):
        writeError(w, r, http.StatusConflict, "CONFLICT", "email already exists")
    default:
        h.logger.ErrorContext(r.Context(), "request failed",
            "error", err, "trace_id", TraceIDFrom(r.Context()),
            "method", r.Method, "path", r.URL.Path,
        )
        writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
    }
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
    r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

    dec := json.NewDecoder(r.Body)
    dec.DisallowUnknownFields()
    if err := dec.Decode(dst); err != nil {
        return fmt.Errorf("invalid JSON body: %w", err)
    }
    if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
        return errors.New("body must contain a single JSON object")
    }
    return nil
}

type errorDetail struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    TraceID string `json:"trace_id,omitempty"`
}

type errorResponse struct {
    Error errorDetail `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
    writeJSON(w, status, errorResponse{
        Error: errorDetail{Code: code, Message: message, TraceID: TraceIDFrom(r.Context())},
    })
}
```

要点：`ServeMux` 的 `"METHOD /path/{id}"` 配合 `r.PathValue("id")`；`r.Context()` 必须传给 service/repository；请求体限 1 MiB，`DisallowUnknownFields` 让拼错字段报错；错误映射只在 handler 做一次；`WriteHeader` 只能调用一次。

---

## 16.8 中间件：日志、恢复、CORS、TraceID

`internal/httpapi/middleware.go`：

```go
package httpapi

import (
    "context"
    "log/slog"
    "net/http"
    "runtime/debug"
    "strconv"
    "time"
)

type contextKey int

const traceIDKey contextKey = iota

func WithTraceID(ctx context.Context, id string) context.Context {
    return context.WithValue(ctx, traceIDKey, id)
}

func TraceIDFrom(ctx context.Context) string {
    id, _ := ctx.Value(traceIDKey).(string)
    return id
}

func Chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
    for i := len(middlewares) - 1; i >= 0; i-- {
        h = middlewares[i](h)
    }
    return h
}

func RequestID() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            id := r.Header.Get("X-Request-ID")
            if id == "" {
                id = strconv.FormatInt(time.Now().UnixNano(), 10)
            }
            w.Header().Set("X-Request-ID", id)
            next.ServeHTTP(w, r.WithContext(WithTraceID(r.Context(), id)))
        })
    }
}

type responseWriter struct {
    http.ResponseWriter
    status      int
    wroteHeader bool
}

func (w *responseWriter) WriteHeader(code int) {
    if w.wroteHeader {
        return
    }
    w.wroteHeader = true
    w.status = code
    w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriter) Write(p []byte) (int, error) {
    if !w.wroteHeader {
        w.WriteHeader(http.StatusOK)
    }
    return w.ResponseWriter.Write(p)
}

func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
            next.ServeHTTP(rw, r)

            logger.InfoContext(r.Context(), "http request",
                "trace_id", TraceIDFrom(r.Context()),
                "method", r.Method, "path", r.URL.Path,
                "status", rw.status,
                "duration_ms", time.Since(start).Milliseconds(),
            )
        })
    }
}

func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            defer func() {
                if rec := recover(); rec != nil {
                    logger.ErrorContext(r.Context(), "panic recovered",
                        "trace_id", TraceIDFrom(r.Context()),
                        "panic", rec, "stack", string(debug.Stack()),
                    )
                    if rw, ok := w.(interface{ Written() bool }); ok && rw.Written() {
                        return
                    }
                    writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
                }
            }()
            next.ServeHTTP(w, r)
        })
    }
}

func CORS(origin string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("Access-Control-Allow-Origin", origin)
            if r.Method == http.MethodOptions {
                w.WriteHeader(http.StatusNoContent)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

要点：中间件签名是 `func(http.Handler) http.Handler`；`Chain` 从后往前包裹，书写顺序就是执行顺序；`responseWriter` 记录状态码；Recovery 必须在日志中间件内层；`CORS("*")` 只适合公开 API。

---

## 16.9 main：组装与优雅关闭

`cmd/api/main.go`：

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "log/slog"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "example.com/miniapi/internal/database"
    "example.com/miniapi/internal/httpapi"
    "example.com/miniapi/internal/repository"
    "example.com/miniapi/internal/service"
)

func main() {
    if err := run(); err != nil {
        slog.Error("fatal", "error", err)
        os.Exit(1)
    }
}

func run() error {
    var level slog.Level
    if err := level.UnmarshalText([]byte(env("APP_LOG_LEVEL", "INFO"))); err != nil {
        return fmt.Errorf("APP_LOG_LEVEL: %w", err)
    }
    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
    slog.SetDefault(logger)

    shutdownTimeout := 10 * time.Second
    if raw := os.Getenv("APP_SHUTDOWN_TIMEOUT"); raw != "" {
        d, err := time.ParseDuration(raw)
        if err != nil {
            return fmt.Errorf("APP_SHUTDOWN_TIMEOUT: %w", err)
        }
        shutdownTimeout = d
    }

    db, err := database.Open(env("APP_DSN", "file:miniapi.db"))
    if err != nil {
        return err
    }
    defer db.Close()

    if err := database.Migrate(context.Background(), db); err != nil {
        return err
    }

    repo := repository.NewUserRepository(db)
    svc := service.NewUserService(repo)
    handler := httpapi.NewHandler(svc, logger)

    routes := httpapi.Chain(
        handler.Routes(),
        httpapi.RequestID(),
        httpapi.Logging(logger),
        httpapi.Recover(logger),
        httpapi.CORS("*"),
    )

    srv := &http.Server{
        Addr:              env("APP_ADDR", ":8080"),
        Handler:           routes,
        ReadHeaderTimeout: 5 * time.Second,
        ReadTimeout:       15 * time.Second,
        WriteTimeout:      15 * time.Second,
        IdleTimeout:       60 * time.Second,
    }

    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    errCh := make(chan error, 1)
    go func() {
        logger.Info("server listening", "addr", srv.Addr)
        if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
            errCh <- err
        }
    }()

    select {
    case err := <-errCh:
        return fmt.Errorf("server: %w", err)
    case <-ctx.Done():
        logger.Info("shutdown signal received")
    }

    shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
    defer cancel()
    if err := srv.Shutdown(shutdownCtx); err != nil {
        return fmt.Errorf("shutdown: %w", err)
    }

    logger.Info("server stopped")
    return nil
}

func env(key, fallback string) string {
    if value, ok := os.LookupEnv(key); ok {
        return value
    }
    return fallback
}
```

组装顺序：配置 → 日志 → 数据库 → repository → service → handler → 路由 → `http.Server`。这就是「main 是组装车间」。

启动 `go run ./cmd/api` 后，用 `curl` 验证：

```bash
curl -i -X POST http://127.0.0.1:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"Alice@Example.com"}'
```

成功返回 `201`、`Location: /api/v1/users/1` 和用户 JSON；重复邮箱返回 `409 CONFLICT`，空姓名返回 `400 VALIDATION_ERROR`，不存在的 ID 返回 `404 NOT_FOUND`，非法 ID 返回 `400 INVALID_ID`。

---

## 16.10 测试：httptest + 表驱动

Handler 依赖 `UserService` 接口，测试时用 stub 替换，不需要数据库。

`internal/httpapi/handler_test.go`：

```go
package httpapi_test

import (
    "context"
    "io"
    "log/slog"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"

    "example.com/miniapi/internal/httpapi"
    "example.com/miniapi/internal/model"
)

type stubService struct {
    create func(context.Context, model.CreateUserInput) (*model.User, error)
}

func (s *stubService) Create(ctx context.Context, in model.CreateUserInput) (*model.User, error) {
    return s.create(ctx, in)
}
func (s *stubService) GetByID(context.Context, int64) (*model.User, error) { panic("not expected") }
func (s *stubService) List(context.Context, int, int) ([]model.User, error) { panic("not expected") }

func TestCreateUser(t *testing.T) {
    tests := []struct {
        name       string
        body       string
        svc        *stubService
        wantStatus int
        wantCode   string
    }{
        {
            name: "created",
            body: `{"name":"Alice","email":"alice@example.com"}`,
            svc: &stubService{create: func(_ context.Context, in model.CreateUserInput) (*model.User, error) {
                return &model.User{ID: 1, Name: in.Name, Email: in.Email}, nil
            }},
            wantStatus: http.StatusCreated,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            logger := slog.New(slog.NewTextHandler(io.Discard, nil))
            handler := httpapi.NewHandler(tt.svc, logger)

            req := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(tt.body))
            req.Header.Set("Content-Type", "application/json")
            rec := httptest.NewRecorder()
            handler.Routes().ServeHTTP(rec, req)

            if rec.Code != tt.wantStatus {
                t.Fatalf("status = %d, want %d; body = %s", rec.Code, tt.wantStatus, rec.Body.String())
            }
            if tt.wantCode != "" && !strings.Contains(rec.Body.String(), tt.wantCode) {
                t.Fatalf("body = %s, want code %s", rec.Body.String(), tt.wantCode)
            }
        })
    }
}
```

运行：

```bash
go test ./...
go test -race ./...
```

补充：service 用 fake repository 测校验和错误翻译；repository 用 `:memory:` SQLite 测增查和唯一冲突；`httptest.NewServer` 适合测完整 HTTP 客户端行为，`NewRecorder` 适合测 handler。

---

## 16.11 与 Gin/GORM 的等价写法

Gin 的 `r.POST("/users", h)`、`r.GET("/users/:id", h)` 对应本章的 `mux.HandleFunc`；`c.Param("id")` 对应 `r.PathValue("id")`；`c.Query`/`c.DefaultQuery` 对应 `r.URL.Query().Get`；`ShouldBindJSON` 对应 `json.Decoder`；`c.JSON` 对应 `writeJSON`。中间件从 `func(http.Handler) http.Handler` 变成 `func(*gin.Context)`，用 `c.Next()` 控制顺序。

GORM 对应写法是 `db.WithContext(ctx).Create(&user)`、`First(&user, id)`、`Limit(limit).Offset(offset).Find(&users)`。`errors.Is(err, gorm.ErrRecordNotFound)` 映射 404；开启 `TranslateError: true` 后，唯一键冲突可用 `gorm.ErrDuplicatedKey`。Gin/GORM 替代路由、绑定、渲染和 SQL 样板，但不会替你分层或决定错误格式。

---

## 16.12 常见错误与陷阱

| 错误/现象 | 原因 | 修复 |
|-----------|------|------|
| `superfluous response.WriteHeader` | 多次写状态码 | 只调用一次 |
| 客户端收到 `null` 而不是 `[]` | 空切片是 `nil` | `make([]T, 0)` |
| 查询卡住 | SQLite 单连接下 rows 未关闭 | 先 `rows.Close()` |
| `:memory:` 测试看不到数据 | 每个连接一个内存库 | `SetMaxOpenConns(1)` |
| `sql.ErrNoRows` 被当成 500 | 没有单独判断 | 返回 `ErrNotFound` |
| 请求取消后数据库还在跑 | 用了 `Background` | 传 `r.Context()` |

---

## 16.13 练习题

### 基础题

1. 按本章结构创建项目，跑通 Create/Get/List。
2. 用 `curl` 验证 201、400、404、409。
3. 给列表加 `limit` 上限和 `offset` 校验，用环境变量配置地址、DSN、日志级别。
4. 跑通 `go test -race ./...`。

### 进阶题

5. 增加 `PATCH` 和 `DELETE /api/v1/users/{id}`。
6. 用事务保证创建用户和写审计日志原子化。
7. 加超时中间件，超过 3 秒返回 504。
8. 用 `httptest.NewServer` 写端到端测试，用 fake repository 测 service。
9. 把 SQLite 换成 MySQL，处理驱动、DSN、唯一键差异；用 `pprof` 看一次列表请求。

### 读文档题

10. 执行 `go doc net/http.ServeMux`、`go doc database/sql.DB.QueryContext`、`go doc database/sql.ErrNoRows`，回答方法模式、rows 关闭责任和 `ErrNoRows` 的时机。
11. 查阅 `modernc.org/sqlite` 文档，回答驱动名和 DSN。
12. 查阅 Gin 的 `ShouldBindJSON`，回答失败时是否自动写 400。

---

## 16.14 验收清单

- [ ] 项目可运行，Create/Get/List 可用，状态码和错误格式符合约定；
- [ ] handler / service / repository 分层清晰；
- [ ] 所有数据库调用都带 `context`，有日志、恢复、CORS、TraceID 中间件并实现优雅关闭；
- [ ] handler 有表驱动测试，能解释每层职责和 Gin/GORM 的简化点；
- [ ] `go test -race ./...` 通过。

---

## 16.15 本章小结

| 主题 | 一句话解释 | 易错点 |
|------|------------|--------|
| REST | 名词资源 + HTTP 方法 | 不要动词 URL |
| 分层 | handler → service → repository | 依赖不能反向 |
| `database/sql` | 连接池 + Context + 占位符 | 忘关 rows、忘查 `ErrNoRows` |
| 统一错误 | code + message + trace_id | 不暴露内部细节 |
| 测试/关闭 | `httptest` + 表驱动 + stub；Shutdown + 超时 | 依赖真实数据库、先关库 |

---

## 16.16 下一章预告
第 17 章：**如何读懂 Gin 文档与源码**。你会把本章的 `net/http` 版本逐项映射到 Gin 的类型、路由、参数、响应和中间件，并学会顺着 `gin.go`、`routergroup.go`、`context.go` 读源码。
