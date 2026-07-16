# GoDP — 本地生活点评与秒杀系统

一个使用 Go 构建的本地生活后端练习项目，覆盖用户认证、商铺、点评、点赞、优惠券秒杀与异步订单处理。项目既可以作为单体服务运行，也提供了按业务拆分的服务入口，并通过 Nginx 统一暴露 API。

> 当前定位：用于学习和演示的后端项目，不建议未经加固直接用于生产环境。

## 功能

- 用户注册、登录与个人资料查询
- bcrypt 密码哈希与 JWT 身份认证
- 商铺创建、详情查询、分类与分页列表
- 点评发布、分页查询与点赞/取消点赞
- 优惠券创建与秒杀请求
- Redis Lua 脚本原子扣减库存，数据库侧校验重复下单
- Kafka 异步投递及消费订单
- MySQL/GORM 数据持久化与自动迁移
- 单体和多服务两种启动方式
- Docker Compose 一键启动完整依赖及 API 网关
- 健康检查与 HTTP 优雅关闭

## 技术栈

- Go 1.25
- Gin
- GORM + MySQL 8
- Redis 7
- Kafka + Sarama
- JWT + bcrypt
- Nginx
- Docker / Docker Compose

## 架构概览

```text
Client
  |
  v
Nginx Gateway :8081
  |-- user-service   -> 注册 / 登录 / 用户资料
  |-- shop-service   -> 商铺
  |-- review-service -> 点评 / 点赞
  `-- order-service  -> 优惠券 / 秒杀 -> Redis -> Kafka -> MySQL

共享基础设施：MySQL、Redis、Kafka（当前各服务共享同一个数据库）
```

服务拆分目前属于“模块化单体的多进程部署”：各入口复用同一套 `internal` 代码和数据源，并不是完全独立自治的微服务。

## 目录结构

```text
.
|-- cmd/
|   |-- server/          # 单体服务入口
|   |-- user-service/    # 用户服务入口
|   |-- shop-service/    # 商铺服务入口
|   |-- review-service/  # 点评服务入口
|   |-- order-service/   # 秒杀及订单服务入口
|   `-- migrate/         # 数据库迁移入口
|-- internal/
|   |-- app/             # 应用装配、启动和优雅关闭
|   |-- config/          # 环境变量读取与校验
|   |-- database/        # MySQL、Redis、Kafka 与迁移
|   |-- handler/         # HTTP Handler 和路由
|   |-- middleware/      # JWT 中间件
|   |-- model/           # GORM 模型
|   |-- repository/      # 数据访问层
|   `-- service/         # 业务逻辑层
|-- pkg/
|   |-- crypto/          # 密码哈希
|   `-- jwt/             # Token 生成与解析
|-- docker/              # 镜像、Compose 与 Nginx 配置
|-- .env.example
|-- go.mod
`-- go.sum
```

## 快速开始（推荐）

### 前置条件

- Docker Desktop
- Docker Compose v2

### 1. 创建环境变量文件

PowerShell：

```powershell
Copy-Item .env.example .env
```

Linux/macOS：

```bash
cp .env.example .env
```

编辑 `.env`，至少替换以下值：

```dotenv
MYSQL_ROOT_PASSWORD=your-local-root-password
DB_PASSWORD=your-local-app-password
JWT_SECRET=replace-with-a-random-secret-at-least-32-characters
```

不要提交 `.env`。生产环境应使用独立的密钥管理方案。

### 2. 启动

从项目根目录执行：

```bash
docker compose --env-file .env -f docker/docker-compose.yml up --build
```

首次启动会构建服务、等待 MySQL 就绪、执行迁移，然后启动业务服务和网关。

### 3. 验证

```bash
curl http://localhost:8081/health
```

预期响应：

```json
{"status":"ok","service":"api-gateway"}
```

API 基础地址为 `http://localhost:8081`。

### 4. 停止

```bash
docker compose --env-file .env -f docker/docker-compose.yml down
```

同时删除 MySQL 和 Redis 数据卷：

```bash
docker compose --env-file .env -f docker/docker-compose.yml down -v
```

## 本地开发

本地直接运行 Go 服务时，需要先准备可访问的 MySQL、Redis 和 Kafka，并把 `.env.example` 中的容器主机名改为本机地址，例如：

```dotenv
DB_HOST=localhost
REDIS_HOST=localhost
KAFKA_BROKERS=localhost:9092
```

注意：程序不会自动读取 `.env` 文件，需要在当前 Shell 中导入环境变量，或由 IDE 运行配置注入。

先迁移数据库：

```bash
go run ./cmd/migrate
```

运行包含全部路由的单体服务：

```bash
go run ./cmd/server
```

默认监听 `http://localhost:8080`。单体入口也会初始化 Kafka 并启动订单消费者，因此 Kafka 必须可用。

也可以分别运行：

