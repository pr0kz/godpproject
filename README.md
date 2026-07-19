# GoDP — 本地生活点评与秒杀微服务

GoDP 是一个使用 Go 构建的本地生活后端项目，包含用户、商铺、点评和秒杀订单四个独立服务。每个服务拥有自己的领域模型、仓储、业务逻辑、HTTP 接口、数据库和迁移，不允许通过其他服务的数据库访问跨领域数据。

> 当前定位：微服务架构学习与演示项目。上线生产环境前仍需补充权限模型、可观测性、限流熔断、密钥管理和完整集成测试。

## 功能

- 用户注册、登录和个人资料查询
- bcrypt 密码哈希与 JWT 身份认证
- 商铺创建、详情查询、分类和分页列表
- Redis 商铺缓存、空值缓存和缓存击穿保护
- 点评发布、分页查询、点赞与取消点赞
- Review Service 通过 Kafka 发布点评领域事件
- 优惠券创建和秒杀请求
- Redis Lua 原子扣减秒杀库存
- Kafka 异步订单消息
- 数据库唯一索引保障一人一单
- 四个服务独立数据库、配置、迁移和容器镜像
- Nginx API Gateway 统一路由
- HTTP 优雅关闭

## 技术栈

- Go 1.25
- Gin
- GORM + MySQL 8
- Redis 7
- Kafka + Sarama
- JWT + bcrypt
- Nginx
- Docker / Docker Compose

## 架构

```text
                         +----------------+      +---------+
Client -> Nginx Gateway ->| User Service   |----->| user_db |
                         +----------------+      +---------+

                         +----------------+      +---------+
Client -> Nginx Gateway ->| Shop Service   |----->| shop_db |
                         |                |-----> Redis
                         +----------------+      +---------+

                         +----------------+      +-----------+
Client -> Nginx Gateway ->| Review Service |----->| review_db |
                         |                |-----> Kafka
                         +----------------+      +-----------+

                         +----------------+      +----------+
Client -> Nginx Gateway ->| Order Service  |----->| order_db |
                         |                |-----> Redis
                         |                |-----> Kafka
                         +----------------+      +----------+
```

### 服务边界

- **User Service**：注册、登录、用户资料、密码哈希和 JWT 签发，只访问 `user_db`。
- **Shop Service**：商铺管理和商铺缓存，只访问 `shop_db` 及 Shop Redis。
- **Review Service**：点评与点赞，只访问 `review_db`，通过 Kafka 发布 `review.created`、`review.liked` 和 `review.unliked` 事件。
- **Order Service**：优惠券、秒杀、订单消息和一人一单校验，只访问 `order_db`、Order Redis 和 Kafka。
- **API Gateway**：只负责反向代理和路由，不承载业务逻辑。

服务之间不共享业务模型、Repository 或 Service，也不能查询或修改其他服务的数据库。

## 目录结构

```text
.
├── services/
│   ├── user/
│   │   ├── cmd/main.go
│   │   ├── internal/
│   │   │   ├── handler/
│   │   │   ├── service/
│   │   │   ├── repository/
│   │   │   ├── model/
│   │   │   └── migrate/
│   │   └── Dockerfile
│   ├── shop/
│   │   ├── cmd/main.go
│   │   ├── internal/
│   │   │   ├── handler/
│   │   │   ├── service/
│   │   │   ├── repository/
│   │   │   ├── model/
│   │   │   └── migrate/
│   │   └── Dockerfile
│   ├── review/
│   │   ├── cmd/main.go
│   │   ├── internal/
│   │   │   ├── handler/
│   │   │   ├── service/
│   │   │   ├── repository/
│   │   │   ├── model/
│   │   │   ├── consumer/
│   │   │   └── migrate/
│   │   └── Dockerfile
│   └── order/
│       ├── cmd/main.go
│       ├── internal/
│       │   ├── handler/
│       │   ├── service/
│       │   ├── repository/
│       │   ├── model/
│       │   ├── consumer/
│       │   └── migrate/
│       └── Dockerfile
├── pkg/
│   ├── auth/          # JWT HTTP 中间件
│   ├── config/        # 服务配置加载与校验
│   ├── crypto/        # 密码哈希
│   ├── database/      # 通用数据库连接
│   ├── httpx/         # HTTP Server 和优雅关闭
│   ├── jwt/           # Token 签发与验证
│   └── messaging/     # Kafka 基础设施
├── api/
│   ├── events/        # 跨服务事件协议
│   └── contracts/     # 同步通信协议预留目录
├── docker/
│   ├── docker-compose.yml
│   └── nginx.conf
├── go.mod
└── go.sum
```

