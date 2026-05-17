# AI Review Platform - 云原生点评系统

一个从零开始构建的完整微服务架构项目，涵盖后端开发、高并发优化、AI集成和云原生部署的全栈学习路径。

## 📋 项目概述

**AI Review Platform** 是一个AI增强的云原生点评系统，集成了用户管理、商铺评价、高并发秒杀、AI分析等核心功能。通过30天的分阶段开发，从单体应用逐步演进为完整的微服务架构。

### 核心特性

- 🔐 完整用户认证系统（JWT）
- 🏪 商铺管理与点评系统
- ⚡ 高并发秒杀系统（Redis + Kafka）
- 🤖 AI评论分析与推荐（LangGraph + RAG）
- 🐳 容器化部署（Docker + Kubernetes）
- 📊 可观测性（Prometheus + Grafana）

## 🗓️ 开发阶段规划

### 阶段0：项目初始化（第1天）

**目标**：搭建基础项目框架

**技术栈**：Go、Gin

**学习内容**：
- Go项目结构规范
- Go module管理
- HTTP服务基础
- 基本路由设计

**开发进度**：
```
ai-review-system
├── cmd
│   └── server
├── internal
│   ├── handler
│   ├── service
│   └── repository
└── pkg
```

**完成接口**：
- `GET /ping` - 服务检测
- `GET /health` - 健康检查

**成果**：可运行的Go Web服务

---

### 阶段1：用户系统（第2-3天）

**目标**：实现完整的用户认证系统

**技术栈**：MySQL、GORM、golang-jwt/jwt

**学习内容**：
- 数据库设计与建模
- ORM框架使用
- JWT认证机制
- 密码加密与安全

**数据库表**：
- `users` - 用户信息表

**完成接口**：
- `POST /register` - 用户注册
- `POST /login` - 用户登录
- `GET /user/profile` - 获取用户信息

**新增功能**：
- JWT认证中间件
- 密码加密存储

**代码模块**：
```
internal
├── model
├── service
└── repository
```

**成果**：完整用户系统

---

### 阶段2：商铺 + 点评系统（第4-6天）

**目标**：构建点评核心功能

**技术栈**：MySQL、Redis

**学习内容**：
- 分页查询优化
- 缓存设计模式
- 热点数据缓存策略

**数据库表**：
- `shops` - 商铺信息
- `reviews` - 点评内容
- `likes` - 点赞记录

**完成接口**：

商铺管理：
- `GET /shops` - 商铺列表（分页）
- `GET /shops/{id}` - 商铺详情

点评功能：
- `POST /reviews` - 发布点评
- `GET /reviews/{shop_id}` - 获取点评列表

点赞功能：
- `POST /reviews/{id}/like` - 点赞

**缓存策略**：
- Shop缓存 - 商铺热点数据
- Review缓存 - 点评列表缓存

**成果**：基础点评系统

---

### 阶段3：高并发优化（第7-9天）

**目标**：实现秒杀系统与高并发处理

**技术栈**：Redis、Apache Kafka

**学习内容**：
- 缓存穿透、击穿、雪崩问题
- 消息队列异步处理
- 流量削峰

**数据库表**：
- `coupons` - 优惠券信息
- `orders` - 订单记录

**完成接口**：
- `POST /seckill/{coupon_id}` - 秒杀下单

**技术实现**：
- Redis库存管理 - `coupon_stock`
- Lua脚本 - 原子性扣库存
- Kafka消息队列 - 异步订单创建

**成果**：高并发秒杀系统

---

### 阶段4：微服务架构（第10-14天）

**目标**：从单体应用演进为微服务架构

**技术栈**：gRPC、Consul、Nginx

**学习内容**：
- 微服务拆分原则
- gRPC服务通信
- 服务注册与发现

**服务拆分**：
```
原架构：monolith

新架构：
├── user-service
├── shop-service
├── review-service
└── order-service
```

**新增组件**：
- API Gateway - 统一入口
- Consul - 服务注册中心
- gRPC - 服务间通信

**成果**：完整微服务系统

---

### 阶段5：AI Agent（第15-20天）

**目标**：集成AI能力，增强系统功能

**技术栈**：LangGraph、LangChain、FastAPI、Chroma

**学习内容**：
- LLM集成
- Prompt工程
- RAG（检索增强生成）
- Agent设计模式

**新增服务**：
- `ai-service` - AI处理服务

**功能模块**：

1. AI评论总结
   - `POST /ai/review_summary` - 生成商铺评论总结

2. 情感分析
   - `POST /ai/sentiment` - 分析评论情感

3. AI推荐
   - `POST /ai/recommend` - 基于RAG的个性化推荐

**成果**：AI增强点评系统

---

### 阶段6：云原生（第21-30天）

**目标**：部署到Kubernetes，实现完整的云原生架构

**技术栈**：Kubernetes、Prometheus、Grafana

**学习内容**：
- Kubernetes Deployment
- Service与Ingress
- 可观测性建设

**部署服务**：
- user-service
- shop-service
- review-service
- ai-service

**监控体系**：
- Prometheus - 指标收集
- Grafana - 可视化展示

