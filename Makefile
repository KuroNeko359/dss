.PHONY: up down server client test clean

# 启动基础设施 (Postgres)
up:
	docker-compose up -d

# 停止并移除基础设施
down:
	docker-compose down

# 启动服务端 (同时开启 TCP 和 Web 控制面板)
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
