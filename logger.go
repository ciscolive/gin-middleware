package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// LogLayout 日志layout
type LogLayout struct {
	Time      time.Time
	Metadata  map[string]interface{} // 存储自定义原数据
	Path      string                 // 访问路径
	Query     string                 // 携带query
	Body      string                 // 携带body数据
	IP        string                 // ip地址
	UserAgent string                 // 代理
	Error     string                 // 错误
	Cost      time.Duration          // 花费时间
	Source    string                 // 来源
}

type Logger struct {
	Filter        func(c *gin.Context) bool               // Filter 用户自定义过滤
	FilterKeyword func(layout *LogLayout) bool            // FilterKeyword 关键字过滤(key)
	AuthProcess   func(c *gin.Context, layout *LogLayout) // AuthProcess 鉴权处理
	Print         func(LogLayout)                         // 日志处理
	Source        string                                  // Source 服务唯一标识
}

func (l Logger) SetLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		var body []byte
		if l.Filter != nil && !l.Filter(c) {
			body, _ = c.GetRawData() // 将原body塞回去
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		}
		c.Next()
		cost := time.Since(start)
		layout := LogLayout{
			Time:      time.Now(),
			Path:      path,
			Query:     query,
			IP:        c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
			Error:     strings.TrimRight(c.Errors.ByType(gin.ErrorTypePrivate).String(), "\n"),
			Cost:      cost,
			Source:    l.Source,
		}
		if l.Filter != nil && !l.Filter(c) {
			layout.Body = string(body)
		}
		l.AuthProcess(c, &layout) // 处理鉴权需要的信息
		if l.FilterKeyword != nil {
			l.FilterKeyword(&layout) // 自行判断key/value 脱敏等
		}
		if l.AuthProcess != nil {
			l.AuthProcess(c, &layout) // 处理鉴权需要的信息
		}
		l.Print(layout) // 自行处理日志
	}
}

func DefaultLogger() gin.HandlerFunc {
	return Logger{
		Print: func(layout LogLayout) {
			v, _ := json.Marshal(layout) // 标准输出,k8s做收集
			fmt.Println(string(v))
		},
		Source: "GVA",
	}.SetLoggerMiddleware()
}
