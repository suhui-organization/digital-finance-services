# digital-finance-services — AI 开发代理规范

## 项目身份
- **名称**: 数字金融彩票 - Go 后端 API
- **端口**: 16080
- **技术栈**: Go 1.25 + Gin + PostgreSQL + Redis + JWT
- **包管理器**: go mod

## 关键约定
- 遵循 `../.clinerules`（根 Rule）和本目录 `.clinerules`
- 严格分层: handler → service → repository → database
- 统一响应格式: `{ "code": 0, "message": "success", "data": {} }`
- 密码使用 bcrypt 加密，JWT token 设置合理过期时间
- 数据库操作使用参数化查询，防止 SQL 注入

## 快速命令
- `go run ./cmd/server` — 启动开发服务器 (端口 16080)
- `go vet ./...` — 代码静态分析
- `go build ./...` — 编译检查
- `go test ./...` — 运行测试

## 分层架构
```
handler → service → repository → database
    ↓        ↓
  (DTO)    (Model)
```
- handler: HTTP 请求/响应处理，参数验证
- service: 业务逻辑，事务管理
- repository: 数据访问，SQL 操作

## 关联服务
- PostgreSQL: localhost:15432
- Redis: localhost:16379
- AI 服务: http://localhost:16081