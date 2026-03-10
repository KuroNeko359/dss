.PHONY: up down server client test clean

# 启动整个系统 (Docker 部署)
deploy:
	docker-compose up -d --build

# 停止整个系统
stop:
	docker-compose down

# 启动服务端 (本地开发，使用 Postgres 模式)
server:
	go run cmd/server/main.go -storage postgres -conn "postgres://postgres:postgres@localhost:5432/dss_metadata?sslmode=disable"

# 启动服务端 (内存模式，方便快速演示 Web 界面)
web-demo:
	go run cmd/server/main.go -storage memory

# 运行测试客户端
client:
	go run cmd/client/main.go

# 运行所有测试
test:
	go test ./...

# 清理生成的二进制文件或数据目录
clean:
	rm -rf my_storage data
