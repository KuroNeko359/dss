package server

import (
	"dss/internal/storage"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"path/filepath"
)

// WebServer 提供了 HTTP 接口和静态资源服务
type WebServer struct {
	addr    string
	storage storage.Storage
}

// NewWebServer 创建一个新的 Web 服务端实例
func NewWebServer(addr string, storage storage.Storage) *WebServer {
	return &WebServer{
		addr:    addr,
		storage: storage,
	}
}

// Start 启动 HTTP 服务
func (ws *WebServer) Start() error {
	mux := http.NewServeMux()

	// 静态文件服务
	fs := http.FileServer(http.Dir("./web"))
	mux.Handle("/", fs)

	// API 路由
	mux.HandleFunc("/api/storage", ws.handleStorage)
	mux.HandleFunc("/api/upload", ws.handleUpload)
	mux.HandleFunc("/api/list", ws.handleList)

	log.Printf("Web server listening on http://localhost%s\n", ws.addr)
	return http.ListenAndServe(ws.addr, mux)
}

// handleStorage 处理 KV 存取请求
func (ws *WebServer) handleStorage(w http.ResponseWriter, r *http.Request) {
	// 简单的 CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" && r.Method != "POST" {
		http.Error(w, "missing key", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		val, err := ws.storage.Get(key)
		if err != nil {
			if err == storage.ErrNotFound {
				http.Error(w, "not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// 根据文件名后缀自动设置 Content-Type
		ext := filepath.Ext(key)
		contentType := mime.TypeByExtension(ext)
		if contentType == "" {
			// 如果无法识别，则尝试探测内容
			contentType = http.DetectContentType(val)
		}
		w.Header().Set("Content-Type", contentType)
		w.Write(val)

	case "POST":
		var data struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := ws.storage.Put(data.Key, []byte(data.Value)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "success")

	case "DELETE":
		if err := ws.storage.Delete(key); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "success")

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleUpload 处理文件上传
func (ws *WebServer) handleUpload(w http.ResponseWriter, r *http.Request) {
	// 简单的 CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 限制文件大小 (10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "invalid file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 存储文件，使用文件名作为 Key
	if err := ws.storage.Put(header.Filename, content); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "success")
}

// handleList 返回所有存储的键列表
func (ws *WebServer) handleList(w http.ResponseWriter, r *http.Request) {
	// 简单的 CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != "GET" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	keys, err := ws.storage.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if keys == nil {
		keys = []string{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(keys)
}
