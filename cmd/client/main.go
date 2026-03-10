// Package main 存储系统的测试客户端入口
package main

import (
	"dss/pkg/protocol"
	"flag"
	"fmt"
	"log"
	"net"
)

func main() {
	// 解析命令行参数
	port := flag.Int("port", 8080, "要连接的服务端端口")
	flag.Parse()

	addr := fmt.Sprintf("localhost:%d", *port)

	// 建立 TCP 连接
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// 1. 执行 PUT 操作测试
	key := "hello"
	val := []byte("world")
	req := &protocol.Request{Op: protocol.PUT, Key: key, Value: val}
	if err := protocol.SendRequest(conn, req); err != nil {
		log.Fatal(err)
	}
	res, err := protocol.ReceiveResponse(conn)
	if err != nil {
		log.Fatal(err)
	}
	if res.Success {
		fmt.Printf("PUT 成功: %s -> %s\n", key, string(val))
	} else {
		fmt.Printf("PUT 失败: %v\n", res.Error)
	}

	// 2. 执行 GET 操作测试
	req = &protocol.Request{Op: protocol.GET, Key: key}
	if err := protocol.SendRequest(conn, req); err != nil {
		log.Fatal(err)
	}
	res, err = protocol.ReceiveResponse(conn)
	if err != nil {
		log.Fatal(err)
	}
	if res.Success {
		fmt.Printf("GET 成功: %s -> %s\n", key, string(res.Value))
	} else {
		fmt.Printf("GET 失败: %v\n", res.Error)
	}

	// 3. 执行文件存储测试 (模拟上传一个 .txt 文件)
	fileName := "docs/readme.txt"
	fileContent := []byte("这是一个分布式存储系统，支持 KV 和文件存储。")
	fmt.Printf("\n正在上传文件: %s\n", fileName)

	req = &protocol.Request{Op: protocol.PUT, Key: fileName, Value: fileContent}
	if err := protocol.SendRequest(conn, req); err != nil {
		log.Fatal(err)
	}
	res, err = protocol.ReceiveResponse(conn)
	if err != nil {
		log.Fatal(err)
	}
	if res.Success {
		fmt.Println("文件上传成功！")
	}

	// 4. 执行文件下载测试
	fmt.Printf("正在下载文件: %s\n", fileName)
	req = &protocol.Request{Op: protocol.GET, Key: fileName}
	if err := protocol.SendRequest(conn, req); err != nil {
		log.Fatal(err)
	}
	res, err = protocol.ReceiveResponse(conn)
	if err != nil {
		log.Fatal(err)
	}
	if res.Success {
		fmt.Printf("下载成功，内容为: %s\n", string(res.Value))
	} else {
		fmt.Printf("下载失败: %v\n", res.Error)
	}
}
