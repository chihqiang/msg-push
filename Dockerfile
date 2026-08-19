# syntax=docker/dockerfile:1

# ============================================================
# 阶段 1：前端构建（Vue 3 + Vite，产物经 go:embed 打入后端二进制）
# ============================================================
FROM node:22-alpine AS web-builder

WORKDIR /web

# 先拷贝依赖清单，利用构建缓存
COPY ui/package.json ui/package-lock.json ./
RUN npm ci

# 拷贝前端源码并构建（产物输出到 /web/dist）
COPY ui/ ./
RUN npm run build

# ============================================================
# 阶段 2：Go 编译
# ============================================================
# infra-go 的 orm 包引入了 gorm.io/driver/sqlite（依赖 mattn/go-sqlite3，需要 CGO），
# 因此保留 CGO；编译时静态链接 musl，产物可直接在 alpine 上运行。
FROM golang:1.25-alpine AS builder

# SQLite 驱动（cgo）编译需要 gcc（alpine 为 musl-gcc）
RUN apk add --no-cache build-base

WORKDIR /src
# ENV GOPROXY=https://goproxy.cn,direct

# 先拷贝依赖清单，下载 Go 依赖
COPY go.mod go.sum ./
RUN go mod download

# 拷贝全部源码（ui/dist 通过 go:embed 编译进二进制）
COPY . .

# 前端产物由 node 阶段生成（源码仓库的 dist 通常已被 git 忽略）
COPY --from=web-builder /web/dist ./ui/dist

# 编译：-trimpath 保证可复现，-s -w 去除调试信息；
# netgo/osusergo + 外部静态链接，避免 glibc NSS 依赖，得到纯 musl 静态二进制
RUN CGO_ENABLED=1 go build -trimpath \
      -tags "netgo osusergo" \
      -ldflags="-s -w -linkmode external -extldflags -static" \
      -o /out/msg-push .

# ============================================================
# 阶段 3：运行镜像
# ============================================================
FROM alpine:3.20 AS runtime

# ca-certificates：HTTPS 调用短信服务商 / Webhook 回调
# tzdata：时区支持（配合 TZ 环境变量）
RUN apk add --no-cache ca-certificates tzdata \
    && update-ca-certificates

# 非 root 运行（降低安全风险）
RUN adduser -D -u 10001 appuser

WORKDIR /app

COPY --from=builder /out/msg-push /app/msg-push
# 容器配置：连接信息通过环境变量覆盖（${VAR} 占位符）
COPY deploy/docker-config.yaml /app/config.yaml

# 环境变量默认值（对应 deploy/docker-compose.yml 中的 mysql/redis 服务名）
ENV APP_ENV=prod \
    SERVER_HOST=0.0.0.0 \
    SERVER_PORT=8080 \
    DB_HOST=mysql \
    DB_PORT=3306 \
    DB_USER=root \
    DB_PASSWORD=msgpush123456 \
    DB_NAME=msg_push \
    REDIS_ADDR=redis:6379 \
    TASKQ_REDIS_ADDR=redis:6379 \
    JWT_SECRET=change-me-in-prod \
    ACCOUNT_SEED_USERNAME=admin \
    ACCOUNT_SEED_PASSWORD=admin123 \
    TZ=Asia/Shanghai

# 赋予运行用户写权限（SQLite 文件模式时需要）
RUN chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8080/health >/dev/null 2>&1 || exit 1

# 仅使用 CMD（不设 ENTRYPOINT），容器启动即运行消息推送服务
CMD ["/app/msg-push", "-config", "/app/config.yaml"]