```bash
go run ./cmd/user-service
go run ./cmd/shop-service
go run ./cmd/review-service
go run ./cmd/order-service
```

分别运行时，各进程需要配置不同的 `PORT`；Docker Compose 场景中它们位于独立容器，因此都可使用 8080。

## 配置

| 变量 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `APP_ENV` | 否 | `development` | 运行环境；`production` 会启用额外配置校验 |
| `SERVICE_ROLE` | 否 | 空 | 路由角色：`user`、`shop`、`review`、`order` |
| `PORT` | 否 | `8080` | HTTP 监听端口 |
| `DB_HOST` | 否 | `localhost` | MySQL 地址 |
| `DB_PORT` | 否 | `3306` | MySQL 端口 |
| `DB_NAME` | 否 | `ai_review_system` | 数据库名 |
| `DB_USER` | 是 | 无 | 数据库用户 |
| `DB_PASSWORD` | 是 | 无 | 数据库密码 |
| `REDIS_HOST` | 否 | `localhost` | Redis 地址 |
| `REDIS_PORT` | 否 | `6379` | Redis 端口 |
| `KAFKA_BROKERS` | 否 | `localhost:9092` | Broker 列表，多个地址用逗号分隔 |
| `JWT_SECRET` | 是 | 无 | JWT 密钥，至少 32 个字符 |

## API

除公开接口外，请携带：

```http
Authorization: Bearer <token>
Content-Type: application/json
```

| 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| `GET` | `/health` | 否 | 健康检查 |
| `GET` | `/ping` | 否 | 应用存活检查 |
| `POST` | `/register` | 否 | 注册用户 |
| `POST` | `/login` | 否 | 登录并获取 JWT |
| `GET` | `/user/profile` | 是 | 查询当前用户 |
| `GET` | `/shops` | 否 | 商铺分页列表，可传 `page`、`page_size`、`category` |
| `GET` | `/shops/:id` | 否 | 商铺详情 |
| `POST` | `/shops` | 是 | 创建商铺 |
| `GET` | `/reviews/:shop_id` | 否 | 查询商铺点评，可传 `page`、`page_size` |
| `POST` | `/reviews` | 是 | 发布点评 |
| `POST` | `/reviews/:id/like` | 是 | 点赞或取消点赞 |
| `POST` | `/coupons` | 是 | 创建优惠券 |
| `POST` | `/seckill/:coupon_id` | 是 | 提交秒杀请求，成功返回 202 |

### 调用示例

注册：

```bash
curl -X POST http://localhost:8081/register \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","email":"alice@example.com","password":"secret123"}'
```

登录：

```bash
curl -X POST http://localhost:8081/login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"secret123"}'
```

创建商铺：

```bash
curl -X POST http://localhost:8081/shops \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Demo Cafe","category":"cafe","address":"No. 1 Demo Road","description":"A demo shop","avg_price":35}'
```

创建点评：

```bash
curl -X POST http://localhost:8081/reviews \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"shop_id":1,"content":"Good experience","score":5}'
```

创建优惠券时，`begin_time` 与 `end_time` 使用 Unix 秒级时间戳：

```bash
curl -X POST http://localhost:8081/coupons \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"50% off","stock":100,"discount":0.5,"begin_time":1767225600,"end_time":1893456000}'
```

发起秒杀：

```bash
curl -X POST http://localhost:8081/seckill/1 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 测试

```bash
go test ./...
```

如果所在网络无法访问 `proxy.golang.org`，可先配置可用的 Go 模块代理，再重试依赖下载和测试。例如：

```bash
go env -w GOPROXY=https://goproxy.cn,direct
```

是否使用第三方代理应由你所在组织的安全策略决定。

## 当前限制与改进方向

- 商铺、点评等 Handler 在请求内直接创建 Service，依赖注入尚未统一，测试替身不易接入。
- 各服务共享数据库与内部代码，服务边界偏部署层，距离真正的微服务自治仍有差距。
- 创建商铺和优惠券仅要求登录，没有管理员或商户角色授权。
- API 错误格式和错误码尚未统一，部分内部错误会直接返回给客户端。
- 缺少请求限流、熔断、链路追踪、结构化日志与 Prometheus 指标。
- 自动化测试覆盖较低，秒杀并发、Redis/Kafka 故障与幂等场景尤其需要集成测试。
- 分页参数缺少严格的边界校验；图片字段仍是字符串，缺少独立资源模型。
- Kafka 使用 ZooKeeper 模式，后续可考虑迁移到 KRaft，并完善重试、死信队列和消费幂等。
- Compose 中间件端口直接暴露到宿主机，生产部署应收紧网络、凭据与访问控制。
- 仓库根目录存在若干手工请求 JSON，建议迁移为自动化 API/集成测试数据。

## License

当前仓库未声明开源许可证。在添加明确的 `LICENSE` 前，请勿默认将其视为可自由再分发的软件。
