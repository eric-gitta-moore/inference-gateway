# 构建阶段
FROM golang:1.24-alpine AS builder

USER root
WORKDIR /app

# 复制go.mod和go.sum文件（如果存在）
COPY go.* .

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# 运行阶段
FROM alpine:latest AS prod

WORKDIR /app

# 安装Python3
RUN apk add --no-cache python3

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .
# 复制健康检查脚本
COPY scripts/healthcheck.py scripts/

# 暴露端口
EXPOSE 8080

# 运行应用
ENV IMMICH_API=http://192.168.1.100:3003
ENV MT_PHOTOS_API=http://192.168.1.100:8060
ENV MT_PHOTOS_API_KEY=mt_photos_ai_extra

ENV GIN_MODE=release
CMD ["./main"]

HEALTHCHECK CMD python3 scripts/healthcheck.py