`pkg` 只包含通用基础设施，不包含任何业务模型、业务仓储或业务服务。

## 快速开始

### 前置条件

- Go 1.25（本地运行时）
- Docker Desktop
- Docker Compose v2

### 环境变量

Docker Compose 至少需要以下变量：

```dotenv
MYSQL_ROOT_PASSWORD=replace-with-root-password
USER_DB_PASSWORD=replace-with-user-db-password
SHOP_DB_PASSWORD=replace-with-shop-db-password
REVIEW_DB_PASSWORD=replace-with-review-db-password
ORDER_DB_PASSWORD=replace-with-order-db-password
JWT_SECRET=replace-with-a-random-secret-at-least-32-characters
```

PowerShell 示例：

```powershell
$env:MYSQL_ROOT_PASSWORD = "local-root-password"
$env:USER_DB_PASSWORD = "local-user-password"
$env:SHOP_DB_PASSWORD = "local-shop-password"
$env:REVIEW_DB_PASSWORD = "local-review-password"
$env:ORDER_DB_PASSWORD = "local-order-password"
$env:JWT_SECRET = "replace-with-at-least-32-random-characters"
```

不要提交真实密码、JWT 密钥或 `.env` 文件。生产环境应使用专用密钥管理系统。

### 启动容器

```bash
docker compose -f docker/docker-compose.yml up --build
```

Gateway 默认监听：

```text
http://localhost:8081
```

验证 Gateway：

```bash
curl http://localhost:8081/health
```

停止服务：

```bash
docker compose -f docker/docker-compose.yml down
```

同时删除数据库和 Redis 数据卷：

```bash
docker compose -f docker/docker-compose.yml down -v
```

## 本地运行

服务不会自动读取 `.env`，需要通过 Shell 或 IDE 运行配置注入环境变量。

每个服务支持通用变量，也支持以服务名为前缀的变量。服务专属变量优先级更高，例如 `USER_DB_HOST` 优先于 `DB_HOST`。

### User Service

所需资源：MySQL、JWT。

```bash
go run ./services/user/cmd
```

默认数据库名为 `user_db`。

### Shop Service

所需资源：MySQL、Redis、Kafka 配置和 JWT。

```bash
go run ./services/shop/cmd
```

默认数据库名为 `shop_db`。

### Review Service

所需资源：MySQL、Kafka Producer 和 JWT。

```bash
go run ./services/review/cmd
```

默认数据库名为 `review_db`。

### Order Service

所需资源：MySQL、Redis、Kafka Producer 和 JWT。

```bash
go run ./services/order/cmd
```

默认数据库名为 `order_db`。

本地同时运行多个服务时，应为每个服务设置不同端口，例如：

```dotenv
USER_PORT=8082
SHOP_PORT=8083
REVIEW_PORT=8084
ORDER_PORT=8085
```

## 配置

所有服务支持以下数据库变量：

- `DB_HOST`，默认 `localhost`
- `DB_PORT`，默认 `3306`
- `DB_NAME`，默认 `<service>_db`
- `DB_USER`，必填
- `DB_PASSWORD`，必填
- `PORT`，默认 `8080`
- `JWT_SECRET`，必填且至少 32 个字符

可在变量名前添加服务前缀：

```text
USER_DB_HOST
SHOP_DB_HOST
REVIEW_DB_HOST
ORDER_DB_HOST
SHOP_REDIS_HOST
ORDER_REDIS_HOST
REVIEW_KAFKA_BROKERS
ORDER_KAFKA_BROKERS
```

Redis 配置：

- `REDIS_HOST`
- `REDIS_PORT`，默认 `6379`
- `REDIS_PASSWORD`

Kafka 配置：

- `KAFKA_BROKERS`，多个 Broker 使用逗号分隔

### 服务资源要求

