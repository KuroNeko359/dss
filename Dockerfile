# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 复制 go.mod 和 go.sum (如果有)
COPY go.mod ./
RUN go mod download

# 复制源代码
COPY . .

# 构建二进制文件
RUN CGO_ENABLED=0 GOOS=linux go build -o dss-server ./cmd/server/main.go

# 运行阶段
FROM alpine:latest

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/dss-server .
# 复制静态 Web 资源
COPY --from=builder /app/web ./web

# 暴露 TCP 和 Web 端口
EXPOSE 8080 8000

# 启动应用
ENTRYPOINT ["./dss-server"]
