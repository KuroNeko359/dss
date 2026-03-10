// Package storage 提供了分布式存储系统的存储引擎接口及实现
package storage

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

var (
	// ErrNotFound 当请求的键不存在时返回此错误
	ErrNotFound = errors.New("key not found")
)

// Storage 定义了存储引擎的基本操作接口
type Storage interface {
	// Put 存储一个键值对
	Put(key string, value []byte) error
	// Get 根据键获取对应的值
	Get(key string) ([]byte, error)
	// Delete 根据键删除对应的键值对
	Delete(key string) error
	// List 返回存储中所有的键
	List() ([]string, error)
}

// MemoryStorage 是一个基于内存的线程安全存储实现
type MemoryStorage struct {
	mu   sync.RWMutex
	data map[string][]byte
}

// NewMemoryStorage 创建并返回一个新的内存存储实例
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string][]byte),
	}
}

// Put 实现了 Storage 接口，将数据存入内存
func (s *MemoryStorage) Put(key string, value []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	return nil
}

// Get 实现了 Storage 接口，从内存中检索数据
func (s *MemoryStorage) Get(key string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.data[key]
	if !ok {
		return nil, ErrNotFound
	}
	return val, nil
}

// Delete 实现了 Storage 接口，从内存中删除数据
func (s *MemoryStorage) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	return nil
}

// List 实现了 Storage 接口，返回内存中所有的键
func (s *MemoryStorage) List() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys, nil
}

// FileStorage 是一个基于磁盘文件的存储实现，适用于存储较大文件
type FileStorage struct {
	rootDir string
}

// NewFileStorage 创建并返回一个新的文件存储实例
func NewFileStorage(rootDir string) (*FileStorage, error) {
	// 确保根目录存在
	if err := os.MkdirAll(rootDir, 0755); err != nil {
		return nil, err
	}
	return &FileStorage{rootDir: rootDir}, nil
}

// Put 将数据写入磁盘文件，键作为文件名
func (s *FileStorage) Put(key string, value []byte) error {
	path := filepath.Join(s.rootDir, key)
	// 确保父目录存在（支持 key 中带有路径，如 "docs/test.txt"）
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, value, 0644)
}

// Get 从磁盘读取文件内容
func (s *FileStorage) Get(key string) ([]byte, error) {
	path := filepath.Join(s.rootDir, key)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return data, nil
}

// Delete 从磁盘删除文件
func (s *FileStorage) Delete(key string) error {
	path := filepath.Join(s.rootDir, key)
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// List 返回文件存储中的所有键（相对于根目录的路径）
func (s *FileStorage) List() ([]string, error) {
	var keys []string
	err := filepath.Walk(s.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			rel, err := filepath.Rel(s.rootDir, path)
			if err != nil {
				return err
			}
			keys = append(keys, rel)
		}
		return nil
	})
	return keys, err
}

// PostgresStorage 是一个基于 PostgreSQL 的存储实现，用于持久化元数据或小对象
type PostgresStorage struct {
	db *sql.DB
}

// NewPostgresStorage 创建并返回一个新的 PostgreSQL 存储实例，包含重试逻辑
func NewPostgresStorage(connStr string) (*PostgresStorage, error) {
	var db *sql.DB
	var err error

	// 增加重试逻辑，等待数据库启动
	for i := 0; i < 5; i++ {
		db, err = sql.Open("postgres", connStr)
		if err == nil {
			if err = db.Ping(); err == nil {
				break
			}
		}
		log.Printf("正在等待数据库启动 (重试 %d/5)...", i+1)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		return nil, err
	}

	// 创建基础表结构
	query := `
	CREATE TABLE IF NOT EXISTS storage (
		key TEXT PRIMARY KEY,
		value BYTEA,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.Exec(query); err != nil {
		return nil, err
	}

	return &PostgresStorage{db: db}, nil
}

// Put 将键值对存入 PostgreSQL
func (s *PostgresStorage) Put(key string, value []byte) error {
	query := `
	INSERT INTO storage (key, value, updated_at)
	VALUES ($1, $2, CURRENT_TIMESTAMP)
	ON CONFLICT (key) DO UPDATE
	SET value = EXCLUDED.value, updated_at = CURRENT_TIMESTAMP;`
	_, err := s.db.Exec(query, key, value)
	return err
}

// Get 从 PostgreSQL 读取数据
func (s *PostgresStorage) Get(key string) ([]byte, error) {
	var value []byte
	query := `SELECT value FROM storage WHERE key = $1`
	err := s.db.QueryRow(query, key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return value, nil
}

// Delete 从 PostgreSQL 删除数据
func (s *PostgresStorage) Delete(key string) error {
	query := `DELETE FROM storage WHERE key = $1`
	_, err := s.db.Exec(query, key)
	return err
}

// List 从 PostgreSQL 获取所有键
func (s *PostgresStorage) List() ([]string, error) {
	query := `SELECT key FROM storage ORDER BY updated_at DESC`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, nil
}
