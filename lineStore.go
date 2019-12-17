package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"sync"
)

// 保存程序读取的文件inode、行数
// 初始化加载本地记录文件
// 记录数据保存到本地文件

// 日志文件监听记录
type LogFileStatus struct {
	logStore map[uint64]uint64
	file     string
	mu       sync.RWMutex
}

// LogFileStatus生成函数
func NewLogFileStatus(file string) *LogFileStatus {
	lineStore := &LogFileStatus{
		logStore: make(map[uint64]uint64),
		file:     file,
	}
	if err := lineStore.Load(); err != nil {
		log.Fatal(err)
	}

	return lineStore

}

// 设置文件读取行数
func (l *LogFileStatus) Set(inode uint64, line uint64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logStore[inode] = line
}

//获取文件读取行数
func (l *LogFileStatus) Get(inode uint64) (uint64, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	line, ok := l.logStore[inode]
	if !ok {
		return 0, false
	}
	return line, true
}

// 文件监听数量
func (l *LogFileStatus) Count() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.logStore)
}

//文件监听记录保存到文件
func (l *LogFileStatus) Save() error {
	f, _ := os.OpenFile(l.file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	defer f.Close()
	e := json.NewEncoder(f)
	err := e.Encode(l.logStore)
	if err != nil {
		return err
	}
	return nil
}

// 从文件加载文件监听记录
func (l *LogFileStatus) Load() error {
	f, _ := os.OpenFile(l.file, os.O_RDONLY|os.O_CREATE, 0644)
	defer f.Close()

	d := json.NewDecoder(f)
	m := make(map[uint64]uint64)
	err := d.Decode(&m)
	if err == nil {
		l.logStore = m
		return nil
	}
	if err == io.EOF {
		return nil
	}
	return err
}
