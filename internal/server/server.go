// Package server 提供了存储节点的网络服务端实现
package server

import (
	"dss/internal/storage"
	"dss/pkg/protocol"
	"fmt"
	"io"
	"log"
	"net"
)

// Server 表示一个存储节点服务器
type Server struct {
	addr    string          // 服务器监听地址
	storage storage.Storage // 后端存储引擎
}

// NewServer 创建并返回一个新的存储节点服务器实例
func NewServer(addr string, storage storage.Storage) *Server {
	return &Server{
		addr:    addr,
		storage: storage,
	}
}

// Start 启动 TCP 服务器并开始监听请求
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	log.Printf("Server listening on %s\n", s.addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Accept error: %v\n", err)
			continue
		}
		// 为每个连接启动一个 goroutine 处理
		go s.handleConnection(conn)
	}
}

// handleConnection 处理单个 TCP 连接的长连接请求
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		// 循环读取请求
		req, err := protocol.ReceiveRequest(conn)
		if err != nil {
			if err != io.EOF {
				log.Printf("Receive error: %v\n", err)
			}
			return
		}

		// 处理请求并获取响应
		res := s.processRequest(req)
		// 发送响应
		if err := protocol.SendResponse(conn, res); err != nil {
			log.Printf("Send error: %v\n", err)
			return
		}
	}
}

// processRequest 根据请求类型调用存储引擎执行相应操作
func (s *Server) processRequest(req *protocol.Request) *protocol.Response {
	res := &protocol.Response{Success: true}

	switch req.Op {
	case protocol.PUT:
		if err := s.storage.Put(req.Key, req.Value); err != nil {
			res.Success = false
			res.Error = err.Error()
		}
	case protocol.GET:
		val, err := s.storage.Get(req.Key)
		if err != nil {
			res.Success = false
			res.Error = err.Error()
		} else {
			res.Value = val
		}
	case protocol.DELETE:
		if err := s.storage.Delete(req.Key); err != nil {
			res.Success = false
			res.Error = err.Error()
		}
	default:
		res.Success = false
		res.Error = fmt.Sprintf("Unknown operation: %v", req.Op)
	}

	return res
}
