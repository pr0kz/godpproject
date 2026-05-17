# Docker 配置文件

## 项目容器化说明

### 快速启动

```bash
# 启动所有服务（MySQL + Redis + API）
docker-compose up -d

# 查看日志
docker-compose logs -f api

# 停止所有服务
docker-compose down

# 清理所有数据（包括数据库）
docker-compose down -v
```

### 服务信息

| 服务 | 容器名 | 端口 | 镜像 |
|------|--------|------|------|
| API | ai-review-api | 8080 | 本地构建 |
| MySQL | ai-review-mysql | 3306 | mysql:8.0 |
| Redis | ai-review-redis | 6379 | redis:7.0-alpine |

### 环境变量

在 `docker-compose.yml` 中配置：
- `DB_USER`: MySQL 用户名（默认：root）
- `DB_PASSWORD`: MySQL 密码（默认：root）
- `DB_HOST`: MySQL 主机（默认：mysql）
- `DB_PORT`: MySQL 端口（默认：3306）
- `DB_NAME`: 数据库名（默认：ai_review_system）
- `JWT_SECRET`: JWT 密钥（需修改为生产密钥）
- `PORT`: API 端口（默认：8080）

### 数据持久化

- MySQL 数据存储在 `mysql_data` 卷
- Redis 数据存储在 `redis_data` 卷

### 健康检查

所有服务都配置了健康检查：
- API：检查 `/health` 端点
- MySQL：检查 mysqladmin ping
- Redis：检查 redis-cli ping

### 网络

所有服务连接到 `ai-review-network` 桥接网络，可以通过服务名相互通信。

### 构建镜像

```bash
# 手动构建 API 镜像
docker build -t ai-review-api:latest .

# 查看镜像
docker images | grep ai-review
```

### 常见问题

**Q: 端口已被占用**
```bash
# 修改 docker-compose.yml 中的端口映射
# 例如：将 "8080:8080" 改为 "9090:8080"
```

**Q: 数据库连接失败**
```bash
# 确保 MySQL 容器已启动
docker-compose ps

# 查看 MySQL 日志
docker-compose logs mysql
```

**Q: 如何进入容器**
```bash
# 进入 API 容器
docker exec -it ai-review-api sh

# 进入 MySQL 容器
docker exec -it ai-review-mysql mysql -u root -proot
```

### 生产部署建议

1. 修改 `JWT_SECRET` 为强密钥
2. 修改 MySQL 密码
3. 使用环境变量文件 `.env`
4. 配置持久化存储
5. 设置资源限制
6. 使用私有镜像仓库
7. 配置日志驱动

