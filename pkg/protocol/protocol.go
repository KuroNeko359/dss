// Package protocol 定义了分布式存储系统节点间通讯的自定义二进制协议
package protocol

import (
	"encoding/gob"
	"io"
)

// Op 定义了存储操作的类型
type Op int

const (
	// PUT 写入/更新操作
	PUT Op = iota
	// GET 读取操作
	GET
	// DELETE 删除操作
	DELETE
)

// Request 表示从客户端发送到服务端的请求数据结构
type Request struct {
	Op    Op     // 操作类型
	Key   string // 键
	Value []byte // 值（仅 PUT 操作需要）
}

// Response 表示从服务端返回给客户端的响应数据结构
type Response struct {
	Success bool   // 操作是否成功
	Value   []byte // 获取的值（仅 GET 操作成功时返回）
	Error   string // 错误信息（若 Success 为 false）
}

// SendRequest 将请求对象序列化并写入输出流
func SendRequest(w io.Writer, req *Request) error {
	return gob.NewEncoder(w).Encode(req)
}

// ReceiveRequest 从输入流中反序列化并读取请求对象
func ReceiveRequest(r io.Reader) (*Request, error) {
	var req Request
	if err := gob.NewDecoder(r).Decode(&req); err != nil {
		return nil, err
	}
	return &req, nil
}

// SendResponse 将响应对象序列化并写入输出流
func SendResponse(w io.Writer, res *Response) error {
	return gob.NewEncoder(w).Encode(res)
}

// ReceiveResponse 从输入流中反序列化并读取响应对象
func ReceiveResponse(r io.Reader) (*Response, error) {
	var res Response
	if err := gob.NewDecoder(r).Decode(&res); err != nil {
		return nil, err
	}
	return &res, nil
}