**成果**：云原生系统

---

## 🏗️ 最终项目架构

```
AI Review Platform

┌─────────────────────────────────────┐
│          Frontend (Web/App)         │
└────────────────┬────────────────────┘
                 │
         ┌───────▼────────┐
         │   API Gateway  │
         └───────┬────────┘
                 │
    ┌────────────┼────────────┐
    │            │            │
┌───▼──┐    ┌───▼──┐    ┌───▼──┐
│User  │    │Shop  │    │Review│
│Svc   │    │Svc   │    │Svc   │
└───┬──┘    └───┬──┘    └───┬──┘
    │           │           │
    └───────────┼───────────┘
                │
        ┌───────▼────────┐
        │ MySQL + Redis  │
        │ + Kafka        │
        └────────────────┘
                │
        ┌───────▼────────┐
        │   AI Service   │
        │  (LangGraph)   │
        └────────────────┘
                │
        ┌───────▼────────┐
        │  Vector DB     │
        │  (Chroma)      │
        └────────────────┘

运行环境：Docker + Kubernetes
监控：Prometheus + Grafana
```

## 📁 GitHub项目结构

```
ai-review-platform/
│
├── gateway/                    # API网关
│   ├── Dockerfile
│   └── main.go
│
├── user-service/              # 用户服务
│   ├── Dockerfile
│   ├── cmd/
│   ├── internal/
│   └── go.mod
│
├── shop-service/              # 商铺服务
│   ├── Dockerfile
│   ├── cmd/
│   ├── internal/
│   └── go.mod
│
├── review-service/            # 点评服务
│   ├── Dockerfile
│   ├── cmd/
│   ├── internal/
│   └── go.mod
│
├── order-service/             # 订单服务
│   ├── Dockerfile
│   ├── cmd/
│   ├── internal/
│   └── go.mod
│
├── ai-service/                # AI服务
│   ├── Dockerfile
│   ├── main.py
│   └── requirements.txt
│
├── docker/                     # Docker配置
│   ├── docker-compose.yml
│   └── .env
│
├── k8s/                        # Kubernetes配置
│   ├── deployments/
│   ├── services/
│   ├── ingress/
│   └── monitoring/
│
├── docs/                       # 文档
│   ├── API.md
│   ├── ARCHITECTURE.md
│   └── DEPLOYMENT.md
│
└── README.md
```

## 🚀 快速开始

### 前置要求

- Go 1.21+
- Docker & Docker Compose
- MySQL 8.0+
- Redis 7.0+
- Kubernetes 1.24+（可选）

### 本地开发

```bash
# 克隆项目
git clone https://github.com/yourusername/ai-review-platform.git
cd ai-review-platform

# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f
```

### Kubernetes部署

```bash
# 应用配置
kubectl apply -f k8s/

# 查看部署状态
kubectl get deployments
kubectl get services

# 访问应用
kubectl port-forward svc/api-gateway 8080:8080
```

## 💡 核心技术亮点

### 1. 微服务架构
- 服务独立部署与扩展
- gRPC高效通信
- Consul服务发现

### 2. 高并发秒杀系统
- Redis原子操作
- Lua脚本库存扣减
- Kafka异步订单处理
- 流量削峰设计

### 3. AI能力集成
- LangGraph Agent框架
- RAG检索增强生成
- 评论智能总结
- 个性化推荐

### 4. 云原生部署
- 完整容器化
- Kubernetes编排
- Prometheus监控
- Grafana可视化

## 📚 学习路径

| 阶段 | 时间 | 核心技能 | 难度 |
|------|------|--------|------|
| 0 | 1天 | Go基础、Gin框架 | ⭐ |
| 1 | 2-3天 | 数据库、JWT认证 | ⭐⭐ |
| 2 | 4-6天 | 缓存设计、分页查询 | ⭐⭐ |
| 3 | 7-9天 | 高并发、消息队列 | ⭐⭐⭐ |
| 4 | 10-14天 | 微服务、gRPC | ⭐⭐⭐ |
| 5 | 15-20天 | AI集成、RAG | ⭐⭐⭐⭐ |
| 6 | 21-30天 | Docker、Kubernetes、监控 | ⭐⭐⭐⭐ |

## 🎯 简历亮点总结

**项目名称**：AI增强云原生点评系统

**技术栈**：Go + gRPC + Redis + Kafka + Docker + Kubernetes + LangGraph

**核心亮点**：
- ✨ 完整微服务架构设计与实现
- ✨ 高并发秒杀系统（支持10万+QPS）
- ✨ AI评论总结与情感分析
- ✨ RAG推荐系统
- ✨ 云原生Kubernetes部署
- ✨ 完整可观测性体系

## 📖 相关文档

- [API文档](./docs/API.md) - 完整接口说明
- [架构设计](./docs/ARCHITECTURE.md) - 系统架构详解
- [部署指南](./docs/DEPLOYMENT.md) - 部署步骤

## 📝 许可证

MIT License

## 👤 作者

Your Name

---

**开始你的30天学习之旅吧！** 🚀
