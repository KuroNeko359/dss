// Package main 存储节点服务端的启动入口
package main

import (
	"dss/internal/server"
	"dss/internal/storage"
	"flag"
	"fmt"
	"log"
)

func main() {
	// 解析命令行参数
	port := flag.Int("port", 8080, "监听端口")
	storageType := flag.String("storage", "memory", "存储类型: memory, file 或 postgres")
	dataDir := flag.String("dir", "./data", "文件存储的根目录 (仅在 storage=file 时有效)")
	connStr := flag.String("conn", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable", "PostgreSQL 连接字符串 (仅在 storage=postgres 时有效)")
	flag.Parse()

	addr := fmt.Sprintf(":%d", *port)

	// 根据参数选择存储引擎
	var s storage.Storage
	var err error

	if *storageType == "file" {
		s, err = storage.NewFileStorage(*dataDir)
		if err != nil {
			log.Fatalf("无法初始化文件存储: %v", err)
		}
		log.Printf("使用文件存储引擎，根目录: %s", *dataDir)
	} else if *storageType == "postgres" {
		s, err = storage.NewPostgresStorage(*connStr)
		if err != nil {
			log.Fatalf("无法初始化 PostgreSQL 存储: %v", err)
		}
		log.Printf("使用 PostgreSQL 存储引擎")
	} else {
		s = storage.NewMemoryStorage()
		log.Printf("使用内存存储引擎")
	}

	// 创建并启动 TCP 服务器
	srv := server.NewServer(addr, s)
	go func() {
		if err := srv.Start(); err != nil {
			log.Printf("TCP 服务器错误: %v", err)
		}
	}()

	// 启动 Web 控制面板
	webSrv := server.NewWebServer(":8000", s)
	if err := webSrv.Start(); err != nil {
		log.Fatal(err)
	}
}
