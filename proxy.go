package middleware

import (
	"bufio"
	"github.com/ciscolive/gin-admin/global"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
)

// ProxyHandler 反向代理接口，项目路由定义一个空函数
func ProxyHandler(urlPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		realServer, err := url.Parse(urlPath)
		if err != nil {
			global.Logger.Info("URL.PARSE解析异常：请检查输入路径")
			c.String(http.StatusInternalServerError, "error")
			c.Abort()
			return
		}
		global.Logger.Info("切入反向代理中间件上下文")

		// step 1: modify realServer path
		req := c.Request
		req.URL.Scheme = realServer.Scheme
		req.URL.Host = realServer.Host
		req.URL.Path = realServer.Path

		// step 2: use http.Transport to do request to real server.
		transport := http.DefaultTransport
		res, err := transport.RoundTrip(req) //nolint:bodyclose
		if err != nil {
			global.Logger.Error("反代请求后端接口报错", zap.Error(err))
			c.String(http.StatusInternalServerError, "error")
			c.Abort()
			return
		}
		global.Logger.Info("反向代理中间件工作正常")
		c.Next()

		// step 3: return real server response to upstream.
		for k, v1 := range res.Header {
			for _, v2 := range v1 {
				c.Header(k, v2)
			}
		}
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(res.Body)
		_, _ = bufio.NewReader(res.Body).WriteTo(c.Writer)
	}
}