| 服务 | MySQL | Redis | Kafka | JWT |
| --- | --- | --- | --- | --- |
| User | `user_db` | 不需要 | 不需要 | 签发和验证 |
| Shop | `shop_db` | Shop Redis | 当前配置要求 | 验证 |
| Review | `review_db` | 不需要 | Producer | 验证 |
| Order | `order_db` | Order Redis | Producer | 验证 |

## 数据库所有权

```text
user_db
└── users

shop_db
└── shops

review_db
├── reviews
└── likes

order_db
├── coupons
└── orders
```

各服务启动时只执行自身 `internal/migrate` 中的迁移。

`order_db.orders` 保留 `(user_id, coupon_id)` 唯一索引，它是“一人一单”的最终一致性保障，不能只依赖 Redis。

## Kafka 事件

Review Service 向 `review.events` 发布统一事件结构：

```json
{
  "event_id": "100-1767225600000000000",
  "event_type": "review.created",
  "occurred_at": 1767225600,
  "source": "review-service",
  "payload": {
    "review_id": 100,
    "shop_id": 10,
    "score": 5
  }
}
```

事件类型定义在 `api/events/review.go`：

- `review.created`
- `review.liked`
- `review.unliked`

Order Service 向 `orders.created` 发布订单创建消息。

## API

需要鉴权的接口必须携带：

```http
Authorization: Bearer <token>
Content-Type: application/json
```

| 方法 | 路径 | 服务 | 鉴权 | 说明 |
| --- | --- | --- | --- | --- |
| `GET` | `/health` | Gateway | 否 | Gateway 健康检查 |
| `POST` | `/register` | User | 否 | 注册 |
| `POST` | `/login` | User | 否 | 登录并签发 JWT |
| `GET` | `/user/profile` | User | 是 | 当前用户资料 |
| `GET` | `/shops` | Shop | 否 | 商铺分页列表 |
| `GET` | `/shops/:id` | Shop | 否 | 商铺详情 |
| `POST` | `/shops` | Shop | 是 | 创建商铺 |
| `GET` | `/reviews/:shop_id` | Review | 否 | 商铺点评列表 |
| `POST` | `/reviews` | Review | 是 | 发布点评 |
| `POST` | `/reviews/:id/like` | Review | 是 | 点赞或取消点赞 |
| `POST` | `/coupons` | Order | 是 | 创建优惠券 |
| `POST` | `/seckill/:coupon_id` | Order | 是 | 提交秒杀请求 |

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
  -d '{"name":"Demo Cafe","category":"cafe","address":"No. 1 Demo Road","avg_price":35}'
```

发布点评：

```bash
curl -X POST http://localhost:8081/reviews \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"shop_id":1,"content":"Good experience","score":5}'
```

创建优惠券：

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

## 测试与检查

运行所有 Go 测试和编译检查：

```bash
go test ./...
```

验证 Compose 配置前，需要先设置必填密码和 JWT 环境变量：

```bash
docker compose -f docker/docker-compose.yml config --quiet
```

如果当前网络无法访问 Go 官方模块代理，可以为当前 Shell 临时设置其他可用代理。是否使用第三方代理应遵循所在组织的安全策略。

## 当前限制

- JWT 当前使用共享 HS256 Secret；更严格的部署建议改为 RS256，由 User Service 持有私钥，其他服务只持有公钥。
- Shop 对点评事件的消费和基于 `event_id` 的幂等记录仍需进一步完善。
- Order 的 Kafka Consumer、重试、死信队列和持久化幂等消费仍需进一步完善。
- 当前服务在启动时执行自身迁移；生产环境建议改为独立、可审计的迁移 Job。
- 创建商铺和优惠券目前仅要求登录，尚未实现管理员或商户角色授权。
- 缺少限流、熔断、链路追踪、Prometheus 指标和集中式结构化日志。
- 自动化测试覆盖仍然有限，需要增加 Redis、Kafka、MySQL 和并发秒杀集成测试。
- Kafka 当前使用 ZooKeeper 模式，后续可迁移到 KRaft。

## License

当前仓库尚未声明开源许可证。在添加明确的 `LICENSE` 前，请勿默认将其视为可自由再分发的软件。
