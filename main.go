package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// 文件读取行数记录模块
//读取日志文件、检索指定信息模块
//警告发送模块

// 间隔时间检查是否有新更新的文件需要监听处理
const TimingCheck = 60 * 2

// 文件超过一段时间未更新，释放文件资源指针
const OutTime = 60 * 10

//读取文件间隔时间
const IntervalTime = 60

// 正在监听切片
var listenerList map[uint64]string

// 停止监听通知channel
var listener chan uint64

// 文件读取行数记录
var lineStore *LogFileStatus

func main() {
	// 正在监听的文件inode
	listenerList = make(map[uint64]string)
	// 通知文件停止监听的chan
	listener = make(chan uint64, 1)
	go func(listener chan uint64, listenerList *map[uint64]string) {
		for {
			inode := <-listener
			if _, ok := (*listenerList)[inode]; ok {
				delete(*listenerList, inode)
			}
		}
	}(listener, &listenerList)

	// 优雅退出，退出时保存文件读取行数记录到文件
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGQUIT)
	go shutdown(quit)

	// 存储文件读取行数的结构体
	lineStore = NewLogFileStatus("./lineStore")

	// 配置参数
	conf := getConf()

	// 定时轮询检查是否有需要监控的文件
	for {
		for _, c := range conf {
			pathList := fileList(c.Path)
			// 运行文件监控处理
			regularRun(pathList, c.Title, c.Level, c.Url)
		}
		time.Sleep(TimingCheck * time.Second)
		// 文件读取行数保存到文件
		err := lineStore.Save()
		if err != nil {
			log.Fatal(err)
		}
	}

}

// 优雅退出
func shutdown(quit <-chan os.Signal) {
	<-quit
	err := lineStore.Save()
	if err != nil {
		log.Fatal("退出程序保存文件读取行数失败", err)
	}
	fmt.Println("程序已关闭")
	os.Exit(0)
}

//配置文件信息获取需要监听的文件列表
func fileList(path []string) []string {
	var pathList []string
	for _, p := range path {
		if strings.Contains(p, "*") {
			m, err := filepath.Glob(p)
			if err != nil {
				log.Fatal(err)
			}
			if m != nil {
				pathList = append(pathList, m...)
			}
		} else {
			pathList = append(pathList, p)
		}
	}
	return pathList
}

// 检测一定时间内有更新的文件进行监听
func regularRun(pathList []string, title string, level []string, url string) {
	t := (time.Now().Unix()) - (60 * 2) // 监听文件的更新时间必须大于这个时间
	for _, file := range pathList {
		fileInfo, err := os.Stat(file)
		if err != nil {
			log.Fatalln(err)
		}
		if fileInfo.IsDir() {
			log.Fatal(file + "是个文件夹，需要的是文件")
		}
		if fileInfo.ModTime().Unix() < t {
			continue
		}

		inode := FileInode(fileInfo)
		if _, ok := listenerList[inode]; ok {
			continue
		}
		listenerList[inode] = file

		// 文件监听处理方法
		handler := &Handler{
			file:      file,
			lineStore: lineStore,
			title:     title,
			level:     level,
			url:       url,
		}
		go handler.Handler(listener, inode)
		log.Println("添加活跃日志文件" + file + "到监控...")
	}
}

// 获取文件inode
func FileInode(fileInfo os.FileInfo) uint64 {
	stat, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		log.Fatalln(fileInfo.Name() + "获取文件元信息失败")
	}
	return stat.Ino
}
