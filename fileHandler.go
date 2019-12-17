package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	file      string // 文件路径
	lineStore *LogFileStatus
	title     string   // 告警标题
	level     []string // 告警匹配字段
	notify    []string // 需要报警的日志
	url       string   // 钉钉通知的url
}

// 文件监听处理程序
func (h *Handler) Handler(listener chan uint64, inode uint64) {
	log.Println("文件监听处理程序")
	defer func() {
		listener <- inode
	}()
	log.Printf(h.file)
	f, err := os.Open(h.file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	//isExistLine := false // 是否有历史读取行数
	br := bufio.NewReader(f)

	line, ok := h.lineStore.Get(inode)
	if !ok {
		for i := 0; ; i++ {
			_, _, err := br.ReadLine()
			if err == io.EOF {
				line = uint64(i)
				h.lineStore.Set(inode, line) // 设置inode记录
				break
			}
			if err != nil {
				log.Println(h.file+":读取失败", err)
				return
			}
		}
	} else {
		for i := 0; uint64(i) < line; i++ {
			_, _, err := br.ReadLine()
			if err == io.EOF {
				line = uint64(i)
				h.lineStore.Set(inode, line) // 设置inode记录
				break
			}
			if err != nil {
				log.Println(h.file+":读取失败", err)
				return
			}

			h.lineStore.Set(inode, line) // 设置inode记录
		}
	}

	upTime := 0 // 文件累计无更新时间
	for {
		l, err := br.ReadString('\n')
		if err == io.EOF {
			log.Printf(h.file + "文件读取完成，暂时休眠")
			if upTime > OutTime {
				log.Printf(h.file + "文件长时间未更新，关闭文件资源，停止监控")
				return
			}
			h.Notify()                   // 告警
			h.lineStore.Set(inode, line) // 设置inode记录
			upTime += IntervalTime
			time.Sleep(IntervalTime * time.Second)
			log.Printf(h.file + "停止休眠，开始尝试读取文件")
			continue
		}
		if err != nil {
			return
		}
		line += 1
		upTime = 0
		// 处理日志文件数据
		h.MatchLine(l)
	}
}

// 处理日志文件数据
func (h *Handler) MatchLine(l string) {

	for _, v := range h.level {
		if strings.Contains(strings.ToLower(l), strings.ToLower(v)) {
			// 需要通知的日志
			h.notify = append(h.notify, l)
		}
	}

}

/*
"msgtype": "text",
    "text": {
        "content": "我就是我, 是不一样的烟火@156xxxx8827"
    },
    "at": {
        "atMobiles": [
            "156xxxx8827",
            "189xxxx8325"
        ],
        "isAtAll": false
    }
*/
type Notify struct {
	MsgType string `json:"msgtype"`
	Text    Text   `json:"text"`
	At      At     `json:"at"`
}
type Text struct {
	Content string `json:"content"`
}
type At struct {
	AtMobiles []string `json:"atMobiles"`
	IsAtAll   bool     `json:"isAtAll"`
}

// 发送告警通知
func (h *Handler) Notify() {
	if len(h.notify) <= 0 {
		return
	}
	text := "【告警】" + h.title + "\n" +
		"错误次数：" + strconv.Itoa(len(h.notify)) + "\n" +
		"内容：" + h.notify[0]
	data := Notify{
		MsgType: "text",
		Text: Text{
			Content: text,
		},
		At: At{
			AtMobiles: nil,
			IsAtAll:   false,
		},
	}
	j, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
		return
	}
	client := &http.Client{}

	//fmt.Println(h.url)
	req, err := http.NewRequest("POST", h.url, bytes.NewBuffer(j))
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Set("Content-Type", "application/json;charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		log.Println("请求钉钉通知失败：", err)
		return
	}
	defer resp.Body.Close()
	fmt.Println("钉钉通知返回码：" + string(resp.StatusCode))

	h.notify = nil
}

//func (h *Handler) Notify() {
//	if len(h.notify) <= 0 {
//		return
//	}
//	n := len(h.notify)
//	text := "【告警】" + h.title + "\n" +
//		"错误次数：" + strconv.Itoa(n) + "\n" +
//		"内容：" + h.notify[0]
//	fmt.Println(text)
//	h.notify = nil
//}
